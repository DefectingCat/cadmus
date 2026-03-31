<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# templates

## Purpose
Templ 模板目录，定义页面布局和组件。

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `layouts/` | 页面布局模板 (see `layouts/AGENTS.md`) |
| `pages/` | 页面组件 (see `pages/AGENTS.md`) |
| `partials/` | 可复用片段组件（待创建） |

## For AI Agents

### Working In This Directory
- `.templ` 文件通过 `templ generate` 生成 `*_templ.go`
- 生成的 Go 文件不应手动编辑
- 模板是类型安全的，编译时检查

### Templ Syntax Basics
```templ
// 定义组件
templ HomePage(title string) {
    <html>
        <head>
            <title>{ title }</title>
        </head>
        <body>
            <h1>Hello, World!</h1>
        </body>
    </html>
}

// 条件渲染
if user != nil {
    <span>{ user.Name }</span>
}

// 循环
for _, item := range items {
    <li>{ item.Name }</li>
}

// 引用子组件
@layouts.BaseLayout(title) {
    <main>Content</main>
}
```

### Build Commands
```bash
# 生成所有 templ 文件
templ generate

# watch 模式
templ generate --watch
```

### File Naming Convention
| Pattern | Example |
|---------|---------|
| `{name}.templ` | `base.templ` |
| 生成的 Go 文件 | `base_templ.go` |

<!-- MANUAL: -->