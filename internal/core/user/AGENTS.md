<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# user - 用户领域模型

## Purpose

`user` 目录包含 Cadmus 博客系统的用户、角色、权限管理的核心领域模型。该模块实现了基于角色的访问控制（RBAC）体系，支持多角色和细粒度权限控制。

主要用途：
- 定义用户实体及其生命周期状态
- 定义角色实体，支持多角色类型（管理员、编辑、作者、订阅者等）
- 定义权限实体，采用 "资源。操作" 格式的细粒度权限
- 提供密码加密存储方法（bcrypt 算法）
- 定义语义化错误类型，便于错误处理

## Files

| File | Purpose |
|------|---------|
| `models.go` | 领域实体定义（User、Role、Permission）、状态枚举、错误类型、密码加密方法 |
| `repository.go` | 数据访问接口定义（UserRepository、RoleRepository、PermissionRepository） |
| `models_test.go` | 单元测试：密码加密/验证、状态验证、错误类型测试 |

## Domain Models

### User 用户实体

表示博客系统中的用户账户，包含基本信息、角色关联和状态。

```go
type User struct {
    ID           uuid.UUID  // 用户唯一标识符（UUID）
    Username     string     // 用户名，用于登录和 URL 展示
    Email        string     // 邮箱地址，用于登录和通知
    PasswordHash string     // bcrypt 加密的密码哈希（不序列化到 JSON）
    AvatarURL    string     // 头像 URL（可选）
    Bio          string     // 个人简介（可选）
    RoleID       uuid.UUID  // 关联的角色 ID
    Status       UserStatus // 用户状态（active/banned/pending）
    CreatedAt    time.Time  // 创建时间（UTC）
    UpdatedAt    time.Time  // 最后修改时间（UTC）
}
```

### UserStatus 用户状态枚举

定义用户在系统中的生命周期状态：

| Status | Value | Description |
|--------|-------|-------------|
| StatusActive | "active" | 正常活跃用户，可以正常登录和使用系统 |
| StatusBanned | "banned" | 已封禁用户，禁止登录，保留数据用于审计 |
| StatusPending | "pending" | 待激活用户，已注册但未完成邮箱验证等激活流程 |

### Role 角色实体

表示用户的角色类型，每个角色关联一组权限：

```go
type Role struct {
    ID          uuid.UUID  // 角色唯一标识符
    Name        string     // 角色内部名称（如 "admin"、"editor"、"subscriber"）
    DisplayName string     // 角色显示名称（如 "管理员"、"编辑"、"订阅者"）
    Permissions []Permission // 角色拥有的权限列表
    IsDefault   bool       // 是否为默认角色（新注册用户自动分配）
    CreatedAt   time.Time  // 创建时间
}
```

### Permission 权限实体

表示系统中的细粒度权限，采用 "资源。操作" 格式：

```go
type Permission struct {
    ID          uuid.UUID  // 权限唯一标识符
    Name        string     // 权限名称（如 "post.create"、"comment.delete"）
    Description string     // 权限描述
    Category    string     // 权限分类（如 "post"、"comment"、"user"、"theme"、"plugin"）
    CreatedAt   time.Time  // 创建时间
}
```

### 错误类型

语义化错误定义，便于调用方进行错误处理：

| Error | Code | Message |
|-------|------|---------|
| ErrUserNotFound | user_not_found | 用户不存在 |
| ErrUserAlreadyExists | user_already_exists | 用户已存在 |
| ErrInvalidCredentials | invalid_credentials | 无效的凭证 |
| ErrInvalidStatus | invalid_status | 无效的用户状态 |
| ErrRoleNotFound | role_not_found | 角色不存在 |
| ErrPermissionDenied | permission_denied | 权限不足 |

### 密码加密方法

```go
// SetPassword 设置用户密码（使用 bcrypt 加密）
func (u *User) SetPassword(password string) error

// CheckPassword 验证用户密码
func (u *User) CheckPassword(password string) bool
```

## Repository Interfaces

### UserRepository

用户数据访问接口，定义用户实体的 CRUD 操作和查询方法：

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, offset, limit int) ([]*User, int, error)
}
```

### RoleRepository

角色数据访问接口，角色通常由系统初始化时创建，运行时主要是查询操作：

```go
type RoleRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*Role, error)           // 不含权限列表
    GetByName(ctx context.Context, name string) (*Role, error)          // 含权限列表
    GetAll(ctx context.Context) ([]*Role, error)
    GetDefault(ctx context.Context) (*Role, error)
    GetWithPermissions(ctx context.Context, id uuid.UUID) (*Role, error)
}
```

### PermissionRepository

权限数据访问接口，权限由系统定义，运行时主要是查询操作：

```go
type PermissionRepository interface {
    GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]Permission, error)
    GetAll(ctx context.Context) ([]Permission, error)
    GetByCategory(ctx context.Context, category string) ([]Permission, error)
    CheckPermission(ctx context.Context, roleID uuid.UUID, permissionName string) (bool, error)
}
```

## Subdirectories

无子目录。

## For AI Agents

### 开发指南

#### 1. 添加新用户状态

在 `models.go` 中添加新的状态常量，并更新 `IsValid()` 方法：

```go
const (
    StatusActive   UserStatus = "active"
    StatusBanned   UserStatus = "banned"
    StatusPending  UserStatus = "pending"
    StatusArchived UserStatus = "archived" // 新增状态
)

func (s UserStatus) IsValid() bool {
    switch s {
    case StatusActive, StatusBanned, StatusPending, StatusArchived: // 添加新状态
        return true
    default:
        return false
    }
}
```

#### 2. 添加新的错误类型

在 `models.go` 的错误定义区域添加：

```go
var (
    // 新增错误
    ErrUserLocked = &UserError{Code: "user_locked", Message: "用户已锁定"}
)
```

#### 3. 添加新的 Repository 方法

在 `repository.go` 中添加接口方法，确保：
- 所有方法接受 `context.Context` 作为第一个参数
- 使用 `uuid.UUID` 作为 ID 类型
- 返回预定义的语义化错误类型

```go
// 示例：添加按状态筛选用户的方法
GetByStatus(ctx context.Context, status UserStatus, offset, limit int) ([]*User, int, error)
```

#### 4. 编写单元测试

在 `models_test.go` 中测试新增功能：

```go
func TestUserStatus_IsValid(t *testing.T) {
    tests := []struct {
        status   UserStatus
        expected bool
    }{
        {StatusActive, true},
        {StatusBanned, true},
        {StatusPending, true},
        {UserStatus("invalid"), false},
    }

    for _, tt := range tests {
        t.Run(string(tt.status), func(t *testing.T) {
            assert.Equal(t, tt.expected, tt.status.IsValid())
        })
    }
}
```

### 实现注意事项

- **并发安全**: Repository 实现必须保证并发安全性
- **错误处理**: 所有错误必须使用 `models.go` 中定义的语义化错误类型
- **超时控制**: 所有接口方法都支持 `context.Context` 进行超时控制
- **密码安全**: 密码必须使用 `SetPassword()` 方法加密，禁止明文存储

### 跨模块引用

- `post`、`comment`、`media` 等模块通过 `uuid.UUID` 引用用户的 `ID` 字段
- 避免在 `user` 模块中直接引用其他业务模块，保持领域模型的独立性
- 业务逻辑应放在 `internal/service/` 层，Repository 只负责数据访问
