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

// 缓存 TTL 配置
const (
	PermissionCacheTTL = 1 * time.Hour // 权限缓存 TTL
	RoleCacheTTL       = 1 * time.Hour // 角色信息缓存 TTL
)

// PermissionCache 权限缓存服务
type PermissionCache struct {
	cache    *cache.Service
	permRepo user.PermissionRepository
	client   *redis.Client // 用于 SCAN 操作
}

// NewPermissionCache 创建权限缓存服务
func NewPermissionCache(cacheService *cache.Service, permRepo user.PermissionRepository, redisClient *redis.Client) *PermissionCache {
	return &PermissionCache{
		cache:    cacheService,
		permRepo: permRepo,
		client:   redisClient,
	}
}

// GetPermission 检查用户是否拥有指定权限（带缓存）
// 返回 true 表示有权限，false 表示无权限
func (pc *PermissionCache) GetPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	key := cache.BuildUserPermsKey(roleID.String(), permission)

	// 尝试从缓存获取
	result, err := pc.cache.Get(ctx, key)
	if err == nil {
		// 缓存命中
		return result == "true", nil
	}

	// 缓存未命中，查询数据库
	hasPerm, err := pc.permRepo.CheckPermission(ctx, roleID, permission)
	if err != nil {
		return false, fmt.Errorf("failed to check permission from db: %w", err)
	}

	// 缓存结果
	cacheValue := "false"
	if hasPerm {
		cacheValue = "true"
	}
	if err := pc.cache.Set(ctx, key, cacheValue, PermissionCacheTTL); err != nil {
		// 缓存失败不影响业务，生产环境应记录日志
	}

	return hasPerm, nil
}

// GetRolePermissions 获取角色的所有权限（带缓存）
func (pc *PermissionCache) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]user.Permission, error) {
	key := cache.BuildRolePermsKey(roleID.String())

	// 尝试从缓存获取
	result, err := pc.cache.Get(ctx, key)
	if err == nil {
		// 缓存命中，反序列化
		var perms []user.Permission
		if err := json.Unmarshal([]byte(result), &perms); err != nil {
			// 反序列化失败，清除缓存并重新查询
			pc.cache.Delete(ctx, key)
		} else {
			return perms, nil
		}
	}

	// 缓存未命中，查询数据库
	perms, err := pc.permRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions from db: %w", err)
	}

	// 缓存结果
	if len(perms) > 0 {
		data, err := json.Marshal(perms)
		if err == nil {
			pc.cache.Set(ctx, key, string(data), RoleCacheTTL)
		}
	}

	return perms, nil
}

// InvalidateUserPermissions 清除用户所有权限缓存
// 需要扫描匹配 cadmus:user:perms:{userID}:* 的所有 key
func (pc *PermissionCache) InvalidateUserPermissions(ctx context.Context, userID uuid.UUID) error {
	pattern := fmt.Sprintf("cadmus:user:perms:%s:*", userID.String())
	return pc.deleteByPattern(ctx, pattern)
}

// InvalidateRolePermissions 清除角色权限缓存
func (pc *PermissionCache) InvalidateRolePermissions(ctx context.Context, roleID uuid.UUID) error {
	// 删除角色权限列表缓存
	rolePermsKey := cache.BuildRolePermsKey(roleID.String())

	// 同时删除所有使用该角色的用户权限缓存
	pattern := fmt.Sprintf("cadmus:user:perms:%s:*", roleID.String())

	keys := []string{rolePermsKey}
	matchedKeys, err := pc.scanKeys(ctx, pattern)
	if err == nil {
		keys = append(keys, matchedKeys...)
	}

	if len(keys) > 0 {
		return pc.cache.Delete(ctx, keys...)
	}
	return nil
}

// InvalidateAllPermissions 清除所有权限相关缓存
func (pc *PermissionCache) InvalidateAllPermissions(ctx context.Context) error {
	patterns := []string{
		"cadmus:user:perms:*",
		"cadmus:role:perms:*",
	}

	for _, pattern := range patterns {
		if err := pc.deleteByPattern(ctx, pattern); err != nil {
			return err
		}
	}
	return nil
}

// scanKeys 扫描匹配 pattern 的所有 key
func (pc *PermissionCache) scanKeys(ctx context.Context, pattern string) ([]string, error) {
	if pc.client == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	var keys []string
	iter := pc.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

// deleteByPattern 删除匹配 pattern 的所有 key
func (pc *PermissionCache) deleteByPattern(ctx context.Context, pattern string) error {
	keys, err := pc.scanKeys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		return pc.cache.Delete(ctx, keys...)
	}
	return nil
}