<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# notify

## Purpose
通知领域模型，定义通知渠道接口和配置。

## Key Files
| File | Description |
|------|-------------|
| `models.go` | NotificationChannel 接口、Notification 结构 |
| `config.go` | 通知配置 |

## For AI Agents

### Working In This Directory
- 支持多种通知渠道：邮件、Discord、Slack 等
- 插件可扩展 NotificationChannel 接口

### Interfaces
```go
type NotificationChannel interface {
    Name() string
    Send(to string, subject string, body string) error
}

type Notification struct {
    To      string
    Subject string
    Body    string
    Type    NotificationType // email/webhook/push
}
```

### Notification Types
| Type | Use Case |
|------|----------|
| `comment` | 新评论通知 |
| `reply` | 回复通知 |
| `post` | 新文章通知 |
| `system` | 系统通知 |

### Built-in Channels
| Channel | Implementation |
|---------|----------------|
| Email | `internal/services/email_channel.go` |
| Discord | 插件扩展 |
| Slack | 插件扩展 |

### Email Configuration
```go
type EmailConfig struct {
    SMTPHost     string
    SMTPPort     int
    SMTPUser     string
    SMTPPassword string
    FromAddress  string
    FromName     string
}
```

<!-- MANUAL: -->