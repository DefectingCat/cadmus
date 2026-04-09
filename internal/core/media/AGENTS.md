<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# media - 媒体文件领域模型

## Purpose

`media` 目录包含 Cadmus 系统媒体文件管理的核心领域模型定义。该模块负责：

- 媒体文件实体定义（图片、文档等）
- 上传输入参数结构
- 媒体列表筛选条件
- 数据访问 Repository 接口
- 语义化错误类型定义
- MIME 类型白名单管理

**不包含**具体实现。实现位于 `internal/database/`（数据库持久化）和 `internal/service/`（业务逻辑）。

## Key Files

| File | Purpose |
|------|---------|
| `models.go` | 媒体实体 `Media`、上传输入 `UploadInput`、筛选条件 `MediaListFilters`、错误类型、MIME 白名单 |
| `repository.go` | `MediaRepository` 接口定义（CRUD 和查询方法） |
| `models_test.go` | 错误类型和 MIME 类型工具函数的单元测试 |

## Domain Models

### Media 实体

表示系统中上传的媒体文件，包含文件信息和元数据：

```go
type Media struct {
    ID           uuid.UUID  // 唯一标识符
    UploaderID   uuid.UUID  // 上传者用户 ID
    Filename     string     // 系统生成的唯一文件名
    OriginalName string     // 用户上传时的原始文件名
    FilePath     string     // 相对于存储根目录的路径
    URL          string     // 完整可访问 URL
    MimeType     string     // MIME 类型（如 "image/jpeg"）
    Size         int64      // 文件大小（字节）
    Width        *int       // 图片宽度（仅图片类型）
    Height       *int       // 图片高度（仅图片类型）
    AltText      *string    // 无障碍访问替代文本
    CreatedAt    time.Time  // UTC 上传时间
}
```

### UploadInput 结构

上传文件请求的输入参数：

```go
type UploadInput struct {
    UploaderID   uuid.UUID
    OriginalName string
    MimeType     string
    Size         int64
    AltText      *string  // 可选
}
```

### MediaListFilters 结构

媒体查询的筛选条件：

```go
type MediaListFilters struct {
    UploaderID uuid.UUID  // 按上传者筛选
    MimeType   string     // 按 MIME 类型筛选（支持前缀匹配，如 "image/"）
}
```

## Repository Interface

`MediaRepository` 接口定义媒体数据访问方法：

```go
type MediaRepository interface {
    // 创建媒体记录
    Create(ctx context.Context, input *UploadInput, filename, filepath, url string, width, height *int) (*Media, error)
    
    // 根据 ID 获取媒体
    GetByID(ctx context.Context, id uuid.UUID) (*Media, error)
    
    // 获取用户上传的所有媒体
    GetByUploaderID(ctx context.Context, uploaderID uuid.UUID) ([]*Media, error)
    
    // 删除媒体记录
    Delete(ctx context.Context, id uuid.UUID) error
    
    // 分页获取媒体列表
    List(ctx context.Context, filters *MediaListFilters, offset, limit int) ([]*Media, error)
    
    // 统计媒体数量
    Count(ctx context.Context, filters *MediaListFilters) (int, error)
}
```

## Error Types

语义化错误类型，便于错误处理：

| Error | Code | Message |
|-------|------|---------|
| `ErrMediaNotFound` | `media_not_found` | 媒体文件不存在 |
| `ErrInvalidMimeType` | `invalid_mime_type` | 不支持的文件类型 |
| `ErrFileSizeTooLarge` | `file_size_too_large` | 文件大小超过限制 |
| `ErrPermissionDenied` | `permission_denied` | 权限不足 |
| `ErrUploadFailed` | `upload_failed` | 上传失败 |

`MediaError` 类型实现 `error` 和 `errors.Is` 接口，支持错误比较：

```go
if errors.Is(err, media.ErrMediaNotFound) {
    // 处理文件不存在的情况
}
```

## MIME Type Management

### AllowedMimeTypes 白名单

支持的文件类型：

- **图片**: `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `image/svg+xml`
- **文档**: `application/pdf`, `application/msword`, `application/vnd.openxmlformats-officedocument.wordprocessingml.document`
- **其他**: `application/zip`, `text/plain`

### IsImageMimeType 工具函数

判断 MIME 类型是否为图片，用于决定是否提取尺寸信息：

```go
if media.IsImageMimeType(input.MimeType) {
    // 提取图片宽高
    width, height := getImageDimensions(file)
    // ...
}
```

## Subdirectories

无子目录。

## For AI Agents

### 开发媒体相关功能时

1. **添加新的媒体类型**：在 `AllowedMimeTypes` 中添加 MIME 类型
2. **扩展 Media 实体**：在 `models.go` 中添加字段，注意使用指针类型表示可选字段
3. **新增查询方法**：在 `repository.go` 中定义接口方法，实现位于 `internal/database/`
4. **添加错误类型**：在 `models.go` 中定义新的 `Err*` 变量，遵循命名约定

### Repository 实现规范

- 实现在 `internal/database/media_repository.go`
- 所有方法必须支持 `context.Context` 超时控制
- 返回错误必须使用 `models.go` 中定义的语义化错误
- 文件删除操作应同时删除数据库记录和实际文件

### 跨模块依赖

- `Media.UploaderID` 引用 `core/user` 模块的 `User` 实体 ID
- 避免循环依赖：`media` 模块不应直接引用 `user` 模块的类型定义

### 测试要点

- 错误类型的 `Error()` 和 `Is()` 方法行为
- `IsImageMimeType` 对各种 MIME 类型的识别准确性
- `AllowedMimeTypes` 白名单与 `IsImageMimeType` 的一致性
