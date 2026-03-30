package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User 用户信息（用于认证服务返回）
type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	RoleID       uuid.UUID
	Status       string
}

// UserStatus 用户状态常量
const (
	UserStatusActive  = "active"
	UserStatusBanned  = "banned"
	UserStatusPending = "pending"
)

// UserRepository 用户数据仓库接口
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) error
}

// TokenBlacklist token 黑名单接口（可选）
type TokenBlacklist interface {
	Add(ctx context.Context, token string, expiry time.Time) error
	Exists(ctx context.Context, token string) bool
}

// AuthService 认证服务
type AuthService struct {
	jwtService *JWTService
	userRepo   UserRepository
	blacklist  TokenBlacklist // 可选
}

// NewAuthService 创建新的认证服务
func NewAuthService(jwtService *JWTService, userRepo UserRepository) *AuthService {
	return &AuthService{
		jwtService: jwtService,
		userRepo:   userRepo,
	}
}

// WithBlacklist 设置 token 黑名单
func (s *AuthService) WithBlacklist(blacklist TokenBlacklist) *AuthService {
	s.blacklist = blacklist
	return s
}

// LoginResult 登录结果
type LoginResult struct {
	Token string
	User  *User
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	// 查找用户
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 检查用户状态
	if user.Status == UserStatusBanned {
		return nil, errors.New("user is banned")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 生成 token
	token, err := s.jwtService.Generate(user.ID, user.RoleID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token: token,
		User:  user,
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*User, error) {
	// 检查邮箱是否已存在
	if existing, _ := s.userRepo.GetByEmail(ctx, email); existing != nil {
		return nil, errors.New("email already registered")
	}

	// 检查用户名是否已存在
	if existing, _ := s.userRepo.GetByUsername(ctx, username); existing != nil {
		return nil, errors.New("username already taken")
	}

	// 哈希密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		Status:       UserStatusPending, // 默认待激活状态
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// Logout 用户登出（可选：加入黑名单）
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if s.blacklist == nil {
		return nil // 无黑名单时直接返回
	}

	// 获取 token 过期时间
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return err // token 无效，无需加入黑名单
	}

	// 将 token 加入黑名单
	expiry := claims.ExpiresAt.Time
	return s.blacklist.Add(ctx, token, expiry)
}

// IsTokenBlacklisted 检查 token 是否在黑名单中
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) bool {
	if s.blacklist == nil {
		return false
	}
	return s.blacklist.Exists(ctx, token)
}

// ValidateToken 验证 token 并返回用户信息
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*Claims, *User, error) {
	// 检查黑名单
	if s.IsTokenBlacklisted(ctx, token) {
		return nil, nil, errors.New("token is blacklisted")
	}

	// 验证 token
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return nil, nil, err
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, errors.New("user not found")
	}

	// 检查用户状态
	if user.Status == UserStatusBanned {
		return nil, nil, errors.New("user is banned")
	}

	return claims, user, nil
}