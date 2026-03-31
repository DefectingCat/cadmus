// Package post 文章、分类、标签管理模块
package post

import (
	"context"

	"github.com/google/uuid"
)

// PostRepository 文章数据访问接口
type PostRepository interface {
	// Create 创建新文章
	Create(ctx context.Context, post *Post) error

	// Update 更新文章
	Update(ctx context.Context, post *Post) error

	// Delete 删除文章
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取文章
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)

	// GetBySlug 根据 Slug 获取文章
	GetBySlug(ctx context.Context, slug string) (*Post, error)

	// List 分页获取文章列表，支持筛选
	List(ctx context.Context, filters PostListFilters, offset, limit int) ([]*Post, int, error)

	// GetByAuthor 获取作者的文章列表
	GetByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int) ([]*Post, int, error)

	// GetByCategory 获取分类下的文章列表
	GetByCategory(ctx context.Context, categoryID uuid.UUID, offset, limit int) ([]*Post, int, error)

	// GetBySeries 获取系列下的文章列表
	GetBySeries(ctx context.Context, seriesID uuid.UUID, offset, limit int) ([]*Post, int, error)

	// Search 全文搜索文章
	Search(ctx context.Context, query string, offset, limit int) ([]*Post, int, error)

	// IncrementViewCount 增加浏览计数
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	// IncrementLikeCount 增加点赞计数
	IncrementLikeCount(ctx context.Context, id uuid.UUID) error

	// CreateVersion 创建文章版本记录
	CreateVersion(ctx context.Context, version *PostVersion) error

	// GetVersions 获取文章版本历史
	GetVersions(ctx context.Context, postID uuid.UUID) ([]*PostVersion, error)

	// GetVersionByNumber 根据版本号获取特定版本
	GetVersionByNumber(ctx context.Context, postID uuid.UUID, version int) (*PostVersion, error)
}

// CategoryRepository 分类数据访问接口
type CategoryRepository interface {
	// Create 创建新分类
	Create(ctx context.Context, category *Category) error

	// Update 更新分类
	Update(ctx context.Context, category *Category) error

	// Delete 删除分类
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取分类
	GetByID(ctx context.Context, id uuid.UUID) (*Category, error)

	// GetBySlug 根据 Slug 获取分类
	GetBySlug(ctx context.Context, slug string) (*Category, error)

	// GetAll 获取所有分类
	GetAll(ctx context.Context) ([]*Category, error)

	// GetChildren 获取子分类
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*Category, error)

	// GetRootCategories 获取顶级分类（parent_id IS NULL）
	GetRootCategories(ctx context.Context) ([]*Category, error)

	// GetPostCount 统计分类下文章数
	GetPostCount(ctx context.Context, categoryID uuid.UUID) (int, error)
}

// TagRepository 标签数据访问接口
type TagRepository interface {
	// Create 创建新标签
	Create(ctx context.Context, tag *Tag) error

	// Delete 删除标签
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取标签
	GetByID(ctx context.Context, id uuid.UUID) (*Tag, error)

	// GetBySlug 根据 Slug 获取标签
	GetBySlug(ctx context.Context, slug string) (*Tag, error)

	// GetByName 根据名称获取标签
	GetByName(ctx context.Context, name string) (*Tag, error)

	// GetAll 获取所有标签
	GetAll(ctx context.Context) ([]*Tag, error)

	// AddPostTag 为文章添加标签
	AddPostTag(ctx context.Context, postID, tagID uuid.UUID) error

	// RemovePostTag 移除文章标签
	RemovePostTag(ctx context.Context, postID, tagID uuid.UUID) error

	// GetPostTags 获取文章的所有标签
	GetPostTags(ctx context.Context, postID uuid.UUID) ([]*Tag, error)

	// GetPostCount 统计标签下文章数
	GetPostCount(ctx context.Context, tagID uuid.UUID) (int, error)
}

// SeriesRepository 系列数据访问接口
type SeriesRepository interface {
	// Create 创建新系列
	Create(ctx context.Context, series *Series) error

	// Update 更新系列
	Update(ctx context.Context, series *Series) error

	// Delete 删除系列
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取系列
	GetByID(ctx context.Context, id uuid.UUID) (*Series, error)

	// GetBySlug 根据 Slug 获取系列
	GetBySlug(ctx context.Context, slug string) (*Series, error)

	// GetByAuthor 获取作者的系列列表
	GetByAuthor(ctx context.Context, authorID uuid.UUID) ([]*Series, error)

	// GetAll 获取所有系列
	GetAll(ctx context.Context) ([]*Series, error)
}

// PostLikeRepository 文章点赞仓库接口
type PostLikeRepository interface {
	// CreateIfNotExists 创建点赞记录（使用 ON CONFLICT DO NOTHING），返回是否实际创建
	// 同时原子更新文章的点赞计数
	CreateIfNotExists(ctx context.Context, postID, userID uuid.UUID) (created bool, err error)

	// DeleteIfExists 删除点赞记录（返回是否实际删除）
	// 同时原子更新文章的点赞计数
	DeleteIfExists(ctx context.Context, postID, userID uuid.UUID) (deleted bool, err error)

	// Exists 检查用户是否已点赞文章
	Exists(ctx context.Context, postID, userID uuid.UUID) (bool, error)

	// CountByPostID 统计文章的点赞数量
	CountByPostID(ctx context.Context, postID uuid.UUID) (int, error)

	// GetByUserID 获取用户的所有点赞记录
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*PostLike, error)
}