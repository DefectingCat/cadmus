// Package auth 提供了 Cadmus 认证和授权功能的核心实现。
//
// 该文件包含 JWT（JSON Web Token）相关的核心逻辑，包括：
//   - JWT token 的生成和签名
//   - JWT token 的验证和解析
//   - JWT token 的刷新机制
//   - 自定义 Claims 结构定义
//
// 主要用途：
//
//	用于实现无状态的用户认证，通过 JWT token 标识用户身份和权限。
//
// 注意事项：
//   - 使用 HMAC-SHA256 算法进行签名
//   - Token 包含唯一 jti，支持黑名单机制
//   - Secret 密钥必须保密且长度足够（至少 32 字符）
//
// 作者：xfy
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims JWT token 的自定义声明结构。
//
// 该结构扩展了标准的 JWT RegisteredClaims，添加了用户相关的自定义字段。
// 用于在 token 中携带用户身份和权限信息。
//
// 字段说明：
//   - UserID: 用户唯一标识符，用于查询用户数据
//   - RoleID: 用户角色标识符，用于权限判断
//   - RegisteredClaims: 标准 JWT 声明，包括过期时间、签发者等
type Claims struct {
	UserID uuid.UUID `json:"user_id"`  // 用户唯一标识符
	RoleID uuid.UUID `json:"role_id"`  // 用户角色标识符
	jwt.RegisteredClaims                 // 标准 JWT 声明
}

// GetJWTID 获取 JWT ID（jti claim）。
//
// jti 是 JWT 的唯一标识符，用于实现 token 黑名单机制。
// 每个生成的 token 都包含唯一的 jti。
//
// 返回值：
//   - 返回 jti 字符串，若未设置则返回空字符串
//
// 使用示例：
//   jti := claims.GetJWTID()
//   if jti != "" {
//       // 使用 jti 进行黑名单操作
//   }
func (c *Claims) GetJWTID() string {
	if c.ID == "" {
		return ""
	}
	return c.ID
}

// JWTService JWT 认证服务，提供 token 生成和验证功能。
//
// 该服务封装了 JWT 的核心操作，使用配置的密钥进行签名和验证。
// 支持生成包含用户信息的 token，以及验证和刷新已有 token。
//
// 注意事项：
//   - 密钥必须保密，泄露将导致认证系统失效
//   - 建议使用强密钥（至少 32 字符）并定期轮换
type JWTService struct {
	Config JWTConfig  // JWT 配置，包含密钥、过期时间等
}

// NewJWTService 创建新的 JWT 服务实例。
//
// 该函数初始化 JWTService，使用提供的配置进行后续操作。
//
// 参数：
//   - config: JWT 配置，包含签名密钥、过期时间、签发者等信息
//
// 返回值：
//   - 返回初始化完成的 JWTService 实例
//
// 使用示例：
//   config := auth.JWTConfig{
//       Secret: "your-secret-key",
//       Expiry: 24 * time.Hour,
//       Issuer: "cadmus",
//   }
//   jwtSvc := auth.NewJWTService(config)
//
// 注意事项：
//   - 密钥长度建议至少 32 字符
//   - 生产环境密钥应从环境变量或密钥管理服务获取
func NewJWTService(config JWTConfig) *JWTService {
	return &JWTService{Config: config}
}

// Generate 生成 JWT token，包含唯一 jti。
//
// 该方法创建一个新的 JWT token，包含用户 ID、角色 ID 和标准声明。
// 每个 token 都分配唯一的 jti，支持后续的黑名单管理。
//
// 参数：
//   - userID: 用户唯一标识符
//   - roleID: 用户角色标识符
//
// 返回值：
//   - tokenString: 签名后的 JWT token 字符串
//   - jti: token 的唯一标识符，用于黑名单操作
//   - err: 生成失败时返回错误
//
// 使用示例：
//   token, jti, err := jwtSvc.Generate(user.ID, user.RoleID)
//   if err != nil {
//       // 处理生成失败
//   }
//   // 存储 jti 用于后续追踪或黑名单操作
//
// 注意事项：
//   - Token 使用 HMAC-SHA256 签名
//   - jti 使用 UUID 格式，保证全局唯一性
func (s *JWTService) Generate(userID, roleID uuid.UUID) (string, string, error) {
	now := time.Now()
	jti := uuid.New().String() // 生成唯一 JWT ID

	// 构建 claims 结构
	// 包含用户标识、角色标识和标准时间声明
	claims := Claims{
		UserID: userID,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,                                    // 设置 jti claim
			Issuer:    s.Config.Issuer,                        // 签发者标识
			IssuedAt:  jwt.NewNumericDate(now),                // 签发时间
			ExpiresAt: jwt.NewNumericDate(now.Add(s.Config.Expiry)),  // 过期时间
			NotBefore: jwt.NewNumericDate(now),                // 生效时间
		},
	}

	// 使用 HMAC-SHA256 算法签名
	// HS256 是对称加密，验证时使用相同密钥
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.Config.Secret))
	return tokenString, jti, err
}

// Validate 验证 JWT token 并返回 claims。
//
// 该方法解析 token 字符串，验证签名和声明有效性。
// 检查签名算法、过期时间、生效时间等标准声明。
//
// 参数：
//   - tokenString: JWT token 字符串
//
// 返回值：
//   - claims: 解析后的声明信息，包含用户 ID 和角色 ID
//   - err: 可能的错误包括：
//       - "unexpected signing method": 签名算法不匹配
//       - "invalid token claims": claims 解析失败
//       - token 过期或格式无效
//
// 使用示例：
//   claims, err := jwtSvc.Validate(tokenString)
//   if err != nil {
//       // 处理验证失败
//   }
//   userID := claims.UserID  // 获取用户 ID
//
// 注意事项：
//   - 验证时会检查 token 是否过期
//   - 确保密钥与生成时使用的密钥一致
func (s *JWTService) Validate(tokenString string) (*Claims, error) {
	// 使用密钥验证签名
	// 同时检查签名算法是否为 HMAC
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.Config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 提取并验证 claims
	// 确保 claims 类型正确且 token 有效
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// Refresh 刷新 JWT token。
//
// 该方法在原有 token 基础上生成新的 token，延长认证有效期。
// 支持在 token 过期后一定时间内仍可刷新（由 RefreshExpiry 配置）。
//
// 参数：
//   - tokenString: 原 JWT token 字符串（可能已过期）
//
// 返回值：
//   - tokenString: 新的 JWT token 字符串
//   - jti: 新 token 的唯一标识符
//   - err: 可能的错误包括：
//       - 原 token 无效或格式错误
//       - "token refresh window expired": 超出刷新时间窗口
//
// 使用示例：
//   newToken, newJti, err := jwtSvc.Refresh(oldToken)
//   if err != nil {
//       // 处理刷新失败，可能需要重新登录
//   }
//
// 注意事项：
//   - 刷新后原 token 不会自动失效，建议配合黑名单机制
//   - 刷新时间窗口通常比 token 有效期更长（如 7 天）
func (s *JWTService) Refresh(tokenString string) (string, string, error) {
	// 验证原 token 的签名和结构
	// 即使过期也能解析 claims
	claims, err := s.Validate(tokenString)
	if err != nil {
		return "", "", err
	}

	// 检查刷新时间窗口
	// 允许在过期后 RefreshExpiry 时间内刷新
	now := time.Now()
	if claims.ExpiresAt != nil {
		// 计算刷新截止时间：过期时间 + 刷新窗口
		expiryTime := claims.ExpiresAt.Time
		if now.After(expiryTime.Add(s.Config.RefreshExpiry)) {
			return "", "", errors.New("token refresh window expired")
		}
	}

	// 使用原 claims 中的用户信息生成新 token
	return s.Generate(claims.UserID, claims.RoleID)
}