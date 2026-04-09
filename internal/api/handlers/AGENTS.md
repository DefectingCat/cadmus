<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# handlers

## Purpose

`handlers` 目录包含所有 HTTP 请求处理器实现。每个处理器负责处理特定资源领域的 HTTP 请求，包括请求解析、参数验证、调用服务层和格式化响应。

## Key Files

| File | Purpose |
|------|---------|
| `admin.go` | 后台管理 API：角色管理、用户管理、批量操作、排序更新 |
| `auth.go` | 用户认证 API：注册、登录、登出、令牌刷新、当前用户信息 |
| `category.go` | 分类管理 API：分类 CRUD、层级分类、文章数量统计 |
| `comment.go` | 评论管理 API：评论 CRUD、树形结构、点赞、审核、通知 |
| `media.go` | 媒体文件 API：文件上传、列表查询、删除管理 |
| `middleware.go` | 中间件工具：JWT 认证、权限检查、错误响应、Context 工具 |
| `post.go` | 文章管理 API：文章 CRUD、发布管理、版本控制、点赞 |
| `rss.go` | RSS 订阅 API：RSS Feed 生成、分类筛选 |
| `search.go` | 搜索 API：全文搜索、搜索建议 |
| `tag.go` | 标签管理 API：标签 CRUD、文章数量统计 |

## Subdirectories

无子目录

## Architecture Guide for AI Agents

### Handler 构造模式

所有 Handler 遵循统一的构造模式：

```go
// 1. 定义 Handler 结构体，持有 Service 引用
type PostHandler struct {
    postService services.PostService
}

// 2. 构造函数注入 Service
func NewPostHandler(postService services.PostService) *PostHandler {
    return &PostHandler{postService: postService}
}

// 3. 可选：支持额外依赖的构造函数
func NewPostHandlerWithNotifications(
    postService services.PostService,
    notificationService services.NotificationService,
) *PostHandler {
    return &PostHandler{
        postService:        postService,
        notificationService: notificationService,
    }
}
```

### 请求处理流程

```go
func (h *Handler) Method(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 路径参数解析（Go 1.22+）
    idStr := r.PathValue("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        WriteAPIError(w, "BAD_REQUEST", "无效的 ID", nil, http.StatusBadRequest)
        return
    }
    
    // 2. 查询参数解析
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    
    // 3. 请求体解析
    var req RequestStruct
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
        return
    }
    
    // 4. 获取认证用户 ID
    userID, err := GetUserID(ctx)
    if err != nil {
        WriteAPIError(w, "UNAUTHORIZED", "未登录", nil, http.StatusUnauthorized)
        return
    }
    
    // 5. 调用 Service 层
    result, err := h.service.Method(ctx, params)
    
    // 6. 统一响应格式
    WriteJSON(w, result, http.StatusOK)
}
```

### 路由映射

| Handler | HTTP Method | Route | Description |
|---------|-------------|-------|-------------|
| `AuthHandler` | POST | `/api/v1/auth/register` | 用户注册 |
| `AuthHandler` | POST | `/api/v1/auth/login` | 用户登录 |
| `AuthHandler` | POST | `/api/v1/auth/logout` | 用户登出 |
| `AuthHandler` | POST | `/api/v1/auth/refresh` | 刷新令牌 |
| `AuthHandler` | GET | `/api/v1/auth/me` | 当前用户信息 |
| `PostHandler` | GET | `/api/v1/posts` | 文章列表 |
| `PostHandler` | GET | `/api/v1/posts/{slug}` | 文章详情 |
| `PostHandler` | POST | `/api/v1/posts` | 创建文章 |
| `PostHandler` | PUT | `/api/v1/posts/{id}` | 更新文章 |
| `PostHandler` | DELETE | `/api/v1/posts/{id}` | 删除文章 |
| `PostHandler` | POST | `/api/v1/posts/{id}/publish` | 发布文章 |
| `PostHandler` | POST | `/api/v1/posts/{id}/like` | 点赞文章 |
| `PostHandler` | DELETE | `/api/v1/posts/{id}/like` | 取消点赞 |
| `CommentHandler` | GET | `/api/v1/comments/post/{postId}` | 文章评论树 |
| `CommentHandler` | POST | `/api/v1/comments` | 发表评论 |
| `CommentHandler` | PUT | `/api/v1/comments/{id}` | 编辑评论 |
| `CommentHandler` | DELETE | `/api/v1/comments/{id}` | 删除评论 |
| `CommentHandler` | POST | `/api/v1/comments/{id}/like` | 点赞评论 |
| `CommentHandler` | PUT | `/api/v1/comments/{id}/approve` | 批准评论 |
| `CategoryHandler` | GET | `/api/v1/categories` | 分类列表 |
| `CategoryHandler` | GET | `/api/v1/categories/{slug}` | 分类详情 |
| `CategoryHandler` | POST | `/api/v1/categories` | 创建分类 |
| `CategoryHandler` | PUT | `/api/v1/categories/{id}` | 更新分类 |
| `CategoryHandler` | DELETE | `/api/v1/categories/{id}` | 删除分类 |
| `TagHandler` | GET | `/api/v1/tags` | 标签列表 |
| `TagHandler` | GET | `/api/v1/tags/{slug}` | 标签详情 |
| `TagHandler` | POST | `/api/v1/tags` | 创建标签 |
| `TagHandler` | DELETE | `/api/v1/tags/{id}` | 删除标签 |
| `MediaHandler` | POST | `/api/v1/media/upload` | 上传文件 |
| `MediaHandler` | GET | `/api/v1/media` | 媒体列表 |
| `MediaHandler` | GET | `/api/v1/media/{id}` | 媒体详情 |
| `MediaHandler` | DELETE | `/api/v1/media/{id}` | 删除媒体 |
| `SearchHandler` | GET | `/api/v1/search` | 全文搜索 |
| `SearchHandler` | GET | `/api/v1/search/suggestions` | 搜索建议 |
| `RSSHandler` | GET | `/api/v1/rss` | RSS Feed |
| `AdminHandler` | GET | `/api/v1/admin/roles` | 角色列表 |
| `AdminHandler` | POST | `/api/v1/admin/roles` | 创建角色 |
| `AdminHandler` | PUT | `/api/v1/admin/roles/{id}` | 更新角色 |
| `AdminHandler` | DELETE | `/api/v1/admin/roles/{id}` | 删除角色 |
| `AdminHandler` | GET | `/api/v1/admin/users` | 用户列表 |
| `AdminHandler` | PUT | `/api/v1/admin/users/{id}/ban` | 封禁用户 |
| `AdminHandler` | POST | `/api/v1/admin/batch` | 批量操作 |

### 错误码规范

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `INVALID_REQUEST` | 400 | 请求格式错误 |
| `VALIDATION_ERROR` | 400 | 必填字段缺失或验证失败 |
| `UNAUTHORIZED` | 401 | 未登录 |
| `AUTH_FAILED` | 401 | 认证失败 |
| `TOKEN_REVOKED` | 401 | 令牌已被撤销 |
| `FORBIDDEN` | 403 | 权限不足 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `ALREADY_EXISTS` | 409 | 资源已存在 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |

### 工具函数

| Function | Purpose |
|----------|---------|
| `WriteJSON(w, data, status)` | 写入 JSON 响应 |
| `WriteAPIError(w, code, message, details, status)` | 写入 API 错误响应 |
| `GetUserID(ctx)` | 从上下文获取用户 ID |
| `GetUserRoleID(ctx)` | 从上下文获取用户角色 ID |
| `ExtractToken(r)` | 从请求中提取 JWT Token |
| `GetRequestID()` | 获取请求 ID（用于追踪） |

### 中间件链

```
Request → AuthMiddleware → PermissionMiddleware → Handler
```

- **AuthMiddleware**: JWT Token 验证，注入用户 ID 到 context
- **PermissionMiddleware**: 基于角色的权限检查
- **AdminMiddleware**: 管理员权限检查（检查 `admin:access` 权限）

## Dependencies

### 内部依赖

```
handlers/ → services/ → core/
handlers/ → database/ (部分 Handler 直接使用 Repository)
```

### 外部依赖

| Package | Purpose |
|---------|---------|
| `net/http` | HTTP 服务器 |
| `encoding/json` | JSON 编解码 |
| `github.com/google/uuid` | UUID 处理 |

## For AI Agents

### 开发指南

1. **新增 API 端点**
   - 在对应的 Handler 文件中添加方法
   - 使用统一的错误处理和响应格式
   - 路径参数使用 `r.PathValue("param")`
   - 查询参数使用 `r.URL.Query().Get("key")`

2. **定义请求/响应结构体**
   ```go
   type CreateRequest struct {
       Name string `json:"name"`
       Slug string `json:"slug"`
   }
   
   type Response struct {
       ID        uuid.UUID `json:"id"`
       Name      string    `json:"name"`
       CreatedAt string    `json:"created_at"`
   }
   ```

3. **错误处理**
   - 参数验证错误返回 `VALIDATION_ERROR` 或 `BAD_REQUEST`
   - 资源不存在返回 `NOT_FOUND`
   - 权限不足返回 `PERMISSION_DENIED`
   - 服务层错误返回 `INTERNAL_ERROR`

4. **分页处理**
   - 默认页码为 1，默认每页 20 条
   - 最大每页数量限制为 100
   - 响应包含 `page`、`page_size`、`total` 字段

### 代码约定

- Handler 方法名：大写动词 `List`, `Get`, `Create`, `Update`, `Delete`
- 请求结构体：`<Action><Resource>Request` 如 `CreatePostRequest`
- 响应结构体：`<Resource>Response` 如 `PostResponse`
- 时间格式：RFC3339 `2006-01-02T15:04:05Z07:00`

### 测试策略

- Handler 层使用 `httptest` 进行集成测试
- Mock Service 依赖避免真实数据库调用
- 验证 HTTP 状态码和响应 JSON 结构

<!-- MANUAL: -->
