<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# media

## Purpose
媒体文件领域模型，定义媒体实体和 Repository 接口。

## Key Files
| File | Description |
|------|-------------|
| `models.go` | Media 结构定义 |
| `repository.go` | MediaRepository 接口定义 |

## For AI Agents

### Working In This Directory
- 支持图片、视频、文档等多种类型
- URL 由 Service 层生成，包含 base URL

### Data Model
```go
type Media struct {
    ID           uuid.UUID
    UploaderID   uuid.UUID
    Filename     string        // 存储文件名
    OriginalName string        // 原始文件名
    Filepath     string        // 文件路径
    URL          string        // 访问 URL
    MimeType     string        // MIME 类型
    Size         int64         // 字节数
    Width        *int          // 图片宽度
    Height       *int          // 图片高度
    AltText      string        // 替代文本
    Metadata     map[string]any // 其他元数据
    CreatedAt    time.Time
}
```

### Supported MIME Types
| Type | Extensions |
|------|------------|
| Image | jpg, jpeg, png, gif, webp, svg |
| Video | mp4, webm, mov |
| Document | pdf, doc, docx, xls, xlsx |
| Archive | zip, tar, gz |

### Repository Interface
```go
type MediaRepository interface {
    Create(ctx context.Context, media *Media) error
    GetByID(ctx context.Context, id uuid.UUID) (*Media, error)
    GetByUploader(ctx context.Context, uploaderID uuid.UUID, offset, limit int) ([]*Media, int, error)
    List(ctx context.Context, offset, limit int) ([]*Media, int, error)
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### Upload Directory Structure
```
uploads/
├── images/
│   └── {year}/
│       └── {month}/
│           └── {uuid}.{ext}
├── videos/
└── documents/
```

<!-- MANUAL: -->