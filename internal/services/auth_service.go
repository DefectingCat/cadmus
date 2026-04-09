// Package services 提供认证服务的实现。
//
// 该文件包含用户认证相关的核心逻辑，包括：
//   - 用户登录验证（邮箱/密码校验）
//   - Token 生成与刷新
//   - Token 黑名单管理（支持登出失效）
//   - Token 验证与用户信息获取
//
// 主要用途：
//
//	用于处理用户认证流程，确保系统安全访问控制。
//
// 安全设计：
//   - 登录失败返回统一错误消息，避免信息泄露
//   - 支持 Token 黑名单，实现即时登出失效
//   - 检查用户状态，禁止被封禁用户登录
//
// 作者：xfy
package services

import (
	"context"
	"errors"
	"time"

	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/user"
)

// AuthService 认证业务服务接口。
//
// 该接口定义了用户认证的核心操作，包括登录、登出、Token 刷新和验证。
// 所有方法均为并发安全，可在多个 goroutine 中同时调用。
type AuthService interface {
	// Login 用户登录，验证邮箱和密码，返回 Token 和用户信息。
	//
	// 参数：
	//   - ctx: 上下文，用于控制超时和取消操作
	//   - email: 用户邮箱地址
	//   - password: 用户密码（明文，内部进行哈希比对）
	//
	// 返回值：
	//   - token: JWT Token 字符串，用于后续请求认证
	//   - user: 登录成功的用户对象
	//   - err: 可能的错误包括：
	//       - "invalid credentials": 邮箱或密码错误（统一消息，避免信息泄露）
	//       - "user is banned": 用户已被封禁
	//
	// 使用示例：
	//   token, user, err := authService.Login(ctx, "user@example.com", "password")
	Login(ctx context.Context, email, password string) (string, *user.User, error)

	// Logout 用户登出，将 Token 加入黑名单使其失效。
	//
	// 参数：
	//   - ctx: 上下文
	//   - token: 需失效的 JWT Token
	//
	// 返回值：
	//   - err: Token 无效或无 jti 时返回错误
	//
	// 注意事项：
	//   - 需要配置黑名单才能生效，否则静默返回 nil
	Logout(ctx context.Context, token string) error

	// Refresh 刷新 Token，生成新的有效 Token。
	//
	// 参数：
	//   - token: 当前有效的 JWT Token
	//
	// 返回值：
	//   - 新的 JWT Token 字符串
	//   - err: Token 无效或过期时返回错误
	Refresh(token string) (string, error)

	// ValidateToken 验证 Token 并返回 Claims 和用户信息。
	//
	// 参数：
	//   - ctx: 上下文
	//   - token: 待验证的 JWT Token
	//
	// 返回值：
	//   - claims: Token 中的声明信息（用户 ID、角色 ID 等）
	//   - user: 用户对象
	//   - err: Token 无效、在黑名单中或用户不存在时返回错误
	ValidateToken(ctx context.Context, token string) (*auth.Claims, *user.User, error)
}

// TokenBlacklist Token 黑名单接口。
//
// 该接口定义了 Token 黑名单的操作，用于实现登出后 Token 即时失效。
// 黑名单通常基于 Redis 或内存实现，存储 Token 的 jti（JWT ID）。
//
// 实现要求：
//   - 所有方法必须是并发安全的
//   - AddToBlacklist 应设置与 Token 过期时间一致的 TTL
type TokenBlacklist interface {
	// AddToBlacklist 将 Token 加入黑名单。
	//
	// 参数：
	//   - ctx: 上下文
	//   - tokenID: Token 的 jti（JWT ID）
	//   - expiry: Token 过期时间，用于设置黑名单条目的 TTL
	AddToBlacklist(ctx context.Context, tokenID string, expiry time.Time) error

	// IsBlacklisted 检查 Token 是否在黑名单中。
	//
	// 参数：
	//   - ctx: 上下文
	//   - tokenID: Token 的 jti
	//
	// 返回值：
	//   - true 表示 Token 已失效，应拒绝访问
	IsBlacklisted(ctx context.Context, tokenID string) bool
}

// authServiceImpl 认证服务的具体实现。
//
// 该结构体实现了 AuthService 接口，依赖 UserRepository 和 JWTService。
// 可选支持 Token 黑名单，通过 blacklist 字段配置。
type authServiceImpl struct {
	// userRepo 用户数据仓库，用于查询用户信息
	userRepo user.UserRepository

	// jwtService JWT 服务，用于 Token 生成和验证
	jwtService *auth.JWTService

	// blacklist Token 黑名单（可选），用于实现登出失效
	blacklist TokenBlacklist
}

// NewAuthService 创建认证服务。
//
// 该函数创建一个不带黑名单的基础认证服务，适用于不需要即时登出失效的场景。
//
// 参数：
//   - userRepo: 用户数据仓库
//   - jwtService: JWT 服务实例
//
// 返回值：
//   - AuthService: 认证服务接口实例
func NewAuthService(userRepo user.UserRepository, jwtService *auth.JWTService) AuthService {
	return &authServiceImpl{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// NewAuthServiceWithBlacklist 创建带黑名单的认证服务。
//
// 该函数创建一个支持 Token 黑名单的认证服务，适用于需要即时登出失效的场景。
// 黑名单会在每次 Token 验证时检查，增加额外的存储查询开销。
//
// 参数：
//   - userRepo: 用户数据仓库
//   - jwtService: JWT 服务实例
//   - blacklist: Token 黑名单实现（通常基于 Redis）
//
// 返回值：
//   - AuthService: 包含黑名单功能的认证服务
//
// 注意事项：
//   - 黑名单存储需配置合理的 TTL，避免无限增长
func NewAuthServiceWithBlacklist(userRepo user.UserRepository, jwtService *auth.JWTService, blacklist TokenBlacklist) AuthService {
	return &authServiceImpl{
		userRepo:   userRepo,
		jwtService: jwtService,
		blacklist:  blacklist,
	}
}

// WithBlacklist 设置 Token 黑名单，返回更新后的服务实例。
//
// 该方法用于在已创建的认证服务上动态添加黑名单功能。
// 采用链式调用风格，便于服务配置。
//
// 参数：
//   - blacklist: Token 黑名单实现
//
// 返回值：
//   - AuthService: 更新后的认证服务实例（实际返回同一实例）
func (s *authServiceImpl) WithBlacklist(blacklist TokenBlacklist) AuthService {
	s.blacklist = blacklist
	return s
}

// Login 用户登录，验证邮箱和密码。
//
// 该方法执行以下步骤：
//  1. 根据邮箱查询用户
//  2. 检查用户状态（是否被封禁）
//  3. 验证密码哈希
//  4. 生成 JWT Token
//
// 安全设计：
//   - 登录失败返回统一错误消息 "invalid credentials"，避免信息泄露
//   - 被封禁用户返回 "user is banned"，提示管理员处理
//
// 参数：
//   - ctx: 上下文，用于控制超时
//   - email: 用户邮箱地址
//   - password: 用户密码（明文）
//
// 返回值：
//   - token: JWT Token 字符串
//   - user: 登录成功的用户对象
//   - err: 登录失败时返回错误
func (s *authServiceImpl) Login(ctx context.Context, email, password string) (string, *user.User, error) {
	// 查找用户
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// 返回统一错误消息，避免信息泄露（用户不存在）
		return "", nil, errors.New("invalid credentials")
	}

	// 检查用户状态：被封禁用户禁止登录
	if u.Status == user.StatusBanned {
		return "", nil, errors.New("user is banned")
	}

	// 验证密码哈希
	if !u.CheckPassword(password) {
		// 返回统一错误消息，避免信息泄露（密码错误）
		return "", nil, errors.New("invalid credentials")
	}

	// 生成 JWT Token
	token, _, err := s.jwtService.Generate(u.ID, u.RoleID)
	if err != nil {
		return "", nil, err
	}

	return token, u, nil
}

// Logout 用户登出，将 Token 加入黑名单。
//
// 该方法将当前 Token 的 jti 加入黑名单，使其立即失效。
// 后续使用该 Token 的请求将被拒绝。
//
// 参数：
//   - ctx: 上下文
//   - token: 需失效的 JWT Token
//
// 返回值：
//   - err: 可能的错误包括：
//   - Token 无效
//   - Token 无 jti 字段
//
// 注意事项：
//   - 未配置黑名单时静默返回 nil，不影响调用方
func (s *authServiceImpl) Logout(ctx context.Context, token string) error {
	// 未配置黑名单时直接返回，不执行登出逻辑
	if s.blacklist == nil {
		return nil // 无黑名单时直接返回
	}

	// 解析 Token 获取过期时间和 jti
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return err // token 无效，无需加入黑名单
	}

	// 将 Token 的 jti 加入黑名单，过期时间与 Token 一致
	expiry := claims.ExpiresAt.Time
	jti := claims.GetJWTID()
	if jti == "" {
		return errors.New("token has no jti")
	}
	return s.blacklist.AddToBlacklist(ctx, jti, expiry)
}

// Refresh 刷新 Token，生成新的有效 Token。
//
// 该方法使用当前 Token 生成一个新的 Token，延长会话有效期。
// 新 Token 的 Claims 内容与原 Token 相同，但过期时间更新。
//
// 参数：
//   - token: 当前有效的 JWT Token
//
// 返回值：
//   - 新的 JWT Token 字符串
//   - err: Token 无效或过期时返回错误
func (s *authServiceImpl) Refresh(token string) (string, error) {
	newToken, _, err := s.jwtService.Refresh(token)
	return newToken, err
}

// ValidateToken 验证 Token 并返回 Claims 和用户信息。
//
// 该方法执行完整的 Token 验证流程：
//  1. JWT 格式和签名验证
//  2. 黑名单检查（如已配置）
//  3. 用户状态检查（是否被封禁）
//
// 参数：
//   - ctx: 上下文
//   - token: 待验证的 JWT Token
//
// 返回值：
//   - claims: Token 中的声明信息
//   - user: 用户对象
//   - err: 可能的错误包括：
//   - Token 无效或过期
//   - Token 在黑名单中
//   - 用户不存在
//   - 用户已被封禁
//
// 使用示例：
//
//	claims, user, err := authService.ValidateToken(ctx, token)
//	if err != nil {
//	    // Token 无效，返回 401 错误
//	    return
//	}
//	// 使用 user.ID 获取当前用户身份
func (s *authServiceImpl) ValidateToken(ctx context.Context, token string) (*auth.Claims, *user.User, error) {
	// 步骤1: 验证 JWT Token 格式和签名
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return nil, nil, err
	}

	// 步骤2: 检查黑名单（使用 jti 标识）
	if s.blacklist != nil {
		jti := claims.GetJWTID()
		if jti != "" && s.blacklist.IsBlacklisted(ctx, jti) {
			return nil, nil, errors.New("token is blacklisted")
		}
	}

	// 步骤3: 获取用户信息
	u, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, errors.New("user not found")
	}

	// 步骤4: 检查用户状态（被封禁用户禁止访问）
	if u.Status == user.StatusBanned {
		return nil, nil, errors.New("user is banned")
	}

	return claims, u, nil
}

// IsTokenBlacklisted 检查 Token 是否在黑名单中。
//
// 该方法是一个辅助方法，用于检查 Token 的失效状态。
// 主要用于调试和测试场景，生产代码通常通过 ValidateToken 完成验证。
//
// 参数：
//   - ctx: 上下文
//   - token: 待检查的 JWT Token
//
// 返回值：
//   - true: Token 已失效（在黑名单中）
//   - false: Token 未失效或黑名单未配置
func (s *authServiceImpl) IsTokenBlacklisted(ctx context.Context, token string) bool {
	// 未配置黑名单时返回 false
	if s.blacklist == nil {
		return false
	}

	// 解析 Token 获取 jti
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return false // Token 无效视为未黑名单
	}

	// 检查 jti 是否在黑名单中
	jti := claims.GetJWTID()
	if jti == "" {
		return false // 无 jti 无法检查黑名单
	}
	return s.blacklist.IsBlacklisted(ctx, jti)
}
