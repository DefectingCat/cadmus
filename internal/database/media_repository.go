// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含媒体数据仓库的核心逻辑，包括：
//   - 媒体文件 CRUD 操作（创建、查询、删除）
//   - 按上传者查询媒体列表
//   - 媒体列表分页和筛选
//   - 媒体统计计数
//
// 主要用途：
//
//	用于管理用户上传的媒体文件元数据，支持图片、视频等多种类型。
//
// 注意事项：
//   - 媒体删除为硬删除，需确保文件已从存储中删除
//   - 支持按 MIME 类型筛选（如 image/%）
//   - 图片宽高信息可选（非图片类型为 nil）
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
	"rua.plus/cadmus/internal/core/media"
)

// MediaRepository 媒体数据仓库实现。
//
// 负责媒体文件元数据的 CRUD 操作，支持按上传者和类型筛选。
// 所有操作通过连接池执行，确保高效的数据访问。
type MediaRepository struct {
	// pool 数据库连接池
	pool *Pool
}

// NewMediaRepository 创建媒体仓库
func NewMediaRepository(pool *Pool) *MediaRepository {
	return &MediaRepository{pool: pool}
}

// Create 创建媒体记录
func (r *MediaRepository) Create(ctx context.Context, input *media.UploadInput, filename, filepath, url string, width, height *int) (*media.Media, error) {
	query := `
		INSERT INTO media (
			id, uploader_id, filename, original_name, filepath, url, mime_type, size, width, height, alt_text, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	id := uuid.New()
	now := time.Now()

	_, err := r.pool.Exec(ctx, query,
		id,
		input.UploaderID,
		filename,
		input.OriginalName,
		filepath,
		url,
		input.MimeType,
		input.Size,
		width,
		height,
		input.AltText,
		now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create media: %w", err)
	}

	return &media.Media{
		ID:           id,
		UploaderID:   input.UploaderID,
		Filename:     filename,
		OriginalName: input.OriginalName,
		FilePath:     filepath,
		URL:          url,
		MimeType:     input.MimeType,
		Size:         input.Size,
		Width:        width,
		Height:       height,
		AltText:      input.AltText,
		CreatedAt:    now,
	}, nil
}

// GetByID 根据 ID 获取媒体
func (r *MediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*media.Media, error) {
	query := `
		SELECT id, uploader_id, filename, original_name, filepath, url, mime_type, size, width, height, alt_text, created_at
		FROM media WHERE id = $1
	`

	m, err := r.scanMedia(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, media.ErrMediaNotFound
		}
		return nil, fmt.Errorf("failed to get media by id: %w", err)
	}
	return m, nil
}

// GetByUploaderID 获取用户上传的所有媒体
func (r *MediaRepository) GetByUploaderID(ctx context.Context, uploaderID uuid.UUID) ([]*media.Media, error) {
	query := `
		SELECT id, uploader_id, filename, original_name, filepath, url, mime_type, size, width, height, alt_text, created_at
		FROM media WHERE uploader_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, uploaderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get media by uploader id: %w", err)
	}
	defer rows.Close()

	return r.scanMedias(rows)
}

// Delete 删除媒体记录
func (r *MediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM media WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}

	if result.RowsAffected() == 0 {
		return media.ErrMediaNotFound
	}

	return nil
}

// List 分页获取媒体列表
func (r *MediaRepository) List(ctx context.Context, filters *media.MediaListFilters, offset, limit int) ([]*media.Media, error) {
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)
	argIndex := 1

	if filters != nil {
		if filters.UploaderID != uuid.Nil {
			whereClause += fmt.Sprintf(" AND uploader_id = $%d", argIndex)
			args = append(args, filters.UploaderID)
			argIndex++
		}
		if filters.MimeType != "" {
			whereClause += fmt.Sprintf(" AND mime_type LIKE $%d", argIndex)
			args = append(args, filters.MimeType+"%")
			argIndex++
		}
	}

	query := fmt.Sprintf(`
		SELECT id, uploader_id, filename, original_name, filepath, url, mime_type, size, width, height, alt_text, created_at
		FROM media %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list media: %w", err)
	}
	defer rows.Close()

	return r.scanMedias(rows)
}

// Count 统计媒体数量
func (r *MediaRepository) Count(ctx context.Context, filters *media.MediaListFilters) (int, error) {
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)
	argIndex := 1

	if filters != nil {
		if filters.UploaderID != uuid.Nil {
			whereClause += fmt.Sprintf(" AND uploader_id = $%d", argIndex)
			args = append(args, filters.UploaderID)
			argIndex++
		}
		if filters.MimeType != "" {
			whereClause += fmt.Sprintf(" AND mime_type LIKE $%d", argIndex)
			args = append(args, filters.MimeType+"%")
			argIndex++
		}
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM media %s`, whereClause)

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count media: %w", err)
	}

	return count, nil
}

// scanMedia 扫描单行媒体数据
func (r *MediaRepository) scanMedia(row pgx.Row) (*media.Media, error) {
	m := &media.Media{}
	err := row.Scan(
		&m.ID,
		&m.UploaderID,
		&m.Filename,
		&m.OriginalName,
		&m.FilePath,
		&m.URL,
		&m.MimeType,
		&m.Size,
		&m.Width,
		&m.Height,
		&m.AltText,
		&m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// scanMedias 扫描多行媒体数据
func (r *MediaRepository) scanMedias(rows pgx.Rows) ([]*media.Media, error) {
	medias := make([]*media.Media, 0)
	for rows.Next() {
		m := &media.Media{}
		err := rows.Scan(
			&m.ID,
			&m.UploaderID,
			&m.Filename,
			&m.OriginalName,
			&m.FilePath,
			&m.URL,
			&m.MimeType,
			&m.Size,
			&m.Width,
			&m.Height,
			&m.AltText,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media: %w", err)
		}
		medias = append(medias, m)
	}
	return medias, nil
}