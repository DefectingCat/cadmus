package media

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *MediaError
		expected string
	}{
		{"media not found", ErrMediaNotFound, "媒体文件不存在"},
		{"invalid mime type", ErrInvalidMimeType, "不支持的文件类型"},
		{"file size too large", ErrFileSizeTooLarge, "文件大小超过限制"},
		{"permission denied", ErrPermissionDenied, "权限不足"},
		{"upload failed", ErrUploadFailed, "上传失败"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestMediaError_Is(t *testing.T) {
	tests := []struct {
		name     string
		err      *MediaError
		target   error
		expected bool
	}{
		{
			name:     "same error code",
			err:      ErrMediaNotFound,
			target:   ErrMediaNotFound,
			expected: true,
		},
		{
			name:     "different error code",
			err:      ErrMediaNotFound,
			target:   ErrInvalidMimeType,
			expected: false,
		},
		{
			name:     "non-MediaError target",
			err:      ErrMediaNotFound,
			target:   errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil target",
			err:      ErrMediaNotFound,
			target:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMediaError_Code(t *testing.T) {
	tests := []struct {
		name         string
		err          *MediaError
		expectedCode string
	}{
		{"media_not_found code", ErrMediaNotFound, "media_not_found"},
		{"invalid_mime_type code", ErrInvalidMimeType, "invalid_mime_type"},
		{"file_size_too_large code", ErrFileSizeTooLarge, "file_size_too_large"},
		{"permission_denied code", ErrPermissionDenied, "permission_denied"},
		{"upload_failed code", ErrUploadFailed, "upload_failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedCode, tt.err.Code)
		})
	}
}

func TestIsImageMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected bool
	}{
		// Image types - should return true
		{"jpeg", "image/jpeg", true},
		{"png", "image/png", true},
		{"gif", "image/gif", true},
		{"webp", "image/webp", true},
		{"svg", "image/svg+xml", true},

		// Non-image types - should return false
		{"pdf", "application/pdf", false},
		{"doc", "application/msword", false},
		{"docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", false},
		{"zip", "application/zip", false},
		{"text", "text/plain", false},
		{"html", "text/html", false},
		{"json", "application/json", false},
		{"empty", "", false},
		{"unknown", "unknown/type", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsImageMimeType(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllowedMimeTypes(t *testing.T) {
	// Test that all image MIME types in AllowedMimeTypes are recognized by IsImageMimeType
	imageTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}

	for _, mt := range imageTypes {
		t.Run("allowed_image_"+mt, func(t *testing.T) {
			assert.True(t, AllowedMimeTypes[mt], "MIME type %s should be in AllowedMimeTypes", mt)
			assert.True(t, IsImageMimeType(mt), "MIME type %s should be recognized as image", mt)
		})
	}

	// Test that document types are allowed but not images
	docTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/zip",
		"text/plain",
	}

	for _, mt := range docTypes {
		t.Run("allowed_doc_"+mt, func(t *testing.T) {
			assert.True(t, AllowedMimeTypes[mt], "MIME type %s should be in AllowedMimeTypes", mt)
			assert.False(t, IsImageMimeType(mt), "MIME type %s should not be recognized as image", mt)
		})
	}

	// Test that disallowed types are not in the map
	disallowedTypes := []string{
		"application/exe",
		"application/x-sh",
		"video/mp4",
	}

	for _, mt := range disallowedTypes {
		t.Run("disallowed_"+mt, func(t *testing.T) {
			assert.False(t, AllowedMimeTypes[mt], "MIME type %s should not be in AllowedMimeTypes", mt)
		})
	}
}
