<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# web

## 用途

前端资源目录，包含完整的 React + TypeScript 前端代码、Templ 模板引擎生成的模板文件、以及构建后的静态资源。

## 目录结构

| 目录 | 用途 |
|------|------|
| `frontend/` | TypeScript 源码 + esbuild 构建配置 |
| `static/` | 构建输出的静态资源（JS/CSS） |
| `templates/` | Templ 模板文件（布局、页面、组件） |

## 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| TypeScript | ^5.0.0 | 类型安全的 JavaScript |
| React | ^19.0.0 | UI 组件库 |
| esbuild | ^0.27.4 | 快速打包构建 |
| Tailwind CSS | ^4.2.2 | 原子化 CSS |
| Templ | v0.3.1001 | 类型安全 Go 模板引擎 |
| Bun | latest | 包管理器 + 运行时 |

## 前端开发指南

### 项目初始化

```bash
cd web/frontend
bun install
```

### 构建命令

```bash
# 构建全部资源（JS + CSS）
cd web/frontend && bun run build:all

# 仅构建主应用
bun run build

# 仅构建管理后台
bun run build:admin

# 构建 CSS
bun run build:css

# 开发模式（监听变化）
bun run dev
bun run dev:css
```

### 代码质量

```bash
# 类型检查
bun run typecheck

# 格式化代码
bun run fmt

# Lint 检查
bun run lint
bun run check
```

### 构建输出流程

```
web/frontend/src/*.ts    → esbuild → web/static/dist/*.js
web/frontend/src/*.css   → Tailwind → web/static/dist/*.css
web/templates/*.templ    → templ generate → *_templ.go
```

### 入口文件

| 入口 | 输出 | 用途 |
|------|------|------|
| `src/main.ts` | `dist/main.js` | 主应用入口 |
| `src/admin/main.ts` | `dist/admin/main.js` | 管理后台入口 |
| `src/admin/posts/list.ts` | `dist/admin/posts.js` | 文章列表页 |
| `src/admin/posts/editor.ts` | `dist/admin/post-edit.js` | 文章编辑器 |
| `src/admin/media/index.ts` | `dist/admin-media.js` | 媒体管理 |
| `src/admin/comments/list.ts` | `dist/admin/comments.js` | 评论管理 |

## 静态资源服务

后端通过以下路由提供静态文件：

| 路由 | 目录 |
|------|------|
| `/static/` | `web/static/` |
| `/uploads/` | `./uploads/` |

## 给 AI Agent 的提示

- 修改 TypeScript 代码后，运行 `bun run build` 或 `bun run dev` 进行构建
- 修改 CSS 后，运行 `bun run build:css` 进行构建
- 修改 Templ 模板后，运行 `templ generate` 生成 Go 代码
- 完整构建使用：`make build/frontend`
