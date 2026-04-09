// Package plugin 提供插件引擎核心模块。
//
// 该包实现了编译时接口注册模式的插件系统，参考 Alist 的生产验证实现。
// 不使用 Go 原生 plugin 包，避免跨平台兼容性问题。
//
// # 核心组件
//
// 该包包含以下核心组件：
//   - Plugin 接口：定义插件必须实现的 Info() 和 Init() 方法
//   - PluginInfo：插件元信息结构体
//   - PluginContext：插件运行时上下文（DB、Cache、Services、Config）
//   - PluginRegistry：全局插件注册表
//
// # 使用方式
//
// 在插件包中定义实现：
//
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
// 通过 init() 自动注册：
//
//	func init() {
//	    plugin.RegisterPlugin(func() plugin.Plugin {
//	        return &MyPlugin{}
//	    })
//	}
//
// 在 main.go 中 blank import 触发注册：
//
//	import _ "rua.plus/cadmus/plugins/my-plugin"
//
// # 插件生命周期
//
// 1. 编译时：通过 blank import 触发 init() 注册插件
// 2. 启动时：调用 RegisterPluginWithContext 初始化插件
// 3. 运行时：通过 GetPlugin 获取插件实例
//
// # 依赖管理
//
// 插件可以在 PluginInfo.Dependencies 中声明依赖的其他插件 ID。
// 系统会在注册时检查依赖是否存在，缺失依赖会记录警告日志。
//
// # 设计理念
//
// 使用编译时注册模式的优势：
//   - 类型安全：所有插件实现相同的接口
//   - 跨平台：不依赖 Go 原生 plugin 包
//   - 简单部署：插件编译到主程序，无需额外文件
//   - 易于测试：可以通过 mock 实现进行测试
//
// 作者：xfy
package plugin
