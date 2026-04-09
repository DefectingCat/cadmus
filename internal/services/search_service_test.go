package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"rua.plus/cadmus/internal/core/search"
	"rua.plus/cadmus/test/mocks"
)

func TestSearchService_Search(t *testing.T) {
	t.Run("successful search", func(t *testing.T) {
		mockRepo := new(mocks.MockSearchRepository)
		svc := NewSearchService(mockRepo)

		filters := search.SearchFilters{Query: "golang"}
		results := []search.SearchResult{
			{Title: "Go Tutorial", Rank: 0.9},
			{Title: "Go Advanced", Rank: 0.8},
		}
		mockRepo.On("Search", mock.Anything, "golang", filters, 0, 20).Return(results, 2, nil)

		resp, err := svc.Search(context.Background(), filters, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(resp.Results))
		assert.Equal(t, 2, resp.Total)
		assert.Equal(t, 1, resp.Page)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty query returns error", func(t *testing.T) {
		mockRepo := new(mocks.MockSearchRepository)
		svc := NewSearchService(mockRepo)

		filters := search.SearchFilters{Query: ""}
		_, err := svc.Search(context.Background(), filters, 1, 20)

		assert.Error(t, err)
		assert.ErrorIs(t, err, search.ErrEmptyQuery)
	})

	t.Run("query too long returns error", func(t *testing.T) {
		mockRepo := new(mocks.MockSearchRepository)
		svc := NewSearchService(mockRepo)

		longQuery := make([]byte, 101)
		for i := range longQuery {
			longQuery[i] = 'a'
		}
		filters := search.SearchFilters{Query: string(longQuery)}
		_, err := svc.Search(context.Background(), filters, 1, 20)

		assert.Error(t, err)
		assert.ErrorIs(t, err, search.ErrQueryTooLong)
	})

	t.Run("corrects invalid pagination params", func(t *testing.T) {
		mockRepo := new(mocks.MockSearchRepository)
		svc := NewSearchService(mockRepo)

		filters := search.SearchFilters{Query: "test"}
		results := []search.SearchResult{}
		mockRepo.On("Search", mock.Anything, "test", filters, 0, 20).Return(results, 0, nil)

		// page < 1 should be corrected to 1
		// pageSize > 100 should be corrected to 20
		resp, err := svc.Search(context.Background(), filters, 0, 200)

		assert.NoError(t, err)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 20, resp.PageSize)
		mockRepo.AssertExpectations(t)
	})
}

func TestSearchService_SearchByCategory(t *testing.T) {
	t.Run("search within category", func(t *testing.T) {
		mockRepo := new(mocks.MockSearchRepository)
		svc := NewSearchService(mockRepo)

		categoryID := uuid.New()
		results := []search.SearchResult{
			{Title: "Category Post", Rank: 0.9},
		}
		mockRepo.On("SearchByCategory", mock.Anything, "test", categoryID, 0, 20).Return(results, 1, nil)

		resp, err := svc.SearchByCategory(context.Background(), "test", categoryID, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Results))
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty query returns error", func(t *testing.T) {
		mockRepo := new(mocks.MockSearchRepository)
		svc := NewSearchService(mockRepo)

		_, err := svc.SearchByCategory(context.Background(), "", uuid.New(), 1, 20)

		assert.Error(t, err)
		assert.ErrorIs(t, err, search.ErrEmptyQuery)
	})
}

func TestSearchService_SearchByAuthor(t *testing.T) {
	t.Run("search by author", func(t *testing.T) {
		mockRepo := new(mocks.MockSearchRepository)
		svc := NewSearchService(mockRepo)

		authorID := uuid.New()
		results := []search.SearchResult{
			{Title: "Author Post", Rank: 0.9},
		}
		mockRepo.On("SearchByAuthor", mock.Anything, "test", authorID, 0, 20).Return(results, 1, nil)

		resp, err := svc.SearchByAuthor(context.Background(), "test", authorID, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Results))
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty query returns error", func(t *testing.T) {
		mockRepo := new(mocks.MockSearchRepository)
		svc := NewSearchService(mockRepo)

		_, err := svc.SearchByAuthor(context.Background(), "", uuid.New(), 1, 20)

		assert.Error(t, err)
		assert.ErrorIs(t, err, search.ErrEmptyQuery)
	})
}

func TestSearchService_GetSuggestions(t *testing.T) {
	t.Run("returns suggestions", func(t *testing.T) {
		mockRepo := new(mocks.MockSearchRepository)
		svc := NewSearchService(mockRepo)

		suggestions := []string{"golang", "go tutorial", "go basics"}
		mockRepo.On("GetSuggestions", mock.Anything, "go", 5).Return(suggestions, nil)

		result, err := svc.GetSuggestions(context.Background(), "go", 5)

		assert.NoError(t, err)
		assert.Equal(t, suggestions, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("corrects invalid limit", func(t *testing.T) {
		mockRepo := new(mocks.MockSearchRepository)
		svc := NewSearchService(mockRepo)

		suggestions := []string{}
		mockRepo.On("GetSuggestions", mock.Anything, "test", 5).Return(suggestions, nil)

		// limit > 10 should be corrected to 5
		result, err := svc.GetSuggestions(context.Background(), "test", 20)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		_ = result
	})
}
