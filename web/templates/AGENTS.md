<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# templates

## 用途

Templ 模板文件目录，包含使用 [Templ](https://templ.guide) 类型安全模板引擎编写的页面布局、页面组件和可复用片段。模板在编译时生成类型安全的 Go 代码。

## 目录结构

| 目录 | 用途 |
|------|------|
| `layouts/` | 页面布局模板（BaseLayout、AdminLayout） |
| `pages/` | 页面组件（首页、文章列表、文章详情、管理后台） |
| `partials/` | 可复用片段组件（导航、卡片、表格等） |

## 子目录说明

### layouts/

定义整体页面框架结构，包含：
- `base.templ` - 前台基础布局（HTML 结构、CSS/JS 引用、Header、Footer）
- `admin.templ` - 后台管理布局（侧边栏、导航）

详见 [layouts/AGENTS.md](./layouts/AGENTS.md)

### pages/

定义具体页面视图，包含：
- `home.templ` - 首页
- `post_list.templ` - 文章列表页
- `post_detail.templ` - 文章详情页
- `admin/` - 管理后台页面（仪表盘、文章管理、评论管理等）

详见 [pages/AGENTS.md](./pages/AGENTS.md)

### partials/

存放可复用的 UI 组件片段：
- 导航组件
- 卡片组件
- 表格组件
- 分页组件
- 表单组件

## Templ 语法指南

### 基本组件定义

```templ
// 定义模板组件，参数类型安全
templ ComponentName(param1 string, param2 int) {
    <div class="component">
        <h1>{ param1 }</h1>
        <p>Count: { param2 }</p>
    </div>
}
```

### 变量插值

```templ
// 简单的变量插值
<h1>{ title }</h1>
<p>{ user.Name }</p>

// 表达式插值
<span>{ len(items) } 项</span>
<span>{ strings.ToUpper(name) }</span>

// 属性插值
<div class={ className }></div>
<input type="text" value={ defaultValue }/>
```

### 条件渲染

```templ
// if/else
if user != nil {
    <span class="username">{ user.Name }</span>
} else {
    <span class="guest">访客</span>
}

// switch
switch status {
case "active":
    <span class="text-green-500">活跃</span>
case "pending":
    <span class="text-yellow-500">待审核</span>
default:
    <span class="text-gray-500">未知</span>
}
```

### 循环渲染

```templ
// for 循环
<ul>
    for _, item := range items {
        <li>{ item.Name }</li>
    }
</ul>

// 带索引的循环
for i, post := range posts {
    <article class={ "post-" + strconv.Itoa(i) }>
        <h2>{ post.Title }</h2>
    </article>
}
```

### 子组件调用

```templ
// 调用子组件
@Header()
@PostCard(post)
@Pagination(pagination)

// 带 children 的布局组件
@layouts.BaseLayout(title) {
    <main class="container">
        <h1>页面内容</h1>
        { children... }
    </main>
}
```

### 组件作为参数

```templ
// 接收 templ.Component 参数
templ CardWrapper(title string, content templ.Component) {
    <div class="card">
        <h3>{ title }</h3>
        <div class="card-body">
            @content
        </div>
    </div>
}

// 使用时
@CardWrapper("标题") {
    <p>自定义内容</p>
    <button>操作按钮</button>
}
```

### HTML 属性

```templ
// 标准属性
<div id="main" class="container" data-value={ value }>

// 条件属性
<button disabled={ !isEnabled }>提交</button>

// 动态类名
<div class={ "btn " + variant }>

// Tailwind 类名
<div class="flex items-center space-x-4 p-4 bg-white rounded-lg shadow">
```

### 脚本和样式

```templ
// 内联脚本（需要转义）
<script>
    console.log("初始化");
</script>

// 外部资源引用
<link rel="stylesheet" href="/static/dist/main.css"/>
<script src="/static/dist/main.js" defer></script>
```

## 开发指南

### 生成 Go 代码

```bash
# 在项目根目录运行
templ generate

# Watch 模式（开发时自动重新生成）
templ generate --watch

# 指定目录
templ generate -path ./web/templates
```

### 文件命名规范

| 模式 | 示例 | 生成文件 |
|------|------|----------|
| `{name}.templ` | `base.templ` | `base_templ.go` |
| `{name}_page.templ` | `home_page.templ` | `home_page_templ.go` |

### 包声明

每个 `.templ` 文件顶部需要声明包名：

```templ
package pages

import "rua.plus/cadmus/web/templates/layouts"

templ HomePage(title string) {
    @layouts.BaseLayout(title) {
        // 页面内容
    }
}
```

### 类型安全数据传递

```go
// 定义数据结构
type Post struct {
    ID        uint
    Title     string
    Content   string
    CreatedAt time.Time
}

type Pagination struct {
    CurrentPage  int
    TotalPages   int
    TotalItems   int
}
```

```templ
// 在模板中使用
templ PostList(posts []*Post, pagination Pagination) {
    for _, post := range posts {
        @PostCard(post.ID, post.Title, post.CreatedAt)
    }
    @PaginationNav(pagination.CurrentPage, pagination.TotalPages)
}
```

## 构建流程

```
web/templates/*.templ    → templ generate → *_templ.go    → Go 编译 → 可执行文件
```

### 完整构建命令

```bash
# 生成所有模板
templ generate

# 构建前端资源
cd web/frontend && bun run build:all

# 构建 Go 应用（包含生成的模板）
go build -o cadmus ./cmd/...
```

## 常见模式

### 页面组件模式

```templ
templ PageComponent(title string, data *PageData) {
    @layouts.BaseLayout(title) {
        @Header()
        <main class="container mx-auto px-4 py-8">
            <h1 class="text-3xl font-bold mb-6">{ title }</h1>
            
            // 条件渲染内容
            if data != nil && len(data.Items) > 0 {
                <div class="grid gap-4">
                    for _, item := range data.Items {
                        @ItemCard(item)
                    }
                </div>
            } else {
                <p class="text-gray-500">暂无数据</p>
            }
        </main>
        @Footer()
    }
}
```

### 管理后台模式

```templ
templ AdminPage(title string, content templ.Component) {
    @layouts.AdminLayout(title) {
        <div class="admin-container">
            <aside class="sidebar">
                @AdminSidebar()
            </aside>
            <main class="flex-1 p-6">
                @content
            </main>
        </div>
    }
}
```

### 分页导航模式

```templ
templ PaginationNav(current int, total int) {
    if total > 1 {
        <nav class="pagination">
            if current > 1 {
                <a href="?page={ current - 1 }">上一页</a>
            }
            <span>第 { current } / { total } 页</span>
            if current < total {
                <a href="?page={ current + 1 }">下一页</a>
            }
        </nav>
    }
}
```

## 注意事项

1. **不要手动编辑生成的 `*_templ.go` 文件** - 这些文件由 `templ generate` 自动生成
2. **模板保持简洁** - 业务逻辑应放在 Handler 或服务层，模板仅负责渲染
3. **使用类型安全参数** - 避免在模板中进行复杂的数据处理
4. **复用组件** - 将通用 UI 提取到 `partials/` 目录
5. **保持一致的命名** - 组件名使用 PascalCase，文件名使用 snake_case

## 参考资源

- [Templ 官方文档](https://templ.guide)
- [Templ 语法参考](https://templ.guide/syntax-and-usage/statements)
- [Templ 项目示例](https://github.com/a-h/templ/tree/main/examples)
