package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/internal/core/notify"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/core/user"
)

// MockNotificationChannel is a mock implementation of NotificationChannel
type MockNotificationChannel struct {
	mock.Mock
}

func (m *MockNotificationChannel) Send(n *notify.Notification) error {
	args := m.Called(n)
	return args.Error(0)
}

func TestNotificationService_SendCommentNotification(t *testing.T) {
	t.Run("sends notification when commenter is not post author", func(t *testing.T) {
		mockChannel := new(MockNotificationChannel)
		svc := NewNotificationService(mockChannel)

		postAuthorID := uuid.New()
		commenterID := uuid.New()
		p := &post.Post{
			ID:        uuid.New(),
			Title:     "Test Post",
			Slug:      "test-post",
			AuthorID:  postAuthorID,
			CreatedAt: time.Now(),
		}
		c := &comment.Comment{
			ID:      uuid.New(),
			PostID:  p.ID,
			UserID:  commenterID,
			Content: "Great article!",
		}
		postAuthor := &user.User{
			ID:       postAuthorID,
			Email:    "author@example.com",
			Username: "author",
		}
		commenter := &user.User{
			ID:       commenterID,
			Username: "commenter",
		}

		mockChannel.On("Send", mock.Anything).Return(nil)

		err := svc.SendCommentNotification(context.Background(), c, p, postAuthor, commenter)

		assert.NoError(t, err)
		mockChannel.AssertExpectations(t)
	})

	t.Run("skips notification when commenter is post author", func(t *testing.T) {
		mockChannel := new(MockNotificationChannel)
		svc := NewNotificationService(mockChannel)

		authorID := uuid.New()
		p := &post.Post{
			ID:       uuid.New(),
			Title:    "My Post",
			AuthorID: authorID,
		}
		c := &comment.Comment{
			ID:     uuid.New(),
			PostID: p.ID,
			UserID: authorID, // Same as post author
		}
		author := &user.User{ID: authorID, Email: "author@example.com"}

		err := svc.SendCommentNotification(context.Background(), c, p, author, author)

		assert.NoError(t, err)
		mockChannel.AssertNotCalled(t, "Send")
	})

	t.Run("skips notification when post author is nil", func(t *testing.T) {
		mockChannel := new(MockNotificationChannel)
		svc := NewNotificationService(mockChannel)

		p := &post.Post{ID: uuid.New(), AuthorID: uuid.New()}
		c := &comment.Comment{ID: uuid.New(), UserID: uuid.New()}

		err := svc.SendCommentNotification(context.Background(), c, p, nil, nil)

		assert.NoError(t, err)
		mockChannel.AssertNotCalled(t, "Send")
	})

	t.Run("uses anonymous name when commenter is nil", func(t *testing.T) {
		mockChannel := new(MockNotificationChannel)
		svc := NewNotificationService(mockChannel)

		postAuthorID := uuid.New()
		commenterID := uuid.New()
		p := &post.Post{ID: uuid.New(), Title: "Test", AuthorID: postAuthorID}
		c := &comment.Comment{ID: uuid.New(), PostID: p.ID, UserID: commenterID, Content: "Hi"}
		postAuthor := &user.User{ID: postAuthorID, Email: "author@example.com"}

		mockChannel.On("Send", mock.MatchedBy(func(n *notify.Notification) bool {
			data, ok := n.Data["comment"].(notify.CommentNotificationData)
			return ok && data.CommentAuthor == "匿名用户"
		})).Return(nil)

		err := svc.SendCommentNotification(context.Background(), c, p, postAuthor, nil)

		assert.NoError(t, err)
		mockChannel.AssertExpectations(t)
	})
}

func TestNotificationService_SendReplyNotification(t *testing.T) {
	t.Run("sends notification when replier is not parent author", func(t *testing.T) {
		mockChannel := new(MockNotificationChannel)
		svc := NewNotificationService(mockChannel)

		parentAuthorID := uuid.New()
		replierID := uuid.New()
		p := &post.Post{ID: uuid.New(), Title: "Test Post", Slug: "test"}
		parent := &comment.Comment{
			ID:     uuid.New(),
			UserID: parentAuthorID,
			Content: "Original comment",
		}
		reply := &comment.Comment{
			ID:      uuid.New(),
			UserID:  replierID,
			Content: "Reply content",
		}
		parentAuthor := &user.User{ID: parentAuthorID, Email: "parent@example.com"}
		replier := &user.User{ID: replierID, Username: "replier"}

		mockChannel.On("Send", mock.Anything).Return(nil)

		err := svc.SendReplyNotification(context.Background(), reply, parent, p, replier, parentAuthor)

		assert.NoError(t, err)
		mockChannel.AssertExpectations(t)
	})

	t.Run("skips notification when replier is parent author", func(t *testing.T) {
		mockChannel := new(MockNotificationChannel)
		svc := NewNotificationService(mockChannel)

		authorID := uuid.New()
		p := &post.Post{ID: uuid.New()}
		parent := &comment.Comment{ID: uuid.New(), UserID: authorID}
		reply := &comment.Comment{ID: uuid.New(), UserID: authorID}
		author := &user.User{ID: authorID}

		err := svc.SendReplyNotification(context.Background(), reply, parent, p, author, author)

		assert.NoError(t, err)
		mockChannel.AssertNotCalled(t, "Send")
	})

	t.Run("skips notification when parent author is nil", func(t *testing.T) {
		mockChannel := new(MockNotificationChannel)
		svc := NewNotificationService(mockChannel)

		p := &post.Post{ID: uuid.New()}
		parent := &comment.Comment{ID: uuid.New(), UserID: uuid.New()}
		reply := &comment.Comment{ID: uuid.New(), UserID: uuid.New()}

		err := svc.SendReplyNotification(context.Background(), reply, parent, p, nil, nil)

		assert.NoError(t, err)
		mockChannel.AssertNotCalled(t, "Send")
	})
}

func TestNotificationService_Send(t *testing.T) {
	t.Run("sends notification with channel", func(t *testing.T) {
		mockChannel := new(MockNotificationChannel)
		svc := NewNotificationService(mockChannel)

		n := &notify.Notification{
			ID:        uuid.New(),
			Type:      notify.NotificationComment,
			Recipient: "user@example.com",
			Subject:   "Test",
		}

		mockChannel.On("Send", n).Return(nil)

		err := svc.Send(context.Background(), n)

		assert.NoError(t, err)
		mockChannel.AssertExpectations(t)
	})

	t.Run("skips when channel is nil", func(t *testing.T) {
		svc := NewNotificationService(nil)

		n := &notify.Notification{Recipient: "user@example.com"}

		err := svc.Send(context.Background(), n)

		assert.NoError(t, err)
	})

	t.Run("skips when recipient is empty", func(t *testing.T) {
		mockChannel := new(MockNotificationChannel)
		svc := NewNotificationService(mockChannel)

		n := &notify.Notification{Recipient: ""}

		err := svc.Send(context.Background(), n)

		assert.NoError(t, err)
		mockChannel.AssertNotCalled(t, "Send")
	})

	t.Run("returns error from channel", func(t *testing.T) {
		mockChannel := new(MockNotificationChannel)
		svc := NewNotificationService(mockChannel)

		n := &notify.Notification{Recipient: "user@example.com"}
		mockChannel.On("Send", n).Return(assert.AnError)

		err := svc.Send(context.Background(), n)

		assert.Error(t, err)
		mockChannel.AssertExpectations(t)
	})
}