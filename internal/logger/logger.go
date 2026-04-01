// Package logger 提供全局日志器封装实现。
//
// 该文件封装 charmbracelet/log 库，提供统一的日志接口，包括：
//   - 全局日志器实例
//   - 日志级别常量
//   - 便捷日志函数
//   - 环境变量配置支持
//
// 主要用途：
//
//	为系统提供统一的日志接口，支持通过环境变量配置日志级别、
//	调用者信息和时间戳显示。
//
// 环境变量配置：
//   - LOG_LEVEL: 日志级别 (debug/info/warn/error/fatal)
//   - LOG_SHOW_CALLER: 是否显示调用者信息 (true/false)
//   - LOG_SHOW_TIMESTAMP: 是否显示时间戳 (true/false，默认 true)
//
// 使用示例：
//
//	// 在 main.go 中初始化
//	logger.Setup()
//
//	// 使用日志函数
//	logger.Info("服务启动")
//	logger.Debug("调试信息", "key", value)
//	logger.Error("发生错误")
//
// 注意事项：
//   - 必须在程序启动时调用 Setup() 初始化
//   - 所有日志函数都是并发安全的
//   - Fatal 系列函数会退出程序
//
// 作者：xfy
package logger

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

// 全局日志器实例。
var (
	// defaultLogger 默认日志器实例
	// 在 Setup() 中初始化
	defaultLogger *log.Logger
)

// Level 日志级别类型别名。
//
// 复用 charmbracelet/log 的 Level 类型。
type Level = log.Level

// 日志级别常量。
const (
	// DebugLevel 调试级别，用于开发调试信息
	DebugLevel = log.DebugLevel

	// InfoLevel 信息级别，用于常规运行信息
	InfoLevel = log.InfoLevel

	// WarnLevel 警告级别，用于潜在问题提示
	WarnLevel = log.WarnLevel

	// ErrorLevel 错误级别，用于错误报告
	ErrorLevel = log.ErrorLevel

	// FatalLevel 致命级别，用于致命错误（会退出程序）
	FatalLevel = log.FatalLevel
)

// Setup 初始化全局日志器。
//
// 从环境变量读取配置，初始化全局日志器实例。
// 应在程序启动时调用一次。
//
// 配置项：
//   - LOG_LEVEL: 日志级别，默认 "info"
//   - LOG_SHOW_CALLER: 显示调用者信息，默认 false
//   - LOG_SHOW_TIMESTAMP: 显示时间戳，默认 true
//
// 使用示例：
//
//	func main() {
//	    logger.Setup()
//	    // ... 其他初始化
//	}
func Setup() {
	defaultLogger = log.New(os.Stderr)

	// 配置日志级别
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if levelStr == "" {
		levelStr = "info"
	}

	var level log.Level
	switch levelStr {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn", "warning":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	case "fatal":
		level = log.FatalLevel
	default:
		level = log.InfoLevel
	}
	defaultLogger.SetLevel(level)

	// 配置调用者显示
	showCaller := os.Getenv("LOG_SHOW_CALLER")
	if showCaller == "true" || showCaller == "1" {
		defaultLogger.SetReportCaller(true)
	}

	// 配置时间戳显示
	showTimestamp := os.Getenv("LOG_SHOW_TIMESTAMP")
	if showTimestamp != "false" && showTimestamp != "0" {
		defaultLogger.SetReportTimestamp(true)
	}
}

// Debug 记录调试级别日志。
//
// 用于输出开发调试信息，生产环境通常不显示。
//
// 参数：
//   - msg: 日志消息，支持任意类型
func Debug(msg interface{}) {
	defaultLogger.Debug(msg)
}

// Info 记录信息级别日志。
//
// 用于输出常规运行信息，如服务启动、请求完成等。
//
// 参数：
//   - msg: 日志消息，支持任意类型
func Info(msg interface{}) {
	defaultLogger.Info(msg)
}

// Warn 记录警告级别日志。
//
// 用于输出潜在问题警告，不影响程序运行但需要关注。
//
// 参数：
//   - msg: 日志消息，支持任意类型
func Warn(msg interface{}) {
	defaultLogger.Warn(msg)
}

// Error 记录错误级别日志。
//
// 用于输出错误信息，程序可继续运行但需要处理。
//
// 参数：
//   - msg: 日志消息，支持任意类型
func Error(msg interface{}) {
	defaultLogger.Error(msg)
}

// Fatal 记录致命级别日志并退出程序。
//
// 用于输出致命错误信息，记录后程序立即退出。
// 仅用于不可恢复的错误场景。
//
// 参数：
//   - msg: 日志消息，支持任意类型
func Fatal(msg interface{}) {
	defaultLogger.Fatal(msg)
}

// Fatalf 记录格式化致命日志并退出程序。
//
// 支持格式化字符串，记录后程序立即退出。
//
// 参数：
//   - format: 格式化字符串
//   - args: 格式化参数
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// Printf 记录信息级别的格式化日志。
//
// 提供与标准库 log.Printf 兼容的接口。
//
// 参数：
//   - format: 格式化字符串
//   - args: 格式化参数
func Printf(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Println 记录信息级别的日志（带换行）。
//
// 提供与标准库 log.Println 兼容的接口。
// 多个参数会用空格连接。
//
// 参数：
//   - args: 日志内容，支持多个参数
func Println(args ...interface{}) {
	msg := ""
	for i, arg := range args {
		if i > 0 {
			msg += " "
		}
		msg += toString(arg)
	}
	defaultLogger.Info(msg)
}

// toString 将值转换为字符串。
//
// 支持字符串、字节切片和其他类型。
// 其他类型返回空字符串。
//
// 参数：
//   - v: 待转换的值
//
// 返回值：
//   - 转换后的字符串
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return ""
	}
}

// SetLevel 动态设置日志级别。
//
// 在运行时修改全局日志器的级别。
//
// 参数：
//   - level: 新的日志级别
//
// 使用示例：
//
//	logger.SetLevel(logger.DebugLevel) // 开启调试日志
func SetLevel(level log.Level) {
	defaultLogger.SetLevel(level)
}

// GetLevel 获取当前日志级别。
//
// 返回当前全局日志器的日志级别。
//
// 返回值：
//   - 当前日志级别
func GetLevel() log.Level {
	return defaultLogger.GetLevel()
}
