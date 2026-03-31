package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/database"
)

// TagHandler 标签 API 处理器
type TagHandler struct {
	repo *database.TagRepository
}

// NewTagHandler 创建标签处理器
func NewTagHandler(repo *database.TagRepository) *TagHandler {
	return &TagHandler{repo: repo}
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// TagResponse 标签响应
type TagResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	PostCount int       `json:"post_count,omitempty"`
	CreatedAt string    `json:"created_at"`
}

// TagListResponse 标签列表响应
type TagListResponse struct {
	Tags  []TagResponse `json:"tags"`
	Total int           `json:"total"`
}

// List 标签列表
// GET /api/v1/tags
func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tags, err := h.repo.GetAll(ctx)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取标签列表失败", nil, http.StatusInternalServerError)
		return
	}

	// 获取每个标签的文章数
	responses := make([]TagResponse, 0, len(tags))
	for _, t := range tags {
		count, _ := h.repo.GetPostCount(ctx, t.ID)
		responses = append(responses, toTagResponse(t, count))
	}

	WriteJSON(w, TagListResponse{
		Tags:  responses,
		Total: len(responses),
	}, http.StatusOK)
}

// Get 标签详情
// GET /api/v1/tags/:slug
func (h *TagHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 URL 获取 slug（路由参数）
	slug := r.PathValue("slug")
	if slug == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "缺少标签标识", nil, http.StatusBadRequest)
		return
	}

	tag, err := h.repo.GetBySlug(ctx, slug)
	if err != nil {
		WriteAPIError(w, "TAG_NOT_FOUND", "标签不存在", nil, http.StatusNotFound)
		return
	}

	count, _ := h.repo.GetPostCount(ctx, tag.ID)
	WriteJSON(w, toTagResponse(tag, count), http.StatusOK)
}

// Create 创建标签（需权限）
// POST /api/v1/tags
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

	if err := h.repo.Create(ctx, tag); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "创建标签失败", []string{err.Error()}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toTagResponse(tag, 0), http.StatusCreated)
}

// Delete 删除标签
// DELETE /api/v1/tags/:id
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

	if err := h.repo.Delete(ctx, id); err != nil {
		if err == post.ErrTagNotFound {
			WriteAPIError(w, "TAG_NOT_FOUND", "标签不存在", nil, http.StatusNotFound)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "删除标签失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "标签已删除"}, http.StatusOK)
}

// toTagResponse 转换 Tag 到 TagResponse
func toTagResponse(t *post.Tag, postCount int) TagResponse {
	return TagResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		PostCount: postCount,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}