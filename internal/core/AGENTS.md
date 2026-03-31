<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# core

## Purpose
核心领域模型目录，定义业务实体和 Repository 接口。

## Key Files
| File | Description |
|------|-------------|
| 无顶层文件 | 按领域分子目录 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `user/` | 用户、角色、权限领域模型 (see `user/AGENTS.md`) |
| `post/` | 文章、分类、标签、系列领域模型 (see `post/AGENTS.md`) |
| `comment/` | 评论领域模型 (see `comment/AGENTS.md`) |
| `media/` | 媒体文件领域模型 (see `media/AGENTS.md`) |
| `search/` | 搜索领域模型 (see `search/AGENTS.md`) |
| `rss/` | RSS 领域模型 (see `rss/AGENTS.md`) |
| `notify/` | 通知领域模型 (see `notify/AGENTS.md`) |
| `block/` | 块编辑器领域模型 (待创建) |

## For AI Agents

### Working In This Directory
- **仅定义模型和接口**，不包含实现
- Repository 接口定义在此，实现在 `internal/database/`
- 领域模型应纯净，不依赖外部框架

### Domain Model Structure
```go
// core/post/models.go
type Post struct {
    ID          uuid.UUID
    AuthorID    uuid.UUID
    Title       string
    Slug        string
    Content     BlockDocument
    Status      PostStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// core/post/repository.go
type PostRepository interface {
    Create(ctx context.Context, post *Post) error
    GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
    GetBySlug(ctx context.Context, slug string) (*Post, error)
    List(ctx context.Context, filters PostListFilters, offset, limit int) ([]*Post, int, error)
    Update(ctx context.Context, post *Post) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### Cross-Aggregate Rules
- `core/` 下的各模块可相互引用类型
- 但应避免循环依赖
- Repository 接口只定义数据访问，不包含业务逻辑

<!-- MANUAL: -->