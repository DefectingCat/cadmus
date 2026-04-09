<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# static

## Purpose
静态资源输出目录，存放构建后的前端资源。

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `dist/` | esbuild 和 Tailwind 构建输出 |

## For AI Agents

### Working In This Directory
- **构建输出目录**，不应手动编辑
- 内容由 `make build/frontend` 生成
- 通过 `/static/` 路由对外服务

### Output Files
| File | Source | Description |
|------|--------|-------------|
| `dist/main.js` | `frontend/src/main.ts` | 前台 JS |
| `dist/admin/main.js` | `frontend/src/admin/main.ts` | 后台 JS |
| `dist/editor.js` | `frontend/src/editor/index.ts` | 编辑器 JS |
| `dist/styles.css` | `frontend/src/styles/main.css` | 前台 CSS |
| `dist/admin.css` | `frontend/src/styles/admin.css` | 后台 CSS |

### HTTP Routes (main.go)
```go
mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
```

### Cache Headers (生产环境建议)
```
Cache-Control: public, max-age=31536000, immutable
```

<!-- MANUAL: -->