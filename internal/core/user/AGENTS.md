<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# user

## Purpose
用户领域模型，定义用户、角色、权限实体和 Repository 接口。

## Key Files
| File | Description |
|------|-------------|
| `models.go` | User、Role、Permission 结构定义 |
| `repository.go` | UserRepository、RoleRepository 接口定义 |

## For AI Agents

### Working In This Directory
- 仅定义模型和接口，实现在 `internal/database/`
- 用户密码哈希使用 bcrypt，不在模型中处理

### Data Models
```go
type User struct {
    ID           uuid.UUID
    Username     string
    Email        string
    PasswordHash string
    AvatarURL    string
    Bio          string
    RoleID       uuid.UUID
    Status       UserStatus // active/banned/pending
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type Role struct {
    ID          uuid.UUID
    Name        string
    DisplayName string
    IsDefault   bool
    CreatedAt   time.Time
}

type Permission struct {
    ID          uuid.UUID
    Name        string    // post.create, comment.moderate 等
    Description string
    Category    string
}
```

### Repository Interfaces
```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
    Update(ctx context.Context, user *User) error
    List(ctx context.Context, offset, limit int) ([]*User, int, error)
}

type RoleRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
    GetByName(ctx context.Context, name string) (*Role, error)
    GetDefault(ctx context.Context) (*Role, error)
    List(ctx context.Context) ([]*Role, error)
}
```

### Permission Names Convention
| Pattern | Example | Description |
|---------|---------|-------------|
| `{entity}.{action}` | `post.create` | 创建文章 |
| `{entity}.{action}` | `comment.moderate` | 审核评论 |
| `{entity}.{action}` | `user.ban` | 封禁用户 |

<!-- MANUAL: -->