<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# mermaid-block

## Purpose
Mermaid 图表块插件，支持文章内容中的流程图、序列图等渲染。

## Key Files
| File | Description |
|------|-------------|
| `plugin.go` | 插件实现：Info、Init、init() 注册 |

## For AI Agents

### Working In This Directory
- 示范性插件，展示插件开发模式
- 当前仅注册元信息，待 BlockTypeRegistry 实现后完善

### Plugin Metadata
| Field | Value |
|-------|-------|
| ID | `mermaid-block` |
| Name | Mermaid 图表块 |
| Version | 1.0.0 |
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

### Usage in Articles (计划)
```json
{
    "type": "mermaid",
    "data": {
        "code": "graph TD\n    A[Start] --> B[End]"
    }
}
```

### Frontend Rendering (计划)
```html
<div class="mermaid-block">
    <pre class="mermaid">
        graph TD
            A[Start] --> B[End]
    </pre>
</div>
<script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
<script>mermaid.initialize({startOnLoad: true});</script>
```

<!-- MANUAL: -->