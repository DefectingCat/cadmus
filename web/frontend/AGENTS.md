<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# frontend

## Purpose

TypeScript 前端代码目录，使用 **esbuild** 打包，**TailwindCSS v4** 样式，**Biome** 代码格式化与检查。

## Key Files

| File | Description |
|------|-------------|
| `package.json` | npm 依赖和构建脚本 |
| `tsconfig.json` | TypeScript 配置（ES2022, strict 模式） |
| `tailwind.config.js` | TailwindCSS 配置 |
| `biome.json` | Biome 格式化与 Lint 配置 |

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `src/` | TypeScript 源代码（见 `src/AGENTS.md`） |
| `src/main.ts` | 主入口文件 |
| `src/admin/` | 管理后台模块 |
| `src/api/` | API 客户端层 |
| `src/components/` | UI 组件 |
| `src/editor/` | 块编辑器模块 |
| `src/styles/` | CSS 样式文件 |
| `src/utils/` | 工具函数 |
| `static/` | 构建输出目录 |
| `node_modules/` | npm 依赖（不提交） |

## For AI Agents

### Working In This Directory

- 使用 **bun** 作为包管理器
- `bun install` 安装依赖
- 所有构建脚本通过 `bun run` 执行

### Build Commands

| Command | Description |
|---------|-------------|
| `bun run build:all` | 构建所有模块（主站 + 管理后台 + CSS） |
| `bun run build` | 构建主入口 `src/main.ts` → `../static/dist/` |
| `bun run build:admin` | 构建管理后台入口 → `../static/dist/admin/main.js` |
| `bun run build:admin-posts` | 构建文章列表页 → `posts.js` |
| `bun run build:admin-post-edit` | 构建文章编辑器 → `post-edit.js` |
| `bun run build:admin-media` | 构建媒体管理模块 |
| `bun run build:admin-comments` | 构建评论管理模块 |
| `bun run build:css` | 构建主站样式 → `styles.css` |
| `bun run build:admin-css` | 构建管理后台样式 → `admin.css` |
| `bun run dev` | 开发模式（watch + sourcemap） |
| `bun run dev:css` | CSS 开发模式 |

### Development Mode

```bash
# JS 热重载
bun run dev

# CSS 热重载  
bun run dev:css
```

### Code Quality

| Command | Description |
|---------|-------------|
| `bun run typecheck` | TypeScript 类型检查 |
| `bun run fmt` | Biome 格式化代码 |
| `bun run fmt:check` | Biome 格式化检查 |
| `bun run lint` | Biome Lint 检查 |
| `bun run lint:fix` | Biome Lint 自动修复 |
| `bun run check` | Biome 完整检查（格式 + Lint） |
| `bun run check:fix` | Biome 完整检查并自动修复 |

### Entry Points

| Entry | Output | Purpose |
|-------|--------|---------|
| `src/main.ts` | `../static/dist/main.js` | 前台交互 |
| `src/admin/main.ts` | `../static/dist/admin/main.js` | 后台管理主入口 |
| `src/admin/posts/list.ts` | `../static/dist/admin/posts.js` | 文章列表页 |
| `src/admin/posts/editor.ts` | `../static/dist/admin/post-edit.js` | 文章编辑器 |
| `src/admin/media/index.ts` | `../static/dist/admin-media.js` | 媒体管理 |
| `src/admin/comments/list.ts` | `../static/dist/admin/comments.js` | 评论管理 |
| `src/editor/index.ts` | `../static/dist/editor.js` | 块编辑器 |

### Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `esbuild` | ^0.27.4 | 快速打包构建 |
| `tailwindcss` | ^4.2.2 | 原子化 CSS 框架 |
| `@tailwindcss/cli` | ^4.2.2 | TailwindCSS CLI 工具 |
| `typescript` | ^6.0.2 | 类型系统 |
| `@biomejs/biome` | ^2.4.10 | 格式化与 Lint |

### Code Style (biome.json)

- **indentWidth**: 2 空格
- **lineWidth**: 100 字符
- **quoteStyle**: 双引号 `"`
- **semicolons**: 始终使用 `;`
- **trailingCommas**: ES5（对象/数组末尾允许逗号）

### Adding New Modules

1. 在 `src/` 下创建新目录
2. 创建入口 TypeScript 文件
3. 在 `package.json` 的 `scripts` 中添加构建命令
4. 运行 `bun run typecheck` 验证类型
5. 运行 `bun run check:fix` 格式化代码

示例构建命令：
```bash
bunx esbuild src/new-module/main.ts --bundle --outfile=../static/dist/new-module.js --minify
```

<!-- MANUAL: -->
