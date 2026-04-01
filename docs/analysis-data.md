# Cadmus 数据层分析报告

## 概述

Cadmus 项目采用典型的三层架构设计，数据层位于 `internal/database/` 目录，负责与 PostgreSQL 数据库的所有交互。本报告分析数据模型设计、数据库迁移策略、Repository 层实现、UUID 主键使用以及数据访问模式。

---

## 1. 数据模型设计

数据模型定义在 `internal/core/` 下的各模块中，遵循领域驱动设计原则，每个模块独立管理其实体定义。

### 1.1 模块分布

| 模块 | 文件位置 | 主要实体 |
|------|----------|----------|
| user | `internal/core/user/models.go` | User, Role, Permission |
| post | `internal/core/post/models.go` | Post, Category, Tag, Series, PostVersion, PostLike |
| comment | `internal/core/comment/models.go` | Comment, CommentLike |
| media | `internal/core/media/models.go` | Media |
| notify | `internal/core/notify/models.go` | Notification |
| search | `internal/core/search/models.go` | SearchResult, SearchFilters |
| rss | `internal/core/rss/models.go` | RSSFeed, RSSChannel, RSSItem |

### 1.2 核心实体结构分析

#### User 实体

```go
type User struct {
    ID           uuid.UUID  `json:"id"`
    Username     string     `json:"username"`
    Email        string     `json:"email"`
    PasswordHash string     `json:"-"`            // 安全考虑，不暴露密码哈希
    AvatarURL    string     `json:"avatar_url,omitempty"`
    Bio          string     `json:"bio,omitempty"`
    RoleID       uuid.UUID  `json:"role_id"`      // 外键关联 Role
    Status       UserStatus `json:"status"`       // 枚举: active/banned/pending
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}
```

**设计特点：**
- 使用 `json:"-"` 隐藏敏感字段（PasswordHash）
- 内置密码哈希方法（SetPassword/CheckPassword），使用 bcrypt
- 用户状态采用枚举类型并带验证方法 `IsValid()`
- 自定义错误类型实现 `errors.Is` 接口，便于错误匹配

#### Post 实体

```go
type Post struct {
    ID           uuid.UUID  `json:"id"`
    AuthorID     uuid.UUID  `json:"author_id"`
    Title        string     `json:"title"`
    Slug         string     `json:"slug"`         // URL友好标识
    Content      []byte     `json:"content"`      // BlockDocument JSON (JSONB)
    ContentText  string     `json:"content_text"` // 纯文本用于搜索
    Excerpt      string     `json:"excerpt"`      // 文章摘要
    CategoryID   uuid.UUID  `json:"category_id"`  // 可为空
    Tags         []*Tag     `json:"tags,omitempty"` // 不存储在表中，关联查询
    Status       PostStatus `json:"status"`       // draft/published/scheduled/private
    PublishAt    *time.Time `json:"publish_at"`   // 定时发布
    FeaturedImage string    `json:"featured_image,omitempty"`
    SEOMeta      SEOMeta    `json:"seo_meta"`     // 内嵌结构体
    ViewCount    int        `json:"view_count"`
    LikeCount    int        `json:"like_count"`
    CommentCount int        `json:"comment_count"`
    SeriesID     *uuid.UUID `json:"series_id"`    // 可为空
    SeriesOrder  int        `json:"series_order"`
    IsPaid       bool       `json:"is_paid"`      // 付费文章标志
    Price        *float64   `json:"price"`        // 付费价格
    Version      int        `json:"version"`      // 版本号，用于版本控制
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}
```

**设计特点：**
- 内容存储为 JSONB（`[]byte`），支持 BlockDocument 结构化内容
- 分离 `Content`（JSON）和 `ContentText`（纯文本），优化全文搜索
- SEO 元数据内嵌结构体设计
- 版本控制字段（Version）支持文章历史版本管理
- 付费内容支持（IsPaid/Price）
- 统计字段（ViewCount/LikeCount/CommentCount）冗余存储，减少聚合查询

#### Comment 实体

```go
type Comment struct {
    ID        uuid.UUID     `json:"id"`
    PostID    uuid.UUID     `json:"post_id"`
    UserID    uuid.UUID     `json:"user_id"`
    ParentID  *uuid.UUID    `json:"parent_id"`   // 父评论ID，支持嵌套
    Depth     int           `json:"depth"`       // 嵌套深度（最大5层）
    Content   string        `json:"content"`
    Status    CommentStatus `json:"status"`      // pending/approved/spam/deleted
    LikeCount int           `json:"like_count"`
    CreatedAt time.Time     `json:"created_at"`
    UpdatedAt time.Time     `json:"updated_at"`
}
```

**设计特点：**
- 支持嵌套评论（ParentID + Depth）
- 深度限制（最大5层）在数据库层和代码层双重约束
- 软删除设计（Status = deleted）
- 评论状态审核机制

### 1.3 枚举设计模式

所有枚举类型采用统一设计模式：

```go
type UserStatus string

const (
    StatusActive   UserStatus = "active"
    StatusBanned   UserStatus = "banned"
    StatusPending  UserStatus = "pending"
)

func (s UserStatus) IsValid() bool {
    switch s {
    case StatusActive, StatusBanned, StatusPending:
        return true
    default:
        return false
    }
}
```

**优点：**
- 类型安全（string 类型别名）
- 序列化友好（直接 JSON 编码）
- 验证方法内置
- 数据库存储紧凑

### 1.4 错误设计

每个模块定义独立的错误类型：

```go
type UserError struct {
    Code    string
    Message string
}

func (e *UserError) Error() string { return e.Message }

func (e *UserError) Is(target error) bool {
    t, ok := target.(*UserError)
    if !ok { return false }
    return e.Code == t.Code
}
```

**设计特点：**
- 错误码 + 消息结构
- 实现 `errors.Is` 接口，支持错误精确匹配
- 预定义常见错误常量，统一错误处理

---

## 2. 数据库迁移策略

迁移文件位于 `migrations/` 目录，采用编号命名规范。

### 2.1 迁移文件列表

| 编号 | 文件名 | 主要内容 |
|------|--------|----------|
| 001 | init.up.sql | 基础表：users, roles, permissions, role_permissions |
| 002 | create_posts.up.sql | 文章相关：posts, categories, tags, series, post_tags, post_versions |
| 003 | create_comments.up.sql | 评论相关：comments, comment_likes |
| 004 | create_post_likes.up.sql | 文章点赞：post_likes |
| 005 | create_media.up.sql | 媒体文件：media |

### 2.2 迁移设计特点

#### UUID 扩展启用

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

所有主键使用 `uuid_generate_v4()` 自动生成。

#### 自动更新触发器

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';
```

每个表创建对应触发器：

```sql
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

#### 全文搜索支持

```sql
ALTER TABLE posts ADD COLUMN search_vector tsvector;
CREATE INDEX posts_search_idx ON posts USING GIN(search_vector);

CREATE OR REPLACE FUNCTION update_posts_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(NEW.content_text, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(NEW.excerpt, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

**搜索权重设计：**
- 权重 A（最高）：标题
- 权重 B：内容纯文本
- 权重 C：摘要

#### 评论深度自动计算

```sql
CREATE OR REPLACE FUNCTION calculate_comment_depth()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.depth := 0;
    ELSE
        SELECT depth + 1 INTO NEW.depth
        FROM comments WHERE id = NEW.parent_id;
        IF NEW.depth > 5 THEN
            RAISE EXCEPTION 'Comment depth exceeds maximum limit of 5';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

### 2.3 索引策略

**单列索引：**
- 所有外键字段（author_id, role_id, post_id 等）
- 查询频繁字段（slug, status, email, username）

**复合索引：**
- `posts_search_idx`：GIN 索引支持全文搜索
- 条件索引：`idx_posts_publish_at WHERE status = 'published'`

**唯一约束：**
- `users_username_key`, `users_email_key`
- `posts_slug_key`
- `tags_name_key`, `tags_slug_key`
- `categories_slug_key`

### 2.4 初始数据植入

迁移脚本包含默认数据：

**权限定义（21个）：**
- 文章权限：post.create, post.edit, post.delete, post.publish, post.view
- 评论权限：comment.create, comment.edit, comment.delete, comment.view
- 用户权限：user.view, user.edit, user.delete, user.manage
- 主题权限：theme.view, theme.edit, theme.install
- 插件权限：plugin.view, plugin.edit, plugin.install
- 系统权限：system.admin, system.settings

**角色定义（5个）：**
- admin：所有权限
- editor：文章和评论权限
- author：创建编辑自己内容
- user：基础权限（默认角色）
- guest：只读权限

---

## 3. Repository 层实现

Repository 层位于 `internal/database/`，共 10 个仓库文件。

### 3.1 Repository 文件列表

| 文件 | Repository 结构体 | 主要职责 |
|------|-------------------|----------|
| user_repository.go | UserRepository | 用户 CRUD、查询 |
| role_repository.go | RoleRepository | 角色 CRUD、权限管理 |
| permission_repository.go | PermissionRepository | 权限查询 |
| post_repository.go | PostRepository, PostLikeRepository | 文章 CRUD、版本管理、点赞 |
| category_repository.go | CategoryRepository | 分类树管理 |
| tag_repository.go | TagRepository | 标签管理、文章标签关联 |
| series_repository.go | SeriesRepository | 系列管理 |
| comment_repository.go | CommentRepository, CommentLikeRepository | 评论 CRUD、点赞 |
| media_repository.go | MediaRepository | 媒体文件管理 |
| search_repository.go | SearchRepository | 全文搜索 |

### 3.2 连接池设计

```go
type Pool struct {
    *pgxpool.Pool
}

func NewPool(ctx context.Context, cfg Config) (*Pool, error) {
    poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
    // 配置连接池参数
    poolCfg.MaxConns = int32(cfg.MaxOpenConns)      // 默认 25
    poolCfg.MinConns = int32(cfg.MaxIdleConns)      // 默认 5
    poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime   // 默认 5分钟
    poolCfg.MaxConnIdleTime = cfg.ConnMaxIdleTime   // 默认 10分钟

    pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
    // 验证连接
    if err := pool.Ping(ctx); err != nil { ... }
    return &Pool{pool}, nil
}
```

**连接池参数：**
- MaxConns：最大连接数（25）
- MinConns：最小空闲连接（5）
- MaxConnLifetime：连接最大生命周期（5分钟）
- MaxConnIdleTime：空闲连接超时（10分钟）

### 3.3 事务管理

```go
type TransactionManager struct {
    pool *Pool
}

func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
    tx, err := tm.pool.Begin(ctx)
    if err != nil { return err }

    defer tx.Rollback(ctx)  // 安全回滚

    if err := fn(tx); err != nil {
        return err  // 回滚在 defer 中执行
    }

    return tx.Commit(ctx)
}
```

**事务使用示例（RoleRepository.SetPermissions）：**

```go
func (r *RoleRepository) SetPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
    if r.txManager != nil {
        return r.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
            return r.setPermissionsWithTx(ctx, tx, roleID, permissionIDs)
        })
    }
    // 传统事务方式备用
    tx, err := r.pool.Begin(ctx)
    ...
}
```

### 3.4 Repository 实现模式

**标准 CRUD 模式：**

```go
type UserRepository struct {
    pool *Pool
}

func NewUserRepository(pool *Pool) *UserRepository {
    return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
    query := `INSERT INTO users (...) VALUES (...)`
    // UUID 自动生成
    if u.ID == uuid.Nil {
        u.ID = uuid.New()
    }
    // 时间自动填充
    now := time.Now()
    _, err := r.pool.Exec(ctx, query, ...)
    // 错误处理：唯一约束冲突
    if isUniqueViolation(err, "users_username_key") {
        return user.ErrUserAlreadyExists
    }
    return nil
}
```

**错误处理模式：**

```go
func isUniqueViolation(err error, constraintName string) bool {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
    }
    return false
}
```

**扫描辅助方法：**

```go
func (r *UserRepository) scanUser(ctx context.Context, row pgx.Row) (*user.User, error) {
    u := &user.User{}
    err := row.Scan(&u.ID, &u.Username, ...)
    if err != nil { return nil, err }
    return u, nil
}

func (r *UserRepository) scanUserFromRow(row pgx.Rows) (*user.User, error) {
    // 用于批量查询
}
```

### 3.5 高级查询特性

**分页查询（UserRepository.List）：**

```go
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, int, error) {
    // 先获取总数
    countQuery := `SELECT COUNT(*) FROM users`
    var total int
    r.pool.QueryRow(ctx, countQuery).Scan(&total)

    // 分页查询
    query := `SELECT ... FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
    rows, err := r.pool.Query(ctx, query, limit, offset)
    ...
    return users, total, nil
}
```

**动态筛选（PostRepository.List）：**

```go
func (r *PostRepository) List(ctx context.Context, filters post.PostListFilters, offset, limit int) ([]*post.Post, int, error) {
    whereClause := "WHERE 1=1"
    args := make([]interface{}, 0)
    argIndex := 1

    if filters.Status != "" {
        whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
        args = append(args, filters.Status)
        argIndex++
    }
    // 其他筛选条件...
    // 标签筛选使用 EXISTS 子查询
    if filters.TagID != uuid.Nil {
        whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM post_tags WHERE post_tags.post_id = posts.id AND post_tags.tag_id = $%d)", argIndex)
    }
    // 全文搜索
    if filters.Search != "" {
        whereClause += fmt.Sprintf(" AND search_vector @@ websearch_to_tsquery('simple', $%d)", argIndex)
    }
    ...
}
```

**批量操作（PostRepository.CountByAuthors）：**

```go
func (r *PostRepository) CountByAuthors(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]int, error) {
    query := `
        SELECT author_id, COUNT(*) as count
        FROM posts WHERE author_id = ANY($1)
        GROUP BY author_id
    `
    rows, err := r.pool.Query(ctx, query, userIDs)
    ...
    result := make(map[uuid.UUID]int)
    for rows.Next() {
        var authorID uuid.UUID
        var count int
        rows.Scan(&authorID, &count)
        result[authorID] = count
    }
    return result, nil
}
```

**原子操作（PostLikeRepository.CreateIfNotExists）：**

```go
func (r *PostLikeRepository) CreateIfNotExists(ctx context.Context, postID, userID uuid.UUID) (created bool, err error) {
    query := `
        INSERT INTO post_likes (id, post_id, user_id, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (post_id, user_id) DO NOTHING
    `
    result, err := r.pool.Exec(ctx, query, id, postID, userID, now)
    created = result.RowsAffected() > 0

    // 只有实际创建时才更新计数
    if created {
        updateQuery := `UPDATE posts SET like_count = like_count + 1 WHERE id = $1`
        r.pool.Exec(ctx, updateQuery, postID)
    }
    return created, nil
}
```

### 3.6 全文搜索实现

```go
func (r *SearchRepository) Search(ctx context.Context, query string, filters search.SearchFilters, offset, limit int) ([]search.SearchResult, int, error) {
    baseQuery := `
        SELECT p.id, p.title, p.slug, p.excerpt, p.author_id, p.category_id, p.status,
               ts_rank(p.search_vector, websearch_to_tsquery('simple', $1)) as rank,
               p.created_at, p.updated_at
        FROM posts p
        WHERE p.search_vector @@ websearch_to_tsquery('simple', $1)
    `
    // 动态筛选条件...
    // 按相关性排序
    fullQuery := baseQuery + whereClause + " ORDER BY rank DESC LIMIT ..."
    ...
}
```

**搜索建议功能：**

```go
func (r *SearchRepository) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
    sql := `
        SELECT DISTINCT word
        FROM (
            SELECT unnest(string_to_array(lower(title), ' ')) as word
            FROM posts WHERE status = 'published' AND lower(title) LIKE lower($1)
        ) sub
        WHERE length(word) >= 2
        ORDER BY word LIMIT $2
    `
    ...
}
```

---

## 4. PostgreSQL UUID 主键使用

### 4.1 UUID 生成策略

**数据库层生成：**

```sql
id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
```

**代码层生成：**

```go
if u.ID == uuid.Nil {
    u.ID = uuid.New()
}
```

**双重生成保障：**
- 数据库默认值作为后备
- 代码层优先生成，确保返回给调用者

### 4.2 UUID 使用场景

| 场景 | 使用方式 |
|------|----------|
| 主键 | 所有实体 ID |
| 外键 | 关联关系（author_id, role_id, post_id 等） |
| 可空外键 | parent_id, category_id, series_id |
| 查询参数 | GetByID, Delete 等方法 |

### 4.3 UUID 零值处理

Go 的 `uuid.UUID` 类型底层是 `[16]byte`，零值为 `uuid.Nil`（全0）。

**可空外键处理：**

```go
// PostRepository.Create
var categoryID interface{}
if p.CategoryID != uuid.Nil {
    categoryID = p.CategoryID
}
// 插入时传入 interface{}，nil 表示数据库 NULL
```

**外键定义：**

```sql
category_id UUID REFERENCES categories(id) ON DELETE SET NULL
-- 可为空，删除分类时置空
```

---

## 5. 数据访问模式评估

### 5.1 架构模式

**Repository 模式特点：**
- 领域模型与数据访问分离
- 每个 Repository 对应一个聚合根
- 统一的 CRUD 接口

**依赖方向：**
```
internal/core (领域模型) ← internal/database (Repository)
```
Repository 依赖领域模型，符合 DDD 原则。

### 5.2 查询优化策略

**冗余计数字段：**
- Post 实体的 view_count, like_count, comment_count
- Comment 实体的 like_count
- 避免高频聚合查询

**索引覆盖：**
- 所有常用查询字段建立索引
- 条件索引优化特定查询（如定时发布）

**批量查询：**
- `CountByAuthors` 批量统计
- `GetByIDs` 批量获取
- `GetLikesBatch` 批量检查点赞状态

**全文搜索：**
- PostgreSQL tsvector + GIN 索引
- 自动更新触发器
- 权重分层（标题 > 内容 > 摘要）

### 5.3 数据一致性保障

**事务使用：**
- RoleRepository.SetPermissions：删除旧权限 + 插入新权限
- 点赞操作：创建点赞记录 + 更新计数（两步操作）

**唯一约束：**
- 防止重复数据（slug, username, email）
- 点赞/评论点赞的唯一约束

**外键约束：**
- CASCADE 删除：posts → author 删除时级联
- SET NULL：posts → category 删除时置空
- RESTRICT：users → role 删除时阻止

**并发控制：**
- `ON CONFLICT DO NOTHING`：幂等点赞操作
- `RowsAffected()` 检查：确认操作生效

### 5.4 软删除策略

**用户删除：**
```go
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
    query := `UPDATE users SET status = 'banned' WHERE id = $1`
    // 软删除，保留数据
}
```

**评论删除：**
```go
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
    return r.UpdateStatus(ctx, id, comment.StatusDeleted)
    // 软删除，标记为 deleted
}
```

**文章删除：**
```go
func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
    query := `DELETE FROM posts WHERE id = $1`
    // 硬删除
}
```

**不同策略原因：**
- 用户/评论：可能需要历史记录或审计
- 文章：版本历史表已保存版本，主表可硬删除

### 5.5 潜在改进建议

1. **批量插入优化**
   - 当前权限设置逐条插入
   - 可优化为批量 INSERT VALUES

2. **连接池监控**
   - 已有 Stats() 方法
   - 可添加健康检查端点

3. **查询缓存**
   - 权限查询可缓存（角色权限变更较少）
   - 分类/标签树可缓存

4. **读写分离准备**
   - Repository 模式便于扩展
   - 可添加 ReadPool 和 WritePool

5. **审计日志**
   - 当前无审计表
   - 可添加操作日志记录

---

## 6. 总结

Cadmus 数据层设计总体评价：

**优点：**
- 清晰的 Repository 模式实现
- UUID 主键统一使用
- 完善的索引策略
- 全文搜索集成
- 错误处理规范（自定义错误类型 + errors.Is）
- 事务管理封装
- 软删除与硬删除策略合理区分

**设计亮点：**
- 评论嵌套深度双重约束（数据库触发器 + 代码验证）
- 搜索向量自动更新
- 冗余计数字段减少聚合查询
- ON CONFLICT 幂等操作

**可改进方向：**
- 批量操作优化
- 查询缓存引入
- 审计日志补充
- 健康监控集成