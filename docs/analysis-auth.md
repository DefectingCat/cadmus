# Cadmus 认证与权限系统分析报告

## 1. JWT 认证实现

### 1.1 核心组件

JWT 认证由 `internal/auth/` 包实现，包含以下核心文件：

| 文件 | 功能 |
|------|------|
| `jwt.go` | JWT Token 生成、验证、刷新 |
| `config.go` | JWT 配置管理 |
| `service.go` | 认证服务（登录、注册、登出） |
| `blacklist.go` | Token 黑名单（撤销机制） |
| `permission_cache.go` | 权限缓存服务 |

### 1.2 JWT Claims 结构

```go
type Claims struct {
    UserID uuid.UUID `json:"user_id"`
    RoleID uuid.UUID `json:"role_id"`
    jwt.RegisteredClaims  // 包含 jti, issuer, expiry 等
}
```

**关键特性：**
- 每个 Token 包含唯一 `jti` (JWT ID)，用于黑名单机制
- Token 承载 `UserID` 和 `RoleID`，避免每次请求查库
- 使用 HMAC-SHA256 (`HS256`) 签名算法

### 1.3 Token 生命周期

```
生成 → 使用 → 验证 → 刷新/登出
```

**生成 (`jwt.go:37-56`)：**
- 生成唯一 `jti` (UUID)
- 设置 `nbf` (Not Before) 防止提前使用
- 默认有效期：24 小时
- 刷新窗口：过期后 7 天内可刷新

**验证 (`jwt.go:59-77`)：**
- 验证签名算法为 HMAC
- 解析 Claims 并返回用户信息
- 自动检查 `exp`, `nbf`, `iss` 等标准 claims

**刷新 (`jwt.go:80-97`)：**
- 过期后 7 天内可刷新
- 生成新 Token 和新 jti

### 1.4 配置安全要求

```go
// config.go:19-34
func DefaultJWTConfig() (JWTConfig, error) {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        return JWTConfig{}, errors.New("JWT_SECRET environment variable is required")
    }
    if len(secret) < 32 {
        return JWTConfig{}, errors.New("JWT_SECRET must be at least 32 characters")
    }
    // ...
}
```

**安全措施：**
- 强制从环境变量读取密钥（禁止硬编码）
- 密钥最小长度 32 字符
- 提供 `MustJWTConfig()` 仅用于测试环境

### 1.5 Token 黑名单机制

```go
// blacklist.go
type TokenBlacklist interface {
    AddToBlacklist(ctx context.Context, tokenID string, expiry time.Time) error
    IsBlacklisted(ctx context.Context, tokenID string) bool
}
```

**实现特点：**
- 基于 Redis 存储，key 格式：`cadmus:jwt:blacklist:{jti}`
- TTL 自动过期（与 Token 过期时间一致）
- 避免永久存储已过期 Token

**登出流程 (`service.go:121-139`)：**
1. 验证 Token 获取 Claims
2. 提取 `jti` 和过期时间
3. 将 `jti` 加入黑名单，TTL = Token 剩余有效期

---

## 2. RBAC 权限模型

### 2.1 数据模型

```
User ──┬── Role ──┬── Permission
       │          │
       │          └── category (分组)
       │
       └── Status (active/banned/pending)
```

**核心实体 (`internal/core/user/models.go`)：**

| 实体 | 关键字段 |
|------|----------|
| `User` | ID, Username, Email, RoleID, Status |
| `Role` | ID, Name, DisplayName, Permissions[], IsDefault |
| `Permission` | ID, Name, Description, Category |

### 2.2 权限命名规范

采用 `resource.action` 格式：

```
post.create, post.edit, post.delete, post.publish, post.view
comment.create, comment.edit, comment.delete, comment.view
user.view, user.edit, user.delete, user.manage
theme.view, theme.edit, theme.install
plugin.view, plugin.edit, plugin.install
system.admin, system.settings
```

### 2.3 预置角色权限矩阵

| 角色 | 权限范围 | IsDefault |
|------|----------|-----------|
| admin | 全部权限 | false |
| editor | post.*, comment.* | false |
| author | post.create/edit/view, comment.*, user.view/edit | false |
| user | post.view, comment.view/create, user.view/edit | **true** |
| guest | post.view, comment.view, user.view | false |

### 2.4 数据库结构 (`migrations/001_init.up.sql`)

```sql
-- 多对多关联表
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 用户角色外键（RESTRICT 防止误删角色）
role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT
```

**索引优化：**
- `idx_permissions_category` - 按类别查询
- `idx_permissions_name` - 权限名快速查找
- `idx_roles_is_default` - 默认角色查询
- `idx_users_role` - 用户角色关联查询

### 2.5 Repository 实现

**PermissionRepository (`internal/database/permission_repository.go`)：**

| 方法 | 功能 |
|------|------|
| `GetByRoleID` | 获取角色所有权限 |
| `GetAll` | 获取所有权限 |
| `GetByCategory` | 按类别获取权限 |
| `CheckPermission` | 检查角色是否拥有指定权限 |

**RoleRepository (`internal/database/role_repository.go`)：**

| 方法 | 功能 |
|------|------|
| `GetByID` | 根据 ID 获取角色（不含权限） |
| `GetByName` | 根据名称获取角色（含权限） |
| `GetDefault` | 获取默认角色 |
| `GetWithPermissions` | 获取角色及权限 |
| `SetPermissions` | 设置角色权限（事务保护） |
| `Create/Delete` | 角色管理 |

---

## 3. 权限缓存策略

### 3.1 缓存架构

```
请求 → 中间件 → PermissionCache → Redis → 数据库
                   ↓
                缓存命中 → 直接返回
                   ↓
                缓存未命中 → 查库 → 写入缓存
```

### 3.2 缓存 Key 设计 (`internal/cache/keys.go`)

| Key 格式 | 用途 | TTL |
|----------|------|-----|
| `cadmus:user:perms:{roleID}:{permission}` | 单权限检查结果 | 1h |
| `cadmus:role:perms:{roleID}` | 角色权限列表 | 1h |
| `cadmus:jwt:blacklist:{jti}` | Token 黑名单 | Token 剩余有效期 |

### 3.3 缓存服务特性

**基础缓存 (`internal/cache/service.go`)：**
- `GetWithNullProtection` - 缓存穿透防护（缓存空值标记）
- `GetWithMutex` - 缓存击穿防护（分布式锁）
- 空值标记 TTL：5 分钟

**权限缓存 (`internal/auth/permission_cache.go`)：**

```go
const (
    PermissionCacheTTL = 1 * time.Hour
    RoleCacheTTL       = 1 * time.Hour
)
```

**缓存失效方法：**

| 方法 | 场景 |
|------|------|
| `InvalidateUserPermissions` | 用户权限变更 |
| `InvalidateRolePermissions` | 角色权限变更 |
| `InvalidateAllPermissions` | 全局权限重置 |

使用 Redis `SCAN` 命令批量删除匹配 key。

### 3.4 缓存流程详解

**GetPermission (`permission_cache.go:40-66`)：**
```
1. 尝试从 Redis 获取 cadmus:user:perms:{roleID}:{permission}
2. 命中 → 返回 bool 结果
3. 未命中 → 查询数据库 CheckPermission
4. 缓存结果 "true" 或 "false"
5. 缓存失败不影响业务（静默处理）
```

**GetRolePermissions (`permission_cache.go:69-100`)：**
```
1. 尝试获取 cadmus:role:perms:{roleID}
2. 命中 → JSON 反序列化返回
3. 反序列化失败 → 清除缓存重新查询
4. 未命中 → 查库并 JSON 序列化缓存
```

---

## 4. 认证中间件流程

### 4.1 中间件类型 (`internal/api/handlers/middleware.go`)

| 中间件 | 功能 | 适用场景 |
|--------|------|----------|
| `AuthMiddleware` | 基础 JWT 验证 | 一般认证 |
| `AuthMiddlewareWithBlacklist` | JWT + 黑名单检查 | 需要登出撤销 |
| `PermissionMiddleware` | 认证 + 权限（直查库） | 权限实时性要求高 |
| `CachedPermissionMiddleware` | 认证 + 权限（带缓存） | 一般权限检查 |
| `AdminMiddleware` | 管理员权限检查 | 管理后台 |
| `RequirePermissionMiddleware` | 通用权限检查 | 需先通过认证 |

### 4.2 认证流程

```
HTTP Request
    ↓
ExtractToken (Bearer header / Cookie)
    ↓
JWT Validate (签名 + Claims)
    ↓
[可选] Blacklist Check (jti)
    ↓
Context 注入 (UserID, RoleID)
    ↓
Handler
```

**Token 提取策略 (`middleware.go:175-193`)：**
```go
func ExtractToken(r *http.Request) string {
    // 1. Authorization: Bearer xxx
    // 2. Cookie: auth_token=xxx
}
```

### 4.3 权限检查流程

```
认证通过 → 获取 RoleID from Context
    ↓
PermissionCache.GetPermission(roleID, permission)
    ↓
[缓存命中] → 返回结果
    ↓
[缓存未命中] → DB CheckPermission → 写入缓存
    ↓
有权限 → Handler
无权限 → 403 Forbidden
```

### 4.4 Context Key 类型安全

```go
type ctxKey string  // 自定义类型避免冲突

const (
    ctxUserID   ctxKey = "user_id"
    ctxUserRole ctxKey = "user_role_id"
)
```

---

## 5. 安全性评估

### 5.1 安全措施清单

| 类别 | 措施 | 实现位置 |
|------|------|----------|
| **密钥安全** | 环境变量强制，禁止硬编码 | `config.go:20-23` |
| **密钥强度** | 最小 32 字符要求 | `config.go:24-26` |
| **签名安全** | 强制 HMAC 算法验证 | `jwt.go:61-64` |
| **Token 撤销** | JTI 黑名单机制 | `blacklist.go` |
| **密码存储** | bcrypt 哈希 | `service.go:99`, `models.go:93-100` |
| **用户状态** | banned 状态拒绝登录 | `service.go:65-67` |
| **错误模糊** | 统一 "invalid credentials" | `service.go:61,71` |

### 5.2 潜在风险与建议

| 风险 | 当前状态 | 建议 |
|------|----------|------|
| **密钥泄露** | 依赖环境变量 | 添加密钥轮换机制 |
| **Token 劫持** | 无 HTTPS 强制 | 生产环境强制 HTTPS |
| **刷新安全** | 过期后仍可刷新 | 考虑 refresh_token 独立机制 |
| **黑名单依赖** | 需要 Redis | 无 Redis 时黑名单失效 |
| **权限时效** | 缓存 1 小时 | 权限变更需主动失效缓存 |
| **密码复杂度** | 未验证 | 添加密码强度校验 |

### 5.3 OWASP Top 10 对照

| 漏洞类型 | 状态 |
|----------|------|
| A01 Broken Access Control | ✅ RBAC + 缓存双重保护 |
| A02 Cryptographic Failures | ✅ bcrypt 密码，HMAC JWT |
| A03 Injection | ⚠️ 权限名需参数化（已使用） |
| A07 Identification and Authentication | ✅ JWT + 黑名单 + 状态检查 |

### 5.4 生产部署建议

1. **强制 HTTPS** - Token 传输加密
2. **密钥轮换** - 定期更换 JWT_SECRET，旧 Token 自然过期
3. **监控黑名单** - 大量撤销可能表示攻击
4. **权限审计日志** - 记录权限检查失败事件
5. **Rate Limiting** - 登录接口防暴力破解
6. **CORS 配置** - 限制 Token 暴露范围

---

## 6. 架构图

### 6.1 认证流程图

```
┌─────────┐    ┌──────────┐    ┌────────────┐    ┌──────────┐
│ Client  │───▶│Middleware│───▶│ JWTService │───▶│ Handler  │
└─────────┘    └──────────┘    └────────────┘    └──────────┘
                    │                 │
                    │                 ▼
                    │         ┌────────────┐
                    │         │ Blacklist  │
                    │         │  (Redis)   │
                    │         └────────────┘
                    ▼
           ┌────────────────┐
           │ PermissionCache│
           │    (Redis)     │
           └────────────────┘
                    │
                    ▼
           ┌────────────────┐
           │  Repository    │
           │  (PostgreSQL)  │
           └────────────────┘
```

### 6.2 RBAC 模型图

```
┌──────────────────────────────────────────────────────────┐
│                      User                                │
│  ┌─────────────────────────────────────────────────────┐│
│  │ id, username, email, password_hash, role_id, status ││
│  └─────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────┘
                           │
                           │ N:1
                           ▼
┌──────────────────────────────────────────────────────────┐
│                      Role                                │
│  ┌─────────────────────────────────────────────────────┐│
│  │ id, name, display_name, is_default                  ││
│  └─────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────┘
                           │
                           │ N:M (role_permissions)
                           ▼
┌──────────────────────────────────────────────────────────┐
│                   Permission                             │
│  ┌─────────────────────────────────────────────────────┐│
│  │ id, name, description, category                     ││
│  └─────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────┘
```

---

## 7. 总结

Cadmus 认证与权限系统具备以下特点：

**优势：**
- JWT 无状态认证，服务端压力小
- JTI 黑名单实现 Token 撤销
- RBAC 模型清晰，支持细粒度权限
- Redis 缓存提升权限检查性能
- 缓存穿透/击穿防护机制完善
- 密码 bcrypt 存储，密钥环境变量管理

**待改进：**
- 缺少独立的 refresh_token 机制
- 无密码复杂度校验
- 黑名单依赖 Redis 可用性
- 缺少登录 Rate Limiting

**总体评价：** 系统设计合理，安全措施到位，适合中小型应用。生产部署需补充 HTTPS 强制、密钥轮换、登录限流等加固措施。