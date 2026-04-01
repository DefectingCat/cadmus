package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter 基于 Redis 滑动窗口的限流器
type RateLimiter struct {
	client *redis.Client
	limit  int           // 窗口内允许的最大请求数
	window time.Duration // 滑动窗口时长
}

// NewRateLimiter 创建限流器
func NewRateLimiter(client *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// Allow 检查是否允许请求（滑动窗口算法）
// 使用 Redis ZSET 实现滑动窗口：
// 1. 以时间戳为 score，请求 ID 为 member
// 2. 清除窗口外的旧记录
// 3. 检查当前窗口内的记录数是否超过限制
func (rl *RateLimiter) Allow(ctx context.Context, key string) bool {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	// 使用 pipeline 减少网络往返
	pipe := rl.client.Pipeline()

	// 1. 移除窗口外的旧记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// 2. 获取当前窗口内的记录数
	countCmd := pipe.ZCard(ctx, key)

	// 3. 执行 pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		// Redis 错误时允许请求通过（fail-open 策略）
		return true
	}

	count, err := countCmd.Result()
	if err != nil {
		return true
	}

	// 检查是否超过限制
	if count >= int64(rl.limit) {
		return false
	}

	// 4. 添加当前请求记录
	nowNano := now.UnixNano()
	member := fmt.Sprintf("%d:%d", nowNano, time.Now().Nanosecond())

	pipe2 := rl.client.Pipeline()
	pipe2.ZAdd(ctx, key, redis.Z{Score: float64(nowNano), Member: member})
	// 设置过期时间，避免僵尸 key
	pipe2.Expire(ctx, key, rl.window+time.Minute)

	_, err = pipe2.Exec(ctx)
	if err != nil {
		return true
	}

	return true
}

// Remaining 获取剩余可用请求次数
func (rl *RateLimiter) Remaining(ctx context.Context, key string) int {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	// 清除旧记录并获取当前计数
	pipe := rl.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))
	countCmd := pipe.ZCard(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return rl.limit // 错误时返回最大限制
	}

	count, err := countCmd.Result()
	if err != nil {
		return rl.limit
	}

	remaining := rl.limit - int(count)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RateLimitMiddleware 创建限流中间件
// keyFunc 用于生成限流 key，通常基于 IP 或用户 ID
func RateLimitMiddleware(limiter *RateLimiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)

			if !limiter.Allow(r.Context(), key) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(limiter.window).Unix()))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"code":"RATE_LIMIT_EXCEEDED","message":"请求过于频繁，请稍后再试","details":null}`))
				return
			}

			// 设置限流信息头
			remaining := limiter.Remaining(r.Context(), key)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(limiter.window).Unix()))

			next.ServeHTTP(w, r)
		})
	}
}

// IPKeyFunc 基于 IP 的限流 key 生成函数
func IPKeyFunc(prefix string) func(*http.Request) string {
	return func(r *http.Request) string {
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
			if ip != "" {
				// X-Forwarded-For 可能包含多个 IP，取第一个
				if idx := len(ip); idx > 0 {
					for i, c := range ip {
						if c == ',' {
							ip = ip[:i]
							break
						}
					}
				}
			}
		}
		if ip == "" {
			ip = r.RemoteAddr
			// 移除端口
			if idx := len(ip); idx > 0 {
				for i := len(ip) - 1; i >= 0; i-- {
					if ip[i] == ':' {
						ip = ip[:i]
						break
					}
				}
			}
		}
		return fmt.Sprintf("ratelimit:%s:%s", prefix, ip)
	}
}

// UserKeyFunc 基于用户 ID 的限流 key 生成函数
func UserKeyFunc(prefix string, getUserID func(*http.Request) string) func(*http.Request) string {
	return func(r *http.Request) string {
		userID := getUserID(r)
		if userID == "" {
			// 未认证用户使用 IP
			return IPKeyFunc(prefix)(r)
		}
		return fmt.Sprintf("ratelimit:%s:user:%s", prefix, userID)
	}
}

// 预定义限流策略
var (
	// LoginLimit 登录限流：10 次/分钟
	LoginLimit = 10
	// LoginWindow 登录窗口：1 分钟
	LoginWindow = time.Minute

	// PublicAPILimit 公开 API 限流：60 次/分钟
	PublicAPILimit = 60
	// PublicAPIWindow 公开 API 窗口：1 分钟
	PublicAPIWindow = time.Minute

	// UserActionLimit 用户操作限流：100 次/分钟
	UserActionLimit = 100
	// UserActionWindow 用户操作窗口：1 分钟
	UserActionWindow = time.Minute
)