package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/database"
)

// CategoryHandler 分类 API 处理器
type CategoryHandler struct {
	repo *database.CategoryRepository
}

// NewCategoryHandler 创建分类处理器
func NewCategoryHandler(repo *database.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	SortOrder   int        `json:"sort_order,omitempty"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	SortOrder   int        `json:"sort_order,omitempty"`
}

// CategoryResponse 分类响应
type CategoryResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	SortOrder   int        `json:"sort_order"`
	PostCount   int        `json:"post_count,omitempty"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

// CategoryListResponse 分类列表响应
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Total      int                `json:"total"`
}

// List 分类列表
// GET /api/v1/categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	categories, err := h.repo.GetAll(ctx)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取分类列表失败", nil, http.StatusInternalServerError)
		return
	}

	// 获取每个分类的文章数
	responses := make([]CategoryResponse, 0, len(categories))
	for _, c := range categories {
		count, _ := h.repo.GetPostCount(ctx, c.ID)
		responses = append(responses, toCategoryResponse(c, count))
	}

	WriteJSON(w, CategoryListResponse{
		Categories: responses,
		Total:      len(responses),
	}, http.StatusOK)
}

// Get 分类详情
// GET /api/v1/categories/:slug
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 URL 获取 slug（路由参数）
	slug := r.PathValue("slug")
	if slug == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "缺少分类标识", nil, http.StatusBadRequest)
		return
	}

	category, err := h.repo.GetBySlug(ctx, slug)
	if err != nil {
		WriteAPIError(w, "CATEGORY_NOT_FOUND", "分类不存在", nil, http.StatusNotFound)
		return
	}

	count, _ := h.repo.GetPostCount(ctx, category.ID)
	WriteJSON(w, toCategoryResponse(category, count), http.StatusOK)
}

// Create 创建分类（需权限）
// POST /api/v1/categories
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Slug == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "分类名称和 slug 为必填项", nil, http.StatusBadRequest)
		return
	}

	// 创建分类实体
	category := &post.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
	}

	if err := h.repo.Create(ctx, category); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "创建分类失败", []string{err.Error()}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toCategoryResponse(category, 0), http.StatusCreated)
}

// Update 更新分类
// PUT /api/v1/categories/:id
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 URL 获取 id
	idStr := r.PathValue("id")
	if idStr == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "缺少分类 ID", nil, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "VALIDATION_ERROR", "无效的分类 ID", nil, http.StatusBadRequest)
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Slug == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "分类名称和 slug 为必填项", nil, http.StatusBadRequest)
		return
	}

	// 获取现有分类
	category, err := h.repo.GetByID(ctx, id)
	if err != nil {
		WriteAPIError(w, "CATEGORY_NOT_FOUND", "分类不存在", nil, http.StatusNotFound)
		return
	}

	// 更新字段
	category.Name = req.Name
	category.Slug = req.Slug
	category.Description = req.Description
	category.ParentID = req.ParentID
	category.SortOrder = req.SortOrder

	if err := h.repo.Update(ctx, category); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "更新分类失败", []string{err.Error()}, http.StatusInternalServerError)
		return
	}

	count, _ := h.repo.GetPostCount(ctx, category.ID)
	WriteJSON(w, toCategoryResponse(category, count), http.StatusOK)
}

// Delete 删除分类
// DELETE /api/v1/categories/:id
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 URL 获取 id
	idStr := r.PathValue("id")
	if idStr == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "缺少分类 ID", nil, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "VALIDATION_ERROR", "无效的分类 ID", nil, http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		if err == post.ErrCategoryNotFound {
			WriteAPIError(w, "CATEGORY_NOT_FOUND", "分类不存在", nil, http.StatusNotFound)
			return
		}
		// 检查是否是业务限制错误
		WriteAPIError(w, "DELETE_FAILED", err.Error(), nil, http.StatusBadRequest)
		return
	}

	WriteJSON(w, map[string]string{"message": "分类已删除"}, http.StatusOK)
}

// toCategoryResponse 转换 Category 到 CategoryResponse
func toCategoryResponse(c *post.Category, postCount int) CategoryResponse {
	return CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		ParentID:    c.ParentID,
		SortOrder:   c.SortOrder,
		PostCount:   postCount,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}