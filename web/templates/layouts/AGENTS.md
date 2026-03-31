<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# layouts

## Purpose
页面布局模板，定义整体页面框架结构。

## Key Files
| File | Description |
|------|-------------|
| `base.templ` | 前台基础布局：HTML 结构、CSS/JS 引用 |
| `admin.templ` | 后台管理布局：侧边栏、导航 |

## For AI Agents

### Working In This Directory
- 布局组件接收子内容 `templ.Component` 类型
- 包含公共资源引用（CSS、JS）

### Base Layout Structure
```templ
templ BaseLayout(title string, content templ.Component) {
    <!DOCTYPE html>
    <html lang="zh-CN">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <title>{ title }</title>
            <link rel="stylesheet" href="/static/dist/styles.css"/>
        </head>
        <body>
            @Header()
            <main>
                @content
            </main>
            @Footer()
            <script src="/static/dist/main.js"></script>
        </body>
    </html>
}
```

### Admin Layout Structure
```templ
templ AdminLayout(title string, content templ.Component) {
    <!DOCTYPE html>
    <html lang="zh-CN">
        <head>...</head>
        <body class="admin-body">
            <aside>@AdminSidebar()</aside>
            <main>@content</main>
            <script src="/static/dist/admin/main.js"></script>
        </body>
    </html>
}
```

### Component Slots
| Pattern | Description |
|---------|-------------|
| `{ children... }` | 子内容插槽 |
| `@component` | 引用子组件 |
| `templ.Component` | 组件类型参数 |

<!-- MANUAL: -->