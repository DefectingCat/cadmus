<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# admin - 管理后台页面模板目录

## 用途

`admin/` 目录存放 Cadmus 博客平台的管理后台页面模板。所有模板使用 Templ 引擎编写，编译时生成类型安全的 Go 代码。

这些页面模板提供完整的管理功能界面：
- 仪表盘：平台数据概览和快捷操作
- 文章管理：文章 CRUD、状态管理、分类标签
- 评论审核：评论审核、批量操作
- 媒体库：文件上传、浏览、管理
- 插件管理：插件启用/禁用
- 主题管理：主题切换和配置

## 关键文件

| 文件 | 功能 | 核心组件 |
|------|------|----------|
| `dashboard.templ` | 仪表盘页面 | `DashboardPage(stats, recentPosts, recentComments, pendingComments)` |
| `admin_posts.templ` | 文章列表页 | `PostListPage()`, `PostTableRow()`, `PostStatusLabel()`, `PostPagination()` |
| `admin_post_edit.templ` | 文章编辑页 | `PostEditPage()`, 辅助函数 (`getPageTitle`, `joinKeywordsStr`, `formatPriceStr`) |
| `admin_comments.templ` | 评论管理页 | `AdminCommentsPage()`, `CommentListItem()`, `CommentStatusTab()` |
| `admin_media.templ` | 媒体库页面 | `AdminMediaPage()`, `MediaItem()`, `formatFileSize()` |
| `admin_plugins.templ` | 插件管理页 | `AdminPluginsPage()`, `PluginItem()` |
| `admin_themes.templ` | 主题管理页 | `AdminThemesPage()`, `ThemeCard()` |

## 文件详解

### dashboard.templ

管理后台首页，展示平台核心指标和快捷操作。

**数据结构：**

```go
type DashboardStats struct {
    TotalPosts      int     // 文章总数
    PostsChange     int     // 本周新增
    TotalViews      int     // 总浏览量
    ViewsChange     int     // 浏览量增长百分比
    TotalComments   int     // 评论总数
    TotalUsers      int     // 用户总数
    UsersChange     int     // 本周新增用户
    StorageUsed     string  // 已用存储
    StorageTotal    string  // 总存储
    ActiveTheme     string  // 活跃主题
    ActivePlugins   int     // 活跃插件数
}
```

**主组件：**

```templ
templ DashboardPage(stats DashboardStats, recentPosts []*post.Post, recentComments []*comment.Comment, pendingComments int)
```

**页面结构：**
- 4 个统计卡片（文章、浏览、评论、用户）
- 快捷操作区（新建文章、上传媒体、审核评论、系统设置）
- 最近文章列表（表格形式）
- 最近评论动态
- 系统状态检查

---

### admin_posts.templ

文章管理列表页面，支持搜索、筛选、批量操作。

**主组件：**

```templ
templ PostListPage(
    posts []*post.Post,
    total int,
    page int,
    pageSize int,
    statuses []string,
    categories []*post.Category
)
```

**子组件：**

| 组件 | 功能 |
|------|------|
| `PostTableRow(p *post.Post)` | 渲染文章表格行 |
| `PostStatusLabel(status post.PostStatus)` | 渲染状态标签（已发布/草稿/定时/私密） |
| `PostPagination(total, page, pageSize)` | 分页导航 |

**功能特性：**
- 文章表格（标题、Slug、状态、浏览数、评论数、创建时间）
- 搜索框（按标题搜索）
- 筛选器（状态、分类）
- 批量操作栏（批量发布、批量删除）
- 加载状态覆盖层
- 空状态处理

**支持的状态：**

| 状态 | 显示 |
|------|------|
| `post.StatusPublished` | 已发布（绿色） |
| `post.StatusDraft` | 草稿（灰色） |
| `post.StatusScheduled` | 定时发布（蓝色） |
| `post.StatusPrivate` | 私密（黄色） |

---

### admin_post_edit.templ

文章编辑页面，支持新建和编辑模式。

**主组件：**

```templ
templ PostEditPage(
    p *post.Post,
    isNew bool,
    categories []*post.Category,
    tags []*post.Tag,
    series []*post.Series
)
```

**辅助函数：**

| 函数 | 功能 |
|------|------|
| `getPageTitle(isNew bool) string` | 返回"新建文章"或"编辑文章" |
| `joinKeywordsStr(keywords []string) string` | 连接 SEO 关键词 |
| `formatPriceStr(price *float64) string` | 格式化价格显示 |
| `isSelected(currentID, optionID string) bool` | 判断下拉框选中状态 |

**表单区域：**

1. **基本信息** - 标题、URL Slug、分类、状态、摘要、特色图片
2. **文章内容** - 内容编辑器（带字数统计）
3. **标签管理** - 现有标签展示、添加新标签
4. **SEO 设置** - SEO 标题、关键词、描述
5. **付费设置** - 付费文章开关、价格
6. **文章系列** - 所属系列、系列内排序（如有系列）

**功能特性：**
- 自动保存状态显示
- 快捷键（Ctrl+S 保存）
- 预览功能
- 版本历史（编辑模式下显示）
- 从媒体库选择特色图片

---

### admin_comments.templ

评论审核管理页面，支持状态标签页筛选和批量操作。

**主组件：**

```templ
templ AdminCommentsPage(
    title string,
    status string,
    comments []*comment.Comment,
    page int,
    perPage int,
    total int
)
```

**辅助函数：**

| 函数 | 功能 |
|------|------|
| `intToString(n int) string` | 整数转字符串用于 URL 拼接 |
| `getStatusClass(status string) string` | 获取状态标签样式类 |
| `getStatusText(status string) string` | 获取状态中文文本 |
| `getTabClass(currentStatus, tabStatus string) string` | 获取标签页样式 |

**状态标签页：**

| 状态 | URL 参数 | 显示文本 |
|------|----------|----------|
| 待审核 | `pending` | 待审核 |
| 已通过 | `approved` | 已通过 |
| 已拒绝 | `spam` | 已拒绝 |
| 已删除 | `deleted` | 已删除 |

**功能特性：**
- 状态标签页导航
- 批量操作工具栏（全选、批量通过/拒绝/删除）
- 确认对话框
- 单条评论操作（通过、拒绝、删除）

---

### admin_media.templ

媒体库管理页面，支持拖拽上传和网格浏览。

**数据结构：**

```go
type MediaPageData struct {
    Title   string
    Medias  []*media.Media
    Total   int
    Page    int
    PerPage int
    UserID  uuid.UUID
}
```

**主组件：**

```templ
templ AdminMediaPage(data MediaPageData)
```

**子组件：**

```templ
templ MediaItem(m *media.Media)
```

**功能特性：**
- 拖拽上传区域
- 文件类型筛选（图片/文档/其他）
- 网格视图（响应式：2/3/4/5 列）
- 多选功能
- 媒体详情弹窗
- 插入到文章功能
- 分页导航

**支持的文件类型：**
- 图片：JPEG, PNG, GIF, WebP, SVG
- 文档：PDF, DOC, DOCX, TXT
- 最大文件大小：10MB

---

### admin_plugins.templ

插件管理页面，显示已安装插件并支持启用/禁用。

**数据结构：**

```go
type PluginInfo struct {
    ID          string
    Name        string
    Version     string
    Author      string
    Description string
    Enabled     bool
}
```

**主组件：**

```templ
templ AdminPluginsPage(title string, plugins []PluginInfo)
```

**子组件：**

```templ
templ PluginItem(p PluginInfo)
```

**功能特性：**
- 插件列表（开关控制启用/禁用）
- 配置按钮（用于插件配置）
- 空状态提示
- 插件开发说明提示

---

### admin_themes.templ

主题管理页面，支持主题切换和配置。

**数据结构：**

```go
type ThemeInfo struct {
    ID          string
    Name        string
    Version     string
    Author      string
    Description string
    Active      bool
}
```

**主组件：**

```templ
templ AdminThemesPage(title string, themes []ThemeInfo)
```

**子组件：**

```templ
templ ThemeCard(t ThemeInfo)
```

**功能特性：**
- 主题卡片网格展示（3 列布局）
- 主题预览图
- 激活按钮/已启用状态
- 配置按钮
- 主题切换 API 调用（`PUT /api/v1/admin/themes/active`）

---

## 子目录

`admin/` 目录无子目录，所有管理后台页面模板均在同一层级。

## 开发指南

### 创建新的管理页面

```templ
package admin

import (
    "rua.plus/cadmus/internal/core/post"
    "rua.plus/cadmus/web/templates/layouts"
)

// NewAdminPage 新管理页面组件
templ NewAdminPage(title string, data *PageData) {
    @layouts.AdminLayout(title, "menu-key") {
        <!-- 顶部导航栏 -->
        <header class="bg-white shadow-sm border-b border-gray-200">
            <nav class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div class="flex justify-between h-16">
                    <div class="flex items-center">
                        <a href="/" class="text-xl font-bold text-gray-900">Cadmus Admin</a>
                    </div>
                    <div class="flex items-center space-x-4">
                        <a href="/admin/posts" class={ "text-gray-600 hover:text-gray-900", templ.KV("text-blue-600 font-medium", isActive) }>文章管理</a>
                        <a href="/admin/media" class="text-gray-600 hover:text-gray-900">媒体库</a>
                        <a href="/admin/comments" class="text-gray-600 hover:text-gray-900">评论审核</a>
                        <a href="/admin/plugins" class="text-gray-600 hover:text-gray-900">插件</a>
                        <a href="/admin/themes" class="text-gray-600 hover:text-gray-900">主题</a>
                    </div>
                </div>
            </nav>
        </header>

        <!-- 主内容区 -->
        <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <!-- 页面内容 -->
        </main>
    }
}
```

### 页面组件结构

每个管理页面模板应遵循以下结构：

```
1. package 声明 (package admin)
2. imports（核心模型、布局、工具函数）
3. 数据结构定义（如需要）
4. 主页面组件（{Feature}Page 后缀）
5. 子组件（按功能命名，如 {Feature}Item, {Feature}List）
6. 辅助函数（状态显示、格式化等）
```

### 顶部导航栏标准结构

所有管理页面应包含统一的顶部导航栏：

```templ
<header class="bg-white shadow-sm border-b border-gray-200">
    <nav class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
            <div class="flex items-center">
                <a href="/" class="text-xl font-bold text-gray-900">Cadmus Admin</a>
            </div>
            <div class="flex items-center space-x-4">
                <a href="/admin/posts" class={ "text-gray-600 hover:text-gray-900", templ.KV("text-blue-600 font-medium", isActive) }>文章管理</a>
                <a href="/admin/media" class="text-gray-600 hover:text-gray-900">媒体库</a>
                <a href="/admin/comments" class="text-gray-600 hover:text-gray-900">评论审核</a>
                <a href="/admin/plugins" class="text-gray-600 hover:text-gray-900">插件</a>
                <a href="/admin/themes" class="text-gray-600 hover:text-gray-900">主题</a>
            </div>
        </div>
    </nav>
</header>
```

### 状态标签模式

```templ
// 状态标签组件
templ StatusLabel(status string) {
    switch status {
    case "published":
        <span class="px-2 py-1 text-xs rounded-full bg-green-100 text-green-800">已发布</span>
    case "draft":
        <span class="px-2 py-1 text-xs rounded-full bg-gray-100 text-gray-800">草稿</span>
    case "pending":
        <span class="px-2 py-1 text-xs rounded-full bg-yellow-100 text-yellow-800">待审核</span>
    default:
        <span class="px-2 py-1 text-xs rounded-full bg-gray-100 text-gray-700">{ status }</span>
    }
}
```

### 批量操作模式

```templ
<!-- 批量操作栏 -->
<div id="bulk-actions" class="bg-blue-50 rounded-lg p-4 mb-4 flex items-center gap-4">
    <span class="text-blue-800">已选择 <span id="selected-count">0</span> 条</span>
    <button id="bulk-action-1" class="px-3 py-1 bg-green-600 text-white rounded hover:bg-green-700 text-sm">
        批量操作 1
    </button>
    <button id="bulk-action-2" class="px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700 text-sm">
        批量操作 2
    </button>
    <button id="bulk-cancel" class="px-3 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 text-sm">
        取消选择
    </button>
</div>
```

### 分页实现

```templ
// 分页导航组件
templ PaginationNav(current int, total int) {
    if total <= 1 {
        return
    }
    <nav class="flex justify-center items-center gap-2">
        // 上一页
        if current > 1 {
            <a href="?page={ current - 1 }" class="px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50">
                上一页
            </a>
        } else {
            <span class="px-4 py-2 bg-gray-100 border border-gray-300 rounded-lg text-gray-400 cursor-not-allowed">
                上一页
            </span>
        }
        
        // 页码
        for i := 1; i <= total; i++ {
            if i == current {
                <span class="px-4 py-2 bg-blue-600 text-white rounded-lg">{ i }</span>
            } else {
                <a href="?page={ i }" class="px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50">
                    { i }
                </a>
            }
        }
        
        // 下一页
        if current < total {
            <a href="?page={ current + 1 }" class="px-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50">
                下一页
            </a>
        } else {
            <span class="px-4 py-2 bg-gray-100 border border-gray-300 rounded-lg text-gray-400 cursor-not-allowed">
                下一页
            </span>
        }
    </nav>
}
```

### 空状态处理

```templ
if len(items) == 0 {
    <div class="text-center py-12 text-gray-500">
        <svg class="w-16 h-16 mx-auto text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <!-- 适当的图标 -->
        </svg>
        <p>暂无数据</p>
        <p class="text-sm mt-1">点击"新建"按钮添加第一条数据</p>
    </div>
} else {
    <!-- 正常渲染列表 -->
}
```

### 与后端交互模式

```templ
<!-- 后端数据注入到前端 JS -->
<script>
    window.pageData = {
        page: { data.Page },
        perPage: { data.PerPage },
        total: { data.Total },
        userId: "{ data.UserID.String() }"
    };
</script>
<script src="/static/dist/admin/feature.js"></script>
```

## 编译与测试

### 生成 Go 代码

```bash
# 在项目根目录执行
templ generate -path ./web/templates/pages/admin

# Watch 模式（开发时）
templ generate -path ./web/templates/pages/admin --watch
```

### 验证编译

```bash
# 检查生成的代码能否编译
go build ./...

# 运行模板测试
go test ./web/templates/...
```

## 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 文件名 | `snake_case.templ` | `admin_posts.templ`, `admin_media.templ` |
| 组件名 | `PascalCase` | `DashboardPage`, `PostListPage` |
| 页面组件 | `{Feature}Page` | `AdminCommentsPage`, `AdminThemesPage` |
| 子组件 | 描述性功能名 | `PostTableRow`, `MediaItem`, `ThemeCard` |
| 辅助函数 | `camelCase` | `formatFileSize`, `getStatusClass` |

## 与 Router 集成

```go
// cmd/server/main.go 或 router 文件
mux.Handle("GET /admin", handlers.DashboardHandler)
mux.Handle("GET /admin/posts", handlers.AdminPostListHandler)
mux.Handle("GET /admin/posts/new", handlers.AdminPostNewHandler)
mux.Handle("GET /admin/posts/{id}/edit", handlers.AdminPostEditHandler)
mux.Handle("GET /admin/comments", handlers.AdminCommentHandler)
mux.Handle("GET /admin/media", handlers.AdminMediaHandler)
mux.Handle("GET /admin/plugins", handlers.AdminPluginHandler)
mux.Handle("GET /admin/themes", handlers.AdminThemeHandler)
```

```go
// handlers/dashboard.go
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    stats := getDashboardStats(ctx)
    recentPosts, _ := postService.ListRecent(5)
    recentComments, _ := commentService.ListRecent(5)
    pendingComments, _ := commentService.CountPending()
    
    page := admin.DashboardPage(stats, recentPosts, recentComments, pendingComments)
    page.Render(ctx, w)
}
```

## 注意事项

1. **使用 AdminLayout** - 管理页面可选择使用 `@layouts.AdminLayout()` 统一布局
2. **保持模板简洁** - 仅负责渲染，业务逻辑放在 Handler 或服务层
3. **类型安全** - 使用强类型结构体传递数据，避免 `interface{}`
4. **组件复用** - 通用 UI 组件（状态标签、分页、批量操作栏）应提取为独立组件
5. **空状态处理** - 所有列表都应处理空数据状态
6. **响应式设计** - 使用 Tailwind 响应式类名（`sm:`, `md:`, `lg:`）
7. **不要编辑生成文件** - `*_templ.go` 文件由 `templ generate` 自动生成
8. **导航高亮** - 当前页面对应的导航项应添加 `text-blue-600 font-medium` 样式

## 参考资源

- [Templ 官方文档](https://templ.guide)
- [父目录文档](../AGENTS.md)
- [布局模板](../layouts/AGENTS.md)
- [组件模板](../partials/AGENTS.md)
- [项目设计文档](../../../../docs/design.md)
