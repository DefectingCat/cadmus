<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# middleware

## 用途

`middleware` 目录包含 Cadmus 应用的 HTTP 中间件实现，提供横切关注点的复用能力。

**主要用途**：
- 请求限流控制，防止 API 滥用
- 认证和授权检查
- 请求日志记录
- 错误恢复和降级处理

## 文件列表

| 文件 | 功能 |
|------|------|
| `ratelimit.go` | 基于 Redis 的滑动窗口限流器实现 |

### ratelimit.go

基于 Redis ZSET 实现滑动窗口限流算法，支持：

- **IP 级别限流**：通过 `IPKeyFunc` 从请求提取客户端 IP
- **用户级别限流**：通过 `UserKeyFunc` 基于认证用户 ID
- **Fail-Open 策略**：Redis 故障时允许请求通过，保证服务可用性

**核心组件**：

```go
// RateLimiter - 限流器实例，通过 NewRateLimiter 创建
type RateLimiter struct {
    client *redis.Client   // Redis 客户端
    limit  int             // 窗口内最大请求数
    window time.Duration   // 滑动窗口时长
}

// RateLimitMiddleware - 创建 HTTP 限流中间件
func RateLimitMiddleware(limiter *RateLimiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler

// IPKeyFunc - 基于 IP 的限流键生成器
func IPKeyFunc(prefix string) func(*http.Request) string

// UserKeyFunc - 基于用户 ID 的限流键生成器
func UserKeyFunc(prefix string, getUserID func(*http.Request) string) func(*http.Request) string
```

**预定义限流策略**：

| 策略 | 限制次数 | 窗口时长 | 适用场景 |
|------|----------|----------|----------|
| `LoginLimit` | 10 | 1 分钟 | 登录接口，防止暴力破解 |
| `PublicAPILimit` | 60 | 1 分钟 | 公开 API，无需认证 |
| `UserActionLimit` | 100 | 1 分钟 | 已认证用户的常规操作 |

**HTTP 响应头**：

| 响应头 | 说明 |
|--------|------|
| `X-RateLimit-Limit` | 窗口内最大请求数 |
| `X-RateLimit-Remaining` | 剩余可用次数 |
| `X-RateLimit-Reset` | 窗口重置时间（Unix 时间戳） |

## 中间件链

请求处理流程：

```
Request → RateLimitMiddleware → AuthMiddleware → PermissionMiddleware → Handler
```

## 无子目录

当前目录无子目录。

## AI Agent 开发指南

### 新增中间件

1. **在 `middleware/` 目录下创建新文件**
   - 如 `auth.go`、`logging.go`
   - 遵循标准中间件签名：`func(next http.Handler) http.Handler`

2. **中间件构造模式**

   ```go
   // 1. 定义中间件配置结构体
   type Config struct {
       Option string
   }
   
   // 2. 构造函数返回中间件函数
   func NewMiddleware(config Config) func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               // 前置处理
               
               next.ServeHTTP(w, r)
               
               // 后置处理
           })
       }
   }
   ```

3. **中间件应用**

   ```go
   // 在路由注册时包装 Handler
   mux.Handle("/api/posts", RateLimitMiddleware(limiter, IPKeyFunc("api"))(postsHandler))
   ```

### 限流器使用

```go
// 1. 创建限流器（60 次/分钟）
limiter := NewRateLimiter(redisClient, 60, time.Minute)

// 2. 应用限流中间件（按 IP 限流）
mux.Handle("/api/public", RateLimitMiddleware(limiter, IPKeyFunc("public"))(handler))

// 3. 应用限流中间件（按用户 ID 限流）
mux.Handle("/api/user", RateLimitMiddleware(limiter, UserKeyFunc("user", getUserID))(handler))
```

### 开发规范

- **中间件职责单一**：每个中间件只处理一个横切关注点
- **可组合性**：中间件应能被链式组合
- **错误处理**：中间件中的错误应记录日志，避免中断请求链
- **性能敏感**：中间件对每个请求执行，避免耗时操作

### 测试策略

```go
// 使用 httptest 测试中间件
func TestRateLimitMiddleware(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    wrapped := RateLimitMiddleware(limiter, IPKeyFunc("test"))(handler)
    
    req := httptest.NewRequest("GET", "/test", nil)
    rr := httptest.NewRecorder()
    wrapped.ServeHTTP(rr, req)
    
    // 验证响应状态码和限流响应头
    if got := rr.Code; got != http.StatusOK {
        t.Errorf("status = %v, want %v", got, http.StatusOK)
    }
}
```

### 依赖

| 依赖 | 用途 |
|------|------|
| `github.com/redis/go-redis/v9` | Redis 客户端，用于限流存储 |
| `net/http` | HTTP 中间件标准接口 |
| `context` | 请求上下文传递 |

<!-- MANUAL: -->
