// Package handlers 提供了 Cadmus API 的 HTTP 处理器实现。
//
// 该文件包含标签相关的核心逻辑，包括：
//   - 标签的 CRUD 操作
//   - 标签文章数量统计
//
// 主要用途：
//
//	用于管理博客文章的标签，支持文章多标签关联。
//
// 注意事项：
//   - 创建和删除标签需要管理员权限
//   - 标签不能重名
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

// TagHandler 标签 API 处理器。
//
// 该处理器负责处理所有标签相关的 HTTP 请求。
type TagHandler struct {
	// service 标签服务
	service services.TagService
}

// NewTagHandler 创建标签处理器。
//
// 参数：
//   - service: 标签服务
//
// 返回值：
//   - *TagHandler: 新创建的标签处理器实例
func NewTagHandler(service services.TagService) *TagHandler {
	return &TagHandler{service: service}
}

// CreateTagRequest 创建标签请求结构体。
type CreateTagRequest struct {
	// Name 标签名称（必填）
	Name string `json:"name"`

	// Slug URL 别名（必填）
	Slug string `json:"slug"`
}

// TagResponse 标签响应结构体。
type TagResponse struct {
	// ID 标签唯一标识符
	ID uuid.UUID `json:"id"`

	// Name 标签名称
	Name string `json:"name"`

	// Slug URL 别名
	Slug string `json:"slug"`

	// PostCount 文章数量
	PostCount int `json:"post_count,omitempty"`

	// CreatedAt 创建时间
	CreatedAt string `json:"created_at"`
}

// TagListResponse 标签列表响应结构体。
type TagListResponse struct {
	// Tags 标签列表
	Tags []TagResponse `json:"tags"`

	// Total 标签总数
	Total int `json:"total"`
}

// List 标签列表。
//
// 获取所有标签列表，包含每个标签的文章数量。
//
// 路由：GET /api/v1/tags
//
// 返回值（通过响应体）：
//   - tags: 标签列表
//   - total: 标签总数
func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tags, err := h.service.GetAll(ctx)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取标签列表失败", nil, http.StatusInternalServerError)
		return
	}

	// 获取每个标签的文章数
	responses := make([]TagResponse, 0, len(tags))
	for _, t := range tags {
		count, _ := h.service.GetPostCount(ctx, t.ID)
		responses = append(responses, toTagResponse(t, count))
	}

	WriteJSON(w, TagListResponse{
		Tags:  responses,
		Total: len(responses),
	}, http.StatusOK)
}

// Get 标签详情。
//
// 根据 slug 获取标签的详细信息。
//
// 路由：GET /api/v1/tags/{slug}
//
// 参数：
//   - slug: 标签 URL 别名（路径参数）
//
// 返回值（通过响应体）：
//   - 标签详细信息
//
// 可能的错误：
//   - VALIDATION_ERROR: 缺少标签标识
//   - TAG_NOT_FOUND: 标签不存在
func (h *TagHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 URL 获取 slug（路由参数）
	slug := r.PathValue("slug")
	if slug == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "缺少标签标识", nil, http.StatusBadRequest)
		return
	}

	tag, err := h.service.GetBySlug(ctx, slug)
	if err != nil {
		WriteAPIError(w, "TAG_NOT_FOUND", "标签不存在", nil, http.StatusNotFound)
		return
	}

	count, _ := h.service.GetPostCount(ctx, tag.ID)
	WriteJSON(w, toTagResponse(tag, count), http.StatusOK)
}

// Create 创建标签（需权限）。
//
// 创建新的标签。需要管理员权限。
//
// 路由：POST /api/v1/tags
//
// 参数（通过请求体）：
//   - name: 标签名称（必填）
//   - slug: URL 别名（必填）
//
// 返回值（通过响应体）：
//   - 新创建的标签信息
func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Slug == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "标签名称和 slug 为必填项", nil, http.StatusBadRequest)
		return
	}

	// 创建标签实体
	tag := &post.Tag{
		Name: req.Name,
		Slug: req.Slug,
	}

	if err := h.service.Create(ctx, tag); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "创建标签失败", []string{err.Error()}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toTagResponse(tag, 0), http.StatusCreated)
}

// Delete 删除标签。
//
// 删除指定的标签。需要管理员权限。
// 删除标签不会删除关联的文章，只是解除关联。
//
// 路由：DELETE /api/v1/tags/{id}
//
// 参数：
//   - id: 标签 ID（路径参数）
//
// 返回值（通过响应体）：
//   - message: 删除成功提示
//
// 可能的错误：
//   - VALIDATION_ERROR: 无效的标签 ID
//   - TAG_NOT_FOUND: 标签不存在
//   - INTERNAL_ERROR: 删除失败
func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 URL 获取 id
	idStr := r.PathValue("id")
	if idStr == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "缺少标签 ID", nil, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "VALIDATION_ERROR", "无效的标签 ID", nil, http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		if err == post.ErrTagNotFound {
			WriteAPIError(w, "TAG_NOT_FOUND", "标签不存在", nil, http.StatusNotFound)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "删除标签失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "标签已删除"}, http.StatusOK)
}

// toTagResponse 转换 Tag 实体到 TagResponse 响应结构。
//
// 参数：
//   - t: Tag 实体指针
//   - postCount: 文章数量
//
// 返回值：
//   - TagResponse: 标签响应结构
func toTagResponse(t *post.Tag, postCount int) TagResponse {
	return TagResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		PostCount: postCount,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
