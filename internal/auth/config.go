// Package auth 提供了 Cadmus 认证和授权功能的核心实现。
//
// 该文件包含 JWT 配置相关的核心逻辑，包括：
//   - JWT 配置结构定义
//   - 默认配置生成函数
//   - 安全配置验证
//
// 主要用途：
//
//	用于管理和验证 JWT 认证所需的配置参数。
//
// 注意事项：
//   - JWT Secret 必须从环境变量获取，禁止硬编码
//   - Secret 长度必须至少 32 字符以确保安全性
//   - 生产环境应使用强密钥并定期轮换
//
// 作者：xfy
package auth

import (
	"errors"
	"os"
	"time"
)

// JWTConfig JWT 配置结构，定义 JWT token 的核心参数。
//
// 该结构包含 JWT 生成和验证所需的所有配置项。
// 配置值可通过 YAML 文件或环境变量设置。
//
// 字段说明：
//   - Secret: 签名密钥，必须保密且长度足够
//   - Expiry: Token 有效期，默认 24 小时
//   - RefreshExpiry: 刷新窗口期，默认 7 天
//   - Issuer: 签发者标识，用于 token 验证
//
// 注意事项：
//   - Secret 泄露将导致认证系统完全失效
//   - 建议使用密钥管理服务存储 Secret
type JWTConfig struct {
	Secret        string        `yaml:"secret"`         // 签名密钥，用于 HMAC-SHA256
	Expiry        time.Duration `yaml:"expiry"`         // Token 有效期，默认 24 小时
	RefreshExpiry time.Duration `yaml:"refresh_expiry"` // 刷新窗口期，默认 7 天
	Issuer        string        `yaml:"issuer"`         // 签发者标识，默认 "cadmus"
}

// DefaultJWTConfig 返回默认 JWT 配置。
//
// 该函数从环境变量 JWT_SECRET 读取密钥，并应用默认值。
// 会对密钥长度进行安全验证，确保符合最低安全要求。
//
// 返回值：
//   - config: JWT 配置对象，包含有效配置值
//   - err: 可能的错误包括：
//   - "JWT_SECRET environment variable is required": 未设置环境变量
//   - "JWT_SECRET must be at least 32 characters": 密钥长度不足
//
// 使用示例：
//
//	config, err := auth.DefaultJWTConfig()
//	if err != nil {
//	    // 处理配置错误，可能需要终止程序
//	}
//	jwtSvc := auth.NewJWTService(config)
//
// 注意事项：
//   - 必须设置环境变量 JWT_SECRET
//   - 生产环境密钥建议 64 字符以上
//   - 此函数在应用启动时调用一次即可
func DefaultJWTConfig() (JWTConfig, error) {
	// 从环境变量读取密钥
	// 禁止硬编码密钥，确保安全性
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return JWTConfig{}, errors.New("JWT_SECRET environment variable is required")
	}

	// 验证密钥长度
	// 32 字符是 HMAC-SHA256 的最低安全要求
	// 更长的密钥可提供更强的安全性
	if len(secret) < 32 {
		return JWTConfig{}, errors.New("JWT_SECRET must be at least 32 characters for security")
	}

	// 返回包含默认值的配置
	// 有效期 24 小时，刷新窗口 7 天，符合常见实践
	return JWTConfig{
		Secret:        secret,
		Expiry:        24 * time.Hour,     // Token 有效期：24 小时
		RefreshExpiry: 7 * 24 * time.Hour, // 刷新窗口：7 天
		Issuer:        "cadmus",           // 签发者标识
	}, nil
}

// MustJWTConfig 返回 JWT 配置，如果 JWT_SECRET 未设置则 panic。
//
// 该函数是 DefaultJWTConfig 的简化版本，配置失败时直接 panic。
// 适用于测试环境或开发阶段，简化错误处理。
//
// 返回值：
//   - config: JWT 配置对象
//
// 使用示例：
//
//	// 测试环境
//	config := auth.MustJWTConfig()
//	jwtSvc := auth.NewJWTService(config)
//
// 注意事项：
//   - 仅用于测试或开发环境，生产环境应使用 DefaultJWTConfig
//   - Panic 会导致程序终止，确保仅在启动阶段调用
func MustJWTConfig() JWTConfig {
	cfg, err := DefaultJWTConfig()
	if err != nil {
		// 配置失败时 panic，简化测试代码
		panic(err)
	}
	return cfg
}
