// Package user 用户、角色、权限管理模块
package user

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	// Create 创建新用户
	Create(ctx context.Context, user *User) error

	// GetByID 根据 ID 获取用户
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Update 更新用户信息
	Update(ctx context.Context, user *User) error

	// Delete 删除用户
	Delete(ctx context.Context, id uuid.UUID) error

	// List 分页获取用户列表
	List(ctx context.Context, offset, limit int) ([]*User, int, error)
}

// RoleRepository 角色数据访问接口
type RoleRepository interface {
	// GetByID 根据 ID 获取角色（不含权限列表）
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)

	// GetByName 根据名称获取角色（含权限列表）
	GetByName(ctx context.Context, name string) (*Role, error)

	// GetAll 获取所有角色
	GetAll(ctx context.Context) ([]*Role, error)

	// GetDefault 获取默认角色
	GetDefault(ctx context.Context) (*Role, error)

	// GetWithPermissions 根据ID获取角色及其权限
	GetWithPermissions(ctx context.Context, id uuid.UUID) (*Role, error)
}

// PermissionRepository 权限数据访问接口
type PermissionRepository interface {
	// GetByRoleID 获取角色的所有权限
	GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]Permission, error)

	// GetAll 获取所有权限
	GetAll(ctx context.Context) ([]Permission, error)

	// GetByCategory 获取指定类别的权限
	GetByCategory(ctx context.Context, category string) ([]Permission, error)

	// CheckPermission 检查角色是否拥有指定权限
	CheckPermission(ctx context.Context, roleID uuid.UUID, permissionName string) (bool, error)
}