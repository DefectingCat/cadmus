// Package auth 提供了 Cadmus 认证和授权功能的核心实现。
//
// 该文件包含权限缓存相关的核心逻辑，包括：
//   - 权限查询结果的缓存管理
//   - 角色权限信息的缓存存储
//   - 缓存失效和批量清理机制
//
// 主要用途：
//
//	用于提升权限验证性能，减少数据库查询压力。
//
// 注意事项：
//   - 缓存依赖 Redis 存储
//   - 缓存 TTL 设置需权衡性能和数据一致性
//   - 权限变更时需主动失效缓存
//
// 作者：xfy
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rua.plus/cadmus/internal/cache"
	"rua.plus/cadmus/internal/core/user"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// 缓存 TTL 配置常量。
//
// 这些常量定义了权限缓存的默认过期时间，需根据业务场景调整：
//   - PermissionCacheTTL: 单个权限验证结果的缓存时间
//   - RoleCacheTTL: 角色权限列表的缓存时间
const (
	// PermissionCacheTTL 权限缓存过期时间。
	// 设为 1 小时，平衡查询频率和数据一致性。
	PermissionCacheTTL = 1 * time.Hour // 权限缓存 TTL

	// RoleCacheTTL 角色信息缓存过期时间。
	// 设为 1 小时，角色权限变更频率通常较低。
	RoleCacheTTL = 1 * time.Hour // 角色信息缓存 TTL
)

// PermissionCache 权限缓存服务，提供权限信息的缓存管理。
//
// 该服务封装了权限验证的缓存逻辑，减少数据库查询次数。
// 支持单个权限查询缓存、角色权限列表缓存，以及缓存失效管理。
//
// 注意事项：
//   - 缓存命中时直接返回结果，不查询数据库
//   - 缓存失败不影响业务流程，降级到数据库查询
//   - 权限变更后需调用失效方法清除缓存
type PermissionCache struct {
	cache    *cache.Service            // 缓存服务，用于存取操作
	permRepo user.PermissionRepository // 权限数据仓库，用于数据库查询
	client   *redis.Client             // Redis 客户端，用于 SCAN 操作
}

// NewPermissionCache 创建权限缓存服务实例。
//
// 该函数初始化 PermissionCache，绑定缓存服务、权限仓库和 Redis 客户端。
// Redis 客户端用于批量失效缓存时的 SCAN 操作。
//
// 参数：
//   - cacheService: 缓存服务实例，用于缓存存取
//   - permRepo: 权限数据仓库，用于数据库查询
//   - redisClient: Redis 客户端，用于 SCAN 操作（可选，仅批量失效时需要）
//
// 返回值：
//   - 返回初始化完成的 PermissionCache 实例
//
// 使用示例：
//
//	permCache := auth.NewPermissionCache(cacheSvc, permRepo, redisClient)
//	hasPerm, err := permCache.GetPermission(ctx, roleID, "user:read")
//
// 注意事项：
//   - 若不使用批量失效功能，redisClient 可为 nil
//   - 确保 cacheService 和 permRepo 已正确初始化
func NewPermissionCache(cacheService *cache.Service, permRepo user.PermissionRepository, redisClient *redis.Client) *PermissionCache {
	return &PermissionCache{
		cache:    cacheService,
		permRepo: permRepo,
		client:   redisClient,
	}
}

// GetPermission 检查用户是否拥有指定权限（带缓存）。
//
// 该方法优先从缓存获取权限验证结果，缓存未命中时查询数据库并缓存结果。
// 使用 roleID 和 permission 组合作为缓存 key。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - roleID: 用户角色 ID
//   - permission: 权限标识符，如 "user:read"、"content:write"
//
// 返回值：
//   - hasPerm: true 表示有权限，false 表示无权限
//   - err: 可能的错误包括数据库查询失败
//
// 使用示例：
//
//	hasPerm, err := permCache.GetPermission(ctx, roleID, "user:read")
//	if err != nil {
//	    // 处理查询错误
//	}
//	if !hasPerm {
//	    // 拒绝操作
//	}
//
// 注意事项：
//   - 缓存命中时返回 true/false，不访问数据库
//   - 缓存写入失败不影响返回结果
//   - 结果缓存 1 小时（PermissionCacheTTL）
func (pc *PermissionCache) GetPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	key := cache.BuildUserPermsKey(roleID.String(), permission)

	// 步骤1: 尝试从缓存获取
	// 命中时直接返回，避免数据库查询
	result, err := pc.cache.Get(ctx, key)
	if err == nil {
		// 缓存命中
		return result == "true", nil
	}

	// 步骤2: 缓存未命中，查询数据库
	// 从权限仓库获取真实的权限状态
	hasPerm, err := pc.permRepo.CheckPermission(ctx, roleID, permission)
	if err != nil {
		return false, fmt.Errorf("failed to check permission from db: %w", err)
	}

	// 步骤3: 缓存结果
	// 缓存失败不影响业务，生产环境应记录日志
	cacheValue := "false"
	if hasPerm {
		cacheValue = "true"
	}
	if err := pc.cache.Set(ctx, key, cacheValue, PermissionCacheTTL); err != nil {
		// 缓存失败不影响业务，生产环境应记录日志
	}

	return hasPerm, nil
}

// GetRolePermissions 获取角色的所有权限（带缓存）。
//
// 该方法获取指定角色的完整权限列表，优先从缓存读取。
// 权限列表以 JSON 格式存储在缓存中。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - roleID: 角色 ID
//
// 返回值：
//   - perms: 权限对象列表
//   - err: 可能的错误包括数据库查询失败、JSON 反序列化失败
//
// 使用示例：
//
//	perms, err := permCache.GetRolePermissions(ctx, roleID)
//	if err != nil {
//	    // 处理错误
//	}
//	for _, perm := range perms {
//	    // 处理每个权限
//	}
//
// 注意事项：
//   - 反序列化失败时会清除缓存并重新查询
//   - 空权限列表不会写入缓存
func (pc *PermissionCache) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]user.Permission, error) {
	key := cache.BuildRolePermsKey(roleID.String())

	// 步骤1: 尝试从缓存获取
	result, err := pc.cache.Get(ctx, key)
	if err == nil {
		// 缓存命中，反序列化 JSON
		var perms []user.Permission
		if err := json.Unmarshal([]byte(result), &perms); err != nil {
			// 反序列化失败，清除无效缓存并重新查询
			pc.cache.Delete(ctx, key)
		} else {
			return perms, nil
		}
	}

	// 步骤2: 缓存未命中，查询数据库
	// 从权限仓库获取角色的所有权限
	perms, err := pc.permRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions from db: %w", err)
	}

	// 步骤3: 缓存结果
	// 仅当有权限数据时才缓存，避免缓存空值
	if len(perms) > 0 {
		data, err := json.Marshal(perms)
		if err == nil {
			pc.cache.Set(ctx, key, string(data), RoleCacheTTL)
		}
	}

	return perms, nil
}

// InvalidateUserPermissions 清除用户所有权限缓存。
//
// 该方法删除指定用户相关的所有权限缓存记录。
// 需要扫描匹配模式的 key 并批量删除。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - userID: 用户 ID
//
// 返回值：
//   - err: 扫描或删除失败时返回错误
//
// 使用示例：
//
//	err := permCache.InvalidateUserPermissions(ctx, userID)
//	if err != nil {
//	    // 记录日志，但不影响业务
//	}
//
// 注意事项：
//   - 需要 Redis 客户端支持 SCAN 操作
//   - 大量匹配 key 时可能耗时较长
func (pc *PermissionCache) InvalidateUserPermissions(ctx context.Context, userID uuid.UUID) error {
	// 构建匹配模式：扫描用户相关的所有权限 key
	pattern := fmt.Sprintf("cadmus:user:perms:%s:*", userID.String())
	return pc.deleteByPattern(ctx, pattern)
}

// InvalidateRolePermissions 清除角色权限缓存。
//
// 该方法删除角色权限列表缓存，以及使用该角色的用户权限缓存。
// 用于角色权限变更后的缓存同步。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - roleID: 角色 ID
//
// 返回值：
//   - err: 删除失败时返回错误
//
// 使用示例：
//
//	err := permCache.InvalidateRolePermissions(ctx, roleID)
//	if err != nil {
//	    // 记录日志
//	}
//
// 注意事项：
//   - 同时删除角色权限列表和用户权限缓存
//   - 确保 Redis 客户端可用
func (pc *PermissionCache) InvalidateRolePermissions(ctx context.Context, roleID uuid.UUID) error {
	// 删除角色权限列表缓存
	rolePermsKey := cache.BuildRolePermsKey(roleID.String())

	// 同时删除所有使用该角色的用户权限缓存
	// 扫描匹配该角色的用户权限 key
	pattern := fmt.Sprintf("cadmus:user:perms:%s:*", roleID.String())

	keys := []string{rolePermsKey}
	matchedKeys, err := pc.scanKeys(ctx, pattern)
	if err == nil {
		keys = append(keys, matchedKeys...)
	}

	// 执行批量删除
	if len(keys) > 0 {
		return pc.cache.Delete(ctx, keys...)
	}
	return nil
}

// InvalidateAllPermissions 清除所有权限相关缓存。
//
// 该方法批量删除所有权限缓存，用于系统重置或大规模权限变更。
// 清理范围包括用户权限和角色权限两类缓存。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//
// 返回值：
//   - err: 扫描或删除失败时返回错误
//
// 使用示例：
//
//	err := permCache.InvalidateAllPermissions(ctx)
//	if err != nil {
//	    // 记录日志
//	}
//
// 注意事项：
//   - 操作范围大，可能耗时较长
//   - 执行后所有权限需重新从数据库查询
func (pc *PermissionCache) InvalidateAllPermissions(ctx context.Context) error {
	// 定义需要清理的缓存模式
	patterns := []string{
		"cadmus:user:perms:*", // 所有用户权限缓存
		"cadmus:role:perms:*", // 所有角色权限缓存
	}

	// 遍历模式执行批量删除
	for _, pattern := range patterns {
		if err := pc.deleteByPattern(ctx, pattern); err != nil {
			return err
		}
	}
	return nil
}

// scanKeys 扫描匹配 pattern 的所有 key。
//
// 该私有方法使用 Redis SCAN 命令遍历匹配指定模式的 key。
// 用于批量删除缓存时的 key 查找。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - pattern: Redis key 匹配模式
//
// 返回值：
//   - keys: 匹配的 key 列表
//   - err: Redis 客户端不可用或扫描失败时返回错误
func (pc *PermissionCache) scanKeys(ctx context.Context, pattern string) ([]string, error) {
	// 检查 Redis 客户端是否可用
	if pc.client == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	// 使用 SCAN 命令遍历 key
	// SCAN 是增量式迭代，不会阻塞 Redis
	var keys []string
	iter := pc.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	// 检查迭代过程中的错误
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

// deleteByPattern 删除匹配 pattern 的所有 key。
//
// 该私有方法先扫描匹配的 key，然后批量删除。
// 是失效方法的底层实现。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - pattern: Redis key 匹配模式
//
// 返回值：
//   - err: 扫描或删除失败时返回错误
func (pc *PermissionCache) deleteByPattern(ctx context.Context, pattern string) error {
	// 扫描匹配的 key
	keys, err := pc.scanKeys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	// 执行批量删除
	if len(keys) > 0 {
		return pc.cache.Delete(ctx, keys...)
	}
	return nil
}
