// Package defaulttheme 提供 Cadmus 默认主题的实现。
//
// 该文件实现了 Cadmus 博客平台的默认主题，使用 Tailwind CSS 样式。
// 主题提供各个页面组件的渲染实现。
//
// 功能特性：
//   - 使用 Tailwind CSS 提供现代化样式
//   - 响应式设计，支持移动端和桌面端
//   - 支持 templ 模板的类型安全渲染
//
// 启用方式：
//
// 通过 blank import 触发主题注册：
//
//	import _ "rua.plus/cadmus/themes/default"
//
// 注意事项：
//   - 主题在 init() 函数中自动注册到全局注册表
//   - 主题 ID 为 "default"
//
// 作者：xfy
package defaulttheme

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"rua.plus/cadmus/internal/theme"
	"rua.plus/cadmus/web/templates/layouts"
	"rua.plus/cadmus/web/templates/pages"
)

// DefaultTheme 默认主题实现。
//
// 实现 theme.ThemeComponents 接口，提供各个页面组件的渲染方法。
type DefaultTheme struct{}

// 确保实现 ThemeComponents 接口。
//
// 编译时检查，如果 DefaultTheme 未实现 ThemeComponents 接口的所有方法，
// 编译将失败。
var _ theme.ThemeComponents = (*DefaultTheme)(nil)

// New 创建默认主题实例。
//
// 返回主题元信息和组件实现的完整 Theme 对象。
//
// 返回值：
//   - Theme: 主题对象，包含 ID、名称、版本、作者、描述和组件实现
func New() theme.Theme {
	return theme.Theme{
		ID:          "default",
		Name:        "Cadmus Default",
		Version:     "1.0.0",
		Author:      "Cadmus Team",
		Description: "Cadmus 默认主题，使用 Tailwind CSS 样式",
		Components:  &DefaultTheme{},
	}
}

// Layout 整体布局框架。
//
// 返回页面的基础 HTML 结构组件。
//
// 返回值：
//   - templ.Component: 布局组件
func (t *DefaultTheme) Layout() templ.Component {
	return layouts.BaseLayout("Cadmus")
}

// Header 头部组件。
//
// 返回网站顶部导航栏组件，包含 logo、菜单、用户信息等。
//
// 返回值：
//   - templ.Component: 头部组件
func (t *DefaultTheme) Header() templ.Component {
	return layouts.Header()
}

// Footer 底部组件。
//
// 返回网站底部区域组件，包含版权信息、链接等。
//
// 返回值：
//   - templ.Component: 底部组件
func (t *DefaultTheme) Footer() templ.Component {
	return layouts.Footer()
}

// PostList 文章列表页。
//
// 返回文章列表页面组件，用于展示文章列表。
// 默认返回空列表作为占位。
//
// 返回值：
//   - templ.Component: 文章列表组件
func (t *DefaultTheme) PostList() templ.Component {
	// 返回空列表作为默认
	return pages.PostListPage("文章列表", nil, pages.Pagination{}, nil, nil)
}

// PostDetail 文章详情页。
//
// 返回文章详情页面组件，用于展示单篇文章内容。
// 默认返回空详情作为占位。
//
// 返回值：
//   - templ.Component: 文章详情组件
func (t *DefaultTheme) PostDetail() templ.Component {
	// 返回空详情作为默认
	return pages.PostDetailPage(nil, nil, nil, nil)
}

// CategoryPage 分类页。
//
// 返回分类页面组件，用于展示分类列表或分类下的文章。
// 当前返回 nil，表示暂未实现。
//
// 返回值：
//   - templ.Component: 分类页组件，当前为 nil
func (t *DefaultTheme) CategoryPage() templ.Component {
	// 暂时返回 nil，后续可扩展
	return nil
}

// TagPage 标签页。
//
// 返回标签页面组件，用于展示标签云或标签下的文章。
// 当前返回 nil，表示暂未实现。
//
// 返回值：
//   - templ.Component: 标签页组件，当前为 nil
func (t *DefaultTheme) TagPage() templ.Component {
	// 暂时返回 nil，后续可扩展
	return nil
}

// Sidebar 侧边栏。
//
// 返回侧边栏组件，用于展示分类、标签、归档等信息。
// 当前返回 nil，侧边栏已集成在页面组件中。
//
// 返回值：
//   - templ.Component: 侧边栏组件，当前为 nil
func (t *DefaultTheme) Sidebar() templ.Component {
	// 暂时返回 nil，侧边栏已集成在页面组件中
	return nil
}

// ErrorPage 错误页面。
//
// 返回错误页面组件，用于展示 404、500 等错误。
//
// 返回值：
//   - templ.Component: 错误页组件
func (t *DefaultTheme) ErrorPage() templ.Component {
	return errorPage()
}

// errorPage 简单的错误页面组件。
//
// 提供基本的错误页面 HTML，显示 500 错误信息。
// 使用内联 HTML 而非 templ 模板，简化实现。
//
// 返回值：
//   - templ.Component: 错误页组件
func errorPage() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>错误 - Cadmus</title>
	<link rel="stylesheet" href="/static/dist/main.css">
</head>
<body class="min-h-screen bg-gray-50 flex items-center justify-center">
	<div class="text-center">
		<h1 class="text-6xl font-bold text-gray-900 mb-4">500</h1>
		<p class="text-xl text-gray-600 mb-8">服务器内部错误</p>
		<a href="/" class="text-blue-600 hover:text-blue-800">返回首页</a>
	</div>
</body>
</html>`))
		return err
	})
}

// init 在编译时自动注册主题。
//
// 通过 blank import 触发此函数执行，将主题注册到全局注册表。
// 如果注册失败（如主题已存在），错误会被忽略。
func init() {
	// 自动注册到全局主题注册表
	if err := theme.Register(New()); err != nil {
		// 注册失败时打印日志（主题已存在是正常情况）
		// 在实际项目中可以使用 log 包
	}
}