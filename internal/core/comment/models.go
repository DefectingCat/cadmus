// Package comment 评论管理模块
package comment

import (
	"time"

	"github.com/google/uuid"
)

// CommentStatus 评论状态枚举
type CommentStatus string

const (
	StatusPending  CommentStatus = "pending"  // 待审核
	StatusApproved CommentStatus = "approved" // 已批准
	StatusSpam     CommentStatus = "spam"     // 垃圾评论
	StatusDeleted  CommentStatus = "deleted"  // 已删除
)

// IsValid 检查评论状态是否有效
func (s CommentStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusSpam, StatusDeleted:
		return true
	default:
		return false
	}
}

// Comment 评论实体
type Comment struct {
	ID        uuid.UUID     `json:"id"`
	PostID    uuid.UUID     `json:"post_id"`
	UserID    uuid.UUID     `json:"user_id"`
	ParentID  *uuid.UUID    `json:"parent_id"`  // 父评论ID（可为空，表示顶层评论）
	Depth     int           `json:"depth"`      // 嵌套深度（0为顶层）
	Content   string        `json:"content"`    // 评论内容
	Status    CommentStatus `json:"status"`     // 评论状态
	LikeCount int           `json:"like_count"` // 点赞数
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// CommentLike 评论点赞记录实体
type CommentLike struct {
	ID        uuid.UUID `json:"id"`
	CommentID uuid.UUID `json:"comment_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateCommentInput 创建评论的输入结构
type CreateCommentInput struct {
	PostID   uuid.UUID  `json:"post_id"`
	UserID   uuid.UUID  `json:"user_id"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	Content  string     `json:"content"`
}

// CommentListFilters 评论列表筛选条件
type CommentListFilters struct {
	PostID   uuid.UUID
	UserID   uuid.UUID
	Status   CommentStatus
	ParentID *uuid.UUID // 筛选某评论的子评论
	Depth    int        // 筛选特定深度
}

// 常见错误定义
var (
	ErrCommentNotFound     = &CommentError{Code: "comment_not_found", Message: "评论不存在"}
	ErrCommentAlreadyExists = &CommentError{Code: "comment_already_exists", Message: "评论已存在"}
	ErrInvalidStatus       = &CommentError{Code: "invalid_status", Message: "无效的评论状态"}
	ErrMaxDepthExceeded    = &CommentError{Code: "max_depth_exceeded", Message: "评论嵌套深度超过限制"}
	ErrParentNotFound      = &CommentError{Code: "parent_not_found", Message: "父评论不存在"}
	ErrPermissionDenied    = &CommentError{Code: "permission_denied", Message: "权限不足"}
	ErrEmptyContent        = &CommentError{Code: "empty_content", Message: "评论内容不能为空"}
	ErrPostNotFound        = &CommentError{Code: "post_not_found", Message: "文章不存在"}
	ErrUserNotFound        = &CommentError{Code: "user_not_found", Message: "用户不存在"}
	ErrAlreadyLiked        = &CommentError{Code: "already_liked", Message: "已点赞该评论"}
	ErrNotLiked            = &CommentError{Code: "not_liked", Message: "未点赞该评论"}
)

// CommentError 评论模块自定义错误
type CommentError struct {
	Code    string
	Message string
}

func (e *CommentError) Error() string {
	return e.Message
}

// Is 实现 errors.Is 接口
func (e *CommentError) Is(target error) bool {
	t, ok := target.(*CommentError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}