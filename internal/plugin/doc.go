// Package plugin 插件引擎核心模块
//
// 提供编译时接口注册模式的插件系统，参考 Alist 的生产验证实现。
// 不使用 Go 原生 plugin 包，避免跨平台兼容性问题。
//
// 核心组件：
//   - Plugin 接口：定义插件必须实现的 Info() 和 Init() 方法
//   - PluginInfo：插件元信息结构体
//   - PluginContext：插件运行时上下文（DB、Cache、Services、Config）
//   - PluginRegistry：全局插件注册表
//
// 使用方式：
//
//	// 在插件包中定义实现
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
//	// 通过 init() 自动注册
//	func init() {
//	    plugin.RegisterPlugin(func() plugin.Plugin {
//	        return &MyPlugin{}
//	    })
//	}
//
//	// 在 main.go 中 blank import 触发注册
//	import _ "rua.plus/cadmus/plugins/my-plugin"
package plugin