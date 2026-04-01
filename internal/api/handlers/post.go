// Package handlers 提供了 Cadmus API 的 HTTP 处理器实现。
//
// 该文件包含文章相关的核心逻辑，包括：
//   - 文章的 CRUD 操作（创建、读取、更新、删除）
//   - 文章发布和版本管理
//   - 文章点赞功能
//   - 文章搜索功能
//
// 主要用途：
//
//	用于处理博客文章的完整生命周期管理，从创建到发布到版本控制。
//
// 注意事项：
//   - 创建和更新文章需要用户认证
//   - 发布文章需要通过审核流程
//   - 文章版本历史自动保存
//
// 作者：xfy
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

// PostHandler 文章 API 处理器。
//
// 该处理器负责处理所有文章相关的 HTTP 请求，包括文章的增删改查、
// 发布管理、版本控制和点赞功能。
//
// 注意事项：
//   - 需要注入 PostService 处理业务逻辑
//   - 部分操作需要用户认证和权限验证
type PostHandler struct {
	// postService 文章服务，处理文章业务逻辑
	postService services.PostService
}

// NewPostHandler 创建文章处理器。
//
// 创建一个新的文章处理器实例。
//
// 参数：
//   - postService: 文章服务，处理文章业务逻辑
//
// 返回值：
//   - *PostHandler: 新创建的文章处理器实例
func NewPostHandler(postService services.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

// CreatePostRequest 创建文章请求结构体。
//
// 包含创建文章所需的所有字段，包括 SEO 元数据和付费设置。
type CreatePostRequest struct {
	// Title 文章标题，必填
	Title string `json:"title"`

	// Slug 文章 URL 别名，必填，用于友好 URL
	Slug string `json:"slug"`

	// Content 文章内容，支持 Markdown 格式
	Content string `json:"content"`

	// ContentText 文章纯文本内容，用于搜索索引（可选）
	ContentText string `json:"content_text,omitempty"`

	// Excerpt 文章摘要，用于列表展示（可选）
	Excerpt string `json:"excerpt,omitempty"`

	// CategoryID 所属分类 ID（可选）
	CategoryID *uuid.UUID `json:"category_id,omitempty"`

	// Status 文章状态：draft, published, archived 等
	Status string `json:"status,omitempty"`

	// FeaturedImage 特色图片 URL（可选）
	FeaturedImage string `json:"featured_image,omitempty"`

	// SEOTitle SEO 标题（可选）
	SEOTitle string `json:"seo_title,omitempty"`

	// SEODescription SEO 描述（可选）
	SEODescription string `json:"seo_description,omitempty"`

	// SEOKeywords SEO 关键词列表（可选）
	SEOKeywords []string `json:"seo_keywords,omitempty"`

	// SeriesID 所属系列 ID（可选）
	SeriesID *uuid.UUID `json:"series_id,omitempty"`

	// SeriesOrder 在系列中的排序（可选）
	SeriesOrder int `json:"series_order,omitempty"`

	// IsPaid 是否为付费内容
	IsPaid bool `json:"is_paid,omitempty"`

	// Price 付费价格（当 IsPaid 为 true 时有效）
	Price *float64 `json:"price,omitempty"`

	// TagIDs 关联的标签 ID 列表
	TagIDs []uuid.UUID `json:"tag_ids,omitempty"`
}

// UpdatePostRequest 更新文章请求结构体。
//
// 包含更新文章所需的所有字段。
type UpdatePostRequest struct {
	// Title 文章标题
	Title string `json:"title"`

	// Slug 文章 URL 别名
	Slug string `json:"slug"`

	// Content 文章内容
	Content string `json:"content"`

	// ContentText 文章纯文本内容（可选）
	ContentText string `json:"content_text,omitempty"`

	// Excerpt 文章摘要（可选）
	Excerpt string `json:"excerpt,omitempty"`

	// CategoryID 所属分类 ID（可选）
	CategoryID *uuid.UUID `json:"category_id,omitempty"`

	// Status 文章状态
	Status string `json:"status,omitempty"`

	// FeaturedImage 特色图片 URL（可选）
	FeaturedImage string `json:"featured_image,omitempty"`

	// SEOTitle SEO 标题（可选）
	SEOTitle string `json:"seo_title,omitempty"`

	// SEODescription SEO 描述（可选）
	SEODescription string `json:"seo_description,omitempty"`

	// SEOKeywords SEO 关键词列表（可选）
	SEOKeywords []string `json:"seo_keywords,omitempty"`

	// SeriesID 所属系列 ID（可选）
	SeriesID *uuid.UUID `json:"series_id,omitempty"`

	// SeriesOrder 在系列中的排序
	SeriesOrder int `json:"series_order,omitempty"`

	// IsPaid 是否为付费内容
	IsPaid bool `json:"is_paid"`

	// Price 付费价格
	Price *float64 `json:"price,omitempty"`

	// TagIDs 关联的标签 ID 列表
	TagIDs []uuid.UUID `json:"tag_ids,omitempty"`
}

// PostResponse 文章响应结构体。
//
// 包含文章的完整信息，用于 API 响应。
type PostResponse struct {
	// ID 文章唯一标识符
	ID uuid.UUID `json:"id"`

	// AuthorID 作者 ID
	AuthorID uuid.UUID `json:"author_id"`

	// Title 文章标题
	Title string `json:"title"`

	// Slug 文章 URL 别名
	Slug string `json:"slug"`

	// Content 文章内容
	Content string `json:"content"`

	// Excerpt 文章摘要
	Excerpt string `json:"excerpt,omitempty"`

	// CategoryID 所属分类 ID
	CategoryID *uuid.UUID `json:"category_id,omitempty"`

	// Status 文章状态
	Status string `json:"status"`

	// PublishAt 发布时间
	PublishAt *time.Time `json:"publish_at,omitempty"`

	// FeaturedImage 特色图片 URL
	FeaturedImage string `json:"featured_image,omitempty"`

	// SEOMeta SEO 元数据
	SEOMeta SEOMetaResp `json:"seo_meta,omitempty"`

	// ViewCount 浏览次数
	ViewCount int `json:"view_count"`

	// LikeCount 点赞次数
	LikeCount int `json:"like_count"`

	// CommentCount 评论次数
	CommentCount int `json:"comment_count"`

	// SeriesID 所属系列 ID
	SeriesID *uuid.UUID `json:"series_id,omitempty"`

	// SeriesOrder 在系列中的排序
	SeriesOrder int `json:"series_order"`

	// IsPaid 是否为付费内容
	IsPaid bool `json:"is_paid"`

	// Price 付费价格
	Price *float64 `json:"price,omitempty"`

	// Version 文章版本号
	Version int `json:"version"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// SEOMetaResp SEO 元数据响应结构体。
type SEOMetaResp struct {
	// Title SEO 标题
	Title string `json:"title,omitempty"`

	// Description SEO 描述
	Description string `json:"description,omitempty"`

	// Keywords SEO 关键词
	Keywords []string `json:"keywords,omitempty"`
}

// PostListResponse 文章列表响应结构体。
//
// 包含分页的文章列表和总数。
type PostListResponse struct {
	// Posts 文章列表
	Posts []PostResponse `json:"posts"`

	// Total 文章总数
	Total int `json:"total"`

	// Page 当前页码
	Page int `json:"page"`

	// PageSize 每页数量
	PageSize int `json:"page_size"`
}

// List 文章列表。
//
// 获取文章列表，支持分页和筛选。
// 可以按状态、作者、分类等条件筛选文章。
//
// 路由：GET /api/v1/posts
//
// 查询参数：
//   - page: 页码，默认 1
//   - page_size: 每页数量，默认 20，最大 100
//   - status: 文章状态筛选
//   - author_id: 作者 ID 筛选
//   - category_id: 分类 ID 筛选
//   - search: 搜索关键词
//
// 返回值（通过响应体）：
//   - posts: 文章列表
//   - total: 总数
//   - page: 当前页码
//   - page_size: 每页数量
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

// Get 文章详情。
//
// 根据 slug 获取文章的详细信息。
// 访问文章时会自动增加浏览量。
//
// 路由：GET /api/v1/posts/{slug}
//
// 参数：
//   - slug: 文章 URL 别名（路径参数）
//
// 返回值（通过响应体）：
//   - 文章完整信息
//
// 可能的错误：
//   - BAD_REQUEST: 缺少文章标识
//   - NOT_FOUND: 文章不存在
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

// Create 创建文章。
//
// 创建一篇新文章。需要用户认证。
// 创建后文章默认为草稿状态，需要调用 Publish 接口发布。
//
// 路由：POST /api/v1/posts
//
// 参数（通过请求体）：
//   - title: 文章标题（必填）
//   - slug: URL 别名（必填）
//   - content: 文章内容（必填）
//   - 其他可选字段见 CreatePostRequest
//
// 返回值（通过响应体）：
//   - 新创建的文章信息
//
// 可能的错误：
//   - UNAUTHORIZED: 未登录
//   - BAD_REQUEST: 请求格式错误
//   - INTERNAL_ERROR: 创建失败
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

// Update 更新文章。
//
// 更新已有的文章内容。需要用户认证且为文章作者。
// 更新会自动创建新版本。
//
// 路由：PUT /api/v1/posts/{id}
//
// 参数：
//   - id: 文章 ID（路径参数）
//   - 其他字段通过请求体传递
//
// 返回值（通过响应体）：
//   - 更新后的文章信息
//
// 可能的错误：
//   - BAD_REQUEST: 无效的文章 ID 或请求格式
//   - NOT_FOUND: 文章不存在
//   - INTERNAL_ERROR: 更新失败
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

// Delete 删除文章。
//
// 删除指定的文章。需要用户认证且为文章作者或管理员。
//
// 路由：DELETE /api/v1/posts/{id}
//
// 参数：
//   - id: 文章 ID（路径参数）
//
// 可能的错误：
//   - BAD_REQUEST: 无效的文章 ID
//   - INTERNAL_ERROR: 删除失败
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

// Publish 发布文章。
//
// 将草稿状态的文章发布为公开状态。
// 发布后文章将对外可见。
//
// 路由：POST /api/v1/posts/{id}/publish
//
// 参数：
//   - id: 文章 ID（路径参数）
//
// 返回值（通过响应体）：
//   - message: 发布成功提示
//
// 可能的错误：
//   - BAD_REQUEST: 无效的文章 ID
//   - INTERNAL_ERROR: 发布失败
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

// Search 搜索文章。
//
// 根据关键词搜索文章标题和内容。
// 支持分页查询。
//
// 路由：GET /api/v1/search
//
// 查询参数：
//   - q: 搜索关键词（必填）
//   - page: 页码，默认 1
//   - page_size: 每页数量，默认 20
//
// 返回值（通过响应体）：
//   - posts: 匹配的文章列表
//   - total: 匹配总数
//
// 可能的错误：
//   - BAD_REQUEST: 缺少搜索关键词
//   - INTERNAL_ERROR: 搜索失败
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

// Versions 获取文章版本历史。
//
// 获取指定文章的所有历史版本记录。
// 用于版本对比和回滚操作。
//
// 路由：GET /api/v1/posts/{id}/versions
//
// 参数：
//   - id: 文章 ID（路径参数）
//
// 返回值（通过响应体）：
//   - 版本历史列表
//
// 可能的错误：
//   - BAD_REQUEST: 无效的文章 ID
//   - INTERNAL_ERROR: 获取版本历史失败
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

// RollbackRequest 回滚请求结构体。
type RollbackRequest struct {
	// Version 目标版本号
	Version int `json:"version"`
}

// Rollback 回滚到指定版本。
//
// 将文章内容回滚到指定的历史版本。
// 回滚后会创建一个新版本，保留当前版本历史。
//
// 路由：POST /api/v1/posts/{id}/rollback
//
// 参数：
//   - id: 文章 ID（路径参数）
//   - version: 目标版本号（请求体）
//
// 返回值（通过响应体）：
//   - message: 回滚成功提示
//
// 可能的错误：
//   - BAD_REQUEST: 无效的文章 ID 或版本号
//   - NOT_FOUND: 版本不存在
//   - INTERNAL_ERROR: 回滚失败
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

// GetUserPosts 获取用户的文章列表。
//
// 获取指定用户发布的所有文章。
// 支持按状态筛选和分页。
//
// 路由：GET /api/v1/users/{id}/posts
//
// 参数：
//   - id: 用户 ID（路径参数）
//   - page: 页码，默认 1
//   - limit: 每页数量，默认 20
//   - status: 文章状态筛选（可选）
//
// 返回值（通过响应体）：
//   - posts: 文章列表
//   - total: 总数
//
// 可能的错误：
//   - BAD_REQUEST: 无效的用户 ID 或状态
//   - INTERNAL_ERROR: 获取失败
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

// Like 点赞文章。
//
// 为指定文章点赞。需要用户认证。
// 每个用户只能对同一篇文章点赞一次。
//
// 路由：POST /api/v1/posts/{id}/like
//
// 参数：
//   - id: 文章 ID（路径参数）
//
// 返回值（通过响应体）：
//   - message: 点赞成功提示
//
// 可能的错误：
//   - BAD_REQUEST: 无效的文章 ID
//   - UNAUTHORIZED: 未登录
//   - ALREADY_LIKED: 已点赞过该文章
//   - NOT_FOUND: 文章不存在
//   - INTERNAL_ERROR: 点赞失败
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

// Unlike 取消点赞。
//
// 取消对文章的点赞。需要用户认证。
//
// 路由：DELETE /api/v1/posts/{id}/like
//
// 参数：
//   - id: 文章 ID（路径参数）
//
// 返回值（通过响应体）：
//   - message: 取消点赞成功提示
//
// 可能的错误：
//   - BAD_REQUEST: 无效的文章 ID
//   - UNAUTHORIZED: 未登录
//   - NOT_LIKED: 未点赞过该文章
//   - NOT_FOUND: 文章不存在
//   - INTERNAL_ERROR: 取消点赞失败
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

// toPostResponse 转换 Post 实体到 PostResponse 响应结构。
//
// 将内部的 Post 实体转换为对外暴露的 PostResponse 结构，
// 包含所有必要的字段格式化。
//
// 参数：
//   - p: Post 实体指针
//
// 返回值：
//   - PostResponse: 文章响应结构
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
