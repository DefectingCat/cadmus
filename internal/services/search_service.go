package services

import (
	"context"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/search"
)

// SearchService 搜索服务接口
type SearchService interface {
	// Search 全文搜索
	Search(ctx context.Context, filters search.SearchFilters, page, pageSize int) (*search.SearchResponse, error)

	// SearchByCategory 在指定分类下搜索
	SearchByCategory(ctx context.Context, query string, categoryID uuid.UUID, page, pageSize int) (*search.SearchResponse, error)

	// SearchByAuthor 搜索指定作者的文章
	SearchByAuthor(ctx context.Context, query string, authorID uuid.UUID, page, pageSize int) (*search.SearchResponse, error)

	// GetSuggestions 获取搜索建议
	GetSuggestions(ctx context.Context, query string, limit int) ([]string, error)
}

// searchServiceImpl 搜索服务实现
type searchServiceImpl struct {
	repo search.SearchRepository
}

// NewSearchService 创建搜索服务
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