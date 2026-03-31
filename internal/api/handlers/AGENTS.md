<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# handlers

## Purpose
REST API 处理器，负责 HTTP 请求解析和响应序列化。

## Key Files
| File | Description |
|------|-------------|
| `auth.go` | 认证处理器：注册、登录、注销、刷新、当前用户 |
| `post.go` | 文章处理器：CRUD、发布、回滚、点赞、版本历史 |
| `comment.go` | 评论处理器：CRUD、点赞、审核、批量操作 |
| `media.go` | 媒体处理器：上传、删除、列表 |
| `search.go` | 搜索处理器：全文搜索、搜索建议 |
| `rss.go` | RSS 处理器：Feed 生成 |
| `admin.go` | 管理员处理器：角色管理、用户管理、批量操作 |
| `middleware.go` | 中间件：认证、权限检查、Token 黑名单 |

## For AI Agents

### Working In This Directory
- Handler 仅负责 HTTP 层，业务逻辑在 Service 层
- 使用 Service Container 获取服务实例
- 错误响应使用 `WriteAPIError` 统一格式

### Handler Pattern
```go
type PostHandler struct {
    service *services.PostService
}

func NewPostHandler(service *services.PostService) *PostHandler {
    return &PostHandler{service: service}
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. 解析请求
    var req CreatePostRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 2. 调用 Service
    post, err := h.service.Create(r.Context(), req)
    if err != nil {
        WriteAPIError(w, "CREATE_FAILED", err.Error(), nil, http.StatusBadRequest)
        return
    }

    // 3. 返回响应
    json.NewEncoder(w).Encode(post)
}
```

### Middleware Chain
```
AuthMiddlewareWithBlacklist → CachedPermissionMiddleware → Handler
```

### API Error Format
```json
{
    "code": "VALIDATION_ERROR",
    "message": "请求参数无效",
    "details": ["标题不能为空"],
    "request_id": "abc-123"
}
```

<!-- MANUAL: -->