<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# plugins

## Purpose
可插拔扩展模块目录，通过编译时注册模式加载插件。

## Key Files
| File | Description |
|------|-------------|
| `doc.go` | 插件系统文档 |
| `README.md` | 插件开发说明 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `mermaid-block/` | Mermaid 图表块插件 (see `mermaid-block/AGENTS.md`) |

## For AI Agents

### Working In This Directory
- **加载机制**: Blank import 触发 `init()` 自动注册
- **注册位置**: `internal/plugin/registry.go` 的 `RegisterPlugin`
- 新增插件需在 `cmd/server/main.go` 中添加 blank import

### Plugin Development Pattern
```go
// plugins/my-plugin/plugin.go
package my_plugin

import "rua.plus/cadmus/internal/plugin"

type MyPlugin struct {}

func (p *MyPlugin) Info() plugin.PluginInfo { ... }
func (p *MyPlugin) Init(ctx *plugin.PluginContext) error { ... }

func init() {
    plugin.RegisterPlugin(func() plugin.Plugin { return &MyPlugin{} })
}
```

### Plugin Types
| Type | Interface | Example |
|------|-----------|---------|
| BlockType | 自定义内容块 | mermaid-block |
| AuthProvider | 第三方登录 | github-auth (待开发) |
| NotificationChannel | 通知渠道 | discord-notify (待开发) |

<!-- MANUAL: -->