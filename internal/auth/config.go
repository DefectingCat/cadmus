package auth

import "time"

// JWTConfig JWT 配置结构
type JWTConfig struct {
	Secret     string        `yaml:"secret"`
	Expiry     time.Duration `yaml:"expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
	Issuer     string        `yaml:"issuer"`
}

// DefaultJWTConfig 返回默认 JWT 配置
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:        "cadmus-default-secret-change-in-production",
		Expiry:        24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "cadmus",
	}
}