<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# cmd

## Purpose
应用入口点目录，包含服务器启动代码。

## Key Files
| File | Description |
|------|-------------|
| `server/main.go` | HTTP 服务器入口，初始化数据库、Redis、JWT、路由 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `server/` | 主服务器入口点 (see `server/AGENTS.md`) |

## For AI Agents

### Working In This Directory
- 修改入口点逻辑时，需要重新编译：`make build/backend`
- 环境变量配置在此读取，新增配置需添加 `getEnvOrDefault` 调用
- 新增 API 路由需在 `main.go` 的 `mux.HandleFunc` 中添加

### Architecture Notes
- 使用标准库 `net/http` 的 `ServeMux` 路由
- 中间件链通过函数包装实现
- 优雅关闭通过 `signal.Notify` + `server.Shutdown` 实现

<!-- MANUAL: -->