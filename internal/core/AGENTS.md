<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# core - 领域模型层

## Purpose
`core` 目录包含 Cadmus 系统的核心领域模型定义。每个子目录代表一个独立的业务领域（Bounded Context），包含：
- **models.go**: 领域实体、值对象、枚举类型定义
- **repository.go**: Repository 接口定义（数据访问抽象）
- **errors.go**: 领域特定的错误类型定义
- **service.go**: 领域服务接口（如适用）

该目录**不包含**具体实现，仅定义接口和类型。实现位于 `internal/database/`（数据库实现）或 `internal/service/`（业务逻辑实现）。

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `user/` | 用户、角色、权限领域模型 |
| `post/` | 文章、分类、标签、系列领域模型 |
| `comment/` | 评论、回复、点赞领域模型 |
| `media/` | 媒体文件、上传管理领域模型 |
| `search/` | 全文搜索、搜索建议领域模型 |
| `rss/` | RSS 订阅生成领域模型 |
| `notify/` | 通知配置、消息推送领域模型 |

## Repository 接口定义规范

### 通用约定
所有 Repository 接口遵循以下命名和操作约定：

```go
// 基本 CRUD 操作
Create(ctx context.Context, entity *Entity) error
GetByID(ctx context.Context, id uuid.UUID) (*Entity, error)
Update(ctx context.Context, entity *Entity) error
Delete(ctx context.Context, id uuid.UUID) error

// 分页查询
List(ctx context.Context, filters Filters, offset, limit int) ([]*Entity, int, error)

// 特定查询
GetBySlug(ctx context.Context, slug string) (*Entity, error)
GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Entity, error)
```

### 各模块 Repository 接口

#### user.Repository
```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, offset, limit int) ([]*User, int, error)
}

type RoleRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
    GetByName(ctx context.Context, name string) (*Role, error)
    GetAll(ctx context.Context) ([]*Role, error)
    GetDefault(ctx context.Context) (*Role, error)
    GetWithPermissions(ctx context.Context, id uuid.UUID) (*Role, error)
}

type PermissionRepository interface {
    GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]Permission, error)
    GetAll(ctx context.Context) ([]Permission, error)
    GetByCategory(ctx context.Context, category string) ([]Permission, error)
    CheckPermission(ctx context.Context, roleID uuid.UUID, permissionName string) (bool, error)
}
```

#### post.Repository
```go
type PostRepository interface {
    Create(ctx context.Context, post *Post) error
    Update(ctx context.Context, post *Post) error
    Delete(ctx context.Context, id uuid.UUID) error
    GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
    GetBySlug(ctx context.Context, slug string) (*Post, error)
    List(ctx context.Context, filters PostListFilters, offset, limit int) ([]*Post, int, error)
    GetByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int) ([]*Post, int, error)
    GetByCategory(ctx context.Context, categoryID uuid.UUID, offset, limit int) ([]*Post, int, error)
    GetBySeries(ctx context.Context, seriesID uuid.UUID, offset, limit int) ([]*Post, int, error)
    Search(ctx context.Context, query string, offset, limit int) ([]*Post, int, error)
    IncrementViewCount(ctx context.Context, id uuid.UUID) error
    IncrementLikeCount(ctx context.Context, id uuid.UUID) error
    CreateVersion(ctx context.Context, version *PostVersion) error
    GetVersions(ctx context.Context, postID uuid.UUID) ([]*PostVersion, error)
    GetVersionByNumber(ctx context.Context, postID uuid.UUID, version int) (*PostVersion, error)
}

type CategoryRepository interface {
    Create(ctx context.Context, category *Category) error
    Update(ctx context.Context, category *Category) error
    Delete(ctx context.Context, id uuid.UUID) error
    GetByID(ctx context.Context, id uuid.UUID) (*Category, error)
    GetBySlug(ctx context.Context, slug string) (*Category, error)
    GetAll(ctx context.Context) ([]*Category, error)
    GetChildren(ctx context.Context, parentID uuid.UUID) ([]*Category, error)
    GetRootCategories(ctx context.Context) ([]*Category, error)
    GetPostCount(ctx context.Context, categoryID uuid.UUID) (int, error)
}

type TagRepository interface {
    Create(ctx context.Context, tag *Tag) error
    Delete(ctx context.Context, id uuid.UUID) error
    GetByID(ctx context.Context, id uuid.UUID) (*Tag, error)
    GetBySlug(ctx context.Context, slug string) (*Tag, error)
    GetByName(ctx context.Context, name string) (*Tag, error)
    GetAll(ctx context.Context) ([]*Tag, error)
    AddPostTag(ctx context.Context, postID, tagID uuid.UUID) error
    RemovePostTag(ctx context.Context, postID, tagID uuid.UUID) error
    GetPostTags(ctx context.Context, postID uuid.UUID) ([]*Tag, error)
    GetPostCount(ctx context.Context, tagID uuid.UUID) (int, error)
}

type SeriesRepository interface {
    Create(ctx context.Context, series *Series) error
    Update(ctx context.Context, series *Series) error
    Delete(ctx context.Context, id uuid.UUID) error
    GetByID(ctx context.Context, id uuid.UUID) (*Series, error)
    GetBySlug(ctx context.Context, slug string) (*Series, error)
    GetByAuthor(ctx context.Context, authorID uuid.UUID) ([]*Series, error)
    GetAll(ctx context.Context) ([]*Series, error)
}

type PostLikeRepository interface {
    CreateIfNotExists(ctx context.Context, postID, userID uuid.UUID) (created bool, err error)
    DeleteIfExists(ctx context.Context, postID, userID uuid.UUID) (deleted bool, err error)
    Exists(ctx context.Context, postID, userID uuid.UUID) (bool, error)
    CountByPostID(ctx context.Context, postID uuid.UUID) (int, error)
    GetByUserID(ctx context.Context, userID uuid.UUID) ([]*PostLike, error)
}
```

#### comment.Repository
```go
type CommentRepository interface {
    Create(ctx context.Context, input *CreateCommentInput) (*Comment, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Comment, error)
    GetByPostID(ctx context.Context, postID uuid.UUID, filters *CommentListFilters) ([]*Comment, error)
    GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Comment, error)
    GetChildren(ctx context.Context, parentID uuid.UUID) ([]*Comment, error)
    Update(ctx context.Context, comment *Comment) error
    UpdateStatus(ctx context.Context, id uuid.UUID, status CommentStatus) error
    Delete(ctx context.Context, id uuid.UUID) error
    CountByPostID(ctx context.Context, postID uuid.UUID) (int, error)
    CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
    List(ctx context.Context, filters *CommentListFilters, offset, limit int) ([]*Comment, error)
}

type CommentLikeRepository interface {
    Create(ctx context.Context, commentID, userID uuid.UUID) (*CommentLike, error)
    CreateIfNotExists(ctx context.Context, commentID, userID uuid.UUID) (created bool, err error)
    GetByCommentAndUser(ctx context.Context, commentID, userID uuid.UUID) (*CommentLike, error)
    Delete(ctx context.Context, commentID, userID uuid.UUID) error
    DeleteIfExists(ctx context.Context, commentID, userID uuid.UUID) (deleted bool, err error)
    Exists(ctx context.Context, commentID, userID uuid.UUID) (bool, error)
    CountByCommentID(ctx context.Context, commentID uuid.UUID) (int, error)
    GetByUserID(ctx context.Context, userID uuid.UUID) ([]*CommentLike, error)
}
```

#### media.Repository
```go
type MediaRepository interface {
    Create(ctx context.Context, input *UploadInput, filename, filepath, url string, width, height *int) (*Media, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Media, error)
    GetByUploaderID(ctx context.Context, uploaderID uuid.UUID) ([]*Media, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, filters *MediaListFilters, offset, limit int) ([]*Media, error)
    Count(ctx context.Context, filters *MediaListFilters) (int, error)
}
```

#### search.Repository
```go
type SearchRepository interface {
    Search(ctx context.Context, query string, filters SearchFilters, offset, limit int) ([]SearchResult, int, error)
    SearchByCategory(ctx context.Context, query string, categoryID uuid.UUID, offset, limit int) ([]SearchResult, int, error)
    SearchByAuthor(ctx context.Context, query string, authorID uuid.UUID, offset, limit int) ([]SearchResult, int, error)
    GetSuggestions(ctx context.Context, query string, limit int) ([]string, error)
}
```

#### rss/models.go
RSS 模块仅包含模型定义，Repository 接口位于 `post.Repository` 中（通过 `GetRSSFeed` 方法提供）。

#### notify/models.go
通知模块当前仅包含配置和模型定义，Repository 接口待扩展。

## For AI Agents

### 开发新领域模型时的步骤
1. 在 `internal/core/` 下创建新的子目录（如 `newfeature/`）
2. 创建 `models.go` 定义实体类型、枚举、常量
3. 创建 `repository.go` 定义数据访问接口
4. 创建 `errors.go` 定义领域特定错误（如 `ErrNotFound`、`ErrAlreadyExists`）
5. 确保接口方法都接受 `context.Context` 作为第一个参数
6. 使用 `uuid.UUID` 作为所有实体的 ID 类型

### Repository 实现规范
- 实现在 `internal/database/` 目录下的对应 package 中
- 必须支持 PostgreSQL/pgx 作为存储后端
- 所有错误必须使用 `core/{module}` 中定义的语义化错误类型
- 查询结果必须正确处理分页（offset/limit）和总数统计

### 跨模块依赖规则
- `core/` 下的模块可以相互引用类型（如 `post.Post` 引用 `user.User` 的 AuthorID）
- 避免循环依赖：如果 A 模块引用 B 模块，B 模块不应引用 A 模块
- Repository 接口只定义数据访问方法，不包含业务逻辑
- 业务逻辑应放在 `internal/service/` 层

### 错误类型命名约定
```go
// 通用错误模式
ErrNotFound          // 实体不存在
ErrAlreadyExists     // 实体已存在（创建时）
ErrPermissionDenied  // 权限不足
ErrInvalidInput      // 输入参数无效

// 特定实体错误（使用前缀区分）
ErrUserNotFound
ErrPostNotFound
ErrCommentNotFound
// ...
```
