package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config Redis 连接配置
type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`

	// 连接池配置
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	MaxRetries   int           `yaml:"max_retries"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// DefaultConfig 返回默认 Redis 配置
func DefaultConfig() Config {
	return Config{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     25,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// RedisClient Redis 客户端封装
type RedisClient struct {
	client *redis.Client
	config Config
}

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(cfg Config) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{
		client: client,
		config: cfg,
	}, nil
}

// Addr 构建 Redis 地址
func (c Config) Addr() string {
	return c.Host + ":" + itoa(c.Port)
}

// itoa 简单的整数转字符串
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Client 获取底层 redis.Client
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// Close 关闭连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Get 获取缓存值
func (r *RedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return r.client.Get(ctx, key)
}

// Set 设置缓存值
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.client.Set(ctx, key, value, expiration)
}

// SetNX 仅当 key 不存在时设置
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return r.client.SetNX(ctx, key, value, expiration)
}

// Del 删除 key
func (r *RedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return r.client.Del(ctx, keys...)
}

// Exists 检查 key 是否存在
func (r *RedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	return r.client.Exists(ctx, keys...)
}

// Expire 设置 key 过期时间
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return r.client.Expire(ctx, key, expiration)
}

// TTL 获取 key 过期时间
func (r *RedisClient) TTL(ctx context.Context, key string) *redis.DurationCmd {
	return r.client.TTL(ctx, key)
}

// Incr 自增
func (r *RedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	return r.client.Incr(ctx, key)
}

// Decr 自减
func (r *RedisClient) Decr(ctx context.Context, key string) *redis.IntCmd {
	return r.client.Decr(ctx, key)
}

// HGet 获取哈希字段
func (r *RedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return r.client.HGet(ctx, key, field)
}

// HSet 设置哈希字段
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.client.HSet(ctx, key, values...)
}

// HDel 删除哈希字段
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	return r.client.HDel(ctx, key, fields...)
}

// HGetAll 获取所有哈希字段
func (r *RedisClient) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	return r.client.HGetAll(ctx, key)
}

// LPush 列表左推
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.client.LPush(ctx, key, values...)
}

// RPush 列表右推
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.client.RPush(ctx, key, values...)
}

// LRange 获取列表范围
func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return r.client.LRange(ctx, key, start, stop)
}

// SAdd 集合添加成员
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return r.client.SAdd(ctx, key, members...)
}

// SMembers 获取集合成员
func (r *RedisClient) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return r.client.SMembers(ctx, key)
}

// SIsMember 检查成员是否存在
func (r *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	return r.client.SIsMember(ctx, key, member)
}

// Ping 测试连接
func (r *RedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	return r.client.Ping(ctx)
}