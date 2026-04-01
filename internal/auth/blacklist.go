// Package auth 提供了 Cadmus 认证和授权功能的核心实现。
//
// 该文件包含 Token 黑名单相关的核心逻辑，包括：
//   - Token 黑名单接口定义
//   - Redis 实现的黑名单存储
//   - 黑名单 key 的构建和管理
//
// 主要用途：
//
//	用于实现安全的用户登出机制，使已注销的 token 无法再次使用。
//
// 注意事项：
//   - 黑名单功能依赖 Redis 存储
//   - 黑名单记录会随 token 过期自动清理
//   - 需确保 Redis 服务高可用
//
// 作者：xfy
package auth

import (
	"context"
	"fmt"
	"time"

	"rua.plus/cadmus/internal/cache"
)

// TokenBlacklist token 黑名单接口（基于 jti）。
//
// 该接口定义了 token 黑名单的核心操作，支持将 token 加入黑名单
// 和检查 token 是否在黑名单中。使用 JWT ID（jti）作为唯一标识。
//
// 实现要求：
//   - 所有方法必须是并发安全的
//   - AddToBlacklist 应支持自动过期清理
//   - IsBlacklisted 应高效查询，避免性能瓶颈
type TokenBlacklist interface {
	// AddToBlacklist 将 token 加入黑名单。
	//
	// 参数：
	//   - ctx: 上下文，用于超时控制
	//   - tokenID: JWT ID（jti），token 的唯一标识
	//   - expiry: token 的过期时间，用于设置 TTL
	//
	// 返回值：
	//   - err: 写入失败时返回错误
	AddToBlacklist(ctx context.Context, tokenID string, expiry time.Time) error

	// IsBlacklisted 检查 token 是否在黑名单中。
	//
	// 参数：
	//   - ctx: 上下文，用于超时控制
	//   - tokenID: JWT ID（jti）
	//
	// 返回值：
	//   - true: token 在黑名单中
	//   - false: token 不在黑名单中
	IsBlacklisted(ctx context.Context, tokenID string) bool
}

// RedisTokenBlacklist Redis 实现的 token 黑名单。
//
// 该结构使用 Redis 作为存储后端，实现高效的黑名单操作。
// 利用 Redis 的 TTL 功能自动清理过期的黑名单记录。
//
// 注意事项：
//   - 依赖 RedisClient 进行存储操作
//   - 黑名单 key 使用统一前缀便于管理
type RedisTokenBlacklist struct {
	client *cache.RedisClient  // Redis 客户端，用于存储操作
}

// NewRedisTokenBlacklist 创建 Redis token 黑名单实例。
//
// 该函数初始化 RedisTokenBlacklist，绑定 Redis 客户端。
//
// 参数：
//   - client: Redis 客户端实例，已配置好连接
//
// 返回值：
//   - 返回初始化完成的 RedisTokenBlacklist 实例
//
// 使用示例：
//   blacklist := auth.NewRedisTokenBlacklist(redisClient)
//   authSvc := auth.NewAuthService(jwtSvc, userRepo).WithBlacklist(blacklist)
//
// 注意事项：
//   - 确保 Redis 客户端已正确配置和连接
//   - Redis 服务应具备高可用性
func NewRedisTokenBlacklist(client *cache.RedisClient) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{client: client}
}

// blacklistKeyPrefix 黑名单 key 的统一前缀。
//
// 使用统一前缀便于：
//   - 区分黑名单 key 和其他业务 key
//   - 批量管理和清理黑名单数据
//   - 监控和统计黑名单使用情况
const blacklistKeyPrefix = "cadmus:jwt:blacklist"

// buildBlacklistKey 构建黑名单 Redis key。
//
// 将 token ID（jti）转换为完整的 Redis key 格式。
//
// 参数：
//   - tokenID: JWT ID（jti）
//
// 返回值：
//   - 格式为 "cadmus:jwt:blacklist:{tokenID}" 的 key
func buildBlacklistKey(tokenID string) string {
	return fmt.Sprintf("%s:%s", blacklistKeyPrefix, tokenID)
}

// AddToBlacklist 将 token jti 加入黑名单。
//
// 该方法将 token 的唯一标识写入 Redis，设置 TTL 为 token 的剩余有效期。
// 当 token 自然过期时，黑名单记录也会自动删除，避免存储浪费。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - tokenID: JWT ID（jti），token 的唯一标识
//   - expiry: token 的过期时间，用于计算 TTL
//
// 返回值：
//   - err: Redis 写入失败时返回错误
//   - 若 token 已过期，返回 nil（无需加入黑名单）
//
// 使用示例：
//   err := blacklist.AddToBlacklist(ctx, jti, claims.ExpiresAt.Time)
//   if err != nil {
//       // 处理写入失败
//   }
//
// 注意事项：
//   - TTL 基于 token 过期时间计算，自动清理
//   - 已过期的 token 无需加入黑名单
func (b *RedisTokenBlacklist) AddToBlacklist(ctx context.Context, tokenID string, expiry time.Time) error {
	key := buildBlacklistKey(tokenID)

	// 计算 TTL：从当前时间到过期时间的剩余时长
	// 过期后自动删除，无需手动清理
	ttl := time.Until(expiry)
	if ttl <= 0 {
		// token 已过期，无需加入黑名单
		// 自然过期的 token 无法再使用，黑名单无意义
		return nil
	}

	// 使用 "1" 作为值，key 本身携带语义
	// TTL 确保过期自动清理
	return b.client.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted 检查 token jti 是否在黑名单中。
//
// 该方法查询 Redis 检查指定 token 是否已被注销。
// 是认证中间件判断 token 有效性的关键步骤。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - tokenID: JWT ID（jti）
//
// 返回值：
//   - true: token 在黑名单中，已被注销
//   - false: token 不在黑名单中，可正常使用
//
// 使用示例：
//   if blacklist.IsBlacklisted(ctx, jti) {
//       // 拒绝请求，token 已被注销
//   }
//
// 注意事项：
//   - 查询失败时返回 false，避免误拒绝有效请求
//   - 高频调用场景建议优化 Redis 连接池配置
func (b *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, tokenID string) bool {
	key := buildBlacklistKey(tokenID)
	result := b.client.Exists(ctx, key)
	return result.Val() > 0
}