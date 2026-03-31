<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# admin

## Purpose
管理后台页面模板，包含仪表盘和各项管理功能页面。

## Key Files
| File | Description |
|------|-------------|
| `dashboard.templ` | 仪表盘页面：统计数据、最近内容 |
| `admin_posts.templ` | 文章管理页面 |
| `admin_post_edit.templ` | 文章编辑页面 |
| `admin_comments.templ` | 评论管理页面 |
| `admin_media.templ` | 媒体管理页面 |
| `admin_plugins.templ` | 插件管理页面 |
| `admin_themes.templ` | 主题管理页面 |

## For AI Agents

### Working In This Directory
- 所有页面使用 `AdminLayout` 布局
- 数据由 Handler 传递，模板仅负责渲染

### Admin Pages
| Page | Route | Description |
|------|-------|-------------|
| Dashboard | `/admin` | 统计概览 |
| Posts | `/admin/posts` | 文章列表 |
| Post Edit | `/admin/posts/{id}/edit` | 文章编辑 |
| Comments | `/admin/comments` | 评论审核 |
| Media | `/admin/media` | 媒体管理 |
| Plugins | `/admin/plugins` | 插件配置 |
| Themes | `/admin/themes` | 主题切换 |

### Dashboard Stats Structure
```go
type DashboardStats struct {
    TotalPosts    int
    TotalViews    int
    TotalComments int
    TotalUsers    int
    ActiveTheme   string
}
```

### Page Component Pattern
```templ
templ DashboardPage(stats DashboardStats, recentPosts []*Post, recentComments []*Comment, pendingCount int) {
    @layouts.AdminLayout("仪表盘") {
        <div class="dashboard">
            <div class="stats-grid">
                <div class="stat-card">
                    <h3>文章总数</h3>
                    <span class="stat-value">{ stats.TotalPosts }</span>
                </div>
                <!-- ... more stats -->
            </div>
            <div class="recent-content">
                @RecentPostsTable(recentPosts)
                @RecentCommentsTable(recentComments)
            </div>
        </div>
    }
}
```

### JavaScript Integration
| Page | JS Module | Data Attributes |
|------|-----------|-----------------|
| Posts | `admin/posts/list` | `data-page="admin-posts"` |
| Comments | `admin/comments/list` | `data-page="admin-comments"` |
| Media | `admin/media/upload`, `admin/media/list` | `data-page="admin-media"` |

<!-- MANUAL: -->