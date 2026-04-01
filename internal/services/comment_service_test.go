package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/test/mocks"
)

// ============================================================
// CommentService Tests
// ============================================================

func createTestComment(content string, status comment.CommentStatus, depth int) *comment.Comment {
	return &comment.Comment{
		ID:        uuid.New(),
		PostID:    uuid.New(),
		UserID:    uuid.New(),
		Content:   content,
		Status:    status,
		Depth:     depth,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestCommentService_CreateComment(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		input := &comment.CreateCommentInput{
			PostID:  uuid.New(),
			UserID:  uuid.New(),
			Content: "This is a test comment",
		}
		expectedComment := &comment.Comment{
			ID:      uuid.New(),
			PostID:  input.PostID,
			UserID:  input.UserID,
			Content: input.Content,
			Status:  comment.StatusPending,
			Depth:   0,
		}

		mockCommentRepo.On("Create", mock.Anything, input).Return(expectedComment, nil)

		result, err := svc.CreateComment(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, expectedComment, result)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("empty content returns error", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		input := &comment.CreateCommentInput{
			PostID:  uuid.New(),
			UserID:  uuid.New(),
			Content: "",
		}

		_, err := svc.CreateComment(context.Background(), input)

		assert.Error(t, err)
		assert.ErrorIs(t, err, comment.ErrEmptyContent)
	})

	t.Run("parent not found returns error", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		parentID := uuid.New()
		input := &comment.CreateCommentInput{
			PostID:   uuid.New(),
			UserID:   uuid.New(),
			ParentID: &parentID,
			Content:  "This is a reply",
		}

		mockCommentRepo.On("GetByID", mock.Anything, parentID).Return(nil, comment.ErrCommentNotFound)

		_, err := svc.CreateComment(context.Background(), input)

		assert.Error(t, err)
		assert.ErrorIs(t, err, comment.ErrParentNotFound)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("max depth exceeded returns error", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		parentID := uuid.New()
		input := &comment.CreateCommentInput{
			PostID:   uuid.New(),
			UserID:   uuid.New(),
			ParentID: &parentID,
			Content:  "This is a reply",
		}

		parentComment := createTestComment("parent", comment.StatusApproved, MaxCommentDepth)

		mockCommentRepo.On("GetByID", mock.Anything, parentID).Return(parentComment, nil)

		_, err := svc.CreateComment(context.Background(), input)

		assert.Error(t, err)
		assert.ErrorIs(t, err, comment.ErrMaxDepthExceeded)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("reply increments depth", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		parentID := uuid.New()
		input := &comment.CreateCommentInput{
			PostID:   uuid.New(),
			UserID:   uuid.New(),
			ParentID: &parentID,
			Content:  "This is a reply",
		}

		parentComment := createTestComment("parent", comment.StatusApproved, 2)
		parentComment.ID = parentID

		expectedComment := &comment.Comment{
			ID:      uuid.New(),
			PostID:  input.PostID,
			UserID:  input.UserID,
			Content: input.Content,
			Status:  comment.StatusPending,
			Depth:   3,
		}

		mockCommentRepo.On("GetByID", mock.Anything, parentID).Return(parentComment, nil)
		mockCommentRepo.On("Create", mock.Anything, input).Return(expectedComment, nil)

		result, err := svc.CreateComment(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, expectedComment, result)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_GetCommentByID(t *testing.T) {
	t.Run("returns comment when found", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusApproved, 0)
		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)

		result, err := svc.GetCommentByID(context.Background(), testComment.ID)

		assert.NoError(t, err)
		assert.Equal(t, testComment, result)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_GetCommentsByPost(t *testing.T) {
	t.Run("returns comment tree", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		postID := uuid.New()
		userID := uuid.New()

		// Create a tree: root -> child1, child2
		root := createTestComment("root", comment.StatusApproved, 0)
		root.PostID = postID
		root.UserID = userID

		child1 := createTestComment("child1", comment.StatusApproved, 1)
		child1.PostID = postID
		child1.UserID = userID
		child1.ParentID = &root.ID

		child2 := createTestComment("child2", comment.StatusApproved, 1)
		child2.PostID = postID
		child2.UserID = userID
		child2.ParentID = &root.ID

		comments := []*comment.Comment{root, child1, child2}

		mockCommentRepo.On("GetByPostID", mock.Anything, postID, mock.Anything).Return(comments, nil)

		result, err := svc.GetCommentsByPost(context.Background(), postID)

		assert.NoError(t, err)
		assert.Len(t, result, 1) // One root node
		assert.Len(t, result[0].Children, 2) // Two children
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_ApproveComment(t *testing.T) {
	t.Run("approves pending comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusPending, 0)
		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)
		mockCommentRepo.On("UpdateStatus", mock.Anything, testComment.ID, comment.StatusApproved).Return(nil)

		err := svc.ApproveComment(context.Background(), testComment.ID)

		assert.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("cannot approve deleted comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusDeleted, 0)
		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)

		err := svc.ApproveComment(context.Background(), testComment.ID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "已删除")
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("comment not found returns error", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		id := uuid.New()
		mockCommentRepo.On("GetByID", mock.Anything, id).Return(nil, comment.ErrCommentNotFound)

		err := svc.ApproveComment(context.Background(), id)

		assert.Error(t, err)
		assert.ErrorIs(t, err, comment.ErrCommentNotFound)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_RejectComment(t *testing.T) {
	t.Run("rejects pending comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusPending, 0)
		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)
		mockCommentRepo.On("UpdateStatus", mock.Anything, testComment.ID, comment.StatusSpam).Return(nil)

		err := svc.RejectComment(context.Background(), testComment.ID)

		assert.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("cannot reject deleted comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusDeleted, 0)
		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)

		err := svc.RejectComment(context.Background(), testComment.ID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "已删除")
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_DeleteComment(t *testing.T) {
	t.Run("author can delete own comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusApproved, 0)
		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)
		mockCommentRepo.On("Delete", mock.Anything, testComment.ID).Return(nil)

		err := svc.DeleteComment(context.Background(), testComment.ID, testComment.UserID)

		assert.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("non-author cannot delete comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusApproved, 0)
		otherUserID := uuid.New()

		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)

		err := svc.DeleteComment(context.Background(), testComment.ID, otherUserID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, comment.ErrPermissionDenied)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("comment not found returns error", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		id := uuid.New()
		mockCommentRepo.On("GetByID", mock.Anything, id).Return(nil, comment.ErrCommentNotFound)

		err := svc.DeleteComment(context.Background(), id, uuid.New())

		assert.Error(t, err)
		assert.ErrorIs(t, err, comment.ErrCommentNotFound)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_LikeComment(t *testing.T) {
	t.Run("successful like", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusApproved, 0)
		userID := uuid.New()

		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)
		mockLikeRepo.On("CreateIfNotExists", mock.Anything, testComment.ID, userID).Return(true, nil)

		err := svc.LikeComment(context.Background(), testComment.ID, userID)

		assert.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
		mockLikeRepo.AssertExpectations(t)
	})

	t.Run("cannot like non-approved comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusPending, 0)
		userID := uuid.New()

		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)

		err := svc.LikeComment(context.Background(), testComment.ID, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "已批准")
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("already liked returns error", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusApproved, 0)
		userID := uuid.New()

		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)
		mockLikeRepo.On("CreateIfNotExists", mock.Anything, testComment.ID, userID).Return(false, nil)

		err := svc.LikeComment(context.Background(), testComment.ID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, comment.ErrAlreadyLiked)
		mockCommentRepo.AssertExpectations(t)
		mockLikeRepo.AssertExpectations(t)
	})

	t.Run("comment not found returns error", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		id := uuid.New()
		mockCommentRepo.On("GetByID", mock.Anything, id).Return(nil, comment.ErrCommentNotFound)

		err := svc.LikeComment(context.Background(), id, uuid.New())

		assert.Error(t, err)
		assert.ErrorIs(t, err, comment.ErrCommentNotFound)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_UnlikeComment(t *testing.T) {
	t.Run("successful unlike", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusApproved, 0)
		userID := uuid.New()

		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)
		mockLikeRepo.On("DeleteIfExists", mock.Anything, testComment.ID, userID).Return(true, nil)

		err := svc.UnlikeComment(context.Background(), testComment.ID, userID)

		assert.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
		mockLikeRepo.AssertExpectations(t)
	})

	t.Run("not liked returns error", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("test content", comment.StatusApproved, 0)
		userID := uuid.New()

		mockCommentRepo.On("GetByID", mock.Anything, testComment.ID).Return(testComment, nil)
		mockLikeRepo.On("DeleteIfExists", mock.Anything, testComment.ID, userID).Return(false, nil)

		err := svc.UnlikeComment(context.Background(), testComment.ID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, comment.ErrNotLiked)
		mockCommentRepo.AssertExpectations(t)
		mockLikeRepo.AssertExpectations(t)
	})
}

func TestCommentService_BatchOperations(t *testing.T) {
	t.Run("batch approve comments", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		id1 := uuid.New()
		id2 := uuid.New()
		comment1 := createTestComment("comment1", comment.StatusPending, 0)
		comment1.ID = id1
		comment2 := createTestComment("comment2", comment.StatusPending, 0)
		comment2.ID = id2

		mockCommentRepo.On("GetByID", mock.Anything, id1).Return(comment1, nil)
		mockCommentRepo.On("UpdateStatus", mock.Anything, id1, comment.StatusApproved).Return(nil)
		mockCommentRepo.On("GetByID", mock.Anything, id2).Return(comment2, nil)
		mockCommentRepo.On("UpdateStatus", mock.Anything, id2, comment.StatusApproved).Return(nil)

		err := svc.BatchApproveComments(context.Background(), []uuid.UUID{id1, id2})

		assert.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("batch delete comments", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		id1 := uuid.New()
		id2 := uuid.New()
		comment1 := createTestComment("comment1", comment.StatusApproved, 0)
		comment1.ID = id1
		comment2 := createTestComment("comment2", comment.StatusApproved, 0)
		comment2.ID = id2

		mockCommentRepo.On("GetByID", mock.Anything, id1).Return(comment1, nil)
		mockCommentRepo.On("Delete", mock.Anything, id1).Return(nil)
		mockCommentRepo.On("GetByID", mock.Anything, id2).Return(comment2, nil)
		mockCommentRepo.On("Delete", mock.Anything, id2).Return(nil)

		err := svc.BatchDeleteComments(context.Background(), []uuid.UUID{id1, id2})

		assert.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_CountCommentsByPost(t *testing.T) {
	t.Run("returns count", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		postID := uuid.New()
		mockCommentRepo.On("CountByPostID", mock.Anything, postID).Return(5, nil)

		count, err := svc.CountCommentsByPost(context.Background(), postID)

		assert.NoError(t, err)
		assert.Equal(t, 5, count)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_GetCommentsByUser(t *testing.T) {
	t.Run("returns user comments", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		userID := uuid.New()
		comments := []*comment.Comment{
			createTestComment("comment1", comment.StatusApproved, 0),
			createTestComment("comment2", comment.StatusApproved, 0),
		}
		mockCommentRepo.On("GetByUserID", mock.Anything, userID).Return(comments, nil)

		result, err := svc.GetCommentsByUser(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_UpdateComment(t *testing.T) {
	t.Run("updates comment content", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		testComment := createTestComment("old content", comment.StatusApproved, 0)
		mockCommentRepo.On("Update", mock.Anything, testComment).Return(nil)

		err := svc.UpdateComment(context.Background(), testComment)

		assert.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_IsLiked(t *testing.T) {
	t.Run("returns true when liked", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		commentID := uuid.New()
		userID := uuid.New()
		mockLikeRepo.On("Exists", mock.Anything, commentID, userID).Return(true, nil)

		liked, err := svc.IsLiked(context.Background(), commentID, userID)

		assert.NoError(t, err)
		assert.True(t, liked)
		mockLikeRepo.AssertExpectations(t)
	})

	t.Run("returns false when not liked", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		commentID := uuid.New()
		userID := uuid.New()
		mockLikeRepo.On("Exists", mock.Anything, commentID, userID).Return(false, nil)

		liked, err := svc.IsLiked(context.Background(), commentID, userID)

		assert.NoError(t, err)
		assert.False(t, liked)
		mockLikeRepo.AssertExpectations(t)
	})
}

func TestCommentService_GetLikesBatch(t *testing.T) {
	t.Run("returns likes map", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		userID := uuid.New()
		commentIDs := []uuid.UUID{uuid.New(), uuid.New()}
		// Mock Exists for each comment ID (fallback implementation)
		mockLikeRepo.On("Exists", mock.Anything, commentIDs[0], userID).Return(true, nil)
		mockLikeRepo.On("Exists", mock.Anything, commentIDs[1], userID).Return(false, nil)

		result, err := svc.GetLikesBatch(context.Background(), commentIDs, userID)

		assert.NoError(t, err)
		assert.True(t, result[commentIDs[0]])
		assert.False(t, result[commentIDs[1]])
		mockLikeRepo.AssertExpectations(t)
	})
}

func TestCommentService_GetCommentsByStatus(t *testing.T) {
	t.Run("returns comments by status", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		comments := []*comment.Comment{
			createTestComment("pending1", comment.StatusPending, 0),
			createTestComment("pending2", comment.StatusPending, 0),
		}
		filters := &comment.CommentListFilters{Status: comment.StatusPending}
		mockCommentRepo.On("List", mock.Anything, filters, 0, 20).Return(comments, nil)

		result, total, err := svc.GetCommentsByStatus(context.Background(), comment.StatusPending, 0, 20)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, total)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestCommentService_BatchRejectComments(t *testing.T) {
	t.Run("batch reject comments", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockLikeRepo := new(mocks.MockCommentLikeRepository)
		svc := NewCommentService(mockCommentRepo, mockLikeRepo)

		id1 := uuid.New()
		id2 := uuid.New()
		comment1 := createTestComment("comment1", comment.StatusPending, 0)
		comment1.ID = id1
		comment2 := createTestComment("comment2", comment.StatusPending, 0)
		comment2.ID = id2

		mockCommentRepo.On("GetByID", mock.Anything, id1).Return(comment1, nil)
		mockCommentRepo.On("UpdateStatus", mock.Anything, id1, comment.StatusSpam).Return(nil)
		mockCommentRepo.On("GetByID", mock.Anything, id2).Return(comment2, nil)
		mockCommentRepo.On("UpdateStatus", mock.Anything, id2, comment.StatusSpam).Return(nil)

		err := svc.BatchRejectComments(context.Background(), []uuid.UUID{id1, id2})

		assert.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})
}
