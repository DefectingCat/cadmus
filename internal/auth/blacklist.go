package auth

import (
	"context"
	"fmt"
	"time"

	"rua.plus/cadmus/internal/cache"
)

// TokenBlacklist token 黑名单接口（基于 jti）
type TokenBlacklist interface {
	AddToBlacklist(ctx context.Context, tokenID string, expiry time.Time) error
	IsBlacklisted(ctx context.Context, tokenID string) bool
}

// RedisTokenBlacklist Redis 实现的 token 黑名单
type RedisTokenBlacklist struct {
	client *cache.RedisClient
}

// NewRedisTokenBlacklist 创建 Redis token 黑名单
func NewRedisTokenBlacklist(client *cache.RedisClient) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{client: client}
}

// blacklistKeyPrefix 黑名单 key 前缀
const blacklistKeyPrefix = "cadmus:jwt:blacklist"

// buildBlacklistKey 构建黑名单 key
func buildBlacklistKey(tokenID string) string {
	return fmt.Sprintf("%s:%s", blacklistKeyPrefix, tokenID)
}

// AddToBlacklist 将 token jti 加入黑名单
// expiry 是 token 的过期时间，Redis key 会在此时间后自动删除
func (b *RedisTokenBlacklist) AddToBlacklist(ctx context.Context, tokenID string, expiry time.Time) error {
	key := buildBlacklistKey(tokenID)
	// 计算从现在到过期时间的持续时间
	ttl := time.Until(expiry)
	if ttl <= 0 {
		// token 已过期，无需加入黑名单
		return nil
	}
	// 使用 "1" 作为值，key 本身携带了意义
	return b.client.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted 检查 token jti 是否在黑名单中
func (b *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, tokenID string) bool {
	key := buildBlacklistKey(tokenID)
	result := b.client.Exists(ctx, key)
	return result.Val() > 0
}