<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# src - 前端 TypeScript 源码目录

## Purpose

本目录包含 Cadmus 前端应用的全部 TypeScript 源代码，采用模块化架构设计，支持主站和管理后台两个独立应用。使用原生 JavaScript/TypeScript 开发，零框架依赖，通过 esbuild 打包构建。

## Key Files

| File | Description |
|------|-------------|
| `main.ts` | 前端主入口，加载全局样式并初始化 admin 模块 |
| `global.d.ts` | 全局类型声明，支持 CSS/SCSS 模块导入 |
| `admin/main.ts` | 管理后台主入口 |

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `admin/` | 管理后台模块（文章、评论、媒体管理） |
| `api/` | API 客户端层，封装后端接口调用 |
| `components/` | 可复用 UI 组件库 |
| `editor/` | 块编辑器核心模块 |
| `styles/` | CSS 样式文件（TailwindCSS v4） |
| `utils/` | 通用工具函数 |

## Directory Details

### admin/

管理后台功能模块：

| Subdirectory | Purpose |
|--------------|---------|
| `posts/` | 文章管理（`list.ts` 列表页、`editor.ts` 编辑器） |
| `comments/` | 评论管理（`list.ts` 列表页） |
| `media/` | 媒体资源管理（`index.ts`、`list.ts`、`upload.ts`） |

### api/

API 客户端封装：

| File | Description |
|------|-------------|
| `client.ts` | HTTP 客户端基础封装 |
| `posts.ts` | 文章相关 API |
| `comments.ts` | 评论相关 API |

### components/

可复用 UI 组件库（待扩展）

### editor/

块编辑器核心模块（待扩展）

### styles/

样式文件：

| File | Description |
|------|-------------|
| `main.css` | 主站样式 |
| `admin.css` | 管理后台样式 |

### utils/

工具函数：

| File | Description |
|------|-------------|
| `dom.ts` | DOM 操作工具函数 |

## For AI Agents

### Development Guidelines

1. **TypeScript 严格模式** - 所有代码必须通过 `bun run typecheck` 类型检查
2. **零框架依赖** - 使用原生 JavaScript/TypeScript，不引入前端框架
3. **模块化** - 新功能应在独立子目录中开发
4. **API 封装** - 后端调用必须通过 `api/` 层，禁止直接 fetch
5. **样式规范** - 使用 TailwindCSS v4，样式文件置于 `styles/`

### Adding New Features

```bash
# 1. 创建新模块目录
mkdir -p src/new-feature

# 2. 编写 TypeScript 代码

# 3. 在入口文件中导入新模块（main.ts 或 admin/main.ts）

# 4. 运行类型检查
bun run typecheck

# 5. 格式化代码
bun run fmt
```

### Entry Points

| File | Bundle Output | Loaded By |
|------|---------------|-----------|
| `main.ts` | `../static/dist/main.js` | 所有前台页面 |
| `admin/main.ts` | `../static/dist/admin/main.js` | 管理后台首页 |
| `admin/posts/list.ts` | `../static/dist/admin/posts.js` | 文章列表页 |
| `admin/posts/editor.ts` | `../static/dist/admin/post-edit.js` | 文章编辑页 |
| `admin/comments/list.ts` | `../static/dist/admin/comments.js` | 评论管理页 |
| `admin/media/index.ts` | `../static/dist/admin-media.js` | 媒体管理页 |
| `editor/index.ts` | `../static/dist/editor.js` | 块编辑器组件 |

### Import Pattern

```typescript
// 相对导入
import { apiClient } from './api/client';
import { createPost } from './api/posts';

// 工具函数导入
import { createElement, onEvent } from './utils/dom';

// CSS 导入（esbuild 原生支持）
import './styles/main.css';
```

### Module Structure

```
src/
├── main.ts              # 前台入口
├── global.d.ts          # 全局类型声明
├── admin/               # 后台管理
│   ├── main.ts          # 后台入口
│   ├── posts/           # 文章管理
│   │   ├── list.ts      # 文章列表
│   │   └── editor.ts    # 文章编辑器
│   ├── comments/        # 评论管理
│   │   └── list.ts      # 评论列表
│   └── media/           # 媒体管理
│       ├── index.ts     # 媒体主页
│       ├── list.ts      # 媒体列表
│       └── upload.ts    # 媒体上传
├── api/                 # API 调用
│   ├── client.ts        # 基础请求封装
│   ├── posts.ts         # 文章 API
│   └── comments.ts      # 评论 API
├── utils/               # 工具函数
│   └── dom.ts           # DOM 操作
├── components/          # UI 组件（待扩展）
├── editor/              # 块编辑器（待扩展）
└── styles/              # 样式文件
    ├── main.css         # 前台样式
    └── admin.css        # 后台样式
```

<!-- MANUAL: -->
