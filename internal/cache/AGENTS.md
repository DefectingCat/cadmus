<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# cache

## Purpose
Redis 缓存服务，提供热点数据缓存和权限缓存支持。

## Key Files
| File | Description |
|------|-------------|
| `redis.go` | Redis 客户端初始化和配置 |
| `service.go` | 缓存服务封装：Get/Set/Delete 操作 |
| `keys.go` | 缓存 Key 常量和构建函数 |

## For AI Agents

### Working In This Directory
- 使用 `go-redis/v9` 客户端
- Key 命名遵循 `cadmus:{entity}:{id}:{variant}` 格式
- 支持缓存穿透和击穿防护

### Cache Key Naming Convention
| Key Pattern | Description | TTL |
|-------------|-------------|-----|
| `cadmus:post:detail:{slug}:v{version}` | 文章详情 | 10 分钟 |
| `cadmus:post:list:{category}:{page}:{sort}` | 文章列表 | 5 分钟 |
| `cadmus:user:info:{user_id}` | 用户信息 | 30 分钟 |
| `cadmus:user:perms:{user_id}:{permission}` | 用户权限 | 1 小时 |
| `cadmus:role:info:{role_id}` | 角色信息 | 1 小时 |
| `cadmus:site:config` | 站点配置 | 1 小时 |

### Service Interface
```go
type Service interface {
    Get(ctx context.Context, key string) *redis.StringCmd
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration)
    Del(ctx context.Context, keys ...string)
    SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) bool
}
```

### Cache Invalidation
| Event | Action |
|-------|--------|
| 文章更新 | 删除 `post:detail:*` 和 `post:list:*` |
| 用户更新 | 删除 `user:info:{id}` |
| 角色权限变更 | 删除 `user:perms:*` 和 `role:info:*` |

<!-- MANUAL: -->