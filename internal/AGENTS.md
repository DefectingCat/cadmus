<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# internal

## Purpose
私有应用代码目录，包含核心业务逻辑、API 处理器、认证、缓存、数据库访问层。

## Key Files
| File | Description |
|------|-------------|
| 无顶层文件 | 按模块分目录组织 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `api/handlers/` | REST API 处理器 (see `api/handlers/AGENTS.md`) |
| `auth/` | JWT 认证 + Token 黑名单 + 权限缓存 (see `auth/AGENTS.md`) |
| `cache/` | Redis 缓存服务 (see `cache/AGENTS.md`) |
| `core/` | 核心领域模型：user/post/comment/media/search/rss/notify/block (see `core/AGENTS.md`) |
| `database/` | PostgreSQL 数据访问层 (see `database/AGENTS.md`) |
| `plugin/` | 插件注册引擎 (see `plugin/AGENTS.md`) |
| `services/` | 业务服务层 (see `services/AGENTS.md`) |
| `theme/` | 主题注册引擎 (see `theme/AGENTS.md`) |

## For AI Agents

### Working In This Directory
- **分层架构**: core (领域模型) → database (数据访问) → services (业务逻辑) → api/handlers (HTTP 处理)
- **依赖方向**: handlers → services → database → core
- **禁止跨层调用**: handlers 不应直接调用 database，需通过 services

### Architecture Pattern
```
Request → Handler → Service → Repository → Database
                ↓
              Cache (Redis)
```

### Module Organization
| Layer | Responsibility |
|-------|----------------|
| `core/` | 领域模型定义 + Repository 接口 |
| `database/` | Repository 实现 (pgx) |
| `services/` | 业务逻辑封装 + 跨 Repository 协调 |
| `api/handlers/` | HTTP 请求处理 + 响应格式化 |

<!-- MANUAL: -->