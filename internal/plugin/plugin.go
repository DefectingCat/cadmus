// Package plugin 提供插件引擎核心模块的实现。
//
// 该文件定义插件系统的核心接口和类型，包括：
//   - Plugin: 插件必须实现的接口
//   - PluginInfo: 插件元信息结构
//   - PluginContext: 插件运行时上下文
//   - PluginConstructor: 插件构造函数类型
//
// 主要用途：
//
//	定义编译时接口注册模式的插件系统，支持通过 blank import
//	在编译时自动注册插件，避免使用 Go 原生 plugin 包的跨平台兼容性问题。
//
// 设计理念：
//
//	参考 Alist 的生产验证实现，使用接口注册模式替代动态加载，
//	确保类型安全和跨平台一致性。
//
// 使用示例：
//
//	// 1. 在插件包中定义实现
//	type MyPlugin struct{}
//
//	func (p *MyPlugin) Info() plugin.PluginInfo {
//	    return plugin.PluginInfo{
//	        ID: "my-plugin",
//	        Name: "我的插件",
//	        Version: "1.0.0",
//	    }
//	}
//
//	func (p *MyPlugin) Init(ctx *plugin.PluginContext) error {
//	    // 使用 ctx.DB、ctx.Cache 等
//	    return nil
//	}
//
//	// 2. 通过 init() 自动注册
//	func init() {
//	    plugin.RegisterPlugin(func() plugin.Plugin {
//	        return &MyPlugin{}
//	    })
//	}
//
//	// 3. 在 main.go 中 blank import 触发注册
//	import _ "rua.plus/cadmus/plugins/my-plugin"
//
// 作者：xfy
package plugin

import (
	"rua.plus/cadmus/internal/cache"
	"rua.plus/cadmus/internal/database"
	"rua.plus/cadmus/internal/services"
)

// Plugin 插件接口。
//
// 所有插件必须实现此接口。接口定义了插件的基本生命周期方法：
//   - Info(): 返回插件元信息，用于展示和依赖管理
//   - Init(): 初始化插件，接收运行时上下文
//
// 设计说明：
//   使用构造函数模式（PluginConstructor）而非直接实例化，
//   便于在注册时验证元信息，在初始化时注入依赖。
type Plugin interface {
	// Info 返回插件元信息。
	//
	// 该方法在插件注册时调用，用于：
	//   - 验证插件 ID 唯一性
	//   - 检查依赖关系
	//   - 展示插件列表
	//
	// 返回值：
	//   - PluginInfo: 插件的元信息
	Info() PluginInfo

	// Init 初始化插件。
	//
	// 在插件注册后或应用启动时调用，接收运行时上下文。
	// 插件应在此方法中完成：
	//   - 注册自定义路由
	//   - 初始化数据库表或缓存
	//   - 设置定时任务
	//   - 注册自定义块类型（待 BlockTypeRegistry 实现）
	//
	// 参数：
	//   - ctx: 插件运行时上下文，包含数据库、缓存、服务等
	//
	// 返回值：
	//   - err: 初始化失败时返回错误，会导致注册失败
	Init(ctx *PluginContext) error
}

// PluginInfo 插件元信息。
//
// 包含插件的基本信息和依赖声明。
type PluginInfo struct {
	// ID 插件唯一标识符。
	// 建议使用反向域名格式，如 "com.example.my-plugin"
	// 或简短的 kebab-case 格式，如 "github-auth"
	ID string

	// Name 插件显示名称。
	// 用于在管理界面展示，如 "GitHub OAuth 登录"
	Name string

	// Version 插件版本号。
	// 建议使用语义化版本，如 "1.0.0"
	Version string

	// Author 插件作者。
	// 可以是个人名称或组织名称
	Author string

	// Description 插件功能描述。
	// 详细说明插件的作用和使用场景
	Description string

	// Dependencies 依赖的其他插件 ID 列表。
	// 系统会在注册时检查依赖是否存在，缺失依赖会记录警告
	Dependencies []string
}

// PluginContext 插件运行时上下文。
//
// 提供插件访问系统资源的能力，通过依赖注入实现解耦。
// 系统在初始化插件时注入此上下文。
type PluginContext struct {
	// DB 数据库连接池。
	// 插件可使用此连接进行数据持久化
	DB *database.Pool

	// Cache 缓存服务。
	// 插件可使用缓存提升性能
	Cache cache.CacheService

	// Registry 块类型注册表。
	// 待 block 模块实现后，用于注册自定义块类型
	// 当前使用 any 类型占位，后续替换为具体类型
	Registry any

	// Services 业务服务容器。
	// 提供对核心业务服务的访问，如用户服务、文章服务等
	Services *services.Container

	// Config 插件配置。
	// 从配置文件加载的插件特定配置项
	Config map[string]any
}

// PluginConstructor 插件构造函数。
//
// 用于创建插件实例的工厂函数。采用函数类型而非直接使用 Plugin 接口，
// 便于在注册时延迟实例化，支持依赖检查后再创建实例。
//
// 使用示例：
//   plugin.RegisterPlugin(func() plugin.Plugin {
//       return &MyPlugin{}
//   })
type PluginConstructor func() Plugin