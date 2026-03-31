<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# auth

## Purpose
JWT 认证模块，处理令牌生成、验证、黑名单和权限缓存。

## Key Files
| File | Description |
|------|-------------|
| `jwt.go` | JWT 服务：令牌生成、验证、刷新 |
| `blacklist.go` | Token 黑名单：Redis 实现的注销令牌管理 |
| `permission_cache.go` | 权限缓存：Redis 实现的权限检查缓存 |
| `service.go` | 认证服务：登录、注册、密码验证 |
| `config.go` | JWT 配置：密钥管理、过期时间 |

## For AI Agents

### Working In This Directory
- JWT 密钥从环境变量 `JWT_SECRET` 或文件读取
- Token 黑名单使用 Redis SET 存储
- 权限缓存 TTL 默认 1 小时

### JWT Token Structure
```json
{
    "user_id": "uuid",
    "role_id": "uuid",
    "exp": 1234567890,
    "iat": 1234567890
}
```

### Redis Keys
| Key Pattern | Description | TTL |
|-------------|-------------|-----|
| `cadmus:token:blacklist:{token_hash}` | 黑名单令牌 | Token 剩余有效期 |
| `cadmus:user:perms:{user_id}:{permission}` | 用户权限缓存 | 1 小时 |
| `cadmus:role:info:{role_id}` | 角色信息缓存 | 1 小时 |

### Key Functions
| Function | Description |
|----------|-------------|
| `GenerateToken(userID, roleID)` | 生成 JWT |
| `Validate(token)` | 验证并解析 Claims |
| `Refresh(token)` | 刷新令牌 |
| `IsBlacklisted(token)` | 检查黑名单 |
| `AddToBlacklist(token)` | 添加到黑名单 |
| `HasPermission(ctx, userID, permission)` | 检查权限 |

<!-- MANUAL: -->