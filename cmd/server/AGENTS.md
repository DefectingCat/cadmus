<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# cmd/server

## 用途

`cmd/server` 目录包含 Cadmus 博客平台的主服务器入口点。该目录是应用程序的核心启动模块，负责：

- 加载并验证所有配置项（数据库、Redis、JWT、服务器、上传）
- 初始化基础设施组件（数据库连接池、Redis 客户端）
- 构建数据访问层（Repositories）和服务层（Services）
- 注册所有 HTTP 路由和中间件链
- 启动 HTTP 服务器并处理优雅关闭

该目录是运行 Cadmus 博客平台的唯一入口点，所有组件的依赖注入和生命周期管理都在此协调完成。

## 关键文件

| 文件 | 功能 |
|------|------|
| `main.go` | 程序入口点，负责组件初始化、依赖注入、HTTP 服务器启动和优雅关闭。包含版本信息打印、基础设施初始化、Repository/Service/Handler 创建、路由配置等完整启动流程 |
| `config.go` | 配置加载模块，定义 `Config`、`ServerConfig`、`UploadConfig` 结构体，从环境变量读取配置参数并提供默认值。包含 `loadConfig()`、`loadJWTConfig()` 等函数 |
| `routes.go` | 路由注册模块，定义 `RouteDeps` 依赖聚合结构体，配置所有 API 路由（认证、文章、分类、标签、评论、媒体、RSS、搜索、管理后台）和中间件链（限流、认证、权限检查） |

## 子目录

无

## 架构说明

### 启动流程

```
main.go (入口)
├── 1. 打印版本信息 (printVersionInfo)
├── 2. 加载配置 (loadConfig → config.go)
│   ├── database.Config (PostgreSQL 连接池)
│   ├── cache.Config (Redis 连接池)
│   ├── auth.JWTConfig (JWT 认证)
│   ├── ServerConfig (HTTP 服务器)
│   └── UploadConfig (文件上传)
├── 3. 基础设施初始化
│   ├── database.NewPool() → pgx/v5 连接池
│   └── cache.NewRedisClient() → go-redis/v9 客户端
├── 4. Repository 层创建 (internal/database)
│   ├── UserRepository, RoleRepository, PermissionRepository
│   ├── PostRepository, CategoryRepository, TagRepository, SeriesRepository
│   ├── CommentRepository, CommentLikeRepository, PostLikeRepository
│   ├── MediaRepository, SearchRepository
│   └── TransactionManager
├── 5. 认证组件初始化 (internal/auth)
│   ├── JWTService (令牌签发/验证)
│   ├── TokenBlacklist (Redis 黑名单)
│   └── PermissionCache (权限缓存)
├── 6. 限流器初始化 (internal/api/middleware)
│   ├── LoginLimiter (登录限流)
│   ├── PublicLimiter (公开 API 限流)
│   └── UserLimiter (用户操作限流)
├── 7. Service 容器创建 (internal/services)
│   ├── AuthService, UserService
│   ├── PostService, CategoryService, TagService
│   ├── CommentService, MediaService, RSSService, SearchService
│   └── AdminService
├── 8. Handler 层创建 (internal/api/handlers)
│   ├── AuthHandler, PostHandler, CategoryHandler, TagHandler
│   ├── CommentHandler, MediaHandler, RSSHandler, SearchHandler
│   └── AdminHandler
├── 9. 路由注册 (routes.go)
│   ├── setupAuthRoutes() - 认证 API
│   ├── setupPostRoutes() - 文章 CRUD
│   ├── setupCategoryRoutes() - 分类管理
│   ├── setupTagRoutes() - 标签管理
│   ├── setupCommentRoutes() - 评论及审核
│   ├── setupMediaRoutes() - 媒体上传
│   ├── setupRSSRoutes() - RSS Feed
│   ├── setupSearchRoutes() - 全文搜索
│   ├── setupAdminRoutes() - 管理后台 API
│   ├── setupAdminPages() - 管理后台页面 (Templ)
│   ├── setupStaticRoutes() - 静态文件服务
│   └── setupHealthRoute() - 健康检查
└── 10. HTTP 服务器启动与优雅关闭
```

### 配置项

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `PORT` | 8080 | HTTP 服务端口 |
| `DB_HOST` | localhost | PostgreSQL 主机 |
| `DB_PORT` | 5432 | PostgreSQL 端口 |
| `DB_NAME` | cadmus | 数据库名称 |
| `DB_USER` | cadmus | 数据库用户名 |
| `DB_PASSWORD` | "" | 数据库密码 |
| `DB_SSLMODE` | disable | SSL 连接模式 |
| `REDIS_HOST` | localhost | Redis 主机 |
| `REDIS_PORT` | 6379 | Redis 端口 |
| `REDIS_PASSWORD` | "" | Redis 密码 |
| `UPLOAD_DIR` | ./uploads | 上传文件存储目录 |
| `BASE_URL` | http://localhost:8080 | 服务基础 URL |

### 编译版本注入

通过 `-ldflags` 在编译时注入版本信息：

```bash
go build -ldflags "-X main.version=v1.0.0 -X main.gitCommit=abc123 -X main.gitBranch=master -X main.buildTime=2026-04-09T12:00:00Z -X main.goVersion=go1.22 -X main.buildPlatform=linux/amd64" -o bin/server ./cmd/server
```

版本变量：
- `version` - 应用版本号（如 "v1.0.0"）
- `gitCommit` - Git 提交哈希
- `gitBranch` - Git 分支名
- `buildTime` - 构建时间（RFC3339 格式）
- `goVersion` - 编译使用的 Go 版本
- `buildPlatform` - 目标平台（如 "linux/amd64"）

## 给 AI Agent 的指南

### 编译与运行

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

### 常见修改场景

**添加新路由**
1. 在 `routes.go` 中创建新的 `setupXxxRoutes()` 函数
2. 在 `setupRoutes()` 中调用该函数
3. 确保 `RouteDeps` 结构体包含所需依赖

**添加新配置项**
1. 在 `config.go` 中扩展相应的配置结构体
2. 在 `loadConfig()` 中使用 `getEnvOrDefault()` 读取环境变量
3. 更新配置表格文档

**修改中间件链**
1. 在对应的 `setupXxxRoutes()` 函数中调整中间件组合
2. 中间件顺序：限流 → 认证 → 权限检查 → Handler

### 关键注意事项

- **环境变量优先级**: 所有配置项均支持环境变量覆盖，需确保 `.env` 文件正确配置
- **依赖初始化顺序**: Repository → Service → Handler，不可颠倒
- **限流器依赖 Redis**: 确保 Redis 服务在服务器启动前可用
- **JWT 配置必须**: JWT 密钥缺失会导致启动时 panic
- **上传目录权限**: 确保 `UPLOAD_DIR` 存在且具有写权限

### 调试技巧

1. **数据库连接失败**: 检查 `DB_*` 环境变量和 PostgreSQL 服务状态
2. **Redis 连接失败**: 检查 `REDIS_HOST` 和 `REDIS_PORT`，确认 Redis 服务运行正常
3. **JWT 配置错误**: 确认 `JWT_SECRET` 或等效环境变量已正确设置
4. **端口占用**: 修改 `PORT` 环境变量或使用不同端口
5. **权限问题**: 检查 `UPLOAD_DIR` 目录是否存在且具有写权限
