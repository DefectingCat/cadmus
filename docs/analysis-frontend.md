# Cadmus 前端构建系统分析报告

## 1. 构建流程概述

Cadmus 采用现代前端构建工具链：**Bun + esbuild + Tailwind CSS v4**。

### 1.1 构建工具选型

| 工具 | 版本 | 用途 |
|------|------|------|
| Bun | - | JS 运行时和包管理器 |
| esbuild | 0.27.4 | TypeScript/JavaScript 打包 |
| Tailwind CSS | 4.2.2 | CSS 框架和样式处理 |
| TypeScript | 6.0.2 | 类型检查 |
| Biome | 2.4.10 | 代码格式化和 lint |

### 1.2 Makefile 构建目标

```makefile
# 构建前端资源（Makefile:43-61）
build/frontend:
    # 主入口打包
    cd web/frontend && bun esbuild src/main.ts \
        --bundle --outdir=../static/dist --minify --format=esm

    # Admin 入口打包
    cd web/frontend && bun esbuild src/admin/main.ts \
        --bundle --outdir=../static/dist/admin --minify --format=esm

    # 主样式处理
    cd web/frontend && bunx @tailwindcss/cli \
        -i src/styles/main.css -o ../static/dist/styles.css --minify

    # Admin 样式处理
    cd web/frontend && bunx @tailwindcss/cli \
        -i src/styles/admin.css -o ../static/dist/admin.css --minify
```

### 1.3 入口点架构

前端采用多入口点设计，按功能模块分离：

| 入口文件 | 输出路径 | 功能 |
|----------|----------|------|
| `src/main.ts` | `static/dist/main.js` | 前台站点主入口 |
| `src/admin/main.ts` | `static/dist/admin/main.js` | 后台管理主入口 |
| `src/admin/media/index.ts` | `static/dist/admin-media.js` | 媒体库管理 |
| `src/admin/comments/list.ts` | `static/dist/admin/comments.js` | 评论管理 |
| `src/editor/index.ts` | `static/dist/editor.js` | 编辑器模块 |

### 1.4 开发模式支持

```json
// package.json scripts
{
  "dev": "esbuild src/main.ts --bundle --outdir=../static/dist --watch --sourcemap=inline --format=esm",
  "dev:css": "bunx @tailwindcss/cli -i src/styles/main.css -o ../static/dist/styles.css --watch"
}
```

支持 watch 模式和 sourcemap，便于开发调试。

---

## 2. templ 模板组件设计

Cadmus 使用 [templ](https://templ.guide/) 作为 Go 端模板引擎，实现类型安全的 HTML 渲染。

### 2.1 目录结构

```
web/templates/
├── layouts/           # 布局模板
│   ├── base.templ     # 前台基础布局
│   ├── admin.templ    # 后台管理布局
│   └── *_templ.go     # 生成文件
├── pages/             # 页面模板
│   ├── home.templ     # 首页
│   ├── post_list.templ
│   ├── post_detail.templ
│   └── admin/         # 后台页面
│       ├── dashboard.templ
│       ├── admin_posts.templ
│       ├── admin_media.templ
│       ├── admin_comments.templ
│       ├── admin_themes.templ
│       └── admin_plugins.templ
└── partials/          # 可复用组件（待扩展）
```

### 2.2 基础布局设计

**前台布局** (`layouts/base.templ`):

```go
templ BaseLayout(title string) {
    <!DOCTYPE html>
    <html lang="zh-CN">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <title>{ title }</title>
            <link rel="stylesheet" href="/static/dist/main.css"/>
        </head>
        <body class="min-h-screen bg-gray-50">
            { children... }
            <script src="/static/dist/main.js"></script>
        </body>
    </html>
}
```

**后台布局** (`layouts/admin.templ`):

采用 Modern Elegant Design，包含：
- 侧边栏导航 (`AdminSidebar`)
- 顶部导航栏 (`AdminHeader`)
- 主题切换按钮
- 响应式移动端支持

### 2.3 templ 特性使用

1. **条件渲染**: `if/else` 语句直接嵌入模板
2. **循环渲染**: `for` 循环生成列表
3. **动态 CSS 类**: `templ.KV("active", condition)` 条件类名
4. **子模板调用**: `@layouts.Header()` 组件嵌套
5. **children 传递**: `{ children... }` 插槽机制

### 2.4 构建集成

```makefile
# Generate templ files
build/templ:
    templ generate
```

templ 文件在编译时生成 Go 代码 (`*_templ.go`)，无需运行时解析。

---

## 3. 主题引擎实现

### 3.1 核心模型 (`internal/theme/models.go`)

```go
// Theme 主题元信息
type Theme struct {
    ID          string          // 主题唯一标识
    Name        string          // 主题名称
    Version     string          // 版本号
    Author      string          // 作者
    Description string          // 描述
    Components  ThemeComponents // 主题组件实现
}

// ThemeComponents 主题必须实现的组件接口
type ThemeComponents interface {
    Layout() templ.Component       // 整体布局框架
    Header() templ.Component       // 头部
    Footer() templ.Component       // 底部
    PostList() templ.Component     // 文章列表页
    PostDetail() templ.Component   // 文章详情页
    CategoryPage() templ.Component // 分类页
    TagPage() templ.Component      // 标签页
    Sidebar() templ.Component      // 侧边栏（可选）
    ErrorPage() templ.Component    // 错误页
}
```

### 3.2 主题注册表 (`internal/theme/registry.go`)

采用单例模式管理主题：

```go
var (
    globalRegistry ThemeRegistry
    once           sync.Once
)

func GetRegistry() ThemeRegistry {
    once.Do(func() {
        globalRegistry = NewThemeRegistry()
    })
    return globalRegistry
}
```

**ThemeRegistry 接口**:

```go
type ThemeRegistry interface {
    Register(theme Theme) error          // 注册主题
    GetActive() (Theme, error)           // 获取激活主题
    SetActive(themeID string) error      // 切换主题
    All() []Theme                        // 列出所有主题
}
```

### 3.3 错误定义

```go
var (
    ErrThemeNotFound      = &ThemeError{Code: "theme_not_found", Message: "主题不存在"}
    ErrThemeAlreadyExists = &ThemeError{Code: "theme_already_exists", Message: "主题已存在"}
    ErrNoActiveTheme      = &ThemeError{Code: "no_active_theme", Message: "未设置激活主题"}
    ErrInvalidThemeID     = &ThemeError{Code: "invalid_theme_id", Message: "无效的主题ID"}
)
```

### 3.4 主题切换机制

后台 Admin CSS 支持 Light/Dark 双主题：

```css
/* Light Theme (Default) */
:root {
  --bg-deep: #fafbfc;
  --accent-primary: #6366f1;
  --text-primary: #1f2937;
  ...
}

/* Dark Theme */
[data-theme="dark"] {
  --bg-deep: #0f1117;
  --accent-primary: #818cf8;
  --text-primary: #f3f4f6;
  ...
}
```

前端 JS 处理切换：

```typescript
// admin/main.ts
function initThemeToggle() {
  const toggleBtn = document.getElementById("theme-toggle");
  toggleBtn.addEventListener("click", () => {
    const currentTheme = document.documentElement.getAttribute("data-theme");
    const newTheme = currentTheme === "dark" ? "light" : "dark";
    document.documentElement.setAttribute("data-theme", newTheme);
    localStorage.setItem("theme", newTheme);
  });
}
```

---

## 4. 插件系统设计

### 4.1 插件引擎架构 (`internal/plugin/`)

采用**编译时注册模式**，不使用 Go 原生 plugin 包，避免跨平台兼容性问题。

**Plugin 接口** (`plugin/plugin.go`):

```go
type Plugin interface {
    Info() PluginInfo          // 返回插件元信息
    Init(ctx *PluginContext) error  // 初始化插件
}

type PluginInfo struct {
    ID          string   // 唯一标识符
    Name        string   // 显示名称
    Version     string   // 版本号
    Author      string   // 作者
    Description string   // 功能描述
    Dependencies []string // 依赖的其他插件 ID
}

type PluginContext struct {
    DB       *database.Pool      // 数据库连接池
    Cache    cache.CacheService  // 缓存服务
    Registry any                 // BlockTypeRegistry
    Services *services.Container // 业务服务容器
    Config   map[string]any      // 插件配置
}
```

### 4.2 插件注册表 (`plugin/registry.go`)

```go
var (
    pluginMap = make(map[string]PluginConstructor)
    mu        sync.RWMutex
)

// RegisterPlugin 注册插件（由插件的 init() 调用）
func RegisterPlugin(ctor PluginConstructor) {
    p := ctor()
    info := p.Info()

    // 检查重复注册
    // 验证依赖关系
    pluginMap[info.ID] = ctor
}
```

### 4.3 示例插件 (`plugins/mermaid-block/`)

```go
type MermaidBlockPlugin struct{}

func (p *MermaidBlockPlugin) Info() plugin.PluginInfo {
    return plugin.PluginInfo{
        ID:          "mermaid-block",
        Name:        "Mermaid 图表块",
        Version:     "1.0.0",
        Author:      "Cadmus Team",
        Description: "支持文章内容中的 Mermaid 图表渲染",
    }
}

func (p *MermaidBlockPlugin) Init(ctx *plugin.PluginContext) error {
    // TODO: 注册自定义块类型
    return nil
}

// init() 中自动注册
func init() {
    plugin.RegisterPlugin(func() plugin.Plugin {
        return &MermaidBlockPlugin{}
    })
}
```

### 4.4 插件启用方式

在 `cmd/server/main.go` 中添加 blank import：

```go
import (
    _ "rua.plus/cadmus/plugins/mermaid-block"
)
```

---

## 5. 静态资源组织

### 5.1 目录结构

```
web/static/
├── dist/              # 构建产物
│   ├── main.js        # 前台 JS
│   ├── main.css       # 前台样式
│   ├── styles.css     # Tailwind 处理后的样式
│   ├── admin.css      # 后台样式（含主题系统）
│   ├── admin/
│   │   └── main.js    # 后台 JS
│   ├── admin-media.js/
│   ├── admin/comments.js/
│   └── assets/        # 其他资源
└── .gitkeep
```

### 5.2 资源引用方式

**模板中引用**:

```go
// layouts/base.templ
<link rel="stylesheet" href="/static/dist/main.css"/>
<script src="/static/dist/main.js"></script>

// layouts/admin.templ
<link rel="stylesheet" href="/static/dist/admin.css"/>
<script src="/static/dist/admin/main.js"></script>
```

### 5.3 字体资源

Admin 主题使用 Google Fonts：

```html
<link rel="preconnect" href="https://fonts.googleapis.com"/>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin/>
```

字体定义：
- **Display**: Fraunces (serif)
- **Body**: Plus Jakarta Sans (sans-serif)

---

## 6. 前端 JavaScript 模块分析

### 6.1 API 客户端 (`src/api/client.ts`)

```typescript
class APIClient {
  private baseUrl = "/api/v1";
  private token: string | null = null;

  // Token 管理（Cookie + localStorage 双重存储）
  private loadToken(): void
  setToken(token: string): void
  clearToken(): void

  // HTTP 方法封装
  async get<T>(path: string, params?: Record<string, string>): Promise<APIResponse<T>>
  async post<T>(path: string, body?: unknown): Promise<APIResponse<T>>
  async put<T>(path: string, body: unknown): Promise<APIResponse<T>>
  async delete<T>(path: string): Promise<APIResponse<T>>
}
```

特性：
- 自动携带 Authorization header
- 统一错误处理 (`APIError` 类型)
- 网络错误封装

### 6.2 DOM 工具函数 (`src/utils/dom.ts`)

提供 jQuery 风格的便捷函数：

```typescript
// 选择器
export const $ = (selector: string, parent?: Element | Document): Element | null
export const $$ = (selector: string, parent?: Element | Document): Element[]

// 元素操作
export const show/hide/toggle(el: Element): void
export const addClass/removeClass/hasClass(el: Element, className: string): void

// 事件绑定
export const on/off(el, event, handler): void
export const delegate(parent, selector, event, handler): void

// 表单处理
export const getFormData(form: HTMLFormElement): Record<string, string>
export const setFormData(form, data): void

// UI 反馈
export const showMessage(message, type): void  // Toast 提示
export const confirm(message): Promise<boolean>  // 确认对话框
```

### 6.3 编辑器增强 (`src/admin/posts/editor.ts`)

`PostEditorManager` 类实现：

| 功能 | 实现方式 |
|------|----------|
| 自动保存 | 30 秒定时器，检测 dirty 状态 |
| 快捷键保存 | Ctrl+S / Cmd+S 监听 |
| 标题转 slug | 自动生成 URL  friendly slug |
| 字数统计 | 实时更新，计算阅读时间 |
| 页面离开提醒 | beforeunload 事件 |
| 状态指示器 | saving/saved/error/dirty 四状态 |
| 发布流程 | 先保存后发布，带确认对话框 |

---

## 7. CSS 主题系统详解

### 7.1 设计变量体系

Admin CSS (`src/styles/admin.css`) 定义了完整的设计系统：

**色彩层级**:
- `--bg-deep/base/card/elevated/hover`: 五级背景层次
- `--accent-primary/secondary/muted`: 三级强调色
- `--text-display/primary/secondary/muted/subtle`: 五级文字层次
- `--border-subtle/default/strong`: 三级边框层次

**阴影系统**:
- `--shadow-xs/sm/md/lg/xl`: 五级阴影层次

**圆角系统**:
- `--radius-sm/md/lg/xl`: 四级圆角尺寸

**间距系统**:
- `--space-xs/sm/md/lg/xl/2xl`: 六级间距尺寸

**过渡动画**:
- `--transition-fast/base/slow`: 三级过渡速度

### 7.2 组件样式

预定义的 UI 组件样式：
- `.card` 卡片组件
- `.stat-card` 统计卡片
- `.data-table` 数据表格
- `.status-badge` 状态徽章
- `.btn-*` 按钮系列
- `.form-*` 表单系列
- `.nav-*` 导航系列
- `.admin-*` 管理后台布局

### 7.3 动画系统

```css
@keyframes fadeIn { ... }
@keyframes fadeInUp { ... }
@keyframes slideInLeft { ... }

.animate-fade-in { animation: fadeIn 0.4s ease forwards; }
.animate-fade-in-up { animation: fadeInUp 0.5s ease forwards; }

.stagger-1 { animation-delay: 0.05s; opacity: 0; }
.stagger-2 { animation-delay: 0.1s; opacity: 0; }
...
```

实现交错动画效果。

### 7.4 响应式断点

```css
@media (max-width: 1280px) { /* 大屏 */ }
@media (max-width: 1024px) { /* 平板 */ }
@media (max-width: 640px)  { /* 手机 */ }
```

---

## 8. 构建流程总结

```
┌─────────────────────────────────────────────────────────────┐
│                    Cadmus 前端构建流程                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  TypeScript 源码                                             │
│  ┌──────────────────┐                                       │
│  │ src/main.ts      │                                       │
│  │ src/admin/*.ts   │                                       │
│  │ src/api/*.ts     │                                       │
│  │ src/utils/*.ts   │                                       │
│  └──────────────────┘                                       │
│           │                                                 │
│           │ esbuild --bundle --minify --format=esm          │
│           ▼                                                 │
│  ┌──────────────────┐                                       │
│  │ static/dist/     │                                       │
│  │ *.js (ESM)       │                                       │
│  └──────────────────┘                                       │
│                                                             │
│  CSS 源码                                                    │
│  ┌──────────────────┐                                       │
│  │ src/styles/      │                                       │
│  │ main.css         │                                       │
│  │ admin.css        │                                       │
│  └──────────────────┘                                       │
│           │                                                 │
│           │ @tailwindcss/cli --minify                       │
│           ▼                                                 │
│  ┌──────────────────┐                                       │
│  │ static/dist/     │                                       │
│  │ styles.css       │                                       │
│  │ admin.css        │                                       │
│  └──────────────────┘                                       │
│                                                             │
│  templ 模板                                                  │
│  ┌──────────────────┐                                       │
│  │ web/templates/   │                                       │
│  │ *.templ          │                                       │
│  └──────────────────┘                                       │
│           │                                                 │
│           │ templ generate                                  │
│           ▼                                                 │
│  ┌──────────────────┐                                       │
│  │ *_templ.go       │                                       │
│  │ (Go 编译时生成)   │                                       │
│  └──────────────────┘                                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 9. 技术亮点与建议

### 亮点

1. **类型安全模板**: templ 提供编译时类型检查，避免运行时模板错误
2. **现代构建工具**: Bun + esbuild 组合提供极速构建体验
3. **CSS 设计系统**: 完整的变量体系支持 Light/Dark 主题切换
4. **模块化 JS**: 多入口点设计，按需加载，避免单一过大 bundle
5. **插件架构**: 编译时注册模式，跨平台兼容，易于扩展

### 建议优化

1. **Tailwind 配置扩展**: 当前配置较为简单，可添加自定义配色和组件类
2. **前端路由**: 可考虑引入轻量路由库增强 SPA 体验
3. **编辑器增强**: 集成 Markdown 或富文本编辑器（如 TipTap）
4. **图片优化**: 添加构建时图片压缩和 WebP 转换
5. **CDN 支持**: 生产环境可配置静态资源 CDN 加速