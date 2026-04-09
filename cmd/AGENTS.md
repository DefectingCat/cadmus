<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# cmd

## Purpose
`cmd` 目录包含 Cadmus 应用的入口点。该目录下的每个子目录对应一个可独立编译的可执行程序。当前仅包含 `server` 子目录，即博客平台的主服务器。

## Key Files
| File | Description |
|------|-------------|
| `server/main.go` | 服务器入口点，负责组件初始化、路由注册、HTTP 服务启动和优雅关闭 |
| `server/config.go` | 配置加载模块，定义配置结构体并从环境变量读取参数 |
| `server/routes.go` | 路由注册模块，配置所有 API 和页面路由及中间件链 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `server/` | 博客平台主服务器（见 `server/AGENTS.md`） |

## Architecture

### 启动流程
1. **版本信息**: 打印编译时注入的版本、Git 提交、构建时间等信息
2. **配置加载**: 从环境变量加载数据库、Redis、JWT、服务器和上传配置
3. **基础设施**: 初始化数据库连接池和 Redis 客户端
4. **数据层**: 创建所有 Repository（用户、角色、文章、评论、媒体等）
5. **认证组件**: 初始化 JWT 服务、Token 黑名单、权限缓存、限流器
6. **服务层**: 构建 Service 容器，聚合所有业务服务
7. **处理器**: 创建所有 HTTP Handler
8. **路由注册**: 配置路由映射和中间件链
9. **服务启动**: 启动 HTTP 服务器并等待中断信号
10. **优雅关闭**: 接收信号后等待 10 秒完成在途请求

### 配置项
| 环境变量 | 默认值 | Description |
|----------|--------|-------------|
| `PORT` | 8080 | HTTP 服务端口 |
| `DB_HOST` | localhost | PostgreSQL 主机 |
| `DB_PORT` | 5432 | PostgreSQL 端口 |
| `DB_NAME` | cadmus | 数据库名 |
| `DB_USER` | cadmus | 数据库用户 |
| `DB_PASSWORD` | "" | 数据库密码 |
| `DB_SSLMODE` | disable | SSL 模式 |
| `REDIS_HOST` | localhost | Redis 主机 |
| `REDIS_PORT` | 6379 | Redis 端口 |
| `REDIS_PASSWORD` | "" | Redis 密码 |
| `UPLOAD_DIR` | ./uploads | 上传文件存储目录 |
| `BASE_URL` | http://localhost:8080 | 基础 URL |

### 依赖关系
```
main.go (入口)
├── loadConfig() → config.go
│   ├── database.Config (PostgreSQL 连接池配置)
│   ├── cache.Config (Redis 连接池配置)
│   ├── auth.JWTConfig (JWT 认证配置)
│   └── ServerConfig / UploadConfig
├── 基础设施初始化
│   ├── database.NewPool() → pgx/v5 连接池
│   └── cache.NewRedisClient() → go-redis/v9 客户端
├── Repository 层 (internal/database)
│   ├── UserRepository, RoleRepository, PermissionRepository
│   ├── PostRepository, CategoryRepository, TagRepository, SeriesRepository
│   ├── CommentRepository, CommentLikeRepository, PostLikeRepository
│   ├── MediaRepository, SearchRepository
│   └── TransactionManager
├── 认证组件 (internal/auth)
│   ├── JWTService (令牌签发/验证)
│   ├── TokenBlacklist (Redis 黑名单)
│   └── PermissionCache (权限缓存)
├── 限流器 (internal/api/middleware)
│   ├── LoginLimiter (登录限流)
│   ├── PublicLimiter (公开 API 限流)
│   └── UserLimiter (用户操作限流)
├── Service 容器 (internal/services)
│   ├── AuthService, UserService
│   ├── PostService, CategoryService, TagService
│   ├── CommentService, MediaService, RSSService, SearchService
│   └── AdminService
├── Handler 层 (internal/api/handlers)
│   ├── AuthHandler, PostHandler, CategoryHandler, TagHandler
│   ├── CommentHandler, MediaHandler, RSSHandler, SearchHandler
│   └── AdminHandler
└── routes.go (路由注册)
    ├── setupAuthRoutes() - 认证 API
    ├── setupPostRoutes() - 文章 CRUD
    ├── setupCategoryRoutes() - 分类管理
    ├── setupTagRoutes() - 标签管理
    ├── setupCommentRoutes() - 评论及审核
    ├── setupMediaRoutes() - 媒体上传
    ├── setupRSSRoutes() - RSS Feed
    ├── setupSearchRoutes() - 全文搜索
    ├── setupAdminRoutes() - 管理后台 API
    ├── setupAdminPages() - 管理后台页面 (Templ)
    ├── setupStaticRoutes() - 静态文件服务
    └── setupHealthRoute() - 健康检查
```

### 编译版本注入
```bash
go build -ldflags "-X main.version=v1.0.0 -X main.gitCommit=abc123 -X main.gitBranch=master -X main.buildTime=2026-04-09T12:00:00Z -X main.goVersion=go1.22 -X main.buildPlatform=linux/amd64" -o bin/server ./cmd/server
```

## For AI Agents

### Working In This Directory

**编译服务器**
```bash
make build
# 或
go build -o bin/server ./cmd/server
```

**开发模式运行**
```bash
go run ./cmd/server
```

**测试配置加载**
```bash
PORT=3000 DB_HOST=localhost go run ./cmd/server
```

### Key Considerations
- **环境变量优先级**: 所有配置项均支持环境变量覆盖。`.env` 文件已集成 godotenv/autoload，加载优先级：进程环境变量 > .env.local > .env > 代码默认值
- **依赖初始化顺序**: Repository → Service → Handler，不可颠倒
- **限流器依赖 Redis**: 确保 Redis 服务在服务器启动前可用
- **JWT 配置必须**: JWT 密钥缺失会导致启动时 panic
- **上传目录权限**: 确保 `UPLOAD_DIR` 存在且可写

### Debugging Tips
1. **数据库连接失败**: 检查 `DB_*` 环境变量和 PostgreSQL 服务状态
2. **Redis 连接失败**: 检查 `REDIS_HOST` 和 `REDIS_PORT`
3. **JWT 配置错误**: 确认 `JWT_SECRET` 或等效环境变量已设置
4. **端口占用**: 修改 `PORT` 环境变量或使用不同端口

### Common Modifications
- **添加新路由**: 在 `routes.go` 中添加 `setupXxxRoutes()` 函数并在 `setupRoutes()` 中调用
- **添加新配置**: 在 `config.go` 中扩展配置结构体并在 `loadConfig()` 中读取环境变量
- **修改中间件链**: 在对应 `setupXxxRoutes()` 函数中调整中间件组合顺序
