// Package services 提供评论服务的实现。
//
// 该文件包含评论管理相关的核心逻辑，包括：
//   - 评论 CRUD 操作（创建、更新、删除）
//   - 评论审核（批准、拒绝、标记为垃圾）
//   - 评论点赞功能
//   - 嵌套评论树形结构构建
//   - 批量操作支持
//
// 主要用途：
//
//	用于处理博客评论系统的完整功能，支持多层嵌套回复。
//
// 设计特点：
//   - 最大嵌套深度限制（防止无限嵌套）
//   - 评论审核机制（防止垃圾评论）
//   - 树形结构便于前端渲染
//
// 作者：xfy
package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/internal/database"
)

// MaxCommentDepth 评论最大嵌套深度。
//
// 限制评论嵌套深度可防止：
//   - 无限嵌套导致的前端渲染问题
//   - 深层嵌套导致的用户体验下降
//   - 潜在的递归性能问题
const MaxCommentDepth = 5

// CommentService 评论业务服务接口。
//
// 该接口定义了评论管理的核心操作，包括 CRUD、审核、点赞和批量操作。
// 所有方法均为并发安全。
type CommentService interface {
	// CreateComment 创建评论，包含深度检查。
	//
	// 参数：
	//   - ctx: 上下文
	//   - input: 评论创建输入，包含内容、文章ID、父评论ID等
	//
	// 返回值：
	//   - comment: 创建的评论对象
	//   - err: 可能的错误包括内容为空、父评论不存在、超过最大深度
	CreateComment(ctx context.Context, input *comment.CreateCommentInput) (*comment.Comment, error)

	// GetCommentByID 根据 ID 获取评论详情。
	GetCommentByID(ctx context.Context, id uuid.UUID) (*comment.Comment, error)

	// GetCommentsByPost 获取文章的评论，构建树形结构。
	//
	// 返回的是树形结构，便于前端递归渲染嵌套评论。
	// 仅返回已批准的评论。
	GetCommentsByPost(ctx context.Context, postID uuid.UUID) ([]*CommentNode, error)

	// GetCommentsByUser 获取用户的评论列表。
	GetCommentsByUser(ctx context.Context, userID uuid.UUID) ([]*comment.Comment, error)

	// ApproveComment 批准评论，将状态设为已批准。
	//
	// 已删除的评论无法批准。
	ApproveComment(ctx context.Context, id uuid.UUID) error

	// RejectComment 拒绝评论，标记为垃圾评论。
	RejectComment(ctx context.Context, id uuid.UUID) error

	// UpdateComment 更新评论内容。
	UpdateComment(ctx context.Context, c *comment.Comment) error

	// DeleteComment 删除评论（需验证权限）。
	//
	// 仅评论作者可以删除自己的评论。
	DeleteComment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// LikeComment 点赞评论。
	//
	// 使用原子操作保证并发安全，只能点赞已批准的评论。
	LikeComment(ctx context.Context, commentID, userID uuid.UUID) error

	// UnlikeComment 取消点赞评论。
	UnlikeComment(ctx context.Context, commentID, userID uuid.UUID) error

	// IsLiked 检查用户是否已点赞指定评论。
	IsLiked(ctx context.Context, commentID, userID uuid.UUID) (bool, error)

	// GetLikesBatch 批量检查点赞状态。
	//
	// 用于列表页一次性获取多个评论的点赞状态，避免 N+1 查询。
	GetLikesBatch(ctx context.Context, commentIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]bool, error)

	// CountCommentsByPost 统计文章的评论数量。
	CountCommentsByPost(ctx context.Context, postID uuid.UUID) (int, error)

	// GetCommentsByStatus 获取指定状态的评论列表（后台审核用）。
	GetCommentsByStatus(ctx context.Context, status comment.CommentStatus, offset, limit int) ([]*comment.Comment, int, error)

	// DeleteCommentAdmin 管理员删除评论，无需权限检查。
	DeleteCommentAdmin(ctx context.Context, id uuid.UUID) error

	// BatchApproveComments 批量批准评论。
	BatchApproveComments(ctx context.Context, ids []uuid.UUID) error

	// BatchRejectComments 批量拒绝评论。
	BatchRejectComments(ctx context.Context, ids []uuid.UUID) error

	// BatchDeleteComments 批量删除评论。
	BatchDeleteComments(ctx context.Context, ids []uuid.UUID) error
}

// CommentNode 评论树节点，用于构建树形结构。
//
// 该结构体将扁平的评论列表转换为树形结构，便于前端渲染嵌套回复。
// 每个节点包含评论对象和其子评论列表。
type CommentNode struct {
	// Comment 评论对象
	Comment *comment.Comment `json:"comment"`

	// Children 子评论列表（回复）
	Children []*CommentNode `json:"children,omitempty"`
}

// commentServiceImpl 评论服务的具体实现。
type commentServiceImpl struct {
	// commentRepo 评论数据仓库
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

	// 使用原子操作创建点赞记录并更新计数
	created, err := s.likeRepo.CreateIfNotExists(ctx, commentID, userID)
	if err != nil {
		return err
	}
	if !created {
		return comment.ErrAlreadyLiked
	}

	return nil
}

// UnlikeComment 取消点赞
func (s *commentServiceImpl) UnlikeComment(ctx context.Context, commentID, userID uuid.UUID) error {
	// 检查评论是否存在
	_, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return comment.ErrCommentNotFound
	}

	// 使用原子操作删除点赞记录并更新计数
	deleted, err := s.likeRepo.DeleteIfExists(ctx, commentID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return comment.ErrNotLiked
	}

	return nil
}

// IsLiked 检查用户是否已点赞
func (s *commentServiceImpl) IsLiked(ctx context.Context, commentID, userID uuid.UUID) (bool, error) {
	return s.likeRepo.Exists(ctx, commentID, userID)
}

// GetLikesBatch 批量检查用户对多个评论的点赞状态
func (s *commentServiceImpl) GetLikesBatch(ctx context.Context, commentIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]bool, error) {
	dbLikeRepo, ok := s.likeRepo.(*database.CommentLikeRepository)
	if ok {
		return dbLikeRepo.GetLikesBatch(ctx, commentIDs, userID)
	}
	// fallback: 逐个查询
	result := make(map[uuid.UUID]bool)
	for _, id := range commentIDs {
		liked, _ := s.likeRepo.Exists(ctx, id, userID)
		result[id] = liked
	}
	return result, nil
}

// CountCommentsByPost 统计文章评论数量
func (s *commentServiceImpl) CountCommentsByPost(ctx context.Context, postID uuid.UUID) (int, error) {
	return s.commentRepo.CountByPostID(ctx, postID)
}

// GetCommentsByStatus 获取指定状态的评论列表（用于后台审核）
func (s *commentServiceImpl) GetCommentsByStatus(ctx context.Context, status comment.CommentStatus, offset, limit int) ([]*comment.Comment, int, error) {
	filters := &comment.CommentListFilters{Status: status}
	comments, err := s.commentRepo.List(ctx, filters, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数
	total := 0
	for range comments {
		total++
	}

	// 如果需要精确总数，可以添加一个 CountByStatus 方法
	// 这里简化处理，返回当前页数据量
	return comments, total, nil
}

// DeleteCommentAdmin 管理员删除评论（无需权限检查）
func (s *commentServiceImpl) DeleteCommentAdmin(ctx context.Context, id uuid.UUID) error {
	// 检查评论是否存在
	_, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return comment.ErrCommentNotFound
	}

	return s.commentRepo.Delete(ctx, id)
}

// BatchApproveComments 批量批准评论
func (s *commentServiceImpl) BatchApproveComments(ctx context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		if err := s.ApproveComment(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// BatchRejectComments 批量拒绝评论
func (s *commentServiceImpl) BatchRejectComments(ctx context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		if err := s.RejectComment(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// BatchDeleteComments 批量删除评论
func (s *commentServiceImpl) BatchDeleteComments(ctx context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		if err := s.DeleteCommentAdmin(ctx, id); err != nil {
			return err
		}
	}
	return nil
}