<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# web

## Purpose
前端资源目录，包含 templ 模板、静态文件、TypeScript 前端代码。

## Key Files
| File | Description |
|------|-------------|
| 无顶层文件 | 按类型分目录组织 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `frontend/` | TypeScript 前端代码 + esbuild 构建 (see `frontend/AGENTS.md`) |
| `static/` | 静态资源输出目录 (see `static/AGENTS.md`) |
| `templates/` | templ 模板文件 (see `templates/AGENTS.md`) |

## For AI Agents

### Working In This Directory
- **构建流程**: `make build/frontend` 执行 esbuild + Tailwind
- **开发模式**: `make live` 启动多进程监听
- **静态服务**: `/static/` 路由指向 `web/static/` 目录

### Build Pipeline
```
web/frontend/src/*.ts → esbuild → web/static/dist/*.js
web/frontend/src/*.css → Tailwind → web/static/dist/*.css
web/templates/*.templ → templ generate → *_templ.go
```

### Static Routes (main.go)
| Route | Directory |
|-------|-----------|
| `/static/` | `web/static/` |
| `/uploads/` | `./uploads/` (可配置) |

<!-- MANUAL: -->