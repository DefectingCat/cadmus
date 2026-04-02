# Cadmus

Cadmus 是一个基于 Go 语言开发的现代博客平台，采用清晰的分层架构设计，
提供完整的文章管理、评论系统、媒体管理和后台管理功能。

---

## 核心特性

### 架构设计

- **分层架构**：Handler -> Service -> Repository，职责清晰分离
- **依赖注入**：通过 Service Container 聚合和管理依赖
- **Repository 模式**：数据访问层抽象，支持多种存储后端
- **插件系统**：编译时注册机制，支持自定义内容块扩展
- **主题系统**：可替换的前端主题，支持自定义页面渲染

### 安全设计

- **密码加密**：使用 bcrypt 算法存储用户密码
- **JWT 认证**：无状态令牌认证，支持令牌刷新机制
- **令牌黑名单**：Redis 存储已注销令牌，防止令牌复用
- **滑动窗口限流**：基于 Redis 的分布式限流，防护暴力破解
- **RBAC 权限体系**：角色-权限映射，支持细粒度访问控制

### 功能模块

- 文章管理：草稿/发布状态、版本控制、回滚支持
- 评论系统：多级评论、点赞机制、审核流程
- 媒体管理：文件上传、元数据存储、关联文章
- 分类与标签：内容组织、slug 映射
- RSS 订阅：标准 RSS/Atom Feed 输出
- 全文搜索：标题/内容搜索、自动补全
- 管理后台：仪表盘、评论审核、角色管理、用户管理

---

## 技术栈

| 类别     | 技术                    |
| -------- | ----------------------- |
| 后端语言 | Go 1.26.1               |
| 数据库   | PostgreSQL (pgx driver) |
| 缓存     | Redis (go-redis)        |
| 认证     | JWT (golang-jwt)        |
| 模板引擎 | templ                   |
| 前端脚本 | TypeScript              |
| 样式框架 | TailwindCSS             |
| 构建工具 | esbuild, bun            |
| 测试框架 | testify                 |

---

## 项目结构

```
cadmus/
├── cmd/
│   └── server/          # 应用入口和路由配置
├── internal/
│   ├── api/             # HTTP 接口层
│   │   ├── handlers/    # 请求处理器
│   │   └── middleware/  # 中间件（限流、认证、权限）
│   ├── auth/            # 认证组件（JWT、黑名单、权限缓存）
│   ├── cache/           # Redis 缓存服务
│   ├── core/            # 领域模型
│   │   ├── comment/     # 评论模型
│   │   ├── media/       # 媒体模型
│   │   ├── notify/      # 通知模型
│   │   ├── post/        # 文章模型
│   │   ├── rss/         # RSS 配置
│   │   ├── search/      # 搜索模型
│   │   └── user/        # 用户模型
│   ├── database/        # Repository 数据层
│   ├── logger/          # 日志组件
│   ├── plugin/          # 插件注册框架
│   ├── services/        # 业务服务层
│   └── theme/           # 主题注册框架
├── migrations/          # 数据库迁移脚本
├── pkg/
│   └── utils/           # 公共工具函数
├── plugins/             # 插件实现
│   └── mermaid-block/   # Mermaid 图表块插件
├── themes/
│   └── default/         # 默认主题
├── web/
│   ├── frontend/        # 前端 TypeScript 源码
│   ├── static/          # 静态资源目录
│   └── templates/       # templ 模板文件
├── bin/                 # 构建输出目录
├── Makefile             # 构建命令
├── go.mod               # Go 模块定义
└── .env.example         # 环境变量示例
```

---

## 快速开始

### 环境要求

- Go 1.26.1+
- PostgreSQL 14+
- Redis 7+
- bun (用于前端构建)

### 安装步骤

1. 克隆仓库

```bash
git clone https://github.com/xfy/cadmus.git
cd cadmus
```

2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，填写实际配置
```

3. 初始化数据库

```bash
# 创建数据库
createdb cadmus

# 执行迁移脚本
psql -d cadmus -f migrations/001_init.up.sql
psql -d cadmus -f migrations/002_create_posts.up.sql
psql -d cadmus -f migrations/003_create_comments.up.sql
psql -d cadmus -f migrations/004_create_post_likes.up.sql
psql -d cadmus -f migrations/005_create_media.up.sql
```

4. 构建并运行

```bash
make build
./bin/server
```

服务将在配置的端口（默认 8080）启动。

---

## 配置说明

通过环境变量进行配置，支持以下参数：

### 服务配置

| 变量         | 说明             | 默认值                  |
| ------------ | ---------------- | ----------------------- |
| `PORT`       | 服务端口         | `8080`                  |
| `BASE_URL`   | 服务基础 URL     | `http://localhost:8080` |
| `UPLOAD_DIR` | 上传文件存储目录 | `./uploads`             |

### PostgreSQL 配置

| 变量          | 说明       | 默认值      |
| ------------- | ---------- | ----------- |
| `DB_HOST`     | 数据库主机 | `localhost` |
| `DB_PORT`     | 数据库端口 | `5432`      |
| `DB_NAME`     | 数据库名称 | `cadmus`    |
| `DB_USER`     | 数据库用户 | `cadmus`    |
| `DB_PASSWORD` | 数据库密码 | `cadmus`    |
| `DB_SSLMODE`  | SSL 模式   | `disable`   |

### Redis 配置

| 变量             | 说明       | 默认值      |
| ---------------- | ---------- | ----------- |
| `REDIS_HOST`     | Redis 主机 | `localhost` |
| `REDIS_PORT`     | Redis 端口 | `6379`      |
| `REDIS_PASSWORD` | Redis 密码 | (空)        |

### 日志配置

| 变量                 | 说明         | 默认值  |
| -------------------- | ------------ | ------- |
| `LOG_LEVEL`          | 日志级别     | `info`  |
| `LOG_SHOW_CALLER`    | 显示调用位置 | `false` |
| `LOG_SHOW_TIMESTAMP` | 显示时间戳   | `true`  |

---

## API 路由概览

### 认证路由 `/api/v1/auth`

| 方法 | 路径        | 说明         | 认证 |
| ---- | ----------- | ------------ | ---- |
| POST | `/register` | 用户注册     | 限流 |
| POST | `/login`    | 用户登录     | 限流 |
| POST | `/logout`   | 用户注销     | 需要 |
| POST | `/refresh`  | 刷新令牌     | -    |
| GET  | `/me`       | 当前用户信息 | 需要 |

### 文章路由 `/api/v1/posts`

| 方法   | 路径             | 说明     | 认证 |
| ------ | ---------------- | -------- | ---- |
| GET    | `/`              | 文章列表 | 公开 |
| GET    | `/{slug}`        | 文章详情 | 公开 |
| POST   | `/`              | 创建文章 | 需要 |
| PUT    | `/{id}`          | 更新文章 | 需要 |
| DELETE | `/{id}`          | 删除文章 | 需要 |
| POST   | `/{id}/publish`  | 发布文章 | 需要 |
| POST   | `/{id}/rollback` | 回滚版本 | 需要 |
| POST   | `/{id}/like`     | 点赞文章 | 需要 |
| GET    | `/{id}/versions` | 版本历史 | 公开 |

### 评论路由 `/api/v1/comments`

| 方法   | 路径             | 说明         | 认证 |
| ------ | ---------------- | ------------ | ---- |
| GET    | `/post/{postId}` | 文章评论列表 | 公开 |
| POST   | `/`              | 创建评论     | 需要 |
| PUT    | `/{id}`          | 更新评论     | 需要 |
| DELETE | `/{id}`          | 删除评论     | 需要 |
| POST   | `/{id}/like`     | 点赞评论     | 需要 |
| PUT    | `/{id}/approve`  | 批准评论     | 权限 |
| PUT    | `/{id}/reject`   | 拒绝评论     | 权限 |

### 管理路由 `/api/v1/admin`

| 方法   | 路径                      | 说明         | 认证   |
| ------ | ------------------------- | ------------ | ------ |
| GET    | `/roles`                  | 角色列表     | 管理员 |
| POST   | `/roles`                  | 创建角色     | 管理员 |
| PUT    | `/roles/{id}`             | 更新角色     | 管理员 |
| DELETE | `/roles/{id}`             | 删除角色     | 管理员 |
| GET    | `/users`                  | 用户列表     | 管理员 |
| PUT    | `/users/{id}/ban`         | 封禁用户     | 管理员 |
| GET    | `/comments`               | 评论管理列表 | 权限   |
| PUT    | `/comments/batch-approve` | 批量批准     | 权限   |

### 其他路由

| 路径                 | 说明           |
| -------------------- | -------------- |
| `/api/v1/categories` | 分类管理       |
| `/api/v1/tags`       | 标签管理       |
| `/api/v1/media`      | 媒体文件管理   |
| `/api/v1/rss`        | RSS Feed       |
| `/api/v1/search`     | 全文搜索       |
| `/admin`             | 管理后台仪表盘 |
| `/admin/comments`    | 评论管理页面   |
| `/health`            | 健康检查       |

---

## 开发指南

### 构建

```bash
# 全量构建（前端 + 后端）
make build

# 仅构建后端
make build/backend

# 仅构建前端
make build/frontend

# 生成 templ 文件
make build/templ

# 显示版本信息
make version
```

### 测试

```bash
# 运行所有测试（含竞态检测）
make test

# 生成覆盖率报告
make test/coverage

# 运行基准测试
make test/bench
```

### 插件开发

插件通过 `blank import` 机制在编译时注册。创建新插件的步骤：

1. 在 `plugins/` 目录下创建插件包
2. 实现 `plugin.BlockPlugin` 接口
3. 在 `init()` 函数中调用 `plugin.Register()`
4. 在 `cmd/server/main.go` 中添加 blank import

示例：

```go
// plugins/my-block/plugin.go
package myblock

import "rua.plus/cadmus/internal/plugin"

func init() {
    plugin.Register(plugin.BlockPlugin{
        Name:        "my-block",
        Description: "自定义内容块",
        Render:      renderFunc,
    })
}
```

### 主题开发

主题通过 `internal/theme` 包注册。创建新主题：

1. 在 `themes/` 目录下创建主题包
2. 实现 `theme.Theme` 接口
3. 在 `init()` 函数中调用 `theme.Register()`
4. 在 `cmd/server/main.go` 中添加 blank import

---

## 部署说明

### 生产环境建议

- 使用环境变量管理敏感配置，避免硬编码
- 配置 PostgreSQL 连接池参数
- 启用 Redis 持久化，确保令牌黑名单可靠性
- 设置适当的限流参数，平衡安全与用户体验
- 使用反向代理（Nginx/Caddy）处理 HTTPS 和静态文件

### 容器化部署

可使用 Docker 进行容器化部署：

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bin/server .
COPY --from=builder /app/web/static ./web/static
EXPOSE 8080
CMD ["./server"]
```

---

## 版本管理

当前版本：`v0.0.1`

版本信息通过 `-ldflags` 在编译时注入：

```bash
go build -ldflags "-s -w \
    -X 'main.version=v0.0.1' \
    -X 'main.gitCommit=$(git rev-parse --short HEAD)' \
    -X 'main.gitBranch=$(git branch --show-current)' \
    -X 'main.buildTime=$(date -u +%Y-%m-%d.%H:%M:%S)' \
    -o bin/server ./cmd/server
```

启动时将显示完整的构建信息。

---

## License

MIT License
