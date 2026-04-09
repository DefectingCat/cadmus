<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# notify - 通知领域模型

## Purpose

`notify` 目录包含 Cadmus 系统的通知领域模型定义。该模块负责：

- 定义通知系统的核心实体和类型
- 抽象通知发送渠道接口
- 提供各类通知场景的数据结构
- 定义 SMTP 配置结构

该目录**不包含**具体实现，仅定义接口和类型。实现位于 `internal/service/`（业务逻辑）或 `internal/database/`（持久化）。

## Key Files

| File | Purpose |
|------|---------|
| `models.go` | 通知实体、通知类型枚举、渠道接口定义 |
| `config.go` | SMTP 配置结构定义 |

## Domain Models

### NotificationType 枚举

定义系统支持的四种通知类型：

| Type | Value | Description |
|------|-------|-------------|
| `NotificationComment` | `"comment"` | 新评论通知，文章收到评论时触发 |
| `NotificationReply` | `"reply"` | 回复通知，用户的评论被回复时触发 |
| `NotificationNewPost` | `"new_post"` | 新文章通知，用于订阅者接收发布提醒 |
| `NotificationSystem` | `"system"` | 系统通知，管理员发送的系统级公告 |

### Notification 实体

核心通知实体，表示一条待发送或已发送的通知记录：

```go
type Notification struct {
    ID        uuid.UUID            // 唯一标识符
    Type      NotificationType     // 通知类型
    Recipient string               // 收件人邮箱
    Subject   string               // 邮件主题
    Content   string               // 邮件正文
    Data      map[string]any       // 附加数据，用于模板渲染
    CreatedAt time.Time            // 创建时间 (UTC)
}
```

**设计说明：**
- ID 由系统自动生成，无需手动设置
- CreatedAt 使用 UTC 时间
- Data 字段用于模板渲染，存储结构化数据
- 所有字段在创建后不可修改（不可变设计）

### NotificationChannel 接口

通知渠道抽象接口，支持多种发送方式：

```go
type NotificationChannel interface {
    Send(notification *Notification) error
}
```

**实现要求：**
- `Send` 方法必须是并发安全的
- 错误返回应包含详细信息

### 通知数据结构

用于模板渲染的结构化数据：

| Struct | Used By | Fields |
|--------|---------|--------|
| `CommentNotificationData` | `NotificationComment` | `PostTitle`, `PostSlug`, `CommentAuthor`, `CommentContent` |
| `ReplyNotificationData` | `NotificationReply` | `PostTitle`, `PostSlug`, `ReplyAuthor`, `ReplyContent`, `ParentContent` |

### SMTP 配置

邮件通知的基础配置：

```go
type SMTPConfig struct {
    Host     string // SMTP 服务器地址
    Port     int    // SMTP 服务器端口
    User     string // SMTP 用户名
    Password string // SMTP 密码
    From     string // 发件人地址
    UseTLS   bool   // 是否使用 TLS
}
```

## No Subdirectories

当前 `notify` 模块无子目录。

## For AI Agents

### 添加新通知类型

1. 在 `models.go` 中添加新的 `NotificationType` 常量：
   ```go
   const (
       // NotificationNewType 新类型说明
       NotificationNewType NotificationType = "new_type"
   )
   ```

2. 创建对应的数据结构（如需要）：
   ```go
   type NewTypeNotificationData struct {
       // 字段定义
   }
   ```

3. 更新此文档的枚举表格

### 实现通知渠道

实现 `NotificationChannel` 接口：

```go
type EmailNotificationChannel struct {
    config *SMTPConfig
}

func (c *EmailNotificationChannel) Send(notification *Notification) error {
    // 实现发送逻辑
    return nil
}
```

### 扩展 Repository 接口

当需要持久化通知记录时，创建 `repository.go`：

```go
type NotificationRepository interface {
    Create(ctx context.Context, notification *Notification) error
    GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
    List(ctx context.Context, filters NotificationFilters, offset, limit int) ([]*Notification, int, error)
    // ... 其他方法
}
```

### 跨模块依赖

- `notify` 可引用 `user`、`post`、`comment` 等模块的类型
- 避免循环依赖：不要在 `user` 等模块中直接引用 `notify`
