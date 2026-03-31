<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# themes

## Purpose
主题系统目录，存放可切换的前端主题模块。

## Key Files
| File | Description |
|------|-------------|
| `README.md` | 主题开发说明 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `default/` | 默认主题 (see `default/AGENTS.md`) |

## For AI Agents

### Working In This Directory
- **加载机制**: Blank import 触发 `init()` 自动注册
- **注册位置**: `internal/theme/registry.go` 的 `RegisterTheme`
- 新增主题需在 `cmd/server/main.go` 中添加 blank import

### Theme Development Pattern
```go
// themes/my-theme/theme.go
package my_theme

import "rua.plus/cadmus/internal/theme"

func init() {
    theme.RegisterTheme(theme.Theme{
        ID: "my-theme",
        Name: "My Theme",
        Components: MyComponents{},
    })
}
```

### Theme Components Interface
| Method | Description |
|--------|-------------|
| `Layout()` | 整体布局框架 |
| `Header()` | 页面头部 |
| `Footer()` | 页面底部 |
| `PostList()` | 文章列表页 |
| `PostDetail()` | 文章详情页 |
| `ErrorPage()` | 错误页 |

<!-- MANUAL: -->