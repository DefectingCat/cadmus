package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims JWT claims 结构
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
	jwt.RegisteredClaims
}

// JWTService JWT 认证服务
type JWTService struct {
	Config JWTConfig
}

// NewJWTService 创建新的 JWT 服务
func NewJWTService(config JWTConfig) *JWTService {
	return &JWTService{Config: config}
}

// Generate 生成 JWT token
func (s *JWTService) Generate(userID, roleID uuid.UUID) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.Config.Expiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Config.Secret))
}

// Validate 验证 JWT token 并返回 claims
func (s *JWTService) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.Config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// Refresh 刷新 JWT token
func (s *JWTService) Refresh(tokenString string) (string, error) {
	claims, err := s.Validate(tokenString)
	if err != nil {
		return "", err
	}

	// 检查 token 是否在可刷新范围内
	now := time.Now()
	if claims.ExpiresAt != nil {
		// 允许在过期后 refresh_expiry 时间内刷新
		expiryTime := claims.ExpiresAt.Time
		if now.After(expiryTime.Add(s.Config.RefreshExpiry)) {
			return "", errors.New("token refresh window expired")
		}
	}

	return s.Generate(claims.UserID, claims.RoleID)
}