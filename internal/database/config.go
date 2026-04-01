// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含数据库配置管理的核心逻辑，包括：
//   - 连接参数的定义和默认值
//   - 连接池参数的配置
//   - DSN 连接字符串的构建
//
// 主要用途：
//
//	用于配置数据库连接参数，支持 YAML 配置文件和代码配置两种方式。
//
// 注意事项：
//   - 默认配置适用于开发环境，生产环境应通过配置文件调整
//   - 连接池参数需要根据实际负载进行调整
//
// 作者：xfy
package database

import "time"

// Config 数据库连接配置结构体。
//
// 包含数据库连接的基础参数（主机、端口、用户等）和
// 连接池配置参数（最大连接数、空闲连接数、超时时间等）。
// 支持 YAML 配置文件映射。
type Config struct {
	// Host 数据库服务器主机地址
	Host string `yaml:"host"`

	// Port 数据库服务器端口
	Port int `yaml:"port"`

	// Name 数据库名称
	Name string `yaml:"name"`

	// User 数据库用户名
	User string `yaml:"user"`

	// Password 数据库密码
	Password string `yaml:"password"`

	// 连接池配置

	// MaxOpenConns 最大打开连接数，控制连接池上限
	MaxOpenConns int `yaml:"max_open_conns"`

	// MaxIdleConns 最大空闲连接数，控制连接池下限
	MaxIdleConns int `yaml:"max_idle_conns"`

	// ConnMaxLifetime 连接最大生命周期，超过此时间的连接会被关闭
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`

	// ConnMaxIdleTime 连接最大空闲时间，超过此时间的空闲连接会被关闭
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`

	// SSL 配置

	// SSLMode SSL 连接模式（disable, require, verify-ca, verify-full）
	SSLMode string `yaml:"ssl_mode"`
}

// DefaultConfig 返回默认数据库配置
func DefaultConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            5432,
		Name:            "cadmus",
		User:            "cadmus",
		Password:        "",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
		SSLMode:         "disable",
	}
}

// DSN 构建数据库连接字符串
func (c Config) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return "host=" + c.Host +
		" port=" + itoa(c.Port) +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.Name +
		" sslmode=" + sslMode
}

// itoa 简单的整数转字符串
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}