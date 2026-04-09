// Package middleware 提供了 Cadmus HTTP 中间件实现。
//
// 该文件包含基于 Redis 的限流器实现，包括：
//   - 滑动窗口限流算法
//   - IP 和用户级别的限流策略
//   - 预定义的限流配置
//
// 主要用途：
//
//	防止 API 滥用，保护系统免受恶意请求攻击。
//
// 注意事项：
//   - 需要 Redis 服务支持
//   - 采用 fail-open 策略（Redis 故障时允许请求通过）
//   - 通过 HTTP 响应头返回限流状态
//
// 作者：xfy
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter 基于 Redis 滑动窗口的限流器。
//
// 使用 Redis ZSET 实现滑动窗口限流算法，
// 支持分布式环境下的统一限流。
//
// 注意事项：
//   - client 为 nil 时降级为允许所有请求
type RateLimiter struct {
	// client Redis 客户端。nil 时降级为允许所有请求。
	client *redis.Client

	// limit 窗口内允许的最大请求数
	limit int

	// window 滑动窗口时长
	window time.Duration
}

// NewRateLimiter 创建限流器。
//
// 参数：
//   - client: Redis 客户端
//   - limit: 窗口内允许的最大请求数
//   - window: 滑动窗口时长
//
// 返回值：
//   - *RateLimiter: 新创建的限流器实例
func NewRateLimiter(client *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// Allow 检查是否允许请求（滑动窗口算法）。
//
// 使用 Redis ZSET 实现滑动窗口：
//  1. 以时间戳为 score，请求 ID 为 member
//  2. 清除窗口外的旧记录
//  3. 检查当前窗口内的记录数是否超过限制
//
// 参数：
//   - ctx: 上下文
//   - key: 限流键名（通常包含 IP 或用户 ID）
//
// 返回值：
//   - bool: true 表示允许请求，false 表示被限流
//
// 注意：
//   - Redis 错误时返回 true（fail-open 策略）
func (rl *RateLimiter) Allow(ctx context.Context, key string) bool {
	// 降级模式：无 Redis 客户端时允许所有请求
	if rl.client == nil {
		return true
	}

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

// Remaining 获取剩余可用请求次数。
//
// 查询当前窗口内还可以发送的请求数量。
//
// 参数：
//   - ctx: 上下文
//   - key: 限流键名
//
// 返回值：
//   - int: 剩余可用次数，最小为 0
func (rl *RateLimiter) Remaining(ctx context.Context, key string) int {
	// 降级模式：无 Redis 客户端时返回最大限制
	if rl.client == nil {
		return rl.limit
	}

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

// RateLimitMiddleware 创建限流中间件。
//
// 创建一个 HTTP 中间件，对请求进行限流控制。
// 被限流时返回 429 状态码，并设置限流相关的响应头。
//
// 参数：
//   - limiter: 限流器实例
//   - keyFunc: 限流键生成函数，用于从请求中提取限流键
//
// 返回值：
//   - 中间件函数
//
// 响应头：
//   - X-RateLimit-Limit: 窗口内最大请求数
//   - X-RateLimit-Remaining: 剩余可用次数
//   - X-RateLimit-Reset: 窗口重置时间（Unix 时间戳）
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
				_, _ = w.Write([]byte(`{"code":"RATE_LIMIT_EXCEEDED","message":"请求过于频繁，请稍后再试","details":null}`))
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

// IPKeyFunc 基于 IP 的限流键生成函数。
//
// 从请求中提取客户端 IP 地址，生成限流键。
// 支持 X-Real-IP、X-Forwarded-For 和 RemoteAddr。
//
// 参数：
//   - prefix: 键前缀，用于区分不同的限流场景
//
// 返回值：
//   - 键生成函数
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

// UserKeyFunc 基于用户 ID 的限流键生成函数。
//
// 从请求中提取用户 ID，生成限流键。
// 如果用户未认证，则回退到 IP 限流。
//
// 参数：
//   - prefix: 键前缀
//   - getUserID: 从请求中获取用户 ID 的函数
//
// 返回值：
//   - 键生成函数
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

// 预定义限流策略。
var (
	// LoginLimit 登录限流：10 次/分钟
	// 防止暴力破解攻击
	LoginLimit = 10

	// LoginWindow 登录窗口：1 分钟
	LoginWindow = time.Minute

	// PublicAPILimit 公开 API 限流：60 次/分钟
	// 适用于无需认证的公开接口
	PublicAPILimit = 60

	// PublicAPIWindow 公开 API 窗口：1 分钟
	PublicAPIWindow = time.Minute

	// UserActionLimit 用户操作限流：100 次/分钟
	// 适用于已认证用户的常规操作
	UserActionLimit = 100

	// UserActionWindow 用户操作窗口：1 分钟
	UserActionWindow = time.Minute
)
