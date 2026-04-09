// Package handlers 提供了 Cadmus API 的 HTTP 处理器实现。
//
// 该文件包含搜索相关的核心逻辑，包括：
//   - 全文搜索功能
//   - 搜索建议功能
//
// 主要用途：
//
//	用于实现博客内容的全文搜索和自动补全建议。
//
// 注意事项：
//   - 搜索功能公开访问，无需认证
//   - 支持多种筛选条件
//   - 搜索关键词有长度限制
//
// 作者：xfy
package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/search"
	"rua.plus/cadmus/internal/services"
)

// SearchHandler 搜索 API 处理器。
//
// 该处理器负责处理搜索相关的 HTTP 请求，包括全文搜索和搜索建议。
type SearchHandler struct {
	// searchService 搜索服务，处理搜索业务逻辑
	searchService services.SearchService
}

// NewSearchHandler 创建搜索处理器。
//
// 参数：
//   - searchService: 搜索服务，处理搜索业务逻辑
//
// 返回值：
//   - *SearchHandler: 新创建的搜索处理器实例
func NewSearchHandler(searchService services.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// Search 全文搜索。
//
// 根据关键词搜索文章内容，支持多种筛选条件。
// 搜索范围包括文章标题、内容等字段。
//
// 路由：GET /api/v1/search
//
// 查询参数：
//   - q: 搜索关键词（必填）
//   - status: 文章状态筛选
//   - category: 分类筛选
//   - tag: 标签筛选
//   - author_id: 作者 ID 筛选
//   - page: 页码，默认 1
//   - page_size: 每页数量，默认 20，最大 100
//
// 返回值（通过响应体）：
//   - 搜索结果列表和分页信息
//
// 可能的错误：
//   - BAD_REQUEST: 缺少搜索关键词或关键词为空/过长
//   - INTERNAL_ERROR: 搜索失败
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 解析搜索参数
	query := r.URL.Query().Get("q")
	if query == "" {
		WriteAPIError(w, "BAD_REQUEST", "缺少搜索关键词", nil, http.StatusBadRequest)
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建过滤条件
	filters := search.SearchFilters{
		Query:    query,
		Status:   r.URL.Query().Get("status"),
		Category: r.URL.Query().Get("category"),
		Tag:      r.URL.Query().Get("tag"),
	}

	// 解析作者 ID
	if authorID := r.URL.Query().Get("author_id"); authorID != "" {
		if id, err := uuid.Parse(authorID); err == nil {
			filters.AuthorID = id
		}
	}

	// 执行搜索
	response, err := h.searchService.Search(ctx, filters, page, pageSize)
	if err != nil {
		if err == search.ErrEmptyQuery {
			WriteAPIError(w, "BAD_REQUEST", "搜索关键词不能为空", nil, http.StatusBadRequest)
			return
		}
		if err == search.ErrQueryTooLong {
			WriteAPIError(w, "BAD_REQUEST", "搜索关键词过长", nil, http.StatusBadRequest)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "搜索失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, response, http.StatusOK)
}

// Suggestions 获取搜索建议。
//
// 根据输入的关键词前缀，返回搜索建议列表。
// 用于实现搜索框的自动补全功能。
//
// 路由：GET /api/v1/search/suggestions
//
// 查询参数：
//   - q: 搜索关键词前缀
//   - limit: 返回数量限制，默认 5，最大 10
//
// 返回值（通过响应体）：
//   - 建议字符串数组
func (h *SearchHandler) Suggestions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query().Get("q")
	if query == "" {
		WriteJSON(w, []string{}, http.StatusOK)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 10 {
		limit = 5
	}

	suggestions, err := h.searchService.GetSuggestions(ctx, query, limit)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取搜索建议失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, suggestions, http.StatusOK)
}
