package auth

import (
	"errors"
	"os"
	"time"
)

// JWTConfig JWT 配置结构
type JWTConfig struct {
	Secret        string        `yaml:"secret"`
	Expiry        time.Duration `yaml:"expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
	Issuer        string        `yaml:"issuer"`
}

// DefaultJWTConfig 返回默认 JWT 配置
// Secret 必须从环境变量 JWT_SECRET 读取，否则返回错误
func DefaultJWTConfig() (JWTConfig, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return JWTConfig{}, errors.New("JWT_SECRET environment variable is required")
	}
	if len(secret) < 32 {
		return JWTConfig{}, errors.New("JWT_SECRET must be at least 32 characters for security")
	}

	return JWTConfig{
		Secret:        secret,
		Expiry:        24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "cadmus",
	}, nil
}

// MustJWTConfig 返回 JWT 配置，如果 JWT_SECRET 未设置则 panic
// 仅用于测试或开发环境
func MustJWTConfig() JWTConfig {
	cfg, err := DefaultJWTConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}