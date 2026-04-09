<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# post - 文章领域模型

## Purpose

`post` 目录包含 Cadmus 博客系统的文章管理核心领域模型。定义文章、分类、标签、系列相关的数据实体、值对象、枚举类型及数据访问接口。

**主要职责**：
- 定义文章实体的完整数据结构（标题、内容、状态、SEO 元数据等）
- 定义分类、标签、系列的实体结构
- 定义文章版本历史记录和点赞记录
- 定义 Repository 接口抽象数据访问层
- 定义领域特定的错误类型

**不包含**：具体实现逻辑（位于 `internal/database/` 和 `internal/service/`）

## Files

| File | Purpose |
|------|---------|
| `models.go` | 领域实体、枚举、错误类型定义 |
| `repository.go` | Repository 接口定义（数据访问抽象） |
| `models_test.go` | 模型层单元测试 |

## Domain Models

### 实体概览

| Entity | Description |
|--------|-------------|
| `Post` | 文章实体，包含内容、元数据、状态等完整信息 |
| `Category` | 分类实体，支持父子层级关系 |
| `Tag` | 标签实体，用于文章横向标记和聚合 |
| `Series` | 文章系列实体，用于组织连载文章 |
| `PostVersion` | 文章版本历史，支持回溯和恢复 |
| `PostLike` | 文章点赞记录 |
| `SEOMeta` | SEO 元数据值对象 |
| `PostListFilters` | 文章列表筛选条件 |

### PostStatus 枚举

| Status | Description |
|--------|-------------|
| `StatusDraft` | 草稿状态，仅作者可见 |
| `StatusPublished` | 已发布状态，公开可见 |
| `StatusScheduled` | 定时发布，等待设定时间到达 |
| `StatusPrivate` | 私密状态，仅特定用户可见 |

### Post 结构体

```go
type Post struct {
    ID            uuid.UUID      // 文章唯一标识符
    AuthorID      uuid.UUID      // 作者用户 ID
    Title         string         // 文章标题
    Slug          string         // URL 别名
    Content       []byte         // 文章内容 (BlockDocument JSONB)
    ContentText   string         // 纯文本，用于全文搜索
    Excerpt       string         // 摘要
    CategoryID    uuid.UUID      // 分类 ID (零值表示未分类)
    Tags          []*Tag         // 标签列表
    Status        PostStatus     // 文章状态
    PublishAt     *time.Time     // 定时发布时间
    FeaturedImage string         // 特色图片 URL
    SEOMeta       SEOMeta        // SEO 元数据
    ViewCount     int            // 浏览次数
    LikeCount     int            // 点赞次数
    CommentCount  int            // 评论次数
    SeriesID      *uuid.UUID     // 系列 ID (可为空)
    SeriesOrder   int            // 系列内排序
    IsPaid        bool           // 是否付费文章
    Price         *float64       // 价格 (IsPaid=true 时有效)
    Version       int            // 版本号
    CreatedAt     time.Time      // 创建时间 (UTC)
    UpdatedAt     time.Time      // 最后修改时间 (UTC)
}
```

### Category 结构体

```go
type Category struct {
    ID          uuid.UUID  // 分类唯一标识符
    Name        string     // 分类名称
    Slug        string     // URL 别名
    Description string     // 分类描述
    ParentID    *uuid.UUID // 父分类 ID (空表示顶级)
    SortOrder   int        // 排序权重
    CreatedAt   time.Time  // 创建时间
    UpdatedAt   time.Time  // 最后修改时间
}
```

### Tag 结构体

```go
type Tag struct {
    ID        uuid.UUID // 标签唯一标识符
    Name      string    // 标签名称
    Slug      string    // URL 别名
    CreatedAt time.Time // 创建时间
}
```

### Series 结构体

```go
type Series struct {
    ID          uuid.UUID  // 系列唯一标识符
    AuthorID    uuid.UUID  // 创建者用户 ID
    Title       string     // 系列标题
    Slug        string     // URL 别名
    Description string     // 系列描述
    CoverImage  string     // 封面图片 URL
    CreatedAt   time.Time  // 创建时间
    UpdatedAt   time.Time  // 最后修改时间
}
```

### PostVersion 结构体

```go
type PostVersion struct {
    ID        uuid.UUID // 版本记录 ID
    PostID    uuid.UUID // 所属文章 ID
    Version   int       // 版本号
    Content   []byte    // 内容快照
    CreatorID uuid.UUID // 创建者 ID
    Note      string    // 版本说明
    CreatedAt time.Time // 创建时间
}
```

### PostLike 结构体

```go
type PostLike struct {
    ID        uuid.UUID // 点赞记录 ID
    PostID    uuid.UUID // 文章 ID
    UserID    uuid.UUID // 用户 ID
    CreatedAt time.Time // 点赞时间
}
```

### SEOMeta 结构体

```go
type SEOMeta struct {
    Title       string   `json:"title,omitempty"`       // SEO 标题
    Description string   `json:"description,omitempty"` // SEO 描述
    Keywords    []string `json:"keywords,omitempty"`    // SEO 关键词
}
```

## Repository Interfaces

### PostRepository

文章数据访问接口，支持：
- 基本 CRUD：`Create`、`Update`、`Delete`、`GetByID`、`GetBySlug`
- 查询：`List`（带筛选）、`GetByAuthor`、`GetByCategory`、`GetBySeries`、`Search`
- 统计：`IncrementViewCount`、`IncrementLikeCount`
- 版本管理：`CreateVersion`、`GetVersions`、`GetVersionByNumber`

### CategoryRepository

分类数据访问接口，支持：
- 基本 CRUD：`Create`、`Update`、`Delete`、`GetByID`、`GetBySlug`
- 层级查询：`GetAll`、`GetChildren`、`GetRootCategories`
- 统计：`GetPostCount`

### TagRepository

标签数据访问接口，支持：
- 基本操作：`Create`、`Delete`、`GetByID`、`GetBySlug`、`GetByName`、`GetAll`
- 关联管理：`AddPostTag`、`RemovePostTag`、`GetPostTags`
- 统计：`GetPostCount`

### SeriesRepository

系列数据访问接口，支持：
- 基本 CRUD：`Create`、`Update`、`Delete`、`GetByID`、`GetBySlug`
- 查询：`GetByAuthor`、`GetAll`

### PostLikeRepository

点赞记录数据访问接口，支持：
- 原子操作：`CreateIfNotExists`、`DeleteIfExists`
- 状态查询：`Exists`、`CountByPostID`
- 历史记录：`GetByUserID`

## Error Types

| Error | Code | Message |
|-------|------|---------|
| `ErrPostNotFound` | `post_not_found` | 文章不存在 |
| `ErrPostAlreadyExists` | `post_already_exists` | 文章已存在（Slug 冲突） |
| `ErrInvalidStatus` | `invalid_status` | 无效的文章状态 |
| `ErrCategoryNotFound` | `category_not_found` | 分类不存在 |
| `ErrTagNotFound` | `tag_not_found` | 标签不存在 |
| `ErrSeriesNotFound` | `series_not_found` | 系列不存在 |
| `ErrVersionNotFound` | `version_not_found` | 版本不存在 |
| `ErrPermissionDenied` | `permission_denied` | 权限不足 |
| `ErrPaidContent` | `paid_content` | 此为付费内容，请先购买 |
| `ErrAlreadyLiked` | `already_liked` | 已点赞过该文章 |
| `ErrNotLiked` | `not_liked` | 未点赞过该文章 |

所有错误类型均为 `*PostError`，实现 `error` 接口和 `errors.Is` 接口，支持错误比较。

## PostListFilters 筛选条件

```go
type PostListFilters struct {
    Status     PostStatus // 按状态筛选
    AuthorID   uuid.UUID  // 按作者筛选
    CategoryID uuid.UUID  // 按分类筛选
    SeriesID   uuid.UUID  // 按系列筛选
    TagID      uuid.UUID  // 按标签筛选
    Search     string     // 搜索关键词
}
```

## For AI Agents

### 开发指南

#### 1. 添加新字段到 Post 实体

在 `models.go` 的 `Post` 结构体中添加字段：
```go
type Post struct {
    // ... existing fields ...
    
    // NewField 字段说明
    NewField Type `json:"new_field"`
}
```

同步更新：
- 在 `repository.go` 中检查是否需要新的查询方法
- 在 `models_test.go` 中添加相关测试（如验证逻辑）

#### 2. 添加新的 Repository 方法

在 `repository.go` 中定义接口方法，遵循命名约定：
```go
// GetByField 根据字段获取
GetByField(ctx context.Context, value Type) (*Post, error)

// CountByField 统计数量
CountByField(ctx context.Context, filters Filters) (int, error)
```

实现位于 `internal/database/post/` 目录。

#### 3. 添加新的错误类型

在 `models.go` 中添加：
```go
var (
    ErrNewError = &PostError{Code: "new_error", Message: "错误描述"}
)
```

#### 4. 跨模块引用规则

- 可引用 `user.User`（作者信息）
- 可引用 `comment.Comment`（评论关联）
- 避免循环依赖：`post` 不应被 `user` 直接引用

#### 5. 测试要求

- 所有枚举类型必须测试 `IsValid()` 方法
- 所有错误类型必须测试 `Error()` 和 `Is()` 方法
- 使用 `assert.Equal` 验证预期值

### 使用示例

#### 筛选文章列表
```go
filters := PostListFilters{
    Status:     StatusPublished,
    CategoryID: categoryUUID,
    Search:     "关键词",
}
posts, total, err := repo.List(ctx, filters, 0, 20)
```

#### 错误处理
```go
post, err := repo.GetByID(ctx, id)
if errors.Is(err, post.ErrPostNotFound) {
    // 处理文章不存在的情况
}
```

#### 点赞操作
```go
created, err := likeRepo.CreateIfNotExists(ctx, postID, userID)
if err != nil {
    return err
}
if !created {
    return post.ErrAlreadyLiked
}
```
