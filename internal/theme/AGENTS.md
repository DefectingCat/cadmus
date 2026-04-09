<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# internal/theme - 主题注册框架

## Purpose

`theme` 目录实现 Cadmus 的主题引擎，提供主题注册、切换和渲染的核心框架。使用 `templ` 库实现类型安全的模板渲染，主题通过实现 `ThemeComponents` 接口提供各个页面组件的渲染逻辑。

## Key Files

| File | Description |
|------|-------------|
| `models.go` | 主题核心数据结构：`Theme` 元信息、`ThemeComponents` 接口、`ThemeError` 错误类型 |
| `registry.go` | 主题注册表实现：`ThemeRegistry` 接口、全局注册函数、并发安全的主题管理 |

## Architecture

### 核心接口

```
┌─────────────────────────────────────────────────────────┐
│                    ThemeComponents                       │
├─────────────────────────────────────────────────────────┤
│  Layout()     templ.Component  - 整体布局框架            │
│  Header()     templ.Component  - 头部导航栏              │
│  Footer()     templ.Component  - 底部区域               │
│  PostList()   templ.Component  - 文章列表页             │
│  PostDetail() templ.Component  - 文章详情页             │
│  CategoryPage() templ.Component - 分类页                │
│  TagPage()    templ.Component  - 标签页                 │
│  Sidebar()    templ.Component  - 侧边栏（可选）          │
│  ErrorPage()  templ.Component  - 错误页面               │
└─────────────────────────────────────────────────────────┘
```

### 主题注册流程

```
主题包 init() → theme.Register(Theme) → globalRegistry.Register()
                                          │
                                          ▼
                                  map[string]Theme (并发安全)
```

### 主题切换流程

```
theme.SetActive("dark") → registry.SetActive() → r.activeThemeID = "dark"
                                │
                                ▼
                        theme.GetActive() → 返回 "dark" 主题的 Components
```

## Theme 接口详解

### 必须实现的方法

| 方法 | 说明 | 返回值要求 |
|------|------|------------|
| `Layout()` | 提供页面基础 HTML 结构，包含 head、header、main、footer | 必须返回有效的 `templ.Component` |
| `Header()` | 渲染顶部导航栏，包含 logo、菜单、用户信息 | 必须返回有效的 `templ.Component` |
| `Footer()` | 渲染底部区域，包含版权信息、链接 | 必须返回有效的 `templ.Component` |
| `PostList()` | 渲染文章列表页面，包含卡片、分页 | 必须返回有效的 `templ.Component` |
| `PostDetail()` | 渲染文章详情页面，包含标题、内容、评论 | 必须返回有效的 `templ.Component` |
| `CategoryPage()` | 渲染分类列表或分类下文章列表 | 必须返回有效的 `templ.Component` |
| `TagPage()` | 渲染标签云或标签下文章列表 | 必须返回有效的 `templ.Component` |
| `Sidebar()` | 渲染侧边栏区域（分类、标签、归档） | 可返回 `nil` 表示无侧边栏 |
| `ErrorPage()` | 渲染 404、500 等错误页面 | 必须返回有效的 `templ.Component` |

### 开发自定义主题的步骤

1. **创建主题结构**：在独立目录中创建 `theme.go` 文件
2. **实现 ThemeComponents 接口**：实现所有必需的组件方法
3. **在 init() 中注册**：使用 `theme.Register()` 注册主题

```go
package mytheme

import (
    "github.com/a-h/templ"
    "cadmus/internal/theme"
)

func init() {
    if err := theme.Register(New()); err != nil {
        log.Printf("主题注册失败：%v", err)
    }
}

func New() theme.Theme {
    return theme.Theme{
        ID:          "mytheme",
        Name:        "我的主题",
        Version:     "1.0.0",
        Author:      "开发者名称",
        Description: "主题描述",
        Components:  &myThemeComponents{},
    }
}

type myThemeComponents struct{}

func (c *myThemeComponents) Layout() templ.Component {
    // 实现布局逻辑
}

// ... 实现其他所有方法
```

## For AI Agents

### 开发指南

1. **新增主题**：在 `plugins/theme/` 目录下创建新主题包，实现 `ThemeComponents` 接口
2. **修改现有主题**：定位主题包目录，修改对应的组件实现
3. **切换主题**：使用 `theme.SetActive(themeID)` 切换当前主题
4. **获取主题列表**：使用 `theme.All()` 遍历所有已注册主题

### 使用示例

```go
// 获取所有主题
themes := theme.All()
for _, t := range themes {
    fmt.Printf("主题：%s (版本：%s)\n", t.Name, t.Version)
}

// 切换主题
if err := theme.SetActive("dark"); err != nil {
    log.Printf("切换失败：%v", err)
}

// 获取当前主题
current, err := theme.GetActive()
if err != nil {
    // 使用默认主题逻辑
}

// 渲染主题组件
component := current.Components.Header()
err = component.Render(ctx, w)
```

### 注意事项

- 主题必须在 `init()` 函数中自动注册
- 所有 `ThemeComponents` 方法应该返回可复用的组件实例
- `Sidebar()` 方法可返回 `nil` 表示主题不支持侧边栏
- 注册表操作是并发安全的，但主题组件本身应该是不可变的
- 使用 `templ` 语法编写模板，享受类型安全的渲染

## Dependencies

| Package | Version | 用途 |
|---------|---------|------|
| `github.com/a-h/templ` | (最新版) | 类型安全的模板引擎 |
| `sync` | stdlib | 读写锁保证并发安全 |

## Subdirectories

无子目录。
