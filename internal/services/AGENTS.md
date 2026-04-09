<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# services

## Purpose

`services` 目录实现 Cadmus 应用的业务服务层（Service Layer），采用接口抽象 + 依赖注入模式，将业务逻辑与数据访问层（Repository）和 HTTP 处理器（Handler）解耦。服务层是领域业务规则的核心承载者，负责跨 Repository 的事务协调和数据验证。

## Service Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Handler                           │
│  (internal/api/handlers/)                                   │
│  - 请求参数解析                                             │
│  - 响应序列化                                               │
└─────────────────────────────────────────────────────────────┘
                            │ 调用
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Service 接口层                             │
│  (internal/services/*_service.go)                           │
│  - 业务规则验证                                             │
│  - 跨 Repository 事务协调                                     │
│  - 领域模型转换                                             │
│  - 通知/缓存触发                                            │
└─────────────────────────────────────────────────────────────┘
                            │ 调用
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                 Repository 接口层                             │
│  (internal/core/*/repository.go)                            │
│  - 数据持久化接口定义                                       │
└─────────────────────────────────────────────────────────────┘
                            │ 实现
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              Repository 实现层                                │
│  (internal/database/*_repository.go)                        │
│  - SQL 查询执行                                              │
│  - 事务管理                                                 │
└─────────────────────────────────────────────────────────────┘
```

## Key Files

| File | Purpose | Key Interfaces |
|------|---------|----------------|
| `services.go` | 服务容器 Container，聚合所有服务；依赖注入入口 | `Container`, `AuthService`, `TokenBlacklist` |
| `auth_service.go` | 用户认证服务：登录/登出/Token 验证/黑名单管理 | `AuthService` |
| `user_service.go` | 用户管理服务：注册/查询/更新/删除 | `UserService` |
| `post_service.go` | 文章管理服务：CRUD/发布/版本管理/点赞 | `PostService`, `CategoryService`, `TagService`, `SeriesService` |
| `comment_service.go` | 评论管理服务：CRUD/审核/点赞/树形结构构建 | `CommentService`, `CommentNode` |
| `media_service.go` | 媒体文件服务：上传/查询/删除/MIME 验证 | `MediaService` |
| `notification_service.go` | 通知推送服务：评论通知/回复通知 | `NotificationService` |
| `email_channel.go` | 邮件通知渠道实现：SMTP 发送/模板渲染 | `EmailChannel` |
| `rss_service.go` | RSS 订阅源生成：RSS 2.0 XML 生成/分类筛选 | `RSSService` |
| `search_service.go` | 全文搜索服务：多条件筛选/搜索建议 | `SearchService` |

## No Subdirectories

`services` 目录采用扁平结构，所有服务文件平铺在根目录下，便于查找和维护。

## Service Container 模式

项目使用 `Container` 结构聚合所有服务，通过构造函数实现渐进式功能启用：

```go
// 基础容器（仅用户和认证）
container := services.NewContainer(userRepo, roleRepo, jwtService)

// 带黑名单的容器（支持登出即时失效）
container := services.NewContainerWithBlacklist(userRepo, roleRepo, jwtService, blacklist)

// 完整容器（包含文章、评论、媒体、通知等）
container := services.NewContainerWithNotifications(
    userRepo, roleRepo, jwtService, blacklist,
    postRepo, categoryRepo, tagRepo, seriesRepo,
    commentRepo, commentLikeRepo,
    mediaRepo, uploadDir, baseURL,
    postLikeRepo, searchRepo,
    notificationChannel,
)
```

## Service Pattern Guide for AI Agents

### 1. 接口定义规范

所有服务必须通过接口抽象，便于测试 Mock：

```go
type UserService interface {
    Register(ctx context.Context, username, email, password string) (*user.User, error)
    GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
    // ...
}
```

### 2. 实现结构体命名

采用 `接口名 + Impl` 模式，私有可见性：

```go
type userServiceImpl struct {
    userRepo user.UserRepository
    roleRepo user.RoleRepository
}
```

### 3. 构造函数命名

- 基础版本：`New<ServiceName>`（如 `NewUserService`）
- 可选功能：`New<ServiceName>With<Feature>`（如 `NewPostServiceWithLikes`）

### 4. 业务规则位置

**Service 层负责**：
- 输入验证（必填字段、格式校验）
- 权限检查（用户状态、资源归属）
- 跨 Repository 协调（如创建用户时获取默认角色）
- 通知触发（评论后通知作者）

**Repository 层仅负责**：
- 单一模型的 CRUD
- SQL 查询构建
- 事务内数据一致性

### 5. 错误处理模式

```go
// Service 层返回语义化错误
if input.Content == "" {
    return nil, comment.ErrEmptyContent
}

// 跨 Repository 错误转换
if err := s.userRepo.Create(ctx, newUser); err != nil {
    return nil, err // 直接返回 Repository 错误
}
```

### 6. 并发安全要求

所有公开方法必须并发安全：
- 避免修改结构体字段（不可变设计）
- Repository 调用本身应是并发安全的
- 点赞等计数操作使用原子方法（`CreateIfNotExists`）

### 7. 树形结构构建

评论嵌套使用 `CommentNode` 构建森林结构：

```go
type CommentNode struct {
    Comment  *comment.Comment
    Children []*CommentNode
}

// buildCommentTree 将扁平列表转换为树
func (s *commentServiceImpl) buildCommentTree(comments []*comment.Comment) []*CommentNode {
    nodeMap := make(map[uuid.UUID]*CommentNode)
    // ... 构建树形逻辑
}
```

## Dependencies

### 内部依赖

```
services/
├── 依赖 core/       ← 领域模型、Repository 接口定义
├── 依赖 database/   ← Repository 实现（可选批量查询优化）
├── 依赖 auth/       ← JWT 服务（Token 生成/验证）
└── 依赖 logger/     ← 日志记录（媒体删除失败警告）
```

### 外部依赖

| Package | 用途 |
|---------|------|
| `github.com/google/uuid` | UUID 生成 |
| `text/template` | 邮件模板渲染 |
| `net/smtp` | SMTP 邮件发送 |
| `encoding/xml` | RSS XML 序列化 |

## Development Workflow for AI Agents

### 新增 Service

1. 在 `core/<domain>/repository.go` 定义 Repository 接口
2. 在 `database/<domain>_repository.go` 实现 Repository
3. 在 `services/<domain>_service.go` 定义 Service 接口和实现
4. 在 `services/services.go` 添加字段到 `Container` 并更新构造函数
5. 在 `api/handlers/<domain>_handler.go` 调用 Service

### 修改业务规则

仅修改 `services/<domain>_service.go` 中的对应方法，不要改动 Handler 或 Repository。

### 单元测试

测试 Service 层时使用 Mock Repository：

```go
// 示例：测试用户注册
func TestUserService_Register(t *testing.T) {
    mockUserRepo := &mocks.UserRepository{}
    mockRoleRepo := &mocks.RoleRepository{}
    svc := services.NewUserService(mockUserRepo, mockRoleRepo)
    
    // 执行测试...
}
```

## For AI Agents

- 所有 Service 方法必须并发安全
- 不要在 Service 层直接操作 HTTP 请求/响应
- 不要在 Repository 层包含业务验证逻辑
- 使用 `Container` 统一管理依赖，避免全局变量
- 新增可选功能时使用 `With<Feature>` 链式方法
