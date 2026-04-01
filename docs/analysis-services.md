# Cadmus 服务层分析报告

## 1. 服务容器模式（Container）

### 1.1 设计理念

Cadmus 采用 **服务容器模式** 来管理所有业务服务的依赖注入和生命周期。容器定义在 `internal/services/services.go`：

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

### 1.2 容器构建工厂方法

容器提供多个渐进式工厂方法，支持按需组装服务：

| 工厂方法 | 包含服务 | 适用场景 |
|---------|---------|---------|
| `NewContainer` | User + Auth | 最小容器，仅需认证 |
| `NewContainerWithBlacklist` | User + Auth + Token黑名单 | 需要登出失效的场景 |
| `NewContainerWithPosts` | + Post + Category + Tag + Series | 博客核心功能 |
| `NewContainerWithComments` | + Comment | 评论系统 |
| `NewContainerWithMedia` | + Media + RSS + Search | 媒体管理与搜索 |
| `NewContainerWithNotifications` | + Notification | 全功能容器 |

### 1.3 设计优势

1. **渐进式组装**：通过不同工厂方法，可按需创建最小依赖容器
2. **依赖集中管理**：所有服务依赖在容器初始化时一次性注入
3. **接口隔离**：容器字段均为接口类型，便于测试和替换实现
4. **私有服务暴露**：`jwtService` 私有，通过 `JWTService()` 方法访问，避免滥用

---

## 2. 服务清单与业务逻辑

项目共定义 **12 个服务接口**，其中 10 个为核心业务服务：

### 2.1 用户域服务

| 服务 | 核心职责 | 关键方法 |
|-----|---------|---------|
| `UserService` | 用户注册、查询、管理 | `Register`, `GetByID`, `GetByEmail`, `Delete` |
| `AuthService` | 认证、Token 管理 | `Login`, `Logout`, `Refresh`, `ValidateToken` |

**业务逻辑亮点**：

- **UserService.Register**：
  - 验证必填字段
  - 检查邮箱/用户名重复
  - 获取默认角色（依赖 RoleRepo）
  - 密码哈希处理
  - 设置初始状态为 `Pending`

- **AuthService.Login**：
  - 查询用户并验证密码
  - 检查用户状态（Banned 用户拒绝登录）
  - 生成 JWT Token
  - 支持 Token 黑名单（可选）

### 2.2 内容域服务

| 服务 | 核心职责 | 关键方法 |
|-----|---------|---------|
| `PostService` | 文章 CRUD、发布、版本管理 | `Create`, `Publish`, `Schedule`, `CreateVersion`, `Rollback`, `LikePost` |
| `CategoryService` | 分类管理 | `Create`, `GetChildren`, `GetPostCount` |
| `TagService` | 标签管理 | `Create`, `GetBySlug`, `GetPostCount` |
| `SeriesService` | 系列管理 | `Create`, `GetByAuthor` |
| `CommentService` | 评论 CRUD、审核、点赞 | `CreateComment`, `ApproveComment`, `LikeComment`, `BatchApproveComments` |

**业务逻辑亮点**：

- **PostService.Create**：
  - 验证标题/Slug 必填
  - 验证状态有效性
  - 默认状态设为 `Draft`
  - 关联标签（多对多）

- **PostService.Publish vs Schedule**：
  - `Publish`：立即设置 `StatusPublished` + 记录 `PublishAt`
  - `Schedule`：设置 `StatusScheduled` + 指定发布时间

- **CommentService.CreateComment**：
  - 嵌套深度检查（`MaxCommentDepth = 5`）
  - 自动计算评论深度
  - 支持树形结构构建

- **CommentService 批量操作**：
  - `BatchApproveComments`, `BatchRejectComments`, `BatchDeleteComments`
  - 用于后台审核场景

### 2.3 辅助域服务

| 服务 | 核心职责 | 关键方法 |
|-----|---------|---------|
| `MediaService` | 文件上传、媒体管理 | `Upload`, `Delete`（权限校验） |
| `NotificationService` | 评论/回复通知 | `SendCommentNotification`, `SendReplyNotification` |
| `RSSService` | RSS Feed 生成 | `GenerateFeed`, `GenerateFeedForCategory` |
| `SearchService` | 全文搜索 | `Search`, `SearchByCategory`, `SearchByAuthor`, `GetSuggestions` |

**业务逻辑亮点**：

- **MediaService.Upload**：
  - 文件大小限制（10MB）
  - MIME 类型白名单校验
  - 生成唯一文件名（UUID）
  - 自动创建上传目录
  - 图片尺寸提取（预留）
  - 上传失败时清理物理文件

- **RSSService.GenerateFeed**：
  - 批量查询分类避免 N+1 问题
  - 生成 RSS 2.0 标准 XML
  - 支持分类过滤

- **SearchService.Search**：
  - 关键词长度验证（1-100字符）
  - 分页参数校验与默认值
  - PageSize 上限 100

---

## 3. 服务间依赖与协作

### 3.1 依赖关系图

```
Container (聚合根)
    │
    ├── UserService ←─── UserRepository + RoleRepository
    │
    ├── AuthService ←─── UserRepository + JWTService + TokenBlacklist(可选)
    │       │
    │       └── 登录时调用 UserRepository.GetByEmail
    │       └── 验证时调用 UserRepository.GetByID
    │
    ├── PostService ←─── PostRepository + CategoryRepository + TagRepository + SeriesRepository + PostLikeRepository(可选)
    │       │
    │       ├── Create 时调用 TagRepository.AddPostTag
    │       ├── Update 时调用 TagRepository (先删后加)
    │       └── LikePost 时调用 PostLikeRepository.CreateIfNotExists
    │
    ├── CommentService ←─── CommentRepository + CommentLikeRepository
    │       │
    │       ├── GetCommentsByPost 构建树形结构
    │       └── GetLikesBatch 批量查询点赞状态
    │
    ├── MediaService ←─── MediaRepository + 配置(uploadDir, baseURL)
    │
    ├── NotificationService ←─── NotificationChannel (接口)
    │       │
    │       └── 通过 EmailChannel 实现 SMTP 发送
    │
    ├── RSSService ←─── PostRepository + CategoryRepository
    │       │
    │       └── 批量查询分类避免 N+1
    │
    └── SearchService ←─── SearchRepository
```

### 3.2 服务协作示例

**评论通知流程**（CommentService → NotificationService）：

```
Handler.Comment.Create
    │
    ├── CommentService.CreateComment (创建评论)
    │
    └── NotificationService.SendCommentNotification (通知文章作者)
            │
            ├── 检查是否自己评论自己的文章（跳过）
            └── 构建 Notification → EmailChannel.Send
```

---

## 4. 错误处理策略

### 4.1 错误类型定义

服务层错误主要来自 `core` 包的领域错误定义：

| 模块 | 主要错误 | 定义位置 |
|-----|---------|---------|
| user | `ErrUserAlreadyExists`, `ErrInvalidStatus` | `internal/core/user` |
| post | `ErrPostNotFound`, `ErrInvalidStatus`, `ErrAlreadyLiked`, `ErrNotLiked` | `internal/core/post` |
| comment | `ErrCommentNotFound`, `ErrEmptyContent`, `ErrParentNotFound`, `ErrMaxDepthExceeded`, `ErrPermissionDenied`, `ErrAlreadyLiked` | `internal/core/comment` |
| media | `ErrFileSizeTooLarge`, `ErrInvalidMimeType`, `ErrPermissionDenied` | `internal/core/media` |
| search | `ErrEmptyQuery`, `ErrQueryTooLong` | `internal/core/search` |

### 4.2 错误处理模式

**1. 参数验证错误（早期返回）**：
```go
if username == "" || email == "" || password == "" {
    return nil, errors.New("username, email and password are required")
}
```

**2. 业务规则错误（领域错误）**：
```go
if c.Status == comment.StatusDeleted {
    return errors.New("无法批准已删除的评论")
}
```

**3. 权限校验错误**：
```go
if c.UserID != userID {
    return comment.ErrPermissionDenied
}
```

**4. 依赖操作错误（传递）**：
```go
if err := s.userRepo.Create(ctx, newUser); err != nil {
    return nil, err  // 直接传递 Repository 错误
}
```

**5. 可选依赖静默处理**：
```go
if s.blacklist == nil {
    return nil  // 无黑名单时直接返回
}

if s.channel == nil {
    return nil  // 无通知渠道时静默跳过
}
```

### 4.3 错误包装（部分使用）：

```go
return nil, fmt.Errorf("failed to open uploaded file: %w", err)
```

---

## 5. 服务层与 Repository 层关系

### 5.1 分层职责划分

| 层级 | 职责 | 示例 |
|-----|-----|-----|
| **Handler** | HTTP 解析/响应、权限检查、调用 Service | 解析 JSON → 调用 `PostService.Create` |
| **Service** | 业务逻辑、验证、跨 Repository 协调 | 验证标题、默认状态、关联标签 |
| **Repository** | 数据持久化、单一表/实体操作 | `INSERT INTO posts`, `SELECT * FROM tags` |

### 5.2 Service 对 Repository 的调用模式

**1. 单 Repository 调用**：
```go
func (s *userServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
    return s.userRepo.GetByID(ctx, id)  // 直接委托
}
```

**2. 多 Repository 协调**：
```go
func (s *postServiceImpl) Create(ctx context.Context, p *post.Post, tagIDs []uuid.UUID) error {
    // 1. 创建文章
    if err := s.postRepo.Create(ctx, p); err != nil {
        return err
    }
    // 2. 关联标签（另一个 Repository）
    for _, tagID := range tagIDs {
        if err := s.tagRepo.AddPostTag(ctx, p.ID, tagID); err != nil {
            return err
        }
    }
    return nil
}
```

**3. 跨 Repository 查询优化**：
```go
// RSSService 批量查询分类避免 N+1
categoriesMap := make(map[uuid.UUID]*post.Category)
dbCategoryRepo, ok := s.categoryRepo.(*database.CategoryRepository)
if ok && len(categoryIDs) > 0 {
    categoriesMap, _ = dbCategoryRepo.GetByIDs(ctx, categoryIDs)  // 批量查询
}
```

### 5.3 Repository 接口定义位置

接口定义在 `internal/core/{domain}/repository.go`，实现在 `internal/database/{entity}_repository.go`：

```
internal/core/post/repository.go      ← 接口定义
internal/database/post_repository.go  ← GORM 实现
```

这种分离使得：
- Service 层只依赖接口（易于测试 Mock）
- 实现层可替换（GORM → SQLx 等）

---

## 6. 接口设计模式

### 6.1 接口命名规范

所有服务接口以 `Service` 结尾，实现结构体以 `ServiceImpl` 结尾：

```go
type UserService interface { ... }
type userServiceImpl struct { ... }  // 小写，不导出
```

### 6.2 工厂函数命名

```go
func NewUserService(...) UserService           // 基础版本
func NewAuthServiceWithBlacklist(...) AuthService  // 增强版本
func NewPostServiceWithLikes(...) PostService     // 可选功能版本
```

### 6.3 可选依赖注入

通过 Builder 模式或可选参数：

```go
// 方式1：带可选参数的工厂函数
func NewAuthServiceWithBlacklist(userRepo, jwtService, blacklist) AuthService

// 方式2：With 方法（链式配置）
func (s *authServiceImpl) WithBlacklist(blacklist TokenBlacklist) AuthService
```

---

## 7. 代码组织总结

### 7.1 服务层文件结构

```
internal/services/
├── services.go           ← Container + 工厂方法
├── auth_service.go       ← AuthService
├── user_service.go       ← UserService
├── post_service.go       ← PostService + CategoryService + TagService + SeriesService
├── comment_service.go    ← CommentService + CommentNode
├── media_service.go      ← MediaService
├── notification_service.go ← NotificationService
├── rss_service.go        ← RSSService
├── search_service.go     ← SearchService
├── email_channel.go      ← EmailChannel（NotificationChannel 实现）
```

### 7.2 设计特点总结

| 特点 | 说明 |
|-----|-----|
| **依赖注入** | 通过 Container 集中管理，工厂方法渐进组装 |
| **接口隔离** | 所有服务定义为接口，实现不导出 |
| **单一职责** | 每个服务聚焦一个领域，跨 Repo 协调 |
| **渐进式组装** | 多个工厂方法支持最小依赖容器 |
| **可选依赖** | 黑名单、点赞、通知等可选注入 |
| **错误定义集中** | 领域错误在 core 包定义，Service 传递/包装 |
| **批量优化** | RSS 服务避免 N+1，Comment 服务批量点赞查询 |

---

## 8. 改进建议

### 8.1 现存问题

1. **批量操作性能**：`BatchApproveComments` 循环调用单条操作，应改为批量 SQL
2. **错误包装不一致**：部分使用 `fmt.Errorf("...: %w", err)`，部分直接返回
3. **GetCommentsByStatus 计数**：`total` 计算方式不准确（仅计算当前页）

### 8.2 建议改进

1. **统一错误包装**：使用 `errors.Is` 和 `errors.As` 进行错误判断
2. **批量操作优化**：Repository 层添加 `BatchUpdateStatus` 方法
3. **事务支持**：跨 Repository 操作（如 PostService.Create）应支持事务传递

---

*分析完成时间：2026-04-01*