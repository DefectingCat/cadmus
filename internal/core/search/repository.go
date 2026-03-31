// Package search 全文搜索模块
package search

import (
	"context"

	"github.com/google/uuid"
)

// SearchRepository 搜索数据访问接口
type SearchRepository interface {
	// Search 全文搜索文章
	// query: 搜索关键词
	// filters: 过滤条件（分类、作者、状态、标签）
	// offset, limit: 分页参数
	Search(ctx context.Context, query string, filters SearchFilters, offset, limit int) ([]SearchResult, int, error)

	// SearchByCategory 在指定分类下搜索
	SearchByCategory(ctx context.Context, query string, categoryID uuid.UUID, offset, limit int) ([]SearchResult, int, error)

	// SearchByAuthor 搜索指定作者的文章
	SearchByAuthor(ctx context.Context, query string, authorID uuid.UUID, offset, limit int) ([]SearchResult, int, error)

	// GetSuggestions 获取搜索建议（基于历史搜索或热门关键词）
	GetSuggestions(ctx context.Context, query string, limit int) ([]string, error)
}