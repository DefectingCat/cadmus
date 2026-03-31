// Package theme 主题引擎模块
package theme

import (
	"github.com/a-h/templ"
)

// Theme 主题元信息
type Theme struct {
	ID          string          // 主题唯一标识
	Name        string          // 主题名称
	Version     string          // 版本号
	Author      string          // 作者
	Description string          // 描述
	Components  ThemeComponents // 主题组件实现
}

// ThemeComponents 主题必须实现的组件接口
type ThemeComponents interface {
	Layout() templ.Component       // 整体布局框架
	Header() templ.Component       // 头部
	Footer() templ.Component       // 底部
	PostList() templ.Component     // 文章列表页
	PostDetail() templ.Component   // 文章详情页
	CategoryPage() templ.Component // 分类页
	TagPage() templ.Component      // 标签页
	Sidebar() templ.Component      // 侧边栏（可选）
	ErrorPage() templ.Component    // 错误页
}

// 常见错误定义
var (
	ErrThemeNotFound      = &ThemeError{Code: "theme_not_found", Message: "主题不存在"}
	ErrThemeAlreadyExists = &ThemeError{Code: "theme_already_exists", Message: "主题已存在"}
	ErrNoActiveTheme      = &ThemeError{Code: "no_active_theme", Message: "未设置激活主题"}
	ErrInvalidThemeID     = &ThemeError{Code: "invalid_theme_id", Message: "无效的主题ID"}
)

// ThemeError 主题模块自定义错误
type ThemeError struct {
	Code    string
	Message string
}

func (e *ThemeError) Error() string {
	return e.Message
}

// Is 实现 errors.Is 接口
func (e *ThemeError) Is(target error) bool {
	t, ok := target.(*ThemeError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}