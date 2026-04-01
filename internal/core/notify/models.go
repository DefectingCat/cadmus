// Package notify 提供邮件通知功能的实现。
//
// 该文件包含通知系统的核心数据模型，包括：
//   - 通知类型枚举定义
//   - 通知实体结构
//   - 通知渠道接口
//   - 各类型通知的数据结构
//
// 主要用途：
//
//	用于处理博客系统中的各类通知场景，如新评论提醒、回复通知、新文章发布等。
//
// 注意事项：
//   - NotificationChannel 接口由具体实现（如 EmailChannel）完成
//   - 通知数据结构用于模板渲染，需与邮件模板配合使用
//
// 作者：xfy
package notify

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType 通知类型枚举。
//
// 定义系统中支持的通知类型，每种类型对应不同的通知场景和模板。
type NotificationType string

// 通知类型常量定义。
const (
	// NotificationComment 新评论通知，当文章收到新评论时触发
	NotificationComment NotificationType = "comment"

	// NotificationReply 回复通知，当用户的评论被回复时触发
	NotificationReply NotificationType = "reply"

	// NotificationNewPost 新文章通知，用于订阅者接收新文章发布提醒
	NotificationNewPost NotificationType = "new_post"

	// NotificationSystem 系统通知，用于管理员发送的系统级公告
	NotificationSystem NotificationType = "system"
)

// Notification 通知实体。
//
// 表示一条待发送或已发送的通知记录，包含收件人、主题、内容等核心信息。
// 所有字段在创建后不可修改（不可变设计）。
//
// 注意事项：
//   - ID 由系统自动生成，无需手动设置
//   - CreatedAt 使用 UTC 时间
//   - Data 字段用于模板渲染，存储结构化数据
type Notification struct {
	// ID 通知的唯一标识符，格式为 UUID
	ID uuid.UUID `json:"id"`

	// Type 通知类型，决定使用的邮件模板
	Type NotificationType `json:"type"`

	// Recipient 收件人邮箱地址
	Recipient string `json:"recipient"`

	// Subject 邮件主题，显示在邮件标题中
	Subject string `json:"subject"`

	// Content 邮件正文内容，可以是纯文本或 HTML
	Content string `json:"content"`

	// Data 附加数据，用于模板渲染时的变量替换
	// 如评论通知中的文章标题、评论作者等
	Data map[string]any `json:"data"`

	// CreatedAt 创建时间，使用 UTC 时间戳
	CreatedAt time.Time `json:"created_at"`
}

// NotificationChannel 通知渠道接口。
//
// 该接口抽象了不同通知发送渠道的共性操作，支持：
//   - 邮件发送（SMTP）
//   - 其他通知服务（如 webhook、推送服务）
//
// 实现要求：
//   - Send 方法必须是并发安全的
//   - 错误返回应包含详细信息便于排查
type NotificationChannel interface {
	// Send 发送通知。
	//
	// 参数：
	//   - notification: 通知对象，包含收件人、主题、内容等信息
	//
	// 返回值：
	//   - err: 发送失败时返回错误，如网络问题、收件人无效等
	Send(notification *Notification) error
}

// CommentNotificationData 评论通知数据。
//
// 用于渲染评论通知邮件模板，包含文章和评论相关信息。
// 与 NotificationComment 类型配合使用。
type CommentNotificationData struct {
	// PostTitle 文章标题，用于邮件正文展示
	PostTitle string `json:"post_title"`

	// PostSlug 文章 Slug，用于构建文章链接
	PostSlug string `json:"post_slug"`

	// CommentAuthor 评论者名称
	CommentAuthor string `json:"comment_author"`

	// CommentContent 评论内容摘要
	CommentContent string `json:"comment_content"`
}

// ReplyNotificationData 回复通知数据。
//
// 用于渲染回复通知邮件模板，包含原评论和回复信息。
// 与 NotificationReply 类型配合使用。
type ReplyNotificationData struct {
	// PostTitle 文章标题
	PostTitle string `json:"post_title"`

	// PostSlug 文章 Slug
	PostSlug string `json:"post_slug"`

	// ReplyAuthor 回复者名称
	ReplyAuthor string `json:"reply_author"`

	// ReplyContent 回复内容
	ReplyContent string `json:"reply_content"`

	// ParentContent 原评论内容，用于上下文展示
	ParentContent string `json:"parent_content"`
}