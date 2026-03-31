package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/services"
)

// PostHandler 文章 API 处理器
type PostHandler struct {
	postService services.PostService
}

// NewPostHandler 创建文章处理器
func NewPostHandler(postService services.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

// CreatePostRequest 创建文章请求
type CreatePostRequest struct {
	Title         string      `json:"title"`
	Slug          string      `json:"slug"`
	Content       string      `json:"content"`
	ContentText   string      `json:"content_text,omitempty"`
	Excerpt       string      `json:"excerpt,omitempty"`
	CategoryID    *uuid.UUID  `json:"category_id,omitempty"`
	Status        string      `json:"status,omitempty"`
	FeaturedImage string      `json:"featured_image,omitempty"`
	SEOTitle      string      `json:"seo_title,omitempty"`
	SEODescription string     `json:"seo_description,omitempty"`
	SEOKeywords   []string    `json:"seo_keywords,omitempty"`
	SeriesID      *uuid.UUID  `json:"series_id,omitempty"`
	SeriesOrder   int         `json:"series_order,omitempty"`
	IsPaid        bool        `json:"is_paid,omitempty"`
	Price         *float64    `json:"price,omitempty"`
	TagIDs        []uuid.UUID `json:"tag_ids,omitempty"`
}

// UpdatePostRequest 更新文章请求
type UpdatePostRequest struct {
	Title         string      `json:"title"`
	Slug          string      `json:"slug"`
	Content       string      `json:"content"`
	ContentText   string      `json:"content_text,omitempty"`
	Excerpt       string      `json:"excerpt,omitempty"`
	CategoryID    *uuid.UUID  `json:"category_id,omitempty"`
	Status        string      `json:"status,omitempty"`
	FeaturedImage string      `json:"featured_image,omitempty"`
	SEOTitle      string      `json:"seo_title,omitempty"`
	SEODescription string     `json:"seo_description,omitempty"`
	SEOKeywords   []string    `json:"seo_keywords,omitempty"`
	SeriesID      *uuid.UUID  `json:"series_id,omitempty"`
	SeriesOrder   int         `json:"series_order,omitempty"`
	IsPaid        bool        `json:"is_paid"`
	Price         *float64    `json:"price,omitempty"`
	TagIDs        []uuid.UUID `json:"tag_ids,omitempty"`
}

// PostResponse 文章响应
type PostResponse struct {
	ID            uuid.UUID   `json:"id"`
	AuthorID      uuid.UUID   `json:"author_id"`
	Title         string      `json:"title"`
	Slug          string      `json:"slug"`
	Content       string      `json:"content"`
	Excerpt       string      `json:"excerpt,omitempty"`
	CategoryID    *uuid.UUID  `json:"category_id,omitempty"`
	Status        string      `json:"status"`
	PublishAt     *time.Time  `json:"publish_at,omitempty"`
	FeaturedImage string      `json:"featured_image,omitempty"`
	SEOMeta       SEOMetaResp `json:"seo_meta,omitempty"`
	ViewCount     int         `json:"view_count"`
	LikeCount     int         `json:"like_count"`
	CommentCount  int         `json:"comment_count"`
	SeriesID      *uuid.UUID  `json:"series_id,omitempty"`
	SeriesOrder   int         `json:"series_order"`
	IsPaid        bool        `json:"is_paid"`
	Price         *float64    `json:"price,omitempty"`
	Version       int         `json:"version"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type SEOMetaResp struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

// PostListResponse 文章列表响应
type PostListResponse struct {
	Posts     []PostResponse `json:"posts"`
	Total     int            `json:"total"`
	Page      int            `json:"page"`
	PageSize  int            `json:"page_size"`
}

// List 文章列表
// GET /api/v1/posts
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 解析分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 解析筛选参数
	filters := post.PostListFilters{
		Status:   post.PostStatus(r.URL.Query().Get("status")),
		Search:   r.URL.Query().Get("search"),
	}

	if authorID := r.URL.Query().Get("author_id"); authorID != "" {
		if id, err := uuid.Parse(authorID); err == nil {
			filters.AuthorID = id
		}
	}
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		if id, err := uuid.Parse(categoryID); err == nil {
			filters.CategoryID = id
		}
	}

	posts, total, err := h.postService.List(ctx, filters, page, pageSize)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取文章列表失败", nil, http.StatusInternalServerError)
		return
	}

	responses := make([]PostResponse, 0, len(posts))
	for _, p := range posts {
		responses = append(responses, toPostResponse(p))
	}

	WriteJSON(w, PostListResponse{
		Posts:    responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK)
}

// Get 文章详情
// GET /api/v1/posts/{slug}
func (h *PostHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := r.PathValue("slug")
	if slug == "" {
		WriteAPIError(w, "BAD_REQUEST", "缺少文章标识", nil, http.StatusBadRequest)
		return
	}

	p, err := h.postService.GetBySlug(ctx, slug)
	if err != nil {
		WriteAPIError(w, "NOT_FOUND", "文章不存在", nil, http.StatusNotFound)
		return
	}

	// 增加浏览量
	go h.postService.IncrementViewCount(ctx, p.ID)

	WriteJSON(w, toPostResponse(p), http.StatusOK)
}

// Create 创建文章
// POST /api/v1/posts
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取当前用户 ID
	userID, err := GetUserID(ctx)
	if err != nil {
		WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "BAD_REQUEST", "请求格式错误", nil, http.StatusBadRequest)
		return
	}

	p := &post.Post{
		AuthorID:      userID,
		Title:         req.Title,
		Slug:          req.Slug,
		Content:       []byte(req.Content),
		ContentText:   req.ContentText,
		Excerpt:       req.Excerpt,
		FeaturedImage: req.FeaturedImage,
		SeriesID:      req.SeriesID,
		SeriesOrder:   req.SeriesOrder,
		IsPaid:        req.IsPaid,
		Price:         req.Price,
	}

	if req.CategoryID != nil {
		p.CategoryID = *req.CategoryID
	}
	if req.Status != "" {
		p.Status = post.PostStatus(req.Status)
	}
	p.SEOMeta.Title = req.SEOTitle
	p.SEOMeta.Description = req.SEODescription
	p.SEOMeta.Keywords = req.SEOKeywords

	if err := h.postService.Create(ctx, p, req.TagIDs); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "创建文章失败: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toPostResponse(p), http.StatusCreated)
}

// Update 更新文章
// PUT /api/v1/posts/{id}
func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的文章ID", nil, http.StatusBadRequest)
		return
	}

	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "BAD_REQUEST", "请求格式错误", nil, http.StatusBadRequest)
		return
	}

	// 获取现有文章
	p, err := h.postService.GetByID(ctx, id)
	if err != nil {
		WriteAPIError(w, "NOT_FOUND", "文章不存在", nil, http.StatusNotFound)
		return
	}

	// 更新字段
	p.Title = req.Title
	p.Slug = req.Slug
	p.Content = []byte(req.Content)
	p.ContentText = req.ContentText
	p.Excerpt = req.Excerpt
	p.FeaturedImage = req.FeaturedImage
	p.SeriesID = req.SeriesID
	p.SeriesOrder = req.SeriesOrder
	p.IsPaid = req.IsPaid
	p.Price = req.Price

	if req.CategoryID != nil {
		p.CategoryID = *req.CategoryID
	}
	if req.Status != "" {
		p.Status = post.PostStatus(req.Status)
	}
	p.SEOMeta.Title = req.SEOTitle
	p.SEOMeta.Description = req.SEODescription
	p.SEOMeta.Keywords = req.SEOKeywords

	if err := h.postService.Update(ctx, p, req.TagIDs); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "更新文章失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toPostResponse(p), http.StatusOK)
}

// Delete 删除文章
// DELETE /api/v1/posts/{id}
func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的文章ID", nil, http.StatusBadRequest)
		return
	}

	if err := h.postService.Delete(ctx, id); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "删除文章失败", nil, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Publish 发布文章
// POST /api/v1/posts/{id}/publish
func (h *PostHandler) Publish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的文章ID", nil, http.StatusBadRequest)
		return
	}

	if err := h.postService.Publish(ctx, id); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "发布文章失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "文章已发布"}, http.StatusOK)
}

// Search 搜索文章
// GET /api/v1/search
func (h *PostHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("q")
	if query == "" {
		WriteAPIError(w, "BAD_REQUEST", "缺少搜索关键词", nil, http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	posts, total, err := h.postService.Search(ctx, query, page, pageSize)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "搜索失败", nil, http.StatusInternalServerError)
		return
	}

	responses := make([]PostResponse, 0, len(posts))
	for _, p := range posts {
		responses = append(responses, toPostResponse(p))
	}

	WriteJSON(w, PostListResponse{
		Posts:    responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK)
}

// Versions 获取版本历史
// GET /api/v1/posts/{id}/versions
func (h *PostHandler) Versions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的文章ID", nil, http.StatusBadRequest)
		return
	}

	versions, err := h.postService.GetVersions(ctx, id)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取版本历史失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, versions, http.StatusOK)
}

// RollbackRequest 回滚请求
type RollbackRequest struct {
	Version int `json:"version"`
}

// Rollback 回滚到指定版本
// POST /api/v1/posts/{id}/rollback
func (h *PostHandler) Rollback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的文章ID", nil, http.StatusBadRequest)
		return
	}

	var req RollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "BAD_REQUEST", "请求格式错误", nil, http.StatusBadRequest)
		return
	}

	if req.Version < 1 {
		WriteAPIError(w, "BAD_REQUEST", "版本号必须大于0", nil, http.StatusBadRequest)
		return
	}

	if err := h.postService.Rollback(ctx, id, req.Version); err != nil {
		if err.Error() == "版本不存在" {
			WriteAPIError(w, "NOT_FOUND", "版本不存在", nil, http.StatusNotFound)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "回滚失败: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "文章已回滚到版本 " + strconv.Itoa(req.Version)}, http.StatusOK)
}

// GetUserPosts 获取用户的文章列表
// GET /api/v1/users/{id}/posts
func (h *PostHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 解析用户 ID
	idStr := r.PathValue("id")
	authorID, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的用户ID", nil, http.StatusBadRequest)
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 解析状态筛选参数
	status := post.PostStatus(r.URL.Query().Get("status"))
	if status != "" && !status.IsValid() {
		WriteAPIError(w, "BAD_REQUEST", "无效的文章状态", nil, http.StatusBadRequest)
		return
	}

	posts, total, err := h.postService.GetByAuthor(ctx, authorID, status, page, limit)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取用户文章列表失败", nil, http.StatusInternalServerError)
		return
	}

	responses := make([]PostResponse, 0, len(posts))
	for _, p := range posts {
		responses = append(responses, toPostResponse(p))
	}

	WriteJSON(w, PostListResponse{
		Posts:    responses,
		Total:    total,
		Page:     page,
		PageSize: limit,
	}, http.StatusOK)
}

// Like 点赞文章
// POST /api/v1/posts/{id}/like
func (h *PostHandler) Like(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	postID, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的文章ID", nil, http.StatusBadRequest)
		return
	}

	// 获取当前用户 ID
	userID, err := GetUserID(ctx)
	if err != nil {
		WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
		return
	}

	if err := h.postService.LikePost(ctx, postID, userID); err != nil {
		if err.Error() == "已点赞过该文章" {
			WriteAPIError(w, "ALREADY_LIKED", "已点赞过该文章", nil, http.StatusBadRequest)
			return
		}
		if err.Error() == "文章不存在" {
			WriteAPIError(w, "NOT_FOUND", "文章不存在", nil, http.StatusNotFound)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "点赞失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "点赞成功"}, http.StatusOK)
}

// Unlike 取消点赞
// DELETE /api/v1/posts/{id}/like
func (h *PostHandler) Unlike(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	postID, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的文章ID", nil, http.StatusBadRequest)
		return
	}

	// 获取当前用户 ID
	userID, err := GetUserID(ctx)
	if err != nil {
		WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
		return
	}

	if err := h.postService.UnlikePost(ctx, postID, userID); err != nil {
		if err.Error() == "未点赞过该文章" {
			WriteAPIError(w, "NOT_LIKED", "未点赞过该文章", nil, http.StatusBadRequest)
			return
		}
		if err.Error() == "文章不存在" {
			WriteAPIError(w, "NOT_FOUND", "文章不存在", nil, http.StatusNotFound)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "取消点赞失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "取消点赞成功"}, http.StatusOK)
}

func toPostResponse(p *post.Post) PostResponse {
	var categoryID *uuid.UUID
	if p.CategoryID != uuid.Nil {
		categoryID = &p.CategoryID
	}

	return PostResponse{
		ID:            p.ID,
		AuthorID:      p.AuthorID,
		Title:         p.Title,
		Slug:          p.Slug,
		Content:       string(p.Content),
		Excerpt:       p.Excerpt,
		CategoryID:    categoryID,
		Status:        string(p.Status),
		PublishAt:     p.PublishAt,
		FeaturedImage: p.FeaturedImage,
		SEOMeta: SEOMetaResp{
			Title:       p.SEOMeta.Title,
			Description: p.SEOMeta.Description,
			Keywords:    p.SEOMeta.Keywords,
		},
		ViewCount:     p.ViewCount,
		LikeCount:     p.LikeCount,
		CommentCount:  p.CommentCount,
		SeriesID:      p.SeriesID,
		SeriesOrder:   p.SeriesOrder,
		IsPaid:        p.IsPaid,
		Price:         p.Price,
		Version:       p.Version,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
