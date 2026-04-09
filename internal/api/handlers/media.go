// Package handlers 提供了 Cadmus API 的 HTTP 处理器实现。
//
// 该文件包含媒体文件相关的核心逻辑，包括：
//   - 文件上传和存储
//   - 媒体文件管理（列表、删除）
//   - 图片尺寸信息处理
//
// 主要用途：
//
//	用于处理用户上传的媒体文件管理，支持图片、视频等多种格式。
//
// 注意事项：
//   - 上传需要用户认证
//   - 有文件大小和类型限制
//   - 只能删除自己上传的文件
//
// 作者：xfy
package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/media"
	"rua.plus/cadmus/internal/services"
)

// MediaHandler 媒体 API 处理器。
//
// 该处理器负责处理所有媒体文件相关的 HTTP 请求，包括文件上传、
// 列表查询和删除操作。
type MediaHandler struct {
	// mediaService 媒体服务，处理媒体业务逻辑
	mediaService services.MediaService
}

// NewMediaHandler 创建媒体处理器。
//
// 参数：
//   - mediaService: 媒体服务，处理媒体业务逻辑
//
// 返回值：
//   - *MediaHandler: 新创建的媒体处理器实例
func NewMediaHandler(mediaService services.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

// MediaResponse 媒体响应结构体。
type MediaResponse struct {
	// ID 媒体唯一标识符
	ID uuid.UUID `json:"id"`

	// UploaderID 上传者 ID
	UploaderID uuid.UUID `json:"uploader_id"`

	// Filename 存储文件名
	Filename string `json:"filename"`

	// OriginalName 原始文件名
	OriginalName string `json:"original_name"`

	// URL 访问 URL
	URL string `json:"url"`

	// MimeType MIME 类型
	MimeType string `json:"mime_type"`

	// Size 文件大小（字节）
	Size int64 `json:"size"`

	// Width 图片宽度（可选）
	Width *int `json:"width,omitempty"`

	// Height 图片高度（可选）
	Height *int `json:"height,omitempty"`

	// AltText 替代文本（可选）
	AltText *string `json:"alt_text,omitempty"`

	// CreatedAt 创建时间
	CreatedAt string `json:"created_at"`
}

// MediaListResponse 媒体列表响应结构体。
type MediaListResponse struct {
	// Medias 媒体列表
	Medias []MediaResponse `json:"media"`

	// Total 总数
	Total int `json:"total"`
}

// Upload 上传文件。
//
// 处理文件上传请求，支持图片、视频等多种格式。
// 最大文件大小为 10MB。需要用户认证。
//
// 路由：POST /api/v1/media/upload
//
// 参数：
//   - file: 上传的文件（multipart form）
//   - alt_text: 替代文本（可选）
//
// 返回值（通过响应体）：
//   - 新创建的媒体信息
//
// 可能的错误：
//   - UNAUTHORIZED: 未登录
//   - BAD_REQUEST: 无法解析表单或未找到文件
//   - FILE_TOO_LARGE: 文件大小超过限制
//   - INVALID_FILE_TYPE: 不支持的文件类型
//   - INTERNAL_ERROR: 上传失败
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取当前用户 ID
	userID, err := GetUserID(ctx)
	if err != nil {
		WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
		return
	}

	// 解析 multipart form（最大 10MB）
	err = r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无法解析表单数据", nil, http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "未找到上传文件", nil, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 获取可选的 alt_text
	var altText *string
	if text := r.FormValue("alt_text"); text != "" {
		altText = &text
	}

	// 上传文件
	m, err := h.mediaService.Upload(ctx, userID, header, altText)
	if err != nil {
		if errors.Is(err, media.ErrFileSizeTooLarge) {
			WriteAPIError(w, "FILE_TOO_LARGE", "文件大小超过限制（最大 10MB）", nil, http.StatusBadRequest)
			return
		}
		if errors.Is(err, media.ErrInvalidMimeType) {
			WriteAPIError(w, "INVALID_FILE_TYPE", "不支持的文件类型", nil, http.StatusBadRequest)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "上传失败: "+err.Error(), nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toMediaResponse(m), http.StatusCreated)
}

// Delete 删除文件。
//
// 删除指定的媒体文件。需要用户认证且只能删除自己上传的文件。
//
// 路由：DELETE /api/v1/media/{id}
//
// 参数：
//   - id: 媒体 ID（路径参数）
//
// 可能的错误：
//   - UNAUTHORIZED: 未登录
//   - BAD_REQUEST: 无效的媒体 ID
//   - NOT_FOUND: 媒体文件不存在
//   - PERMISSION_DENIED: 只能删除自己上传的文件
//   - INTERNAL_ERROR: 删除失败
func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的媒体ID", nil, http.StatusBadRequest)
		return
	}

	// 获取当前用户 ID
	userID, err := GetUserID(ctx)
	if err != nil {
		WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
		return
	}

	if err := h.mediaService.Delete(ctx, id, userID); err != nil {
		if errors.Is(err, media.ErrMediaNotFound) {
			WriteAPIError(w, "NOT_FOUND", "媒体文件不存在", nil, http.StatusNotFound)
			return
		}
		if errors.Is(err, media.ErrPermissionDenied) {
			WriteAPIError(w, "PERMISSION_DENIED", "只能删除自己上传的文件", nil, http.StatusForbidden)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "删除失败", nil, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List 获取媒体列表。
//
// 获取媒体文件列表，支持分页和筛选。
// 可以筛选当前用户上传的文件。
//
// 路由：GET /api/v1/media
//
// 查询参数：
//   - offset: 偏移量，默认 0
//   - limit: 数量限制，默认 20，最大 100
//   - mine: 是否只获取当前用户的文件（true/false）
//   - type: MIME 类型筛选
//
// 返回值（通过响应体）：
//   - media: 媒体列表
//   - total: 总数
func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取当前用户 ID（可选，用于筛选）
	userID, _ := GetUserID(ctx)

	// 解析分页参数
	offset := 0
	limit := 20
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := parseInt(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := parseInt(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// 解析筛选条件
	filters := &media.MediaListFilters{}
	if userID != uuid.Nil && r.URL.Query().Get("mine") == "true" {
		filters.UploaderID = userID
	}
	if mimeType := r.URL.Query().Get("type"); mimeType != "" {
		filters.MimeType = mimeType
	}

	medias, total, err := h.mediaService.List(ctx, filters, offset, limit)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取媒体列表失败", nil, http.StatusInternalServerError)
		return
	}

	responses := make([]MediaResponse, 0, len(medias))
	for _, m := range medias {
		responses = append(responses, toMediaResponse(m))
	}

	WriteJSON(w, MediaListResponse{
		Medias: responses,
		Total:  total,
	}, http.StatusOK)
}

// Get 获取单个媒体信息。
//
// 根据媒体 ID 获取详细信息。
//
// 路由：GET /api/v1/media/{id}
//
// 参数：
//   - id: 媒体 ID（路径参数）
//
// 返回值（通过响应体）：
//   - 媒体详细信息
//
// 可能的错误：
//   - BAD_REQUEST: 无效的媒体 ID
//   - NOT_FOUND: 媒体文件不存在
//   - INTERNAL_ERROR: 获取失败
func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "BAD_REQUEST", "无效的媒体ID", nil, http.StatusBadRequest)
		return
	}

	m, err := h.mediaService.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, media.ErrMediaNotFound) {
			WriteAPIError(w, "NOT_FOUND", "媒体文件不存在", nil, http.StatusNotFound)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "获取媒体信息失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toMediaResponse(m), http.StatusOK)
}

// toMediaResponse 转换 Media 实体到 MediaResponse 响应结构。
//
// 参数：
//   - m: Media 实体指针
//
// 返回值：
//   - MediaResponse: 媒体响应结构
func toMediaResponse(m *media.Media) MediaResponse {
	return MediaResponse{
		ID:           m.ID,
		UploaderID:   m.UploaderID,
		Filename:     m.Filename,
		OriginalName: m.OriginalName,
		URL:          m.URL,
		MimeType:     m.MimeType,
		Size:         m.Size,
		Width:        m.Width,
		Height:       m.Height,
		AltText:      m.AltText,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// parseInt 解析整数字符串。
//
// 将字符串转换为整数，支持负数。
// 不使用 strconv.Atoi 以避免导入额外包。
//
// 参数：
//   - s: 要解析的字符串
//
// 返回值：
//   - int: 解析结果
//   - error: 解析失败时的错误
func parseInt(s string) (int, error) {
	var n int
	var negative bool
	for i, c := range s {
		if i == 0 && c == '-' {
			negative = true
			continue
		}
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return 0, errors.New("invalid integer")
		}
	}
	if negative {
		n = -n
	}
	return n, nil
}
