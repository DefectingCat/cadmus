// Package handlers 提供了 Cadmus API 的 HTTP 处理器实现。
//
// 该文件包含评论相关的核心逻辑，包括：
//   - 评论和回复的 CRUD 操作
//   - 评论树形结构处理
//   - 评论点赞功能
//   - 评论审核管理
//   - 评论通知发送
//
// 主要用途：
//
//	用于处理文章评论的完整生命周期管理，支持嵌套回复和审核流程。
//
// 注意事项：
//   - 发表评论需要用户认证
//   - 评论支持多级嵌套，但有深度限制
//   - 管理员可以审核、批准或拒绝评论
//
// 作者：xfy
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

// CommentHandler 评论 API 处理器。
//
// 该处理器负责处理所有评论相关的 HTTP 请求，包括评论的增删改查、
// 点赞功能和审核管理。支持树形结构的嵌套评论。
//
// 注意事项：
//   - 需要注入 CommentService 处理业务逻辑
//   - 可选注入 NotificationService 实现评论通知
type CommentHandler struct {
	// commentService 评论服务，处理评论业务逻辑
	commentService services.CommentService

	// notificationService 通知服务，发送评论通知（可选）
	notificationService services.NotificationService

	// postService 文章服务，用于通知时获取文章信息
	postService services.PostService

	// userService 用户服务，用于通知时获取用户信息
	userService services.UserService
}

// NewCommentHandler 创建评论处理器。
//
// 创建一个基础的评论处理器，不支持通知功能。
//
// 参数：
//   - commentService: 评论服务，处理评论业务逻辑
//
// 返回值：
//   - *CommentHandler: 新创建的评论处理器实例
func NewCommentHandler(commentService services.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// NewCommentHandlerWithNotifications 创建带通知功能的评论处理器。
//
// 创建一个完整的评论处理器，支持评论通知功能。
// 当用户发表评论或回复时，会异步发送通知给文章作者或被回复用户。
//
// 参数：
//   - commentService: 评论服务
//   - notificationService: 通知服务
//   - postService: 文章服务，用于获取文章信息
//   - userService: 用户服务，用于获取用户信息
//
// 返回值：
//   - *CommentHandler: 新创建的完整功能评论处理器实例
func NewCommentHandlerWithNotifications(
	commentService services.CommentService,
	notificationService services.NotificationService,
	postService services.PostService,
	userService services.UserService,
) *CommentHandler {
	return &CommentHandler{
		commentService:      commentService,
		notificationService: notificationService,
		postService:         postService,
		userService:         userService,
	}
}

// CreateCommentRequest 创建评论请求结构体。
type CreateCommentRequest struct {
	// PostID 文章 ID，必填
	PostID uuid.UUID `json:"post_id"`

	// ParentID 父评论 ID，回复时填写（可选）
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// Content 评论内容，必填
	Content string `json:"content"`
}

// UpdateCommentRequest 更新评论请求结构体。
type UpdateCommentRequest struct {
	// Content 新的评论内容
	Content string `json:"content"`
}

// CommentResponse 评论响应结构体。
type CommentResponse struct {
	// ID 评论唯一标识符
	ID uuid.UUID `json:"id"`

	// PostID 所属文章 ID
	PostID uuid.UUID `json:"post_id"`

	// UserID 评论者 ID
	UserID uuid.UUID `json:"user_id"`

	// ParentID 父评论 ID（顶层评论为 nil）
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// Depth 嵌套深度，顶层评论为 0
	Depth int `json:"depth"`

	// Content 评论内容
	Content string `json:"content"`

	// Status 评论状态
	Status string `json:"status"`

	// LikeCount 点赞数
	LikeCount int `json:"like_count"`

	// CreatedAt 创建时间
	CreatedAt string `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt string `json:"updated_at"`
}

// CommentNodeResponse 评论树节点响应结构体。
//
// 用于返回树形结构的评论数据，包含子评论列表。
type CommentNodeResponse struct {
	// Comment 评论内容
	Comment CommentResponse `json:"comment"`

	// Children 子评论列表
	Children []*CommentNodeResponse `json:"children,omitempty"`

	// IsLiked 当前用户是否已点赞
	IsLiked bool `json:"is_liked,omitempty"`
}

// CommentListResponse 评论列表响应结构体。
type CommentListResponse struct {
	// Comments 评论树列表
	Comments []*CommentNodeResponse `json:"comments"`

	// Total 评论总数
	Total int `json:"total"`
}

// GetByPost 获取文章评论列表（树形结构）。
//
// 获取指定文章的所有评论，以树形结构返回。
// 支持显示当前用户的点赞状态。
//
// 路由：GET /api/v1/comments/post/{postId}
//
// 参数：
//   - postId: 文章 ID（路径参数）
//
// 返回值（通过响应体）：
//   - comments: 评论树列表
//   - total: 评论总数
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

	// 收集所有评论 ID 用于批量查询点赞状态
	commentIDs := collectCommentIDs(nodes)

	// 批量查询点赞状态（避免 N+1）
	likesMap := make(map[uuid.UUID]bool)
	if userID != uuid.Nil && len(commentIDs) > 0 {
		likesMap, _ = h.commentService.GetLikesBatch(ctx, commentIDs, userID)
	}

	// 转换为响应格式
	responses := make([]*CommentNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		responses = append(responses, toCommentNodeResponseWithLikes(node, likesMap))
	}

	WriteJSON(w, CommentListResponse{
		Comments: responses,
		Total:    len(responses),
	}, http.StatusOK)
}

// Create 发表评论。
//
// 发表新评论或回复。需要用户认证。
// 支持多级嵌套回复，但有深度限制。
// 发表成功后会异步发送通知。
//
// 路由：POST /api/v1/comments
//
// 参数（通过请求体）：
//   - post_id: 文章 ID（必填）
//   - parent_id: 父评论 ID（回复时必填）
//   - content: 评论内容（必填）
//
// 返回值（通过响应体）：
//   - 新创建的评论信息
//
// 可能的错误：
//   - UNAUTHORIZED: 未登录
//   - BAD_REQUEST: 请求格式错误或缺少必填字段
//   - VALIDATION_ERROR: 评论嵌套深度超过限制
//   - NOT_FOUND: 文章或父评论不存在
//   - INTERNAL_ERROR: 创建失败
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

// Update 编辑评论。
//
// 编辑已有的评论内容。需要用户认证且只能编辑自己的评论。
//
// 路由：PUT /api/v1/comments/{id}
//
// 参数：
//   - id: 评论 ID（路径参数）
//   - content: 新的评论内容（请求体）
//
// 返回值（通过响应体）：
//   - 更新后的评论信息
//
// 可能的错误：
//   - UNAUTHORIZED: 未登录
//   - BAD_REQUEST: 无效的评论 ID 或请求格式
//   - NOT_FOUND: 评论不存在
//   - PERMISSION_DENIED: 只能编辑自己的评论
//   - INTERNAL_ERROR: 更新失败
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

// Delete 删除评论。
//
// 删除指定的评论。需要用户认证且只能删除自己的评论。
//
// 路由：DELETE /api/v1/comments/{id}
//
// 参数：
//   - id: 评论 ID（路径参数）
//
// 可能的错误：
//   - UNAUTHORIZED: 未登录
//   - BAD_REQUEST: 无效的评论 ID
//   - NOT_FOUND: 评论不存在
//   - PERMISSION_DENIED: 只能删除自己的评论
//   - INTERNAL_ERROR: 删除失败
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

// Like 点赞评论。
//
// 为指定评论点赞。需要用户认证。
// 每个用户只能对同一评论点赞一次。
//
// 路由：POST /api/v1/comments/{id}/like
//
// 参数：
//   - id: 评论 ID（路径参数）
//
// 返回值（通过响应体）：
//   - message: 点赞成功提示
//
// 可能的错误：
//   - UNAUTHORIZED: 未登录
//   - BAD_REQUEST: 无效的评论 ID
//   - NOT_FOUND: 评论不存在
//   - ALREADY_LIKED: 已点赞该评论
//   - INTERNAL_ERROR: 点赞失败
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

// Unlike 取消点赞。
//
// 取消对评论的点赞。需要用户认证。
//
// 路由：DELETE /api/v1/comments/{id}/like
//
// 参数：
//   - id: 评论 ID（路径参数）
//
// 返回值（通过响应体）：
//   - message: 取消点赞成功提示
//
// 可能的错误：
//   - UNAUTHORIZED: 未登录
//   - BAD_REQUEST: 无效的评论 ID
//   - NOT_FOUND: 评论不存在
//   - NOT_LIKED: 未点赞该评论
//   - INTERNAL_ERROR: 取消点赞失败
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

// Approve 审核批准评论（需权限）。
//
// 批准待审核的评论，使其公开可见。需要管理员权限。
//
// 路由：PUT /api/v1/comments/{id}/approve
//
// 参数：
//   - id: 评论 ID（路径参数）
//
// 返回值（通过响应体）：
//   - message: 批准成功提示
//
// 可能的错误：
//   - BAD_REQUEST: 无效的评论 ID
//   - NOT_FOUND: 评论不存在
//   - INTERNAL_ERROR: 审核失败
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

// Reject 审核拒绝评论（需权限）。
//
// 拒绝待审核的评论，使其不可见。需要管理员权限。
//
// 路由：PUT /api/v1/comments/{id}/reject
//
// 参数：
//   - id: 评论 ID（路径参数）
//
// 返回值（通过响应体）：
//   - message: 拒绝成功提示
//
// 可能的错误：
//   - BAD_REQUEST: 无效的评论 ID
//   - NOT_FOUND: 评论不存在
//   - INTERNAL_ERROR: 审核失败
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

// AdminListComments 获取评论列表（管理员）。
//
// 获取所有评论的列表，支持按状态筛选和分页。
// 需要管理员权限。
//
// 路由：GET /api/v1/admin/comments
//
// 查询参数：
//   - status: 评论状态筛选，默认为 pending
//   - page: 页码，默认 1
//   - per_page: 每页数量，默认 20，最大 100
//
// 返回值（通过响应体）：
//   - comments: 评论列表
//   - total: 总数
//   - page: 当前页码
//   - per_page: 每页数量
func (h *CommentHandler) AdminListComments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取查询参数
	statusStr := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")

	// 默认值
	page := 1
	perPage := 20
	if pageStr != "" {
		if p, err := parseInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr != "" {
		if p, err := parseInt(perPageStr); err == nil && p > 0 && p <= 100 {
			perPage = p
		}
	}

	// 解析状态
	var status comment.CommentStatus
	if statusStr != "" {
		status = comment.CommentStatus(statusStr)
		if !status.IsValid() {
			WriteAPIError(w, "BAD_REQUEST", "无效的评论状态", nil, http.StatusBadRequest)
			return
		}
	} else {
		status = comment.StatusPending // 默认获取待审核评论
	}

	offset := (page - 1) * perPage
	comments, total, err := h.commentService.GetCommentsByStatus(ctx, status, offset, perPage)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取评论列表失败", nil, http.StatusInternalServerError)
		return
	}

	// 构建响应
	responses := make([]CommentResponse, 0, len(comments))
	for _, c := range comments {
		responses = append(responses, toCommentResponse(c))
	}

	WriteJSON(w, AdminCommentListResponse{
		Comments: responses,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
	}, http.StatusOK)
}

// BatchApproveRequest 批量操作请求结构体。
type BatchApproveRequest struct {
	// IDs 评论 ID 列表
	IDs []string `json:"ids"`
}

// BatchApprove 批量批准评论。
//
// 批量批准多条待审核的评论。需要管理员权限。
//
// 路由：PUT /api/v1/admin/comments/batch-approve
//
// 参数（通过请求体）：
//   - ids: 评论 ID 列表
//
// 返回值（通过响应体）：
//   - message: 批量审核成功提示
//
// 可能的错误：
//   - BAD_REQUEST: 请求格式错误或无效的评论 ID
//   - VALIDATION_ERROR: 未选择要审核的评论
//   - INTERNAL_ERROR: 批量审核失败
func (h *CommentHandler) BatchApprove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req BatchApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "BAD_REQUEST", "请求格式错误", nil, http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		WriteAPIError(w, "VALIDATION_ERROR", "请选择要审核的评论", nil, http.StatusBadRequest)
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			WriteAPIError(w, "BAD_REQUEST", "无效的评论ID: "+idStr, nil, http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}

	if err := h.commentService.BatchApproveComments(ctx, ids); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "批量审核失败: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "批量审核成功"}, http.StatusOK)
}

// BatchReject 批量拒绝评论。
//
// 批量拒绝多条待审核的评论。需要管理员权限。
//
// 路由：PUT /api/v1/admin/comments/batch-reject
//
// 参数（通过请求体）：
//   - ids: 评论 ID 列表
//
// 返回值（通过响应体）：
//   - message: 批量拒绝成功提示
//
// 可能的错误：
//   - BAD_REQUEST: 请求格式错误或无效的评论 ID
//   - VALIDATION_ERROR: 未选择要审核的评论
//   - INTERNAL_ERROR: 批量拒绝失败
func (h *CommentHandler) BatchReject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req BatchApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "BAD_REQUEST", "请求格式错误", nil, http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		WriteAPIError(w, "VALIDATION_ERROR", "请选择要审核的评论", nil, http.StatusBadRequest)
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			WriteAPIError(w, "BAD_REQUEST", "无效的评论ID: "+idStr, nil, http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}

	if err := h.commentService.BatchRejectComments(ctx, ids); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "批量拒绝失败: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "批量拒绝成功"}, http.StatusOK)
}

// BatchDelete 批量删除评论。
//
// 批量删除多条评论。需要管理员权限。
//
// 路由：DELETE /api/v1/admin/comments/batch-delete
//
// 参数（通过请求体）：
//   - ids: 评论 ID 列表
//
// 返回值（通过响应体）：
//   - message: 批量删除成功提示
//
// 可能的错误：
//   - BAD_REQUEST: 请求格式错误或无效的评论 ID
//   - VALIDATION_ERROR: 未选择要删除的评论
//   - INTERNAL_ERROR: 批量删除失败
func (h *CommentHandler) BatchDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req BatchApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "BAD_REQUEST", "请求格式错误", nil, http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		WriteAPIError(w, "VALIDATION_ERROR", "请选择要删除的评论", nil, http.StatusBadRequest)
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			WriteAPIError(w, "BAD_REQUEST", "无效的评论ID: "+idStr, nil, http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}

	if err := h.commentService.BatchDeleteComments(ctx, ids); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "批量删除失败: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "批量删除成功"}, http.StatusOK)
}

// AdminDeleteComment 管理员删除单条评论。
//
// 管理员删除指定的评论，无需是评论作者。
//
// 路由：DELETE /api/v1/admin/comments/{id}
//
// 参数：
//   - id: 评论 ID（路径参数）
//
// 可能的错误：
//   - BAD_REQUEST: 无效的评论 ID
//   - NOT_FOUND: 评论不存在
//   - INTERNAL_ERROR: 删除失败
func (h *CommentHandler) AdminDeleteComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的评论ID", nil, http.StatusBadRequest)
		return
	}

	if err := h.commentService.DeleteCommentAdmin(ctx, id); err != nil {
		if errors.Is(err, comment.ErrCommentNotFound) {
			WriteAPIError(w, "NOT_FOUND", "评论不存在", nil, http.StatusNotFound)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "删除评论失败", nil, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminCommentListResponse 管理员评论列表响应结构体。
type AdminCommentListResponse struct {
	// Comments 评论列表
	Comments []CommentResponse `json:"comments"`

	// Total 评论总数
	Total int `json:"total"`

	// Page 当前页码
	Page int `json:"page"`

	// PerPage 每页数量
	PerPage int `json:"per_page"`
}

// toCommentResponse 转换 Comment 实体到 CommentResponse 响应结构。
//
// 将内部的 Comment 实体转换为对外暴露的 CommentResponse 结构。
//
// 参数：
//   - c: Comment 实体指针
//
// 返回值：
//   - CommentResponse: 评论响应结构
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

// toCommentNodeCommentResponse 转换 CommentNode 中的评论为响应格式。
//
// 用于树形结构中转换单个评论节点，不包含点赞状态。
//
// 参数：
//   - c: Comment 实体指针
//
// 返回值：
//   - CommentResponse: 评论响应结构
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

// collectCommentIDs 从评论树中收集所有评论 ID。
//
// 递归遍历评论树，收集所有评论的 ID，用于批量查询点赞状态。
//
// 参数：
//   - nodes: 评论树节点列表
//
// 返回值：
//   - []uuid.UUID: 所有评论 ID 列表
func collectCommentIDs(nodes []*services.CommentNode) []uuid.UUID {
	ids := make([]uuid.UUID, 0)
	for _, node := range nodes {
		ids = append(ids, node.Comment.ID)
		ids = append(ids, collectCommentIDs(node.Children)...)
	}
	return ids
}

// toCommentNodeResponseWithLikes 转换评论树节点为响应格式。
//
// 使用预查询的点赞状态映射，递归转换整个评论树。
// 这种方式避免了 N+1 查询问题，提高性能。
//
// 参数：
//   - node: 评论树节点
//   - likesMap: 预查询的点赞状态映射（评论 ID -> 是否点赞）
//
// 返回值：
//   - *CommentNodeResponse: 评论树节点响应结构
func toCommentNodeResponseWithLikes(node *services.CommentNode, likesMap map[uuid.UUID]bool) *CommentNodeResponse {
	resp := &CommentNodeResponse{
		Comment:  toCommentNodeCommentResponse(node.Comment),
		IsLiked:  likesMap[node.Comment.ID],
		Children: make([]*CommentNodeResponse, 0, len(node.Children)),
	}

	// 递归转换子评论
	for _, child := range node.Children {
		resp.Children = append(resp.Children, toCommentNodeResponseWithLikes(child, likesMap))
	}

	return resp
}

// sendCommentNotification 发送评论通知（异步）。
//
// 在后台异步发送评论通知：
//   - 回复评论时：通知被回复的用户
//   - 顶层评论时：通知文章作者
//
// 该方法在 goroutine 中执行，不阻塞主请求。
// 如果未配置通知服务，则直接返回不做任何处理。
//
// 参数：
//   - ctx: 上下文（未使用，保留用于未来扩展）
//   - c: 新创建的评论
//   - postID: 文章 ID
//   - parentID: 父评论 ID（回复时有效）
//   - commentUserID: 评论者 ID
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
