// Package notify 邮件通知模块
package notify

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType 通知类型枚举
type NotificationType string

const (
	NotificationComment    NotificationType = "comment"    // 新评论通知
	NotificationReply      NotificationType = "reply"      // 回复通知
	NotificationNewPost    NotificationType = "new_post"   // 新文章通知
	NotificationSystem     NotificationType = "system"     // 系统通知
)

// Notification 通知实体
type Notification struct {
	ID        uuid.UUID       `json:"id"`
	Type      NotificationType `json:"type"`
	Recipient string          `json:"recipient"`  // 收件人邮箱
	Subject   string          `json:"subject"`    // 邮件主题
	Content   string          `json:"content"`    // 邮件内容
	Data      map[string]any  `json:"data"`       // 附加数据（用于模板渲染）
	CreatedAt time.Time       `json:"created_at"`
}

// NotificationChannel 通知渠道接口
type NotificationChannel interface {
	// Send 发送通知
	Send(notification *Notification) error
}

// CommentNotificationData 评论通知数据
type CommentNotificationData struct {
	PostTitle    string `json:"post_title"`
	PostSlug     string `json:"post_slug"`
	CommentAuthor string `json:"comment_author"`
	CommentContent string `json:"comment_content"`
}

// ReplyNotificationData 回复通知数据
type ReplyNotificationData struct {
	PostTitle     string `json:"post_title"`
	PostSlug      string `json:"post_slug"`
	ReplyAuthor   string `json:"reply_author"`
	ReplyContent  string `json:"reply_content"`
	ParentContent string `json:"parent_content"`
}