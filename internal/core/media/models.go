// Package media 媒体文件管理模块
package media

import (
	"time"

	"github.com/google/uuid"
)

// Media 媒体文件实体
type Media struct {
	ID           uuid.UUID `json:"id"`
	UploaderID   uuid.UUID `json:"uploader_id"`
	Filename     string    `json:"filename"`      // 存储文件名（唯一）
	OriginalName string    `json:"original_name"` // 原始文件名
	FilePath     string    `json:"filepath"`      // 文件存储路径
	URL          string    `json:"url"`           // 访问 URL
	MimeType     string    `json:"mime_type"`     // MIME 类型
	Size         int64     `json:"size"`          // 文件大小（字节）
	Width        *int      `json:"width"`         // 图片宽度（可选）
	Height       *int      `json:"height"`        // 图片高度（可选）
	AltText      *string   `json:"alt_text"`      // 替代文本（可选）
	CreatedAt    time.Time `json:"created_at"`
}

// UploadInput 上传文件输入
type UploadInput struct {
	UploaderID   uuid.UUID
	OriginalName string
	MimeType     string
	Size         int64
	AltText      *string
}

// MediaListFilters 媒体列表筛选条件
type MediaListFilters struct {
	UploaderID uuid.UUID
	MimeType   string // 可按类型筛选（如 "image/", "application/"）
}

// 常见错误定义
var (
	ErrMediaNotFound      = &MediaError{Code: "media_not_found", Message: "媒体文件不存在"}
	ErrInvalidMimeType    = &MediaError{Code: "invalid_mime_type", Message: "不支持的文件类型"}
	ErrFileSizeTooLarge   = &MediaError{Code: "file_size_too_large", Message: "文件大小超过限制"}
	ErrPermissionDenied   = &MediaError{Code: "permission_denied", Message: "权限不足"}
	ErrUploadFailed       = &MediaError{Code: "upload_failed", Message: "上传失败"}
)

// MediaError 媒体模块自定义错误
type MediaError struct {
	Code    string
	Message string
}

func (e *MediaError) Error() string {
	return e.Message
}

// Is 实现 errors.Is 接口
func (e *MediaError) Is(target error) bool {
	t, ok := target.(*MediaError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// AllowedMimeTypes 允许上传的 MIME 类型
var AllowedMimeTypes = map[string]bool{
	// 图片类型
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/svg+xml": true,
	// 文档类型
	"application/pdf": true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	// 其他
	"application/zip": true,
	"text/plain": true,
}

// IsImageMimeType 判断是否为图片类型
func IsImageMimeType(mimeType string) bool {
	return mimeType == "image/jpeg" ||
		mimeType == "image/png" ||
		mimeType == "image/gif" ||
		mimeType == "image/webp" ||
		mimeType == "image/svg+xml"
}