package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/user"
	"rua.plus/cadmus/test/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// 测试用的固定密钥（至少 32 字符）
const testSecret = "test-secret-key-for-testing-32-chars"

// mockUserRepository 是 user.UserRepository 的 mock 实现
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*user.User), args.Get(1).(int), args.Error(2)
}

// mockRoleRepository 是 user.RoleRepository 的 mock 实现
type mockRoleRepository struct {
	mock.Mock
}

func (m *mockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Role), args.Error(1)
}

func (m *mockRoleRepository) GetByName(ctx context.Context, name string) (*user.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Role), args.Error(1)
}

func (m *mockRoleRepository) GetAll(ctx context.Context) ([]*user.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.Role), args.Error(1)
}

func (m *mockRoleRepository) GetDefault(ctx context.Context) (*user.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Role), args.Error(1)
}

func (m *mockRoleRepository) GetWithPermissions(ctx context.Context, id uuid.UUID) (*user.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Role), args.Error(1)
}

// mockTokenBlacklist 是 TokenBlacklist 的 mock 实现
type mockTokenBlacklist struct {
	mock.Mock
}

func (m *mockTokenBlacklist) AddToBlacklist(ctx context.Context, tokenID string, expiry time.Time) error {
	args := m.Called(ctx, tokenID, expiry)
	return args.Error(0)
}

func (m *mockTokenBlacklist) IsBlacklisted(ctx context.Context, tokenID string) bool {
	args := m.Called(ctx, tokenID)
	return args.Bool(0)
}

// createTestJWTService 创建测试用的 JWT 服务
func createTestJWTService(t *testing.T) *auth.JWTService {
	return auth.NewJWTService(auth.JWTConfig{
		Secret:        testSecret,
		Expiry:        time.Hour,
		RefreshExpiry: 24 * time.Hour,
		Issuer:        "test",
	})
}

// createTestUser 创建测试用户
func createTestUser(email, password string, status user.UserStatus) *user.User {
	u := &user.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    email,
		Status:   status,
		RoleID:   uuid.New(),
	}
	u.SetPassword(password)
	return u
}

func TestAuthService_Login(t *testing.T) {
	jwtService := createTestJWTService(t)

	t.Run("successful login returns token and user", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		svc := NewAuthService(mockRepo, jwtService)

		testUser := createTestUser("test@example.com", "password123", user.StatusActive)
		mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

		token, u, err := svc.Login(context.Background(), "test@example.com", "password123")

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Equal(t, testUser, u)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found returns invalid credentials", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		svc := NewAuthService(mockRepo, jwtService)

		mockRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, errors.New("not found"))

		token, u, err := svc.Login(context.Background(), "notfound@example.com", "password")

		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
		assert.Empty(t, token)
		assert.Nil(t, u)
		mockRepo.AssertExpectations(t)
	})

	t.Run("wrong password returns invalid credentials", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		svc := NewAuthService(mockRepo, jwtService)

		testUser := createTestUser("test@example.com", "correctpassword", user.StatusActive)
		mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

		token, u, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
		assert.Empty(t, token)
		assert.Nil(t, u)
		mockRepo.AssertExpectations(t)
	})

	t.Run("banned user returns user is banned", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		svc := NewAuthService(mockRepo, jwtService)

		testUser := createTestUser("banned@example.com", "password123", user.StatusBanned)
		mockRepo.On("GetByEmail", mock.Anything, "banned@example.com").Return(testUser, nil)

		token, u, err := svc.Login(context.Background(), "banned@example.com", "password123")

		assert.Error(t, err)
		assert.Equal(t, "user is banned", err.Error())
		assert.Empty(t, token)
		assert.Nil(t, u)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_Logout(t *testing.T) {
	jwtService := createTestJWTService(t)

	t.Run("without blacklist returns nil", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		svc := NewAuthService(mockRepo, jwtService)

		err := svc.Logout(context.Background(), "any-token")

		assert.NoError(t, err)
	})

	t.Run("with blacklist adds token to blacklist", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		mockBlacklist := new(mockTokenBlacklist)
		svc := NewAuthServiceWithBlacklist(mockRepo, jwtService, mockBlacklist)

		// 先生成一个有效 token
		userID := uuid.New()
		roleID := uuid.New()
		token, _, err := jwtService.Generate(userID, roleID)
		require.NoError(t, err)

		// 验证 token 获取 claims
		claims, err := jwtService.Validate(token)
		require.NoError(t, err)

		mockBlacklist.On("AddToBlacklist", mock.Anything, claims.GetJWTID(), claims.ExpiresAt.Time).Return(nil)

		err = svc.Logout(context.Background(), token)

		assert.NoError(t, err)
		mockBlacklist.AssertExpectations(t)
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		mockBlacklist := new(mockTokenBlacklist)
		svc := NewAuthServiceWithBlacklist(mockRepo, jwtService, mockBlacklist)

		err := svc.Logout(context.Background(), "invalid-token")

		assert.Error(t, err)
	})
}

func TestAuthService_Refresh(t *testing.T) {
	jwtService := createTestJWTService(t)
	mockRepo := new(mockUserRepository)
	svc := NewAuthService(mockRepo, jwtService)

	t.Run("refresh valid token returns new token", func(t *testing.T) {
		userID := uuid.New()
		roleID := uuid.New()
		token, _, err := jwtService.Generate(userID, roleID)
		require.NoError(t, err)

		newToken, err := svc.Refresh(token)

		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
		assert.NotEqual(t, token, newToken)
	})

	t.Run("refresh invalid token returns error", func(t *testing.T) {
		_, err := svc.Refresh("invalid-token")

		assert.Error(t, err)
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	jwtService := createTestJWTService(t)

	t.Run("valid token returns claims and user", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		svc := NewAuthService(mockRepo, jwtService)

		testUser := createTestUser("test@example.com", "password", user.StatusActive)
		token, _, err := jwtService.Generate(testUser.ID, testUser.RoleID)
		require.NoError(t, err)

		mockRepo.On("GetByID", mock.Anything, testUser.ID).Return(testUser, nil)

		claims, u, err := svc.ValidateToken(context.Background(), token)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, testUser, u)
		assert.Equal(t, testUser.ID, claims.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		svc := NewAuthService(mockRepo, jwtService)

		_, _, err := svc.ValidateToken(context.Background(), "invalid-token")

		assert.Error(t, err)
	})

	t.Run("blacklisted token returns error", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		mockBlacklist := new(mockTokenBlacklist)
		svc := NewAuthServiceWithBlacklist(mockRepo, jwtService, mockBlacklist)

		testUser := createTestUser("test@example.com", "password", user.StatusActive)
		token, _, err := jwtService.Generate(testUser.ID, testUser.RoleID)
		require.NoError(t, err)

		claims, err := jwtService.Validate(token)
		require.NoError(t, err)

		mockBlacklist.On("IsBlacklisted", mock.Anything, claims.GetJWTID()).Return(true)

		_, _, err = svc.ValidateToken(context.Background(), token)

		assert.Error(t, err)
		assert.Equal(t, "token is blacklisted", err.Error())
		mockBlacklist.AssertExpectations(t)
	})

	t.Run("user not found returns error", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		svc := NewAuthService(mockRepo, jwtService)

		userID := uuid.New()
		roleID := uuid.New()
		token, _, err := jwtService.Generate(userID, roleID)
		require.NoError(t, err)

		mockRepo.On("GetByID", mock.Anything, userID).Return(nil, errors.New("not found"))

		_, _, err = svc.ValidateToken(context.Background(), token)

		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("banned user returns error", func(t *testing.T) {
		mockRepo := new(mockUserRepository)
		svc := NewAuthService(mockRepo, jwtService)

		testUser := createTestUser("banned@example.com", "password", user.StatusBanned)
		token, _, err := jwtService.Generate(testUser.ID, testUser.RoleID)
		require.NoError(t, err)

		mockRepo.On("GetByID", mock.Anything, testUser.ID).Return(testUser, nil)

		_, _, err = svc.ValidateToken(context.Background(), token)

		assert.Error(t, err)
		assert.Equal(t, "user is banned", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)
		svc := NewUserService(mockUserRepo, mockRoleRepo)

		defaultRole := &user.Role{ID: uuid.New(), Name: "user"}
		mockUserRepo.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
		mockUserRepo.On("GetByUsername", mock.Anything, "newuser").Return(nil, errors.New("not found"))
		mockRoleRepo.On("GetDefault", mock.Anything).Return(defaultRole, nil)
		mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		u, err := svc.Register(context.Background(), "newuser", "new@example.com", "password123")

		assert.NoError(t, err)
		assert.NotNil(t, u)
		assert.Equal(t, "newuser", u.Username)
		assert.Equal(t, "new@example.com", u.Email)
		assert.Equal(t, user.StatusPending, u.Status)
		assert.NotEmpty(t, u.PasswordHash)
		mockUserRepo.AssertExpectations(t)
		mockRoleRepo.AssertExpectations(t)
	})

	t.Run("missing fields returns error", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)
		svc := NewUserService(mockUserRepo, mockRoleRepo)

		_, err := svc.Register(context.Background(), "", "email@example.com", "password")
		assert.Error(t, err)
		assert.Equal(t, "username, email and password are required", err.Error())

		_, err = svc.Register(context.Background(), "user", "", "password")
		assert.Error(t, err)

		_, err = svc.Register(context.Background(), "user", "email@example.com", "")
		assert.Error(t, err)
	})

	t.Run("duplicate email returns error", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)
		svc := NewUserService(mockUserRepo, mockRoleRepo)

		existingUser := createTestUser("existing@example.com", "password", user.StatusActive)
		mockUserRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)

		_, err := svc.Register(context.Background(), "newuser", "existing@example.com", "password")

		assert.Error(t, err)
		assert.ErrorIs(t, err, user.ErrUserAlreadyExists)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("duplicate username returns error", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)
		svc := NewUserService(mockUserRepo, mockRoleRepo)

		mockUserRepo.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
		existingUser := createTestUser("other@example.com", "password", user.StatusActive)
		existingUser.Username = "existinguser"
		mockUserRepo.On("GetByUsername", mock.Anything, "existinguser").Return(existingUser, nil)

		_, err := svc.Register(context.Background(), "existinguser", "new@example.com", "password")

		assert.Error(t, err)
		assert.ErrorIs(t, err, user.ErrUserAlreadyExists)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("failed to get default role returns error", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)
		svc := NewUserService(mockUserRepo, mockRoleRepo)

		mockUserRepo.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
		mockUserRepo.On("GetByUsername", mock.Anything, "newuser").Return(nil, errors.New("not found"))
		mockRoleRepo.On("GetDefault", mock.Anything).Return(nil, errors.New("no default role"))

		_, err := svc.Register(context.Background(), "newuser", "new@example.com", "password")

		assert.Error(t, err)
		assert.Equal(t, "failed to get default role", err.Error())
		mockUserRepo.AssertExpectations(t)
		mockRoleRepo.AssertExpectations(t)
	})
}

func TestUserService_GetByID(t *testing.T) {
	mockUserRepo := new(mockUserRepository)
	mockRoleRepo := new(mockRoleRepository)
	svc := NewUserService(mockUserRepo, mockRoleRepo)

	t.Run("returns user when found", func(t *testing.T) {
		testUser := createTestUser("test@example.com", "password", user.StatusActive)
		mockUserRepo.On("GetByID", mock.Anything, testUser.ID).Return(testUser, nil)

		u, err := svc.GetByID(context.Background(), testUser.ID)

		assert.NoError(t, err)
		assert.Equal(t, testUser, u)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		id := uuid.New()
		mockUserRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

		_, err := svc.GetByID(context.Background(), id)

		assert.Error(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetByEmail(t *testing.T) {
	mockUserRepo := new(mockUserRepository)
	mockRoleRepo := new(mockRoleRepository)
	svc := NewUserService(mockUserRepo, mockRoleRepo)

	t.Run("returns user when found", func(t *testing.T) {
		testUser := createTestUser("test@example.com", "password", user.StatusActive)
		mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

		u, err := svc.GetByEmail(context.Background(), "test@example.com")

		assert.NoError(t, err)
		assert.Equal(t, testUser, u)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		mockUserRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, errors.New("not found"))

		_, err := svc.GetByEmail(context.Background(), "notfound@example.com")

		assert.Error(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetByUsername(t *testing.T) {
	mockUserRepo := new(mockUserRepository)
	mockRoleRepo := new(mockRoleRepository)
	svc := NewUserService(mockUserRepo, mockRoleRepo)

	t.Run("returns user when found", func(t *testing.T) {
		testUser := createTestUser("test@example.com", "password", user.StatusActive)
		testUser.Username = "testuser"
		mockUserRepo.On("GetByUsername", mock.Anything, "testuser").Return(testUser, nil)

		u, err := svc.GetByUsername(context.Background(), "testuser")

		assert.NoError(t, err)
		assert.Equal(t, testUser, u)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		mockUserRepo.On("GetByUsername", mock.Anything, "notfound").Return(nil, errors.New("not found"))

		_, err := svc.GetByUsername(context.Background(), "notfound")

		assert.Error(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_Update(t *testing.T) {
	mockUserRepo := new(mockUserRepository)
	mockRoleRepo := new(mockRoleRepository)
	svc := NewUserService(mockUserRepo, mockRoleRepo)

	t.Run("successful update", func(t *testing.T) {
		testUser := createTestUser("test@example.com", "password", user.StatusActive)
		mockUserRepo.On("Update", mock.Anything, testUser).Return(nil)

		err := svc.Update(context.Background(), testUser)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("invalid status returns error", func(t *testing.T) {
		testUser := &user.User{
			ID:       uuid.New(),
			Username: "test",
			Email:    "test@example.com",
			Status:   user.UserStatus("invalid"),
		}

		err := svc.Update(context.Background(), testUser)

		assert.Error(t, err)
		assert.ErrorIs(t, err, user.ErrInvalidStatus)
	})
}

func TestUserService_Delete(t *testing.T) {
	mockUserRepo := new(mockUserRepository)
	mockRoleRepo := new(mockRoleRepository)
	svc := NewUserService(mockUserRepo, mockRoleRepo)

	t.Run("successful delete", func(t *testing.T) {
		id := uuid.New()
		mockUserRepo.On("Delete", mock.Anything, id).Return(nil)

		err := svc.Delete(context.Background(), id)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error when user not found", func(t *testing.T) {
		id := uuid.New()
		mockUserRepo.On("Delete", mock.Anything, id).Return(errors.New("not found"))

		err := svc.Delete(context.Background(), id)

		assert.Error(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_List(t *testing.T) {
	mockUserRepo := new(mockUserRepository)
	mockRoleRepo := new(mockRoleRepository)
	svc := NewUserService(mockUserRepo, mockRoleRepo)

	t.Run("returns paginated users", func(t *testing.T) {
		users := []*user.User{
			createTestUser("user1@example.com", "pass", user.StatusActive),
			createTestUser("user2@example.com", "pass", user.StatusActive),
		}
		mockUserRepo.On("List", mock.Anything, 0, 10).Return(users, 2, nil)

		result, total, err := svc.List(context.Background(), 0, 10)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, 2, total)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns empty list", func(t *testing.T) {
		mockUserRepo.On("List", mock.Anything, 100, 10).Return([]*user.User{}, 0, nil)

		result, total, err := svc.List(context.Background(), 100, 10)

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.Equal(t, 0, total)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestContainer_Constructors(t *testing.T) {
	jwtService := createTestJWTService(t)

	t.Run("NewContainer creates basic container", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)

		container := NewContainer(mockUserRepo, mockRoleRepo, jwtService)

		assert.NotNil(t, container)
		assert.NotNil(t, container.UserService)
		assert.NotNil(t, container.AuthService)
		assert.NotNil(t, container.JWTService())
	})

	t.Run("NewContainerWithBlacklist creates container with blacklist", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)
		mockBlacklist := new(mockTokenBlacklist)

		container := NewContainerWithBlacklist(mockUserRepo, mockRoleRepo, jwtService, mockBlacklist)

		assert.NotNil(t, container)
		assert.NotNil(t, container.AuthService)
	})

	t.Run("NewContainerWithPosts creates container with post services", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)
		mockPostRepo := new(mocks.MockPostRepository)
		mockCategoryRepo := new(mocks.MockCategoryRepository)
		mockTagRepo := new(mocks.MockTagRepository)
		mockSeriesRepo := new(mocks.MockSeriesRepository)

		container := NewContainerWithPosts(
			mockUserRepo, mockRoleRepo, jwtService, nil,
			mockPostRepo, mockCategoryRepo, mockTagRepo, mockSeriesRepo,
		)

		assert.NotNil(t, container)
		assert.NotNil(t, container.PostService)
		assert.NotNil(t, container.CategoryService)
		assert.NotNil(t, container.TagService)
		assert.NotNil(t, container.SeriesService)
	})

	t.Run("NewContainerWithComments creates container with comment service", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)
		mockPostRepo := new(mocks.MockPostRepository)
		mockCategoryRepo := new(mocks.MockCategoryRepository)
		mockTagRepo := new(mocks.MockTagRepository)
		mockSeriesRepo := new(mocks.MockSeriesRepository)
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockCommentLikeRepo := new(mocks.MockCommentLikeRepository)

		container := NewContainerWithComments(
			mockUserRepo, mockRoleRepo, jwtService, nil,
			mockPostRepo, mockCategoryRepo, mockTagRepo, mockSeriesRepo,
			mockCommentRepo, mockCommentLikeRepo,
		)

		assert.NotNil(t, container)
		assert.NotNil(t, container.CommentService)
	})

	t.Run("NewContainerWithMedia creates container with all core services", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)
		mockPostRepo := new(mocks.MockPostRepository)
		mockCategoryRepo := new(mocks.MockCategoryRepository)
		mockTagRepo := new(mocks.MockTagRepository)
		mockSeriesRepo := new(mocks.MockSeriesRepository)
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockCommentLikeRepo := new(mocks.MockCommentLikeRepository)
		mockMediaRepo := new(mocks.MockMediaRepository)
		mockPostLikeRepo := new(mocks.MockPostLikeRepository)
		mockSearchRepo := new(mocks.MockSearchRepository)

		container := NewContainerWithMedia(
			mockUserRepo, mockRoleRepo, jwtService, nil,
			mockPostRepo, mockCategoryRepo, mockTagRepo, mockSeriesRepo,
			mockCommentRepo, mockCommentLikeRepo,
			mockMediaRepo, "/uploads", "https://example.com",
			mockPostLikeRepo, mockSearchRepo,
		)

		assert.NotNil(t, container)
		assert.NotNil(t, container.MediaService)
		assert.NotNil(t, container.RSSService)
		assert.NotNil(t, container.SearchService)
	})

	t.Run("NewContainerWithNotifications creates full container", func(t *testing.T) {
		mockUserRepo := new(mockUserRepository)
		mockRoleRepo := new(mockRoleRepository)
		mockPostRepo := new(mocks.MockPostRepository)
		mockCategoryRepo := new(mocks.MockCategoryRepository)
		mockTagRepo := new(mocks.MockTagRepository)
		mockSeriesRepo := new(mocks.MockSeriesRepository)
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockCommentLikeRepo := new(mocks.MockCommentLikeRepository)
		mockMediaRepo := new(mocks.MockMediaRepository)
		mockPostLikeRepo := new(mocks.MockPostLikeRepository)
		mockSearchRepo := new(mocks.MockSearchRepository)
		mockNotificationChannel := new(MockNotificationChannel)

		container := NewContainerWithNotifications(
			mockUserRepo, mockRoleRepo, jwtService, nil,
			mockPostRepo, mockCategoryRepo, mockTagRepo, mockSeriesRepo,
			mockCommentRepo, mockCommentLikeRepo,
			mockMediaRepo, "/uploads", "https://example.com",
			mockPostLikeRepo, mockSearchRepo,
			mockNotificationChannel,
		)

		assert.NotNil(t, container)
		assert.NotNil(t, container.NotificationService)
	})
}