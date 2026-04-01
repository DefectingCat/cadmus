// Package comment 提供评论管理功能的核心数据模型。
//
// 该文件包含评论系统的核心实体定义，包括：
//   - 评论实体及其状态枚举
//   - 评论点赞记录实体
//   - 创建评论的输入结构
//   - 评论列表筛选条件
//   - 语义化错误类型定义
//
// 主要用途：
//
//	用于博客系统的评论功能，支持嵌套回复、点赞、状态管理等特性。
//
// 注意事项：
//   - CommentStatus 枚举通过 IsValid() 方法验证状态有效性
//   - 嵌套评论通过 ParentID 和 Depth 字段实现，最大深度由业务层控制
//   - 所有实体使用 UUID 作为主键
//
// 作者：xfy
package comment

import (
	"time"

	"github.com/google/uuid"
)

// CommentStatus 评论状态枚举。
//
// 定义评论在系统中的生命周期状态，控制评论的可见性和管理行为。
type CommentStatus string

// 评论状态常量定义。
const (
	// StatusPending 待审核状态，新评论默认状态，需管理员审核后公开
	StatusPending CommentStatus = "pending"

	// StatusApproved 已批准状态，评论公开可见
	StatusApproved CommentStatus = "approved"

	// StatusSpam 垃圾评论状态，被判定为垃圾内容的评论
	StatusSpam CommentStatus = "spam"

	// StatusDeleted 已删除状态，软删除后的评论状态
	StatusDeleted CommentStatus = "deleted"
)

// IsValid 检查评论状态是否有效。
//
// 验证状态值是否为预定义的四种状态之一。
//
// 返回值：
//   - true: 状态有效
//   - false: 状态无效
func (s CommentStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusSpam, StatusDeleted:
		return true
	default:
		return false
	}
}

// Comment 评论实体。
//
// 表示博客文章中的一条评论，支持嵌套回复结构。
// 通过 ParentID 和 Depth 字段实现评论树形结构。
//
// 注意事项：
//   - ID 由系统自动生成，无需手动设置
//   - ParentID 为空表示顶层评论（Depth=0）
//   - ParentID 有值表示回复评论（Depth>=1）
//   - CreatedAt/UpdatedAt 使用 UTC 时间
type Comment struct {
	// ID 评论的唯一标识符，格式为 UUID
	ID uuid.UUID `json:"id"`

	// PostID 所属文章的 ID
	PostID uuid.UUID `json:"post_id"`

	// UserID 评论者的用户 ID
	UserID uuid.UUID `json:"user_id"`

	// ParentID 父评论 ID，为空表示顶层评论
	ParentID *uuid.UUID `json:"parent_id"`

	// Depth 嵌套深度，0 表示顶层评论，>=1 表示回复层级
	Depth int `json:"depth"`

	// Content 评论内容，支持纯文本或 Markdown 格式
	Content string `json:"content"`

	// Status 评论状态，控制可见性
	Status CommentStatus `json:"status"`

	// LikeCount 点赞数统计
	LikeCount int `json:"like_count"`

	// CreatedAt 创建时间，使用 UTC 时间戳
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 最后修改时间，使用 UTC 时间戳
	UpdatedAt time.Time `json:"updated_at"`
}

// CommentLike 评论点赞记录实体。
//
// 记录用户对评论的点赞行为，用于统计和查询用户点赞历史。
// 通过唯一约束防止重复点赞同一评论。
type CommentLike struct {
	// ID 点赞记录的唯一标识符
	ID uuid.UUID `json:"id"`

	// CommentID 被点赞评论的 ID
	CommentID uuid.UUID `json:"comment_id"`

	// UserID 点赞用户的 ID
	UserID uuid.UUID `json:"user_id"`

	// CreatedAt 点赞时间
	CreatedAt time.Time `json:"created_at"`
}

// CreateCommentInput 创建评论的输入结构。
//
// 用于接收创建评论请求的参数，包含必要的信息。
type CreateCommentInput struct {
	// PostID 所属文章的 ID，必填
	PostID uuid.UUID `json:"post_id"`

	// UserID 评论者的用户 ID，必填
	UserID uuid.UUID `json:"user_id"`

	// ParentID 父评论 ID，可选，为空表示顶层评论
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// Content 评论内容，必填，不能为空
	Content string `json:"content"`
}

// CommentListFilters 评论列表筛选条件。
//
// 用于构建评论查询的过滤条件，支持多条件组合筛选。
// 所有字段均为可选，未设置时忽略该条件。
type CommentListFilters struct {
	// PostID 按文章筛选，获取指定文章的评论
	PostID uuid.UUID

	// UserID 按用户筛选，获取指定用户的评论
	UserID uuid.UUID

	// Status 按状态筛选，默认获取已批准的评论
	Status CommentStatus

	// ParentID 筛选指定评论的子评论，用于递归加载
	ParentID *uuid.UUID

	// Depth 筛选特定深度的评论，如只获取顶层评论（Depth=0）
	Depth int
}

// 常见错误定义。
//
// 使用语义化错误类型，便于调用方进行错误处理和判断。
var (
	// ErrCommentNotFound 评论不存在错误，查询评论时 ID 无效时返回
	ErrCommentNotFound = &CommentError{Code: "comment_not_found", Message: "评论不存在"}

	// ErrCommentAlreadyExists 评论已存在错误，重复创建时返回
	ErrCommentAlreadyExists = &CommentError{Code: "comment_already_exists", Message: "评论已存在"}

	// ErrInvalidStatus 无效评论状态错误，设置状态值不合法时返回
	ErrInvalidStatus = &CommentError{Code: "invalid_status", Message: "无效的评论状态"}

	// ErrMaxDepthExceeded 嵌套深度超限错误，回复层级超过系统限制时返回
	ErrMaxDepthExceeded = &CommentError{Code: "max_depth_exceeded", Message: "评论嵌套深度超过限制"}

	// ErrParentNotFound 父评论不存在错误，回复的评论 ID 无效时返回
	ErrParentNotFound = &CommentError{Code: "parent_not_found", Message: "父评论不存在"}

	// ErrPermissionDenied 权限不足错误，用户无权操作评论时返回
	ErrPermissionDenied = &CommentError{Code: "permission_denied", Message: "权限不足"}

	// ErrEmptyContent 评论内容为空错误，提交空评论时返回
	ErrEmptyContent = &CommentError{Code: "empty_content", Message: "评论内容不能为空"}

	// ErrPostNotFound 文章不存在错误，评论的文章 ID 无效时返回
	ErrPostNotFound = &CommentError{Code: "post_not_found", Message: "文章不存在"}

	// ErrUserNotFound 用户不存在错误，评论的用户 ID 无效时返回
	ErrUserNotFound = &CommentError{Code: "user_not_found", Message: "用户不存在"}

	// ErrAlreadyLiked 已点赞错误，用户重复点赞同一评论时返回
	ErrAlreadyLiked = &CommentError{Code: "already_liked", Message: "已点赞该评论"}

	// ErrNotLiked 未点赞错误，取消点赞但未点赞过时返回
	ErrNotLiked = &CommentError{Code: "not_liked", Message: "未点赞该评论"}
)

// CommentError 评论模块自定义错误类型。
//
// 实现 error 和 errors.Is 接口，支持错误比较和类型判断。
// 通过 Code 字段区分不同错误类型，Message 字段提供人类可读描述。
type CommentError struct {
	// Code 错误代码，用于程序化错误判断
	Code string

	// Message 错误消息，用于展示给用户或记录日志
	Message string
}

// Error 实现 error 接口，返回错误消息。
func (e *CommentError) Error() string {
	return e.Message
}

// Is 实现 errors.Is 接口，支持错误类型比较。
//
// 通过比较 Code 字段判断是否为同类型错误，
// 便于使用 errors.Is(err, ErrCommentNotFound) 进行错误判断。
func (e *CommentError) Is(target error) bool {
	t, ok := target.(*CommentError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}