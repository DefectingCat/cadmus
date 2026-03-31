package services

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"rua.plus/cadmus/internal/core/notify"
)

// EmailChannel 邮件通知渠道实现
type EmailChannel struct {
	config *notify.SMTPConfig
}

// NewEmailChannel 创建邮件渠道
func NewEmailChannel(config *notify.SMTPConfig) *EmailChannel {
	return &EmailChannel{
		config: config,
	}
}

// Send 发送邮件通知
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
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
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