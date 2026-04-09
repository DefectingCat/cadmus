// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含搜索数据仓库的核心逻辑，包括：
//   - 全文搜索文章（基于 PostgreSQL tsvector）
//   - 多条件筛选（作者、分类、标签）
//   - 搜索建议（自动补全）
//   - 搜索结果排序（按相关性 rank）
//
// 主要用途：
//
//	用于实现文章全文搜索功能，支持中文分词和多种筛选条件。
//
// 注意事项：
//   - 使用 PostgreSQL 的 websearch_to_tsquery 进行搜索
//   - 默认只搜索已发布的文章
//   - 搜索建议基于文章标题提取关键词
//
// 作者：xfy
package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/search"
)

// SearchRepository 搜索数据仓库实现。
//
// 负责全文搜索功能，基于 PostgreSQL 的 tsvector 实现。
// 支持多条件筛选和搜索建议。
type SearchRepository struct {
	// pool 数据库连接池
	pool *Pool
}

// NewSearchRepository 创建搜索仓库
func NewSearchRepository(pool *Pool) *SearchRepository {
	return &SearchRepository{pool: pool}
}

// Search 全文搜索文章
func (r *SearchRepository) Search(ctx context.Context, query string, filters search.SearchFilters, offset, limit int) ([]search.SearchResult, int, error) {
	// 构建基础查询
	baseQuery := `
		SELECT
			p.id, p.title, p.slug, p.excerpt, p.author_id, p.category_id, p.status,
			ts_rank(p.search_vector, websearch_to_tsquery('simple', $1)) as rank,
			p.created_at, p.updated_at
		FROM posts p
		WHERE p.search_vector @@ websearch_to_tsquery('simple', $1)
	`

	// 构建条件子句
	whereClause := ""
	args := []interface{}{query}
	argIndex := 2

	// 默认只搜索已发布文章
	if filters.Status == "" {
		whereClause = " AND p.status = 'published'"
	} else {
		whereClause = fmt.Sprintf(" AND p.status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	// 作者筛选
	if filters.AuthorID != uuid.Nil {
		whereClause += fmt.Sprintf(" AND p.author_id = $%d", argIndex)
		args = append(args, filters.AuthorID)
		argIndex++
	}

	// 分类筛选
	if filters.Category != "" {
		whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM categories c WHERE c.id = p.category_id AND c.slug = $%d)", argIndex)
		args = append(args, filters.Category)
		argIndex++
	}

	// 标签筛选
	if filters.Tag != "" {
		whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM post_tags pt JOIN tags t ON t.id = pt.tag_id WHERE pt.post_id = p.id AND t.slug = $%d)", argIndex)
		args = append(args, filters.Tag)
		argIndex++
	}

	// 组合查询
	fullQuery := baseQuery + whereClause + fmt.Sprintf(" ORDER BY rank DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// 执行搜索查询
	rows, err := r.pool.Query(ctx, fullQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search posts: %w", err)
	}
	defer rows.Close()

	results := make([]search.SearchResult, 0)
	for rows.Next() {
		var r search.SearchResult
		var categoryID uuid.UUID
		err := rows.Scan(
			&r.ID,
			&r.Title,
			&r.Slug,
			&r.Excerpt,
			&r.AuthorID,
			&categoryID,
			&r.Status,
			&r.Rank,
			&r.CreatedAt,
			&r.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}
		r.CategoryID = categoryID
		results = append(results, r)
	}

	// 获取总数
	countQuery := `
		SELECT COUNT(*)
		FROM posts p
		WHERE p.search_vector @@ websearch_to_tsquery('simple', $1)
	` + whereClause

	countArgs := args[:argIndex-2] // 去掉 limit 和 offset
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	return results, total, nil
}

// SearchByCategory 在指定分类下搜索
func (r *SearchRepository) SearchByCategory(ctx context.Context, query string, categoryID uuid.UUID, offset, limit int) ([]search.SearchResult, int, error) {
	filters := search.SearchFilters{
		Query: query,
	}
	// 需要先获取分类 slug
	var slug string
	err := r.pool.QueryRow(ctx, "SELECT slug FROM categories WHERE id = $1", categoryID).Scan(&slug)
	if err != nil {
		return nil, 0, fmt.Errorf("category not found: %w", err)
	}
	filters.Category = slug

	return r.Search(ctx, query, filters, offset, limit)
}

// SearchByAuthor 搜索指定作者的文章
func (r *SearchRepository) SearchByAuthor(ctx context.Context, query string, authorID uuid.UUID, offset, limit int) ([]search.SearchResult, int, error) {
	filters := search.SearchFilters{
		Query:    query,
		AuthorID: authorID,
	}
	return r.Search(ctx, query, filters, offset, limit)
}

// GetSuggestions 获取搜索建议
func (r *SearchRepository) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	// 基于文章标题提取匹配的关键词
	sql := `
		SELECT DISTINCT word
		FROM (
			SELECT unnest(string_to_array(lower(title), ' ')) as word
			FROM posts
			WHERE status = 'published'
			  AND lower(title) LIKE lower($1)
		) sub
		WHERE length(word) >= 2
		ORDER BY word
		LIMIT $2
	`

	pattern := query + "%"
	rows, err := r.pool.Query(ctx, sql, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}
	defer rows.Close()

	suggestions := make([]string, 0)
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			return nil, fmt.Errorf("failed to scan suggestion: %w", err)
		}
		suggestions = append(suggestions, word)
	}

	return suggestions, nil
}
