<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# rss - RSS 订阅领域模型

## Purpose

`rss` 目录包含 Cadmus 系统 RSS 订阅功能的核心领域模型定义。该模块负责：

- **RSS 2.0 订阅源生成**：定义符合 RSS 2.0 规范的订阅源数据结构
- **订阅源配置**：提供可自定义的订阅源元信息配置
- **时间格式转换**：将 Go 时间类型转换为 RSS 规范的 RFC 822 格式

该模块**仅包含模型定义**，Repository 接口位于 `post.Repository` 中（通过 `GetRSSFeed` 方法提供 RSS 数据获取能力）。

## Key Files

| File | Purpose |
|------|---------|
| `models.go` | RSS 领域模型定义：`RSSFeed`、`RSSChannel`、`RSSItem`、`FeedConfig` 及工具函数 |

## Domain Models

### RSSFeed
RSS 2.0 订阅源根结构，包含版本声明和频道信息。
```go
type RSSFeed struct {
    XMLName xml.Name    // XML 元素名称，自动设置为 "rss"
    Version string      // 版本号，固定为 "2.0"
    Channel *RSSChannel // 频道信息
}
```

### RSSChannel
RSS 2.0 频道结构，包含订阅源的元数据和文章列表。
```go
type RSSChannel struct {
    Title       string     // 频道标题
    Link        string     // 频道链接
    Description string     // 频道描述
    Language    string     // 语言标识 (如 "zh-CN")
    PubDate     string     // 发布时间 (RFC 822 格式)
    LastBuildDate string   // 最后构建时间
    Generator   string     // 生成器标识
    Items       []*RSSItem // 文章列表
}
```

### RSSItem
RSS 2.0 文章项结构，表示单个文章条目。
```go
type RSSItem struct {
    Title       string // 文章标题
    Link        string // 文章链接
    Description string // 文章摘要或描述
    Author      string // 作者信息
    Category    string // 文章分类
    PubDate     string // 发布时间 (RFC 822 格式)
    GUID        string // 文章唯一标识
}
```

### FeedConfig
RSS 订阅源配置结构，用于自定义订阅源元信息。
```go
type FeedConfig struct {
    Title       string // 频道标题
    Link        string // 频道链接
    Description string // 频道描述
    Language    string // 语言标识
    Generator   string // 生成器标识
    BaseURL     string // 文章链接基础 URL
}
```

## Utility Functions

### FormatRFC822Time
将 `time.Time` 转换为 RSS 2.0 规范要求的 RFC 822 格式。
```go
func FormatRFC822Time(t time.Time) string
```

### DefaultFeedConfig
返回默认订阅源配置，适用于 Cadmus 博客平台默认场景。
```go
func DefaultFeedConfig() FeedConfig
```

## For AI Agents

### 开发指南

1. **扩展模型字段**
   - 如需添加新的 RSS 2.0 元素（如 `<image>`、`<cloud>`），在 `RSSChannel` 中添加对应字段
   - 使用 `xml:",attr"` 标签定义 XML 属性，使用 `xml:"element"` 定义子元素
   - 可选字段使用 `omitempty` 标签

2. **遵循 RSS 2.0 规范**
   - 参考官方规范：https://www.rssboard.org/rss-specification
   - 时间格式必须使用 `FormatRFC822Time` 转换为 RFC 822 格式
   - GUID 应使用文章链接作为唯一标识

3. **使用 XML 序列化**
   - 生成 RSS XML 时使用 Go 标准库 `encoding/xml`
   - 确保所有字段的 `xml` 标签正确映射到 RSS 2.0 元素名称

4. **配置自定义**
   - 使用 `DefaultFeedConfig()` 获取默认配置
   - 修改配置字段以自定义订阅源标题、链接、描述等元信息

### 依赖关系

- **无内部依赖**：`rss` 模块不依赖其他 `core/` 子目录
- **外部依赖**：仅依赖 Go 标准库 `encoding/xml` 和 `time`
- **数据源**：RSS 数据通过 `post.Repository.GetRSSFeed` 方法获取

### 注意事项

- RSS 2.0 要求时间使用 RFC 822 格式（如 "Wed, 02 Oct 2024 08:00:00 +0000"）
- `GUID` 字段应使用文章链接，RSS 阅读器用它判断文章是否已读
- 空时间值（`t.IsZero()`）应返回空字符串，避免在 XML 中生成无效时间
