// Package post 提供文章、分类、标签、系列管理的数据访问接口。
//
// 该文件定义内容管理系统的 Repository 接口，包括：
//   - PostRepository: 文章数据访问接口
//   - CategoryRepository: 分类数据访问接口
//   - TagRepository: 标签数据访问接口
//   - SeriesRepository: 系列数据访问接口
//   - PostLikeRepository: 点赞数据访问接口
//
// 主要用途：
//
//	抽象数据访问层，便于实现不同的存储后端（如 PostgreSQL、MySQL），
//	并支持单元测试时使用 mock 实现。
//
// 注意事项：
//   - 所有接口方法必须支持 context.Context 进行超时控制
//   - 返回的错误应使用 models.go 中定义的语义化错误类型
//   - 接口实现必须是并发安全的
//
// 作者：xfy
package post

import (
	"context"

	"github.com/google/uuid"
)

// PostRepository 文章数据访问接口。
//
// 定义文章实体的 CRUD 操作和查询方法，支持复杂的筛选条件和分页。
// 实现该接口的类必须保证所有方法的并发安全性。
type PostRepository interface {
	// Create 创建新文章。
	//
	// 参数：
	//   - ctx: 上下文，用于控制超时和取消操作
	//   - post: 文章对象，必填字段包括 AuthorID、Title、Slug、Status
	//
	// 返回值：
	//   - err: 可能的错误包括 ErrPostAlreadyExists（Slug 冲突）
	Create(ctx context.Context, post *Post) error

	// Update 更新文章。
	//
	// 参数：
	//   - ctx: 上下文
	//   - post: 文章对象，ID 字段必须有效
	//
	// 返回值：
	//   - err: 可能的错误包括 ErrPostNotFound、ErrPermissionDenied
	Update(ctx context.Context, post *Post) error

	// Delete 删除文章。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 文章 ID
	//
	// 返回值：
	//   - err: 可能的错误包括 ErrPostNotFound、ErrPermissionDenied
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取文章。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 文章 ID
	//
	// 返回值：
	//   - post: 文章对象，包含完整信息（含标签列表）
	//   - err: 文章不存在时返回 ErrPostNotFound
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)

	// GetBySlug 根据 Slug 获取文章。
	//
	// Slug 是文章的 URL 别名，用于生成友好的永久链接。
	//
	// 参数：
	//   - ctx: 上下文
	//   - slug: 文章 Slug
	//
	// 返回值：
	//   - post: 文章对象
	//   - err: 文章不存在时返回 ErrPostNotFound
	GetBySlug(ctx context.Context, slug string) (*Post, error)

	// List 分页获取文章列表，支持筛选。
	//
	// 参数：
	//   - ctx: 上下文
	//   - filters: 筛选条件，所有字段可选
	//   - offset: 分页偏移量（从 0 开始）
	//   - limit: 每页数量
	//
	// 返回值：
	//   - posts: 文章列表
	//   - total: 符合条件的总数（用于计算总页数）
	//   - err: 查询错误
	List(ctx context.Context, filters PostListFilters, offset, limit int) ([]*Post, int, error)

	// GetByAuthor 获取作者的文章列表。
	//
	// 参数：
	//   - ctx: 上下文
	//   - authorID: 作者用户 ID
	//   - offset: 分页偏移量
	//   - limit: 每页数量
	//
	// 返回值：
	//   - posts: 文章列表
	//   - total: 作者文章总数
	//   - err: 查询错误
	GetByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int) ([]*Post, int, error)

	// GetByCategory 获取分类下的文章列表。
	//
	// 参数：
	//   - ctx: 上下文
	//   - categoryID: 分类 ID
	//   - offset: 分页偏移量
	//   - limit: 每页数量
	//
	// 返回值：
	//   - posts: 文章列表
	//   - total: 分类下文章总数
	//   - err: 分类不存在时返回 ErrCategoryNotFound
	GetByCategory(ctx context.Context, categoryID uuid.UUID, offset, limit int) ([]*Post, int, error)

	// GetBySeries 获取系列下的文章列表。
	//
	// 返回的文章按 SeriesOrder 字段排序，确保系列文章的顺序正确。
	//
	// 参数：
	//   - ctx: 上下文
	//   - seriesID: 系列 ID
	//   - offset: 分页偏移量
	//   - limit: 每页数量
	//
	// 返回值：
	//   - posts: 文章列表（按 SeriesOrder 排序）
	//   - total: 系列下文章总数
	//   - err: 系列不存在时返回 ErrSeriesNotFound
	GetBySeries(ctx context.Context, seriesID uuid.UUID, offset, limit int) ([]*Post, int, error)

	// Search 全文搜索文章。
	//
	// 使用 ContentText 字段进行全文搜索，支持关键词匹配。
	//
	// 参数：
	//   - ctx: 上下文
	//   - query: 搜索关键词
	//   - offset: 分页偏移量
	//   - limit: 每页数量
	//
	// 返回值：
	//   - posts: 匹配的文章列表（按相关性排序）
	//   - total: 匹配总数
	//   - err: 搜索错误
	Search(ctx context.Context, query string, offset, limit int) ([]*Post, int, error)

	// IncrementViewCount 增加浏览计数。
	//
	// 每次文章被访问时调用，用于统计文章热度。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 文章 ID
	//
	// 返回值：
	//   - err: 文章不存在时返回 ErrPostNotFound
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	// IncrementLikeCount 增加点赞计数。
	//
	// 用户点赞成功后调用，与 PostLikeRepository 配合使用。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 文章 ID
	//
	// 返回值：
	//   - err: 文章不存在时返回 ErrPostNotFound
	IncrementLikeCount(ctx context.Context, id uuid.UUID) error

	// CreateVersion 创建文章版本记录。
	//
	// 每次文章保存时创建新版本，用于版本历史和回溯。
	//
	// 参数：
	//   - ctx: 上下文
	//   - version: 版本记录对象
	//
	// 返回值：
	//   - err: 创建失败错误
	CreateVersion(ctx context.Context, version *PostVersion) error

	// GetVersions 获取文章版本历史。
	//
	// 返回文章的所有版本记录，按版本号降序排列。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//
	// 返回值：
	//   - versions: 版本历史列表
	//   - err: 查询错误
	GetVersions(ctx context.Context, postID uuid.UUID) ([]*PostVersion, error)

	// GetVersionByNumber 根据版本号获取特定版本。
	//
	// 用于查看或恢复历史版本内容。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//   - version: 版本号
	//
	// 返回值：
	//   - postVersion: 版本记录对象
	//   - err: 版本不存在时返回 ErrVersionNotFound
	GetVersionByNumber(ctx context.Context, postID uuid.UUID, version int) (*PostVersion, error)
}

// CategoryRepository 分类数据访问接口。
//
// 定义分类实体的 CRUD 操作和层级关系查询方法。
// 支持父子分类结构，构建树形分类体系。
type CategoryRepository interface {
	// Create 创建新分类。
	//
	// 参数：
	//   - ctx: 上下文
	//   - category: 分类对象，必填字段包括 Name、Slug
	//
	// 返回值：
	//   - err: Slug 冲突时返回 ErrCategoryNotFound 的变体
	Create(ctx context.Context, category *Category) error

	// Update 更新分类。
	//
	// 参数：
	//   - ctx: 上下文
	//   - category: 分类对象，ID 字段必须有效
	//
	// 返回值：
	//   - err: 分类不存在时返回 ErrCategoryNotFound
	Update(ctx context.Context, category *Category) error

	// Delete 删除分类。
	//
	// 注意：删除分类前应检查是否有文章或子分类关联。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 分类 ID
	//
	// 返回值：
	//   - err: 分类不存在时返回 ErrCategoryNotFound
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取分类。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 分类 ID
	//
	// 返回值：
	//   - category: 分类对象
	//   - err: 分类不存在时返回 ErrCategoryNotFound
	GetByID(ctx context.Context, id uuid.UUID) (*Category, error)

	// GetBySlug 根据 Slug 获取分类。
	//
	// 参数：
	//   - ctx: 上下文
	//   - slug: 分类 Slug
	//
	// 返回值：
	//   - category: 分类对象
	//   - err: 分类不存在时返回 ErrCategoryNotFound
	GetBySlug(ctx context.Context, slug string) (*Category, error)

	// GetAll 获取所有分类。
	//
	// 返回系统中所有分类，不分层级。
	//
	// 参数：
	//   - ctx: 上下文
	//
	// 返回值：
	//   - categories: 分类列表
	//   - err: 查询错误
	GetAll(ctx context.Context) ([]*Category, error)

	// GetChildren 获取子分类。
	//
	// 查询指定分类的直接子分类（不递归）。
	//
	// 参数：
	//   - ctx: 上下文
	//   - parentID: 父分类 ID
	//
	// 返回值：
	//   - categories: 子分类列表
	//   - err: 查询错误
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*Category, error)

	// GetRootCategories 获取顶级分类。
//
// 返回所有无父分类的顶级分类（ParentID 为空）。
//
// 参数：
//   - ctx: 上下文
//
// 返回值：
//   - categories: 顶级分类列表
//   - err: 查询错误
GetRootCategories(ctx context.Context) ([]*Category, error)

	// GetPostCount 统计分类下文章数。
	//
	// 用于展示分类的文章数量，辅助用户导航。
	//
	// 参数：
	//   - ctx: 上下文
	//   - categoryID: 分类 ID
	//
	// 返回值：
	//   - count: 文章数量（仅统计已发布文章）
	//   - err: 查询错误
	GetPostCount(ctx context.Context, categoryID uuid.UUID) (int, error)
}

// TagRepository 标签数据访问接口。
//
// 定义标签实体的 CRUD 操作和文章-标签关联管理方法。
// 标签与文章是多对多关系，通过中间表关联。
type TagRepository interface {
	// Create 创建新标签。
	//
	// 参数：
	//   - ctx: 上下文
	//   - tag: 标签对象，必填字段包括 Name、Slug
	//
	// 返回值：
	//   - err: Slug 冲突时返回错误
	Create(ctx context.Context, tag *Tag) error

	// Delete 删除标签。
	//
	// 删除标签时会自动解除与文章的关联关系。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 标签 ID
	//
	// 返回值：
	//   - err: 标签不存在时返回 ErrTagNotFound
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取标签。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 标签 ID
	//
	// 返回值：
	//   - tag: 标签对象
	//   - err: 标签不存在时返回 ErrTagNotFound
	GetByID(ctx context.Context, id uuid.UUID) (*Tag, error)

	// GetBySlug 根据 Slug 获取标签。
	//
	// 参数：
	//   - ctx: 上下文
	//   - slug: 标签 Slug
	//
	// 返回值：
	//   - tag: 标签对象
	//   - err: 标签不存在时返回 ErrTagNotFound
	GetBySlug(ctx context.Context, slug string) (*Tag, error)

	// GetByName 根据名称获取标签。
	//
	// 用于根据用户输入查找已存在的标签。
	//
	// 参数：
	//   - ctx: 上下文
	//   - name: 标签名称
	//
	// 返回值：
	//   - tag: 标签对象
	//   - err: 标签不存在时返回 ErrTagNotFound
	GetByName(ctx context.Context, name string) (*Tag, error)

	// GetAll 获取所有标签。
	//
	// 参数：
	//   - ctx: 上下文
	//
	// 返回值：
	//   - tags: 标签列表
	//   - err: 查询错误
	GetAll(ctx context.Context) ([]*Tag, error)

	// AddPostTag 为文章添加标签。
	//
	// 创建文章-标签关联记录，支持多对多关系。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//   - tagID: 标签 ID
	//
	// 返回值：
	//   - err: 关联已存在时忽略，其他错误返回
	AddPostTag(ctx context.Context, postID, tagID uuid.UUID) error

	// RemovePostTag 移除文章标签。
	//
	// 删除文章-标签关联记录。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//   - tagID: 标签 ID
	//
	// 返回值：
	//   - err: 关联不存在时忽略，其他错误返回
	RemovePostTag(ctx context.Context, postID, tagID uuid.UUID) error

	// GetPostTags 获取文章的所有标签。
	//
	// 查询文章关联的所有标签列表。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//
	// 返回值：
	//   - tags: 标签列表
	//   - err: 查询错误
	GetPostTags(ctx context.Context, postID uuid.UUID) ([]*Tag, error)

	// GetPostCount 统计标签下文章数。
	//
	// 用于展示标签的文章数量，辅助用户筛选。
	//
	// 参数：
	//   - ctx: 上下文
	//   - tagID: 标签 ID
	//
	// 返回值：
	//   - count: 文章数量（仅统计已发布文章）
	//   - err: 查询错误
	GetPostCount(ctx context.Context, tagID uuid.UUID) (int, error)
}

// SeriesRepository 系列数据访问接口。
//
// 定义文章系列的 CRUD 操作和查询方法。
// 系列用于组织连载文章或相关主题文章集合。
type SeriesRepository interface {
	// Create 创建新系列。
	//
	// 参数：
	//   - ctx: 上下文
	//   - series: 系列对象，必填字段包括 AuthorID、Title、Slug
	//
	// 返回值：
	//   - err: Slug 冲突时返回错误
	Create(ctx context.Context, series *Series) error

	// Update 更新系列。
	//
	// 参数：
	//   - ctx: 上下文
	//   - series: 系列对象，ID 字段必须有效
	//
	// 返回值：
	//   - err: 系列不存在时返回 ErrSeriesNotFound
	Update(ctx context.Context, series *Series) error

	// Delete 删除系列。
	//
	// 注意：删除系列不会删除系列内的文章，仅解除文章的系列关联。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 系列 ID
	//
	// 返回值：
	//   - err: 系列不存在时返回 ErrSeriesNotFound
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取系列。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 系列 ID
	//
	// 返回值：
	//   - series: 系列对象
	//   - err: 系列不存在时返回 ErrSeriesNotFound
	GetByID(ctx context.Context, id uuid.UUID) (*Series, error)

	// GetBySlug 根据 Slug 获取系列。
	//
	// 参数：
	//   - ctx: 上下文
	//   - slug: 系列 Slug
	//
	// 返回值：
	//   - series: 系列对象
	//   - err: 系列不存在时返回 ErrSeriesNotFound
	GetBySlug(ctx context.Context, slug string) (*Series, error)

	// GetByAuthor 获取作者的系列列表。
	//
	// 返回指定作者创建的所有系列。
	//
	// 参数：
	//   - ctx: 上下文
	//   - authorID: 作者用户 ID
	//
	// 返回值：
	//   - seriesList: 系列列表
	//   - err: 查询错误
	GetByAuthor(ctx context.Context, authorID uuid.UUID) ([]*Series, error)

	// GetAll 获取所有系列。
	//
	// 参数：
	//   - ctx: 上下文
	//
	// 返回值：
	//   - seriesList: 系列列表
	//   - err: 查询错误
	GetAll(ctx context.Context) ([]*Series, error)
}

// PostLikeRepository 文章点赞数据访问接口。
//
// 定义点赞记录的创建、删除和查询方法。
// 点赞记录用于防止重复点赞和查询用户点赞历史。
type PostLikeRepository interface {
	// CreateIfNotExists 创建点赞记录（使用 ON CONFLICT DO NOTHING）。
	//
	// 使用原子操作确保不会重复点赞，同时更新文章的点赞计数。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - created: true 表示实际创建了记录（新点赞），false 表示已存在
	//   - err: 操作错误
	CreateIfNotExists(ctx context.Context, postID, userID uuid.UUID) (created bool, err error)

	// DeleteIfExists 删除点赞记录（返回是否实际删除）。
	//
	// 使用原子操作确保取消点赞的准确性，同时更新文章的点赞计数。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - deleted: true 表示实际删除了记录（取消点赞），false 表示不存在
	//   - err: 操作错误
	DeleteIfExists(ctx context.Context, postID, userID uuid.UUID) (deleted bool, err error)

	// Exists 检查用户是否已点赞文章。
	//
	// 用于前端展示点赞状态。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - exists: true 表示已点赞
	//   - err: 查询错误
	Exists(ctx context.Context, postID, userID uuid.UUID) (bool, error)

	// CountByPostID 统计文章的点赞数量。
	//
	// 用于展示文章的点赞数，与 Post.LikeCount 字段对应。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//
	// 返回值：
	//   - count: 点赞数量
	//   - err: 查询错误
	CountByPostID(ctx context.Context, postID uuid.UUID) (int, error)

	// GetByUserID 获取用户的所有点赞记录。
	//
	// 用于展示用户的点赞历史或分析用户兴趣。
	//
	// 参数：
	//   - ctx: 上下文
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - likes: 点赞记录列表
	//   - err: 查询错误
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*PostLike, error)
}