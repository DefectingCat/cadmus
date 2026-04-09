package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/internal/core/media"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/core/search"
	"rua.plus/cadmus/internal/core/user"
)

// ============================================================
// User Module Mocks
// ============================================================

// MockUserRepository 是 UserRepository 接口的 mock 实现
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id [16]byte) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id [16]byte) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*user.User), args.Get(1).(int), args.Error(2)
}

// MockRoleRepository 是 RoleRepository 接口的 mock 实现
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id [16]byte) (*user.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Role), args.Error(1)
}

func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*user.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Role), args.Error(1)
}

func (m *MockRoleRepository) GetAll(ctx context.Context) ([]user.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]user.Role), args.Error(1)
}

func (m *MockRoleRepository) GetDefault(ctx context.Context) (*user.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Role), args.Error(1)
}

func (m *MockRoleRepository) GetWithPermissions(ctx context.Context, id [16]byte) (*user.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Role), args.Error(1)
}

// MockTokenBlacklist 是 TokenBlacklist 接口的 mock 实现
type MockTokenBlacklist struct {
	mock.Mock
}

func (m *MockTokenBlacklist) AddToBlacklist(ctx context.Context, tokenID string, expiry int64) error {
	args := m.Called(ctx, tokenID, expiry)
	return args.Error(0)
}

func (m *MockTokenBlacklist) IsBlacklisted(ctx context.Context, tokenID string) bool {
	args := m.Called(ctx, tokenID)
	return args.Bool(0)
}

// ============================================================
// Post Module Mocks
// ============================================================

// MockPostRepository 是 PostRepository 接口的 mock 实现
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(ctx context.Context, p *post.Post) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPostRepository) Update(ctx context.Context, p *post.Post) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.Post), args.Error(1)
}

func (m *MockPostRepository) GetBySlug(ctx context.Context, slug string) (*post.Post, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.Post), args.Error(1)
}

func (m *MockPostRepository) List(ctx context.Context, filters post.PostListFilters, offset, limit int) ([]*post.Post, int, error) {
	args := m.Called(ctx, filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*post.Post), args.Get(1).(int), args.Error(2)
}

func (m *MockPostRepository) GetByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int) ([]*post.Post, int, error) {
	args := m.Called(ctx, authorID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*post.Post), args.Get(1).(int), args.Error(2)
}

func (m *MockPostRepository) GetByCategory(ctx context.Context, categoryID uuid.UUID, offset, limit int) ([]*post.Post, int, error) {
	args := m.Called(ctx, categoryID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*post.Post), args.Get(1).(int), args.Error(2)
}

func (m *MockPostRepository) GetBySeries(ctx context.Context, seriesID uuid.UUID, offset, limit int) ([]*post.Post, int, error) {
	args := m.Called(ctx, seriesID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*post.Post), args.Get(1).(int), args.Error(2)
}

func (m *MockPostRepository) Search(ctx context.Context, query string, offset, limit int) ([]*post.Post, int, error) {
	args := m.Called(ctx, query, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*post.Post), args.Get(1).(int), args.Error(2)
}

func (m *MockPostRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) IncrementLikeCount(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) CreateVersion(ctx context.Context, v *post.PostVersion) error {
	args := m.Called(ctx, v)
	return args.Error(0)
}

func (m *MockPostRepository) GetVersions(ctx context.Context, postID uuid.UUID) ([]*post.PostVersion, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*post.PostVersion), args.Error(1)
}

func (m *MockPostRepository) GetVersionByNumber(ctx context.Context, postID uuid.UUID, version int) (*post.PostVersion, error) {
	args := m.Called(ctx, postID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.PostVersion), args.Error(1)
}

// MockCategoryRepository 是 CategoryRepository 接口的 mock 实现
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(ctx context.Context, c *post.Category) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCategoryRepository) Update(ctx context.Context, c *post.Category) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*post.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetBySlug(ctx context.Context, slug string) (*post.Category, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetAll(ctx context.Context) ([]*post.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*post.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*post.Category, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*post.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetRootCategories(ctx context.Context) ([]*post.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*post.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetPostCount(ctx context.Context, categoryID uuid.UUID) (int, error) {
	args := m.Called(ctx, categoryID)
	return args.Get(0).(int), args.Error(1)
}

// MockTagRepository 是 TagRepository 接口的 mock 实现
type MockTagRepository struct {
	mock.Mock
}

func (m *MockTagRepository) Create(ctx context.Context, t *post.Tag) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*post.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.Tag), args.Error(1)
}

func (m *MockTagRepository) GetBySlug(ctx context.Context, slug string) (*post.Tag, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.Tag), args.Error(1)
}

func (m *MockTagRepository) GetByName(ctx context.Context, name string) (*post.Tag, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.Tag), args.Error(1)
}

func (m *MockTagRepository) GetAll(ctx context.Context) ([]*post.Tag, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*post.Tag), args.Error(1)
}

func (m *MockTagRepository) AddPostTag(ctx context.Context, postID, tagID uuid.UUID) error {
	args := m.Called(ctx, postID, tagID)
	return args.Error(0)
}

func (m *MockTagRepository) RemovePostTag(ctx context.Context, postID, tagID uuid.UUID) error {
	args := m.Called(ctx, postID, tagID)
	return args.Error(0)
}

func (m *MockTagRepository) GetPostTags(ctx context.Context, postID uuid.UUID) ([]*post.Tag, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*post.Tag), args.Error(1)
}

func (m *MockTagRepository) GetPostCount(ctx context.Context, tagID uuid.UUID) (int, error) {
	args := m.Called(ctx, tagID)
	return args.Get(0).(int), args.Error(1)
}

// MockSeriesRepository 是 SeriesRepository 接口的 mock 实现
type MockSeriesRepository struct {
	mock.Mock
}

func (m *MockSeriesRepository) Create(ctx context.Context, s *post.Series) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSeriesRepository) Update(ctx context.Context, s *post.Series) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSeriesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSeriesRepository) GetByID(ctx context.Context, id uuid.UUID) (*post.Series, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.Series), args.Error(1)
}

func (m *MockSeriesRepository) GetBySlug(ctx context.Context, slug string) (*post.Series, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.Series), args.Error(1)
}

func (m *MockSeriesRepository) GetByAuthor(ctx context.Context, authorID uuid.UUID) ([]*post.Series, error) {
	args := m.Called(ctx, authorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*post.Series), args.Error(1)
}

func (m *MockSeriesRepository) GetAll(ctx context.Context) ([]*post.Series, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*post.Series), args.Error(1)
}

// MockPostLikeRepository 是 PostLikeRepository 接口的 mock 实现
type MockPostLikeRepository struct {
	mock.Mock
}

func (m *MockPostLikeRepository) CreateIfNotExists(ctx context.Context, postID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, postID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostLikeRepository) DeleteIfExists(ctx context.Context, postID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, postID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostLikeRepository) Exists(ctx context.Context, postID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, postID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostLikeRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockPostLikeRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*post.PostLike, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*post.PostLike), args.Error(1)
}

// ============================================================
// Comment Module Mocks
// ============================================================

// MockCommentRepository 是 CommentRepository 接口的 mock 实现
type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) Create(ctx context.Context, input *comment.CreateCommentInput) (*comment.Comment, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*comment.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*comment.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*comment.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetByPostID(ctx context.Context, postID uuid.UUID, filters *comment.CommentListFilters) ([]*comment.Comment, error) {
	args := m.Called(ctx, postID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*comment.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*comment.Comment, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*comment.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*comment.Comment, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*comment.Comment), args.Error(1)
}

func (m *MockCommentRepository) Update(ctx context.Context, c *comment.Comment) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCommentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status comment.CommentStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockCommentRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockCommentRepository) List(ctx context.Context, filters *comment.CommentListFilters, offset, limit int) ([]*comment.Comment, error) {
	args := m.Called(ctx, filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*comment.Comment), args.Error(1)
}

// MockCommentLikeRepository 是 CommentLikeRepository 接口的 mock 实现
type MockCommentLikeRepository struct {
	mock.Mock
}

func (m *MockCommentLikeRepository) Create(ctx context.Context, commentID, userID uuid.UUID) (*comment.CommentLike, error) {
	args := m.Called(ctx, commentID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*comment.CommentLike), args.Error(1)
}

func (m *MockCommentLikeRepository) CreateIfNotExists(ctx context.Context, commentID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, commentID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCommentLikeRepository) GetByCommentAndUser(ctx context.Context, commentID, userID uuid.UUID) (*comment.CommentLike, error) {
	args := m.Called(ctx, commentID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*comment.CommentLike), args.Error(1)
}

func (m *MockCommentLikeRepository) Delete(ctx context.Context, commentID, userID uuid.UUID) error {
	args := m.Called(ctx, commentID, userID)
	return args.Error(0)
}

func (m *MockCommentLikeRepository) DeleteIfExists(ctx context.Context, commentID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, commentID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCommentLikeRepository) Exists(ctx context.Context, commentID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, commentID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCommentLikeRepository) CountByCommentID(ctx context.Context, commentID uuid.UUID) (int, error) {
	args := m.Called(ctx, commentID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockCommentLikeRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*comment.CommentLike, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*comment.CommentLike), args.Error(1)
}

// ============================================================
// Media Module Mocks
// ============================================================

// MockMediaRepository 是 MediaRepository 接口的 mock 实现
type MockMediaRepository struct {
	mock.Mock
}

func (m *MockMediaRepository) Create(ctx context.Context, input *media.UploadInput, filename, filepath, url string, width, height *int) (*media.Media, error) {
	args := m.Called(ctx, input, filename, filepath, url, width, height)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*media.Media), args.Error(1)
}

func (m *MockMediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*media.Media, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*media.Media), args.Error(1)
}

func (m *MockMediaRepository) GetByUploaderID(ctx context.Context, uploaderID uuid.UUID) ([]*media.Media, error) {
	args := m.Called(ctx, uploaderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*media.Media), args.Error(1)
}

func (m *MockMediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMediaRepository) List(ctx context.Context, filters *media.MediaListFilters, offset, limit int) ([]*media.Media, error) {
	args := m.Called(ctx, filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*media.Media), args.Error(1)
}

func (m *MockMediaRepository) Count(ctx context.Context, filters *media.MediaListFilters) (int, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).(int), args.Error(1)
}

// ============================================================
// Search Module Mocks
// ============================================================

// MockSearchRepository 是 SearchRepository 接口的 mock 实现
type MockSearchRepository struct {
	mock.Mock
}

func (m *MockSearchRepository) Search(ctx context.Context, query string, filters search.SearchFilters, offset, limit int) ([]search.SearchResult, int, error) {
	args := m.Called(ctx, query, filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]search.SearchResult), args.Get(1).(int), args.Error(2)
}

func (m *MockSearchRepository) SearchByCategory(ctx context.Context, query string, categoryID uuid.UUID, offset, limit int) ([]search.SearchResult, int, error) {
	args := m.Called(ctx, query, categoryID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]search.SearchResult), args.Get(1).(int), args.Error(2)
}

func (m *MockSearchRepository) SearchByAuthor(ctx context.Context, query string, authorID uuid.UUID, offset, limit int) ([]search.SearchResult, int, error) {
	args := m.Called(ctx, query, authorID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]search.SearchResult), args.Get(1).(int), args.Error(2)
}

func (m *MockSearchRepository) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
