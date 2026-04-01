<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-01 | Updated: 2026-04-01 -->

# middleware

## Purpose
HTTP 中间件模块，提供横切关注点的处理。

## Key Files
| File | Description |
|------|-------------|
| `ratelimit.go` | 基于 Redis 滑动窗口的限流器 |

## For AI Agents

### Working In This Directory
- 中间件遵循 `func(http.Handler) http.Handler` 签名
- 限流使用 Redis ZSET 实现滑动窗口算法
- 支持 IP 和用户 ID 两种限流 key 策略

### Rate Limiting Algorithm
```
滑动窗口实现：
1. ZRemRangeByScore 清除窗口外记录
2. ZCard 获取当前计数
3. 检查是否超过限制
4. ZAdd 添加当前请求
5. Expire 设置过期时间
```

### Predefined Limits
| Policy | Limit | Window |
|--------|-------|--------|
| Login | 10 | 1 min |
| Public API | 60 | 1 min |
| User Action | 100 | 1 min |

### Usage
```go
limiter := middleware.NewRateLimiter(redisClient, 60, time.Minute)
handler := middleware.RateLimitMiddleware(limiter, middleware.IPKeyFunc("api"))(nextHandler)
```

## Dependencies

### External
- `github.com/redis/go-redis/v9` - Redis 客户端

<!-- MANUAL: -->