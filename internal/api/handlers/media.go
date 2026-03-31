package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/media"
	"rua.plus/cadmus/internal/services"
)

// MediaHandler 媒体 API 处理器
type MediaHandler struct {
	mediaService services.MediaService
}

// NewMediaHandler 创建媒体处理器
func NewMediaHandler(mediaService services.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

// MediaResponse 媒体响应
type MediaResponse struct {
	ID           uuid.UUID `json:"id"`
	UploaderID   uuid.UUID `json:"uploader_id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	URL          string    `json:"url"`
	MimeType     string    `json:"mime_type"`
	Size         int64     `json:"size"`
	Width        *int      `json:"width,omitempty"`
	Height       *int      `json:"height,omitempty"`
	AltText      *string   `json:"alt_text,omitempty"`
	CreatedAt    string    `json:"created_at"`
}

// MediaListResponse 媒体列表响应
type MediaListResponse struct {
	Medias []MediaResponse `json:"media"`
	Total  int             `json:"total"`
}

// Upload 上传文件
// POST /api/v1/media/upload
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

// Delete 删除文件
// DELETE /api/v1/media/{id}
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

// List 获取媒体列表
// GET /api/v1/media
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

// Get 获取单个媒体信息
// GET /api/v1/media/{id}
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

// toMediaResponse 转换媒体为响应格式
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

// parseInt 解析整数
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