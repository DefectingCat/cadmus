<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# server

## Purpose
HTTP 服务器入口点，负责应用初始化和路由配置。

## Key Files
| File | Description |
|------|-------------|
| `main.go` | 服务启动、依赖注入、路由注册、优雅关闭 |

## For AI Agents

### Working In This Directory
- 主入口文件，所有初始化逻辑在此
- 新增 API 路由需在 `mux.HandleFunc` 中添加
- 新增依赖需在 `NewContainer*` 调用中传递

### Initialization Flow
```
1. 环境变量读取 (PORT, DB_*, REDIS_*, UPLOAD_DIR, BASE_URL)
2. 数据库连接池初始化 (database.NewPool)
3. Redis 连接池初始化 (cache.NewRedisClient)
4. 缓存服务初始化 (cache.NewService)
5. Repository 初始化 (所有数据访问层)
6. JWT 服务初始化 (auth.NewJWTService)
7. Token 黑名单初始化 (auth.NewRedisTokenBlacklist)
8. 权限缓存初始化 (auth.NewPermissionCache)
9. Service 容器初始化 (services.NewContainerWithMedia)
10. Handler 初始化 (所有 HTTP 处理器)
11. 路由注册 (mux.HandleFunc)
12. 服务器启动 (server.ListenAndServe)
13. 优雅关闭 (signal.Notify + server.Shutdown)
```

### Route Groups
| Prefix | Middleware | Description |
|--------|------------|-------------|
| `/api/v1/auth/*` | 无/Token | 认证 API |
| `/api/v1/posts/*` | 无/Token | 文章 API |
| `/api/v1/comments/*` | 无/Token | 评论 API |
| `/api/v1/admin/*` | Token + Admin | 管理员 API |
| `/admin/*` | Token + Admin | 管理后台页面 |
| `/static/*` | 无 | 静态文件服务 |

<!-- MANUAL: -->