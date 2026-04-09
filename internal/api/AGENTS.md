<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# api

## Purpose

`api` 目录是应用的 HTTP 接口层，负责处理所有 REST API 请求。采用标准的 Handler + Service 分层架构，Handler 仅负责 HTTP 协议相关的请求解析和响应格式化，业务逻辑全部委托给 `services` 层处理。

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `handlers/` | HTTP 处理器实现：文章、评论、用户、媒体、搜索、RSS 等 API (see `handlers/AGENTS.md`) |
| `middleware/` | HTTP 中间件：限流器等横切关注点 (see `middleware/AGENTS.md`) |

## Architecture Guide for AI Agents

### Handler 模式

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

// 3. 实现 HTTP 方法（增删改查）
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
    // 1. 解析请求参数
    // 2. 调用 Service 处理业务
    // 3. 返回 JSON 响应或错误
}
```

### 中间件链

请求处理流程：

```
Request → RateLimitMiddleware → AuthMiddleware → PermissionMiddleware → Handler
```

中间件按顺序处理：
1. **限流中间件** - 基于 IP 或用户 ID 限流
2. **认证中间件** - JWT Token 验证，提取用户 ID
3. **权限中间件** - 基于缓存的权限检查
4. **Handler** - 业务逻辑处理

### 请求处理流程

```go
// 典型 Handler 方法结构
func (h *Handler) Method(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 路径参数解析
    idStr := r.PathValue("id")
    id, err := uuid.Parse(idStr)
    
    // 2. 查询参数解析
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    
    // 3. 请求体解析
    var req RequestStruct
    json.NewDecoder(r.Body).Decode(&req)
    
    // 4. 获取认证用户 ID
    userID, err := GetUserID(ctx)
    
    // 5. 调用 Service
    result, err := h.service.Method(ctx, params)
    
    // 6. 统一响应格式
    WriteJSON(w, result, http.StatusOK)
    WriteAPIError(w, "ERROR_CODE", "错误消息", nil, http.StatusInternalServerError)
}
```

### 响应格式

**成功响应**：
```json
{
  "data": { ... },
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

**错误响应**：
```json
{
  "code": "NOT_FOUND",
  "message": "资源不存在",
  "details": null,
  "request_id": "abc-123"
}
```

### 常用工具函数

| 函数 | 用途 |
|------|------|
| `WriteJSON(w, data, status)` | 统一 JSON 响应 |
| `WriteAPIError(w, code, message, details, status)` | 统一错误响应 |
| `GetUserID(ctx)` | 从上下文获取认证用户 ID |
| `r.PathValue("param")` | 获取路径参数 |
| `r.URL.Query().Get("key")` | 获取查询参数 |

## Dependencies

### 内部依赖

```
api/handlers/ → services/ → database/ → core/
api/middleware/ → cache/ (Redis 限流)
```

### 外部依赖

| Package | 用途 |
|---------|------|
| `encoding/json` | JSON 编解码 |
| `net/http` | HTTP 服务器 |
| `github.com/google/uuid` | UUID 处理 |

## For AI Agents

### 开发指南

1. **新增 API 端点**
   - 在对应的 `handlers/*.go` 文件中添加方法
   - 定义请求/响应结构体（带 JSON 标签）
   - 使用 `WriteJSON` 和 `WriteAPIError` 统一响应
   - 错误码使用大写蛇形命名（如 `NOT_FOUND`）

2. **添加新资源 Handler**
   - 创建 `handlers/resource.go` 文件
   - 遵循 `ResourceHandler` + `NewResourceHandler` 模式
   - 注入对应的 `services.ResourceService`

3. **参数验证**
   - 在 Handler 层进行基本验证（必填字段、格式检查）
   - 业务规则验证在 Service 层进行
   - 返回明确的错误码和消息

4. **认证和授权**
   - 需要认证的操作调用 `GetUserID(ctx)`
   - 未登录返回 `UNAUTHORIZED` 错误
   - 权限检查在 Service 层进行

### 代码约定

- Handler 方法名使用大写动词：`List`, `Get`, `Create`, `Update`, `Delete`
- 请求结构体后缀 `Request`：`CreatePostRequest`, `UpdateUserRequest`
- 响应结构体后缀 `Response`：`PostResponse`, `UserListResponse`
- 路径参数使用 `r.PathValue("id")`（Go 1.22+）
- 分页参数：`page`（默认 1），`page_size`（默认 20，最大 100）

### 测试策略

- Handler 层使用 `httptest` 进行集成测试
- 使用 Mock Service 避免真实数据库调用
- 验证响应状态码和 JSON 结构

<!-- MANUAL: -->
