// Package services 提供搜索服务的实现。
//
// 该文件包含全文搜索相关的核心逻辑，包括：
//   - 全文搜索（支持关键词查询）
//   - 分类筛选搜索
//   - 作者筛选搜索
//   - 搜索建议（自动补全）
//
// 主要用途：
//
//	用于实现博客内容的全文搜索功能，提升用户查找效率。
//
// 设计特点：
//   - 支持多维度筛选（分类、作者）
//   - 分页查询优化
//   - 搜索建议提升用户体验
//
// 作者：xfy
package services

import (
	"context"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/search"
)

// SearchService 搜索服务接口。
//
// 该接口定义了全文搜索的操作，支持多维度筛选和搜索建议。
type SearchService interface {
	// Search 全文搜索，支持多条件筛选。
	//
	// 参数：
	//   - ctx: 上下文
	//   - filters: 搜索筛选条件（关键词、分类、作者等）
	//   - page: 页码（从 1 开始）
	//   - pageSize: 每页数量（最大 100）
	//
	// 返回值：
	//   - SearchResponse: 搜索结果，包含文章列表和分页信息
	//   - error: 搜索失败时返回错误
	Search(ctx context.Context, filters search.SearchFilters, page, pageSize int) (*search.SearchResponse, error)

	// SearchByCategory 在指定分类下搜索。
	SearchByCategory(ctx context.Context, query string, categoryID uuid.UUID, page, pageSize int) (*search.SearchResponse, error)

	// SearchByAuthor 搜索指定作者的文章。
	SearchByAuthor(ctx context.Context, query string, authorID uuid.UUID, page, pageSize int) (*search.SearchResponse, error)

	// GetSuggestions 获取搜索建议（自动补全）。
	//
	// 参数：
	//   - ctx: 上下文
	//   - query: 用户输入的关键词前缀
	//   - limit: 返回建议数量（最大 10）
	//
	// 返回值：
	//   - []string: 建议关键词列表
	//   - error: 查询失败
	GetSuggestions(ctx context.Context, query string, limit int) ([]string, error)
}

// searchServiceImpl 搜索服务的具体实现。
type searchServiceImpl struct {
	// repo 搜索数据仓库
	repo search.SearchRepository
}

// NewSearchService 创建搜索服务实例。
func NewSearchService(repo search.SearchRepository) SearchService {
	return &searchServiceImpl{repo: repo}
}

// Search 全文搜索
func (s *searchServiceImpl) Search(ctx context.Context, filters search.SearchFilters, page, pageSize int) (*search.SearchResponse, error) {
	// 验证关键词
	if filters.Query == "" {
		return nil, search.ErrEmptyQuery
	}
	if len(filters.Query) > 100 {
		return nil, search.ErrQueryTooLong
	}

	// 分页处理
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 执行搜索
	results, total, err := s.repo.Search(ctx, filters.Query, filters, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &search.SearchResponse{
		Results:  results,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Query:    filters.Query,
	}, nil
}

// SearchByCategory 在指定分类下搜索
func (s *searchServiceImpl) SearchByCategory(ctx context.Context, query string, categoryID uuid.UUID, page, pageSize int) (*search.SearchResponse, error) {
	if query == "" {
		return nil, search.ErrEmptyQuery
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	results, total, err := s.repo.SearchByCategory(ctx, query, categoryID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &search.SearchResponse{
		Results:  results,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Query:    query,
	}, nil
}

// SearchByAuthor 搜索指定作者的文章
func (s *searchServiceImpl) SearchByAuthor(ctx context.Context, query string, authorID uuid.UUID, page, pageSize int) (*search.SearchResponse, error) {
	if query == "" {
		return nil, search.ErrEmptyQuery
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	results, total, err := s.repo.SearchByAuthor(ctx, query, authorID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &search.SearchResponse{
		Results:  results,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Query:    query,
	}, nil
}

// GetSuggestions 获取搜索建议
func (s *searchServiceImpl) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if limit < 1 || limit > 10 {
		limit = 5
	}
	return s.repo.GetSuggestions(ctx, query, limit)
}
