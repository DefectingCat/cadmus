// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含点赞功能的通用仓库实现，支持文章点赞和评论点赞。
//
// 主要用途：
//
//	通过配置化的方式实现点赞功能的复用，避免代码重复。
//
// 作者：xfy
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LikeConfig 点赞仓库配置。
//
// 通过配置指定表名、目标列名和计数器表名，
// 使同一个 LikeRepository 实现可以用于不同的点赞场景。
type LikeConfig struct {
	// TableName 点赞表名（如 "post_likes", "comment_likes"）
	TableName string
	// TargetColumn 目标列名（如 "post_id", "comment_id"）
	TargetColumn string
	// CounterTable 计数器表名（如 "posts", "comments"）
	CounterTable string
}

// BaseLikeRepository 通用点赞仓库。
//
// 提供点赞的原子操作，包括幂等创建、幂等删除、存在检查和计数。
// 通过 LikeConfig 配置不同场景的具体表和列名。
type BaseLikeRepository struct {
	pool   *Pool
	config LikeConfig
}

// NewBaseLikeRepository 创建通用点赞仓库。
//
// 参数：
//   - pool: 数据库连接池
//   - config: 点赞配置
//
// 返回值：
//   - *BaseLikeRepository: 点赞仓库实例
func NewBaseLikeRepository(pool *Pool, config LikeConfig) *BaseLikeRepository {
	return &BaseLikeRepository{pool: pool, config: config}
}

// CreateIfNotExists 创建点赞记录（幂等），返回是否实际创建。
//
// 使用 ON CONFLICT DO NOTHING 确保幂等性，
// 同时原子更新目标对象的点赞计数。
//
// 参数：
//   - ctx: 上下文
//   - targetID: 目标对象 ID（文章 ID 或评论 ID）
//   - userID: 用户 ID
//
// 返回值：
//   - created: true 表示实际创建了记录（新点赞），false 表示已存在
//   - err: 操作错误
func (r *BaseLikeRepository) CreateIfNotExists(ctx context.Context, targetID, userID uuid.UUID) (created bool, err error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, %s, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (%s, user_id) DO NOTHING
	`, r.config.TableName, r.config.TargetColumn, r.config.TargetColumn)

	id := uuid.New()
	now := time.Now()

	result, err := r.pool.Exec(ctx, query, id, targetID, userID, now)
	if err != nil {
		return false, fmt.Errorf("failed to create like: %w", err)
	}

	created = result.RowsAffected() > 0

	// 只有实际创建点赞记录时才更新计数
	if created {
		updateQuery := fmt.Sprintf(
			"UPDATE %s SET like_count = like_count + 1 WHERE id = $1",
			r.config.CounterTable,
		)
		_, err = r.pool.Exec(ctx, updateQuery, targetID)
		if err != nil {
			return false, fmt.Errorf("failed to update like count: %w", err)
		}
	}

	return created, nil
}

// DeleteIfExists 删除点赞记录（幂等），返回是否实际删除。
//
// 同时原子更新目标对象的点赞计数。
//
// 参数：
//   - ctx: 上下文
//   - targetID: 目标对象 ID
//   - userID: 用户 ID
//
// 返回值：
//   - deleted: true 表示实际删除了记录，false 表示不存在
//   - err: 操作错误
func (r *BaseLikeRepository) DeleteIfExists(ctx context.Context, targetID, userID uuid.UUID) (deleted bool, err error) {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1 AND user_id = $2",
		r.config.TableName, r.config.TargetColumn,
	)

	result, err := r.pool.Exec(ctx, query, targetID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to delete like: %w", err)
	}

	deleted = result.RowsAffected() > 0

	// 只有实际删除点赞记录时才更新计数
	if deleted {
		updateQuery := fmt.Sprintf(
			"UPDATE %s SET like_count = like_count - 1 WHERE id = $1 AND like_count > 0",
			r.config.CounterTable,
		)
		_, err = r.pool.Exec(ctx, updateQuery, targetID)
		if err != nil {
			return false, fmt.Errorf("failed to update like count: %w", err)
		}
	}

	return deleted, nil
}

// Exists 检查用户是否已点赞。
//
// 参数：
//   - ctx: 上下文
//   - targetID: 目标对象 ID
//   - userID: 用户 ID
//
// 返回值：
//   - exists: true 表示已点赞
//   - err: 查询错误
func (r *BaseLikeRepository) Exists(ctx context.Context, targetID, userID uuid.UUID) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1 AND user_id = $2)",
		r.config.TableName, r.config.TargetColumn,
	)

	var exists bool
	err := r.pool.QueryRow(ctx, query, targetID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check like exists: %w", err)
	}
	return exists, nil
}

// CountByTarget 统计目标对象的点赞数。
//
// 参数：
//   - ctx: 上下文
//   - targetID: 目标对象 ID
//
// 返回值：
//   - count: 点赞数量
//   - err: 查询错误
func (r *BaseLikeRepository) CountByTarget(ctx context.Context, targetID uuid.UUID) (int, error) {
	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE %s = $1",
		r.config.TableName, r.config.TargetColumn,
	)

	var count int
	err := r.pool.QueryRow(ctx, query, targetID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count likes: %w", err)
	}
	return count, nil
}

// PostLikeConfig 返回文章点赞配置。
func PostLikeConfig() LikeConfig {
	return LikeConfig{
		TableName:    "post_likes",
		TargetColumn: "post_id",
		CounterTable: "posts",
	}
}

// CommentLikeConfig 返回评论点赞配置。
func CommentLikeConfig() LikeConfig {
	return LikeConfig{
		TableName:    "comment_likes",
		TargetColumn: "comment_id",
		CounterTable: "comments",
	}
}
