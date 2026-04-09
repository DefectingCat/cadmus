package comment

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommentStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   CommentStatus
		expected bool
	}{
		{"pending status", StatusPending, true},
		{"approved status", StatusApproved, true},
		{"spam status", StatusSpam, true},
		{"deleted status", StatusDeleted, true},
		{"invalid status", CommentStatus("invalid"), false},
		{"empty status", CommentStatus(""), false},
		{"unknown status", CommentStatus("archived"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommentError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CommentError
		expected string
	}{
		{"comment not found", ErrCommentNotFound, "评论不存在"},
		{"comment already exists", ErrCommentAlreadyExists, "评论已存在"},
		{"invalid status", ErrInvalidStatus, "无效的评论状态"},
		{"max depth exceeded", ErrMaxDepthExceeded, "评论嵌套深度超过限制"},
		{"parent not found", ErrParentNotFound, "父评论不存在"},
		{"permission denied", ErrPermissionDenied, "权限不足"},
		{"empty content", ErrEmptyContent, "评论内容不能为空"},
		{"post not found", ErrPostNotFound, "文章不存在"},
		{"user not found", ErrUserNotFound, "用户不存在"},
		{"already liked", ErrAlreadyLiked, "已点赞该评论"},
		{"not liked", ErrNotLiked, "未点赞该评论"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestCommentError_Is(t *testing.T) {
	tests := []struct {
		name     string
		err      *CommentError
		target   error
		expected bool
	}{
		{
			name:     "same error code",
			err:      ErrCommentNotFound,
			target:   ErrCommentNotFound,
			expected: true,
		},
		{
			name:     "different error code",
			err:      ErrCommentNotFound,
			target:   ErrCommentAlreadyExists,
			expected: false,
		},
		{
			name:     "non-CommentError target",
			err:      ErrCommentNotFound,
			target:   errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil target",
			err:      ErrCommentNotFound,
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

func TestCommentError_Code(t *testing.T) {
	tests := []struct {
		name         string
		err          *CommentError
		expectedCode string
	}{
		{"comment_not_found code", ErrCommentNotFound, "comment_not_found"},
		{"comment_already_exists code", ErrCommentAlreadyExists, "comment_already_exists"},
		{"invalid_status code", ErrInvalidStatus, "invalid_status"},
		{"max_depth_exceeded code", ErrMaxDepthExceeded, "max_depth_exceeded"},
		{"parent_not_found code", ErrParentNotFound, "parent_not_found"},
		{"permission_denied code", ErrPermissionDenied, "permission_denied"},
		{"empty_content code", ErrEmptyContent, "empty_content"},
		{"post_not_found code", ErrPostNotFound, "post_not_found"},
		{"user_not_found code", ErrUserNotFound, "user_not_found"},
		{"already_liked code", ErrAlreadyLiked, "already_liked"},
		{"not_liked code", ErrNotLiked, "not_liked"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedCode, tt.err.Code)
		})
	}
}
