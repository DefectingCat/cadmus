// Package mermaidblock 提供 Mermaid 图表块插件的实现。
//
// 该文件实现了支持文章内容中 Mermaid 图表渲染的插件。
// Mermaid 是一种基于文本的图表描述语言，支持流程图、序列图、甘特图等。
//
// 功能特性：
//   - 支持在文章中使用 Mermaid 语法编写图表
//   - 支持流程图、序列图、甘特图、类图等多种图表类型
//   - 通过前端 JavaScript 库渲染图表
//
// 启用方式：
//
// 在 main.go 中添加 blank import 触发插件注册：
//
//	import _ "rua.plus/cadmus/plugins/mermaid-block"
//
// 注意事项：
//   - 需要前端引入 Mermaid JavaScript 库
//   - 当前版本仅完成注册，待 BlockTypeRegistry 实现后注册自定义块类型
//
// 作者：xfy
package mermaidblock

import (
	"rua.plus/cadmus/internal/logger"
	"rua.plus/cadmus/internal/plugin"
)

// MermaidBlockPlugin Mermaid 图表块插件实现。
//
// 实现 plugin.Plugin 接口，提供 Mermaid 图表渲染支持。
type MermaidBlockPlugin struct{}

// Info 返回插件元信息。
//
// 返回值：
//   - PluginInfo: 包含插件 ID、名称、版本、作者、描述和依赖信息
func (p *MermaidBlockPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:          "mermaid-block",
		Name:        "Mermaid 图表块",
		Version:     "1.0.0",
		Author:      "Cadmus Team",
		Description: "支持文章内容中的 Mermaid 图表渲染，包括流程图、序列图、甘特图等",
		Dependencies: []string{}, // 无依赖
	}
}

// Init 初始化插件。
//
// 接收运行时上下文，完成插件的初始化工作。
// 当前版本仅记录初始化日志，待 BlockTypeRegistry 实现后注册自定义块类型。
//
// 参数：
//   - ctx: 插件运行时上下文，包含数据库、缓存、服务等
//
// 返回值：
//   - error: 初始化失败时返回错误，当前始终返回 nil
//
// TODO:
//   当 BlockTypeRegistry 实现后，注册自定义块类型：
//
//	if ctx.Registry != nil {
//	    registry := ctx.Registry.(plugin.BlockTypeRegistry)
//	    registry.Register("mermaid", &MermaidBlockType{})
//	}
func (p *MermaidBlockPlugin) Init(ctx *plugin.PluginContext) error {
	logger.Info("[mermaid-block] Plugin initialized with context")

	// TODO: 当 BlockTypeRegistry 实现后，注册自定义块类型
	// 示例：
	// if ctx.Registry != nil {
	//     registry := ctx.Registry.(plugin.BlockTypeRegistry)
	//     registry.Register("mermaid", &MermaidBlockType{})
	// }

	return nil
}

// init 在编译时自动注册插件。
//
// 通过 blank import 触发此函数执行，将插件注册到全局注册表。
func init() {
	plugin.RegisterPlugin(func() plugin.Plugin {
		return &MermaidBlockPlugin{}
	})
}