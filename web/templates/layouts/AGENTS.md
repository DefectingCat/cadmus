<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# 布局模板 (Layouts Templates)

## 目录用途

`web/templates/layouts/` 目录存放项目的基础布局模板组件。这些模板提供统一的 HTML 页面结构，包含页面组件共用的框架代码（如 `<head>`、导航栏、页脚等），确保整个应用保持视觉一致性和结构可维护性。

所有布局模板使用 `.templ` 扩展名，基于 [templ](https://templ.guide) 类型安全的 Go HTML 模板引擎编写。

## 关键文件及功能

### base.templ - 前台基础布局

网站前台页面的基础模板集合：

| 组件 | 功能说明 |
|------|----------|
| `BaseLayout(title)` | 完整 HTML 页面框架，包含 charset、viewport、CSS/JS 引用 |
| `Head(title, description)` | 页面头部元数据（可嵌入到其他布局） |
| `Header()` | 顶部导航栏：包含 Logo、首页、文章、关于链接 |
| `Footer()` | 页脚区域：版权信息、RSS、Sitemap 链接 |

**使用示例:**
```go
templ HomePage() {
    @BaseLayout("首页") {
        @Header()
        <main>页面内容</main>
        @Footer()
    }
}
```

### admin.templ - 后台管理布局

后台管理系统的完整布局模板：

| 组件 | 功能说明 |
|------|----------|
| `AdminLayout(title, currentPage)` | 后台页面主框架，含侧边栏和主内容区 |
| `AdminSidebar(currentPage)` | 左侧导航栏：内容管理、系统管理菜单 |
| `AdminHeader(title)` | 顶部工具栏：页面标题、主题切换、快捷操作 |

**布局结构:**
```
┌─────────────────────────────────────┐
│ AdminLayout                         │
│ ┌─────────────┬───────────────────┐ │
│ │ AdminSidebar│  AdminHeader      │ │
│ │             ├───────────────────┤ │
│ │ - 仪表盘     │  页面内容          │ │
│ │ - 文章管理   │  { children... }  │ │
│ │ - 媒体库     │                   │ │
│ │ - 评论审核   │                   │ │
│ │ - 用户管理   │                   │ │
│ │ - 分类管理   │                   │ │
│ │ - 插件管理   │                   │ │
│ │ - 主题管理   │                   │ │
│ │ - 系统设置   │                   │ │
│ └─────────────┴───────────────────┘ │
└─────────────────────────────────────┘
```

**特性:**
- 支持明暗主题切换（基于 localStorage）
- 响应式侧边栏设计
- 当前页面高亮标识
- SVG 图标导航

## 无子目录

该目录当前没有子目录，所有布局模板均为扁平结构。

## AI Agent 开发指南

### 添加新布局模板

1. 在 `web/templates/layouts/` 下创建新的 `.templ` 文件
2. 定义 `package layouts`
3. 使用 `templ` 关键字声明模板组件
4. 通过 `@ComponentName()` 调用其他布局组件

### 扩展现有布局

- **修改导航菜单**: 在对应 `AdminSidebar` 或 `Header` 的 `<nav>` 内添加 `<a>` 链接
- **添加新组件**: 在同一文件中定义新 `templ` 函数，通过 `@` 语法嵌入主布局
- **调整样式**: 修改 class 属性（使用 Tailwind CSS 类名）

### 注意事项

- 所有模板必须声明 `package layouts`
- 使用 `{ children... }` 插槽接收子组件内容
- 路径引用使用绝对路径（如 `/static/dist/main.css`）
- 中文文本保持 UTF-8 编码
- 新增模板后运行 `templ generate` 生成 Go 代码