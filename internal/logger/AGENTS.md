<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# internal/logger

## 用途

`logger` 目录提供全局日志器封装实现，基于 `charmbracelet/log` 库构建。为整个系统提供统一的日志接口，支持通过环境变量配置日志级别、调用者信息和时间戳显示。

## 关键文件

| 文件 | 功能 |
|------|------|
| `logger.go` | 全局日志器封装：日志器实例、级别常量、便捷函数、环境变量配置 |

## 组件架构

```
┌─────────────────────────────────────────────────────────────┐
│                     应用程序启动                             │
│                        main()                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    logger.Setup() 初始化
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   defaultLogger                             │
│              (基于 charmbracelet/log)                        │
│  - 日志级别：LOG_LEVEL (默认 info)                           │
│  - 调用者显示：LOG_SHOW_CALLER (默认 false)                   │
│  - 时间戳显示：LOG_SHOW_TIMESTAMP (默认 true)                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 全局日志函数                                  │
│  Debug() / Info() / Warn() / Error() / Fatal()              │
└─────────────────────────────────────────────────────────────┘
```

## 环境变量配置

| 变量 | 说明 | 默认值 | 可选值 |
|------|------|--------|--------|
| `LOG_LEVEL` | 日志级别 | `info` | `debug`, `info`, `warn`, `error`, `fatal` |
| `LOG_SHOW_CALLER` | 显示调用者信息 | `false` | `true`, `false`, `1`, `0` |
| `LOG_SHOW_TIMESTAMP` | 显示时间戳 | `true` | `true`, `false`, `1`, `0` |

## 日志级别

| 级别 | 常量 | 用途 |
|------|------|------|
| DEBUG | `DebugLevel` | 开发调试信息，生产环境通常不显示 |
| INFO | `InfoLevel` | 常规运行信息（服务启动、请求完成） |
| WARN | `WarnLevel` | 潜在问题警告，不影响程序运行 |
| ERROR | `ErrorLevel` | 错误信息，程序可继续运行 |
| FATAL | `FatalLevel` | 致命错误，记录后立即退出程序 |

## 使用指南

### 初始化

```go
// 在 main.go 中初始化
func main() {
    logger.Setup()
    // ... 其他初始化
}
```

### 基础用法

```go
// 信息级别日志
logger.Info("服务启动")

// 调试级别日志
logger.Debug("调试信息")

// 警告级别日志
logger.Warn("潜在问题")

// 错误级别日志
logger.Error("发生错误")

// 致命级别日志（会退出程序）
logger.Fatal("致命错误")
```

### 格式化输出

```go
// 格式化信息日志
logger.Printf("用户 %s 登录", username)

// 格式化致命日志
logger.Fatalf("配置错误：%s", configError)

// Println 风格（多个参数用空格连接）
logger.Println("服务运行于端口", port)
```

### 动态级别控制

```go
// 运行时切换日志级别
logger.SetLevel(logger.DebugLevel)  // 开启调试模式

// 获取当前级别
currentLevel := logger.GetLevel()
```

## 注意事项

1. **必须初始化**: 程序启动时必须调用 `logger.Setup()`，否则 `defaultLogger` 为 `nil`
2. **并发安全**: 所有日志函数底层使用 charmbracelet/log，保证并发安全
3. **Fatal 行为**: `Fatal` 和 `Fatalf` 记录日志后会调用 `os.Exit(1)`
4. **类型安全**: 日志函数支持任意类型参数，底层由 charmbracelet/log 处理

## 依赖关系

```
logger/
    └── 依赖 charmbracelet/log (外部库)
```

`logger` 被以下模块依赖：
- `api/` - HTTP 请求处理中的日志记录
- `services/` - 业务逻辑执行日志
- `database/` - 数据库操作日志
