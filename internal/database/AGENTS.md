<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# database

## Purpose
PostgreSQL 数据访问层，实现 Repository 接口。

## Key Files
| File | Description |
|------|-------------|
| `config.go` | 数据库配置结构 |
| `connection.go` | 连接池管理 |
| `user_repository.go` | 用户 Repository 实现 |
| `role_repository.go` | 角色 Repository 实现 |
| `permission_repository.go` | 权限 Repository 实现 |
| `post_repository.go` | 文章 Repository 实现 |
| `category_repository.go` | 分类 Repository 实现 |
| `tag_repository.go` | 标签 Repository 实现 |
| `series_repository.go` | 系列 Repository 实现 |
| `comment_repository.go` | 评论 Repository 实现 |
| `media_repository.go` | 媒体 Repository 实现 |
| `search_repository.go` | 搜索 Repository 实现 |

## For AI Agents

### Working In This Directory
- 使用 `pgx/v5` 驱动，支持连接池
- 实现 `internal/core/*/repository.go` 定义的接口
- SQL 查询使用参数化防止注入

### Connection Pool Config
| Parameter | Default | Description |
|-----------|---------|-------------|
| `MaxConns` | 25 | 最大连接数 |
| `MinConns` | 5 | 最小连接数 |
| `MaxConnLifetime` | 1h | 连接最大生命周期 |
| `MaxConnIdleTime` | 15m | 空闲连接超时 |

### Repository Pattern
```go
type PostRepository struct {
    db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
    return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *Post) error {
    query := `INSERT INTO posts (id, author_id, title, slug, content, status)
              VALUES ($1, $2, $3, $4, $5, $6)`
    _, err := r.db.Exec(ctx, query,
        post.ID, post.AuthorID, post.Title, post.Slug, post.Content, post.Status)
    return err
}
```

### Query Optimization
- 使用 `pgxpool.Pool` 连接池
- 批量操作使用 `CopyFrom` 或事务
- 复杂查询使用 `sqlc` 生成（可选）

<!-- MANUAL: -->