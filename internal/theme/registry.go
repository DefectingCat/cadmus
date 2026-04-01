// Package theme 提供主题注册管理实现。
//
// 该文件包含主题系统的全局注册表和相关操作函数，包括：
//   - ThemeRegistry: 主题管理接口
//   - 全局主题注册表实例
//   - 主题注册和查询的便捷函数
//
// 主要用途：
//
//	管理已注册主题的生命周期，支持主题的注册、切换和遍历。
//
// 注意事项：
//   - 所有注册表操作都是并发安全的
//   - 主题通过 init() 函数在编译时自动注册
//   - 使用单例模式管理全局注册表
//
// 作者：xfy
package theme

import (
	"sync"
)

// 全局主题注册表实例。
var (
	// globalRegistry 全局主题注册表（单例）
	globalRegistry ThemeRegistry

	// once 确保注册表只初始化一次
	once sync.Once
)

// GetRegistry 获取全局主题注册表。
//
// 使用单例模式，首次调用时创建注册表实例，后续调用返回同一实例。
//
// 返回值：
//   - ThemeRegistry: 主题注册表实例
//
// 使用示例：
//   registry := theme.GetRegistry()
//   themes := registry.All()
func GetRegistry() ThemeRegistry {
	once.Do(func() {
		globalRegistry = NewThemeRegistry()
	})
	return globalRegistry
}

// Register 全局注册主题函数。
//
// 供主题包的 init() 调用，将主题注册到全局注册表。
//
// 参数：
//   - theme: 主题对象
//
// 返回值：
//   - err: 注册失败时返回错误，如主题已存在
//
// 使用示例：
//   func init() {
//       if err := theme.Register(New()); err != nil {
//           log.Printf("主题注册失败: %v", err)
//       }
//   }
func Register(theme Theme) error {
	return GetRegistry().Register(theme)
}

// GetActive 获取当前激活主题。
//
// 返回系统当前使用的主题。
//
// 返回值：
//   - Theme: 激活的主题对象
//   - err: 未设置激活主题时返回 ErrNoActiveTheme
//
// 使用示例：
//   active, err := theme.GetActive()
//   if err != nil {
//       // 使用默认主题
//   }
func GetActive() (Theme, error) {
	return GetRegistry().GetActive()
}

// SetActive 设置激活主题。
//
// 切换系统使用的主题。
//
// 参数：
//   - themeID: 主题 ID
//
// 返回值：
//   - err: 主题不存在返回 ErrThemeNotFound
//
// 使用示例：
//   if err := theme.SetActive("dark"); err != nil {
//       log.Printf("切换主题失败: %v", err)
//   }
func SetActive(themeID string) error {
	return GetRegistry().SetActive(themeID)
}

// All 列出所有主题。
//
// 返回系统中所有已注册的主题列表。
//
// 返回值：
//   - []Theme: 主题列表
//
// 使用示例：
//   themes := theme.All()
//   for _, t := range themes {
//       fmt.Println(t.Name)
//   }
func All() []Theme {
	return GetRegistry().All()
}

// ThemeRegistry 主题管理接口。
//
// 定义主题注册表的核心操作方法，支持注册、切换和查询主题。
type ThemeRegistry interface {
	// Register 注册主题。
	//
	// 参数：
	//   - theme: 主题对象
	//
	// 返回值：
	//   - err: 主题已存在返回 ErrThemeAlreadyExists
	Register(theme Theme) error

	// GetActive 获取当前激活主题。
	//
	// 返回值：
	//   - Theme: 激活的主题对象
	//   - err: 未设置激活主题时返回 ErrNoActiveTheme
	GetActive() (Theme, error)

	// SetActive 切换主题。
	//
	// 参数：
	//   - themeID: 主题 ID
	//
	// 返回值：
	//   - err: 主题不存在返回 ErrThemeNotFound
	SetActive(themeID string) error

	// All 列出所有主题。
	//
	// 返回值：
	//   - []Theme: 所有已注册的主题列表
	All() []Theme
}

// themeRegistry 主题注册表实现。
//
// 实现 ThemeRegistry 接口，使用读写锁保证并发安全。
type themeRegistry struct {
	// mu 保护并发访问
	mu sync.RWMutex

	// themes 存储已注册的主题，键为主题 ID
	themes map[string]Theme

	// activeThemeID 当前激活的主题 ID
	activeThemeID string
}

// NewThemeRegistry 创建主题注册表。
//
// 返回值：
//   - ThemeRegistry: 新的注册表实例
func NewThemeRegistry() ThemeRegistry {
	return &themeRegistry{
		themes: make(map[string]Theme),
	}
}

// Register 注册主题。
//
// 将主题添加到注册表。主题 ID 不能为空，且不能重复注册。
//
// 参数：
//   - theme: 主题对象
//
// 返回值：
//   - err: ID 无效返回 ErrInvalidThemeID，已存在返回 ErrThemeAlreadyExists
func (r *themeRegistry) Register(theme Theme) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if theme.ID == "" {
		return ErrInvalidThemeID
	}

	if _, exists := r.themes[theme.ID]; exists {
		return ErrThemeAlreadyExists
	}

	r.themes[theme.ID] = theme
	return nil
}

// GetActive 获取当前激活主题。
//
// 返回值：
//   - Theme: 激活的主题对象
//   - err: 未设置激活主题返回 ErrNoActiveTheme，主题不存在返回 ErrThemeNotFound
func (r *themeRegistry) GetActive() (Theme, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.activeThemeID == "" {
		return Theme{}, ErrNoActiveTheme
	}

	theme, exists := r.themes[r.activeThemeID]
	if !exists {
		return Theme{}, ErrThemeNotFound
	}

	return theme, nil
}

// SetActive 切换主题。
//
// 参数：
//   - themeID: 主题 ID
//
// 返回值：
//   - err: ID 无效返回 ErrInvalidThemeID，主题不存在返回 ErrThemeNotFound
func (r *themeRegistry) SetActive(themeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if themeID == "" {
		return ErrInvalidThemeID
	}

	if _, exists := r.themes[themeID]; !exists {
		return ErrThemeNotFound
	}

	r.activeThemeID = themeID
	return nil
}

// All 列出所有主题。
//
// 返回值：
//   - []Theme: 所有已注册主题的列表
func (r *themeRegistry) All() []Theme {
	r.mu.RLock()
	defer r.mu.RUnlock()

	themes := make([]Theme, 0, len(r.themes))
	for _, theme := range r.themes {
		themes = append(themes, theme)
	}
	return themes
}