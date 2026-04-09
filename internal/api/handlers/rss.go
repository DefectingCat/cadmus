// Package handlers 提供了 Cadmus API 的 HTTP 处理器实现。
//
// 该文件包含 RSS 订阅相关的核心逻辑，包括：
//   - RSS Feed 生成
//   - 分类筛选支持
//
// 主要用途：
//
//	用于生成符合 RSS 2.0 规范的订阅源，供 RSS 阅读器订阅。
//
// 注意事项：
//   - RSS 订阅公开访问，无需认证
//   - 支持按分类筛选文章
//   - 生成的内容会被缓存 1 小时
//
// 作者：xfy
package handlers

import (
	"net/http"

	"rua.plus/cadmus/internal/core/rss"
	"rua.plus/cadmus/internal/services"
)

// RSSHandler RSS 订阅处理器。
//
// 该处理器负责生成 RSS Feed，供用户订阅网站更新。
type RSSHandler struct {
	// rssService RSS 服务，处理 Feed 生成
	rssService services.RSSService

	// config Feed 配置，包含站点信息
	config rss.FeedConfig
}

// NewRSSHandler 创建 RSS 处理器。
//
// 参数：
//   - rssService: RSS 服务，处理 Feed 生成
//   - config: Feed 配置，包含站点标题、链接等信息
//
// 返回值：
//   - *RSSHandler: 新创建的 RSS 处理器实例
func NewRSSHandler(rssService services.RSSService, config rss.FeedConfig) *RSSHandler {
	return &RSSHandler{
		rssService: rssService,
		config:     config,
	}
}

// Feed 生成 RSS 订阅源。
//
// 生成符合 RSS 2.0 规范的 XML 订阅源。
// 支持按分类筛选文章。
//
// 路由：GET /api/v1/rss
//
// 查询参数：
//   - category: 分类标识（可选），用于筛选特定分类的文章
//
// 返回值：
//   - RSS XML 内容，Content-Type 为 application/xml
//   - 响应会被缓存 1 小时（Cache-Control: public, max-age=3600）
//
// 可能的错误：
//   - INTERNAL_ERROR: 生成 RSS 订阅源失败
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
	_, _ = w.Write([]byte(xmlContent))
}
