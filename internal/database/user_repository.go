// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含用户数据仓库的核心逻辑，包括：
//   - 用户 CRUD 操作（创建、查询、更新、删除）
//   - 多种查询方式（ID、邮箱、用户名）
//   - 用户列表分页查询
//   - 唯一约束冲突检测
//
// 主要用途：
//
//	用于管理用户数据的持久化存储，提供服务层与数据库之间的数据访问层。
//
// 注意事项：
//   - 用户删除采用软删除（设置 banned 状态）
//   - 用户名和邮箱有唯一约束
//   - 所有方法都需要传入有效的上下文
//
// 作者：xfy
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"rua.plus/cadmus/internal/core/user"
)

// UserRepository 用户数据仓库实现。
//
// 负责用户数据的 CRUD 操作，支持多种查询方式和分页查询。
// 所有操作通过连接池执行，确保高效的数据访问。
type UserRepository struct {
	// pool 数据库连接池
	pool *Pool
}

// NewUserRepository 创建用户仓库
func NewUserRepository(pool *Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create 创建新用户
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, avatar_url, bio, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	now := u.CreatedAt
	if now.IsZero() {
		now = u.UpdatedAt
	}
	if now.IsZero() {
		now = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		u.ID,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.AvatarURL,
		u.Bio,
		u.RoleID,
		u.Status,
		now,
		now,
	)

	if err != nil {
		// 检查唯一约束冲突
		if IsUniqueViolation(err, "users_username_key") {
			return user.ErrUserAlreadyExists
		}
		if IsUniqueViolation(err, "users_email_key") {
			return user.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	u.CreatedAt = now
	u.UpdatedAt = now
	return nil
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar_url, bio, role_id, status, created_at, updated_at
		FROM users WHERE id = $1
	`

	u, err := r.scanUser(ctx, r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return u, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar_url, bio, role_id, status, created_at, updated_at
		FROM users WHERE email = $1
	`

	u, err := r.scanUser(ctx, r.pool.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return u, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar_url, bio, role_id, status, created_at, updated_at
		FROM users WHERE username = $1
	`

	u, err := r.scanUser(ctx, r.pool.QueryRow(ctx, query, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return u, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users SET username = $2, email = $3, password_hash = $4, avatar_url = $5, bio = $6, role_id = $7, status = $8
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		u.ID,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.AvatarURL,
		u.Bio,
		u.RoleID,
		u.Status,
	)

	if err != nil {
		if IsUniqueViolation(err, "users_username_key") || IsUniqueViolation(err, "users_email_key") {
			return user.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// Delete 删除用户（设置为 banned 状态）
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET status = 'banned' WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// List 分页获取用户列表
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, int, error) {
	// 获取总数
	countQuery := `SELECT COUNT(*) FROM users`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 获取列表
	query := `
		SELECT id, username, email, password_hash, avatar_url, bio, role_id, status, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]*user.User, 0)
	for rows.Next() {
		u, err := r.scanUserFromRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, total, nil
}

// scanUser 扫描单行用户数据
func (r *UserRepository) scanUser(ctx context.Context, row pgx.Row) (*user.User, error) {
	u := &user.User{}
	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.AvatarURL,
		&u.Bio,
		&u.RoleID,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// scanUserFromRow 从 rows 扫描用户数据
func (r *UserRepository) scanUserFromRow(row pgx.Rows) (*user.User, error) {
	u := &user.User{}
	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.AvatarURL,
		&u.Bio,
		&u.RoleID,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}