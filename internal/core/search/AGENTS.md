<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# search

## Purpose
搜索领域模型，定义搜索结果实体和 Repository 接口。

## Key Files
| File | Description |
|------|-------------|
| `models.go` | SearchResult、SearchFilters 结构定义 |
| `repository.go` | SearchRepository 接口定义 |

## For AI Agents

### Working In This Directory
- 使用 PostgreSQL 全文搜索 (pg_trgm)
- 支持标题、内容、摘要搜索

### Data Models
```go
type SearchResult struct {
    ID       uuid.UUID
    Title    string
    Excerpt  string
    Type     string   // post/page/comment
    Score    float64  // 相关度得分
    CreatedAt time.Time
}

type SearchFilters struct {
    Query    string
    Type     string   // 可选：限定类型
    AuthorID *uuid.UUID
    TagIDs   []uuid.UUID
    DateFrom *time.Time
    DateTo   *time.Time
}
```

### Search Implementation
- PostgreSQL `tsvector` + `tsquery`
- GIN 索引加速
- 权重：标题 A > 摘要 B > 内容 C

### Repository Interface
```go
type SearchRepository interface {
    Search(ctx context.Context, query string, filters SearchFilters, offset, limit int) ([]*SearchResult, int, error)
    Suggestions(ctx context.Context, prefix string, limit int) ([]string, error)
}
```

### SQL Example
```sql
SELECT id, title, excerpt, ts_rank(search_vector, to_tsquery('english', $1)) as score
FROM posts
WHERE search_vector @@ to_tsquery('english', $1)
  AND status = 'published'
ORDER BY score DESC
LIMIT 20;
```

<!-- MANUAL: -->