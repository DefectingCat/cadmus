// Package notify 邮件通知模块配置
package notify

// SMTPConfig SMTP 配置结构
type SMTPConfig struct {
	Host     string // SMTP 服务器地址
	Port     int    // SMTP 服务器端口
	User     string // SMTP 用户名
	Password string // SMTP 密码
	From     string // 发件人地址
	UseTLS   bool   // 是否使用 TLS
}

// DefaultSMTPConfig 默认 SMTP 配置
func DefaultSMTPConfig() *SMTPConfig {
	return &SMTPConfig{
		Host:   "localhost",
		Port:   25,
		UseTLS: false,
	}
}