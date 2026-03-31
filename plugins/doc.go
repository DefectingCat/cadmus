// Package plugins Cadmus 插件集合
//
// 此目录包含可选插件的实现，通过编译时注册模式集成到主程序。
// 启用插件需要在 cmd/server/main.go 中添加 blank import。
//
// 示例：
//
//	import (
//	    _ "rua.plus/cadmus/plugins/mermaid-block"
//	)
package plugins