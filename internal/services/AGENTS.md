<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# services

## Purpose
业务服务层，封装业务逻辑，协调 Repository 调用。

## Key Files
| File | Description |
|------|-------------|
| `services.go` | Service Container 定义和构造函数 |
| `user_service.go` | 用户服务：创建、查询、更新 |
| `auth_service.go` | 认证服务：登录、注册、密码验证 |
| `post_service.go` | 文章服务：CRUD、发布、版本管理 |
| `comment_service.go` | 评论服务：CRUD、审核 |
| `media_service.go` | 媒体服务：上传、删除、URL 生成 |
| `search_service.go` | 搜索服务：全文搜索 |
| `rss_service.go` | RSS 服务：Feed 生成 |
| `notification_service.go` | 通知服务：发送通知 |
| `email_channel.go` | 邮件通知渠道实现 |

## For AI Agents

### Working In This Directory
- Service 层是业务逻辑核心
- Handler 调用 Service，Service 调用 Repository
- 使用 Container 管理服务依赖

### Service Container
```go
type Container struct {
    UserService        UserService
    AuthService        AuthService
    PostService        PostService
    CommentService     CommentService
    MediaService       MediaService
    SearchService      SearchService
    RSSService         RSSService
    NotificationService NotificationService
    jwtService         *auth.JWTService
}
```

### Container Constructors
| Function | Services Included |
|----------|-------------------|
| `NewContainer` | User, Auth |
| `NewContainerWithBlacklist` | User, Auth + Token Blacklist |
| `NewContainerWithPosts` | + Post, Category, Tag, Series |
| `NewContainerWithComments` | + Comment |
| `NewContainerWithMedia` | + Media, RSS, Search |
| `NewContainerWithNotifications` | + Notification |

### Service Pattern
```go
type PostService struct {
    postRepo     post.PostRepository
    categoryRepo post.CategoryRepository
    tagRepo      post.TagRepository
    seriesRepo   post.SeriesRepository
}

func NewPostService(...) *PostService { ... }

func (s *PostService) Create(ctx context.Context, req CreatePostInput) (*Post, error) {
    // 1. 业务验证
    // 2. 数据处理
    // 3. 调用 Repository
    // 4. 返回结果
}
```

<!-- MANUAL: -->