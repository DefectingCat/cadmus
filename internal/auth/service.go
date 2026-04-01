// Package auth 提供了 Cadmus 认证和授权功能的核心实现。
//
// 该文件包含认证服务相关的核心逻辑，包括：
//   - 用户登录认证
//   - 用户注册处理
//   - JWT token 生成和验证
//   - Token 黑名单管理（可选）
//
// 主要用途：
//
//	用于处理用户身份认证的完整流程，从登录验证到 token 管理。
//
// 注意事项：
//   - 所有公开方法均为并发安全
//   - 使用前需确保 UserRepository 和 JWTService 已正确初始化
//   - 黑名单功能为可选，需要 Redis 支持
//
// 作者：xfy
package auth

import (
	"context"
	"errors"

	"rua.plus/cadmus/internal/core/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserStatus 用户状态常量（从 user 包导出）
//
// 这些常量用于表示用户账户的不同状态，影响认证流程的行为：
//   - UserStatusActive: 用户正常，可以登录和使用系统
//   - UserStatusBanned: 用户被封禁，无法登录
//   - UserStatusPending: 用户待激活，需要完成激活流程
const (
	UserStatusActive  = user.StatusActive   // 正常状态
	UserStatusBanned  = user.StatusBanned   // 封禁状态
	UserStatusPending = user.StatusPending  // 待激活状态
)

// UserRepository 用户数据仓库接口（使用 user.User）
//
// 该接口定义了用户数据存储的抽象操作，支持不同的存储后端实现。
// 实现要求：
//   - 所有方法必须是并发安全的
//   - 返回的错误应使用语义化错误类型
type UserRepository interface {
	// GetByEmail 根据邮箱查询用户。
	//
	// 参数：
	//   - ctx: 上下文，用于超时控制
	//   - email: 用户邮箱地址
	//
	// 返回值：
	//   - user: 用户对象，未找到时返回 nil
	//   - err: 查询错误
	GetByEmail(ctx context.Context, email string) (*user.User, error)

	// GetByUsername 根据用户名查询用户。
	//
	// 参数：
	//   - ctx: 上下文，用于超时控制
	//   - username: 用户名
	//
	// 返回值：
	//   - user: 用户对象，未找到时返回 nil
	//   - err: 查询错误
	GetByUsername(ctx context.Context, username string) (*user.User, error)

	// GetByID 根据用户 ID 查询用户。
	//
	// 参数：
	//   - ctx: 上下文，用于超时控制
	//   - id: 用户唯一标识符
	//
	// 返回值：
	//   - user: 用户对象，未找到时返回 nil
	//   - err: 查询错误
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)

	// Create 创建新用户。
	//
	// 参数：
	//   - ctx: 上下文，用于超时控制
	//   - user: 用户对象，包含创建所需的数据
	//
	// 返回值：
	//   - err: 创建失败时返回错误
	Create(ctx context.Context, user *user.User) error

	// Update 更新用户信息。
	//
	// 参数：
	//   - ctx: 上下文，用于超时控制
	//   - user: 用户对象，包含更新后的数据
	//
	// 返回值：
	//   - err: 更新失败时返回错误
	Update(ctx context.Context, user *user.User) error
}

// AuthService 认证服务，提供用户认证的核心功能。
//
// 该服务封装了用户登录、注册、token 验证等认证相关操作。
// 支持可选的 token 黑名单功能，用于实现安全的登出机制。
//
// 注意事项：
//   - 通过 WithBlacklist 方法可以启用 token 黑名单功能
//   - 黑名单功能需要 Redis 支持
type AuthService struct {
	jwtService *JWTService      // JWT 服务，用于 token 生成和验证
	userRepo   UserRepository   // 用户数据仓库
	blacklist  TokenBlacklist   // Token 黑名单（可选）
}

// NewAuthService 创建新的认证服务实例。
//
// 该函数初始化 AuthService，绑定 JWT 服务和用户数据仓库。
// 创建的服务默认不启用 token 黑名单功能。
//
// 参数：
//   - jwtService: JWT 服务实例，用于生成和验证 token
//   - userRepo: 用户数据仓库，用于用户数据的存取
//
// 返回值：
//   - 返回初始化完成的 AuthService 实例
//
// 使用示例：
//   jwtSvc := auth.NewJWTService(config)
//   authSvc := auth.NewAuthService(jwtSvc, userRepo)
//
// 注意事项：
//   - 若需启用黑名单功能，需调用 WithBlacklist 方法
func NewAuthService(jwtService *JWTService, userRepo UserRepository) *AuthService {
	return &AuthService{
		jwtService: jwtService,
		userRepo:   userRepo,
	}
}

// WithBlacklist 设置 token 黑名单，启用安全登出功能。
//
// 该方法为 AuthService 配置 token 黑名单，使登出的 token 无法再次使用。
// 需要依赖 Redis 等缓存服务来实现黑名单存储。
//
// 参数：
//   - blacklist: Token 黑名单实现，通常为 RedisTokenBlacklist
//
// 返回值：
//   - 返回配置了黑名单的 AuthService 实例（支持链式调用）
//
// 使用示例：
//   blacklist := auth.NewRedisTokenBlacklist(redisClient)
//   authSvc := auth.NewAuthService(jwtSvc, userRepo).WithBlacklist(blacklist)
//
// 注意事项：
//   - 黑名单功能会增加每次 token 验证时的 Redis 查询开销
//   - 确保 Redis 服务高可用，否则可能影响认证性能
func (s *AuthService) WithBlacklist(blacklist TokenBlacklist) *AuthService {
	s.blacklist = blacklist
	return s
}

// LoginResult 登录结果结构，包含 token 和用户信息。
//
// 该结构用于封装登录成功后返回的数据，便于调用方获取完整认证信息。
type LoginResult struct {
	Token string        // JWT 认证 token，用于后续请求认证
	User  *user.User    // 用户对象，包含用户详细信息
}

// Login 用户登录认证。
//
// 该方法验证用户邮箱和密码，成功后生成 JWT token 并返回用户信息。
// 同时检查用户状态，禁止被封禁用户登录。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - email: 用户邮箱地址
//   - password: 用户密码（明文）
//
// 返回值：
//   - result: 登录结果，包含 token 和用户信息
//   - err: 可能的错误包括：
//       - "invalid credentials": 邮箱或密码错误
//       - "user is banned": 用户被封禁
//       - token 生成失败的其他错误
//
// 使用示例：
//   result, err := authSvc.Login(ctx, "user@example.com", "password")
//   if err != nil {
//       // 处理登录失败
//   }
//   token := result.Token  // 用于后续认证
//
// 注意事项：
//   - 密码验证使用 bcrypt，可抵御暴力破解
//   - 登录失败时返回通用错误消息，避免泄露具体原因
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	// 步骤1: 查找用户
	// 根据邮箱查询用户，不存在时返回通用错误避免枚举攻击
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 步骤2: 检查用户状态
	// 被封禁用户禁止登录，返回明确错误便于前端处理
	if user.Status == UserStatusBanned {
		return nil, errors.New("user is banned")
	}

	// 步骤3: 验证密码
	// 使用 bcrypt 比较哈希值，验证失败返回通用错误
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 步骤4: 生成 token
	// 创建包含用户 ID 和角色 ID 的 JWT token
	token, _, err := s.jwtService.Generate(user.ID, user.RoleID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token: token,
		User:  user,
	}, nil
}

// Register 用户注册。
//
// 该方法创建新用户账户，包括邮箱和用户名唯一性检查、密码哈希处理。
// 新注册用户默认处于待激活状态（Pending）。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - username: 用户名，需唯一
//   - email: 邮箱地址，需唯一
//   - password: 用户密码（明文），将被哈希存储
//
// 返回值：
//   - user: 创建成功的用户对象
//   - err: 可能的错误包括：
//       - "email already registered": 箱已被注册
//       - "username already taken": 用户名已被使用
//       - 密码哈希或数据库操作失败的其他错误
//
// 使用示例：
//   user, err := authSvc.Register(ctx, "newuser", "new@example.com", "password")
//   if err != nil {
//       // 处理注册失败
//   }
//
// 注意事项：
//   - 密码使用 bcrypt.DefaultCost 进行哈希，平衡安全性和性能
//   - 新用户需完成激活流程才能正常使用系统
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*user.User, error) {
	// 步骤1: 检查邮箱唯一性
	// 邮箱重复时返回错误，避免重复注册
	if existing, _ := s.userRepo.GetByEmail(ctx, email); existing != nil {
		return nil, errors.New("email already registered")
	}

	// 步骤2: 检查用户名唯一性
	// 用户名重复时返回错误，确保用户标识唯一
	if existing, _ := s.userRepo.GetByUsername(ctx, username); existing != nil {
		return nil, errors.New("username already taken")
	}

	// 步骤3: 哈希密码
	// 使用 bcrypt 进行安全哈希，防止明文存储
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 步骤4: 创建用户对象
	// 设置初始状态为 Pending，需后续激活
	newUser := &user.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		Status:       user.StatusPending, // 默认待激活状态
	}

	// 步骤5: 持久化用户
	// 保存用户到数据库
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// Logout 用户登出，将 token 加入黑名单。
//
// 该方法解析 token 获取 jti（JWT ID），并将其加入黑名单。
// 黑名单中的 token 在后续验证时会被拒绝。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - token: JWT token 字符串
//
// 返回值：
//   - err: 可能的错误包括：
//       - token 解析失败
//       - 黑名单写入失败
//       - 若未配置黑名单，返回 nil（静默成功）
//
// 使用示例：
//   err := authSvc.Logout(ctx, tokenString)
//   if err != nil {
//       // 处理登出失败
//   }
//
// 注意事项：
//   - 需先通过 WithBlacklist 配置黑名单，否则该方法无效果
//   - 已过期的 token 无需加入黑名单（自动失效）
func (s *AuthService) Logout(ctx context.Context, token string) error {
	// 检查黑名单是否启用
	if s.blacklist == nil {
		return nil // 无黑名单时直接返回
	}

	// 解析 token 获取 claims
	// claims 包含 jti 和过期时间，用于黑名单操作
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return err // token 无效，无需加入黑名单
	}

	// 获取 jti 并加入黑名单
	// 使用过期时间设置 Redis TTL，自动清理过期记录
	jti := claims.GetJWTID()
	if jti == "" {
		return nil // 无 jti，无法加入黑名单
	}
	expiry := claims.ExpiresAt.Time
	return s.blacklist.AddToBlacklist(ctx, jti, expiry)
}

// IsTokenBlacklisted 检查 token 是否在黑名单中。
//
// 该方法验证 token 是否已被注销（加入黑名单）。
// 用于中间件层检查请求携带的 token 是否有效。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - token: JWT token 字符串
//
// 返回值：
//   - true: token 在黑名单中，已被注销
//   - false: token 不在黑名单中，或未配置黑名单
//
// 使用示例：
//   if authSvc.IsTokenBlacklisted(ctx, tokenString) {
//       // 拒绝请求
//   }
//
// 注意事项：
//   - 未配置黑名单时始终返回 false
//   - token 无效或无 jti 时返回 false
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) bool {
	// 检查黑名单是否启用
	if s.blacklist == nil {
		return false
	}

	// 解析 token 获取 jti
	// 无有效 jti 时无法查询黑名单
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

// ValidateToken 验证 token 并返回用户信息。
//
// 该方法执行完整的 token 验证流程：检查黑名单、验证签名、获取用户信息。
// 是认证中间件的核心方法。
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - token: JWT token 字符串
//
// 返回值：
//   - claims: token 中的声明信息，包含用户 ID 和角色 ID
//   - user: 用户对象，包含完整用户信息
//   - err: 可能的错误包括：
//       - "token is blacklisted": token 已被注销
//       - token 签名或格式无效
//       - "user not found": 用户不存在
//       - "user is banned": 用户被封禁
//
// 使用示例：
//   claims, user, err := authSvc.ValidateToken(ctx, tokenString)
//   if err != nil {
//       // 处理验证失败，拒绝请求
//   }
//   // 使用 user.ID 和 user.RoleID 进行后续处理
//
// 注意事项：
//   - 该方法会检查用户状态，封禁用户即使 token 有效也会被拒绝
//   - 验证失败时应返回 401 Unauthorized 响应
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*Claims, *user.User, error) {
	// 步骤1: 检查黑名单
	// 黑名单中的 token 即使签名有效也需拒绝
	if s.IsTokenBlacklisted(ctx, token) {
		return nil, nil, errors.New("token is blacklisted")
	}

	// 步骤2: 验证 token 签名和声明
	// 解析 token 并验证签名完整性
	claims, err := s.jwtService.Validate(token)
	if err != nil {
		return nil, nil, err
	}

	// 步骤3: 获取用户信息
	// 根据 token 中的用户 ID 查询完整用户数据
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, errors.New("user not found")
	}

	// 步骤4: 检查用户状态
	// 封禁用户无法通过认证，需重新验证账户状态
	if user.Status == UserStatusBanned {
		return nil, nil, errors.New("user is banned")
	}

	return claims, user, nil
}