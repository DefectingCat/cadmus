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

// TagRepository 标签数据仓库实现
type TagRepository struct {
	pool *Pool
}

// NewTagRepository 创建标签仓库
func NewTagRepository(pool *Pool) *TagRepository {
	return &TagRepository{pool: pool}
}

// Create 创建新标签
func (r *TagRepository) Create(ctx context.Context, tag *post.Tag) error {
	query := `
		INSERT INTO tags (id, name, slug, created_at)
		VALUES ($1, $2, $3, $4)
	`

	if tag.ID == uuid.Nil {
		tag.ID = uuid.New()
	}

	now := tag.CreatedAt
	if now.IsZero() {
		now = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		tag.ID,
		tag.Name,
		tag.Slug,
		now,
	)

	if err != nil {
		// 检查唯一约束冲突
		if isUniqueViolation(err, "tags_name_key") {
			return fmt.Errorf("tag name already exists: %w", post.ErrTagNotFound)
		}
		if isUniqueViolation(err, "tags_slug_key") {
			return fmt.Errorf("tag slug already exists: %w", post.ErrTagNotFound)
		}
		return fmt.Errorf("failed to create tag: %w", err)
	}

	tag.CreatedAt = now
	return nil
}

// Delete 删除标签（级联删除 post_tags 关联）
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tags WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrTagNotFound
	}

	return nil
}

// GetByID 根据 ID 获取标签
func (r *TagRepository) GetByID(ctx context.Context, id uuid.UUID) (*post.Tag, error) {
	query := `
		SELECT id, name, slug, created_at
		FROM tags WHERE id = $1
	`

	tag, err := r.scanTag(ctx, r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrTagNotFound
		}
		return nil, fmt.Errorf("failed to get tag by id: %w", err)
	}
	return tag, nil
}

// GetBySlug 根据 Slug 获取标签
func (r *TagRepository) GetBySlug(ctx context.Context, slug string) (*post.Tag, error) {
	query := `
		SELECT id, name, slug, created_at
		FROM tags WHERE slug = $1
	`

	tag, err := r.scanTag(ctx, r.pool.QueryRow(ctx, query, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrTagNotFound
		}
		return nil, fmt.Errorf("failed to get tag by slug: %w", err)
	}
	return tag, nil
}

// GetByName 根据名称获取标签
func (r *TagRepository) GetByName(ctx context.Context, name string) (*post.Tag, error) {
	query := `
		SELECT id, name, slug, created_at
		FROM tags WHERE name = $1
	`

	tag, err := r.scanTag(ctx, r.pool.QueryRow(ctx, query, name))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrTagNotFound
		}
		return nil, fmt.Errorf("failed to get tag by name: %w", err)
	}
	return tag, nil
}

// GetAll 获取所有标签
func (r *TagRepository) GetAll(ctx context.Context) ([]*post.Tag, error) {
	query := `
		SELECT id, name, slug, created_at
		FROM tags ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tags: %w", err)
	}
	defer rows.Close()

	tags := make([]*post.Tag, 0)
	for rows.Next() {
		tag, err := r.scanTagFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// AddPostTag 为文章添加标签
func (r *TagRepository) AddPostTag(ctx context.Context, postID, tagID uuid.UUID) error {
	query := `
		INSERT INTO post_tags (post_id, tag_id)
		VALUES ($1, $2)
	`

	_, err := r.pool.Exec(ctx, query, postID, tagID)
	if err != nil {
		// UNIQUE 约束冲突表示关联已存在，忽略即可
		if isUniqueViolation(err, "post_tags_pkey") {
			return nil // 关联已存在，不视为错误
		}
		return fmt.Errorf("failed to add post tag: %w", err)
	}

	return nil
}

// RemovePostTag 移除文章标签
func (r *TagRepository) RemovePostTag(ctx context.Context, postID, tagID uuid.UUID) error {
	query := `
		DELETE FROM post_tags WHERE post_id = $1 AND tag_id = $2
	`

	_, err := r.pool.Exec(ctx, query, postID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove post tag: %w", err)
	}

	return nil
}

// GetPostTags 获取文章的所有标签
func (r *TagRepository) GetPostTags(ctx context.Context, postID uuid.UUID) ([]*post.Tag, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.created_at
		FROM tags t
		INNER JOIN post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = $1
		ORDER BY t.name ASC
	`

	rows, err := r.pool.Query(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post tags: %w", err)
	}
	defer rows.Close()

	tags := make([]*post.Tag, 0)
	for rows.Next() {
		tag, err := r.scanTagFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// GetPostCount 统计标签下文章数
func (r *TagRepository) GetPostCount(ctx context.Context, tagID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM post_tags WHERE tag_id = $1
	`

	var count int
	err := r.pool.QueryRow(ctx, query, tagID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get post count: %w", err)
	}

	return count, nil
}

// scanTag 扫描单行标签数据
func (r *TagRepository) scanTag(ctx context.Context, row pgx.Row) (*post.Tag, error) {
	tag := &post.Tag{}
	err := row.Scan(
		&tag.ID,
		&tag.Name,
		&tag.Slug,
		&tag.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

// scanTagFromRow 从 rows 扫描标签数据
func (r *TagRepository) scanTagFromRow(row pgx.Rows) (*post.Tag, error) {
	tag := &post.Tag{}
	err := row.Scan(
		&tag.ID,
		&tag.Name,
		&tag.Slug,
		&tag.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tag, nil
}