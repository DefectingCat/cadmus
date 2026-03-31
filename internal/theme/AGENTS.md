<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# theme

## Purpose
主题引擎模块，提供主题注册和管理功能。

## Key Files
| File | Description |
|------|-------------|
| `registry.go` | 主题注册表：Register、GetActive、SetActive、All |
| `models.go` | Theme 结构和 ThemeComponents 接口定义 |

## For AI Agents

### Working In This Directory
- 主题通过 blank import 触发 `init()` 自动注册
- 单例模式的 ThemeRegistry
- 主题切换不重启服务

### Theme Interface
```go
type Theme struct {
    ID          string
    Name        string
    Version     string
    Author      string
    Description string
    Components  ThemeComponents
}

type ThemeComponents interface {
    Layout() templ.Component
    Header() templ.Component
    Footer() templ.Component
    PostList() templ.Component
    PostDetail() templ.Component
    CategoryPage() templ.Component
    TagPage() templ.Component
    Sidebar() templ.Component
    ErrorPage() templ.Component
}
```

### Registration Functions
| Function | Description |
|----------|-------------|
| `Register(theme)` | 注册主题 |
| `GetActive()` | 获取当前激活主题 |
| `SetActive(themeID)` | 切换主题 |
| `All()` | 列出所有主题 |

### Theme Development Pattern
```go
// themes/my-theme/theme.go
package mytheme

import "rua.plus/cadmus/internal/theme"

func init() {
    theme.Register(theme.Theme{
        ID:   "my-theme",
        Name: "My Custom Theme",
        Components: &MyComponents{},
    })
}

type MyComponents struct{}

func (c *MyComponents) Layout() templ.Component { ... }
func (c *MyComponents) PostDetail() templ.Component { ... }
// ... 实现其他方法
```

<!-- MANUAL: -->