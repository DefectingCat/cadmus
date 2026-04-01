// Package search 提供全文搜索的数据访问接口。
//
// 该文件定义搜索系统的 Repository 接口。
//
// 主要用途：
//
//	抽象搜索数据访问层，便于实现不同的搜索引擎后端，
//	如 PostgreSQL 全文搜索、Elasticsearch、MeiliSearch 等。
//
// 注意事项：
//   - 所有接口方法必须支持 context.Context 进行超时控制
//   - 搜索结果应按相关性得分降序排列
//   - 接口实现必须是并发安全的
//
// 作者：xfy
package search

import (
	"context"

	"github.com/google/uuid"
)

// SearchRepository 搜索数据访问接口。
//
// 定义搜索操作的方法，支持全文搜索和搜索建议。
// 实现该接口的类必须保证所有方法的并发安全性。
type SearchRepository interface {
	// Search 全文搜索文章。
	//
	// 根据关键词和过滤条件搜索文章，返回按相关性排序的结果列表。
	// 使用数据库全文索引或搜索引擎实现高效检索。
	//
	// 参数：
	//   - ctx: 上下文，用于控制超时和取消操作
	//   - query: 搜索关键词，不能为空
	//   - filters: 过滤条件，支持按分类、作者、状态、标签筛选
	//   - offset: 分页偏移量（从 0 开始）
	//   - limit: 每页数量
	//
	// 返回值：
	//   - results: 搜索结果列表，按相关性得分降序排列
	//   - total: 符合条件的总结果数（用于计算总页数）
	//   - err: 可能的错误包括 ErrEmptyQuery、ErrQueryTooLong
	//
	// 使用示例：
	//   results, total, err := repo.Search(ctx, "golang", SearchFilters{Status: "published"}, 0, 10)
	Search(ctx context.Context, query string, filters SearchFilters, offset, limit int) ([]SearchResult, int, error)

	// SearchByCategory 在指定分类下搜索。
	//
	// 限定搜索范围到特定分类，提高搜索精准度。
	// 等同于 Search 方法中设置 filters.Category 参数。
	//
	// 参数：
	//   - ctx: 上下文
	//   - query: 搜索关键词
	//   - categoryID: 分类 ID
	//   - offset: 分页偏移量
	//   - limit: 每页数量
	//
	// 返回值：
	//   - results: 搜索结果列表
	//   - total: 符合条件的总结果数
	//   - err: 搜索错误
	SearchByCategory(ctx context.Context, query string, categoryID uuid.UUID, offset, limit int) ([]SearchResult, int, error)

	// SearchByAuthor 搜索指定作者的文章。
	//
	// 限定搜索范围到特定作者的文章，用于作者页面搜索功能。
	// 等同于 Search 方法中设置 filters.AuthorID 参数。
	//
	// 参数：
	//   - ctx: 上下文
	//   - query: 搜索关键词
	//   - authorID: 作者用户 ID
	//   - offset: 分页偏移量
	//   - limit: 每页数量
	//
	// 返回值：
	//   - results: 搜索结果列表
	//   - total: 符合条件的总结果数
	//   - err: 搜索错误
	SearchByAuthor(ctx context.Context, query string, authorID uuid.UUID, offset, limit int) ([]SearchResult, int, error)

	// GetSuggestions 获取搜索建议。
	//
	// 根据用户输入的部分关键词，返回搜索建议列表。
	// 建议来源于历史搜索记录或热门关键词。
	// 用于搜索框自动补全功能。
	//
	// 参数：
	//   - ctx: 上下文
	//   - query: 用户输入的部分关键词
	//   - limit: 返回建议数量上限
	//
	// 返回值：
	//   - suggestions: 搜索建议字符串列表
	//   - err: 查询错误
	//
	// 使用示例：
	//   suggestions, err := repo.GetSuggestions(ctx, "go", 5)
	//   // 可能返回: ["golang", "google", "good"]
	GetSuggestions(ctx context.Context, query string, limit int) ([]string, error)
}