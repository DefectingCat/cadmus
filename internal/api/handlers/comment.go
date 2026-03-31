package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/internal/services"
)

// CommentHandler 评论 API 处理器
type CommentHandler struct {
	commentService     services.CommentService
	notificationService services.NotificationService
	postService        services.PostService
	userService        services.UserService
}

// NewCommentHandler 创建评论处理器
func NewCommentHandler(commentService services.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// NewCommentHandlerWithNotifications 创建带通知功能的评论处理器
func NewCommentHandlerWithNotifications(
	commentService services.CommentService,
	notificationService services.NotificationService,
	postService services.PostService,
	userService services.UserService,
) *CommentHandler {
	return &CommentHandler{
		commentService:     commentService,
		notificationService: notificationService,
		postService:        postService,
		userService:        userService,
	}
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	PostID   uuid.UUID  `json:"post_id"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	Content  string     `json:"content"`
}

// UpdateCommentRequest 更新评论请求
type UpdateCommentRequest struct {
	Content string `json:"content"`
}

// CommentResponse 评论响应
type CommentResponse struct {
	ID        uuid.UUID          `json:"id"`
	PostID    uuid.UUID          `json:"post_id"`
	UserID    uuid.UUID          `json:"user_id"`
	ParentID  *uuid.UUID         `json:"parent_id,omitempty"`
	Depth     int                `json:"depth"`
	Content   string             `json:"content"`
	Status    string             `json:"status"`
	LikeCount int                `json:"like_count"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}

// CommentNodeResponse 评论树节点响应
type CommentNodeResponse struct {
	Comment  CommentResponse      `json:"comment"`
	Children []*CommentNodeResponse `json:"children,omitempty"`
	IsLiked  bool                 `json:"is_liked,omitempty"`
}

// CommentListResponse 评论列表响应
type CommentListResponse struct {
	Comments []*CommentNodeResponse `json:"comments"`
	Total    int                    `json:"total"`
}

// GetByPost 获取文章评论列表（树形结构）
// GET /api/v1/comments/post/{postId}
func (h *CommentHandler) GetByPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	postIDStr := r.PathValue("postId")

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的文章ID", nil, http.StatusBadRequest)
		return
	}

	// 获取评论树
	nodes, err := h.commentService.GetCommentsByPost(ctx, postID)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取评论失败", nil, http.StatusInternalServerError)
		return
	}

	// 尝试获取当前用户 ID（可选，用于显示点赞状态）
	userID, _ := GetUserID(ctx)

	// 转换为响应格式
	responses := make([]*CommentNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		responses = append(responses, toCommentNodeResponse(node, userID, h.commentService))
	}

	WriteJSON(w, CommentListResponse{
		Comments: responses,
		Total:    len(responses),
	}, http.StatusOK)
}

// Create 发表评论
// POST /api/v1/comments
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取当前用户 ID
	userID, err := GetUserID(ctx)
	if err != nil {
		WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "BAD_REQUEST", "请求格式错误", nil, http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.PostID == uuid.Nil {
		WriteAPIError(w, "VALIDATION_ERROR", "文章ID为必填项", nil, http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "评论内容为必填项", nil, http.StatusBadRequest)
		return
	}

	input := &comment.CreateCommentInput{
		PostID:   req.PostID,
		UserID:   userID,
		ParentID: req.ParentID,
		Content:  req.Content,
	}

	c, err := h.commentService.CreateComment(ctx, input)
	if err != nil {
		if errors.Is(err, comment.ErrMaxDepthExceeded) {
			WriteAPIError(w, "VALIDATION_ERROR", "评论嵌套深度超过限制", nil, http.StatusBadRequest)
			return
		}
		if errors.Is(err, comment.ErrParentNotFound) {
			WriteAPIError(w, "NOT_FOUND", "父评论不存在", nil, http.StatusNotFound)
			return
		}
		if errors.Is(err, comment.ErrPostNotFound) {
			WriteAPIError(w, "NOT_FOUND", "文章不存在", nil, http.StatusNotFound)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "创建评论失败: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// 发送通知（异步，不影响响应）
	h.sendCommentNotification(ctx, c, req.PostID, req.ParentID, userID)

	WriteJSON(w, toCommentResponse(c), http.StatusCreated)
}

// Update 编辑评论
// PUT /api/v1/comments/{id}
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的评论ID", nil, http.StatusBadRequest)
		return
	}

	// 获取当前用户 ID
	userID, err := GetUserID(ctx)
	if err != nil {
		WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
		return
	}

	var req UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "BAD_REQUEST", "请求格式错误", nil, http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "评论内容为必填项", nil, http.StatusBadRequest)
		return
	}

	// 获取现有评论
	c, err := h.commentService.GetCommentByID(ctx, id)
	if err != nil {
		WriteAPIError(w, "NOT_FOUND", "评论不存在", nil, http.StatusNotFound)
		return
	}

	// 检查权限：只有评论作者可以编辑
	if c.UserID != userID {
		WriteAPIError(w, "PERMISSION_DENIED", "只能编辑自己的评论", nil, http.StatusForbidden)
		return
	}

	// 更新内容
	c.Content = req.Content
	if err := h.commentService.UpdateComment(ctx, c); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "更新评论失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toCommentResponse(c), http.StatusOK)
}

// Delete 删除评论
// DELETE /api/v1/comments/{id}
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的评论ID", nil, http.StatusBadRequest)
		return
	}

	// 获取当前用户 ID
	userID, err := GetUserID(ctx)
	if err != nil {
		WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
		return
	}

	if err := h.commentService.DeleteComment(ctx, id, userID); err != nil {
		if errors.Is(err, comment.ErrCommentNotFound) {
			WriteAPIError(w, "NOT_FOUND", "评论不存在", nil, http.StatusNotFound)
			return
		}
		if errors.Is(err, comment.ErrPermissionDenied) {
			WriteAPIError(w, "PERMISSION_DENIED", "只能删除自己的评论", nil, http.StatusForbidden)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "删除评论失败", nil, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Like 点赞评论
// POST /api/v1/comments/{id}/like
func (h *CommentHandler) Like(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	commentID, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的评论ID", nil, http.StatusBadRequest)
		return
	}

	// 获取当前用户 ID
	userID, err := GetUserID(ctx)
	if err != nil {
		WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
		return
	}

	if err := h.commentService.LikeComment(ctx, commentID, userID); err != nil {
		if errors.Is(err, comment.ErrCommentNotFound) {
			WriteAPIError(w, "NOT_FOUND", "评论不存在", nil, http.StatusNotFound)
			return
		}
		if errors.Is(err, comment.ErrAlreadyLiked) {
			WriteAPIError(w, "ALREADY_LIKED", "已点赞该评论", nil, http.StatusBadRequest)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "点赞失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "点赞成功"}, http.StatusOK)
}

// Unlike 取消点赞
// DELETE /api/v1/comments/{id}/like
func (h *CommentHandler) Unlike(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	commentID, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的评论ID", nil, http.StatusBadRequest)
		return
	}

	// 获取当前用户 ID
	userID, err := GetUserID(ctx)
	if err != nil {
		WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
		return
	}

	if err := h.commentService.UnlikeComment(ctx, commentID, userID); err != nil {
		if errors.Is(err, comment.ErrCommentNotFound) {
			WriteAPIError(w, "NOT_FOUND", "评论不存在", nil, http.StatusNotFound)
			return
		}
		if errors.Is(err, comment.ErrNotLiked) {
			WriteAPIError(w, "NOT_LIKED", "未点赞该评论", nil, http.StatusBadRequest)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "取消点赞失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "取消点赞成功"}, http.StatusOK)
}

// Approve 审核批准评论（需权限）
// PUT /api/v1/comments/{id}/approve
func (h *CommentHandler) Approve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的评论ID", nil, http.StatusBadRequest)
		return
	}

	if err := h.commentService.ApproveComment(ctx, id); err != nil {
		if errors.Is(err, comment.ErrCommentNotFound) {
			WriteAPIError(w, "NOT_FOUND", "评论不存在", nil, http.StatusNotFound)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "审核失败: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "评论已批准"}, http.StatusOK)
}

// Reject 审核拒绝评论（需权限）
// PUT /api/v1/comments/{id}/reject
func (h *CommentHandler) Reject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的评论ID", nil, http.StatusBadRequest)
		return
	}

	if err := h.commentService.RejectComment(ctx, id); err != nil {
		if errors.Is(err, comment.ErrCommentNotFound) {
			WriteAPIError(w, "NOT_FOUND", "评论不存在", nil, http.StatusNotFound)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "审核失败: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "评论已拒绝"}, http.StatusOK)
}

// toCommentResponse 转换评论为响应格式
func toCommentResponse(c *comment.Comment) CommentResponse {
	return CommentResponse{
		ID:        c.ID,
		PostID:    c.PostID,
		UserID:    c.UserID,
		ParentID:  c.ParentID,
		Depth:     c.Depth,
		Content:   c.Content,
		Status:    string(c.Status),
		LikeCount: c.LikeCount,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// toCommentNodeResponse 转换评论树节点为响应格式
func toCommentNodeResponse(node *services.CommentNode, userID uuid.UUID, svc services.CommentService) *CommentNodeResponse {
	resp := &CommentNodeResponse{
		Comment:  toCommentNodeCommentResponse(node.Comment),
		Children: make([]*CommentNodeResponse, 0, len(node.Children)),
	}

	// 如果用户已登录，检查点赞状态
	if userID != uuid.Nil {
		isLiked, _ := svc.IsLiked(context.Background(), node.Comment.ID, userID)
		resp.IsLiked = isLiked
	}

	// 递归转换子评论
	for _, child := range node.Children {
		resp.Children = append(resp.Children, toCommentNodeResponse(child, userID, svc))
	}

	return resp
}

// toCommentNodeCommentResponse 转换 CommentNode 中的评论为响应格式（不含 isLiked）
func toCommentNodeCommentResponse(c *comment.Comment) CommentResponse {
	return CommentResponse{
		ID:        c.ID,
		PostID:    c.PostID,
		UserID:    c.UserID,
		ParentID:  c.ParentID,
		Depth:     c.Depth,
		Content:   c.Content,
		Status:    string(c.Status),
		LikeCount: c.LikeCount,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// sendCommentNotification 发送评论通知（异步）
func (h *CommentHandler) sendCommentNotification(ctx context.Context, c *comment.Comment, postID uuid.UUID, parentID *uuid.UUID, commentUserID uuid.UUID) {
	// 如果没有通知服务，跳过
	if h.notificationService == nil {
		return
	}

	// 在后台异步发送通知
	go func() {
		bgCtx := context.Background()

		// 获取文章信息
		post, err := h.postService.GetByID(bgCtx, postID)
		if err != nil {
			return
		}

		// 获取评论者信息
		commentAuthor, err := h.userService.GetByID(bgCtx, commentUserID)
		if err != nil {
			return
		}

		if parentID != nil {
			// 回复评论：通知被回复用户
			parentComment, err := h.commentService.GetCommentByID(bgCtx, *parentID)
			if err != nil {
				return
			}

			// 获取被回复用户信息
			parentAuthor, err := h.userService.GetByID(bgCtx, parentComment.UserID)
			if err != nil {
				return
			}

			h.notificationService.SendReplyNotification(bgCtx, c, parentComment, post, commentAuthor, parentAuthor)
		} else {
			// 顶层评论：通知文章作者
			postAuthor, err := h.userService.GetByID(bgCtx, post.AuthorID)
			if err != nil {
				return
			}

			h.notificationService.SendCommentNotification(bgCtx, c, post, postAuthor, commentAuthor)
		}
	}()
}