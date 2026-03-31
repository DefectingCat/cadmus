<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# plugin

## Purpose
插件引擎模块，提供插件注册和管理功能。

## Key Files
| File | Description |
|------|-------------|
| `registry.go` | 插件注册表：RegisterPlugin、GetPlugin、AllPlugins |
| `plugin.go` | Plugin 接口定义和 PluginContext |
| `doc.go` | 包文档和使用说明 |

## For AI Agents

### Working In This Directory
- **编译时注册模式**，不使用 Go 原生 plugin 包
- 插件通过 blank import 触发 `init()` 自动注册
- 支持依赖验证

### Plugin Interface
```go
type Plugin interface {
    Info() PluginInfo
    Init(ctx *PluginContext) error
}

type PluginInfo struct {
    ID          string
    Name        string
    Version     string
    Author      string
    Description string
    Dependencies []string
}

type PluginContext struct {
    DB       *sql.DB
    Cache    CacheService
    Registry BlockTypeRegistry
    Services *ServicesRegistry
    Config   map[string]any
}
```

### Extension Interfaces (待实现)
| Interface | Description | Use Case |
|-----------|-------------|----------|
| `BlockType` | 自定义内容块 | mermaid-block, code-block |
| `AuthProvider` | 第三方登录 | github-auth, google-auth |
| `NotificationChannel` | 通知渠道 | discord-notify, slack-notify |
| `CommentFilter` | 评论过滤 | spam-detection |
| `PostRenderer` | 文章渲染扩展 | syntax-highlight |

### Registration Functions
| Function | Description |
|----------|-------------|
| `RegisterPlugin(ctor)` | 注册插件构造函数 |
| `RegisterPluginWithContext(ctor, ctx)` | 注册并初始化 |
| `GetPlugin(id)` | 获取插件实例 |
| `AllPlugins()` | 列出所有插件元信息 |
| `HasPlugin(id)` | 检查插件是否已注册 |

<!-- MANUAL: -->