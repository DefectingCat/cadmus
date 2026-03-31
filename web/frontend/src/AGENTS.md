<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# src

## Purpose
TypeScript 源代码目录，包含前台交互、后台管理和块编辑器代码。

## Key Files
| File | Description |
|------|-------------|
| `main.ts` | 前台入口文件 |
| `admin/main.ts` | 后台管理入口文件 |
| `global.d.ts` | 全局类型声明 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `admin/` | 后台管理模块 (see `admin/AGENTS.md`) |
| `api/` | API 调用封装 (see `api/AGENTS.md`) |
| `components/` | 前台交互组件（待创建） |
| `editor/` | 块编辑器模块（待创建） |
| `styles/` | CSS 样式文件 |
| `utils/` | 工具函数 |

## For AI Agents

### Working In This Directory
- TypeScript 严格模式
- 使用原生 JS，零框架依赖
- esbuild 打包

### Entry Points
| File | Bundle Output |
|------|---------------|
| `main.ts` | `../static/dist/main.js` |
| `admin/main.ts` | `../static/dist/admin/main.js` |
| `editor/index.ts` | `../static/dist/editor.js` |

### Module Organization
```
src/
├── main.ts          # 前台入口
├── admin/           # 后台管理
│   ├── main.ts      # 后台入口
│   ├── posts/       # 文章管理
│   ├── comments/    # 评论管理
│   └── media/       # 媒体管理
├── api/             # API 调用
│   ├── client.ts    # 基础请求
│   ├── posts.ts     # 文章 API
│   └── comments.ts  # 评论 API
├── utils/           # 工具函数
│   └── dom.ts       # DOM 操作
└── styles/          # 样式文件
    ├── main.css     # 前台样式
    └── admin.css    # 后台样式
```

### Import Pattern
```typescript
// 相对导入
import { apiClient } from './api/client';
import { createPost } from './api/posts';

// 类型导入
import type { Post, Comment } from './types';
```

<!-- MANUAL: -->