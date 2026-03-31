<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# Cadmus

## Purpose
Cadmus 是一个基于 Go 的现代博客/CMS 平台，支持文章发布、评论管理、媒体上传、RSS 订阅和全文搜索。采用分层架构设计，前端使用 Templ 模板引擎和 TypeScript/React。

## Key Files
| File | Description |
|------|-------------|
| `go.mod` | Go 模块定义，依赖包括 templ、pgx、jwt、redis |
| `Makefile` | 构建命令：后端编译、前端打包、templ 生成 |
| `cmd/server/main.go` | 服务入口点，初始化数据库、Redis、JWT、路由 |
| `configs/config.example.yaml` | 配置文件模板 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `bin/` | 编译输出目录，存放 server 二进制文件 |
| `cmd/` | 应用入口点 (see `cmd/AGENTS.md`) |
| `configs/` | 配置文件目录 |
| `docs/` | 项目文档 (see `docs/AGENTS.md`) |
| `internal/` | 私有应用代码，核心业务逻辑 (see `internal/AGENTS.md`) |
| `migrations/` | PostgreSQL 数据库迁移脚本 |
| `pkg/` | 公共可导出包（当前为空） |
| `plugins/` | 可插拔扩展模块 (see `plugins/AGENTS.md`) |
| `themes/` | 主题系统 (see `themes/AGENTS.md`) |
| `web/` | 前端资源：模板、静态文件、TypeScript (see `web/AGENTS.md`) |

## For AI Agents

### Working In This Directory
- **构建**: `make build` 编译后端+前端
- **仅后端**: `make build/backend` 或 `go build -o bin/server ./cmd/server`
- **仅前端**: `make build/frontend` (需要 bun)
- **Templ 生成**: `make build/templ` 或 `templ generate`
- **测试**: `make test` 或 `go test ./...`

### Architecture Notes
- **分层架构**: Repository (数据层) → Service (业务层) → Handler (API层)
- **依赖注入**: Service Container 模式管理服务依赖
- **认证**: JWT + Token Blacklist + Permission Cache
- **缓存**: Redis 用于权限缓存和 Token 黑名单
- **插件系统**: Blank import 触发 `init()` 注册插件

### Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | 服务端口 |
| `DB_HOST` | localhost | PostgreSQL 主机 |
| `DB_PORT` | 5432 | PostgreSQL 端口 |
| `DB_NAME` | cadmus | 数据库名 |
| `DB_USER` | cadmus | 数据库用户 |
| `DB_PASSWORD` | "" | 数据库密码 |
| `REDIS_HOST` | localhost | Redis 主机 |
| `REDIS_PORT` | 6379 | Redis 端口 |
| `UPLOAD_DIR` | ./uploads | 上传文件目录 |
| `BASE_URL` | http://localhost:8080 | 基础 URL |

### API Structure
- `/api/v1/auth/*` - 认证 API (注册、登录、注销、刷新)
- `/api/v1/posts/*` - 文章 CRUD + 发布/回滚/点赞
- `/api/v1/comments/*` - 评论 CRUD + 点赞/审核
- `/api/v1/media/*` - 媒体上传管理
- `/api/v1/search/*` - 全文搜索 + 建议
- `/api/v1/rss` - RSS 订阅
- `/api/v1/admin/*` - 管理后台 API (需权限)
- `/admin/*` - 管理后台页面 (Templ 渲染)

### Common Patterns
- Repository 使用 `pgx/v5` 连接池
- Service 封装业务逻辑，返回 domain models
- Handler 使用 Service，处理 HTTP 请求/响应
- Templ 模板生成 Go 代码，无运行时模板解析
- 中间件链：AuthMiddleware → PermissionMiddleware → Handler

## Dependencies

### External
| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/a-h/templ` | v0.3.1001 | 类型安全 HTML 模板 |
| `github.com/jackc/pgx/v5` | v5.9.1 | PostgreSQL 驱动 |
| `github.com/golang-jwt/jwt/v5` | v5.3.1 | JWT 认证 |
| `github.com/redis/go-redis/v9` | v9.18.0 | Redis 客户端 |
| `golang.org/x/crypto` | v0.49.0 | 密码哈希 (bcrypt) |
| `github.com/google/uuid` | v1.6.0 | UUID 生成 |

<!-- MANUAL: 项目特定说明可在此添加 -->