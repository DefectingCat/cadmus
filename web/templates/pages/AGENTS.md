<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# pages

## Purpose
页面组件模板，定义各个页面视图。

## Key Files
| File | Description |
|------|-------------|
| `home.templ` | 首页 |
| `post_list.templ` | 文章列表页 |
| `post_detail.templ` | 文章详情页 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `admin/` | 管理后台页面 (see `admin/AGENTS.md`) |

## For AI Agents

### Working In This Directory
- 页面组件通常由 Handler 调用
- 数据通过参数传递，不在此处查询数据库

### Page Components
| Component | Route | Description |
|-----------|-------|-------------|
| `HomePage` | `/` | 首页 |
| `PostListPage` | `/posts` | 文章列表 |
| `PostDetailPage` | `/posts/{slug}` | 文章详情 |

### Component Pattern
```templ
// 页面组件接收数据和分页信息
templ PostListPage(title string, posts []*Post, pagination Pagination, categories []*Category, tags []*Tag) {
    @layouts.BaseLayout(title) {
        <div class="container">
            <div class="post-list">
                for _, post := range posts {
                    @PostCard(post)
                }
            </div>
            @PaginationNav(pagination)
        </div>
    }
}
```

### Pagination Structure
```go
type Pagination struct {
    CurrentPage  int
    TotalPages   int
    TotalItems   int
    ItemsPerPage int
}
```

<!-- MANUAL: -->