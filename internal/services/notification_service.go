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

// NotificationService 通知服务接口
type NotificationService interface {
	// SendCommentNotification 发送评论通知（通知文章作者）
	// postAuthor: 文章作者（收件人）
	// commentAuthor: 评论者
	SendCommentNotification(ctx context.Context, comment *comment.Comment, post *post.Post, postAuthor *user.User, commentAuthor *user.User) error

	// SendReplyNotification 发送回复通知（通知被回复用户）
	SendReplyNotification(ctx context.Context, reply *comment.Comment, parentComment *comment.Comment, post *post.Post, replyAuthor *user.User, parentAuthor *user.User) error

	// Send 发送通用通知
	Send(ctx context.Context, notification *notify.Notification) error
}

// notificationServiceImpl 通知服务实现
type notificationServiceImpl struct {
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