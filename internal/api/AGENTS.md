<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-01 | Updated: 2026-04-01 -->

# api

## Purpose
REST API 模块，包含 HTTP 处理器和中间件。

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `handlers/` | REST API 处理器：认证、文章、评论、媒体、搜索、RSS (see `handlers/AGENTS.md`) |
| `middleware/` | HTTP 中间件：限流 (see `middleware/AGENTS.md`) |

## For AI Agents

### Working In This Directory
- Handler 负责 HTTP 层：请求解析、响应序列化
- 业务逻辑在 `services/` 层，Handler 仅调用 Service
- 中间件用于横切关注点：认证、限流、日志

### Architecture
```
Request → Middleware Chain → Handler → Service → Repository
```

### Common Patterns

**Handler 结构**:
```go
type PostHandler struct {
    service *services.PostService
}

func NewPostHandler(service *services.PostService) *PostHandler {
    return &PostHandler{service: service}
}
```

**中间件模式**:
```go
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow(r.Context(), key) {
                http.Error(w, "rate limit exceeded", 429)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Dependencies

### Internal
- `internal/services` - 业务逻辑服务
- `internal/auth` - JWT 认证
- `internal/cache` - Redis 限流

<!-- MANUAL: -->