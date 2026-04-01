// Package services 提供通知服务的实现。
//
// 该文件包含通知推送相关的核心逻辑，包括：
//   - 评论通知（通知文章作者）
//   - 回复通知（通知被回复用户）
//   - 通用通知发送
//
// 主要用途：
//
//	用于向用户推送各类通知，提升用户互动体验。
//
// 设计特点：
//   - 支持多种通知渠道（邮件、WebPush 等）
//   - 自动跳过自我通知（评论者即作者）
//   - 渠道可选配置（无渠道时静默跳过）
//
// 作者：xfy
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/internal/core/notify"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/core/user"
)

// NotificationService 通知服务接口。
//
// 该接口定义了通知推送的操作，支持评论通知和回复通知。
type NotificationService interface {
	// SendCommentNotification 发送评论通知给文章作者。
	//
	// 当用户对文章发表评论时，通知文章作者。
	// 如果评论者就是文章作者，则跳过通知。
	//
	// 参数：
	//   - ctx: 上下文
	//   - comment: 评论对象
	//   - post: 文章对象
	//   - postAuthor: 文章作者（收件人）
	//   - commentAuthor: 评论者（用于显示名称）
	//
	// 返回值：
	//   - error: 发送失败时返回错误
	SendCommentNotification(ctx context.Context, comment *comment.Comment, post *post.Post, postAuthor *user.User, commentAuthor *user.User) error

	// SendReplyNotification 发送回复通知给被回复用户。
	//
	// 当用户回复某条评论时，通知原评论作者。
	// 如果回复者就是被回复者，则跳过通知。
	//
	// 参数：
	//   - reply: 回复评论对象
	//   - parentComment: 被回复的评论
	//   - post: 文章对象
	//   - replyAuthor: 回复者
	//   - parentAuthor: 被回复者（收件人）
	SendReplyNotification(ctx context.Context, reply *comment.Comment, parentComment *comment.Comment, post *post.Post, replyAuthor *user.User, parentAuthor *user.User) error

	// Send 发送通用通知。
	//
	// 通过配置的通知渠道发送任意类型的通知。
	// 如果未配置渠道或收件人为空，则静默跳过。
	Send(ctx context.Context, notification *notify.Notification) error
}

// notificationServiceImpl 通知服务的具体实现。
type notificationServiceImpl struct {
	// channel 通知渠道（邮件、WebPush 等）
	channel notify.NotificationChannel
}

// NewNotificationService 创建通知服务
func NewNotificationService(channel notify.NotificationChannel) NotificationService {
	return &notificationServiceImpl{
		channel: channel,
	}
}

// SendCommentNotification 发送评论通知（通知文章作者）
func (s *notificationServiceImpl) SendCommentNotification(ctx context.Context, c *comment.Comment, p *post.Post, postAuthor *user.User, commentAuthor *user.User) error {
	// 如果评论者就是文章作者，不发送通知
	if c.UserID == p.AuthorID {
		return nil
	}

	// 如果没有文章作者信息，跳过通知
	if postAuthor == nil {
		return nil
	}

	// 获取评论者名称
	commentAuthorName := "匿名用户"
	if commentAuthor != nil {
		commentAuthorName = commentAuthor.Username
	}

	notification := &notify.Notification{
		ID:        uuid.New(),
		Type:      notify.NotificationComment,
		Recipient: postAuthor.Email,
		Subject:   fmt.Sprintf("您的文章《%s》收到了新评论", p.Title),
		Data: map[string]any{
			"comment": notify.CommentNotificationData{
				PostTitle:      p.Title,
				PostSlug:       p.Slug,
				CommentAuthor:  commentAuthorName,
				CommentContent: c.Content,
			},
		},
		CreatedAt: time.Now(),
	}

	return s.Send(ctx, notification)
}

// SendReplyNotification 发送回复通知（通知被回复用户）
func (s *notificationServiceImpl) SendReplyNotification(ctx context.Context, reply *comment.Comment, parentComment *comment.Comment, p *post.Post, replyAuthor *user.User, parentAuthor *user.User) error {
	// 如果回复者就是被回复用户，不发送通知
	if reply.UserID == parentComment.UserID {
		return nil
	}

	// 如果无法获取被回复用户信息，跳过通知
	if parentAuthor == nil {
		return nil
	}

	replyAuthorName := "匿名用户"
	if replyAuthor != nil {
		replyAuthorName = replyAuthor.Username
	}

	notification := &notify.Notification{
		ID:        uuid.New(),
		Type:      notify.NotificationReply,
		Recipient: parentAuthor.Email,
		Subject:   fmt.Sprintf("您的评论在文章《%s》中收到了回复", p.Title),
		Data: map[string]any{
			"reply": notify.ReplyNotificationData{
				PostTitle:     p.Title,
				PostSlug:      p.Slug,
				ReplyAuthor:   replyAuthorName,
				ReplyContent:  reply.Content,
				ParentContent: parentComment.Content,
			},
		},
		CreatedAt: time.Now(),
	}

	return s.Send(ctx, notification)
}

// Send 发送通用通知
func (s *notificationServiceImpl) Send(ctx context.Context, notification *notify.Notification) error {
	if s.channel == nil {
		// 如果没有配置通知渠道，静默跳过
		return nil
	}

	if notification.Recipient == "" {
		return nil
	}

	return s.channel.Send(notification)
}