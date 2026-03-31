// Package rss RSS 订阅模块
//
// 提供 RSS 2.0 规范兼容的订阅功能，用于发布文章更新通知
package rss

import (
	"encoding/xml"
	"time"
)

// RSSFeed RSS 2.0 订阅源结构
type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel *RSSChannel `xml:"channel"`
}

// RSSChannel RSS 2.0 频道结构
type RSSChannel struct {
	Title       string    `xml:"title"`              // 频道标题
	Link        string    `xml:"link"`               // 频道链接
	Description string    `xml:"description"`        // 频道描述
	Language    string    `xml:"language,omitempty"` // 语言标识 (如 zh-CN)
	PubDate     string    `xml:"pubDate,omitempty"`  // 发布时间 (RFC 822 格式)
	LastBuildDate string  `xml:"lastBuildDate,omitempty"` // 最后构建时间
	Generator   string    `xml:"generator,omitempty"`     // 生成器标识
	Items       []*RSSItem `xml:"item"`              // 文章列表
}

// RSSItem RSS 2.0 文章项结构
type RSSItem struct {
	Title       string `xml:"title"`              // 文章标题
	Link        string `xml:"link"`               // 文章链接
	Description string `xml:"description"`        // 文章摘要/描述
	Author      string `xml:"author,omitempty"`   // 作者邮箱或名称
	Category    string `xml:"category,omitempty"` // 文章分类
	PubDate     string `xml:"pubDate,omitempty"`  // 发布时间 (RFC 822 格式)
	GUID        string `xml:"guid,omitempty"`     // 文章唯一标识
}

// FeedConfig RSS 订阅源配置
type FeedConfig struct {
	Title       string // 频道标题
	Link        string // 频道链接 (网站首页)
	Description string // 频道描述
	Language    string // 语言标识
	Generator   string // 生成器标识
	BaseURL     string // 文章链接基础 URL
}

// DefaultFeedConfig 默认订阅源配置
func DefaultFeedConfig() FeedConfig {
	return FeedConfig{
		Title:       "Cadmus Blog",
		Link:        "https://cadmus.blog",
		Description: "Cadmus 博客平台最新文章",
		Language:    "zh-CN",
		Generator:   "Cadmus RSS Generator",
		BaseURL:     "https://cadmus.blog/posts",
	}
}

// FormatRFC822Time 将时间格式化为 RFC 822 格式 (RSS 2.0 规范要求)
func FormatRFC822Time(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC822)
}