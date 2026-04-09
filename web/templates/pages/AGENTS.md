<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# pages - 页面模板目录

## 用途

`pages/` 目录存放具体的页面视图模板，每个页面对应一个完整的 HTTP 响应视图。这些模板使用 Templ 引擎编写，编译时生成类型安全的 Go 代码。

页面模板负责：
- 组合布局（`layouts/`）和组件（`partials/`）构建完整页面
- 接收 Handler 传递的类型安全数据
- 渲染页面内容和动态数据
- 处理条件显示（空状态、权限、状态等）

## 关键文件

| 文件 | 功能 | 核心组件 |
|------|------|----------|
| `home.templ` | 首页视图 | `HomePage(title string)` |
| `post_list.templ` | 文章列表页 | `PostListPage()`, `PostCard()`, `PaginationNav()` |
| `post_detail.templ` | 文章详情页 | `PostDetailPage()`, `RelatedPostCard()`, `PostMetaSEO()` |

## 文件详解

### home.templ

首页模板，展示平台概览。

```templ
package pages

import "rua.plus/cadmus/web/templates/layouts"

templ HomePage(title string) {
    @layouts.BaseLayout(title) {
        @layouts.Header()
        <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <h1 class="text-3xl font-bold text-gray-900 mb-4">欢迎来到 Cadmus</h1>
            <p class="text-gray-600">一个现代化的多用户博客平台。</p>
        </main>
        @layouts.Footer()
    }
}
```

### post_list.templ

文章列表页模板，包含文章卡片列表、分类侧边栏、标签云和分页导航。

**核心组件：**

```templ
// 数据结构
type Pagination struct {
    CurrentPage int
    TotalPages  int
    TotalItems  int
    PerPage     int
}

// PostListPage 主页面组件
templ PostListPage(
    title string,
    posts []*post.Post,
    pagination Pagination,
    categories []*post.Category,
    tags []*post.Tag
)

// PostCard 文章卡片
templ PostCard(p *post.Post)

// PaginationNav 分页导航
templ PaginationNav(p Pagination)
```

**功能特性：**
- 文章卡片网格展示（含特色图片、标题、摘要、元信息）
- 侧边栏显示分类和标签
- 分页导航（上一页/下一页/页码）
- 空状态处理（无文章时显示提示）

### post_detail.templ

文章详情页模板，展示完整文章内容和相关信息。

**核心组件：**

```templ
// PostDetailPage 主页面组件
templ PostDetailPage(
    p *post.Post,
    author *user.User,
    tags []*post.Tag,
    relatedPosts []*post.Post
)

// RelatedPostCard 相关文章卡片
templ RelatedPostCard(p *post.Post)

// PostMetaSEO SEO 元信息注入
templ PostMetaSEO(p *post.Post)
```

**功能特性：**
- 文章头部（特色图片、标题、作者、元信息、标签）
- 文章内容区域（支持 JSON 内容和纯文本）
- 底部操作（点赞、分享、版本信息）
- 相关文章推荐
- 付费内容提示（如适用）
- SEO 元信息注入（Open Graph）

## 子目录

### admin/

管理后台页面目录，包含后台管理相关的页面模板。

| 文件 | 功能 |
|------|------|
| `dashboard.templ` | 仪表盘页面 |
| `posts/` | 文章管理页面 |
| `comments/` | 评论管理页面 |
| `users/` | 用户管理页面 |
| `settings/` | 设置页面 |

详见 [admin/AGENTS.md](./admin/AGENTS.md)

## 开发指南

### 创建新页面

```templ
package pages

import (
    "rua.plus/cadmus/web/templates/layouts"
    // 其他导入
)

// NewPage 新页面组件
templ NewPage(title string, data *PageData) {
    @layouts.BaseLayout(title) {
        @layouts.Header()
        <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <!-- 页面内容 -->
        </main>
        @layouts.Footer()
    }
}
```

### 页面组件结构

每个页面模板应遵循以下结构：

```
1. package 声明和 imports
2. 数据结构定义（如需要）
3. 主页面组件（Page 后缀）
4. 子组件（按功能命名）
```

### 数据传递模式

```go
// Handler 中
data := &PageData{
    Title: "页面标题",
    Items: items,
}
pages.NewPage(data).Render(ctx, w)
```

```templ
// 模板中
templ NewPage(data *PageData) {
    @layouts.BaseLayout(data.Title) {
        for _, item := range data.Items {
            @ItemCard(item)
        }
    }
}
```

### 条件渲染模式

```templ
// 空状态处理
if len(items) == 0 {
    <div class="text-center py-12">
        <p class="text-gray-500">暂无数据</p>
    </div>
} else {
    <div class="grid gap-4">
        for _, item := range items {
            @ItemCard(item)
        }
    </div>
}

// 状态显示
switch status {
case "published":
    <span class="text-green-500">已发布</span>
case "draft":
    <span class="text-yellow-500">草稿</span>
default:
    <span class="text-gray-500">未知</span>
}
```

### 分页实现

```templ
// 分页导航组件
templ PaginationNav(current int, total int) {
    if total <= 1 {
        return
    }
    <nav class="flex justify-center items-center space-x-2">
        // 上一页
        if current > 1 {
            <a href="?page={ current - 1 }" class="px-4 py-2">上一页</a>
        }
        
        // 页码
        for i := 1; i <= total; i++ {
            if i == current {
                <span class="px-4 py-2 bg-blue-600 text-white">{ i }</span>
            } else {
                <a href="?page={ i }" class="px-4 py-2">{ i }</a>
            }
        }
        
        // 下一页
        if current < total {
            <a href="?page={ current + 1 }" class="px-4 py-2">下一页</a>
        }
    </nav>
}
```

### 组件复用

将通用 UI 提取到 `partials/` 目录：

```templ
// pages/post_list.templ
@PostCard(post)

// partials/post_card.templ
templ PostCard(p *post.Post) {
    <article class="bg-white rounded-lg shadow-sm">
        <!-- 卡片内容 -->
    </article>
}
```

## 编译与测试

### 生成 Go 代码

```bash
# 在项目根目录
templ generate -path ./web/templates/pages

# Watch 模式
templ generate -path ./web/templates/pages --watch
```

### 验证编译

```bash
# 检查生成的代码
go build ./...

# 运行测试
go test ./web/templates/...
```

## 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 文件名 | `snake_case.templ` | `post_list.templ` |
| 组件名 | `PascalCase` | `PostListPage` |
| 页面组件 | `{PageName}Page` | `HomePage`, `DetailPage` |
| 子组件 | 描述性功能名 | `PostCard`, `PaginationNav` |

## 与 Router 集成

```go
// cmd/server/main.go 或 router 文件
mux.Handle("GET /", pages.HomePage("Cadmus"))
mux.Handle("GET /posts", handlers.PostListHandler)
mux.Handle("GET /posts/{slug}", handlers.PostDetailHandler)
```

```go
// handlers/post_list.go
func PostListHandler(w http.ResponseWriter, r *http.Request) {
    posts, _ := postService.List()
    categories, _ := postService.GetCategories()
    tags, _ := postService.GetTags()
    
    page := pages.PostListPage(title, posts, pagination, categories, tags)
    page.Render(r.Context(), w)
}
```

## 注意事项

1. **保持模板简洁** - 仅负责渲染，业务逻辑放在 Handler 或服务层
2. **类型安全** - 使用强类型结构体传递数据，避免 `interface{}`
3. **组件粒度** - 复杂页面拆分为多个子组件
4. **空状态** - 始终处理空数据状态
5. **响应式** - 使用 Tailwind 响应式类名（`sm:`, `md:`, `lg:`）
6. **SEO** - 在详情页等页面使用 `PostMetaSEO` 组件注入元信息
7. **不要编辑生成文件** - `*_templ.go` 文件由 `templ generate` 自动生成

## 参考资源

- [Templ 官方文档](https://templ.guide)
- [父目录文档](../AGENTS.md)
- [布局模板](../layouts/AGENTS.md)
- [组件模板](../partials/AGENTS.md)
