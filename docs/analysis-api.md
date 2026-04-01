# Cadmus API 层设计分析报告

## 1. API 路由组织 (`cmd/server/routes.go`)

### 1.1 路由架构设计

Cadmus 采用 **模块化路由组织** 模式，通过 `RouteDeps` 结构体聚合所有依赖，实现依赖注入：

```go
type RouteDeps struct {
    // Handlers (9 个)
    AuthHandler    *handlers.AuthHandler
    PostHandler    *handlers.PostHandler
    CategoryHandler *handlers.CategoryHandler
    TagHandler     *handlers.TagHandler
    CommentHandler *handlers.CommentHandler
    MediaHandler   *handlers.MediaHandler
    RSSHandler     *handlers.RSSHandler
    SearchHandler  *handlers.SearchHandler
    AdminHandler   *handlers.AdminHandler

    // Middleware (3 个限流器)
    LoginLimiter  *middleware.RateLimiter
    PublicLimiter *middleware.RateLimiter
    UserLimiter   *middleware.RateLimiter

    // Auth 相关
    JWTService    *auth.JWTService
    TokenBlacklist auth.TokenBlacklist
    PermCache     *auth.PermissionCache

    // Repositories
    PostRepo    *database.PostRepository
    CommentRepo *database.CommentRepository

    // Services
    Services *services.Container

    // Config
    UploadDir string
}
```

### 1.2 路由分组策略

路由按功能域分组，每个域有独立的 `setupXXXRoutes` 函数：

| 分组 | 路径前缀 | 认证要求 | 限流策略 |
|------|----------|----------|----------|
| Auth | `/api/v1/auth` | 部分需要 | LoginLimiter |
| Post | `/api/v1/posts` | 写操作需要 | PublicLimiter/UserLimiter |
| Category | `/api/v1/categories` | 写操作需要 | 无 |
| Tag | `/api/v1/tags` | 写操作需要 | 无 |
| Comment | `/api/v1/comments` | 需要 | 无 |
| Media | `/api/v1/media` | 需要 | 无 |
| RSS | `/api/v1/rss` | 公开 | 无 |
| Search | `/api/v1/search` | 公开 | PublicLimiter |
| Admin | `/api/v1/admin` | 管理员权限 | 无 |
| Static | `/static`, `/uploads` | 公开 | 无 |
| Health | `/health` | 公开 | 无 |

### 1.3 路由注册模式

使用 Go 1.22+ 的 `http.ServeMux` 新语法：

```go
// 公开路由
mux.HandleFunc("GET /api/v1/posts", publicRL(http.HandlerFunc(deps.PostHandler.List)).ServeHTTP)

// 需认证路由（中间件链式组合）
mux.HandleFunc("POST /api/v1/posts", userRL(authMW(http.HandlerFunc(deps.PostHandler.Create))).ServeHTTP)

// 管理员路由（双重中间件）
mux.Handle("GET /api/v1/admin/roles", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.ListRoles))))
```

**中间件组合顺序**：限流 → 认证 → 权限 → Handler

---

## 2. Handler 实现模式

### 2.1 Handler 结构设计

所有 Handler 遵循统一的构造器模式：

```go
type XXXHandler struct {
    service services.XXXService  // 依赖 Service 层
}

func NewXXXHandler(service services.XXXService) *XXXHandler {
    return &XXXHandler{service: service}
}
```

**Handler 列表**：

| Handler | 文件 | 行数 | 主要职责 |
|---------|------|------|----------|
| AuthHandler | auth.go | ~207 | 注册、登录、登出、刷新令牌、获取当前用户 |
| PostHandler | post.go | ~573 | 文章 CRUD、发布、版本管理、点赞 |
| CommentHandler | comment.go | ~690 | 评论 CRUD、审核、批量操作、点赞 |
| CategoryHandler | category.go | ~236 | 分类 CRUD |
| TagHandler | tag.go | ~159 | 标签 CRUD |
| MediaHandler | media.go | ~218 | 文件上传、管理 |
| SearchHandler | search.go | ~100 | 全文搜索、搜索建议 |
| RSSHandler | rss.go | ~44 | RSS Feed 生成 |
| AdminHandler | admin.go | ~711 | 角色管理、用户管理、批量操作、排序更新 |

### 2.2 请求处理模式

#### 请求解析

```go
// JSON 请求体解析
var req CreatePostRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
    return
}
```

#### 路径参数提取

```go
// Go 1.22+ PathValue
slug := r.PathValue("slug")
idStr := r.PathValue("id")

// UUID 解析
id, err := uuid.Parse(idStr)
if err != nil {
    WriteAPIError(w, "VALIDATION_ERROR", "无效的ID", nil, http.StatusBadRequest)
    return
}
```

#### 分页参数处理

```go
page, _ := strconv.Atoi(r.URL.Query().Get("page"))
if page < 1 {
    page = 1
}
pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
if pageSize < 1 || pageSize > 100 {
    pageSize = 20  // 默认值 + 上限保护
}
```

### 2.3 响应格式设计

#### 成功响应

```go
// 单一资源
WriteJSON(w, toPostResponse(p), http.StatusCreated)

// 列表响应（带分页信息）
type PostListResponse struct {
    Posts     []PostResponse `json:"posts"`
    Total     int            `json:"total"`
    Page      int            `json:"page"`
    PageSize  int            `json:"page_size"`
}

// 操作确认
WriteJSON(w, map[string]string{"message": "文章已发布"}, http.StatusOK)

// 无内容响应（删除）
w.WriteHeader(http.StatusNoContent)
```

#### 错误响应（统一格式）

```go
type APIError struct {
    Code      string   `json:"code"`       // 错误代码：AUTH_FAILED, VALIDATION_ERROR
    Message   string   `json:"message"`    // 用户友好消息
    Details   []string `json:"details"`    // 详细错误列表
    RequestID string   `json:"request_id"` // 请求追踪 ID
}

func WriteAPIError(w http.ResponseWriter, code string, message string, details []string, status int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(APIError{
        Code:      code,
        Message:   message,
        Details:   details,
        RequestID: GetRequestID(),
    })
}
```

**错误代码分类**：

| 类别 | 错误代码 | HTTP 状态码 |
|------|----------|-------------|
| 认证 | AUTH_FAILED, UNAUTHORIZED, TOKEN_REVOKED | 401 |
| 权限 | PERMISSION_DENIED | 403 |
| 资源 | NOT_FOUND, USER_NOT_FOUND, POST_NOT_FOUND | 404 |
| 冲突 | USER_EXISTS, ALREADY_LIKED | 409 |
| 验证 | VALIDATION_ERROR, INVALID_REQUEST, BAD_REQUEST | 400 |
| 业务 | USER_BANNED, FILE_TOO_LARGE, RATE_LIMIT_EXCEEDED | 400/403 |
| 系统 | INTERNAL_ERROR | 500 |

### 2.4 上下文信息传递

```go
// 从 context 获取用户 ID
userID, err := GetUserID(ctx)
if err != nil {
    WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
    return
}

// 类型安全的 context key
type ctxKey string
const (
    ctxUserID   ctxKey = "user_id"
    ctxUserRole ctxKey = "user_role_id"
)
```

### 2.5 N+1 查询优化

项目在多处实现了批量查询优化：

```go
// CommentHandler - 批量查询点赞状态
commentIDs := collectCommentIDs(nodes)
likesMap := make(map[uuid.UUID]bool)
if userID != uuid.Nil && len(commentIDs) > 0 {
    likesMap, _ = h.commentService.GetLikesBatch(ctx, commentIDs, userID)
}

// AdminHandler - 批量查询角色和文章数
roleIDs := make([]uuid.UUID, 0, len(users))
userIDs := make([]uuid.UUID, 0, len(users))
rolesMap, _ = dbRoleRepo.GetByIDs(ctx, roleIDs)
postCountsMap, _ = dbPostRepo.CountByAuthors(ctx, userIDs)
```

---

## 3. 中间件设计 (`internal/api/handlers/middleware.go`, `internal/api/middleware/ratelimit.go`)

### 3.1 认证中间件

#### AuthMiddleware（基础认证）

```go
func AuthMiddleware(jwtService *auth.JWTService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := ExtractToken(r)
            claims, err := jwtService.Validate(token)
            // 设置 context
            ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
            ctx = context.WithValue(ctx, ctxUserRole, claims.RoleID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

#### AuthMiddlewareWithBlacklist（带黑名单检查）

支持 JWT 撤销机制：

```go
// 检查 jti 是否在黑名单中
jti := claims.GetJWTID()
if jti != "" && blacklist.IsBlacklisted(ctx, jti) {
    WriteAPIError(w, "TOKEN_REVOKED", "令牌已被撤销", nil, http.StatusUnauthorized)
    return
}
```

#### Token 提取策略

支持多种来源：

```go
func ExtractToken(r *http.Request) string {
    // 1. Authorization header (Bearer token)
    authHeader := r.Header.Get("Authorization")
    if authHeader != "" {
        token, found := strings.CutPrefix(authHeader, "Bearer ")
        if found {
            return token
        }
    }
    
    // 2. Cookie (auth_token)
    cookie, err := r.Cookie("auth_token")
    if err == nil && cookie.Value != "" {
        return cookie.Value
    }
    
    return ""
}
```

### 3.2 权限中间件

#### PermissionMiddleware（直接查库）

```go
func PermissionMiddleware(jwtService *auth.JWTService, permRepo user.PermissionRepository, requiredPerm string) func(http.Handler) http.Handler
```

#### CachedPermissionMiddleware（带缓存）

使用 `PermissionCache` 缓存权限查询结果：

```go
func CachedPermissionMiddleware(jwtService *auth.JWTService, permCache *auth.PermissionCache, requiredPerm string) func(http.Handler) http.Handler
```

#### AdminMiddleware（管理员专用）

检查 `admin:access` 权限：

```go
func AdminMiddleware(permCache *auth.PermissionCache) func(http.Handler) http.Handler {
    hasPerm, err := permCache.GetPermission(ctx, roleID, "admin:access")
}
```

#### RequirePermissionMiddleware（通用权限检查）

```go
func RequirePermissionMiddleware(permCache *auth.PermissionCache, permission string) func(http.Handler) http.Handler
```

### 3.3 限流中间件 (`middleware/ratelimit.go`)

#### 滑动窗口算法实现

基于 Redis ZSET 的滑动窗口限流：

```go
type RateLimiter struct {
    client *redis.Client
    limit  int           // 窗口内最大请求数
    window time.Duration // 窗口时长
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) bool {
    // 1. 移除窗口外的旧记录
    pipe.ZRemRangeByScore(ctx, key, "0", windowStart.UnixNano())
    
    // 2. 获取当前窗口内的记录数
    countCmd := pipe.ZCard(ctx, key)
    
    // 3. 检查是否超过限制
    if count >= limit {
        return false
    }
    
    // 4. 添加当前请求记录
    pipe2.ZAdd(ctx, key, redis.Z{Score: nowNano, Member: member})
    pipe2.Expire(ctx, key, window + time.Minute)
}
```

#### Key 生成策略

```go
// 基于 IP
func IPKeyFunc(prefix string) func(*http.Request) string {
    // 支持 X-Real-IP, X-Forwarded-For, RemoteAddr
}

// 基于用户 ID（未认证时回退到 IP）
func UserKeyFunc(prefix string, getUserID func(*http.Request) string) func(*http.Request) string
```

#### 预定义限流策略

```go
var (
    LoginLimit     = 10        // 登录：10 次/分钟
    LoginWindow    = time.Minute
    
    PublicAPILimit = 60        // 公开 API：60 次/分钟
    PublicAPIWindow = time.Minute
    
    UserActionLimit = 100      // 用户操作：100 次/分钟
    UserActionWindow = time.Minute
)
```

#### 限流响应

```go
w.Header().Set("X-RateLimit-Limit", limit)
w.Header().Set("X-RateLimit-Remaining", remaining)
w.Header().Set("X-RateLimit-Reset", resetTime)
w.WriteHeader(http.StatusTooManyRequests)
w.Write([]byte(`{"code":"RATE_LIMIT_EXCEEDED","message":"请求过于频繁，请稍后再试"}`))
```

---

## 4. API 版本控制策略

### 4.1 当前策略

项目采用 **URL 路径版本控制**：

- 所有 API 路径以 `/api/v1/` 前缀
- 当前仅有一个版本（v1）
- 未实现版本协商或向后兼容机制

### 4.2 版本控制特点

| 特点 | 当前状态 | 说明 |
|------|----------|------|
| 版本前缀 | `/api/v1/` | 固定前缀 |
| 多版本支持 | 未实现 | 无 v2 路径 |
| 版本协商 | 未实现 | 无 Accept header 版本选择 |
| 向后兼容 | 未实现 | 无字段兼容性处理 |
| 版本废弃 | 未实现 | 无废弃通知机制 |

---

## 5. 请求/响应处理方式

### 5.1 请求处理流程

```
Client Request → RateLimiter → AuthMiddleware → PermissionMiddleware → Handler → Service → Repository → Database
                              ↓
                         Context 注入 (userID, roleID)
```

### 5.2 响应处理流程

```
Handler → WriteJSON/WriteAPIError → HTTP Response
         ↓
    设置 Content-Type: application/json
    设置 CORS headers（未在代码中体现，可能在外层）
    设置 RateLimit headers
```

### 5.3 特殊响应处理

#### RSS Feed（XML 响应）

```go
w.Header().Set("Content-Type", "application/xml; charset=utf-8")
w.Header().Set("Cache-Control", "public, max-age=3600") // 1 小时缓存
w.Write([]byte(xmlContent))
```

#### Admin Pages（Templ 模板渲染）

```go
templ.Handler(adminpages.DashboardPage(stats, recentPosts, recentComments, pendingCount)).ServeHTTP(w, r)
```

### 5.4 异步处理模式

```go
// 异步增加浏览量
go h.postService.IncrementViewCount(ctx, p.ID)

// 异步发送通知
go func() {
    bgCtx := context.Background()
    h.notificationService.SendReplyNotification(bgCtx, ...)
}()
```

---

## 6. 设计模式总结

### 6.1 优点

1. **模块化路由组织**：按功能域分组，易于维护和扩展
2. **依赖注入**：通过 `RouteDeps` 聚合依赖，避免全局变量
3. **统一错误格式**：`APIError` 结构提供一致的错误响应
4. **类型安全 context**：使用自定义类型 `ctxKey` 避免冲突
5. **N+1 查询优化**：批量查询点赞状态、角色信息、文章数
6. **限流策略分级**：登录、公开、用户操作不同限流阈值
7. **滑动窗口限流**：Redis ZSET 实现，比固定窗口更平滑
8. **权限缓存**：`CachedPermissionMiddleware` 减少数据库查询
9. **Token 多源提取**：支持 Header 和 Cookie 两种方式
10. **fail-open 策略**：Redis 错误时允许请求通过

### 6.2 可改进点

1. **缺少 CORS 中间件**：跨域处理未在 API 层体现
2. **缺少请求日志中间件**：无请求追踪和性能监控
3. **缺少请求体大小限制**：除文件上传外无全局限制
4. **分页参数解析重复**：每个 Handler 重复解析逻辑
5. **版本控制机制不完善**：仅有 v1，无版本演进策略
6. **RequestID 生成简单**：目前使用 UUID，未与日志系统集成
7. **缺少 API 文档生成**：无 Swagger/OpenAPI 集成
8. **缺少响应压缩**：无 gzip 中间件
9. **部分 Handler 无限流**：Category、Tag、Comment、Media 无限流保护

### 6.3 建议改进

```go
// 1. 添加通用分页解析
type PaginationParams struct {
    Page     int
    PageSize int
    Offset   int
}
func ParsePagination(r *http.Request) PaginationParams { ... }

// 2. 添加 CORS 中间件
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler { ... }

// 3. 添加请求日志中间件
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler { ... }

// 4. 添加全局限流保护
func GlobalRateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler { ... }
```

---

## 7. 路由完整性清单

| 路由 | 方法 | 认证 | 权限 | 限流 |
|------|------|------|------|------|
| `/api/v1/auth/register` | POST | ❌ | ❌ | Login |
| `/api/v1/auth/login` | POST | ❌ | ❌ | Login |
| `/api/v1/auth/logout` | POST | ✅ | ❌ | ❌ |
| `/api/v1/auth/refresh` | POST | ❌ | ❌ | ❌ |
| `/api/v1/auth/me` | GET | ✅ | ❌ | ❌ |
| `/api/v1/posts` | GET | ❌ | ❌ | Public |
| `/api/v1/posts` | POST | ✅ | ❌ | User |
| `/api/v1/posts/{slug}` | GET | ❌ | ❌ | Public |
| `/api/v1/posts/{id}` | PUT | ✅ | ❌ | User |
| `/api/v1/posts/{id}` | DELETE | ✅ | ❌ | User |
| `/api/v1/posts/{id}/publish` | POST | ✅ | ❌ | User |
| `/api/v1/posts/{id}/rollback` | POST | ✅ | ❌ | User |
| `/api/v1/posts/{id}/like` | POST/DELETE | ✅ | ❌ | User |
| `/api/v1/posts/{id}/versions` | GET | ❌ | ❌ | Public |
| `/api/v1/users/{id}/posts` | GET | ❌ | ❌ | Public |
| `/api/v1/categories` | GET | ❌ | ❌ | ❌ |
| `/api/v1/categories` | POST | ✅ | ❌ | ❌ |
| `/api/v1/categories/{slug}` | GET | ❌ | ❌ | ❌ |
| `/api/v1/categories/{id}` | PUT/DELETE | ✅ | ❌ | ❌ |
| `/api/v1/tags` | GET | ❌ | ❌ | ❌ |
| `/api/v1/tags` | POST | ✅ | ❌ | ❌ |
| `/api/v1/tags/{slug}` | GET | ❌ | ❌ | ❌ |
| `/api/v1/tags/{id}` | DELETE | ✅ | ❌ | ❌ |
| `/api/v1/comments/post/{postId}` | GET | ❌ | ❌ | ❌ |
| `/api/v1/comments` | POST | ✅ | ❌ | ❌ |
| `/api/v1/comments/{id}` | PUT/DELETE | ✅ | ❌ | ❌ |
| `/api/v1/comments/{id}/like` | POST/DELETE | ✅ | ❌ | ❌ |
| `/api/v1/comments/{id}/approve` | PUT | ✅ | comment:moderate | ❌ |
| `/api/v1/comments/{id}/reject` | PUT | ✅ | comment:moderate | ❌ |
| `/api/v1/admin/comments` | GET | ✅ | comment:moderate | ❌ |
| `/api/v1/admin/comments/batch-*` | PUT/DELETE | ✅ | comment:moderate | ❌ |
| `/api/v1/media/upload` | POST | ✅ | ❌ | ❌ |
| `/api/v1/media` | GET | ✅ | ❌ | ❌ |
| `/api/v1/media/{id}` | GET/DELETE | ✅ | ❌ | ❌ |
| `/api/v1/rss` | GET | ❌ | ❌ | ❌ |
| `/api/v1/search` | GET | ❌ | ❌ | Public |
| `/api/v1/search/suggestions` | GET | ❌ | ❌ | Public |
| `/api/v1/admin/roles` | GET/POST | ✅ | admin:access | ❌ |
| `/api/v1/admin/roles/{id}` | PUT/DELETE | ✅ | admin:access | ❌ |
| `/api/v1/admin/users` | GET | ✅ | admin:access | ❌ |
| `/api/v1/admin/users/{id}/ban` | PUT | ✅ | admin:access | ❌ |
| `/api/v1/admin/batch` | POST | ✅ | admin:access | ❌ |
| `/api/v1/admin/order` | PUT | ✅ | admin:access | ❌ |
| `/admin` | GET | ✅ | admin:access | ❌ |
| `/admin/comments` | GET | ✅ | admin:access | ❌ |
| `/static/*` | GET | ❌ | ❌ | ❌ |
| `/uploads/*` | GET | ❌ | ❌ | ❌ |
| `/health` | GET | ❌ | ❌ | ❌ |

---

**分析完成时间**：2026-04-01
**分析范围**：API 路由、Handler 实现、中间件设计、版本控制、请求/响应处理