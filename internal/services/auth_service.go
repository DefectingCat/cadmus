package services

import (
	"context"
	"errors"
	"time"

	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/user"
)

// AuthService 认证业务服务接口
type AuthService interface {
	// Login 用户登录，返回 token 和用户信息
	Login(ctx context.Context, email, password string) (string, *user.User, error)

	// Logout 用户登出（可选加入黑名单）
	Logout(ctx context.Context, token string) error

	// Refresh 刷新 token
	Refresh(token string) (string, error)

	// ValidateToken 验证 token 并返回用户信息
	ValidateToken(ctx context.Context, token string) (*auth.Claims, *user.User, error)
}

// TokenBlacklist token 黑名单接口（基于 jti）
type TokenBlacklist interface {
	AddToBlacklist(ctx context.Context, tokenID string, expiry time.Time) error
	IsBlacklisted(ctx context.Context, tokenID string) bool
}

// authServiceImpl 认证服务实现
type authServiceImpl struct {
	userRepo   user.UserRepository
	jwtService *auth.JWTService
	blacklist  TokenBlacklist
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo user.UserRepository, jwtService *auth.JWTService) AuthService {
	return &authServiceImpl{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// NewAuthServiceWithBlacklist 创建带黑名单的认证服务
func NewAuthServiceWithBlacklist(userRepo user.UserRepository, jwtService *auth.JWTService, blacklist TokenBlacklist) AuthService {
	return &authServiceImpl{
		userRepo:   userRepo,
		jwtService: jwtService,
		blacklist:  blacklist,
	}
}

// WithBlacklist 设置 token 黑名单
func (s *authServiceImpl) WithBlacklist(blacklist TokenBlacklist) AuthService {
	s.blacklist = blacklist
	return s
}

// Login 用户登录
func (s *authServiceImpl) Login(ctx context.Context, email, password string) (string, *user.User, error) {
	// 查找用户
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// 检查用户状态
	if u.Status == user.StatusBanned {
		return "", nil, errors.New("user is banned")
	}

	// 验证密码
	if !u.CheckPassword(password) {
		return "", nil, errors.New("invalid credentials")
	}

	// 生成 token
	token, _, err := s.jwtService.Generate(u.ID, u.RoleID)
	if err != nil {
		return "", nil, err
	}

	return token, u, nil
}

// Logout 用户登出
func (s *authServiceImpl) Logout(ctx context.Context, token string) error {
	if s.blacklist == nil {
		return nil // 无黑名单时直接返回
	}

	// 获取 token 过期时间和 jti
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return err // token 无效，无需加入黑名单
	}

	// 将 token jti 加入黑名单
	expiry := claims.ExpiresAt.Time
	jti := claims.GetJWTID()
	if jti == "" {
		return errors.New("token has no jti")
	}
	return s.blacklist.AddToBlacklist(ctx, jti, expiry)
}

// Refresh 刷新 token
func (s *authServiceImpl) Refresh(token string) (string, error) {
	newToken, _, err := s.jwtService.Refresh(token)
	return newToken, err
}

// ValidateToken 验证 token 并返回用户信息
func (s *authServiceImpl) ValidateToken(ctx context.Context, token string) (*auth.Claims, *user.User, error) {
	// 验证 token
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return nil, nil, err
	}

	// 检查黑名单（使用 jti）
	if s.blacklist != nil {
		jti := claims.GetJWTID()
		if jti != "" && s.blacklist.IsBlacklisted(ctx, jti) {
			return nil, nil, errors.New("token is blacklisted")
		}
	}

	// 获取用户信息
	u, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, errors.New("user not found")
	}

	// 检查用户状态
	if u.Status == user.StatusBanned {
		return nil, nil, errors.New("user is banned")
	}

	return claims, u, nil
}

// IsTokenBlacklisted 检查 token 是否在黑名单中
func (s *authServiceImpl) IsTokenBlacklisted(ctx context.Context, token string) bool {
	if s.blacklist == nil {
		return false
	}

	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return false
	}

	jti := claims.GetJWTID()
	if jti == "" {
		return false
	}
	return s.blacklist.IsBlacklisted(ctx, jti)
}