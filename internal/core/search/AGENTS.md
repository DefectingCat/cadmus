<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# search - 全文搜索领域模型

## Purpose

`search` 目录定义 Cadmus 系统全文搜索功能的核心领域模型。包含搜索数据访问接口和结果模型，用于：

- 抽象搜索数据访问层，支持多种后端实现（PostgreSQL 全文搜索、Elasticsearch、MeiliSearch 等）
- 定义搜索操作的标准接口和数据结构

该目录**不包含**具体实现，仅定义接口和类型。实现位于 `internal/database/postgres/` 目录下。

## Key Files

| File | Purpose |
|------|---------|
| `models.go` | 搜索结果实体、过滤条件、响应结构、错误类型定义 |
| `repository.go` | SearchRepository 接口定义 |

## Domain Models

### SearchResult

搜索结果项，包含文章的基本信息和相关性得分：

```go
type SearchResult struct {
    ID         uuid.UUID  // 文章 ID
    Title      string     // 文章标题
    Slug       string     // URL 别名
    Excerpt    string     // 摘要
    AuthorID   uuid.UUID  // 作者 ID
    CategoryID uuid.UUID  // 分类 ID（可选）
    Status     string     // 状态（通常为 "published"）
    Rank       float64    // 相关性得分
    CreatedAt  time.Time  // 创建时间
    UpdatedAt  time.Time  // 最后修改时间
}
```

### SearchFilters

搜索过滤条件，支持多条件组合：

```go
type SearchFilters struct {
    Query    string       // 搜索关键词（必填）
    Category string       // 分类 Slug
    AuthorID uuid.UUID    // 作者 ID
    Status   string       // 文章状态（默认 "published"）
    Tag      string       // 标签 Slug
}
```

### SearchResponse

搜索响应结构：

```go
type SearchResponse struct {
    Results  []SearchResult `json:"results"`  // 搜索结果列表
    Total    int            `json:"total"`    // 总结果数
    Page     int            `json:"page"`     // 当前页码
    PageSize int            `json:"page_size"`// 每页数量
    Query    string         `json:"query"`    // 原始搜索词
}
```

### SearchError

搜索模块自定义错误类型：

```go
type SearchError struct {
    Code    string  // 错误代码
    Message string  // 错误消息
}

// 预定义错误
var (
    ErrEmptyQuery   = &SearchError{Code: "empty_query", Message: "搜索关键词不能为空"}
    ErrQueryTooLong = &SearchError{Code: "query_too_long", Message: "搜索关键词过长"}
)
```

## SearchRepository 接口

```go
type SearchRepository interface {
    // Search 全文搜索文章
    Search(ctx context.Context, query string, filters SearchFilters, offset, limit int) ([]SearchResult, int, error)

    // SearchByCategory 在指定分类下搜索
    SearchByCategory(ctx context.Context, query string, categoryID uuid.UUID, offset, limit int) ([]SearchResult, int, error)

    // SearchByAuthor 搜索指定作者的文章
    SearchByAuthor(ctx context.Context, query string, authorID uuid.UUID, offset, limit int) ([]SearchResult, int, error)

    // GetSuggestions 获取搜索建议
    GetSuggestions(ctx context.Context, query string, limit int) ([]string, error)
}
```

## Subdirectories

无。搜索模块的所有定义均位于根目录下。

## For AI Agents

### 开发指南

1. **实现 SearchRepository 接口时**
   - 必须支持 context.Context 进行超时控制
   - 所有方法必须保证并发安全
   - 搜索结果必须按相关性得分（Rank）降序排列
   - 必须正确处理分页参数（offset/limit）和总数统计（total）

2. **错误处理**
   - 使用预定义的语义化错误类型：
     - `ErrEmptyQuery`: 搜索关键词为空时返回
     - `ErrQueryTooLong`: 关键词超过系统限制时返回
   - 所有错误实现 `error` 接口，支持 `errors.Is` 比较

3. **实现后端时的注意事项**
   - PostgreSQL 实现：使用 `tsvector`/`tsquery` 全文搜索
   - Elasticsearch 实现：使用 `_search` API，注意映射配置
   - MeiliSearch 实现：使用 `search` 方法，配置可搜索属性

4. **接口使用示例**
   ```go
   // 基本搜索
   results, total, err := repo.Search(ctx, "golang", SearchFilters{Status: "published"}, 0, 10)

   // 获取搜索建议
   suggestions, err := repo.GetSuggestions(ctx, "go", 5)
   // 可能返回：["golang", "google", "good"]
   ```

5. **与其他模块的集成**
   - 与 `post` 模块：搜索结果显示文章信息（Title、Slug、Excerpt）
   - 与 `user` 模块：支持按作者搜索（AuthorID 过滤）
   - 与 `post` 模块的分类/标签：支持按分类/标签筛选
