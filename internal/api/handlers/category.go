// Package handlers 提供了 Cadmus API 的 HTTP 处理器实现。
//
// 该文件包含分类相关的核心逻辑，包括：
//   - 分类的 CRUD 操作
//   - 分类文章数量统计
//   - 分类排序
//
// 主要用途：
//
//	用于管理博客文章的分类，支持层级分类结构。
//
// 注意事项：
//   - 创建和更新分类需要管理员权限
//   - 删除分类前需要检查是否有关联文章
//
// 作者：xfy
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/services"
)

// CategoryHandler 分类 API 处理器。
//
// 该处理器负责处理所有分类相关的 HTTP 请求。
type CategoryHandler struct {
	// service 分类服务
	service services.CategoryService
}

// NewCategoryHandler 创建分类处理器。
//
// 参数：
//   - service: 分类服务
//
// 返回值：
//   - *CategoryHandler: 新创建的分类处理器实例
func NewCategoryHandler(service services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// CreateCategoryRequest 创建分类请求结构体。
type CreateCategoryRequest struct {
	// Name 分类名称（必填）
	Name string `json:"name"`

	// Slug URL 别名（必填）
	Slug string `json:"slug"`

	// Description 分类描述（可选）
	Description string `json:"description,omitempty"`

	// ParentID 父分类 ID（可选，用于层级分类）
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// SortOrder 排序顺序
	SortOrder int `json:"sort_order,omitempty"`
}

// UpdateCategoryRequest 更新分类请求结构体。
type UpdateCategoryRequest struct {
	// Name 分类名称
	Name string `json:"name"`

	// Slug URL 别名
	Slug string `json:"slug"`

	// Description 分类描述
	Description string `json:"description,omitempty"`

	// ParentID 父分类 ID
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// SortOrder 排序顺序
	SortOrder int `json:"sort_order,omitempty"`
}

// CategoryResponse 分类响应结构体。
type CategoryResponse struct {
	// ID 分类唯一标识符
	ID uuid.UUID `json:"id"`

	// Name 分类名称
	Name string `json:"name"`

	// Slug URL 别名
	Slug string `json:"slug"`

	// Description 分类描述
	Description string `json:"description,omitempty"`

	// ParentID 父分类 ID
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// SortOrder 排序顺序
	SortOrder int `json:"sort_order"`

	// PostCount 文章数量
	PostCount int `json:"post_count,omitempty"`

	// CreatedAt 创建时间
	CreatedAt string `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt string `json:"updated_at"`
}

// CategoryListResponse 分类列表响应结构体。
type CategoryListResponse struct {
	// Categories 分类列表
	Categories []CategoryResponse `json:"categories"`

	// Total 分类总数
	Total int `json:"total"`
}

// List 分类列表。
//
// 获取所有分类列表，包含每个分类的文章数量。
//
// 路由：GET /api/v1/categories
//
// 返回值（通过响应体）：
//   - categories: 分类列表
//   - total: 分类总数
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	categories, err := h.service.GetAll(ctx)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取分类列表失败", nil, http.StatusInternalServerError)
		return
	}

	// 获取每个分类的文章数
	responses := make([]CategoryResponse, 0, len(categories))
	for _, c := range categories {
		count, _ := h.service.GetPostCount(ctx, c.ID)
		responses = append(responses, toCategoryResponse(c, count))
	}

	WriteJSON(w, CategoryListResponse{
		Categories: responses,
		Total:      len(responses),
	}, http.StatusOK)
}

// Get 分类详情。
//
// 根据 slug 获取分类的详细信息。
//
// 路由：GET /api/v1/categories/{slug}
//
// 参数：
//   - slug: 分类 URL 别名（路径参数）
//
// 返回值（通过响应体）：
//   - 分类详细信息
//
// 可能的错误：
//   - VALIDATION_ERROR: 缺少分类标识
//   - CATEGORY_NOT_FOUND: 分类不存在
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 URL 获取 slug（路由参数）
	slug := r.PathValue("slug")
	if slug == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "缺少分类标识", nil, http.StatusBadRequest)
		return
	}

	category, err := h.service.GetBySlug(ctx, slug)
	if err != nil {
		WriteAPIError(w, "CATEGORY_NOT_FOUND", "分类不存在", nil, http.StatusNotFound)
		return
	}

	count, _ := h.service.GetPostCount(ctx, category.ID)
	WriteJSON(w, toCategoryResponse(category, count), http.StatusOK)
}

// Create 创建分类（需权限）。
//
// 创建新的分类。需要管理员权限。
//
// 路由：POST /api/v1/categories
//
// 参数（通过请求体）：
//   - name: 分类名称（必填）
//   - slug: URL 别名（必填）
//   - description: 描述（可选）
//   - parent_id: 父分类 ID（可选）
//   - sort_order: 排序顺序
//
// 返回值（通过响应体）：
//   - 新创建的分类信息
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

	if err := h.service.Create(ctx, category); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "创建分类失败", []string{err.Error()}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toCategoryResponse(category, 0), http.StatusCreated)
}

// Update 更新分类。
//
// 更新已有分类的信息。需要管理员权限。
//
// 路由：PUT /api/v1/categories/{id}
//
// 参数：
//   - id: 分类 ID（路径参数）
//   - 其他字段通过请求体传递
//
// 返回值（通过响应体）：
//   - 更新后的分类信息
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
	category, err := h.service.GetByID(ctx, id)
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

	if err := h.service.Update(ctx, category); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "更新分类失败", []string{err.Error()}, http.StatusInternalServerError)
		return
	}

	count, _ := h.service.GetPostCount(ctx, category.ID)
	WriteJSON(w, toCategoryResponse(category, count), http.StatusOK)
}

// Delete 删除分类。
//
// 删除指定的分类。需要管理员权限。
// 如果分类下有文章，则无法删除。
//
// 路由：DELETE /api/v1/categories/{id}
//
// 参数：
//   - id: 分类 ID（路径参数）
//
// 返回值（通过响应体）：
//   - message: 删除成功提示
//
// 可能的错误：
//   - VALIDATION_ERROR: 无效的分类 ID
//   - CATEGORY_NOT_FOUND: 分类不存在
//   - DELETE_FAILED: 分类下有文章，无法删除
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

	if err := h.service.Delete(ctx, id); err != nil {
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

// toCategoryResponse 转换 Category 实体到 CategoryResponse 响应结构。
//
// 参数：
//   - c: Category 实体指针
//   - postCount: 文章数量
//
// 返回值：
//   - CategoryResponse: 分类响应结构
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
