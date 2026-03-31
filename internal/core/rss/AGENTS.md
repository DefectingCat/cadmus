<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# rss

## Purpose
RSS 领域模型，定义 RSS Feed 结构。

## Key Files
| File | Description |
|------|-------------|
| `models.go` | Feed、Item 结构定义 |

## For AI Agents

### Working In This Directory
- 生成 RSS 2.0 格式 XML
- 支持自定义 Feed 配置

### Data Models
```go
type Feed struct {
    Title       string
    Link        string
    Description string
    Language    string
    Items       []Item
    LastBuild   time.Time
}

type Item struct {
    Title       string
    Link        string
    Description string
    PubDate     time.Time
    GUID        string
    Author      string
    Categories  []string
}

type FeedConfig struct {
    Title       string
    Link        string
    Description string
    Language    string
    ItemCount   int      // 默认 20
}
```

### RSS 2.0 Output
```xml
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Cadmus Blog</title>
    <link>https://example.com</link>
    <description>Latest posts</description>
    <item>
      <title>Post Title</title>
      <link>https://example.com/posts/slug</link>
      <description>Post excerpt...</description>
      <pubDate>Mon, 31 Mar 2026 00:00:00 +0000</pubDate>
      <guid>https://example.com/posts/slug</guid>
    </item>
  </channel>
</rss>
```

<!-- MANUAL: -->