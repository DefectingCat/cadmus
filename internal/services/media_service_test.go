package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"rua.plus/cadmus/internal/core/media"
	"rua.plus/cadmus/test/mocks"
)

func TestMediaService_GetByID(t *testing.T) {
	t.Run("returns media when found", func(t *testing.T) {
		mockRepo := new(mocks.MockMediaRepository)
		svc := NewMediaService(mockRepo, "/uploads", "https://example.com")

		testMedia := &media.Media{
			ID:           uuid.New(),
			UploaderID:   uuid.New(),
			Filename:     "test.jpg",
			OriginalName: "original.jpg",
			URL:          "https://example.com/uploads/test.jpg",
			MimeType:     "image/jpeg",
		}
		mockRepo.On("GetByID", mock.Anything, testMedia.ID).Return(testMedia, nil)

		result, err := svc.GetByID(context.Background(), testMedia.ID)

		assert.NoError(t, err)
		assert.Equal(t, testMedia, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		mockRepo := new(mocks.MockMediaRepository)
		svc := NewMediaService(mockRepo, "/uploads", "https://example.com")

		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, media.ErrMediaNotFound)

		_, err := svc.GetByID(context.Background(), id)

		assert.Error(t, err)
		assert.ErrorIs(t, err, media.ErrMediaNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestMediaService_GetByUser(t *testing.T) {
	t.Run("returns user media list", func(t *testing.T) {
		mockRepo := new(mocks.MockMediaRepository)
		svc := NewMediaService(mockRepo, "/uploads", "https://example.com")

		userID := uuid.New()
		medias := []*media.Media{
			{ID: uuid.New(), UploaderID: userID, Filename: "file1.jpg"},
			{ID: uuid.New(), UploaderID: userID, Filename: "file2.jpg"},
		}
		mockRepo.On("GetByUploaderID", mock.Anything, userID).Return(medias, nil)

		result, err := svc.GetByUser(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})
}

func TestMediaService_Delete(t *testing.T) {
	t.Run("uploader can delete own media", func(t *testing.T) {
		mockRepo := new(mocks.MockMediaRepository)
		svc := NewMediaService(mockRepo, "/uploads", "https://example.com")

		userID := uuid.New()
		testMedia := &media.Media{
			ID:         uuid.New(),
			UploaderID: userID,
			Filename:   "test.jpg",
			FilePath:   "/uploads/test.jpg",
		}
		mockRepo.On("GetByID", mock.Anything, testMedia.ID).Return(testMedia, nil)
		mockRepo.On("Delete", mock.Anything, testMedia.ID).Return(nil)

		err := svc.Delete(context.Background(), testMedia.ID, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("non-uploader cannot delete media", func(t *testing.T) {
		mockRepo := new(mocks.MockMediaRepository)
		svc := NewMediaService(mockRepo, "/uploads", "https://example.com")

		userID := uuid.New()
		otherUserID := uuid.New()
		testMedia := &media.Media{
			ID:         uuid.New(),
			UploaderID: userID,
			Filename:   "test.jpg",
		}
		mockRepo.On("GetByID", mock.Anything, testMedia.ID).Return(testMedia, nil)

		err := svc.Delete(context.Background(), testMedia.ID, otherUserID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, media.ErrPermissionDenied)
		mockRepo.AssertExpectations(t)
	})

	t.Run("media not found returns error", func(t *testing.T) {
		mockRepo := new(mocks.MockMediaRepository)
		svc := NewMediaService(mockRepo, "/uploads", "https://example.com")

		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, media.ErrMediaNotFound)

		err := svc.Delete(context.Background(), id, uuid.New())

		assert.Error(t, err)
		assert.ErrorIs(t, err, media.ErrMediaNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestMediaService_List(t *testing.T) {
	t.Run("returns paginated media list", func(t *testing.T) {
		mockRepo := new(mocks.MockMediaRepository)
		svc := NewMediaService(mockRepo, "/uploads", "https://example.com")

		filters := &media.MediaListFilters{}
		medias := []*media.Media{
			{ID: uuid.New(), Filename: "file1.jpg"},
			{ID: uuid.New(), Filename: "file2.jpg"},
		}
		mockRepo.On("List", mock.Anything, filters, 0, 10).Return(medias, nil)
		mockRepo.On("Count", mock.Anything, filters).Return(2, nil)

		result, total, err := svc.List(context.Background(), filters, 0, 10)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, total)
		mockRepo.AssertExpectations(t)
	})
}

func TestGenerateUniqueFilename(t *testing.T) {
	t.Run("generates unique filename with extension", func(t *testing.T) {
		filename1 := generateUniqueFilename(".jpg")
		filename2 := generateUniqueFilename(".jpg")

		assert.NotEqual(t, filename1, filename2)
		assert.Contains(t, filename1, ".jpg")
		assert.Contains(t, filename2, ".jpg")
		assert.Len(t, filename1, 40) // UUID (36) + extension (4)
	})
}

func TestMimeTypeFromExt(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".JPG", "image/jpeg"},
		{".png", "image/png"},
		{".gif", "image/gif"},
		{".webp", "image/webp"},
		{".svg", "image/svg+xml"},
		{".pdf", "application/pdf"},
		{".doc", "application/msword"},
		{".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{".zip", "application/zip"},
		{".txt", "text/plain"},
		{".unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run("extension_"+tt.ext, func(t *testing.T) {
			result := mimeTypeFromExt(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}