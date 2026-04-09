<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# internal/plugin - 插件注册框架

## 用途

`plugin` 目录实现编译时接口注册模式的插件系统，支持可插拔扩展。通过 Go 的 blank import 触发 `init()` 函数自动注册插件，避免使用 Go 原生 `plugin` 包的跨平台兼容性问题。

**核心职责**:
- 提供全局插件注册表（并发安全）
- 定义插件接口和运行时上下文
- 支持依赖检查和初始化注入
- 管理插件生命周期（注册、查询、遍历）

## 关键文件

| 文件 | 功能 |
|------|------|
| `plugin.go` | 定义 `Plugin` 接口、`PluginInfo` 元信息、`PluginContext` 上下文、`PluginConstructor` 构造函数 |
| `registry.go` | 实现全局注册表：`RegisterPlugin`、`GetPlugin`、`AllPlugins`、`PluginIDs`、`HasPlugin`、`Count` |
| `doc.go` | 包级别文档和注册表操作说明 |

## 无子目录

该目录为扁平结构，所有文件直接位于 `plugin/` 下。

## BlockPlugin 接口

插件通过实现 `BlockPlugin` 接口来注册自定义内容块类型：

```go
// BlockPlugin 扩展接口（需实现）
type BlockPlugin interface {
    Plugin
    // RegisterBlocks 返回插件提供的自定义块类型
    RegisterBlocks() []BlockType
}

// BlockType 自定义块类型定义
type BlockType struct {
    Name     string      // 块类型名称
    Schema   interface{} // 块数据结构定义
    Renderer func(data interface{}) (string, error) // 渲染函数
}
```

### 插件开发指南

**1. 定义插件结构**

```go
package myplugin

import (
    "rua.plus/cadmus/internal/plugin"
)

type MyPlugin struct {
    // 插件状态
}

func (p *MyPlugin) Info() plugin.PluginInfo {
    return plugin.PluginInfo{
        ID:          "my-plugin",
        Name:        "我的插件",
        Version:     "1.0.0",
        Author:      "作者名",
        Description: "插件功能描述",
        Dependencies: []string{}, // 依赖的其他插件 ID
    }
}

func (p *MyPlugin) Init(ctx *plugin.PluginContext) error {
    // 使用 ctx.DB、ctx.Cache、ctx.Services
    // 注册自定义路由或块类型
    return nil
}
```

**2. 注册插件**

```go
func init() {
    plugin.RegisterPlugin(func() plugin.Plugin {
        return &MyPlugin{}
    })
}
```

**3. 在主程序中引入**

```go
// cmd/server/main.go
import (
    _ "rua.plus/cadmus/plugins/my-plugin" // blank import 触发注册
)
```

### 注册表 API

| 函数 | 说明 |
|------|------|
| `RegisterPlugin(ctor)` | 注册插件构造函数（不初始化） |
| `RegisterPluginWithContext(ctor, ctx)` | 注册并立即初始化插件 |
| `GetPlugin(id)` | 获取插件新实例 |
| `AllPlugins()` | 返回所有插件元信息 |
| `PluginIDs()` | 返回所有插件 ID 列表 |
| `HasPlugin(id)` | 检查插件是否已注册 |
| `Count()` | 返回已注册插件数量 |

### 并发安全

所有注册表操作使用 `sync.RWMutex` 保护：
- 读操作（GetPlugin、AllPlugins 等）使用读锁
- 写操作（RegisterPlugin）使用写锁

### 注意事项

- 插件通过 `init()` 在编译时自动注册
- 重复注册同一插件 ID 会被忽略并记录警告
- 缺失依赖会记录警告但不会阻止注册
- `RegisterPlugin` 仅注册不初始化，`RegisterPluginWithContext` 会调用 `Init()`
