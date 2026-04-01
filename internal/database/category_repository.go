// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含分类数据仓库的核心逻辑，包括：
//   - 分类 CRUD 操作（创建、查询、更新、删除）
//   - 多种查询方式（ID、Slug、父分类）
//   - 层级分类管理（支持父子关系）
//   - 分类排序管理
//
// 主要用途：
//
//	用于管理文章分类数据，支持层级分类结构。
//
// 注意事项：
//   - 分类 Slug 有唯一约束
//   - 删除分类前需检查是否有子分类或文章
//   - 分类支持层级关系（parent_id）
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

// CategoryRepository 分类数据仓库实现。
//
// 负责分类数据的 CRUD 操作，支持层级分类管理和批量操作。
// 所有操作通过连接池执行，确保高效的数据访问。
type CategoryRepository struct {
	// pool 数据库连接池
	pool *Pool
}

// NewCategoryRepository 创建分类仓库
func NewCategoryRepository(pool *Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

// Create 创建新分类
func (r *CategoryRepository) Create(ctx context.Context, category *post.Category) error {
	query := `
		INSERT INTO categories (id, name, slug, description, parent_id, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}

	now := category.CreatedAt
	if now.IsZero() {
		now = category.UpdatedAt
	}
	if now.IsZero() {
		now = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		category.ID,
		category.Name,
		category.Slug,
		category.Description,
		category.ParentID,
		category.SortOrder,
		now,
		now,
	)

	if err != nil {
		if IsUniqueViolation(err, "categories_slug_key") {
			return fmt.Errorf("分类 slug 已存在: %s", category.Slug)
		}
		return fmt.Errorf("failed to create category: %w", err)
	}

	category.CreatedAt = now
	category.UpdatedAt = now
	return nil
}

// Update 更新分类
func (r *CategoryRepository) Update(ctx context.Context, category *post.Category) error {
	query := `
		UPDATE categories SET name = $2, slug = $3, description = $4, parent_id = $5, sort_order = $6
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		category.ID,
		category.Name,
		category.Slug,
		category.Description,
		category.ParentID,
		category.SortOrder,
	)

	if err != nil {
		if IsUniqueViolation(err, "categories_slug_key") {
			return fmt.Errorf("分类 slug 已存在: %s", category.Slug)
		}
		return fmt.Errorf("failed to update category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrCategoryNotFound
	}

	return nil
}

// Delete 删除分类（检查是否有子分类或文章）
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// 检查是否有子分类
	hasChildren, err := r.hasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check children: %w", err)
	}
	if hasChildren {
		return fmt.Errorf("分类下存在子分类，无法删除")
	}

	// 检查是否有文章
	hasPosts, err := r.hasPosts(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check posts: %w", err)
	}
	if hasPosts {
		return fmt.Errorf("分类下存在文章，无法删除")
	}

	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return post.ErrCategoryNotFound
	}

	return nil
}

// GetByID 根据 ID 获取分类
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*post.Category, error) {
	query := `
		SELECT id, name, slug, description, parent_id, sort_order, created_at, updated_at
		FROM categories WHERE id = $1
	`

	category, err := r.scanCategory(ctx, r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category by id: %w", err)
	}
	return category, nil
}

// GetBySlug 根据 Slug 获取分类
func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*post.Category, error) {
	query := `
		SELECT id, name, slug, description, parent_id, sort_order, created_at, updated_at
		FROM categories WHERE slug = $1
	`

	category, err := r.scanCategory(ctx, r.pool.QueryRow(ctx, query, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}
	return category, nil
}

// GetAll 获取所有分类
func (r *CategoryRepository) GetAll(ctx context.Context) ([]*post.Category, error) {
	query := `
		SELECT id, name, slug, description, parent_id, sort_order, created_at, updated_at
		FROM categories ORDER BY sort_order ASC, name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*post.Category, 0)
	for rows.Next() {
		category, err := r.scanCategoryFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetChildren 获取子分类
func (r *CategoryRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*post.Category, error) {
	query := `
		SELECT id, name, slug, description, parent_id, sort_order, created_at, updated_at
		FROM categories WHERE parent_id = $1 ORDER BY sort_order ASC, name ASC
	`

	rows, err := r.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*post.Category, 0)
	for rows.Next() {
		category, err := r.scanCategoryFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetRootCategories 获取顶级分类（parent_id IS NULL）
func (r *CategoryRepository) GetRootCategories(ctx context.Context) ([]*post.Category, error) {
	query := `
		SELECT id, name, slug, description, parent_id, sort_order, created_at, updated_at
		FROM categories WHERE parent_id IS NULL ORDER BY sort_order ASC, name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get root categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*post.Category, 0)
	for rows.Next() {
		category, err := r.scanCategoryFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetPostCount 统计分类下文章数
func (r *CategoryRepository) GetPostCount(ctx context.Context, categoryID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM posts WHERE category_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, categoryID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get post count: %w", err)
	}

	return count, nil
}

// hasChildren 检查是否有子分类
func (r *CategoryRepository) hasChildren(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE parent_id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// hasPosts 检查是否有文章
func (r *CategoryRepository) hasPosts(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE category_id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// scanCategory 扫描单行分类数据
func (r *CategoryRepository) scanCategory(ctx context.Context, row pgx.Row) (*post.Category, error) {
	category := &post.Category{}
	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.ParentID,
		&category.SortOrder,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return category, nil
}

// scanCategoryFromRow 从 rows 扫描分类数据
func (r *CategoryRepository) scanCategoryFromRow(row pgx.Rows) (*post.Category, error) {
	category := &post.Category{}
	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.ParentID,
		&category.SortOrder,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return category, nil
}

// UpdateOrder 批量更新分类排序
func (r *CategoryRepository) UpdateOrder(ctx context.Context, order []uuid.UUID) error {
	query := `UPDATE categories SET sort_order = $2 WHERE id = $1`

	for i, id := range order {
		_, err := r.pool.Exec(ctx, query, id, i)
		if err != nil {
			return fmt.Errorf("failed to update category order: %w", err)
		}
	}

	return nil
}

// GetByIDs 批量获取多个分类
// 返回一个 map，key 是 category ID，value 是分类对象
func (r *CategoryRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*post.Category, error) {
	if len(ids) == 0 {
		return make(map[uuid.UUID]*post.Category), nil
	}

	query := `
		SELECT id, name, slug, description, parent_id, sort_order, created_at, updated_at
		FROM categories WHERE id = ANY($1)
	`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories by ids: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*post.Category)
	for rows.Next() {
		category, err := r.scanCategoryFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		result[category.ID] = category
	}

	return result, nil
}