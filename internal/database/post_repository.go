// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含文章数据仓库的核心逻辑，包括：
//   - 文章 CRUD 操作（创建、查询、更新、删除）
//   - 多种查询方式（ID、Slug、作者、分类、系列）
//   - 文章列表分页和筛选
//   - 文章版本历史管理
//   - 文章点赞关联操作
//
// 主要用途：
//
//	用于管理文章数据的持久化存储，支持完整的文章生命周期管理。
//
// 注意事项：
//   - 文章 Slug 有唯一约束
//   - 文章版本支持乐观锁机制
//   - 删除操作为硬删除，需谨慎使用
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
	"github.com/jackc/pgx/v5/pgconn"
	"rua.plus/cadmus/internal/core/post"
)

// PostRepository 文章数据仓库实现。
//
// 负责文章数据的 CRUD 操作，支持多种查询方式、分页筛选和版本管理。
// 所有操作通过连接池执行，确保高效的数据访问。
type PostRepository struct {
	// pool 数据库连接池
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

// GetVersionByNumber 根据版本号获取特定版本
func (r *PostRepository) GetVersionByNumber(ctx context.Context, postID uuid.UUID, version int) (*post.PostVersion, error) {
	query := `
		SELECT id, post_id, version, content, creator_id, note, created_at
		FROM post_versions WHERE post_id = $1 AND version = $2
	`

	v := &post.PostVersion{}
	err := r.pool.QueryRow(ctx, query, postID, version).Scan(
		&v.ID,
		&v.PostID,
		&v.Version,
		&v.Content,
		&v.CreatorID,
		&v.Note,
		&v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrVersionNotFound
		}
		return nil, fmt.Errorf("failed to get post version: %w", err)
	}

	return v, nil
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

// PostLikeRepository 文章点赞仓库实现
type PostLikeRepository struct {
	pool *Pool
}

// NewPostLikeRepository 创建文章点赞仓库
func NewPostLikeRepository(pool *Pool) *PostLikeRepository {
	return &PostLikeRepository{pool: pool}
}

// CreateIfNotExists 创建点赞记录（使用 ON CONFLICT DO NOTHING），返回是否实际创建
// 同时原子更新文章的点赞计数
func (r *PostLikeRepository) CreateIfNotExists(ctx context.Context, postID, userID uuid.UUID) (created bool, err error) {
	query := `
		INSERT INTO post_likes (id, post_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (post_id, user_id) DO NOTHING
	`

	id := uuid.New()
	now := time.Now()

	result, err := r.pool.Exec(ctx, query, id, postID, userID, now)
	if err != nil {
		return false, fmt.Errorf("failed to create post like: %w", err)
	}

	created = result.RowsAffected() > 0

	// 只有实际创建点赞记录时才更新计数
	if created {
		updateQuery := `UPDATE posts SET like_count = like_count + 1 WHERE id = $1`
		_, err = r.pool.Exec(ctx, updateQuery, postID)
		if err != nil {
			return false, fmt.Errorf("failed to update like count: %w", err)
		}
	}

	return created, nil
}

// DeleteIfExists 删除点赞记录（返回是否实际删除）
// 同时原子更新文章的点赞计数
func (r *PostLikeRepository) DeleteIfExists(ctx context.Context, postID, userID uuid.UUID) (deleted bool, err error) {
	query := `DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, postID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to delete post like: %w", err)
	}

	deleted = result.RowsAffected() > 0

	// 只有实际删除点赞记录时才更新计数
	if deleted {
		updateQuery := `UPDATE posts SET like_count = like_count - 1 WHERE id = $1 AND like_count > 0`
		_, err = r.pool.Exec(ctx, updateQuery, postID)
		if err != nil {
			return false, fmt.Errorf("failed to update like count: %w", err)
		}
	}

	return deleted, nil
}

// Exists 检查用户是否已点赞文章
func (r *PostLikeRepository) Exists(ctx context.Context, postID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = $1 AND user_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, postID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check post like exists: %w", err)
	}
	return exists, nil
}

// CountByPostID 统计文章的点赞数量
func (r *PostLikeRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM post_likes WHERE post_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, postID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count post likes: %w", err)
	}
	return count, nil
}

// GetByUserID 获取用户的所有点赞记录
func (r *PostLikeRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*post.PostLike, error) {
	query := `
		SELECT id, post_id, user_id, created_at
		FROM post_likes WHERE user_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post likes by user id: %w", err)
	}
	defer rows.Close()

	likes := make([]*post.PostLike, 0)
	for rows.Next() {
		l := &post.PostLike{}
		err := rows.Scan(
			&l.ID,
			&l.PostID,
			&l.UserID,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post like: %w", err)
		}
		likes = append(likes, l)
	}
	return likes, nil
}

// CountByAuthor 统计作者的文章数量
func (r *PostRepository) CountByAuthor(ctx context.Context, authorID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM posts WHERE author_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, authorID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts by author: %w", err)
	}

	return count, nil
}

// CountByAuthors 批量统计多个作者的文章数量
// 返回一个 map，key 是 authorID，value 是文章数量
func (r *PostRepository) CountByAuthors(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(userIDs) == 0 {
		return make(map[uuid.UUID]int), nil
	}

	query := `
		SELECT author_id, COUNT(*) as count
		FROM posts WHERE author_id = ANY($1)
		GROUP BY author_id
	`

	rows, err := r.pool.Query(ctx, query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts by authors: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]int)
	for rows.Next() {
		var authorID uuid.UUID
		var count int
		if err := rows.Scan(&authorID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan author count: %w", err)
		}
		result[authorID] = count
	}

	return result, nil
}

// MoveCategory 移动文章到指定分类
func (r *PostRepository) MoveCategory(ctx context.Context, postID, categoryID uuid.UUID) error {
	query := `UPDATE posts SET category_id = $2 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, postID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to move post category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrPostNotFound
	}

	return nil
}

// ChangeStatus 更改文章状态
func (r *PostRepository) ChangeStatus(ctx context.Context, postID uuid.UUID, status string) error {
	query := `UPDATE posts SET status = $2 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, postID, status)
	if err != nil {
		return fmt.Errorf("failed to change post status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrPostNotFound
	}

	return nil
}