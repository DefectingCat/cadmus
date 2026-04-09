// Package cache 提供 Redis 客户端封装实现。
//
// 该文件包含 Redis 客户端的配置和封装，包括：
//   - Config: Redis 连接配置结构
//   - RedisClient: Redis 客户端封装
//   - 常用 Redis 操作的便捷方法
//
// 主要用途：
//
//	封装 go-redis 库，提供统一的 Redis 访问接口，
//	简化配置和连接管理。
//
// 注意事项：
//   - 使用连接池管理连接，提升性能
//   - 支持配置连接超时、读写超时等参数
//   - 创建时会测试连接，确保 Redis 可用
//
// 作者：xfy
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config Redis 连接配置。
//
// 包含 Redis 服务器的连接信息和连接池配置。
type Config struct {
	// Host Redis 服务器地址
	Host string `yaml:"host"`

	// Port Redis 服务器端口
	Port int `yaml:"port"`

	// Password Redis 访问密码，无密码时为空
	Password string `yaml:"password"`

	// DB Redis 数据库编号，默认为 0
	DB int `yaml:"db"`

	// 连接池配置

	// PoolSize 连接池大小，默认 25
	PoolSize int `yaml:"pool_size"`

	// MinIdleConns 最小空闲连接数，默认 5
	MinIdleConns int `yaml:"min_idle_conns"`

	// MaxRetries 命令最大重试次数，默认 3
	MaxRetries int `yaml:"max_retries"`

	// DialTimeout 连接建立超时时间
	DialTimeout time.Duration `yaml:"dial_timeout"`

	// ReadTimeout 读取超时时间
	ReadTimeout time.Duration `yaml:"read_timeout"`

	// WriteTimeout 写入超时时间
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// DefaultConfig 返回默认 Redis 配置。
//
// 提供一套适用于大多数场景的默认配置值。
// 可通过修改返回值的字段自定义配置。
//
// 返回值：
//   - config: 默认配置对象
//
// 使用示例：
//
//	cfg := DefaultConfig()
//	cfg.Host = "redis.example.com" // 自定义地址
//	cfg.Password = "secret"        // 设置密码
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

// Addr 构建 Redis 地址字符串。
//
// 将 Host 和 Port 组合成 "host:port" 格式。
//
// 返回值：
//   - Redis 地址字符串，如 "localhost:6379"
func (c Config) Addr() string {
	return c.Host + ":" + itoa(c.Port)
}

// itoa 简单的整数转字符串。
//
// 避免引入 strconv 包，实现简单的正整数转字符串。
// 仅用于端口号等正整数场景。
//
// 参数：
//   - n: 待转换的正整数
//
// 返回值：
//   - 转换后的字符串
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

// RedisClient Redis 客户端封装。
//
// 封装 go-redis 的 Client，提供类型安全的便捷方法。
// 所有方法直接转发到底层 Client，保持原有行为。
type RedisClient struct {
	// client 底层 go-redis 客户端
	client *redis.Client

	// config 配置信息，用于调试和日志
	config Config
}

// NewRedisClient 创建 Redis 客户端。
//
// 根据配置创建 Redis 客户端并测试连接。
// 创建成功后即可使用。
//
// 参数：
//   - cfg: Redis 连接配置
//
// 返回值：
//   - client: Redis 客户端实例
//   - err: 连接失败时返回错误
//
// 使用示例：
//
//	client, err := NewRedisClient(DefaultConfig())
//	if err != nil {
//	    log.Fatal("Redis 连接失败:", err)
//	}
//	defer client.Close()
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

// Client 获取底层 redis.Client。
//
// 用于执行更复杂的 Redis 操作，如事务、Lua 脚本等。
//
// 返回值：
//   - go-redis 客户端实例
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// Close 关闭连接。
//
// 释放连接池资源，应在程序退出时调用。
//
// 返回值：
//   - err: 关闭失败时返回错误
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Get 获取缓存值。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//
// 返回值：
//   - StringCmd: 命令结果，可通过 .Result() 获取值
func (r *RedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return r.client.Get(ctx, key)
}

// Set 设置缓存值。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 缓存值
//   - expiration: 过期时间
//
// 返回值：
//   - StatusCmd: 命令结果
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.client.Set(ctx, key, value, expiration)
}

// SetNX 仅当键不存在时设置。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 缓存值
//   - expiration: 过期时间
//
// 返回值：
//   - BoolCmd: 命令结果，true 表示设置成功
//
//nolint:staticcheck // SA1019 SetNX is deprecated in favor of Set with NX option, but we keep it for clarity
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return r.client.SetNX(ctx, key, value, expiration)
}

// Del 删除键。
//
// 参数：
//   - ctx: 上下文
//   - keys: 要删除的键列表
//
// 返回值：
//   - IntCmd: 命令结果，返回删除的键数量
func (r *RedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return r.client.Del(ctx, keys...)
}

// Exists 检查键是否存在。
//
// 参数：
//   - ctx: 上下文
//   - keys: 要检查的键列表
//
// 返回值：
//   - IntCmd: 命令结果，返回存在的键数量
func (r *RedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	return r.client.Exists(ctx, keys...)
}

// Expire 设置键的过期时间。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - expiration: 过期时间
//
// 返回值：
//   - BoolCmd: 命令结果，true 表示设置成功
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return r.client.Expire(ctx, key, expiration)
}

// TTL 获取键的剩余过期时间。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//
// 返回值：
//   - DurationCmd: 命令结果，返回剩余时间
func (r *RedisClient) TTL(ctx context.Context, key string) *redis.DurationCmd {
	return r.client.TTL(ctx, key)
}

// Incr 自增键的值。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//
// 返回值：
//   - IntCmd: 命令结果，返回自增后的值
func (r *RedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	return r.client.Incr(ctx, key)
}

// Decr 自减键的值。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//
// 返回值：
//   - IntCmd: 命令结果，返回自减后的值
func (r *RedisClient) Decr(ctx context.Context, key string) *redis.IntCmd {
	return r.client.Decr(ctx, key)
}

// HGet 获取哈希字段值。
//
// 参数：
//   - ctx: 上下文
//   - key: 哈希键
//   - field: 字段名
//
// 返回值：
//   - StringCmd: 命令结果
func (r *RedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return r.client.HGet(ctx, key, field)
}

// HSet 设置哈希字段值。
//
// 参数：
//   - ctx: 上下文
//   - key: 哈希键
//   - values: 字段值对
//
// 返回值：
//   - IntCmd: 命令结果
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.client.HSet(ctx, key, values...)
}

// HDel 删除哈希字段。
//
// 参数：
//   - ctx: 上下文
//   - key: 哈希键
//   - fields: 要删除的字段列表
//
// 返回值：
//   - IntCmd: 命令结果，返回删除的字段数量
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	return r.client.HDel(ctx, key, fields...)
}

// HGetAll 获取哈希所有字段。
//
// 参数：
//   - ctx: 上下文
//   - key: 哈希键
//
// 返回值：
//   - MapStringStringCmd: 命令结果，返回所有字段值对
func (r *RedisClient) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	return r.client.HGetAll(ctx, key)
}

// LPush 列表左侧插入元素。
//
// 参数：
//   - ctx: 上下文
//   - key: 列表键
//   - values: 要插入的值
//
// 返回值：
//   - IntCmd: 命令结果，返回列表长度
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.client.LPush(ctx, key, values...)
}

// RPush 列表右侧插入元素。
//
// 参数：
//   - ctx: 上下文
//   - key: 列表键
//   - values: 要插入的值
//
// 返回值：
//   - IntCmd: 命令结果，返回列表长度
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return r.client.RPush(ctx, key, values...)
}

// LRange 获取列表范围内的元素。
//
// 参数：
//   - ctx: 上下文
//   - key: 列表键
//   - start: 起始索引（0 开始）
//   - stop: 结束索引（-1 表示末尾）
//
// 返回值：
//   - StringSliceCmd: 命令结果，返回元素列表
func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return r.client.LRange(ctx, key, start, stop)
}

// SAdd 集合添加成员。
//
// 参数：
//   - ctx: 上下文
//   - key: 集合键
//   - members: 要添加的成员
//
// 返回值：
//   - IntCmd: 命令结果，返回添加的成员数量
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return r.client.SAdd(ctx, key, members...)
}

// SMembers 获取集合所有成员。
//
// 参数：
//   - ctx: 上下文
//   - key: 集合键
//
// 返回值：
//   - StringSliceCmd: 命令结果，返回所有成员
func (r *RedisClient) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return r.client.SMembers(ctx, key)
}

// SIsMember 检查成员是否在集合中。
//
// 参数：
//   - ctx: 上下文
//   - key: 集合键
//   - member: 要检查的成员
//
// 返回值：
//   - BoolCmd: 命令结果，true 表示成员存在
func (r *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	return r.client.SIsMember(ctx, key, member)
}

// Ping 测试连接。
//
// 发送 PING 命令测试与 Redis 服务器的连接。
//
// 参数：
//   - ctx: 上下文
//
// 返回值：
//   - StatusCmd: 命令结果，正常返回 "PONG"
func (r *RedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	return r.client.Ping(ctx)
}
