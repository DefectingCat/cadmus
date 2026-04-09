<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# plugins

## Purpose
可插拔扩展模块目录，存放 Cadmus 平台的插件实现。插件通过编译时注册模式集成到主程序，支持在 `cmd/server/main.go` 中使用 blank import 启用。

## Key Files
| File | Description |
|------|-------------|
| `doc.go` | 包文档，说明插件系统的集成方式 |
| `README.md` | 插件目录说明和插件结构规范 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `mermaid-block/` | Mermaid 图表块插件，支持流程图等渲染 (see `mermaid-block/AGENTS.md`) |

## For AI Agents

### 插件开发指南

#### Blank Import 注册机制
插件系统使用 Go 的 `init()` 函数自动注册模式。启用插件只需在 `cmd/server/main.go` 中添加 blank import：

```go
import (
    _ "rua.plus/cadmus/plugins/mermaid-block"
)
```

此 import 会触发插件包中的 `init()` 函数执行，自动将插件注册到全局注册表。

#### 插件结构
每个插件应遵循以下结构：

```
plugin-name/
├── plugin.go      # 插件主实现：Plugin 接口 + init() 注册
├── AGENTS.md      # 插件说明文档
└── hooks/         # 钩子实现（可选）
```

#### Plugin 接口
所有插件必须实现 `plugin.Plugin` 接口：

```go
type Plugin interface {
    Info() PluginInfo  // 返回插件元信息
    Init(ctx *PluginContext) error  // 初始化插件
}
```

#### PluginInfo 元数据
```go
type PluginInfo struct {
    ID          string   // 唯一标识符
    Name        string   // 显示名称
    Version     string   // 语义化版本号
    Author      string   // 作者信息
    Description string   // 功能描述
    Dependencies []string // 依赖插件 ID 列表
}
```

#### 开发步骤
1. 在 `plugins/` 下创建新目录（使用 `kebab-case` 命名）
2. 实现 `plugin.go`，定义插件结构体并实现 `Plugin` 接口
3. 添加 `init()` 函数，调用 `plugin.RegisterPlugin()` 注册
4. 在 `cmd/server/main.go` 添加 blank import 启用插件
5. 编写 `AGENTS.md` 文档说明插件用途和用法

#### 示例代码
```go
package myplugin

import (
    "rua.plus/cadmus/internal/plugin"
)

type MyPlugin struct{}

func (p *MyPlugin) Info() plugin.PluginInfo {
    return plugin.PluginInfo{
        ID:          "my-plugin",
        Name:        "我的插件",
        Version:     "1.0.0",
        Author:      "Cadmus Team",
        Description: "插件功能描述",
        Dependencies: []string{},
    }
}

func (p *MyPlugin) Init(ctx *plugin.PluginContext) error {
    // 初始化逻辑
    return nil
}

func init() {
    plugin.RegisterPlugin(func() plugin.Plugin {
        return &MyPlugin{}
    })
}
```

### Architecture Notes
- **编译时注册**: 插件在编译阶段通过 blank import 引入，运行时由 `init()` 自动注册
- **无插件管理**: 当前版本不支持动态加载/卸载，需重新编译启用
- **依赖顺序**: 有依赖的插件需确保依赖项先注册（通过 import 顺序控制）
