// Package user 提供用户、角色、权限管理的核心数据模型。
//
// 该文件包含用户系统的核心实体定义，包括：
//   - 用户实体及其状态枚举
//   - 角色实体
//   - 权限实体
//   - 密码加密相关方法
//   - 语义化错误类型定义
//
// 主要用途：
//
//	用于博客系统的用户管理，支持多角色、细粒度权限控制。
//
// 注意事项：
//   - UserStatus 枚举通过 IsValid() 方法验证状态有效性
//   - 密码使用 bcrypt 算法加密存储，安全性高
//   - 所有实体使用 UUID 作为主键
//
// 作者：xfy
package user

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserStatus 用户状态枚举。
//
// 定义用户在系统中的生命周期状态，控制用户的登录权限和可见性。
type UserStatus string

// 用户状态常量定义。
const (
	// StatusActive 正常活跃用户，可以正常登录和使用系统
	StatusActive UserStatus = "active"

	// StatusBanned 已封禁用户，禁止登录，保留数据用于审计
	StatusBanned UserStatus = "banned"

	// StatusPending 待激活用户，已注册但未完成邮箱验证等激活流程
	StatusPending UserStatus = "pending"
)

// IsValid 检查用户状态是否有效。
//
// 验证状态值是否为预定义的三种状态之一。
//
// 返回值：
//   - true: 状态有效
//   - false: 状态无效
func (s UserStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusBanned, StatusPending:
		return true
	default:
		return false
	}
}

// User 用户实体。
//
// 表示博客系统中的用户账户，包含基本信息、角色关联和状态。
// 密码使用 bcrypt 算法加密存储，不暴露密码哈希值。
//
// 注意事项：
//   - ID 由系统自动生成，无需手动设置
//   - PasswordHash 字段使用 json:"-" 标签，不会序列化到 JSON
//   - CreatedAt/UpdatedAt 使用 UTC 时间
type User struct {
	// ID 用户的唯一标识符，格式为 UUID
	ID uuid.UUID `json:"id"`

	// Username 用户名，用于登录和 URL 展示
	Username string `json:"username"`

	// Email 邮箱地址，用于登录和通知
	Email string `json:"email"`

	// PasswordHash 密码哈希值，使用 bcrypt 加密，不暴露给外部
	PasswordHash string `json:"-"`

	// AvatarURL 头像图片 URL，可选
	AvatarURL string `json:"avatar_url,omitempty"`

	// Bio 个人简介，可选
	Bio string `json:"bio,omitempty"`

	// RoleID 关联的角色 ID，决定用户权限
	RoleID uuid.UUID `json:"role_id"`

	// Status 用户状态，控制登录权限
	Status UserStatus `json:"status"`

	// CreatedAt 创建时间，使用 UTC 时间戳
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 最后修改时间，使用 UTC 时间戳
	UpdatedAt time.Time `json:"updated_at"`
}

// Role 角色实体。
//
// 表示用户的角色类型，如管理员、编辑、作者、订阅者等。
// 每个角色关联一组权限，形成基于角色的访问控制（RBAC）体系。
type Role struct {
	// ID 角色的唯一标识符
	ID uuid.UUID `json:"id"`

	// Name 角色内部名称，如 "admin"、"editor"、"subscriber"
	Name string `json:"name"`

	// DisplayName 角色显示名称，如 "管理员"、"编辑"、"订阅者"
	DisplayName string `json:"display_name"`

	// Permissions 角色拥有的权限列表，可选加载
	Permissions []Permission `json:"permissions,omitempty"`

	// IsDefault 是否为默认角色，新注册用户自动分配默认角色
	IsDefault bool `json:"is_default"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// Permission 权限实体。
//
// 表示系统中的细粒度权限，如创建文章、删除评论等。
// 权限通过角色间接授予用户，支持灵活的权限组合。
type Permission struct {
	// ID 权限的唯一标识符
	ID uuid.UUID `json:"id"`

	// Name 权限名称，采用 "资源.操作" 格式
	// 如 "post.create"、"post.edit"、"comment.delete"
	Name string `json:"name"`

	// Description 权限描述，说明该权限的作用
	Description string `json:"description"`

	// Category 权限分类，用于组织权限列表
	// 如 "post"、"comment"、"user"、"theme"、"plugin"
	Category string `json:"category"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// 常见错误定义。
//
// 使用语义化错误类型，便于调用方进行错误处理和判断。
var (
	// ErrUserNotFound 用户不存在错误，查询用户时 ID 或条件无效时返回
	ErrUserNotFound = &UserError{Code: "user_not_found", Message: "用户不存在"}

	// ErrUserAlreadyExists 用户已存在错误，创建用户时用户名或邮箱冲突时返回
	ErrUserAlreadyExists = &UserError{Code: "user_already_exists", Message: "用户已存在"}

	// ErrInvalidCredentials 无效凭证错误，登录时用户名或密码错误时返回
	ErrInvalidCredentials = &UserError{Code: "invalid_credentials", Message: "无效的凭证"}

	// ErrInvalidStatus 无效用户状态错误，设置状态值不合法时返回
	ErrInvalidStatus = &UserError{Code: "invalid_status", Message: "无效的用户状态"}

	// ErrRoleNotFound 角色不存在错误，关联的角色 ID 无效时返回
	ErrRoleNotFound = &UserError{Code: "role_not_found", Message: "角色不存在"}

	// ErrPermissionDenied 权限不足错误，用户无权执行操作时返回
	ErrPermissionDenied = &UserError{Code: "permission_denied", Message: "权限不足"}
)

// UserError 用户模块自定义错误类型。
//
// 实现 error 和 errors.Is 接口，支持错误比较和类型判断。
// 通过 Code 字段区分不同错误类型，Message 字段提供人类可读描述。
type UserError struct {
	// Code 错误代码，用于程序化错误判断
	Code string

	// Message 错误消息，用于展示给用户或记录日志
	Message string
}

// Error 实现 error 接口，返回错误消息。
func (e *UserError) Error() string {
	return e.Message
}

// Is 实现 errors.Is 接口，支持错误类型比较。
//
// 通过比较 Code 字段判断是否为同类型错误，
// 便于使用 errors.Is(err, ErrUserNotFound) 进行错误判断。
func (e *UserError) Is(target error) bool {
	t, ok := target.(*UserError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// SetPassword 设置用户密码。
//
// 将明文密码使用 bcrypt 算法加密后存储到 PasswordHash 字段。
// bcrypt 是一种自适应哈希算法，具有盐值防彩虹表攻击。
//
// 参数：
//   - password: 明文密码，建议至少 8 个字符
//
// 返回值：
//   - err: 加密失败时返回错误（通常为内存不足等极端情况）
//
// 使用示例：
//   err := user.SetPassword("mySecurePassword123")
//   if err != nil {
//       return err
//   }
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword 验证用户密码。
//
// 比较输入的明文密码与存储的哈希值是否匹配。
// 用于登录验证等场景。
//
// 参数：
//   - password: 待验证的明文密码
//
// 返回值：
//   - true: 密码匹配
//   - false: 密码不匹配
//
// 使用示例：
//   if user.CheckPassword(inputPassword) {
//       // 登录成功
//   } else {
//       // 密码错误
//   }
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}