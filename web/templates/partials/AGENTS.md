<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# partials

## 用途

存放可复用的 Templ 模板片段组件。partials 目录中的模板不包含完整的 HTML 文档结构，仅渲染页面中的特定部分或 UI 组件。

## 当前状态

- **目录状态**: 空目录（仅包含 `.gitkeep` 文件）
- **子目录**: 无

## 开发指南

### 何时使用 partials

将以下类型的组件放入 `partials/` 目录：

| 组件类型 | 示例 | 说明 |
|----------|------|------|
| 导航组件 | `header.templ`, `footer.templ`, `sidebar.templ` | 页面共用的导航结构 |
| 卡片组件 | `post_card.templ`, `user_card.templ` | 数据展示卡片 |
| 表格组件 | `table.templ`, `pagination.templ` | 数据表格和分页 |
| 表单组件 | `input.templ`, `select.templ`, `form_group.templ` | 表单元素 |
| 状态组件 | `badge.templ`, `alert.templ`, `modal.templ` | 状态提示和弹窗 |

### 何时不使用 partials

- **完整页面** → 放入 `pages/` 目录
- **布局框架** → 放入 `layouts/` 目录

### 组件开发规范

#### 1. 文件命名

```
{组件名}.templ         → 标准组件
{组件名}_item.templ   → 列表项组件
{组件名}_group.templ  → 组件组
```

#### 2. 基本结构

```templ
package partials

// 带参数的组件
templ ComponentName(title string, isActive bool) {
    <div class="component">
        <h3>{ title }</h3>
        if isActive {
            <span class="active">活跃</span>
        }
    </div>
}
```

#### 3. 类型安全参数

```templ
// 定义数据结构
type PostSummary struct {
    ID        uint
    Title     string
    CreatedAt string
    Tags      []string
}

// 使用结构化参数
templ PostCard(post PostSummary) {
    <article class="card">
        <h2>{ post.Title }</h2>
        <time>{ post.CreatedAt }</time>
        <div class="tags">
            for _, tag := range post.Tags {
                <span class="tag">{ tag }</span>
            }
        </div>
    </article>
}
```

#### 4. 组件组合

```templ
// 调用其他 partials
@Header(user)
@PostCard(post)
@PaginationNav(current, total)
@Footer()
```

### 开发流程

```bash
# 1. 在 partials 目录创建 .templ 文件
touch web/templates/partials/my_component.templ

# 2. 编写组件代码（使用 templ 语法）

# 3. 生成 Go 代码
templ generate -path ./web/templates/partials

# 4. Watch 模式开发
templ generate --watch -path ./web/templates/partials

# 5. 在页面中引用
# import "rua.plus/cadmus/web/templates/partials"
```

### 测试指南

由于 partials 是纯模板组件，测试应关注：

1. **渲染输出**: 组件是否正确渲染 HTML
2. **边界条件**: 空数据、nil 值的处理
3. **交互性**: 动态类名、条件渲染是否正确

## 参考资源

- [父目录文档](../AGENTS.md)
- [Templ 语法参考](https://templ.guide/syntax-and-usage/statements)
