package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/comment"
)

// MaxCommentDepth 评论最大嵌套深度
const MaxCommentDepth = 5

// CommentService 评论业务服务接口
type CommentService interface {
	// CreateComment 创建评论（包含深度检查）
	CreateComment(ctx context.Context, input *comment.CreateCommentInput) (*comment.Comment, error)

	// GetCommentByID 根据 ID 获取评论
	GetCommentByID(ctx context.Context, id uuid.UUID) (*comment.Comment, error)

	// GetCommentsByPost 获取文章的评论（构建树形结构）
	GetCommentsByPost(ctx context.Context, postID uuid.UUID) ([]*CommentNode, error)

	// GetCommentsByUser 获取用户的评论
	GetCommentsByUser(ctx context.Context, userID uuid.UUID) ([]*comment.Comment, error)

	// ApproveComment 批准评论
	ApproveComment(ctx context.Context, id uuid.UUID) error

	// RejectComment 拒绝评论（标记为垃圾）
	RejectComment(ctx context.Context, id uuid.UUID) error

	// UpdateComment 更新评论内容
	UpdateComment(ctx context.Context, c *comment.Comment) error

	// DeleteComment 删除评论
	DeleteComment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// LikeComment 点赞评论
	LikeComment(ctx context.Context, commentID, userID uuid.UUID) error

	// UnlikeComment 取消点赞
	UnlikeComment(ctx context.Context, commentID, userID uuid.UUID) error

	// IsLiked 检查用户是否已点赞
	IsLiked(ctx context.Context, commentID, userID uuid.UUID) (bool, error)

	// CountCommentsByPost 统计文章评论数量
	CountCommentsByPost(ctx context.Context, postID uuid.UUID) (int, error)
}

// CommentNode 评论树节点（用于构建树形结构）
type CommentNode struct {
	Comment  *comment.Comment `json:"comment"`
	Children []*CommentNode   `json:"children,omitempty"`
}

// commentServiceImpl 评论服务实现
type commentServiceImpl struct {
	commentRepo comment.CommentRepository
	likeRepo    comment.CommentLikeRepository
}

// NewCommentService 创建评论服务
func NewCommentService(
	commentRepo comment.CommentRepository,
	likeRepo comment.CommentLikeRepository,
) CommentService {
	return &commentServiceImpl{
		commentRepo: commentRepo,
		likeRepo:    likeRepo,
	}
}

// CreateComment 创建评论
func (s *commentServiceImpl) CreateComment(ctx context.Context, input *comment.CreateCommentInput) (*comment.Comment, error) {
	// 验证内容不为空
	if input.Content == "" {
		return nil, comment.ErrEmptyContent
	}

	// 检查嵌套深度
	if input.ParentID != nil {
		parent, err := s.commentRepo.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, comment.ErrParentNotFound
		}

		// 检查深度是否超过限制
		if parent.Depth >= MaxCommentDepth {
			return nil, comment.ErrMaxDepthExceeded
		}

		// 新评论深度 = 父评论深度 + 1
		input.ParentID = &parent.ID
	}

	// 创建评论
	c, err := s.commentRepo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// GetCommentByID 根据 ID 获取评论
func (s *commentServiceImpl) GetCommentByID(ctx context.Context, id uuid.UUID) (*comment.Comment, error) {
	return s.commentRepo.GetByID(ctx, id)
}

// GetCommentsByPost 获取文章的评论（构建树形结构）
func (s *commentServiceImpl) GetCommentsByPost(ctx context.Context, postID uuid.UUID) ([]*CommentNode, error) {
	// 获取所有已批准的评论
	filters := &comment.CommentListFilters{
		PostID: postID,
		Status: comment.StatusApproved,
	}
	comments, err := s.commentRepo.GetByPostID(ctx, postID, filters)
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	return s.buildCommentTree(comments), nil
}

// buildCommentTree 构建评论树形结构
func (s *commentServiceImpl) buildCommentTree(comments []*comment.Comment) []*CommentNode {
	// 创建节点映射
	nodeMap := make(map[uuid.UUID]*CommentNode)
	for _, c := range comments {
		nodeMap[c.ID] = &CommentNode{
			Comment:  c,
			Children: []*CommentNode{},
		}
	}

	// 构建树
	rootNodes := []*CommentNode{}
	for _, c := range comments {
		node := nodeMap[c.ID]
		if c.ParentID == nil {
			// 顶层评论
			rootNodes = append(rootNodes, node)
		} else {
			// 子评论，添加到父评论的 children
			if parent, ok := nodeMap[*c.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return rootNodes
}

// GetCommentsByUser 获取用户的评论
func (s *commentServiceImpl) GetCommentsByUser(ctx context.Context, userID uuid.UUID) ([]*comment.Comment, error) {
	return s.commentRepo.GetByUserID(ctx, userID)
}

// ApproveComment 批准评论
func (s *commentServiceImpl) ApproveComment(ctx context.Context, id uuid.UUID) error {
	// 检查评论是否存在
	c, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return comment.ErrCommentNotFound
	}

	// 检查状态是否有效
	if c.Status == comment.StatusDeleted {
		return errors.New("无法批准已删除的评论")
	}

	return s.commentRepo.UpdateStatus(ctx, id, comment.StatusApproved)
}

// RejectComment 拒绝评论（标记为垃圾）
func (s *commentServiceImpl) RejectComment(ctx context.Context, id uuid.UUID) error {
	// 检查评论是否存在
	c, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return comment.ErrCommentNotFound
	}

	// 检查状态是否有效
	if c.Status == comment.StatusDeleted {
		return errors.New("无法拒绝已删除的评论")
	}

	return s.commentRepo.UpdateStatus(ctx, id, comment.StatusSpam)
}

// UpdateComment 更新评论内容
func (s *commentServiceImpl) UpdateComment(ctx context.Context, c *comment.Comment) error {
	return s.commentRepo.Update(ctx, c)
}

// DeleteComment 删除评论
func (s *commentServiceImpl) DeleteComment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// 检查评论是否存在
	c, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return comment.ErrCommentNotFound
	}

	// 检查权限：只有评论作者可以删除自己的评论
	if c.UserID != userID {
		return comment.ErrPermissionDenied
	}

	return s.commentRepo.Delete(ctx, id)
}

// LikeComment 点赞评论
func (s *commentServiceImpl) LikeComment(ctx context.Context, commentID, userID uuid.UUID) error {
	// 检查评论是否存在且已批准
	c, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return comment.ErrCommentNotFound
	}

	if c.Status != comment.StatusApproved {
		return errors.New("只能点赞已批准的评论")
	}

	// 检查是否已点赞
	exists, err := s.likeRepo.Exists(ctx, commentID, userID)
	if err != nil {
		return err
	}
	if exists {
		return comment.ErrAlreadyLiked
	}

	// 创建点赞记录
	_, err = s.likeRepo.Create(ctx, commentID, userID)
	if err != nil {
		return err
	}

	// 更新点赞计数
	c.LikeCount++
	c.UpdatedAt = time.Now()
	return s.commentRepo.Update(ctx, c)
}

// UnlikeComment 取消点赞
func (s *commentServiceImpl) UnlikeComment(ctx context.Context, commentID, userID uuid.UUID) error {
	// 检查评论是否存在
	c, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return comment.ErrCommentNotFound
	}

	// 检查是否已点赞
	exists, err := s.likeRepo.Exists(ctx, commentID, userID)
	if err != nil {
		return err
	}
	if !exists {
		return comment.ErrNotLiked
	}

	// 删除点赞记录
	err = s.likeRepo.Delete(ctx, commentID, userID)
	if err != nil {
		return err
	}

	// 更新点赞计数
	c.LikeCount--
	if c.LikeCount < 0 {
		c.LikeCount = 0
	}
	c.UpdatedAt = time.Now()
	return s.commentRepo.Update(ctx, c)
}

// IsLiked 检查用户是否已点赞
func (s *commentServiceImpl) IsLiked(ctx context.Context, commentID, userID uuid.UUID) (bool, error) {
	return s.likeRepo.Exists(ctx, commentID, userID)
}

// CountCommentsByPost 统计文章评论数量
func (s *commentServiceImpl) CountCommentsByPost(ctx context.Context, postID uuid.UUID) (int, error) {
	return s.commentRepo.CountByPostID(ctx, postID)
}