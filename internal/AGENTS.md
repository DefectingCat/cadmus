<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# internal

## Purpose

`internal` 目录是 Cadmus 应用的核心业务逻辑层，采用 Go 的 `internal` 包可见性规则，确保代码不被外部项目导入。包含完整的分层架构实现：领域模型、数据访问层、业务服务层、API 处理器、认证系统、缓存服务和插件/主题注册引擎。

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `api/` | REST API 处理器和 HTTP 中间件 (see `api/AGENTS.md`) |
| `auth/` | JWT 认证 + Token 黑名单 + 权限缓存 (see `auth/AGENTS.md`) |
| `cache/` | Redis 缓存服务，管理键和缓存策略 (see `cache/AGENTS.md`) |
| `core/` | 核心领域模型定义：user/post/comment/media/search/rss/notify (see `core/AGENTS.md`) |
| `database/` | PostgreSQL 数据访问层，Repository 模式实现 (see `database/AGENTS.md`) |
| `logger/` | 结构化日志记录器 (see `logger/AGENTS.md`) |
| `plugin/` | 插件注册引擎，支持可插拔扩展 (see `plugin/AGENTS.md`) |
| `services/` | 业务服务层，封装业务逻辑和跨 Repository 协调 (see `services/AGENTS.md`) |
| `theme/` | 主题注册引擎，管理主题加载和切换 (see `theme/AGENTS.md`) |

## Architecture Guide for AI Agents

### 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Request                           │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  api/handlers/  - HTTP 请求处理、参数解析、响应格式化         │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  services/ - 业务逻辑封装、事务协调、跨模型操作               │
└─────────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┴───────────────┐
            │                               │
            ▼                               ▼
┌───────────────────────┐       ┌─────────────────────────────┐
│  database/            │       │  cache/                     │
│  Repository 实现        │       │  Redis 缓存服务              │
│  (pgx/v5)             │       │  (go-redis/v9)              │
└───────────────────────┘       └─────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────┐
│  core/ - 领域模型 (Domain Models) + Repository 接口定义      │
└─────────────────────────────────────────────────────────────┘
```

### 依赖注入

项目采用 Service Container 模式管理服务依赖：

```go
// services/services.go 定义 ServiceContainer
type ServiceContainer struct {
    PostService    *post.PostService
    CommentService *comment.CommentService
    UserService    *user.UserService
    // ...
}

// 在 cmd/server/main.go 中组装
container := services.NewServiceContainer(db, redisClient, jwtConfig)
```

**依赖方向规则**:
- `handlers` → `services` → `database/repository` → `core`
- `handlers` 不应直接调用 `database`，必须通过 `services`
- `services` 可调用多个 `repository` 实现跨模型协调

### 核心模块职责

| 模块 | 职责 | 关键文件 |
|------|------|----------|
| `core/` | 领域模型定义、Repository 接口、DTO | `models.go`, `repository.go` |
| `database/` | Repository 实现、SQL 查询、事务管理 | `*_repository.go`, `transaction.go` |
| `services/` | 业务规则、验证、授权检查、通知触发 | `*_service.go` |
| `api/handlers/` | HTTP 路由、请求绑定、响应发送 | `*_handler.go` |
| `auth/` | JWT 签发/验证、Token 黑名单、权限缓存 | `jwt.go`, `permission_cache.go` |
| `cache/` | Redis 连接、键管理、缓存策略 | `keys.go`, `redis.go` |

### 认证流程

```
Request → AuthMiddleware (JWT 验证) → PermissionMiddleware (缓存检查) → Handler
                │                              │
                ▼                              ▼
        auth/jwt.go                   cache/permission_cache.go
```

### 插件系统

使用 Go 的 blank import 触发 `init()` 函数注册插件：

```go
// 在 cmd/server/main.go 中
import (
    _ "cadmus/plugins/search"
    _ "cadmus/plugins/seo"
)
```

## Dependencies

### 内部依赖关系

```
internal/core/        ← 无内部依赖 (仅定义接口和模型)
    ↑
internal/database/    ← 依赖 core/ (实现 Repository 接口)
    ↑
internal/services/    ← 依赖 database/, core/, cache/, auth/
    ↑
internal/api/         ← 依赖 services/, auth/, logger/
```

### 外部依赖

| Package | Version | 用途 |
|---------|---------|------|
| `github.com/jackc/pgx/v5` | v5.9.1 | PostgreSQL 驱动和连接池 |
| `github.com/redis/go-redis/v9` | v9.18.0 | Redis 客户端 |
| `github.com/golang-jwt/jwt/v5` | v5.3.1 | JWT 令牌处理 |
| `golang.org/x/crypto` | v0.49.0 | 密码哈希 (bcrypt) |
| `github.com/google/uuid` | v1.6.0 | UUID 生成 |

## For AI Agents

### 开发指南

1. **新增业务功能**: 
   - 在 `core/` 定义领域模型
   - 在 `database/` 实现 Repository
   - 在 `services/` 编写业务逻辑
   - 在 `api/handlers/` 添加 HTTP 端点

2. **修改数据查询**: 仅修改 `database/` 中的 Repository 实现

3. **修改业务规则**: 仅修改 `services/` 中的对应 Service

4. **添加新 API**: 在 `api/handlers/` 添加 Handler，调用现有 Service

### 测试策略

- **单元测试**: `services/*_test.go` 覆盖业务逻辑
- **Repository Mock**: `test/mocks/database/` 提供 Mock 实现
- **集成测试**: `test/integration/` 验证完整流程

### 常见模式

- Repository 返回 `*core.Model` 或 `[]*core.Model`
- Service 返回 `(Result, error)` 或 `(Result, pagination, error)`
- Handler 统一返回 JSON: `{ code, data, message }`
- 错误处理使用 `database/errors.go` 定义的错误类型
