<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# post

## Purpose
文章领域模型，定义文章、分类、标签、系列实体和 Repository 接口。

## Key Files
| File | Description |
|------|-------------|
| `models.go` | Post、Category、Tag、Series、PostVersion 结构定义 |
| `repository.go` | PostRepository、CategoryRepository 等接口定义 |

## For AI Agents

### Working In This Directory
- 文章内容使用 BlockDocument (JSONB) 存储
- 版本历史通过 PostVersion 管理

### Data Models
```go
type Post struct {
    ID            uuid.UUID
    AuthorID      uuid.UUID
    Title         string
    Slug          string
    Content       BlockDocument  // JSONB
    ContentText   string         // 纯文本，用于搜索
    Excerpt       string
    CategoryID    *uuid.UUID
    Tags          []Tag
    Status        PostStatus     // draft/published/scheduled/private
    PublishAt     *time.Time
    FeaturedImage string
    ViewCount     int
    LikeCount     int
    SeriesID      *uuid.UUID
    SeriesOrder   int
    Version       int
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type PostVersion struct {
    ID        uuid.UUID
    PostID    uuid.UUID
    Version   int
    Content   BlockDocument
    CreatorID uuid.UUID
    Note      string
    CreatedAt time.Time
}

type Category struct {
    ID          uuid.UUID
    Name        string
    Slug        string
    Description string
    ParentID    *uuid.UUID
    SortOrder   int
}

type Tag struct {
    ID   uuid.UUID
    Name string
    Slug string
}

type Series struct {
    ID          uuid.UUID
    AuthorID    uuid.UUID
    Title       string
    Slug        string
    Description string
}
```

### Post Status Flow
```
draft → published → (scheduled/private)
         ↑
      rollback
```

### Repository Interfaces
| Interface | Key Methods |
|-----------|-------------|
| PostRepository | Create, GetByID, GetBySlug, List, Update, Delete |
| CategoryRepository | Create, GetByID, GetBySlug, List, Update, Delete |
| TagRepository | Create, GetByID, GetBySlug, List, Delete |
| SeriesRepository | Create, GetByID, GetBySlug, List, Update, Delete |
| PostLikeRepository | Create, Delete, GetByUserAndPost |

<!-- MANUAL: -->