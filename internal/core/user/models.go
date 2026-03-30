// Package user 用户、角色、权限管理模块
package user

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserStatus 用户状态枚举
type UserStatus string

const (
	StatusActive   UserStatus = "active"   // 正常活跃用户
	StatusBanned   UserStatus = "banned"   // 已封禁用户
	StatusPending  UserStatus = "pending"  // 待激活用户
)

// IsValid 检查用户状态是否有效
func (s UserStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusBanned, StatusPending:
		return true
	default:
		return false
	}
}

// User 用户实体
type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // 不暴露密码哈希
	AvatarURL    string     `json:"avatar_url,omitempty"`
	Bio          string     `json:"bio,omitempty"`
	RoleID       uuid.UUID  `json:"role_id"`
	Status       UserStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Role 角色实体
type Role struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name"`
	Permissions []Permission `json:"permissions,omitempty"`
	IsDefault   bool       `json:"is_default"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Permission 权限实体
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`        // post.create, post.edit, comment.delete 等
	Description string    `json:"description"`
	Category    string    `json:"category"`    // post/comment/user/theme/plugin 等
	CreatedAt   time.Time `json:"created_at"`
}

// 常见错误定义
var (
	ErrUserNotFound      = &UserError{Code: "user_not_found", Message: "用户不存在"}
	ErrUserAlreadyExists = &UserError{Code: "user_already_exists", Message: "用户已存在"}
	ErrInvalidCredentials = &UserError{Code: "invalid_credentials", Message: "无效的凭证"}
	ErrInvalidStatus     = &UserError{Code: "invalid_status", Message: "无效的用户状态"}
	ErrRoleNotFound      = &UserError{Code: "role_not_found", Message: "角色不存在"}
	ErrPermissionDenied  = &UserError{Code: "permission_denied", Message: "权限不足"}
)

// UserError 用户模块自定义错误
type UserError struct {
	Code    string
	Message string
}

func (e *UserError) Error() string {
	return e.Message
}

// Is 实现 errors.Is 接口
func (e *UserError) Is(target error) bool {
	t, ok := target.(*UserError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// SetPassword 设置密码（哈希存储）
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}