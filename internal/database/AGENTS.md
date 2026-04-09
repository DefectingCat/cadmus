<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# Database 数据层开发指南

## 目录用途

`internal/database` 目录实现 Repository 数据访问层，负责：

- PostgreSQL 数据库连接池管理
- 事务控制与自动错误处理
- 实体 CRUD 操作的统一封装
- 通用扫描工具与错误处理

所有数据访问通过 Repository 模式组织，服务层通过 Repository 接口与数据库交互。

---

## 关键文件及功能

### 连接配置

| 文件 | 用途 |
|------|------|
| `config.go` | 数据库连接配置结构体，支持 YAML 配置和 DSN 构建 |
| `connection.go` | 连接池封装，提供 `Pool` 类型和生命周期管理 |
| `transaction.go` | 事务管理器，回调式事务执行接口 |

### 通用工具

| 文件 | 用途 |
|------|------|
| `scanner.go` | 通用扫描工具 `ScanAll`/`ScanOne`，简化数据扫描 |
| `errors.go` | PostgreSQL 错误判断工具（唯一约束、外键冲突） |

### Repository 实现

| 文件 | 功能 |
|------|------|
| `user_repository.go` | 用户 CRUD、按邮箱/用户名查询、分页列表 |
| `post_repository.go` | 文章 CRUD、Slug 查询、分类/系列筛选、版本管理 |
| `comment_repository.go` | 评论 CRUD、文章关联查询 |
| `media_repository.go` | 媒体文件 CRUD |
| `category_repository.go` | 分类树 CRUD |
| `tag_repository.go` | 标签 CRUD、文章关联 |
| `series_repository.go` | 文章系列 CRUD |
| `role_repository.go` | 角色权限 CRUD |
| `permission_repository.go` | 权限 CRUD |
| `like_repository.go` | 点赞通用 Repository（BaseLikeRepository） |
| `search_repository.go` | 全文搜索实现 |

---

## 无子目录

所有文件位于 `internal/database/` 根目录。

---

## 数据层开发指南

### pgx 连接池使用

创建连接池：

```go
cfg := database.Config{
    Host:            "localhost",
    Port:            5432,
    Name:            "cadmus",
    User:            "cadmus",
    Password:        "secret",
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
    ConnMaxIdleTime: 10 * time.Minute,
    SSLMode:         "disable",
}

pool, err := database.NewPool(ctx, cfg)
if err != nil {
    return nil, fmt.Errorf("failed to create pool: %w", err)
}
defer pool.Close()
```

获取连接池统计：

```go
stats := pool.Stats()
// stats.MaxConns() - 最大连接数
// stats.AcquireCount() - 已获取连接数
```

### 事务处理

使用 `TransactionManager` 执行回调式事务：

```go
tm := database.NewTransactionManager(pool)

err := tm.WithTransaction(ctx, func(tx pgx.Tx) error {
    // 在事务中执行多个数据库操作
    if err := userRepo.Create(ctx, user); err != nil {
        return err // 自动回滚
    }
    
    if err := postRepo.Create(ctx, post); err != nil {
        return err // 自动回滚
    }
    
    return nil // 自动提交
})

if err != nil {
    // 事务回滚或提交失败
    return err
}
```

### 唯一约束错误处理

使用 `IsUniqueViolation` 判断唯一约束冲突：

```go
_, err := pool.Exec(ctx, query, args...)
if err != nil {
    if database.IsUniqueViolation(err, "users_username_key") {
        return user.ErrUserAlreadyExists
    }
    if database.IsUniqueViolation(err, "users_email_key") {
        return user.ErrUserAlreadyExists
    }
    return fmt.Errorf("failed to create user: %w", err)
}
```

### 数据扫描工具

单行扫描：

```go
user, err := database.ScanOne(row, func(r database.RowScanner) (*user.User, error) {
    u := &user.User{}
    err := r.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash)
    return u, err
})
```

多行扫描：

```go
rows, err := pool.Query(ctx, query, args...)
if err != nil {
    return nil, err
}

users, err := database.ScanAll(rows, func(r database.RowScanner) (*user.User, error) {
    u := &user.User{}
    err := r.Scan(&u.ID, &u.Username, &u.Email)
    return u, err
})
```

---

## 开发注意事项

1. **连接池安全**: `Pool` 是并发安全的，可在多个 goroutine 中共享
2. **资源释放**: 使用 `Query` 后必须 `defer rows.Close()`
3. **事务生命周期**: 事务回调中不应执行耗时操作，避免长事务
4. **错误包装**: 所有数据库错误应使用 `fmt.Errorf("...: %w", err)` 包装
5. **上下文传递**: 所有 Repository 方法需传入 `context.Context`
