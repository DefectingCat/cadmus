// Package main 提供 Cadmus 博客平台的配置加载功能。
//
// 该文件包含应用配置相关的核心逻辑，包括：
//   - 配置结构体的定义和聚合
//   - 环境变量的读取和默认值处理
//   - JWT 配置的安全加载
//   - 数据库、Redis、服务器等子配置的组装
//
// 主要用途：
//
//	为 Cadmus 服务器提供统一的配置管理，支持通过环境变量灵活配置
//	各组件参数，同时提供合理的默认值。
//
// 注意事项：
//   - 所有配置项均支持环境变量覆盖
//   - JWT 配置加载失败会导致 panic，需确保环境正确配置
//   - 数据库连接池参数已针对生产环境优化
//
// 作者：xfy
package main

import (
	"os"
	"time"

	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/cache"
	"rua.plus/cadmus/internal/database"
)

// Config 应用配置聚合结构体。
//
// 该结构体汇总了 Cadmus 服务器运行所需的所有配置项，
// 包括数据库、缓存、认证、服务器和上传配置。
// 所有配置项均可通过环境变量覆盖默认值。
//
// 使用示例：
//   cfg := loadConfig()
//   pool, err := database.NewPool(ctx, cfg.Database)
type Config struct {
	// Database PostgreSQL 数据库连接配置
	// 包含主机、端口、用户名、密码、连接池参数等
	Database database.Config

	// Redis Redis 缓存服务配置
	// 用于会话存储、限流计数、权限缓存等
	Redis cache.Config

	// JWT JWT 认证配置
	// 包含签名密钥、过期时间、颁发者等
	JWT auth.JWTConfig

	// Server HTTP 服务器配置
	// 包含端口、读写超时、空闲超时等
	Server ServerConfig

	// Upload 文件上传配置
	// 包含存储目录和基础 URL
	Upload UploadConfig
}

// ServerConfig HTTP 服务器配置结构体。
//
// 定义了 HTTP 服务器的运行参数，包括监听端口和各类超时设置。
// 超时参数已针对生产环境优化，防止慢客户端攻击。
type ServerConfig struct {
	// Port 服务器监听端口，默认为 "8080"
	Port string

	// ReadTimeout 读取请求超时时间，默认 15 秒
	// 防止客户端缓慢发送请求头占用连接
	ReadTimeout time.Duration

	// WriteTimeout 写入响应超时时间，默认 15 秒
	// 防止响应写入过程阻塞过久
	WriteTimeout time.Duration

	// IdleTimeout 连接空闲超时时间，默认 60 秒
	// 用于 keep-alive 连接的复用管理
	IdleTimeout time.Duration
}

// UploadConfig 文件上传配置结构体。
//
// 定义了用户上传媒体文件的存储位置和访问 URL。
type UploadConfig struct {
	// Dir 文件存储目录路径，默认为 "./uploads"
	// 确保该目录存在且有写入权限
	Dir string

	// BaseURL 文件访问的基础 URL
	// 用于生成媒体文件的公开访问链接
	BaseURL string
}

// loadConfig 从环境变量加载并组装应用配置。
//
// 该函数读取所有必要的环境变量，为各个组件构建配置对象。
// 对于未设置的环境变量，使用合理的默认值替代。
//
// 返回值：
//   - *Config: 完整的应用配置对象，包含数据库、Redis、JWT、服务器和上传配置
//
// 注意事项：
//   - JWT 配置加载失败会导致 panic
//   - 数据库连接池参数已针对生产环境预设优化值
//   - Redis 连接池参数已针对高频访问场景优化
//
// 环境变量列表：
//   - PORT: 服务器端口（默认 "8080"）
//   - DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, DB_SSLMODE: 数据库配置
//   - REDIS_HOST, REDIS_PORT, REDIS_PASSWORD: Redis 配置
//   - UPLOAD_DIR, BASE_URL: 上传配置
func loadConfig() *Config {
	port := getEnvOrDefault("PORT", "8080")

	// JWT 配置需要单独加载，因为可能失败
	jwtCfg, err := loadJWTConfig()
	if err != nil {
		// 在调用方处理错误
		panic("JWT config error: " + err.Error())
	}

	return &Config{
		Database: database.Config{
			Host:            getEnvOrDefault("DB_HOST", "localhost"),
			Port:            atoi(getEnvOrDefault("DB_PORT", "5432")),
			Name:            getEnvOrDefault("DB_NAME", "cadmus"),
			User:            getEnvOrDefault("DB_USER", "cadmus"),
			Password:        getEnvOrDefault("DB_PASSWORD", ""),
			SSLMode:         getEnvOrDefault("DB_SSLMODE", "disable"),
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		Redis: cache.Config{
			Host:         getEnvOrDefault("REDIS_HOST", "localhost"),
			Port:         atoi(getEnvOrDefault("REDIS_PORT", "6379")),
			Password:     getEnvOrDefault("REDIS_PASSWORD", ""),
			DB:           0,
			PoolSize:     25,
			MinIdleConns: 5,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		JWT: jwtCfg,
		Server: ServerConfig{
			Port:         port,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Upload: UploadConfig{
			Dir:     getEnvOrDefault("UPLOAD_DIR", "./uploads"),
			BaseURL: getEnvOrDefault("BASE_URL", "http://localhost:"+port),
		},
	}
}

// loadJWTConfig 加载 JWT 认证配置。
//
// 该函数从内部 auth 包加载默认的 JWT 配置。
// JWT 密钥等敏感信息应通过环境变量或安全存储提供。
//
// 返回值：
//   - auth.JWTConfig: JWT 配置对象，包含签名密钥、过期时间等
//   - error: 配置加载错误，当密钥等必要信息缺失时返回
func loadJWTConfig() (auth.JWTConfig, error) {
	return auth.DefaultJWTConfig()
}

// getEnvOrDefault 获取环境变量值，若不存在则返回默认值。
//
// 该函数用于安全地读取环境变量，避免空值导致的配置错误。
// 仅当环境变量为空字符串时才使用默认值。
//
// 参数：
//   - key: 环境变量名，如 "PORT"、"DB_HOST"
//   - defaultValue: 默认值，当环境变量未设置时使用
//
// 返回值：
//   - string: 环境变量值或默认值
//
// 使用示例：
//   port := getEnvOrDefault("PORT", "8080")
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// atoi 将字符串转换为整数。
//
// 该函数实现简单的字符串到整数转换，仅处理数字字符。
// 非数字字符会被跳过，而非返回错误。适用于端口等纯数字字符串。
//
// 参数：
//   - s: 待转换的字符串，应仅包含数字字符
//
// 返回值：
//   - int: 转换后的整数，若字符串无数字则返回 0
//
// 使用示例：
//   port := atoi("5432") // 返回 5432
func atoi(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

// parseIntStr 解析整数字符串。
//
// 该函数将字符串转换为整数，遇到非数字字符时提前终止。
// 与 atoi 不同，该函数返回 error 用于错误处理场景。
//
// 参数：
//   - s: 待解析的字符串
//
// 返回值：
//   - int: 解析得到的整数，目前遇错返回 0
//   - error: 当前实现未返回实际错误，保留参数用于未来扩展
//
// 使用示例：
//   page, err := parseIntStr("20")
func parseIntStr(s string) (int, error) {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return 0, nil
		}
	}
	return n, nil
}