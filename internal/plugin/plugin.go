// Package plugin 插件引擎核心模块
//
// 提供编译时接口注册模式的插件系统，参考 Alist 的生产验证实现。
// 不使用 Go 原生 plugin 包，避免跨平台兼容性问题。
package plugin

import (
	"rua.plus/cadmus/internal/cache"
	"rua.plus/cadmus/internal/database"
	"rua.plus/cadmus/internal/services"
)

// Plugin 插件接口 - 所有插件必须实现
type Plugin interface {
	// Info 返回插件元信息
	Info() PluginInfo
	// Init 初始化插件，接收运行时上下文
	Init(ctx *PluginContext) error
}

// PluginInfo 插件元信息
type PluginInfo struct {
	ID          string   // 唯一标识符，如 "github-auth"
	Name        string   // 显示名称，如 "GitHub OAuth 登录"
	Version     string   // 版本号，如 "1.0.0"
	Author      string   // 作者
	Description string   // 功能描述
	Dependencies []string // 依赖的其他插件 ID
}

// PluginContext 插件运行时上下文
type PluginContext struct {
	DB       *database.Pool     // 数据库连接池
	Cache    cache.CacheService // 缓存服务
	Registry any                // BlockTypeRegistry（待 block 模块实现后替换类型）
	Services *services.Container // 业务服务容器
	Config   map[string]any     // 插件配置
}

// PluginConstructor 插件构造函数
type PluginConstructor func() Plugin