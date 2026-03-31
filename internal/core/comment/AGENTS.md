<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# comment

## Purpose
评论领域模型，定义评论实体和 Repository 接口。

## Key Files
| File | Description |
|------|-------------|
| `models.go` | Comment、CommentLike 结构定义 |
| `repository.go` | CommentRepository、CommentLikeRepository 接口定义 |

## For AI Agents

### Working In This Directory
- 支持嵌套回复，最大深度 5 层
- 审核流程：pending → approved/rejected

### Data Models
```go
type Comment struct {
    ID        uuid.UUID
    PostID    uuid.UUID
    UserID    uuid.UUID
    ParentID  *uuid.UUID     // 嵌套回复
    Depth     int            // 嵌套深度 0-5
    Content   string
    Status    CommentStatus  // pending/approved/spam/deleted
    LikeCount int
    CreatedAt time.Time
    UpdatedAt time.Time
}

type CommentLike struct {
    ID        uuid.UUID
    CommentID uuid.UUID
    UserID    uuid.UUID
    CreatedAt time.Time
}
```

### Comment Status Flow
```
用户发表 → pending (待审核)
             ↓
        approved (已通过) / spam (垃圾)
             ↓
           deleted (已删除)
```

### Depth Limitation
- 最大嵌套深度：5 层
- 数据库 CHECK 约束：`depth >= 0 AND depth <= 5`
- 前端渲染时检查深度

### Repository Interfaces
```go
type CommentRepository interface {
    Create(ctx context.Context, comment *Comment) error
    GetByID(ctx context.Context, id uuid.UUID) (*Comment, error)
    GetByPost(ctx context.Context, postID uuid.UUID) ([]*Comment, error)
    GetByStatus(ctx context.Context, status CommentStatus, offset, limit int) ([]*Comment, int, error)
    Update(ctx context.Context, comment *Comment) error
    Delete(ctx context.Context, id uuid.UUID) error
    Approve(ctx context.Context, id uuid.UUID) error
    Reject(ctx context.Context, id uuid.UUID) error
    BatchApprove(ctx context.Context, ids []uuid.UUID) error
    BatchReject(ctx context.Context, ids []uuid.UUID) error
}

type CommentLikeRepository interface {
    Create(ctx context.Context, like *CommentLike) error
    Delete(ctx context.Context, commentID, userID uuid.UUID) error
    GetByUserAndComment(ctx context.Context, commentID, userID uuid.UUID) (*CommentLike, error)
}
```

<!-- MANUAL: -->