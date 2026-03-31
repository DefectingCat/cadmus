// Package defaulttheme 默认主题实现
package defaulttheme

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"rua.plus/cadmus/internal/theme"
	"rua.plus/cadmus/web/templates/layouts"
	"rua.plus/cadmus/web/templates/pages"
)

// DefaultTheme 默认主题实现
type DefaultTheme struct{}

// 确保实现 ThemeComponents 接口
var _ theme.ThemeComponents = (*DefaultTheme)(nil)

// New 创建默认主题
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

// Layout 整体布局框架
func (t *DefaultTheme) Layout() templ.Component {
	return layouts.BaseLayout("Cadmus")
}

// Header 头部组件
func (t *DefaultTheme) Header() templ.Component {
	return layouts.Header()
}

// Footer 底部组件
func (t *DefaultTheme) Footer() templ.Component {
	return layouts.Footer()
}

// PostList 文章列表页
func (t *DefaultTheme) PostList() templ.Component {
	// 返回空列表作为默认
	return pages.PostListPage("文章列表", nil, pages.Pagination{}, nil, nil)
}

// PostDetail 文章详情页
func (t *DefaultTheme) PostDetail() templ.Component {
	// 返回空详情作为默认
	return pages.PostDetailPage(nil, nil, nil, nil)
}

// CategoryPage 分类页
func (t *DefaultTheme) CategoryPage() templ.Component {
	// 暂时返回 nil，后续可扩展
	return nil
}

// TagPage 标签页
func (t *DefaultTheme) TagPage() templ.Component {
	// 暂时返回 nil，后续可扩展
	return nil
}

// Sidebar 侧边栏
func (t *DefaultTheme) Sidebar() templ.Component {
	// 暂时返回 nil，侧边栏已集成在页面组件中
	return nil
}

// ErrorPage 错误页面
func (t *DefaultTheme) ErrorPage() templ.Component {
	return errorPage()
}

// errorPage 简单的错误页面组件
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

func init() {
	// 自动注册到全局主题注册表
	if err := theme.Register(New()); err != nil {
		// 注册失败时打印日志（主题已存在是正常情况）
		// 在实际项目中可以使用 log 包
	}
}