// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含权限数据仓库的核心逻辑，包括：
//   - 权限查询（按角色、按类别）
//   - 权限检查（验证角色是否拥有指定权限）
//   - 权限列表获取
//
// 主要用途：
//
//	用于管理权限数据，支持 RBAC 权限模型的权限检查。
//
// 注意事项：
//   - 权限数据通常由系统初始化时插入，不提供 CRUD 操作
//   - 权限按类别分组便于管理
//   - 权限检查使用 EXISTS 子查询提高效率
//
// 作者：xfy
package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/user"
)

// PermissionRepository 权限数据仓库实现。
//
// 负责权限数据的查询和检查操作。
// 权限数据为只读，由系统初始化时设置。
type PermissionRepository struct {
	// pool 数据库连接池
	pool *Pool
}

// NewPermissionRepository 创建权限仓库
func NewPermissionRepository(pool *Pool) *PermissionRepository {
	return &PermissionRepository{pool: pool}
}

// GetByRoleID 获取角色的所有权限
func (r *PermissionRepository) GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]user.Permission, error) {
	query := `
		SELECT p.id, p.name, p.description, p.category, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.category, p.name
	`

	rows, err := r.pool.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by role id: %w", err)
	}
	defer rows.Close()

	perms := make([]user.Permission, 0)
	for rows.Next() {
		var p user.Permission
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// GetAll 获取所有权限
func (r *PermissionRepository) GetAll(ctx context.Context) ([]user.Permission, error) {
	query := `
		SELECT id, name, description, category, created_at
		FROM permissions ORDER BY category, name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all permissions: %w", err)
	}
	defer rows.Close()

	perms := make([]user.Permission, 0)
	for rows.Next() {
		var p user.Permission
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// GetByCategory 获取指定类别的权限
func (r *PermissionRepository) GetByCategory(ctx context.Context, category string) ([]user.Permission, error) {
	query := `
		SELECT id, name, description, category, created_at
		FROM permissions WHERE category = $1 ORDER BY name
	`

	rows, err := r.pool.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by category: %w", err)
	}
	defer rows.Close()

	perms := make([]user.Permission, 0)
	for rows.Next() {
		var p user.Permission
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// CheckPermission 检查角色是否拥有指定权限
func (r *PermissionRepository) CheckPermission(ctx context.Context, roleID uuid.UUID, permissionName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM role_permissions rp
			INNER JOIN permissions p ON rp.permission_id = p.id
			WHERE rp.role_id = $1 AND p.name = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, roleID, permissionName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return exists, nil
}