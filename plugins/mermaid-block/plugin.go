// Package mermaidblock Mermaid 图表块插件
//
// 提供文章内容中的 Mermaid 图表渲染支持。
// Mermaid 是一种基于文本的图表描述语言，支持流程图、序列图等。
//
// 启用方式：在 main.go 中添加 blank import
//
//	import _ "rua.plus/cadmus/plugins/mermaid-block"
package mermaidblock

import (
	"log"

	"rua.plus/cadmus/internal/plugin"
)

// MermaidBlockPlugin 实现 plugin.Plugin 接口
type MermaidBlockPlugin struct{}

// Info 返回插件元信息
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

// Init 初始化插件
func (p *MermaidBlockPlugin) Init(ctx *plugin.PluginContext) error {
	log.Printf("[mermaid-block] Plugin initialized with context")

	// TODO: 当 BlockTypeRegistry 实现后，注册自定义块类型
	// 示例：
	// if ctx.Registry != nil {
	//     registry := ctx.Registry.(plugin.BlockTypeRegistry)
	//     registry.Register("mermaid", &MermaidBlockType{})
	// }

	return nil
}

// 在 init() 中注册插件（编译时自动执行）
func init() {
	plugin.RegisterPlugin(func() plugin.Plugin {
		return &MermaidBlockPlugin{}
	})
}