# Cadmus 架构分析报告

> 生成时间: 2026-04-01
> 分析范围: 整体架构设计、模块划分、依赖关系

## 1. 架构概览

Cadmus 是一个基于 Go 的现代博客/CMS 平台，采用经典的**分层架构**设计，遵循 Go 项目标准布局规范。

### 架构层次图

```
┌─────────────────────────────────────────────────────────────────┐
│                        Presentation Layer                        │
│  ┌──────────────────┐  ┌──────────────────┐  ┌─────────────────┐ │
│  │  HTTP Handlers   │  │  Templ Templates │  │  Middleware     │ │
│  │  (internal/api)  │  │  (web/templates) │  │  (rate limit)   │ │
│  └──────────────────┘  └──────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                        Business Layer                            │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │              Service Container (internal/services)         │  │
│  │  UserService | AuthService | PostService | CommentService │  │
│  └───────────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                        Domain Layer                              │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │         Core Models & Repository Interfaces (internal/core)│  │
│  │        User | Post | Comment | Media | Search | RSS       │  │
│  └───────────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                        Data Access Layer                         │
│  ┌──────────────────┐  ┌──────────────────┐  ┌─────────────────┐ │
│  │  PostgreSQL      │  │     Redis        │  │   Repositories  │ │
│  │  (pgx/v5 pool)   │  │  (go-redis/v9)   │  │ (internal/db)   │ │
│  └──────────────────┘  └──────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. 模块划分与边界

### 2.1 目录结构

| 目录 | 职责 | 可见性 |
|------|------|--------|
| `cmd/server/` | 应用入口点、配置加载、路由注册 | 入口层 |
| `internal/` | 私有业务代码，不可被外部导入 | 核心层 |
| `plugins/` | 可插拔扩展模块 | 扩展层 |
| `themes/` | 主题系统 | 扩展层 |
| `web/` | 前端模板、静态资源 | 展示层 |
| `migrations/` | 数据库迁移脚本 | 数据层 |

> **注意**: 项目未使用 `pkg/` 目录（公共可导出包），符合"内部优先"原则。

### 2.2 internal 子模块详解

#### `internal/api/` - API 层

```
api/
├── handlers/        # HTTP 请求处理器
│   ├── auth.go      # 认证：注册、登录、注销
│   ├── post.go      # 文章 CRUD
│   ├── comment.go   # 评论 CRUD + 审核
│   ├── media.go     # 媒体上传
│   ├── search.go    # 搜索
│   ├── rss.go       # RSS 订阅
│   ├── admin.go     # 管理后台
│   └── middleware.go # 认证/权限中间件
└── middleware/
    └── ratelimit.go # Redis 滑动窗口限流
```

**职责边界**:
- Handler 仅处理 HTTP 请求解析/响应序列化
- 业务逻辑委托给 Service 层
- 不直接访问 Repository

#### `internal/services/` - 业务服务层

```
services/
├── services.go        # 服务容器定义
├── auth_service.go    # 认证业务逻辑
├── user_service.go    # 用户管理
├── post_service.go    # 文章业务
├── comment_service.go # 评论业务
├── media_service.go   # 媒体管理
├── search_service.go  # 搜索服务
├── rss_service.go     # RSS 生成
└── notification_*.go  # 通知渠道
```

**职责边界**:
- 封装业务逻辑（验证、默认值、业务规则）
- 协调多个 Repository
- 返回 domain models

#### `internal/core/` - 领域模型层

```
core/
├── user/
│   ├── models.go      # User, Role, Permission 实体
│   └── repository.go  # Repository 接口定义
├── post/
│   ├── models.go      # Post, Category, Tag, Series 实体
│   └ repository.go    # Repository 接口定义
├── comment/
├── media/
├── search/
├── rss/
└── notify/
```

**职责边界**:
- 定义领域实体（Entity）和值对象
- 定义 Repository 接口（抽象）
- 不包含具体实现

#### `internal/database/` - 数据访问层

```
database/
├── connection.go       # pgx 连接池管理
├── config.go           # 数据库配置
├── transaction.go      # 事务管理器
├── user_repository.go  # User Repository 实现
├── post_repository.go  # Post Repository 实现
├── comment_repository.go
├── role_repository.go
└── ...
```

**职责边界**:
- 实现 `internal/core/*/repository.go` 中定义的接口
- 封装 SQL 查询和数据持久化
- 使用 `pgx/v5` 连接池

#### `internal/auth/` - 认证模块

```
auth/
├── config.go           # JWT 配置
├── jwt.go              # JWT 生成/验证
├── service.go          # AuthService
├── blacklist.go        # Token 黑名单接口
├── permission_cache.go # 权限缓存 (Redis)
```

**职责边界**:
- JWT token 生命周期管理
- Token 黑名单（登出撤销）
- 权限缓存（减少数据库查询）

#### `internal/cache/` - 缓存模块

```
cache/
├── redis.go      # Redis 客户端封装
├── service.go    # 缓存服务（穿透/击穿防护）
└── keys.go       # Key 生成函数
```

**职责边界**:
- Redis 连接管理
- 缓存穿透防护（空值标记）
- 缓存击穿防护（分布式锁）

#### `internal/plugin/` - 插件系统

```
plugin/
├── doc.go       # 文档和设计说明
├── plugin.go    # Plugin 接口定义
└── registry.go  # 全局注册表
```

**职责边界**:
- 定义 Plugin 接口
- 提供注册/获取机制
- 支持依赖验证

#### `internal/theme/` - 主题系统

```
theme/
├── models.go     # Theme 实体定义
└── registry.go   # 主题注册表
```

---

## 3. 依赖关系与注入方式

### 3.1 依赖方向

遵循**单向依赖**原则:

```
cmd/server → internal/api → internal/services → internal/core → internal/database
                    ↓              ↓
              internal/auth    internal/cache
```

- 上层依赖下层，下层不依赖上层
- `internal/core` 定义接口，`internal/database` 实现接口

### 3.2 依赖注入模式

采用**构造函数注入** + **Service Container** 模式。

#### Service Container 定义

`internal/services/services.go:19-33`:

```go
type Container struct {
    UserService        UserService
    AuthService        AuthService
    PostService        PostService
    CategoryService    CategoryService
    TagService         TagService
    SeriesService      SeriesService
    CommentService     CommentService
    MediaService       MediaService
    NotificationService NotificationService
    RSSService         RSSService
    SearchService      SearchService
    jwtService         *auth.JWTService  // 私有，通过方法暴露
}
```

#### 注入函数层次

Container 提供多个构造函数，按功能渐进式组装:

| 构造函数 | 注入的服务 | 使用场景 |
|---------|-----------|---------|
| `NewContainer` | User, Auth | 最小容器 |
| `NewContainerWithBlacklist` | User, Auth + Token黑名单 | 登出支持 |
| `NewContainerWithPosts` | + Post, Category, Tag, Series | 文章管理 |
| `NewContainerWithComments` | + Comment | 评论系统 |
| `NewContainerWithMedia` | + Media, RSS, Search | 媒体+搜索 |
| `NewContainerWithNotifications` | + Notification | 完整功能 |

**设计意图**: 支持不同部署场景的功能裁剪。

### 3.3 启动流程中的依赖组装

`cmd/server/main.go:45-141` 展示了完整的初始化顺序:

```
┌─────────────────────────────────────────────────────────────┐
│  1. 配置加载                                                 │
│     loadConfig() → Config{Database, Redis, JWT, Server}     │
├─────────────────────────────────────────────────────────────┤
│  2. 基础设施初始化                                           │
│     database.NewPool() → pgx 连接池                         │
│     cache.NewRedisClient() → Redis 客户端                    │
│     cache.NewService() → 缓存服务                            │
├─────────────────────────────────────────────────────────────┤
│  3. Repository 初始化                                        │
│     NewUserRepository, NewPostRepository, ...               │
│     NewTransactionManager → 事务管理器                       │
├─────────────────────────────────────────────────────────────┤
│  4. 认证组件初始化                                           │
│     auth.NewJWTService() → JWT 服务                         │
│     auth.NewRedisTokenBlacklist() → Token 黑名单             │
│     auth.NewPermissionCache() → 权限缓存                     │
├─────────────────────────────────────────────────────────────┤
│  5. 限流器初始化                                             │
│     middleware.NewRateLimiter(login/public/user)            │
├─────────────────────────────────────────────────────────────┤
│  6. Service Container 组装                                  │
│     services.NewContainerWithMedia(...) → 完整容器          │
├─────────────────────────────────────────────────────────────┤
│  7. Handler 初始化                                           │
│     handlers.NewAuthHandler, NewPostHandler, ...            │
├─────────────────────────────────────────────────────────────┤
│  8. 路由配置                                                 │
│     RouteDeps 聚合 → setupRoutes()                          │
├─────────────────────────────────────────────────────────────┤
│  9. HTTP Server 启动                                         │
│     http.Server{Addr, Handler, Timeout} → ListenAndServe    │
└─────────────────────────────────────────────────────────────┘
```

### 3.4 RouteDeps 路由依赖聚合

`cmd/server/routes.go:17-49` 定义了路由层依赖聚合结构:

```go
type RouteDeps struct {
    // Handlers
    AuthHandler    *handlers.AuthHandler
    PostHandler    *handlers.PostHandler
    // ...

    // Middleware
    LoginLimiter   *middleware.RateLimiter
    PublicLimiter  *middleware.RateLimiter
    UserLimiter    *middleware.RateLimiter

    // Auth
    JWTService     *auth.JWTService
    TokenBlacklist auth.TokenBlacklist
    PermCache      *auth.PermissionCache

    // Repositories (for admin pages direct access)
    PostRepo       *database.PostRepository
    CommentRepo    *database.CommentRepository

    // Services
    Services       *services.Container

    // Config
    UploadDir      string
}
```

**设计模式**: 将所有路由需要的依赖聚合到一个结构体，避免全局变量。

---

## 4. 配置管理方式

### 4.1 配置结构

`cmd/server/config.go:12-33`:

```go
type Config struct {
    Database database.Config   // PostgreSQL 配置
    Redis    cache.Config      // Redis 配置
    JWT      auth.JWTConfig    // JWT 密钥配置
    Server   ServerConfig      // HTTP 服务配置
    Upload   UploadConfig      // 上传配置
}
```

### 4.2 加载方式

采用**环境变量驱动**，无配置文件:

```go
func loadConfig() *Config {
    return &Config{
        Database: database.Config{
            Host:     getEnvOrDefault("DB_HOST", "localhost"),
            Port:     atoi(getEnvOrDefault("DB_PORT", "5432")),
            // ...
        },
        Redis: cache.Config{
            Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
            // ...
        },
        Server: ServerConfig{
            Port:         getEnvOrDefault("PORT", "8080"),
            ReadTimeout:  15 * time.Second,
            WriteTimeout: 15 * time.Second,
            IdleTimeout:  60 * time.Second,
        },
    }
}
```

### 4.3 环境变量清单

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PORT` | `8080` | HTTP 服务端口 |
| `DB_HOST` | `localhost` | PostgreSQL 主机 |
| `DB_PORT` | `5432` | PostgreSQL 端口 |
| `DB_NAME` | `cadmus` | 数据库名 |
| `DB_USER` | `cadmus` | 数据库用户 |
| `DB_PASSWORD` | `""` | 数据库密码 |
| `DB_SSLMODE` | `disable` | SSL 模式 |
| `REDIS_HOST` | `localhost` | Redis 主机 |
| `REDIS_PORT` | `6379` | Redis 端口 |
| `REDIS_PASSWORD` | `""` | Redis 密码 |
| `UPLOAD_DIR` | `./uploads` | 上传目录 |
| `BASE_URL` | `http://localhost:8080` | 基础 URL |

---

## 5. 插件与主题系统

### 5.1 插件系统

采用**编译时注册**模式:

```go
// plugins/mermaid-block/plugin.go
func init() {
    plugin.RegisterPlugin(func() plugin.Plugin {
        return &MermaidBlockPlugin{}
    })
}
```

**注册机制** (`internal/plugin/registry.go:17-42`):
- 使用 `init()` 函数自动注册
- Blank import 触发: `_ "rua.plus/cadmus/plugins/mermaid-block"`
- 全局 `pluginMap` 存储构造函数
- 支持依赖验证

### 5.2 主题系统

同样采用编译时注册:

```go
// themes/default/theme.go
func init() {
    theme.Register(Theme{
        ID:          "default",
        Name:        "Default Theme",
        Version:     "1.0.0",
        Description: "Default Cadmus theme",
    })
}
```

**注册表** (`internal/theme/registry.go`):
- 单例模式 (`sync.Once`)
- 支持激活主题切换
- 线程安全 (`sync.RWMutex`)

---

## 6. 中间件架构

### 6.1 认证中间件链

```
Request → RateLimitMiddleware → AuthMiddleware → PermissionMiddleware → Handler
                ↓                      ↓                ↓
           Redis 限流            JWT 验证 + 黑名单    RBAC 权限检查
```

### 6.2 权限中间件类型

`internal/api/handlers/middleware.go` 提供三种权限中间件:

| 中间件 | 检查方式 | 适用场景 |
|--------|---------|---------|
| `PermissionMiddleware` | 直接查库 | 低频、敏感操作 |
| `CachedPermissionMiddleware` | Redis 缓存 | 高频操作 |
| `AdminMiddleware` | 缓存 + `admin:access` | 管理后台入口 |

### 6.3 限流策略

`internal/api/middleware/ratelimit.go:181-197`:

| 限流器 | 配额 | 窗口 | 保护对象 |
|--------|------|------|---------|
| `LoginLimiter` | 10 次 | 1 分钟 | 登录接口 |
| `PublicLimiter` | 60 次 | 1 分钟 | 公开 API |
| `UserLimiter` | 100 次 | 1 分钟 | 认证用户操作 |

**实现**: Redis ZSET 滑动窗口算法。

---

## 7. 架构优缺点评估

### 7.1 优点

| 方面 | 具体表现 |
|------|---------|
| **分层清晰** | Handler → Service → Repository 三层边界明确 |
| **依赖注入** | Service Container 模式，避免全局变量 |
| **接口抽象** | Repository 接口在 `core` 定义，实现在 `database` |
| **插件扩展** | 编译时注册，零运行时开销 |
| **模板性能** | Templ 编译生成 Go 代码，无模板解析开销 |
| **缓存设计** | 穿透/击穿防护，分布式锁 |
| **权限缓存** | Redis 缓存权限，减少数据库压力 |
| **限流完善** | 滑动窗口 + 不同场景策略 |
| **事务管理** | 封装 `TransactionManager`，自动提交/回滚 |

### 7.2 待改进点

| 方面 | 当前状态 | 建议改进 |
|------|---------|---------|
| **Container 构造函数** | 6 个变体函数 | 支持渐进式组装，但可考虑 Builder 模式简化 |
| **配置管理** | 环境变量 + 硬编码默认值 | 可独立为 `internal/config` 包 |
| **日志系统** | 使用 `log` 标准库 | 可引入结构化日志 (zap/zerolog) |
| **错误处理** | 自定义 Error 类型 | 可统一错误码体系 |
| **测试覆盖** | 未发现测试文件 | 需补充单元测试 |
| **API 文档** | 无 OpenAPI/Swagger | 可添加 API 规范文档 |

### 7.3 依赖关系图

```
┌──────────────────────────────────────────────────────────────────┐
│                         External Dependencies                     │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────┐ │
│  │  pgx/v5 │ │go-redis │ │ golang- │ │ google/ │ │   a-h/templ │ │
│  │PostgreSQL│ │   v9    │ │  jwt/v5 │ │  uuid   │ │   (前端)    │ │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └──────┬──────┘ │
└───────┼──────────┼──────────┼──────────┼───────────────┼─────────┘
        │          │          │          │               │
        ▼          ▼          ▼          ▼               ▼
┌──────────────────────────────────────────────────────────────────┐
│                         Internal Modules                          │
│                                                                   │
│  cmd/server ─────────────────────────────────────────────────────│
│       │                                                           │
│       ├── internal/api/handlers ←─────────────────────────────────│
│       │       │                                                   │
│       │       ├── internal/api/middleware ←───────────────────────│
│       │       │                                                   │
│       │       ├── internal/services ←─────────────────────────────│
│       │       │       │                                           │
│       │       │       ├── internal/auth ←─────────────────────────│
│       │       │       │       │                                   │
│       │       │       │       └── internal/cache ←────────────────│
│       │       │       │                                           │
│       │       │       └── internal/core ←─────────────────────────│
│       │       │               │                                   │
│       │       │               └── internal/database ←─────────────│
│       │       │                                                   │
│       │       └── internal/plugin ←───────────────────────────────│
│       │                                                           │
│       ├── internal/theme ←────────────────────────────────────────│
│       │                                                           │
│       └── web/templates ←─────────────────────────────────────────│
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

---

## 8. 总结

Cadmus 采用了成熟的三层架构设计，依赖注入模式清晰，插件和主题系统设计合理。整体代码结构符合 Go 项目最佳实践，模块边界明确。

**核心亮点**:
1. Repository 接口与实现分离，支持未来更换存储
2. Service Container 渐进式组装，支持功能裁剪
3. Redis 用于限流、缓存、Token 黑名单，基础设施统一
4. Templ 编译时模板生成，性能优异

**下一步建议**:
1. 补充单元测试和集成测试
2. 引入结构化日志库
3. 添加 API 文档规范 (OpenAPI)
4. 考虑配置文件支持 (YAML)

---

## 参考文献

- `cmd/server/main.go:45-208` - 启动流程
- `cmd/server/config.go:1-121` - 配置管理
- `cmd/server/routes.go:1-265` - 路由配置
- `internal/services/services.go:1-227` - 服务容器
- `internal/plugin/registry.go:1-128` - 插件注册
- `internal/theme/registry.go:1-129` - 主题注册
- `internal/auth/permission_cache.go:1-174` - 权限缓存
- `internal/cache/service.go:1-262` - 缓存服务
- `internal/api/handlers/middleware.go:1-288` - 认证中间件
- `internal/api/middleware/ratelimit.go:1-197` - 限流实现