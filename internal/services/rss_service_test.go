package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/core/rss"
	"rua.plus/cadmus/test/mocks"
)

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func TestRSSService_GenerateFeed(t *testing.T) {
	t.Run("generates feed without category", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockCategoryRepo := new(mocks.MockCategoryRepository)
		svc := NewRSSService(mockPostRepo, mockCategoryRepo)

		posts := []*post.Post{
			{
				ID:        uuid.New(),
				Title:     "Test Post",
				Slug:      "test-post",
				Excerpt:   "Test excerpt",
				Status:    post.StatusPublished,
				CreatedAt: parseTime("2024-01-01T00:00:00Z"),
			},
		}
		mockPostRepo.On("List", mock.Anything, post.PostListFilters{Status: post.StatusPublished}, 0, 20).Return(posts, 1, nil)

		config := rss.FeedConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "Test blog description",
			BaseURL:     "https://example.com/posts",
		}

		feed, err := svc.GenerateFeed(context.Background(), config, "")

		assert.NoError(t, err)
		assert.Contains(t, feed, "Test Blog")
		assert.Contains(t, feed, "Test Post")
		assert.Contains(t, feed, "test-post")
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("generates feed with category", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockCategoryRepo := new(mocks.MockCategoryRepository)
		svc := NewRSSService(mockPostRepo, mockCategoryRepo)

		categoryID := uuid.New()
		category := &post.Category{
			ID:   categoryID,
			Name: "Technology",
			Slug: "tech",
		}
		posts := []*post.Post{
			{
				ID:         uuid.New(),
				Title:      "Tech Post",
				Slug:       "tech-post",
				Excerpt:    "Tech excerpt",
				Status:     post.StatusPublished,
				CategoryID: categoryID,
				CreatedAt:  parseTime("2024-01-01T00:00:00Z"),
			},
		}

		mockCategoryRepo.On("GetBySlug", mock.Anything, "tech").Return(category, nil)
		filters := post.PostListFilters{Status: post.StatusPublished, CategoryID: categoryID}
		mockPostRepo.On("List", mock.Anything, filters, 0, 20).Return(posts, 1, nil)

		config := rss.FeedConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "Test blog description",
			BaseURL:     "https://example.com/posts",
		}

		feed, err := svc.GenerateFeed(context.Background(), config, "tech")

		assert.NoError(t, err)
		assert.Contains(t, feed, "Test Blog")
		assert.Contains(t, feed, "Tech Post")
		mockCategoryRepo.AssertExpectations(t)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("handles post repo error", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockCategoryRepo := new(mocks.MockCategoryRepository)
		svc := NewRSSService(mockPostRepo, mockCategoryRepo)

		mockPostRepo.On("List", mock.Anything, post.PostListFilters{Status: post.StatusPublished}, 0, 20).Return(nil, 0, assert.AnError)

		config := rss.FeedConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "Test description",
		}

		_, err := svc.GenerateFeed(context.Background(), config, "")

		assert.Error(t, err)
		mockPostRepo.AssertExpectations(t)
	})
}

func TestRSSService_GenerateFeedForCategory(t *testing.T) {
	t.Run("generates feed for valid category ID", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockCategoryRepo := new(mocks.MockCategoryRepository)
		svc := NewRSSService(mockPostRepo, mockCategoryRepo)

		categoryID := uuid.New()
		category := &post.Category{
			ID:          categoryID,
			Name:        "Technology",
			Slug:        "tech",
			Description: "Tech articles",
		}

		mockCategoryRepo.On("GetByID", mock.Anything, categoryID).Return(category, nil)
		mockCategoryRepo.On("GetBySlug", mock.Anything, "tech").Return(category, nil)
		posts := []*post.Post{
			{
				ID:         uuid.New(),
				Title:      "Tech Post",
				Slug:       "tech-post",
				Excerpt:    "Tech excerpt",
				Status:     post.StatusPublished,
				CategoryID: categoryID,
				CreatedAt:  parseTime("2024-01-01T00:00:00Z"),
			},
		}
		filters := post.PostListFilters{Status: post.StatusPublished, CategoryID: categoryID}
		mockPostRepo.On("List", mock.Anything, filters, 0, 20).Return(posts, 1, nil)

		config := rss.FeedConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "Test blog",
			BaseURL:     "https://example.com/posts",
		}

		feed, err := svc.GenerateFeedForCategory(context.Background(), config, categoryID.String())

		assert.NoError(t, err)
		assert.Contains(t, feed, "Test Blog - Technology")
		mockCategoryRepo.AssertExpectations(t)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("fallback to full feed for invalid UUID", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockCategoryRepo := new(mocks.MockCategoryRepository)
		svc := NewRSSService(mockPostRepo, mockCategoryRepo)

		posts := []*post.Post{}
		mockPostRepo.On("List", mock.Anything, post.PostListFilters{Status: post.StatusPublished}, 0, 20).Return(posts, 0, nil)

		config := rss.FeedConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "Test blog",
		}

		feed, err := svc.GenerateFeedForCategory(context.Background(), config, "invalid-uuid")

		assert.NoError(t, err)
		assert.Contains(t, feed, "Test Blog")
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("fallback to full feed when category not found", func(t *testing.T) {
		mockPostRepo := new(mocks.MockPostRepository)
		mockCategoryRepo := new(mocks.MockCategoryRepository)
		svc := NewRSSService(mockPostRepo, mockCategoryRepo)

		categoryID := uuid.New()
		mockCategoryRepo.On("GetByID", mock.Anything, categoryID).Return(nil, assert.AnError)
		posts := []*post.Post{}
		mockPostRepo.On("List", mock.Anything, post.PostListFilters{Status: post.StatusPublished}, 0, 20).Return(posts, 0, nil)

		config := rss.FeedConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "Test blog",
		}

		feed, err := svc.GenerateFeedForCategory(context.Background(), config, categoryID.String())

		assert.NoError(t, err)
		assert.Contains(t, feed, "Test Blog")
		mockCategoryRepo.AssertExpectations(t)
		mockPostRepo.AssertExpectations(t)
	})
}
