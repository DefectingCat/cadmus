// Package media 提供媒体文件管理的核心数据模型。
//
// 该文件包含媒体系统的核心实体定义，包括：
//   - Media: 媒体文件实体
//   - UploadInput: 上传文件输入结构
//   - MediaListFilters: 媒体列表筛选条件
//   - 允许上传的 MIME 类型定义
//   - 语义化错误类型定义
//
// 主要用途：
//
//	用于博客系统的媒体文件管理，支持图片、文档等文件的上传和存储。
//
// 注意事项：
//   - 文件大小和类型由业务层校验
//   - 图片类文件可存储尺寸信息用于前端展示
//   - 所有实体使用 UUID 作为主键
//
// 作者：xfy
package media

import (
	"time"

	"github.com/google/uuid"
)

// Media 媒体文件实体。
//
// 表示系统中上传的媒体文件，包含文件信息和元数据。
// 支持图片、文档等多种文件类型。
//
// 注意事项：
//   - ID 由系统自动生成，无需手动设置
//   - Filename 是存储在服务器上的唯一文件名
//   - OriginalName 保留用户上传时的原始文件名
//   - CreatedAt 使用 UTC 时间
type Media struct {
	// ID 媒体文件的唯一标识符，格式为 UUID
	ID uuid.UUID `json:"id"`

	// UploaderID 上传者的用户 ID
	UploaderID uuid.UUID `json:"uploader_id"`

	// Filename 存储文件名，系统生成的唯一名称
	// 通常使用 UUID 或时间戳+随机数生成，避免文件名冲突
	Filename string `json:"filename"`

	// OriginalName 原始文件名，保留用户上传时的文件名
	// 用于下载时显示友好名称
	OriginalName string `json:"original_name"`

	// FilePath 文件存储路径，相对于存储根目录
	// 如 "uploads/2024/01/xxx.jpg"
	FilePath string `json:"filepath"`

	// URL 文件访问 URL，完整的可访问地址
	// 如 "https://cdn.example.com/uploads/2024/01/xxx.jpg"
	URL string `json:"url"`

	// MimeType 文件的 MIME 类型
	// 如 "image/jpeg"、"application/pdf"
	MimeType string `json:"mime_type"`

	// Size 文件大小，单位为字节
	Size int64 `json:"size"`

	// Width 图片宽度（像素），仅图片类型有效
	// 非图片文件为 nil
	Width *int `json:"width"`

	// Height 图片高度（像素），仅图片类型有效
	// 非图片文件为 nil
	Height *int `json:"height"`

	// AltText 替代文本，用于图片无障碍访问
	// 显示在图片无法加载时的占位文本
	AltText *string `json:"alt_text"`

	// CreatedAt 上传时间，使用 UTC 时间戳
	CreatedAt time.Time `json:"created_at"`
}

// UploadInput 上传文件输入结构。
//
// 用于接收上传文件请求的参数，包含文件的基本信息。
type UploadInput struct {
	// UploaderID 上传者的用户 ID
	UploaderID uuid.UUID

	// OriginalName 原始文件名
	OriginalName string

	// MimeType 文件的 MIME 类型
	MimeType string

	// Size 文件大小（字节）
	Size int64

	// AltText 替代文本（可选）
	AltText *string
}

// MediaListFilters 媒体列表筛选条件。
//
// 用于构建媒体查询的过滤条件，支持多条件组合。
// 所有字段均为可选，未设置时忽略该条件。
type MediaListFilters struct {
	// UploaderID 按上传者筛选，获取指定用户上传的文件
	UploaderID uuid.UUID

	// MimeType 按 MIME 类型筛选，支持前缀匹配
	// 如 "image/" 筛选所有图片，"application/" 筛选所有文档
	MimeType string
}

// 常见错误定义。
//
// 使用语义化错误类型，便于调用方进行错误处理和判断。
var (
	// ErrMediaNotFound 媒体文件不存在错误，查询文件时 ID 无效时返回
	ErrMediaNotFound = &MediaError{Code: "media_not_found", Message: "媒体文件不存在"}

	// ErrInvalidMimeType 无效 MIME 类型错误，上传不允许的文件类型时返回
	ErrInvalidMimeType = &MediaError{Code: "invalid_mime_type", Message: "不支持的文件类型"}

	// ErrFileSizeTooLarge 文件大小超限错误，上传文件超过系统限制时返回
	ErrFileSizeTooLarge = &MediaError{Code: "file_size_too_large", Message: "文件大小超过限制"}

	// ErrPermissionDenied 权限不足错误，用户无权操作文件时返回
	ErrPermissionDenied = &MediaError{Code: "permission_denied", Message: "权限不足"}

	// ErrUploadFailed 上传失败错误，文件上传过程中出错时返回
	ErrUploadFailed = &MediaError{Code: "upload_failed", Message: "上传失败"}
)

// MediaError 媒体模块自定义错误类型。
//
// 实现 error 和 errors.Is 接口，支持错误比较和类型判断。
// 通过 Code 字段区分不同错误类型，Message 字段提供人类可读描述。
type MediaError struct {
	// Code 错误代码，用于程序化错误判断
	Code string

	// Message 错误消息，用于展示给用户或记录日志
	Message string
}

// Error 实现 error 接口，返回错误消息。
func (e *MediaError) Error() string {
	return e.Message
}

// Is 实现 errors.Is 接口，支持错误类型比较。
//
// 通过比较 Code 字段判断是否为同类型错误，
// 便于使用 errors.Is(err, ErrMediaNotFound) 进行错误判断。
func (e *MediaError) Is(target error) bool {
	t, ok := target.(*MediaError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// AllowedMimeTypes 允许上传的 MIME 类型白名单。
//
// 定义系统支持上传的文件类型，不在列表中的类型将被拒绝。
// 包含常见图片类型、文档类型和压缩文件类型。
var AllowedMimeTypes = map[string]bool{
	// 图片类型
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,

	// 文档类型
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,

	// 其他类型
	"application/zip": true,
	"text/plain":      true,
}

// IsImageMimeType 判断是否为图片 MIME 类型。
//
// 检查给定的 MIME 类型是否属于图片类别。
// 用于判断是否需要提取图片尺寸信息。
//
// 参数：
//   - mimeType: MIME 类型字符串
//
// 返回值：
//   - true: 是图片类型
//   - false: 不是图片类型
//
// 使用示例：
//
//	if IsImageMimeType(file.Header.Get("Content-Type")) {
//	    // 提取图片尺寸
//	}
func IsImageMimeType(mimeType string) bool {
	return mimeType == "image/jpeg" ||
		mimeType == "image/png" ||
		mimeType == "image/gif" ||
		mimeType == "image/webp" ||
		mimeType == "image/svg+xml"
}
