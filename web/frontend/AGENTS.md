<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# frontend

## Purpose
TypeScript 前端代码目录，使用 esbuild 打包，Tailwind CSS 样式。

## Key Files
| File | Description |
|------|-------------|
| `package.json` | npm 依赖和构建脚本 |
| `tsconfig.json` | TypeScript 配置 |
| `tailwind.config.js` | Tailwind CSS 配置 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `src/` | TypeScript 源代码 (see `src/AGENTS.md`) |
| `static/` | 构建输出目录 |
| `node_modules/` | npm 依赖（不提交） |

## For AI Agents

### Working In This Directory
- 使用 bun 作为包管理器
- `bun install` 安装依赖
- 构建命令在 Makefile 中定义

### Build Commands
| Command | Description |
|---------|-------------|
| `make build/frontend` | 生产构建：esbuild + Tailwind |
| `make build/editor` | 单独构建编辑器入口 |
| `make build/templ` | 生成 templ Go 代码 |

### Entry Points
| Entry | Output | Purpose |
|-------|--------|---------|
| `src/main.ts` | `../static/dist/main.js` | 前台交互 |
| `src/admin/main.ts` | `../static/dist/admin/main.js` | 后台管理 |
| `src/editor/index.ts` | `../static/dist/editor.js` | 块编辑器 |

### Development Mode
```bash
# 同时启动所有监听进程
make live

# 或单独启动
# templ watch
templ generate --watch

# esbuild watch
bun esbuild src/main.ts --bundle --outdir=../static/dist --watch --sourcemap=inline

# Tailwind watch
bunx @tailwindcss/cli -i src/styles/main.css -o ../static/dist/styles.css --watch
```

### Dependencies
| Package | Version | Purpose |
|---------|---------|---------|
| `esbuild` | ^0.25.x | 打包工具 |
| `tailwindcss` | ^4.x | CSS 框架 |
| `typescript` | ^6.x | 类型检查 |

<!-- MANUAL: -->