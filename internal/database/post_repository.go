package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"rua.plus/cadmus/internal/core/post"
)

// PostRepository 文章数据仓库实现
type PostRepository struct {
	pool *Pool
}

// NewPostRepository 创建文章仓库
func NewPostRepository(pool *Pool) *PostRepository {
	return &PostRepository{pool: pool}
}

// Create 创建新文章
func (r *PostRepository) Create(ctx context.Context, p *post.Post) error {
	query := `
		INSERT INTO posts (
			id, author_id, title, slug, content, content_text, excerpt,
			category_id, status, publish_at, featured_image,
			seo_title, seo_description, seo_keywords,
			view_count, like_count, comment_count,
			series_id, series_order, is_paid, price, version,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
	`

	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	now := p.CreatedAt
	if now.IsZero() {
		now = p.UpdatedAt
	}
	if now.IsZero() {
		now = time.Now()
	}

	// 处理分类 ID（零值表示无分类）
	var categoryID interface{}
	if p.CategoryID != uuid.Nil {
		categoryID = p.CategoryID
	}

	_, err := r.pool.Exec(ctx, query,
		p.ID,
		p.AuthorID,
		p.Title,
		p.Slug,
		p.Content,
		p.ContentText,
		p.Excerpt,
		categoryID,
		p.Status,
		p.PublishAt,
		p.FeaturedImage,
		p.SEOMeta.Title,
		p.SEOMeta.Description,
		p.SEOMeta.Keywords,
		p.ViewCount,
		p.LikeCount,
		p.CommentCount,
		p.SeriesID,
		p.SeriesOrder,
		p.IsPaid,
		p.Price,
		p.Version,
		now,
		now,
	)

	if err != nil {
		if isUniqueViolation(err, "posts_slug_key") {
			return post.ErrPostAlreadyExists
		}
		return fmt.Errorf("failed to create post: %w", err)
	}

	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// Update 更新文章，version 自增
func (r *PostRepository) Update(ctx context.Context, p *post.Post) error {
	query := `
		UPDATE posts SET
			title = $2, slug = $3, content = $4, content_text = $5, excerpt = $6,
			category_id = $7, status = $8, publish_at = $9, featured_image = $10,
			seo_title = $11, seo_description = $12, seo_keywords = $13,
			series_id = $14, series_order = $15, is_paid = $16, price = $17,
			version = version + 1
		WHERE id = $1
	`

	// 处理分类 ID
	var categoryID interface{}
	if p.CategoryID != uuid.Nil {
		categoryID = p.CategoryID
	}

	result, err := r.pool.Exec(ctx, query,
		p.ID,
		p.Title,
		p.Slug,
		p.Content,
		p.ContentText,
		p.Excerpt,
		categoryID,
		p.Status,
		p.PublishAt,
		p.FeaturedImage,
		p.SEOMeta.Title,
		p.SEOMeta.Description,
		p.SEOMeta.Keywords,
		p.SeriesID,
		p.SeriesOrder,
		p.IsPaid,
		p.Price,
	)

	if err != nil {
		if isUniqueViolation(err, "posts_slug_key") {
			return post.ErrPostAlreadyExists
		}
		return fmt.Errorf("failed to update post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrPostNotFound
	}

	// 获取更新后的版本号
	var newVersion int
	err = r.pool.QueryRow(ctx, "SELECT version FROM posts WHERE id = $1", p.ID).Scan(&newVersion)
	if err == nil {
		p.Version = newVersion
	}

	return nil
}

// Delete 删除文章（硬删除）
func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM posts WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrPostNotFound
	}

	return nil
}

// GetByID 根据 ID 获取文章
func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	query := `
		SELECT id, author_id, title, slug, content, content_text, excerpt,
			category_id, status, publish_at, featured_image,
			seo_title, seo_description, seo_keywords,
			view_count, like_count, comment_count,
			series_id, series_order, is_paid, price, version,
			created_at, updated_at
		FROM posts WHERE id = $1
	`

	p, err := r.scanPost(ctx, r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}
	return p, nil
}

// GetBySlug 根据 Slug 获取文章
func (r *PostRepository) GetBySlug(ctx context.Context, slug string) (*post.Post, error) {
	query := `
		SELECT id, author_id, title, slug, content, content_text, excerpt,
			category_id, status, publish_at, featured_image,
			seo_title, seo_description, seo_keywords,
			view_count, like_count, comment_count,
			series_id, series_order, is_paid, price, version,
			created_at, updated_at
		FROM posts WHERE slug = $1
	`

	p, err := r.scanPost(ctx, r.pool.QueryRow(ctx, query, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post by slug: %w", err)
	}
	return p, nil
}

// List 分页获取文章列表，支持筛选
func (r *PostRepository) List(ctx context.Context, filters post.PostListFilters, offset, limit int) ([]*post.Post, int, error) {
	// 构建基础查询条件
	whereClause := "WHERE 1=1"
	args := make([]interface{}, 0)
	argIndex := 1

	// 添加筛选条件
	if filters.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}
	if filters.AuthorID != uuid.Nil {
		whereClause += fmt.Sprintf(" AND author_id = $%d", argIndex)
		args = append(args, filters.AuthorID)
		argIndex++
	}
	if filters.CategoryID != uuid.Nil {
		whereClause += fmt.Sprintf(" AND category_id = $%d", argIndex)
		args = append(args, filters.CategoryID)
		argIndex++
	}
	if filters.SeriesID != uuid.Nil {
		whereClause += fmt.Sprintf(" AND series_id = $%d", argIndex)
		args = append(args, filters.SeriesID)
		argIndex++
	}
	if filters.TagID != uuid.Nil {
		whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM post_tags WHERE post_tags.post_id = posts.id AND post_tags.tag_id = $%d)", argIndex)
		args = append(args, filters.TagID)
		argIndex++
	}
	if filters.Search != "" {
		whereClause += fmt.Sprintf(" AND search_vector @@ websearch_to_tsquery('simple', $%d)", argIndex)
		args = append(args, filters.Search)
		argIndex++
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM posts %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	// 获取列表
	listQuery := fmt.Sprintf(`
		SELECT id, author_id, title, slug, content, content_text, excerpt,
			category_id, status, publish_at, featured_image,
			seo_title, seo_description, seo_keywords,
			view_count, like_count, comment_count,
			series_id, series_order, is_paid, price, version,
			created_at, updated_at
		FROM posts %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list posts: %w", err)
	}
	defer rows.Close()

	posts := make([]*post.Post, 0)
	for rows.Next() {
		p, err := r.scanPostFromRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, p)
	}

	return posts, total, nil
}

// GetByAuthor 获取作者的文章列表
func (r *PostRepository) GetByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int) ([]*post.Post, int, error) {
	return r.List(ctx, post.PostListFilters{AuthorID: authorID}, offset, limit)
}

// GetByCategory 获取分类下的文章列表
func (r *PostRepository) GetByCategory(ctx context.Context, categoryID uuid.UUID, offset, limit int) ([]*post.Post, int, error) {
	return r.List(ctx, post.PostListFilters{CategoryID: categoryID}, offset, limit)
}

// GetBySeries 获取系列下的文章列表
func (r *PostRepository) GetBySeries(ctx context.Context, seriesID uuid.UUID, offset, limit int) ([]*post.Post, int, error) {
	return r.List(ctx, post.PostListFilters{SeriesID: seriesID}, offset, limit)
}

// Search 全文搜索文章
func (r *PostRepository) Search(ctx context.Context, query string, offset, limit int) ([]*post.Post, int, error) {
	return r.List(ctx, post.PostListFilters{Search: query}, offset, limit)
}

// IncrementViewCount 增加浏览计数
func (r *PostRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE posts SET view_count = view_count + 1 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrPostNotFound
	}

	return nil
}

// IncrementLikeCount 增加点赞计数
func (r *PostRepository) IncrementLikeCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE posts SET like_count = like_count + 1 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment like count: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrPostNotFound
	}

	return nil
}

// CreateVersion 创建文章版本记录
func (r *PostRepository) CreateVersion(ctx context.Context, v *post.PostVersion) error {
	query := `
		INSERT INTO post_versions (id, post_id, version, content, creator_id, note, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}

	now := v.CreatedAt
	if now.IsZero() {
		now = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		v.ID,
		v.PostID,
		v.Version,
		v.Content,
		v.CreatorID,
		v.Note,
		now,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// 版本号已存在，这不是错误，只是忽略
			return nil
		}
		return fmt.Errorf("failed to create post version: %w", err)
	}

	v.CreatedAt = now
	return nil
}

// GetVersions 获取文章版本历史
func (r *PostRepository) GetVersions(ctx context.Context, postID uuid.UUID) ([]*post.PostVersion, error) {
	query := `
		SELECT id, post_id, version, content, creator_id, note, created_at
		FROM post_versions WHERE post_id = $1 ORDER BY version DESC
	`

	rows, err := r.pool.Query(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post versions: %w", err)
	}
	defer rows.Close()

	versions := make([]*post.PostVersion, 0)
	for rows.Next() {
		v := &post.PostVersion{}
		err := rows.Scan(
			&v.ID,
			&v.PostID,
			&v.Version,
			&v.Content,
			&v.CreatorID,
			&v.Note,
			&v.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}

// scanPost 扫描单行文章数据
func (r *PostRepository) scanPost(ctx context.Context, row pgx.Row) (*post.Post, error) {
	p := &post.Post{
		SEOMeta: post.SEOMeta{},
	}

	var categoryID uuid.UUID
	var seoKeywords []string

	err := row.Scan(
		&p.ID,
		&p.AuthorID,
		&p.Title,
		&p.Slug,
		&p.Content,
		&p.ContentText,
		&p.Excerpt,
		&categoryID,
		&p.Status,
		&p.PublishAt,
		&p.FeaturedImage,
		&p.SEOMeta.Title,
		&p.SEOMeta.Description,
		&seoKeywords,
		&p.ViewCount,
		&p.LikeCount,
		&p.CommentCount,
		&p.SeriesID,
		&p.SeriesOrder,
		&p.IsPaid,
		&p.Price,
		&p.Version,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	p.CategoryID = categoryID
	p.SEOMeta.Keywords = seoKeywords

	return p, nil
}

// scanPostFromRow 从 rows 扫描文章数据
func (r *PostRepository) scanPostFromRow(row pgx.Rows) (*post.Post, error) {
	p := &post.Post{
		SEOMeta: post.SEOMeta{},
	}

	var categoryID uuid.UUID
	var seoKeywords []string

	err := row.Scan(
		&p.ID,
		&p.AuthorID,
		&p.Title,
		&p.Slug,
		&p.Content,
		&p.ContentText,
		&p.Excerpt,
		&categoryID,
		&p.Status,
		&p.PublishAt,
		&p.FeaturedImage,
		&p.SEOMeta.Title,
		&p.SEOMeta.Description,
		&seoKeywords,
		&p.ViewCount,
		&p.LikeCount,
		&p.CommentCount,
		&p.SeriesID,
		&p.SeriesOrder,
		&p.IsPaid,
		&p.Price,
		&p.Version,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	p.CategoryID = categoryID
	p.SEOMeta.Keywords = seoKeywords

	return p, nil
}