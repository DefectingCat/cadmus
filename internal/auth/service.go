package auth

import (
	"context"
	"errors"

	"rua.plus/cadmus/internal/core/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserStatus 用户状态常量（从 user 包导出）
const (
	UserStatusActive  = user.StatusActive
	UserStatusBanned  = user.StatusBanned
	UserStatusPending = user.StatusPending
)

// UserRepository 用户数据仓库接口（使用 user.User）
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	Create(ctx context.Context, user *user.User) error
	Update(ctx context.Context, user *user.User) error
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
	User  *user.User
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
	token, _, err := s.jwtService.Generate(user.ID, user.RoleID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token: token,
		User:  user,
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*user.User, error) {
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
	newUser := &user.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		Status:       user.StatusPending, // 默认待激活状态
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// Logout 用户登出（可选：加入黑名单）
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if s.blacklist == nil {
		return nil // 无黑名单时直接返回
	}

	// 获取 token claims（包含 jti 和过期时间）
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return err // token 无效，无需加入黑名单
	}

	// 将 token jti 加入黑名单
	jti := claims.GetJWTID()
	if jti == "" {
		return nil // 无 jti，无法加入黑名单
	}
	expiry := claims.ExpiresAt.Time
	return s.blacklist.AddToBlacklist(ctx, jti, expiry)
}

// IsTokenBlacklisted 检查 token 是否在黑名单中
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) bool {
	if s.blacklist == nil {
		return false
	}

	// 获取 token jti
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return false // token 无效，视为不在黑名单
	}

	jti := claims.GetJWTID()
	if jti == "" {
		return false // 无 jti，不在黑名单
	}

	return s.blacklist.IsBlacklisted(ctx, jti)
}

// ValidateToken 验证 token 并返回用户信息
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*Claims, *user.User, error) {
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