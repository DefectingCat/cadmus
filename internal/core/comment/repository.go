// Package comment 评论管理模块
package comment

import (
	"context"

	"github.com/google/uuid"
)

// CommentRepository 评论仓库接口
type CommentRepository interface {
	// Create 创建新评论
	Create(ctx context.Context, input *CreateCommentInput) (*Comment, error)

	// GetByID 根据ID获取评论
	GetByID(ctx context.Context, id uuid.UUID) (*Comment, error)

	// GetByPostID 获取文章的所有评论（支持筛选）
	GetByPostID(ctx context.Context, postID uuid.UUID, filters *CommentListFilters) ([]*Comment, error)

	// GetByUserID 获取用户的所有评论
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Comment, error)

	// GetChildren 获取评论的子评论
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*Comment, error)

	// Update 更新评论
	Update(ctx context.Context, comment *Comment) error

	// UpdateStatus 更新评论状态
	UpdateStatus(ctx context.Context, id uuid.UUID, status CommentStatus) error

	// Delete 删除评论（软删除，标记为 deleted）
	Delete(ctx context.Context, id uuid.UUID) error

	// CountByPostID 统计文章的评论数量
	CountByPostID(ctx context.Context, postID uuid.UUID) (int, error)

	// CountByUserID 统计用户的评论数量
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// List 分页获取评论列表
	List(ctx context.Context, filters *CommentListFilters, offset, limit int) ([]*Comment, error)
}

// CommentLikeRepository 评论点赞仓库接口
type CommentLikeRepository interface {
	// Create 创建点赞记录
	Create(ctx context.Context, commentID, userID uuid.UUID) (*CommentLike, error)

	// GetByCommentAndUser 获取用户对评论的点赞记录
	GetByCommentAndUser(ctx context.Context, commentID, userID uuid.UUID) (*CommentLike, error)

	// Delete 删除点赞记录（取消点赞）
	Delete(ctx context.Context, commentID, userID uuid.UUID) error

	// Exists 检查用户是否已点赞评论
	Exists(ctx context.Context, commentID, userID uuid.UUID) (bool, error)

	// CountByCommentID 统计评论的点赞数量
	CountByCommentID(ctx context.Context, commentID uuid.UUID) (int, error)

	// GetByUserID 获取用户的所有点赞记录
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*CommentLike, error)
}