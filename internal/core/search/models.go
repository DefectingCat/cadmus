// Package search 提供全文搜索功能的核心数据模型。
//
// 该文件包含搜索系统的核心实体定义，包括：
//   - SearchResult: 搜索结果项结构
//   - SearchFilters: 搜索过滤条件
//   - SearchResponse: 搜索响应结构
//   - 语义化错误类型定义
//
// 主要用途：
//
//	用于博客系统的全文搜索功能，支持文章内容检索和筛选。
//
// 注意事项：
//   - 搜索结果包含相关性得分（Rank），用于结果排序
//   - 支持按分类、作者、状态、标签等条件筛选
//   - 使用 PostgreSQL 全文搜索或其他搜索引擎实现
//
// 作者：xfy
package search

import (
	"time"

	"github.com/google/uuid"
)

// SearchResult 搜索结果项。
//
// 表示单条搜索结果，包含文章的基本信息和相关性得分。
// 用于搜索 API 的返回数据结构。
type SearchResult struct {
	// ID 文章的唯一标识符
	ID uuid.UUID `json:"id"`

	// Title 文章标题，搜索匹配时可能高亮显示关键词
	Title string `json:"title"`

	// Slug 文章 URL 别名，用于生成文章链接
	Slug string `json:"slug"`

	// Excerpt 文章摘要，显示在搜索结果预览中
	Excerpt string `json:"excerpt"`

	// AuthorID 作者的用户 ID
	AuthorID uuid.UUID `json:"author_id"`

	// CategoryID 分类 ID，可选
	CategoryID uuid.UUID `json:"category_id,omitempty"`

	// Status 文章状态，通常为 "published"
	Status string `json:"status"`

	// Rank 搜索相关性得分，用于结果排序
	// 得分越高表示与搜索词的相关性越强
	Rank float64 `json:"rank"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 最后修改时间
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchFilters 搜索过滤条件。
//
// 用于构建搜索查询的过滤条件，支持多条件组合。
// 所有字段均为可选，未设置时忽略该条件。
type SearchFilters struct {
	// Query 搜索关键词，必填，最小长度为 1
	Query string `json:"query"`

	// Category 分类 Slug，按分类筛选结果
	Category string `json:"category,omitempty"`

	// AuthorID 作者 ID，按作者筛选结果
	AuthorID uuid.UUID `json:"author_id,omitempty"`

	// Status 文章状态筛选，默认为 "published"（仅搜索已发布文章）
	Status string `json:"status,omitempty"`

	// Tag 标签 Slug，按标签筛选结果
	Tag string `json:"tag,omitempty"`
}

// SearchResponse 搜索响应结构。
//
// 搜索 API 的标准返回格式，包含结果列表和分页信息。
type SearchResponse struct {
	// Results 搜索结果列表，按相关性得分降序排列
	Results []SearchResult `json:"results"`

	// Total 符合条件的总结果数，用于计算总页数
	Total int `json:"total"`

	// Page 当前页码，从 1 开始
	Page int `json:"page"`

	// PageSize 每页结果数量
	PageSize int `json:"page_size"`

	// Query 原始搜索关键词，用于前端显示
	Query string `json:"query"`
}

// SearchError 搜索模块自定义错误类型。
//
// 实现 error 和 errors.Is 接口，支持错误比较和类型判断。
type SearchError struct {
	// Code 错误代码，用于程序化错误判断
	Code string

	// Message 错误消息，用于展示给用户或记录日志
	Message string
}

// Error 实现 error 接口，返回错误消息。
func (e *SearchError) Error() string {
	return e.Message
}

// 常见错误定义。
//
// 使用语义化错误类型，便于调用方进行错误处理和判断。
var (
	// ErrEmptyQuery 空搜索关键词错误，搜索词为空时返回
	ErrEmptyQuery = &SearchError{Code: "empty_query", Message: "搜索关键词不能为空"}

	// ErrQueryTooLong 搜索关键词过长错误，超过系统限制时返回
	ErrQueryTooLong = &SearchError{Code: "query_too_long", Message: "搜索关键词过长"}
)
