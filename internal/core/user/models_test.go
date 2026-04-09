package user

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   UserStatus
		expected bool
	}{
		{StatusActive, true},
		{StatusBanned, true},
		{StatusPending, true},
		{UserStatus("invalid"), false},
		{UserStatus(""), false},
		{UserStatus("ACTIVE"), false}, // 大小写敏感
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	t.Run("sets password hash", func(t *testing.T) {
		u := &User{}
		err := u.SetPassword("myPassword123")

		assert.NoError(t, err)
		assert.NotEmpty(t, u.PasswordHash)
		assert.NotEqual(t, "myPassword123", u.PasswordHash) // 应该是哈希值
	})

	t.Run("sets different hashes for different passwords", func(t *testing.T) {
		u1 := &User{}
		u2 := &User{}
		u1.SetPassword("password1")
		u2.SetPassword("password2")

		assert.NotEqual(t, u1.PasswordHash, u2.PasswordHash)
	})

	t.Run("handles empty password", func(t *testing.T) {
		u := &User{}
		err := u.SetPassword("")

		assert.NoError(t, err)
		assert.NotEmpty(t, u.PasswordHash) // bcrypt 会哈希空密码
	})
}

func TestUser_CheckPassword(t *testing.T) {
	t.Run("returns true for correct password", func(t *testing.T) {
		u := &User{}
		u.SetPassword("correctPassword")

		assert.True(t, u.CheckPassword("correctPassword"))
	})

	t.Run("returns false for incorrect password", func(t *testing.T) {
		u := &User{}
		u.SetPassword("correctPassword")

		assert.False(t, u.CheckPassword("wrongPassword"))
	})

	t.Run("returns false for empty password hash", func(t *testing.T) {
		u := &User{PasswordHash: ""}

		assert.False(t, u.CheckPassword("anyPassword"))
	})

	t.Run("is case sensitive", func(t *testing.T) {
		u := &User{}
		u.SetPassword("Password")

		assert.False(t, u.CheckPassword("password"))
		assert.False(t, u.CheckPassword("PASSWORD"))
		assert.True(t, u.CheckPassword("Password"))
	})
}

func TestUserError_Error(t *testing.T) {
	t.Run("returns message", func(t *testing.T) {
		err := &UserError{Code: "test_code", Message: "test message"}
		assert.Equal(t, "test message", err.Error())
	})
}

func TestUserError_Is(t *testing.T) {
	t.Run("returns true for same code", func(t *testing.T) {
		err1 := &UserError{Code: "user_not_found", Message: "用户不存在"}
		err2 := &UserError{Code: "user_not_found", Message: "different message"}

		assert.True(t, errors.Is(err1, err2))
	})

	t.Run("returns false for different code", func(t *testing.T) {
		err1 := &UserError{Code: "user_not_found", Message: "用户不存在"}
		err2 := &UserError{Code: "user_already_exists", Message: "用户已存在"}

		assert.False(t, errors.Is(err1, err2))
	})

	t.Run("returns false for different error type", func(t *testing.T) {
		err := &UserError{Code: "test", Message: "test"}
		stdErr := errors.New("standard error")

		assert.False(t, errors.Is(err, stdErr))
	})

	t.Run("works with predefined errors", func(t *testing.T) {
		assert.True(t, errors.Is(ErrUserNotFound, ErrUserNotFound))
		assert.False(t, errors.Is(ErrUserNotFound, ErrUserAlreadyExists))
	})
}
