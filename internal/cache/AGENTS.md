<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# Cache 目录 - Redis 缓存服务

## 目录结构

```
cache/
├── keys.go          # 缓存键常量定义和构建函数
├── redis.go         # Redis 客户端封装
└── service.go       # 缓存服务接口和实现
```

## 目录用途

`cache` 目录提供基于 Redis 的缓存服务，用于：

- 减轻数据库压力，提升系统响应速度
- 缓存文章、用户、角色、配置等热点数据
- 支持缓存穿透和缓存击穿防护
- 提供分布式锁功能

## 关键文件

### keys.go - 缓存键管理

统一管理所有缓存键的命名规范。

**功能：**
- 定义缓存命名空间 `cadmus`
- 提供各类缓存键常量（文章、用户、角色、配置等）
- 提供键构建辅助函数，避免手动拼接

**缓存键格式：**

| 键类型 | 格式 | 示例 |
|--------|------|------|
| 文章详情 | `cadmus:post:detail:{slug}:v{version}` | `cadmus:post:detail:my-article:v1` |
| 文章列表 | `cadmus:post:list:{category}:{page}:{sort}` | `cadmus:post:list:tech:1:created_at` |
| 用户信息 | `cadmus:user:info:{user_id}` | `cadmus:user:info:550e8400` |
| 用户权限 | `cadmus:user:perms:{user_id}:{permission}` | `cadmus:user:perms:550e8400:post.create` |
| 角色信息 | `cadmus:role:info:{role_id}` | `cadmus:role:info:550e8400` |
| 角色权限 | `cadmus:role:perms:{role_id}` | `cadmus:role:perms:550e8400` |
| 站点配置 | `cadmus:site:config` | `cadmus:site:config` |
| 主题配置 | `cadmus:theme:config:{theme_id}` | `cadmus:theme:config:default` |

### redis.go - Redis 客户端封装

封装 `go-redis` 库，提供类型安全的 Redis 访问接口。

**功能：**
- `Config` - Redis 连接配置结构体
- `RedisClient` - 封装底层客户端，提供便捷方法
- 支持连接池管理、超时配置、重试机制

**核心方法：**
- `Get/Set/Del` - 基础键值操作
- `SetNX` - 原子性设置（用于分布式锁）
- `HGet/HSet/HDel` - 哈希操作
- `LPush/RPush/LRange` - 列表操作
- `SAdd/SMembers/SIsMember` - 集合操作
- `Expire/TTL` - 过期时间管理

### service.go - 缓存服务实现

实现 `CacheService` 接口，提供高级缓存功能。

**功能：**
- 基础缓存操作：`Get`, `Set`, `Delete`, `SetNX`
- 批量操作：`GetMulti`, `SetMulti`
- 缓存穿透防护：`GetWithNullProtection`
- 缓存击穿防护：`GetWithMutex`

**常量：**
- `NullMarker` - 空值标记 `"__NULL__"`
- `NullMarkerTTL` - 空值标记过期时间 5 分钟
- `DefaultLockTTL` - 分布式锁过期时间 10 秒
- `LockMaxRetry` - 获取锁最大重试次数 50 次

## AI Agent 缓存使用指南

### 1. 获取缓存实例

```go
import "your-module/internal/cache"

// 创建 Redis 客户端
client, err := cache.NewRedisClient(cache.DefaultConfig())
if err != nil {
    log.Fatal("Redis 连接失败:", err)
}
defer client.Close()

// 创建缓存服务
svc := cache.NewService(client)
```

### 2. 自定义 Redis 配置

```go
cfg := cache.DefaultConfig()
cfg.Host = "redis.example.com"
cfg.Port = 6379
cfg.Password = "your-password"
cfg.PoolSize = 50
```

### 3. 基础缓存操作

```go
ctx := context.Background()

// 设置缓存
err := svc.Set(ctx, "my-key", "my-value", 10*time.Minute)

// 获取缓存
value, err := svc.Get(ctx, "my-key")
if errors.Is(err, cache.ErrNotFound) {
    // 键不存在
}

// 删除缓存
err := svc.Delete(ctx, "my-key")
```

### 4. 使用缓存键构建函数

```go
// 构建文章详情缓存键
key := cache.BuildPostDetailKey("my-article", 1)
// 返回："cadmus:post:detail:my-article:v1"

// 构建用户信息缓存键
key := cache.BuildUserInfoKey(userID)

// 构建角色权限缓存键
key := cache.BuildRolePermsKey(roleID)
```

### 5. 缓存穿透防护

当查询的数据可能不存在时，使用空值保护：

```go
data, err := svc.GetWithNullProtection(ctx, "user:123", func() (interface{}, error) {
    // 数据源查询逻辑
    return db.GetUser(ctx, 123)
}, 5*time.Minute)
```

### 6. 缓存击穿防护

高并发访问热点数据时，使用互斥锁：

```go
data, err := svc.GetWithMutex(ctx, "hot:article:123", func() (interface{}, error) {
    // 数据源查询逻辑
    return db.GetArticle(ctx, 123)
}, 10*time.Minute)
```

### 7. 批量操作

```go
// 批量获取
keys := []string{"key1", "key2", "key3"}
results, err := svc.GetMulti(ctx, keys)

// 批量设置
items := map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
}
err := svc.SetMulti(ctx, items, 10*time.Minute)
```

### 8. 使用空值标记判断

```go
value, _ := svc.Get(ctx, key)
if cache.IsNullMarker(value) {
    // 缓存的是空值标记，数据实际不存在
}
```

## 注意事项

1. **所有缓存键使用 `cadmus` 命名空间前缀**
2. **设置合理的 TTL**：避免数据永不过期
3. **使用防护机制**：根据场景选择穿透或击穿防护
4. **并发安全**：所有方法均支持并发调用
5. **错误处理**：检查 `ErrNotFound`, `ErrNullValue` 等特定错误

## 无子目录
