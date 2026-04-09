// Package cache 提供 Redis 缓存服务的实现。
//
// 该文件包含缓存服务的核心实现，包括：
//   - CacheService 接口定义
//   - Service 缓存服务实现
//   - 缓存穿透和缓存击穿防护机制
//   - 批量操作支持
//
// 主要用途：
//
//	为系统提供高性能缓存服务，减轻数据库压力，提升响应速度。
//
// 注意事项：
//   - 所有方法必须是并发安全的
//   - 支持缓存穿透防护（缓存空值）
//   - 支持缓存击穿防护（互斥锁）
//   - 使用 Redis 作为存储后端
//
// 作者：xfy
package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// 错误定义。
var (
	// ErrNotFound 缓存键不存在错误
	ErrNotFound = errors.New("cache: key not found")

	// ErrNullValue 空值标记错误，表示缓存的是空值标记而非实际数据
	ErrNullValue = errors.New("cache: null value marker")

	// ErrLockFailed 获取锁失败错误
	ErrLockFailed = errors.New("cache: failed to acquire lock")

	// ErrLockTimeout 获取锁超时错误
	ErrLockTimeout = errors.New("cache: lock acquisition timeout")
)

// 空值标记和锁相关常量。
const (
	// NullMarker 空值标记，用于缓存穿透防护
	// 当数据不存在时，缓存此标记而非不缓存
	NullMarker = "__NULL__"

	// NullMarkerTTL 空值标记的过期时间，较短以避免长期占用
	NullMarkerTTL = 5 * time.Minute

	// LockValue 锁的值，固定为 "1"
	LockValue = "1"

	// DefaultLockTTL 锁的默认过期时间
	DefaultLockTTL = 10 * time.Second

	// LockWaitInterval 获取锁的重试间隔
	LockWaitInterval = 100 * time.Millisecond

	// LockMaxRetry 获取锁的最大重试次数，最多等待 5 秒
	LockMaxRetry = 50
)

// CacheService 缓存服务接口。
//
// 定义缓存操作的标准方法，支持基础操作、防护操作和批量操作。
type CacheService interface {
	// 基础操作
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)

	// 带防护的操作
	GetWithNullProtection(ctx context.Context, key string, fetchFn func() (interface{}, error), ttl time.Duration) (interface{}, error)
	GetWithMutex(ctx context.Context, key string, fetchFn func() (interface{}, error), ttl time.Duration) (interface{}, error)

	// 批量操作
	GetMulti(ctx context.Context, keys []string) (map[string]string, error)
	SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error

	// 辅助方法
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

// Service 缓存服务实现。
//
// 实现 CacheService 接口，封装 Redis 客户端提供缓存功能。
type Service struct {
	client *RedisClient
}

// NewService 创建缓存服务实例。
//
// 参数：
//   - client: Redis 客户端实例
//
// 返回值：
//   - 缓存服务实例
//
// 使用示例：
//
//	client, _ := NewRedisClient(DefaultConfig())
//	svc := NewService(client)
func NewService(client *RedisClient) *Service {
	return &Service{
		client: client,
	}
}

// Get 获取缓存值。
//
// 从缓存中获取指定键的值。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//
// 返回值：
//   - value: 缓存值字符串
//   - err: 键不存在返回 ErrNotFound，空值标记返回 ErrNullValue
func (s *Service) Get(ctx context.Context, key string) (string, error) {
	result, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound
		}
		return "", err
	}
	// 检查是否为空值标记
	if result == NullMarker {
		return "", ErrNullValue
	}
	return result, nil
}

// Set 设置缓存值。
//
// 将值存入缓存，设置过期时间。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 缓存值（支持字符串、[]byte 等类型）
//   - ttl: 过期时间
//
// 返回值：
//   - err: 设置失败错误
func (s *Service) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

// Delete 删除缓存。
//
// 删除一个或多个缓存键。
//
// 参数：
//   - ctx: 上下文
//   - keys: 要删除的缓存键列表
//
// 返回值：
//   - err: 删除失败错误
func (s *Service) Delete(ctx context.Context, keys ...string) error {
	return s.client.Del(ctx, keys...).Err()
}

// SetNX 仅当键不存在时设置。
//
// 用于实现分布式锁或防重复操作。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 缓存值
//   - ttl: 过期时间
//
// 返回值：
//   - ok: true 表示设置成功（键不存在），false 表示键已存在
//   - err: 操作错误
func (s *Service) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, ttl).Result()
}

// GetWithNullProtection 缓存穿透防护：缓存空值。
//
// 当数据不存在时，缓存一个空值标记，防止大量请求穿透到数据库。
// 适用于查询结果可能为空的场景。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - fetchFn: 数据获取函数，当缓存未命中时调用
//   - ttl: 缓存过期时间
//
// 返回值：
//   - data: 缓存数据或 fetchFn 返回的数据
//   - err: 数据不存在返回 ErrNotFound，其他错误原样返回
//
// 使用示例：
//
//	data, err := svc.GetWithNullProtection(ctx, "user:123", func() (interface{}, error) {
//	    return db.GetUser(123)
//	}, 5*time.Minute)
func (s *Service) GetWithNullProtection(ctx context.Context, key string, fetchFn func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// 步骤1: 尝试从缓存获取
	result, err := s.Get(ctx, key)
	if err == nil {
		return result, nil
	}

	// 检查是否为空值标记
	if errors.Is(err, ErrNullValue) {
		return nil, ErrNotFound
	}

	// 步骤2: 缓存未命中，从数据源获取
	data, err := fetchFn()
	if err != nil {
		return nil, err
	}

	// 步骤3: 缓存结果
	if data == nil {
		// 数据不存在，缓存空值标记，使用较短的 TTL
		if err := s.Set(ctx, key, NullMarker, NullMarkerTTL); err != nil {
			return nil, err
		}
		return nil, ErrNotFound
	}

	// 缓存实际数据
	if err := s.Set(ctx, key, data, ttl); err != nil {
		return nil, err
	}

	return data, nil
}

// GetWithMutex 缓存击穿防护：互斥锁。
//
// 当热点键失效时，只有一个请求去查询数据库，其他请求等待。
// 适用于高并发访问热点数据的场景。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - fetchFn: 数据获取函数，当缓存未命中时调用
//   - ttl: 缓存过期时间
//
// 返回值：
//   - data: 缓存数据或 fetchFn 返回的数据
//   - err: 获取锁超时返回 ErrLockTimeout，其他错误原样返回
//
// 使用示例：
//
//	data, err := svc.GetWithMutex(ctx, "hot:article:123", func() (interface{}, error) {
//	    return db.GetArticle(123)
//	}, 10*time.Minute)
func (s *Service) GetWithMutex(ctx context.Context, key string, fetchFn func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// 步骤1: 尝试从缓存获取
	result, err := s.Get(ctx, key)
	if err == nil {
		return result, nil
	}
	if errors.Is(err, ErrNullValue) {
		return nil, ErrNotFound
	}

	// 步骤2: 尝试获取分布式锁
	lockKey := key + ":lock"
	acquired, err := s.SetNX(ctx, lockKey, LockValue, DefaultLockTTL)
	if err != nil {
		return nil, err
	}

	if acquired {
		// 成功获取锁
		defer func() {
			_ = s.Delete(ctx, lockKey) //nolint:errcheck // 清理锁，失败不影响业务
		}()

		// 步骤3: 双重检查：再次尝试从缓存获取
		result, err = s.Get(ctx, key)
		if err == nil {
			return result, nil
		}
		if errors.Is(err, ErrNullValue) {
			return nil, ErrNotFound
		}

		// 步骤4: 从数据源获取
		data, err := fetchFn()
		if err != nil {
			return nil, err
		}

		// 步骤5: 缓存结果
		if data == nil {
			if err := s.Set(ctx, key, NullMarker, NullMarkerTTL); err != nil {
				return nil, err
			}
			return nil, ErrNotFound
		}

		if err := s.Set(ctx, key, data, ttl); err != nil {
			return nil, err
		}

		return data, nil
	}

	// 步骤6: 未获取锁，等待其他 goroutine 完成后重试
	for i := 0; i < LockMaxRetry; i++ {
		time.Sleep(LockWaitInterval)

		// 再次尝试从缓存获取
		result, err = s.Get(ctx, key)
		if err == nil {
			return result, nil
		}
		if errors.Is(err, ErrNullValue) {
			return nil, ErrNotFound
		}

		// 检查锁是否已被释放
		exists, err := s.Exists(ctx, lockKey)
		if err != nil {
			continue
		}
		if !exists {
			// 锁已释放但缓存仍未填充，重新尝试获取锁
			return s.GetWithMutex(ctx, key, fetchFn, ttl)
		}
	}

	return nil, ErrLockTimeout
}

// GetMulti 批量获取缓存值。
//
// 批量获取多个键的值，返回存在的键值对。
//
// 参数：
//   - ctx: 上下文
//   - keys: 缓存键列表
//
// 返回值：
//   - results: 键值对映射，仅包含存在的键
//   - err: 操作错误
func (s *Service) GetMulti(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}

	// 使用循环获取（可优化为 MGet）
	results := make(map[string]string)
	for _, key := range keys {
		val, err := s.Get(ctx, key)
		if err == nil {
			results[key] = val
		}
	}

	return results, nil
}

// SetMulti 批量设置缓存值。
//
// 批量设置多个键值对，使用相同的过期时间。
//
// 参数：
//   - ctx: 上下文
//   - items: 键值对映射
//   - ttl: 过期时间
//
// 返回值：
//   - err: 设置失败错误
func (s *Service) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	for key, value := range items {
		if err := s.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}

	return nil
}

// Exists 检查键是否存在。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//
// 返回值：
//   - exists: true 表示键存在
//   - err: 操作错误
func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Expire 设置键的过期时间。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - ttl: 新的过期时间
//
// 返回值：
//   - err: 操作错误
func (s *Service) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.Expire(ctx, key, ttl).Err()
}

// Client 获取底层 Redis 客户端。
//
// 返回 RedisClient 实例，用于执行更复杂的 Redis 操作。
//
// 返回值：
//   - Redis 客户端实例
func (s *Service) Client() *RedisClient {
	return s.client
}

// IsNullMarker 检查值是否为空值标记。
//
// 用于调用方判断从缓存获取的值是否为空值标记。
//
// 参数：
//   - value: 待检查的值
//
// 返回值：
//   - true: 是空值标记
//   - false: 不是空值标记
func IsNullMarker(value string) bool {
	return value == NullMarker
}
