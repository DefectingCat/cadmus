package post

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   PostStatus
		expected bool
	}{
		{"draft status", StatusDraft, true},
		{"published status", StatusPublished, true},
		{"scheduled status", StatusScheduled, true},
		{"private status", StatusPrivate, true},
		{"invalid status", PostStatus("invalid"), false},
		{"empty status", PostStatus(""), false},
		{"unknown status", PostStatus("archived"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPostError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *PostError
		expected string
	}{
		{"post not found", ErrPostNotFound, "文章不存在"},
		{"post already exists", ErrPostAlreadyExists, "文章已存在"},
		{"invalid status", ErrInvalidStatus, "无效的文章状态"},
		{"category not found", ErrCategoryNotFound, "分类不存在"},
		{"tag not found", ErrTagNotFound, "标签不存在"},
		{"series not found", ErrSeriesNotFound, "文章系列不存在"},
		{"version not found", ErrVersionNotFound, "版本不存在"},
		{"permission denied", ErrPermissionDenied, "权限不足"},
		{"paid content", ErrPaidContent, "此为付费内容，请先购买"},
		{"already liked", ErrAlreadyLiked, "已点赞过该文章"},
		{"not liked", ErrNotLiked, "未点赞过该文章"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestPostError_Is(t *testing.T) {
	tests := []struct {
		name     string
		err      *PostError
		target   error
		expected bool
	}{
		{
			name:     "same error code",
			err:      ErrPostNotFound,
			target:   ErrPostNotFound,
			expected: true,
		},
		{
			name:     "different error code",
			err:      ErrPostNotFound,
			target:   ErrPostAlreadyExists,
			expected: false,
		},
		{
			name:     "non-PostError target",
			err:      ErrPostNotFound,
			target:   errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil target",
			err:      ErrPostNotFound,
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

func TestPostError_Code(t *testing.T) {
	tests := []struct {
		name         string
		err          *PostError
		expectedCode string
	}{
		{"post_not_found code", ErrPostNotFound, "post_not_found"},
		{"post_already_exists code", ErrPostAlreadyExists, "post_already_exists"},
		{"invalid_status code", ErrInvalidStatus, "invalid_status"},
		{"category_not_found code", ErrCategoryNotFound, "category_not_found"},
		{"tag_not_found code", ErrTagNotFound, "tag_not_found"},
		{"series_not_found code", ErrSeriesNotFound, "series_not_found"},
		{"version_not_found code", ErrVersionNotFound, "version_not_found"},
		{"permission_denied code", ErrPermissionDenied, "permission_denied"},
		{"paid_content code", ErrPaidContent, "paid_content"},
		{"already_liked code", ErrAlreadyLiked, "already_liked"},
		{"not_liked code", ErrNotLiked, "not_liked"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedCode, tt.err.Code)
		})
	}
}