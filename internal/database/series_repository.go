// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含系列数据仓库的核心逻辑，包括：
//   - 系列 CRUD 操作（创建、查询、更新、删除）
//   - 多种查询方式（ID、Slug、作者）
//   - 系列文章统计
//
// 主要用途：
//
//	用于管理文章系列数据，支持将相关文章组织成系列。
//
// 注意事项：
//   - 系列 Slug 有唯一约束
//   - 删除系列时文章的 series_id 会被置空（ON DELETE SET NULL）
//   - 系列支持封面图片和描述信息
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
	"rua.plus/cadmus/internal/core/post"
)

// SeriesRepository 系列数据仓库实现。
//
// 负责系列数据的 CRUD 操作，支持按作者查询和文章统计。
// 所有操作通过连接池执行，确保高效的数据访问。
type SeriesRepository struct {
	// pool 数据库连接池
	pool *Pool
}

// NewSeriesRepository 创建系列仓库
func NewSeriesRepository(pool *Pool) *SeriesRepository {
	return &SeriesRepository{pool: pool}
}

// Create 创建新系列
func (r *SeriesRepository) Create(ctx context.Context, series *post.Series) error {
	query := `
		INSERT INTO series (id, author_id, title, slug, description, cover_image, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	if series.ID == uuid.Nil {
		series.ID = uuid.New()
	}

	now := time.Now()
	if series.CreatedAt.IsZero() {
		series.CreatedAt = now
	}
	if series.UpdatedAt.IsZero() {
		series.UpdatedAt = now
	}

	_, err := r.pool.Exec(ctx, query,
		series.ID,
		series.AuthorID,
		series.Title,
		series.Slug,
		series.Description,
		series.CoverImage,
		series.CreatedAt,
		series.UpdatedAt,
	)

	if err != nil {
		if IsUniqueViolation(err, "series_slug_key") {
			return post.ErrSeriesNotFound // slug 已存在，使用适当的错误
		}
		return fmt.Errorf("failed to create series: %w", err)
	}

	return nil
}

// Update 更新系列
func (r *SeriesRepository) Update(ctx context.Context, series *post.Series) error {
	query := `
		UPDATE series SET title = $2, slug = $3, description = $4, cover_image = $5
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		series.ID,
		series.Title,
		series.Slug,
		series.Description,
		series.CoverImage,
	)

	if err != nil {
		if IsUniqueViolation(err, "series_slug_key") {
			return post.ErrSeriesNotFound
		}
		return fmt.Errorf("failed to update series: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrSeriesNotFound
	}

	return nil
}

// Delete 删除系列（posts 表的 series_id 会被置空，由数据库 ON DELETE SET NULL 处理）
func (r *SeriesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM series WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete series: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrSeriesNotFound
	}

	return nil
}

// GetByID 根据 ID 获取系列
func (r *SeriesRepository) GetByID(ctx context.Context, id uuid.UUID) (*post.Series, error) {
	query := `
		SELECT id, author_id, title, slug, description, cover_image, created_at, updated_at
		FROM series WHERE id = $1
	`

	series, err := r.scanSeries(ctx, r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrSeriesNotFound
		}
		return nil, fmt.Errorf("failed to get series by id: %w", err)
	}
	return series, nil
}

// GetBySlug 根据 Slug 获取系列
func (r *SeriesRepository) GetBySlug(ctx context.Context, slug string) (*post.Series, error) {
	query := `
		SELECT id, author_id, title, slug, description, cover_image, created_at, updated_at
		FROM series WHERE slug = $1
	`

	series, err := r.scanSeries(ctx, r.pool.QueryRow(ctx, query, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrSeriesNotFound
		}
		return nil, fmt.Errorf("failed to get series by slug: %w", err)
	}
	return series, nil
}

// GetByAuthor 获取作者的系列列表
func (r *SeriesRepository) GetByAuthor(ctx context.Context, authorID uuid.UUID) ([]*post.Series, error) {
	query := `
		SELECT id, author_id, title, slug, description, cover_image, created_at, updated_at
		FROM series WHERE author_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get series by author: %w", err)
	}
	defer rows.Close()

	seriesList := make([]*post.Series, 0)
	for rows.Next() {
		series, err := r.scanSeriesFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan series: %w", err)
		}
		seriesList = append(seriesList, series)
	}

	return seriesList, nil
}

// GetAll 获取所有系列
func (r *SeriesRepository) GetAll(ctx context.Context) ([]*post.Series, error) {
	query := `
		SELECT id, author_id, title, slug, description, cover_image, created_at, updated_at
		FROM series ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all series: %w", err)
	}
	defer rows.Close()

	seriesList := make([]*post.Series, 0)
	for rows.Next() {
		series, err := r.scanSeriesFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan series: %w", err)
		}
		seriesList = append(seriesList, series)
	}

	return seriesList, nil
}

// GetPostCount 统计系列下文章数
func (r *SeriesRepository) GetPostCount(ctx context.Context, seriesID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM posts WHERE series_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, seriesID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get post count: %w", err)
	}

	return count, nil
}

// scanSeries 扫描单行系列数据
func (r *SeriesRepository) scanSeries(ctx context.Context, row pgx.Row) (*post.Series, error) {
	series := &post.Series{}
	err := row.Scan(
		&series.ID,
		&series.AuthorID,
		&series.Title,
		&series.Slug,
		&series.Description,
		&series.CoverImage,
		&series.CreatedAt,
		&series.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return series, nil
}

// scanSeriesFromRow 从 rows 扫描系列数据
func (r *SeriesRepository) scanSeriesFromRow(row pgx.Rows) (*post.Series, error) {
	series := &post.Series{}
	err := row.Scan(
		&series.ID,
		&series.AuthorID,
		&series.Title,
		&series.Slug,
		&series.Description,
		&series.CoverImage,
		&series.CreatedAt,
		&series.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return series, nil
}
