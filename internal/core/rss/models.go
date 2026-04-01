// Package rss 提供 RSS 2.0 规范兼容的订阅功能实现。
//
// 该文件包含 RSS 订阅系统的核心数据结构，包括：
//   - RSSFeed: RSS 2.0 订阅源根结构
//   - RSSChannel: RSS 2.0 频道结构
//   - RSSItem: RSS 2.0 文章项结构
//   - FeedConfig: 订阅源配置结构
//
// 主要用途：
//
//	用于生成博客的 RSS 订阅源，让用户通过 RSS 阅读器订阅文章更新。
//
// 注意事项：
//   - 遵循 RSS 2.0 规范 (https://www.rssboard.org/rss-specification)
//   - 时间格式必须使用 RFC 822 标准
//   - GUID 应为文章的唯一标识，通常使用文章链接
//
// 作者：xfy
package rss

import (
	"encoding/xml"
	"time"
)

// RSSFeed RSS 2.0 订阅源根结构。
//
// RSS 订阅源的顶层 XML 结构，包含版本声明和频道信息。
// 生成 XML 时使用 xml 包的序列化功能。
type RSSFeed struct {
	// XMLName XML 元素名称，自动设置为 "rss"
	XMLName xml.Name `xml:"rss"`

	// Version RSS 版本号，固定为 "2.0"
	Version string `xml:"version,attr"`

	// Channel 频道信息，包含订阅源的元数据和文章列表
	Channel *RSSChannel `xml:"channel"`
}

// RSSChannel RSS 2.0 频道结构。
//
// 表示 RSS 订阅源的频道，包含网站信息和文章列表。
// 频道是订阅源的核心容器，每个订阅源通常只有一个频道。
type RSSChannel struct {
	// Title 频道标题，显示在 RSS 阅读器的订阅列表中
	Title string `xml:"title"`

	// Link 频道链接，指向网站首页或订阅源页面
	Link string `xml:"link"`

	// Description 频道描述，简要说明订阅源的内容主题
	Description string `xml:"description"`

	// Language 语言标识，如 "zh-CN" 表示中文简体
	Language string `xml:"language,omitempty"`

	// PubDate 发布时间，使用 RFC 822 格式 (如 "Wed, 02 Oct 2024 08:00:00 +0000")
	PubDate string `xml:"pubDate,omitempty"`

	// LastBuildDate 最后构建时间，表示订阅源最后更新的时间
	LastBuildDate string `xml:"lastBuildDate,omitempty"`

	// Generator 生成器标识，说明订阅源由什么工具生成
	Generator string `xml:"generator,omitempty"`

	// Items 文章列表，包含订阅源中的所有文章条目
	Items []*RSSItem `xml:"item"`
}

// RSSItem RSS 2.0 文章项结构。
//
// 表示 RSS 订阅源中的单个文章条目，包含文章的元数据和链接。
// 用户点击文章项后跳转到原文页面查看完整内容。
type RSSItem struct {
	// Title 文章标题，显示在 RSS 阅读器的文章列表中
	Title string `xml:"title"`

	// Link 文章链接，指向博客中的原文页面
	Link string `xml:"link"`

	// Description 文章摘要或描述，显示在阅读器的预览区域
	// 可以是纯文本或 HTML 格式
	Description string `xml:"description"`

	// Author 作者信息，可以是作者邮箱或名称
	Author string `xml:"author,omitempty"`

	// Category 文章分类，用于组织文章主题
	Category string `xml:"category,omitempty"`

	// PubDate 发布时间，使用 RFC 822 格式
	PubDate string `xml:"pubDate,omitempty"`

	// GUID 文章唯一标识，通常使用文章链接作为 GUID
	// RSS 阅读器使用 GUID 判断文章是否已读
	GUID string `xml:"guid,omitempty"`
}

// FeedConfig RSS 订阅源配置。
//
// 用于自定义订阅源的元信息，如标题、链接、描述等。
// 通过此配置可以生成不同风格或主题的订阅源。
type FeedConfig struct {
	// Title 频道标题，建议使用网站名称或博客主题
	Title string

	// Link 频道链接，指向网站首页
	Link string

	// Description 频道描述，简要说明内容主题
	Description string

	// Language 语言标识，如 "zh-CN"、"en-US"
	Language string

	// Generator 生成器标识，用于标识生成工具
	Generator string

	// BaseURL 文章链接基础 URL，用于拼接文章 Slug 生成完整链接
	// 如 "https://cadmus.blog/posts" + "/my-article" = "https://cadmus.blog/posts/my-article"
	BaseURL string
}

// DefaultFeedConfig 返回默认订阅源配置。
//
// 提供一套预设的配置值，适用于 Cadmus 博客平台默认场景。
// 可通过修改返回值的字段自定义配置。
//
// 返回值：
//   - config: 默认配置对象，包含预设的标题、链接等信息
//
// 使用示例：
//   config := DefaultFeedConfig()
//   config.Title = "我的博客" // 自定义标题
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

// FormatRFC822Time 将时间格式化为 RFC 822 格式。
//
// RSS 2.0 规范要求时间使用 RFC 822 格式，如 "Wed, 02 Oct 2024 08:00:00 +0000"。
// 此函数将 Go 的 time.Time 转换为规范要求的字符串格式。
//
// 参数：
//   - t: 时间对象，使用任意时区
//
// 返回值：
//   - 格式化后的时间字符串，空时间返回空字符串
//
// 使用示例：
//   pubDate := FormatRFC822Time(post.CreatedAt)
func FormatRFC822Time(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC822)
}