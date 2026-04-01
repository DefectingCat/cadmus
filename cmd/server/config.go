package main

import (
	"os"
	"time"

	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/cache"
	"rua.plus/cadmus/internal/database"
)

// Config 应用配置聚合
type Config struct {
	Database database.Config
	Redis    cache.Config
	JWT      auth.JWTConfig
	Server   ServerConfig
	Upload   UploadConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// UploadConfig 上传配置
type UploadConfig struct {
	Dir     string
	BaseURL string
}

// loadConfig 从环境变量加载配置
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

// loadJWTConfig 加载 JWT 配置
func loadJWTConfig() (auth.JWTConfig, error) {
	return auth.DefaultJWTConfig()
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// atoi 字符串转整数
func atoi(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

// parseIntStr 解析整数字符串
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