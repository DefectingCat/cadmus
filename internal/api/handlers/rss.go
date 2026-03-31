package handlers

import (
	"net/http"

	"rua.plus/cadmus/internal/core/rss"
	"rua.plus/cadmus/internal/services"
)

// RSSHandler RSS 订阅处理器
type RSSHandler struct {
	rssService services.RSSService
	config     rss.FeedConfig
}

// NewRSSHandler 创建 RSS 处理器
func NewRSSHandler(rssService services.RSSService, config rss.FeedConfig) *RSSHandler {
	return &RSSHandler{
		rssService: rssService,
		config:     config,
	}
}

// Feed 生成 RSS 订阅源
// GET /api/v1/rss
func (h *RSSHandler) Feed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取分类筛选参数
	category := r.URL.Query().Get("category")

	// 生成 RSS Feed
	xmlContent, err := h.rssService.GenerateFeed(ctx, h.config, category)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "生成 RSS 订阅源失败", nil, http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600") // 缓存 1 小时
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xmlContent))
}