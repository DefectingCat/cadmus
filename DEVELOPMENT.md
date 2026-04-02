# 开发指南

本文档介绍 Cadmus 项目的本地开发环境配置和开发流程。

---

## 1. 项目概述

### 项目简介

Cadmus 是一个基于 Go 语言开发的现代博客平台，采用清晰的分层架构设计，提供完整的文章管理、评论系统、媒体管理和后台管理功能。

### 技术栈

| 类别     | 技术                    | 版本      |
| -------- | ----------------------- | --------- |
| 后端语言 | Go                      | 1.26.1    |
| 数据库   | PostgreSQL (pgx driver) | 18.3      |
| 缓存     | Redis (go-redis)        | 7+        |
| 认证     | JWT (golang-jwt)        | v5.3.1    |
| 模板引擎 | templ                   | v0.3.1001 |
| 前端脚本 | TypeScript              | 6.0.2     |
| 样式框架 | TailwindCSS             | 4.2.2     |
| 构建工具 | esbuild, bun            | -         |
| 测试框架 | testify                 | v1.11.1   |

---

## 2. 环境配置

### Go 环境要求

- Go 1.26.1 或更高版本
- 确保 `GOPATH` 和 `GOROOT` 正确配置

验证 Go 版本：

```bash
go version
# 输出应为: go version go1.26.x ...
```

### 前端环境要求

- **bun** - 用于前端资源构建和包管理

安装 bun（如未安装）：

```bash
# macOS/Linux
curl -fsSL https://bun.sh/install | bash

# 或通过 npm
npm install -g bun
```

验证安装：

```bash
bun --version
```

### 数据库环境

- PostgreSQL 14+（推荐 18.3）
- Redis 7+

---

## 3. PostgreSQL 开发服务器

项目提供了专用 Dockerfile 用于快速搭建 PostgreSQL 开发环境。

### Dockerfile 说明

`Dockerfile.postgres` 基于 PostgreSQL 18.3 官方镜像（最新稳定版本），预配置了：

- 默认数据库：`cadmus`
- 默认用户：`cadmus`
- 默认密码：`cadmus`
- 健康检查配置
- 数据持久化卷

### 构建镜像

```bash
docker build -f Dockerfile.postgres -t cadmus-postgres .
```

### 运行容器

#### 方式一：基础运行（使用默认配置）

```bash
docker run -d --name cadmus-db -p 5432:5432 -v ./data:/var/lib/postgresql/data cadmus-postgres
```

#### 方式二：自定义环境变量

```bash
docker run -d --name cadmus-db -p 5432:5432 \
  -e POSTGRES_DB=cadmus \
  -e POSTGRES_USER=cadmus \
  -e POSTGRES_PASSWORD=your_password \
  -v ./data:/var/lib/postgresql/data \
  cadmus-postgres
```

**环境变量说明：**

| 变量                | 说明         | 默认值    |
| ------------------- | ------------ | --------- |
| `POSTGRES_DB`       | 数据库名称   | `cadmus`  |
| `POSTGRES_USER`     | 数据库用户   | `cadmus`  |
| `POSTGRES_PASSWORD` | 数据库密码   | `cadmus`  |

### 数据持久化

数据默认存储在项目根目录的 `data/` 文件夹下：

```bash
# 数据目录结构
data/
└── pgdata/           # PostgreSQL 数据文件
```

**查看数据目录：**

```bash
ls -la data/
```

### 连接数据库

```bash
# 使用 psql 连接
psql -h localhost -p 5432 -U cadmus -d cadmus

# 或通过 Docker
docker exec -it cadmus-db psql -U cadmus -d cadmus
```

### 停止和清理

```bash
# 停止容器
docker stop cadmus-db

# 删除容器
docker rm cadmus-db

# 删除数据目录（谨慎操作，会清除所有数据）
rm -rf data/
```

---

## 4. 本地开发启动

### 环境变量配置

1. 复制示例配置文件：

```bash
cp .env.example .env
```

2. 编辑 `.env` 文件，根据实际情况调整配置：

```bash
# 服务配置
PORT=8080
BASE_URL=http://localhost:8080
UPLOAD_DIR=./uploads

# PostgreSQL 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=cadmus
DB_USER=cadmus
DB_PASSWORD=cadmus
DB_SSLMODE=disable

# Redis 缓存配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# 日志配置
LOG_LEVEL=info
LOG_SHOW_CALLER=false
LOG_SHOW_TIMESTAMP=true
```

### 初始化数据库

执行迁移脚本初始化数据库结构：

```bash
# 确保数据库已启动
psql -d cadmus -f migrations/001_init.up.sql
psql -d cadmus -f migrations/002_create_posts.up.sql
psql -d cadmus -f migrations/003_create_comments.up.sql
psql -d cadmus -f migrations/004_create_post_likes.up.sql
psql -d cadmus -f migrations/005_create_media.up.sql
```

或使用一键脚本：

```bash
for f in migrations/*.up.sql; do psql -d cadmus -f "$f"; done
```

### 构建项目

```bash
# 全量构建（前端 + 后端）
make build

# 仅构建后端
make build/backend

# 仅构建前端
make build/frontend

# 生成 templ 模板文件
make build/templ

# 显示版本信息
make version
```

构建产物位于 `bin/` 目录。

### 运行测试

```bash
# 运行所有测试（含竞态检测）
make test

# 生成覆盖率报告
make test/coverage
# 报告输出: coverage.html

# 运行基准测试
make test/bench
```

### 启动服务

```bash
# 直接运行
./bin/server

# 或使用 go run（开发调试）
go run ./cmd/server
```

服务启动后访问：

- 主页：http://localhost:8080
- 管理后台：http://localhost:8080/admin
- 健康检查：http://localhost:8080/health

---

## 5. 目录结构说明

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
│   │   ├── src/
│   │   │   ├── main.ts      # 主入口
│   │   │   ├── admin/       # 管理后台脚本
│   │   │   ├── editor/      # 编辑器脚本
│   │   │   └── styles/      # CSS 样式
│   │   └── package.json     # 前端依赖配置
│   ├── static/          # 静态资源目录
│   │   └── dist/        # 构建输出
│   └── templates/       # templ 模板文件
├── bin/                 # 构建输出目录
├── test/                # 测试辅助文件
├── docs/                # 文档目录
├── Makefile             # 构建命令
├── go.mod               # Go 模块定义
├── .env.example         # 环境变量示例
├── Dockerfile.postgres  # PostgreSQL 开发镜像
└── README.md            # 项目说明
```

**主要目录说明：**

| 目录          | 说明                                     |
| ------------- | ---------------------------------------- |
| `cmd/server/` | 应用启动入口，包含路由注册和依赖初始化   |
| `internal/`   | 内部实现，不对外暴露                     |
| `internal/api/` | HTTP 处理层，接收请求并调用服务层      |
| `internal/services/` | 业务逻辑层，协调数据访问和业务规则 |
| `internal/database/` | 数据访问层，Repository 实现        |
| `internal/core/` | 领域模型定义                          |
| `migrations/` | SQL 迁移脚本，按序号执行                 |
| `web/frontend/` | 前端 TypeScript 源码和样式            |
| `web/templates/` | templ 模板文件                        |
| `plugins/`    | 可选插件，编译时注册                     |
| `themes/`     | 主题实现，可替换前端渲染                 |

---

## 6. 常见问题

### Q: 数据库连接失败？

检查以下项：

1. PostgreSQL 是否正常运行
2. `.env` 中的数据库配置是否正确
3. 数据库用户是否有足够权限
4. SSL 模式配置（本地开发通常设为 `disable`）

```bash
# 测试连接
psql -h localhost -U cadmus -d cadmus
```

### Q: Redis 连接失败？

1. 确保 Redis 服务已启动
2. 检查端口和密码配置
3. 本地开发通常无需密码

```bash
# 启动 Redis（如果未运行）
redis-server

# 测试连接
redis-cli ping
# 应返回: PONG
```

### Q: 前端构建失败？

1. 确保已安装 bun
2. 进入 `web/frontend` 目录检查依赖

```bash
cd web/frontend
bun install
```

### Q: templ 文件未生成？

运行 templ 生成命令：

```bash
make build/templ
# 或
templ generate
```

### Q: 测试失败？

1. 确保数据库已初始化
2. 检查测试所需的表和数据
3. 使用 `-v` 参数查看详细输出

```bash
go test -v -race ./...
```

### Q: 如何添加新的插件？

参考 README.md 中的「插件开发」章节，通过 `blank import` 机制在编译时注册。

### Q: 如何更换主题？

参考 README.md 中的「主题开发」章节，实现 `theme.Theme` 接口并在 `cmd/server/main.go` 中注册。

---

## 附录：依赖版本

### Go 依赖

| 包                      | 版本      | 说明              |
| ----------------------- | --------- | ----------------- |
| `github.com/a-h/templ`  | v0.3.1001 | 模板引擎          |
| `github.com/golang-jwt/jwt/v5` | v5.3.1 | JWT 认证      |
| `github.com/google/uuid` | v1.6.0  | UUID 生成         |
| `github.com/jackc/pgx/v5` | v5.9.1 | PostgreSQL 驱动 |
| `github.com/redis/go-redis/v9` | v9.18.0 | Redis 客户端 |
| `golang.org/x/crypto`   | v0.49.0   | 密码加密（bcrypt）|
| `github.com/stretchr/testify` | v1.11.1 | 测试框架      |

### 前端依赖

| 包             | 版本    | 说明           |
| -------------- | ------- | -------------- |
| `typescript`   | 6.0.2   | TypeScript     |
| `tailwindcss`  | 4.2.2   | CSS 框架       |
| `esbuild`      | 0.27.4  | 构建打包       |
| `@biomejs/biome` | 2.4.10 | 代码检查和格式化 |