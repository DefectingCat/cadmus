package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/test/mocks"
)

// ============================================================
// PostService Tests
// ============================================================

func createTestPost(title, slug string, status post.PostStatus) *post.Post {
	return &post.Post{
		ID:        uuid.New(),
		AuthorID:  uuid.New(),
		Title:     title,
		Slug:      slug,
		Status:    status,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestPostService_Create(t *testing.T) {
	t.Run("successful creation with default status", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockTagRepo := new(mocks.MockTagRepository)
		svc := NewPostService(mockPostRepo, nil, mockTagRepo, nil)

		testPost := createTestPost("Test Title", "test-slug", "")
		mockPostRepo.On("Create", mock.Anything, testPost).Return(nil)

		err := svc.Create(context.Background(), testPost, nil)

		assert.NoError(t, err)
		assert.Equal(t, post.StatusDraft, testPost.Status)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("creation with tags", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockTagRepo := new(mocks.MockTagRepository)
		svc := NewPostService(mockPostRepo, nil, mockTagRepo, nil)

		testPost := createTestPost("Test Title", "test-slug", post.StatusDraft)
		tagID := uuid.New()
		mockPostRepo.On("Create", mock.Anything, testPost).Return(nil)
		mockTagRepo.On("AddPostTag", mock.Anything, testPost.ID, tagID).Return(nil)

		err := svc.Create(context.Background(), testPost, []uuid.UUID{tagID})

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("missing title returns error", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		testPost := createTestPost("", "test-slug", post.StatusDraft)

		err := svc.Create(context.Background(), testPost, nil)

		assert.Error(t, err)
		assert.Equal(t, "title and slug are required", err.Error())
	})

	t.Run("missing slug returns error", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		testPost := createTestPost("Test Title", "", post.StatusDraft)

		err := svc.Create(context.Background(), testPost, nil)

		assert.Error(t, err)
		assert.Equal(t, "title and slug are required", err.Error())
	})

	t.Run("invalid status returns error", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		testPost := createTestPost("Test Title", "test-slug", post.PostStatus("invalid"))

		err := svc.Create(context.Background(), testPost, nil)

		assert.Error(t, err)
		assert.ErrorIs(t, err, post.ErrInvalidStatus)
	})
}

func TestPostService_Update(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockTagRepo := new(mocks.MockTagRepository)
		svc := NewPostService(mockPostRepo, nil, mockTagRepo, nil)

		testPost := createTestPost("Updated Title", "test-slug", post.StatusPublished)
		mockPostRepo.On("Update", mock.Anything, testPost).Return(nil)
		mockTagRepo.On("GetPostTags", mock.Anything, testPost.ID).Return([]*post.Tag{}, nil)

		err := svc.Update(context.Background(), testPost, nil)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("invalid status returns error", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		testPost := createTestPost("Test Title", "test-slug", post.PostStatus("invalid"))

		err := svc.Update(context.Background(), testPost, nil)

		assert.Error(t, err)
		assert.ErrorIs(t, err, post.ErrInvalidStatus)
	})

	t.Run("update replaces tags", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockTagRepo := new(mocks.MockTagRepository)
		svc := NewPostService(mockPostRepo, nil, mockTagRepo, nil)

		testPost := createTestPost("Test Title", "test-slug", post.StatusPublished)
		oldTagID := uuid.New()
		newTagID := uuid.New()
		oldTag := &post.Tag{ID: oldTagID, Name: "old-tag"}

		mockPostRepo.On("Update", mock.Anything, testPost).Return(nil)
		mockTagRepo.On("GetPostTags", mock.Anything, testPost.ID).Return([]*post.Tag{oldTag}, nil)
		mockTagRepo.On("RemovePostTag", mock.Anything, testPost.ID, oldTagID).Return(nil)
		mockTagRepo.On("AddPostTag", mock.Anything, testPost.ID, newTagID).Return(nil)

		err := svc.Update(context.Background(), testPost, []uuid.UUID{newTagID})

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestPostService_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		postID := uuid.New()
		mockPostRepo.On("Delete", mock.Anything, postID).Return(nil)

		err := svc.Delete(context.Background(), postID)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
	})
}

func TestPostService_GetByID(t *testing.T) {
	t.Run("returns post when found", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		testPost := createTestPost("Test Title", "test-slug", post.StatusPublished)
		mockPostRepo.On("GetByID", mock.Anything, testPost.ID).Return(testPost, nil)

		result, err := svc.GetByID(context.Background(), testPost.ID)

		assert.NoError(t, err)
		assert.Equal(t, testPost, result)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		postID := uuid.New()
		mockPostRepo.On("GetByID", mock.Anything, postID).Return(nil, post.ErrPostNotFound)

		_, err := svc.GetByID(context.Background(), postID)

		assert.Error(t, err)
		mockPostRepo.AssertExpectations(t)
	})
}

func TestPostService_Publish(t *testing.T) {
	t.Run("publishes draft post", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		testPost := createTestPost("Test Title", "test-slug", post.StatusDraft)
		mockPostRepo.On("GetByID", mock.Anything, testPost.ID).Return(testPost, nil)
		mockPostRepo.On("Update", mock.Anything, testPost).Return(nil)

		err := svc.Publish(context.Background(), testPost.ID)

		assert.NoError(t, err)
		assert.Equal(t, post.StatusPublished, testPost.Status)
		assert.NotNil(t, testPost.PublishAt)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("returns error when post not found", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		postID := uuid.New()
		mockPostRepo.On("GetByID", mock.Anything, postID).Return(nil, post.ErrPostNotFound)

		err := svc.Publish(context.Background(), postID)

		assert.Error(t, err)
		mockPostRepo.AssertExpectations(t)
	})
}

func TestPostService_Schedule(t *testing.T) {
	t.Run("schedules post for future publication", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		testPost := createTestPost("Test Title", "test-slug", post.StatusDraft)
		publishAt := time.Now().Add(24 * time.Hour)

		mockPostRepo.On("GetByID", mock.Anything, testPost.ID).Return(testPost, nil)
		mockPostRepo.On("Update", mock.Anything, testPost).Return(nil)

		err := svc.Schedule(context.Background(), testPost.ID, publishAt)

		assert.NoError(t, err)
		assert.Equal(t, post.StatusScheduled, testPost.Status)
		assert.NotNil(t, testPost.PublishAt)
		assert.Equal(t, publishAt, *testPost.PublishAt)
		mockPostRepo.AssertExpectations(t)
	})
}

func TestPostService_CreateVersion(t *testing.T) {
	t.Run("creates version snapshot", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		testPost := createTestPost("Test Title", "test-slug", post.StatusPublished)
		testPost.Version = 5
		testPost.Content = []byte("test content")
		creatorID := uuid.New()

		mockPostRepo.On("GetByID", mock.Anything, testPost.ID).Return(testPost, nil)
		mockPostRepo.On("CreateVersion", mock.Anything, mock.Anything).Return(nil)

		err := svc.CreateVersion(context.Background(), testPost.ID, "test note", creatorID)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
	})
}

func TestPostService_Rollback(t *testing.T) {
	t.Run("rollbacks to previous version", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		testPost := createTestPost("Test Title", "test-slug", post.StatusPublished)
		testPost.Content = []byte("current content")

		targetVersion := &post.PostVersion{
			ID:      uuid.New(),
			PostID:  testPost.ID,
			Version: 3,
			Content: []byte("old content"),
		}

		mockPostRepo.On("GetByID", mock.Anything, testPost.ID).Return(testPost, nil)
		mockPostRepo.On("GetVersionByNumber", mock.Anything, testPost.ID, 3).Return(targetVersion, nil)
		mockPostRepo.On("Update", mock.Anything, testPost).Return(nil)

		err := svc.Rollback(context.Background(), testPost.ID, 3)

		assert.NoError(t, err)
		assert.Equal(t, []byte("old content"), testPost.Content)
		mockPostRepo.AssertExpectations(t)
	})
}

func TestPostService_LikePost(t *testing.T) {
	t.Run("successful like", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockLikeRepo := new(mocks.MockPostLikeRepository)
		svc := NewPostServiceWithLikes(mockPostRepo, nil, nil, nil, mockLikeRepo)

		testPost := createTestPost("Test Title", "test-slug", post.StatusPublished)
		userID := uuid.New()

		mockPostRepo.On("GetByID", mock.Anything, testPost.ID).Return(testPost, nil)
		mockLikeRepo.On("CreateIfNotExists", mock.Anything, testPost.ID, userID).Return(true, nil)

		err := svc.LikePost(context.Background(), testPost.ID, userID)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
		mockLikeRepo.AssertExpectations(t)
	})

	t.Run("post not found returns error", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockLikeRepo := new(mocks.MockPostLikeRepository)
		svc := NewPostServiceWithLikes(mockPostRepo, nil, nil, nil, mockLikeRepo)

		postID := uuid.New()
		userID := uuid.New()

		mockPostRepo.On("GetByID", mock.Anything, postID).Return(nil, post.ErrPostNotFound)

		err := svc.LikePost(context.Background(), postID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, post.ErrPostNotFound)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("already liked returns error", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockLikeRepo := new(mocks.MockPostLikeRepository)
		svc := NewPostServiceWithLikes(mockPostRepo, nil, nil, nil, mockLikeRepo)

		testPost := createTestPost("Test Title", "test-slug", post.StatusPublished)
		userID := uuid.New()

		mockPostRepo.On("GetByID", mock.Anything, testPost.ID).Return(testPost, nil)
		mockLikeRepo.On("CreateIfNotExists", mock.Anything, testPost.ID, userID).Return(false, nil)

		err := svc.LikePost(context.Background(), testPost.ID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, post.ErrAlreadyLiked)
		mockPostRepo.AssertExpectations(t)
		mockLikeRepo.AssertExpectations(t)
	})
}

func TestPostService_UnlikePost(t *testing.T) {
	t.Run("successful unlike", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockLikeRepo := new(mocks.MockPostLikeRepository)
		svc := NewPostServiceWithLikes(mockPostRepo, nil, nil, nil, mockLikeRepo)

		testPost := createTestPost("Test Title", "test-slug", post.StatusPublished)
		userID := uuid.New()

		mockPostRepo.On("GetByID", mock.Anything, testPost.ID).Return(testPost, nil)
		mockLikeRepo.On("DeleteIfExists", mock.Anything, testPost.ID, userID).Return(true, nil)

		err := svc.UnlikePost(context.Background(), testPost.ID, userID)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
		mockLikeRepo.AssertExpectations(t)
	})

	t.Run("not liked returns error", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockLikeRepo := new(mocks.MockPostLikeRepository)
		svc := NewPostServiceWithLikes(mockPostRepo, nil, nil, nil, mockLikeRepo)

		testPost := createTestPost("Test Title", "test-slug", post.StatusPublished)
		userID := uuid.New()

		mockPostRepo.On("GetByID", mock.Anything, testPost.ID).Return(testPost, nil)
		mockLikeRepo.On("DeleteIfExists", mock.Anything, testPost.ID, userID).Return(false, nil)

		err := svc.UnlikePost(context.Background(), testPost.ID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, post.ErrNotLiked)
		mockPostRepo.AssertExpectations(t)
		mockLikeRepo.AssertExpectations(t)
	})
}

func TestPostService_List(t *testing.T) {
	t.Run("returns paginated list", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		posts := []*post.Post{
			createTestPost("Post 1", "post-1", post.StatusPublished),
			createTestPost("Post 2", "post-2", post.StatusPublished),
		}
		filters := post.PostListFilters{}
		mockPostRepo.On("List", mock.Anything, filters, 0, 20).Return(posts, 2, nil)

		result, total, err := svc.List(context.Background(), filters, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, 2, total)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("corrects invalid pagination params", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		svc := NewPostService(mockPostRepo, nil, nil, nil)

		posts := []*post.Post{}
		filters := post.PostListFilters{}
		mockPostRepo.On("List", mock.Anything, filters, 0, 20).Return(posts, 0, nil)

		// page < 1 should be corrected to 1
		// pageSize > 100 should be corrected to 20
		result, _, err := svc.List(context.Background(), filters, 0, 200)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockPostRepo.AssertExpectations(t)
	})
}

// ============================================================
// CategoryService Tests
// ============================================================

func TestCategoryService_Create(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		mockRepo := new(mocks.MockCategoryRepository)
		svc := NewCategoryService(mockRepo)

		cat := &post.Category{
			ID:   uuid.New(),
			Name: "Technology",
			Slug: "technology",
		}
		mockRepo.On("Create", mock.Anything, cat).Return(nil)

		err := svc.Create(context.Background(), cat)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryService_GetByID(t *testing.T) {
	t.Run("returns category when found", func(t *testing.T) {
		mockRepo := new(mocks.MockCategoryRepository)
		svc := NewCategoryService(mockRepo)

		cat := &post.Category{ID: uuid.New(), Name: "Tech", Slug: "tech"}
		mockRepo.On("GetByID", mock.Anything, cat.ID).Return(cat, nil)

		result, err := svc.GetByID(context.Background(), cat.ID)

		assert.NoError(t, err)
		assert.Equal(t, cat, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryService_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockRepo := new(mocks.MockCategoryRepository)
		svc := NewCategoryService(mockRepo)

		id := uuid.New()
		mockRepo.On("Delete", mock.Anything, id).Return(nil)

		err := svc.Delete(context.Background(), id)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

// ============================================================
// TagService Tests
// ============================================================

func TestTagService_Create(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		mockRepo := new(mocks.MockTagRepository)
		svc := NewTagService(mockRepo)

		tag := &post.Tag{
			ID:   uuid.New(),
			Name: "Go",
			Slug: "go",
		}
		mockRepo.On("Create", mock.Anything, tag).Return(nil)

		err := svc.Create(context.Background(), tag)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTagService_GetByID(t *testing.T) {
	t.Run("returns tag when found", func(t *testing.T) {
		mockRepo := new(mocks.MockTagRepository)
		svc := NewTagService(mockRepo)

		tag := &post.Tag{ID: uuid.New(), Name: "Go", Slug: "go"}
		mockRepo.On("GetByID", mock.Anything, tag.ID).Return(tag, nil)

		result, err := svc.GetByID(context.Background(), tag.ID)

		assert.NoError(t, err)
		assert.Equal(t, tag, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestTagService_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockRepo := new(mocks.MockTagRepository)
		svc := NewTagService(mockRepo)

		id := uuid.New()
		mockRepo.On("Delete", mock.Anything, id).Return(nil)

		err := svc.Delete(context.Background(), id)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

// ============================================================
// SeriesService Tests
// ============================================================

func TestSeriesService_Create(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		mockRepo := new(mocks.MockSeriesRepository)
		svc := NewSeriesService(mockRepo)

		series := &post.Series{
			ID:       uuid.New(),
			AuthorID: uuid.New(),
			Title:    "Go Tutorial",
			Slug:     "go-tutorial",
		}
		mockRepo.On("Create", mock.Anything, series).Return(nil)

		err := svc.Create(context.Background(), series)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSeriesService_GetByID(t *testing.T) {
	t.Run("returns series when found", func(t *testing.T) {
		mockRepo := new(mocks.MockSeriesRepository)
		svc := NewSeriesService(mockRepo)

		series := &post.Series{ID: uuid.New(), Title: "Go Tutorial", Slug: "go-tutorial"}
		mockRepo.On("GetByID", mock.Anything, series.ID).Return(series, nil)

		result, err := svc.GetByID(context.Background(), series.ID)

		assert.NoError(t, err)
		assert.Equal(t, series, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestSeriesService_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockRepo := new(mocks.MockSeriesRepository)
		svc := NewSeriesService(mockRepo)

		id := uuid.New()
		mockRepo.On("Delete", mock.Anything, id).Return(nil)

		err := svc.Delete(context.Background(), id)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
