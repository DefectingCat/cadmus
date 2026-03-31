package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/search"
	"rua.plus/cadmus/internal/services"
)

// SearchHandler 搜索 API 处理器
type SearchHandler struct {
	searchService services.SearchService
}

// NewSearchHandler 创建搜索处理器
func NewSearchHandler(searchService services.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// Search 全文搜索
// GET /api/v1/search?q=xxx&category=xxx&author_id=xxx&tag=xxx&page=xxx&page_size=xxx
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
		Query:   query,
		Status:  r.URL.Query().Get("status"),
		Category: r.URL.Query().Get("category"),
		Tag:     r.URL.Query().Get("tag"),
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

// Suggestions 获取搜索建议
// GET /api/v1/search/suggestions?q=xxx
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