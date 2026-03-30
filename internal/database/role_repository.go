package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"rua.plus/cadmus/internal/core/user"
)

// RoleRepository 角色数据仓库实现
type RoleRepository struct {
	pool *Pool
}

// NewRoleRepository 创建角色仓库
func NewRoleRepository(pool *Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

// GetByID 根据 ID 获取角色（不含权限列表）
func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.Role, error) {
	query := `
		SELECT id, name, display_name, is_default, created_at
		FROM roles WHERE id = $1
	`

	role, err := r.scanRole(ctx, r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role by id: %w", err)
	}
	return role, nil
}

// GetByName 根据名称获取角色（含权限列表）
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*user.Role, error) {
	query := `
		SELECT id, name, display_name, is_default, created_at
		FROM roles WHERE name = $1
	`

	role, err := r.scanRole(ctx, r.pool.QueryRow(ctx, query, name))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	// 加载权限
	perms, err := r.loadPermissions(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load role permissions: %w", err)
	}
	role.Permissions = perms

	return role, nil
}

// GetAll 获取所有角色
func (r *RoleRepository) GetAll(ctx context.Context) ([]*user.Role, error) {
	query := `
		SELECT id, name, display_name, is_default, created_at
		FROM roles ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*user.Role, 0)
	for rows.Next() {
		role, err := r.scanRoleFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetDefault 获取默认角色
func (r *RoleRepository) GetDefault(ctx context.Context) (*user.Role, error) {
	query := `
		SELECT id, name, display_name, is_default, created_at
		FROM roles WHERE is_default = true LIMIT 1
	`

	role, err := r.scanRole(ctx, r.pool.QueryRow(ctx, query))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get default role: %w", err)
	}
	return role, nil
}

// GetWithPermissions 根据ID获取角色及其权限
func (r *RoleRepository) GetWithPermissions(ctx context.Context, id uuid.UUID) (*user.Role, error) {
	role, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	perms, err := r.loadPermissions(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load role permissions: %w", err)
	}
	role.Permissions = perms

	return role, nil
}

// loadPermissions 加载角色的权限列表
func (r *RoleRepository) loadPermissions(ctx context.Context, roleID uuid.UUID) ([]user.Permission, error) {
	query := `
		SELECT p.id, p.name, p.description, p.category, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.category, p.name
	`

	rows, err := r.pool.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query role permissions: %w", err)
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

// scanRole 扫描单行角色数据
func (r *RoleRepository) scanRole(ctx context.Context, row pgx.Row) (*user.Role, error) {
	role := &user.Role{}
	err := row.Scan(
		&role.ID,
		&role.Name,
		&role.DisplayName,
		&role.IsDefault,
		&role.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return role, nil
}

// scanRoleFromRow 从 rows 扫描角色数据
func (r *RoleRepository) scanRoleFromRow(row pgx.Rows) (*user.Role, error) {
	role := &user.Role{}
	err := row.Scan(
		&role.ID,
		&role.Name,
		&role.DisplayName,
		&role.IsDefault,
		&role.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return role, nil
}