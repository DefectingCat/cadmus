<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# admin/ - 管理后台前端模块

## Purpose

本目录包含 Cadmus 管理后台的全部前端源代码，负责内容管理功能的用户界面实现。管理后台提供文章管理、评论审核、媒体资源上传等核心功能。

技术栈：
- **原生 TypeScript** - 零框架依赖，轻量高效
- **esbuild** - 快速打包构建
- **TailwindCSS v4** - 原子化 CSS 样式
- **模块化架构** - 功能模块独立封装

## Key Files

| File | Description |
|------|-------------|
| `main.ts` | 管理后台主入口，初始化主题切换和侧边栏响应式功能 |

### main.ts 功能说明

```typescript
// 核心功能
- 主题切换 (initThemeToggle) - 明/暗主题切换，支持 localStorage 持久化
- 移动端侧边栏 (initSidebarToggle) - 响应式侧边栏折叠/展开
```

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `comments/` | 评论管理模块 - 评论列表展示、审核、删除 |
| `media/` | 媒体资源管理 - 图片/文件上传、列表浏览 |
| `posts/` | 文章管理模块 - 文章列表、创建、编辑、删除 |

## For AI Agents

### 开发指南

1. **入口点约定** - 每个子目录必须有独立的入口文件（`index.ts` 或 `list.ts`）
2. **API 调用** - 所有后端请求必须通过 `src/api/` 层封装
3. **样式隔离** - 管理后台样式使用 `src/styles/admin.css`
4. **类型安全** - 所有代码必须通过 TypeScript 严格模式检查

### 添加新功能

```bash
# 1. 创建模块目录
mkdir -p src/admin/new-feature

# 2. 创建入口文件
touch src/admin/new-feature/index.ts

# 3. 在 main.ts 中导入新模块
import './new-feature/index';

# 4. 类型检查
bun run typecheck

# 5. 格式化
bun run fmt
```

### 构建输出

| Entry Point | Output Bundle |
|-------------|---------------|
| `admin/main.ts` | `web/frontend/static/dist/admin/main.js` |
| `admin/posts/list.ts` | `web/frontend/static/dist/admin/posts.js` |
| `admin/posts/editor.ts` | `web/frontend/static/dist/admin/post-edit.js` |
| `admin/comments/list.ts` | `web/frontend/static/dist/admin/comments.js` |
| `admin/media/index.ts` | `web/frontend/static/dist/admin-media.js` |

### 模块依赖图

```
admin/main.ts
├── 主题切换 (initThemeToggle)
├── 侧边栏切换 (initSidebarToggle)
└── 功能模块导入
    ├── admin/posts/list.ts
    ├── admin/posts/editor.ts
    ├── admin/comments/list.ts
    └── admin/media/index.ts
```

### 编码规范

```typescript
// 导入顺序
import { apiClient } from '../../api/client';     // API 层优先
import { createElement } from '../../utils/dom';   // 工具函数
import './styles/admin.css';                       // 样式文件

// 导出模式
export { initFeature };  // 命名导出，便于测试
```
