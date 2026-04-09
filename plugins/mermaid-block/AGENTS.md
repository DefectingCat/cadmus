<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# mermaid-block

## Purpose
Mermaid 图表块插件，支持在文章内容中使用 Mermaid 语法编写和渲染图表。Mermaid 是一种基于文本的图表描述语言，支持流程图、序列图、甘特图等多种图表类型。

## Key Files
| File | Description |
|------|-------------|
| `plugin.go` | 插件主实现文件，定义 `MermaidBlockPlugin` 结构，实现 `Info()`、`Init()` 方法和 `init()` 自动注册逻辑 |

## Subdirectories
无子目录。

## For AI Agents

### Plugin Registration
此插件通过 blank import 自动注册到全局插件注册表：
```go
import _ "rua.plus/cadmus/plugins/mermaid-block"
```

### Plugin Info
| Field | Value |
|-------|-------|
| ID | `mermaid-block` |
| Name | Mermaid 图表块 |
| Version | 1.0.0 |
| Author | Cadmus Team |
| Dependencies | 无 |

### Supported Diagram Types
| Type | Description |
|------|-------------|
| Flowchart | 流程图 |
| Sequence | 序列图 |
| Class | 类图 |
| State | 状态图 |
| Gantt | 甘特图 |
| Pie | 饼图 |
| GitGraph | Git 分支图 |

### Implementation Status
- [x] 插件结构定义 (`MermaidBlockPlugin`)
- [x] `Info()` 返回插件元信息
- [x] `Init()` 初始化函数（记录日志）
- [x] `init()` 自动注册到全局插件注册表
- [ ] 自定义块类型注册（待 `BlockTypeRegistry` 实现后添加）

### Frontend Requirements
需要在主题中引入 Mermaid JavaScript 库以支持图表渲染：
```html
<script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
<script>mermaid.initialize({startOnLoad: true});</script>
```

### Usage Example (计划)
```json
{
    "type": "mermaid",
    "data": {
        "code": "graph TD\n    A[Start] --> B[End]"
    }
}
```

### TODO
当 `BlockTypeRegistry` 接口实现后，在 `Init()` 函数中添加：
```go
if ctx.Registry != nil {
    registry := ctx.Registry.(plugin.BlockTypeRegistry)
    registry.Register("mermaid", &MermaidBlockType{})
}
```

### Working In This Directory
- **编译**: `make build` 或 `go build ./...`
- **测试**: `go test ./plugins/mermaid-block/...`
