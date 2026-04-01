// Package theme 提供主题引擎核心模块的实现。
//
// 该文件包含主题系统的核心数据结构，包括：
//   - Theme: 主题元信息结构
//   - ThemeComponents: 主题组件接口
//   - ThemeError: 主题模块错误类型
//
// 主要用途：
//
//	定义主题系统的核心接口和类型，支持多主题切换和自定义主题开发。
//
// 设计理念：
//
//	使用 templ 库实现类型安全的模板渲染，主题通过实现 ThemeComponents
//	接口提供各个页面组件的渲染逻辑。
//
// 注意事项：
//   - 每个主题必须实现 ThemeComponents 接口的所有方法
//   - 主题在 init() 函数中通过 Register() 自动注册
//   - 使用单例模式管理全局主题注册表
//
// 作者：xfy
package theme

import (
	"github.com/a-h/templ"
)

// Theme 主题元信息。
//
// 包含主题的基本信息和组件实现。系统中的每个主题都由 Theme 结构表示。
type Theme struct {
	// ID 主题唯一标识符。
	// 用于主题切换和配置存储，如 "default"、"dark"、"minimal"
	ID string

	// Name 主题显示名称。
	// 用于在管理界面展示，如 "Cadmus Default"、"深色主题"
	Name string

	// Version 主题版本号。
	// 建议使用语义化版本，如 "1.0.0"
	Version string

	// Author 主题作者。
	// 可以是个人名称或组织名称
	Author string

	// Description 主题描述。
	// 说明主题的风格特点和适用场景
	Description string

	// Components 主题组件实现。
	// 必须实现 ThemeComponents 接口的所有方法
	Components ThemeComponents
}

// ThemeComponents 主题必须实现的组件接口。
//
// 定义主题需要提供的各个页面组件。每个方法返回一个 templ 组件，
// 用于渲染对应的页面部分。
//
// 实现要求：
//   - 所有方法必须返回有效的 templ.Component
//   - 组件应该是可复用的，避免在方法内创建新实例
//   - 返回 nil 表示该主题不支持此组件
type ThemeComponents interface {
	// Layout 整体布局框架。
	//
	// 提供页面的基础 HTML 结构，包含 head、header、main、footer 等区域。
	// 其他组件通常嵌入在 Layout 中渲染。
	//
	// 返回值：
	//   - templ.Component: 布局组件
	Layout() templ.Component

	// Header 头部组件。
	//
	// 渲染网站顶部导航栏，包含 logo、菜单、用户信息等。
	//
	// 返回值：
	//   - templ.Component: 头部组件
	Header() templ.Component

	// Footer 底部组件。
	//
	// 渲染网站底部区域，包含版权信息、链接、统计数据等。
	//
	// 返回值：
	//   - templ.Component: 底部组件
	Footer() templ.Component

	// PostList 文章列表页。
	//
	// 渲染文章列表页面，包含文章卡片、分页、筛选等。
	//
	// 返回值：
	//   - templ.Component: 文章列表组件
	PostList() templ.Component

	// PostDetail 文章详情页。
	//
	// 渲染文章详情页面，包含标题、内容、评论等。
	//
	// 返回值：
	//   - templ.Component: 文章详情组件
	PostDetail() templ.Component

	// CategoryPage 分类页。
	//
	// 渲染分类列表或分类下的文章列表页面。
	//
	// 返回值：
	//   - templ.Component: 分类页组件
	CategoryPage() templ.Component

	// TagPage 标签页。
	//
	// 渲染标签云或标签下的文章列表页面。
	//
	// 返回值：
	//   - templ.Component: 标签页组件
	TagPage() templ.Component

	// Sidebar 侧边栏。
	//
	// 渲染侧边栏区域，包含分类、标签、归档、热门文章等。
	// 可选组件，返回 nil 表示主题无侧边栏。
	//
	// 返回值：
	//   - templ.Component: 侧边栏组件，可选
	Sidebar() templ.Component

	// ErrorPage 错误页面。
	//
	// 渲染错误页面，如 404、500 等错误场景。
	//
	// 返回值：
	//   - templ.Component: 错误页组件
	ErrorPage() templ.Component
}

// 常见错误定义。
//
// 使用语义化错误类型，便于调用方进行错误处理和判断。
var (
	// ErrThemeNotFound 主题不存在错误，切换或查询主题时 ID 无效返回
	ErrThemeNotFound = &ThemeError{Code: "theme_not_found", Message: "主题不存在"}

	// ErrThemeAlreadyExists 主题已存在错误，注册主题时 ID 冲突返回
	ErrThemeAlreadyExists = &ThemeError{Code: "theme_already_exists", Message: "主题已存在"}

	// ErrNoActiveTheme 未设置激活主题错误，获取当前主题但未设置时返回
	ErrNoActiveTheme = &ThemeError{Code: "no_active_theme", Message: "未设置激活主题"}

	// ErrInvalidThemeID 无效主题 ID 错误，主题 ID 为空时返回
	ErrInvalidThemeID = &ThemeError{Code: "invalid_theme_id", Message: "无效的主题ID"}
)

// ThemeError 主题模块自定义错误类型。
//
// 实现 error 和 errors.Is 接口，支持错误比较和类型判断。
// 通过 Code 字段区分不同错误类型，Message 字段提供人类可读描述。
type ThemeError struct {
	// Code 错误代码，用于程序化错误判断
	Code string

	// Message 错误消息，用于展示给用户或记录日志
	Message string
}

// Error 实现 error 接口，返回错误消息。
func (e *ThemeError) Error() string {
	return e.Message
}

// Is 实现 errors.Is 接口，支持错误类型比较。
//
// 通过比较 Code 字段判断是否为同类型错误，
// 便于使用 errors.Is(err, ErrThemeNotFound) 进行错误判断。
func (e *ThemeError) Is(target error) bool {
	t, ok := target.(*ThemeError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}