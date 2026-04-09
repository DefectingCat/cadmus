// Package services 提供邮件通知渠道的实现。
//
// 该文件包含邮件发送相关的核心逻辑，包括：
//   - SMTP 邮件发送
//   - 评论通知邮件模板渲染
//   - 回复通知邮件模板渲染
//
// 主要用途：
//
//	用于通过邮件向用户发送各类通知。
//
// 设计特点：
//   - 使用 Go 标准库 net/smtp 发送邮件
//   - 内置通知模板，支持自定义内容
//   - 支持 UTF-8 编码
//
// 作者：xfy
package services

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"rua.plus/cadmus/internal/core/notify"
)

// EmailChannel 邮件通知渠道实现。
//
// 该结构体实现了 NotificationChannel 接口，通过 SMTP 协议发送邮件。
// 支持评论通知和回复通知的模板渲染。
type EmailChannel struct {
	// config SMTP 配置（主机、端口、认证信息）
	config *notify.SMTPConfig
}

// NewEmailChannel 创建邮件渠道实例。
//
// 参数：
//   - config: SMTP 配置对象
//
// 返回值：
//   - *EmailChannel: 邮件渠道实例
func NewEmailChannel(config *notify.SMTPConfig) *EmailChannel {
	return &EmailChannel{
		config: config,
	}
}

// Send 发送邮件通知。
//
// 根据通知类型选择对应的邮件模板进行渲染，
// 然后通过 SMTP 协议发送邮件。
//
// 参数：
//   - notification: 通知对象，包含收件人、主题、类型和数据
//
// 返回值：
//   - error: 发送失败时返回错误
func (e *EmailChannel) Send(notification *notify.Notification) error {
	if e.config == nil {
		return fmt.Errorf("SMTP config is nil")
	}

	if notification.Recipient == "" {
		return fmt.Errorf("recipient email is empty")
	}

	// 构建邮件内容
	body, err := e.renderEmailBody(notification)
	if err != nil {
		return fmt.Errorf("failed to render email body: %w", err)
	}

	// 构建邮件消息
	message := e.buildMessage(notification, body)

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)
	auth := smtp.PlainAuth("", e.config.User, e.config.Password, e.config.Host)

	if err := smtp.SendMail(addr, auth, e.config.From, []string{notification.Recipient}, message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildMessage 构建邮件消息
func (e *EmailChannel) buildMessage(notification *notify.Notification, body string) []byte {
	headers := map[string]string{
		"From":         e.config.From,
		"To":           notification.Recipient,
		"Subject":      notification.Subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}

	message := bytes.NewBuffer(nil)
	for k, v := range headers {
		fmt.Fprintf(message, "%s: %s\r\n", k, v)
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	return message.Bytes()
}

// renderEmailBody 渲染邮件内容
func (e *EmailChannel) renderEmailBody(notification *notify.Notification) (string, error) {
	switch notification.Type {
	case notify.NotificationComment:
		return e.renderCommentNotification(notification)
	case notify.NotificationReply:
		return e.renderReplyNotification(notification)
	default:
		return notification.Content, nil
	}
}

// renderCommentNotification 渲染评论通知邮件
func (e *EmailChannel) renderCommentNotification(notification *notify.Notification) (string, error) {
	tmpl := `您好，

您的文章《{{.PostTitle}}》收到了一条新评论。

评论者：{{.CommentAuthor}}
评论内容：
{{.CommentContent}}

您可以点击以下链接查看文章：
{{.PostURL}}

---
此邮件由系统自动发送，请勿回复。`

	data, ok := notification.Data["comment"].(notify.CommentNotificationData)
	if !ok {
		return notification.Content, nil
	}

	postURL := fmt.Sprintf("/posts/%s", data.PostSlug)
	t, err := template.New("comment").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]string{
		"PostTitle":      data.PostTitle,
		"CommentAuthor":  data.CommentAuthor,
		"CommentContent": data.CommentContent,
		"PostURL":        postURL,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderReplyNotification 渲染回复通知邮件
func (e *EmailChannel) renderReplyNotification(notification *notify.Notification) (string, error) {
	tmpl := `您好，

您在文章《{{.PostTitle}}》下的评论收到了回复。

回复者：{{.ReplyAuthor}}
回复内容：
{{.ReplyContent}}

您的原评论：
{{.ParentContent}}

您可以点击以下链接查看文章：
{{.PostURL}}

---
此邮件由系统自动发送，请勿回复。`

	data, ok := notification.Data["reply"].(notify.ReplyNotificationData)
	if !ok {
		return notification.Content, nil
	}

	postURL := fmt.Sprintf("/posts/%s", data.PostSlug)
	t, err := template.New("reply").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]string{
		"PostTitle":     data.PostTitle,
		"ReplyAuthor":   data.ReplyAuthor,
		"ReplyContent":  data.ReplyContent,
		"ParentContent": data.ParentContent,
		"PostURL":       postURL,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
