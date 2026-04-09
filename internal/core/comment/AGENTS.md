<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# comment - 评论领域模型

## Purpose

`comment` 目录定义博客评论系统的领域模型和数据访问接口。该模块支持：

- **嵌套回复**: 通过 `ParentID` 和 `Depth` 字段实现评论树形结构
- **状态管理**: 审核流程（待审核/已批准/垃圾评论/已删除）
- **点赞功能**: 用户评论点赞记录与统计
- **语义化错误**: 自定义错误类型便于调用方处理

该目录**仅包含接口和类型定义**，不包含具体实现。实现位于 `internal/database/`（数据存储）和 `internal/service/`（业务逻辑）。

## Key Files

| File | Purpose |
|------|---------|
| `models.go` | 领域实体、枚举、输入结构、错误类型定义 |
| `repository.go` | 数据访问接口定义（CommentRepository、CommentLikeRepository） |
| `models_test.go` | 单元测试：状态验证、错误类型测试 |

## Domain Models

### 核心实体

#### Comment (评论)
```go
type Comment struct {
    ID         uuid.UUID       // 评论唯一标识
    PostID     uuid.UUID       // 所属文章 ID
    UserID     uuid.UUID       // 评论者用户 ID
    ParentID   *uuid.UUID      // 父评论 ID（nil 表示顶层评论）
    Depth      int             // 嵌套深度（0=顶层，>=1=回复）
    Content    string          // 评论内容
    Status     CommentStatus   // 评论状态
    LikeCount  int             // 点赞数统计
    CreatedAt  time.Time       // 创建时间（UTC）
    UpdatedAt  time.Time       // 最后修改时间（UTC）
}
```

#### CommentLike (点赞记录)
```go
type CommentLike struct {
    ID        uuid.UUID  // 点赞记录唯一标识
    CommentID uuid.UUID  // 被点赞评论 ID
    UserID    uuid.UUID  // 点赞用户 ID
    CreatedAt time.Time  // 点赞时间
}
```

### 状态枚举

#### CommentStatus
```go
type CommentStatus string

const (
    StatusPending  CommentStatus = "pending"   // 待审核
    StatusApproved CommentStatus = "approved"  // 已批准（公开可见）
    StatusSpam     CommentStatus = "spam"      // 垃圾评论
    StatusDeleted  CommentStatus = "deleted"   // 已删除（软删除）
)
```

验证方法：`status.IsValid()` 返回状态是否有效。

### 输入结构

#### CreateCommentInput
```go
type CreateCommentInput struct {
    PostID   uuid.UUID       // 文章 ID（必填）
    UserID   uuid.UUID       // 用户 ID（必填）
    ParentID *uuid.UUID      // 父评论 ID（可选，nil 表示顶层评论）
    Content  string          // 评论内容（必填，不可为空）
}
```

#### CommentListFilters
```go
type CommentListFilters struct {
    PostID   uuid.UUID       // 按文章筛选
    UserID   uuid.UUID       // 按用户筛选
    Status   CommentStatus   // 按状态筛选
    ParentID *uuid.UUID      // 筛选子评论（递归加载）
    Depth    int             // 筛选特定深度
}
```

### 错误类型

#### 预定义错误
| 错误 | 代码 | 说明 |
|------|------|------|
| `ErrCommentNotFound` | `comment_not_found` | 评论不存在 |
| `ErrCommentAlreadyExists` | `comment_already_exists` | 评论已存在 |
| `ErrInvalidStatus` | `invalid_status` | 无效评论状态 |
| `ErrMaxDepthExceeded` | `max_depth_exceeded` | 嵌套深度超限 |
| `ErrParentNotFound` | `parent_not_found` | 父评论不存在 |
| `ErrPermissionDenied` | `permission_denied` | 权限不足 |
| `ErrEmptyContent` | `empty_content` | 评论内容为空 |
| `ErrPostNotFound` | `post_not_found` | 文章不存在 |
| `ErrUserNotFound` | `user_not_found` | 用户不存在 |
| `ErrAlreadyLiked` | `already_liked` | 已点赞 |
| `ErrNotLiked` | `not_liked` | 未点赞 |

#### CommentError 类型
```go
type CommentError struct {
    Code    string  // 错误代码，用于程序化判断
    Message string  // 错误消息，用于展示
}
```

实现 `error` 和 `errors.Is` 接口，支持 `errors.Is(err, ErrCommentNotFound)` 类型判断。

## Repository Interfaces

### CommentRepository

评论数据访问接口，支持嵌套结构和状态管理。

```go
type CommentRepository interface {
    // 创建
    Create(ctx context.Context, input *CreateCommentInput) (*Comment, error)
    
    // 查询
    GetByID(ctx context.Context, id uuid.UUID) (*Comment, error)
    GetByPostID(ctx context.Context, postID uuid.UUID, filters *CommentListFilters) ([]*Comment, error)
    GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Comment, error)
    GetChildren(ctx context.Context, parentID uuid.UUID) ([]*Comment, error)
    
    // 更新
    Update(ctx context.Context, comment *Comment) error
    UpdateStatus(ctx context.Context, id uuid.UUID, status CommentStatus) error
    
    // 删除
    Delete(ctx context.Context, id uuid.UUID) error
    
    // 统计
    CountByPostID(ctx context.Context, postID uuid.UUID) (int, error)
    CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
    
    // 分页列表
    List(ctx context.Context, filters *CommentListFilters, offset, limit int) ([]*Comment, error)
}
```

### CommentLikeRepository

点赞记录数据访问接口。

```go
type CommentLikeRepository interface {
    // 创建（返回是否实际创建）
    Create(ctx context.Context, commentID, userID uuid.UUID) (*CommentLike, error)
    CreateIfNotExists(ctx context.Context, commentID, userID uuid.UUID) (created bool, err error)
    
    // 查询
    GetByCommentAndUser(ctx context.Context, commentID, userID uuid.UUID) (*CommentLike, error)
    Exists(ctx context.Context, commentID, userID uuid.UUID) (bool, error)
    CountByCommentID(ctx context.Context, commentID uuid.UUID) (int, error)
    GetByUserID(ctx context.Context, userID uuid.UUID) ([]*CommentLike, error)
    
    // 删除（返回是否实际删除）
    Delete(ctx context.Context, commentID, userID uuid.UUID) error
    DeleteIfExists(ctx context.Context, commentID, userID uuid.UUID) (deleted bool, err error)
}
```

## Subdirectories

无子目录。

## For AI Agents

### 开发指南

1. **创建新实体时**:
   - 在 `models.go` 中添加结构体定义
   - 使用 `uuid.UUID` 作为主键类型
   - 添加 `CreatedAt`/`UpdatedAt` 时间戳字段
   - 在 `repository.go` 中添加对应的 Repository 接口

2. **实现业务逻辑时**:
   - Repository 接口只定义数据访问方法，不包含业务逻辑
   - 业务逻辑应放在 `internal/service/` 层实现
   - 所有错误必须使用本目录定义的语义化错误类型

3. **处理嵌套评论时**:
   - 顶层评论：`ParentID = nil`, `Depth = 0`
   - 回复评论：`ParentID != nil`, `Depth >= 1`
   - 注意检查 `ErrMaxDepthExceeded` 错误

4. **处理点赞时**:
   - 优先使用 `CreateIfNotExists` 和 `DeleteIfExists` 避免竞态条件
   - 检查 `ErrAlreadyLiked` 和 `ErrNotLiked` 错误

### 跨模块依赖

- `Comment.PostID` 依赖 `post` 模块的文章存在
- `Comment.UserID` 依赖 `user` 模块的用户存在
- 创建评论前需验证文章和用户有效性（返回 `ErrPostNotFound`/`ErrUserNotFound`）

### Mock 实现

单元测试时实现 `CommentRepository` 和 `CommentLikeRepository` 接口：
- 使用内存存储模拟数据操作
- 返回预定义的错误类型
- 参考 `internal/mocks/comment_repository.go`（如已存在）
