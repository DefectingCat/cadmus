// Package search 全文搜索模块
package search

import (
	"time"

	"github.com/google/uuid"
)

// SearchResult 搜索结果项
type SearchResult struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt"`
	AuthorID    uuid.UUID  `json:"author_id"`
	CategoryID  uuid.UUID  `json:"category_id,omitempty"`
	Status      string     `json:"status"`
	Rank        float64    `json:"rank"`        // 搜索相关性得分
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// SearchFilters 搜索过滤条件
type SearchFilters struct {
	Query     string    `json:"query"`              // 搜索关键词
	Category  string    `json:"category,omitempty"` // 分类 slug 筛选
	AuthorID  uuid.UUID `json:"author_id,omitempty"` // 作者 ID 筛选
	Status    string    `json:"status,omitempty"`   // 文章状态筛选（默认 published）
	Tag       string    `json:"tag,omitempty"`      // 标签 slug 筛选
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Query    string         `json:"query"`
}

// SearchError 搜索模块错误
type SearchError struct {
	Code    string
	Message string
}

func (e *SearchError) Error() string {
	return e.Message
}

// 常见错误定义
var (
	ErrEmptyQuery = &SearchError{Code: "empty_query", Message: "搜索关键词不能为空"}
	ErrQueryTooLong = &SearchError{Code: "query_too_long", Message: "搜索关键词过长"}
)