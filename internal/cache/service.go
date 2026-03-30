package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// 错误定义
var (
	ErrNotFound    = errors.New("cache: key not found")
	ErrNullValue   = errors.New("cache: null value marker")
	ErrLockFailed  = errors.New("cache: failed to acquire lock")
	ErrLockTimeout = errors.New("cache: lock acquisition timeout")
)

// 空值标记常量
const (
	NullMarker       = "__NULL__"
	NullMarkerTTL    = 5 * time.Minute
	LockValue        = "1"
	DefaultLockTTL   = 10 * time.Second
	LockWaitInterval = 100 * time.Millisecond
	LockMaxRetry     = 50 // 最多等待 5 秒
)

// CacheService 缓存服务接口
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

// Service 缓存服务实现
type Service struct {
	client *RedisClient
}

// NewService 创建缓存服务
func NewService(client *RedisClient) *Service {
	return &Service{
		client: client,
	}
}

// Get 获取缓存值
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

// Set 设置缓存值
func (s *Service) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

// Delete 删除缓存
func (s *Service) Delete(ctx context.Context, keys ...string) error {
	return s.client.Del(ctx, keys...).Err()
}

// SetNX 仅当 key 不存在时设置
func (s *Service) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, ttl).Result()
}

// GetWithNullProtection 缓存穿透防护：缓存空值
// 当数据不存在时，缓存一个空值标记，防止大量请求穿透到数据库
func (s *Service) GetWithNullProtection(ctx context.Context, key string, fetchFn func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// 1. 尝试从缓存获取
	result, err := s.Get(ctx, key)
	if err == nil {
		return result, nil
	}

	// 检查是否为空值标记
	if errors.Is(err, ErrNullValue) {
		return nil, ErrNotFound
	}

	// 2. 缓存未命中，从数据源获取
	data, err := fetchFn()
	if err != nil {
		return nil, err
	}

	// 3. 缓存结果
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

// GetWithMutex 缓存击穿防护：互斥锁
// 当热点 key 失效时，只有一个请求去查询数据库，其他请求等待
func (s *Service) GetWithMutex(ctx context.Context, key string, fetchFn func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// 1. 尝试从缓存获取
	result, err := s.Get(ctx, key)
	if err == nil {
		return result, nil
	}
	if errors.Is(err, ErrNullValue) {
		return nil, ErrNotFound
	}

	// 2. 尝试获取分布式锁
	lockKey := key + ":lock"
	acquired, err := s.SetNX(ctx, lockKey, LockValue, DefaultLockTTL)
	if err != nil {
		return nil, err
	}

	if acquired {
		// 成功获取锁
		defer s.Delete(ctx, lockKey)

		// 3. 双重检查：再次尝试从缓存获取
		result, err = s.Get(ctx, key)
		if err == nil {
			return result, nil
		}
		if errors.Is(err, ErrNullValue) {
			return nil, ErrNotFound
		}

		// 4. 从数据源获取
		data, err := fetchFn()
		if err != nil {
			return nil, err
		}

		// 5. 缓存结果
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

	// 6. 未获取锁，等待其他 goroutine 完成后重试
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

// GetMulti 批量获取
func (s *Service) GetMulti(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}

	// 使用 MGet 批量获取
	results := make(map[string]string)
	for _, key := range keys {
		val, err := s.Get(ctx, key)
		if err == nil {
			results[key] = val
		}
	}

	return results, nil
}

// SetMulti 批量设置
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

// Exists 检查 key 是否存在
func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Expire 设置过期时间
func (s *Service) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.Expire(ctx, key, ttl).Err()
}

// Client 获取底层 Redis 客户端
func (s *Service) Client() *RedisClient {
	return s.client
}

// IsNullMarker 检查是否为空值标记
func IsNullMarker(value string) bool {
	return value == NullMarker
}