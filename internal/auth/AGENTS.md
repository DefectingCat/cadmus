<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# auth

## Purpose

认证模块，提供 JWT 令牌签发与验证、Token 黑名单管理、权限缓存三大核心功能。是 Cadmus 应用认证和授权系统的基础组件。

## Key Files

| File | Description |
|------|-------------|
| `config.go` | JWT 配置：从环境变量读取密钥，定义过期时间、签发者等参数 |
| `jwt.go` | JWT 服务：令牌生成、验证、刷新，使用 HMAC-SHA256 签名 |
| `blacklist.go` | Token 黑名单：基于 Redis 实现，管理已注销的令牌 |
| `permission_cache.go` | 权限缓存：缓存角色权限查询结果，减少数据库压力 |
| `service.go` | 认证服务：登录、注册、登出、Token 验证的完整流程 |

## No Subdirectories

本目录无子目录。

## For AI Agents

### 认证开发指南

#### 1. JWT 配置

JWT 密钥必须从环境变量获取，禁止硬编码：

```bash
export JWT_SECRET="your-secret-key-at-least-32-chars"
```

```go
// 获取配置（生产环境）
config, err := auth.DefaultJWTConfig()
if err != nil {
    // 处理配置错误
}

// 创建 JWT 服务
jwtSvc := auth.NewJWTService(config)
```

**配置项**：
- `Secret`: 签名密钥（HMAC-SHA256），至少 32 字符
- `Expiry`: Token 有效期，默认 24 小时
- `RefreshExpiry`: 刷新窗口期，默认 7 天
- `Issuer`: 签发者标识，默认 "cadmus"

#### 2. JWT Token Structure

```json
{
  "user_id": "uuid-string",
  "role_id": "uuid-string",
  "iat": 1234567890,
  "exp": 1234567890,
  "nbf": 1234567890,
  "iss": "cadmus",
  "jti": "unique-jwt-id"
}
```

**Claims 说明**：
- `user_id`: 用户唯一标识符（UUID）
- `role_id`: 用户角色标识符（UUID）
- `jti`: JWT 唯一 ID，用于黑名单机制
- `iat`: 签发时间
- `exp`: 过期时间
- `nbf`: 生效时间
- `iss`: 签发者

#### 3. Token 黑名单（Redis）

**使用场景**：用户登出时将 Token 加入黑名单，使已注销的令牌无法再次使用。

```go
// 创建黑名单服务
blacklist := auth.NewRedisTokenBlacklist(redisClient)

// 登出时将 token 加入黑名单
claims, _ := jwtSvc.Validate(tokenString)
jti := claims.GetJWTID()
err := blacklist.AddToBlacklist(ctx, jti, claims.ExpiresAt.Time)

// 检查 token 是否在黑名单中
if blacklist.IsBlacklisted(ctx, jti) {
    // 拒绝请求
}
```

**Redis Key 格式**：
- `cadmus:jwt:blacklist:{jti}` - 黑名单令牌存储

**TTL**：黑名单记录的 TTL 等于 Token 的剩余有效期，过期后自动删除。

#### 4. Permission Cache（权限缓存）

**使用场景**：缓存角色权限查询结果，减少数据库访问，提升权限验证性能。

```go
// 创建权限缓存服务
permCache := auth.NewPermissionCache(cacheService, permRepo, redisClient)

// 检查权限（带缓存）
hasPerm, err := permCache.GetPermission(ctx, roleID, "user:read")

// 获取角色所有权限（带缓存）
perms, err := permCache.GetRolePermissions(ctx, roleID)

// 权限变更后清除缓存
permCache.InvalidateRolePermissions(ctx, roleID)
permCache.InvalidateUserPermissions(ctx, userID)
```

**Redis Key 格式**：
| Key Pattern | Description | TTL |
|-------------|-------------|-----|
| `cadmus:user:perms:{userID}:{permission}` | 用户权限缓存 | 1 小时 |
| `cadmus:role:perms:{roleID}` | 角色权限列表（JSON） | 1 小时 |

**缓存失效策略**：
- 角色权限变更时调用 `InvalidateRolePermissions`
- 用户角色变更时调用 `InvalidateUserPermissions`
- 系统重置时调用 `InvalidateAllPermissions`

#### 5. 认证服务完整流程

```go
// 创建认证服务
authSvc := auth.NewAuthService(jwtSvc, userRepo)

// 启用黑名单功能（可选）
authSvc = authSvc.WithBlacklist(blacklist)

// 用户登录
result, err := authSvc.Login(ctx, email, password)
// result.Token 用于后续请求认证
// result.User 包含用户信息

// 用户注册
user, err := authSvc.Register(ctx, username, email, password)
// 新用户默认状态为 StatusPending

// 用户登出
err := authSvc.Logout(ctx, tokenString)
// 将 token 加入黑名单

// 验证 token（中间件使用）
claims, user, err := authSvc.ValidateToken(ctx, tokenString)
// 完整验证流程：黑名单检查 → 签名验证 → 用户状态检查
```

### 关键函数参考

| Function | Description |
|----------|-------------|
| `DefaultJWTConfig()` | 从环境变量读取 JWT 配置 |
| `NewJWTService(config)` | 创建 JWT 服务 |
| `jwtSvc.Generate(userID, roleID)` | 生成 JWT token，返回 `(token, jti, error)` |
| `jwtSvc.Validate(token)` | 验证并解析 Claims |
| `jwtSvc.Refresh(token)` | 刷新过期 token |
| `blacklist.AddToBlacklist(ctx, jti, expiry)` | 将 token 加入黑名单 |
| `blacklist.IsBlacklisted(ctx, jti)` | 检查 token 是否在黑名单中 |
| `permCache.GetPermission(ctx, roleID, permission)` | 检查权限（带缓存） |
| `authSvc.Login(ctx, email, password)` | 用户登录 |
| `authSvc.Logout(ctx, token)` | 用户登出 |
| `authSvc.ValidateToken(ctx, token)` | 完整 token 验证 |

### 注意事项

1. **安全性**：
   - JWT_SECRET 必须从环境变量或密钥管理服务获取
   - 密钥长度建议 64 字符以上
   - 生产环境应定期轮换密钥

2. **性能**：
   - 权限缓存 TTL 为 1 小时，权限变更后需主动失效
   - 黑名单检查会增加每次请求的 Redis 查询开销
   - 确保 Redis 服务高可用

3. **并发安全**：
   - 所有公开方法均为并发安全
   - Redis 客户端需配置合适的连接池
