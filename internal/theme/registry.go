// Package theme 主题引擎模块
package theme

import (
	"sync"
)

// 全局主题注册表实例
var (
	globalRegistry ThemeRegistry
	once           sync.Once
)

// GetRegistry 获取全局主题注册表（单例模式）
func GetRegistry() ThemeRegistry {
	once.Do(func() {
		globalRegistry = NewThemeRegistry()
	})
	return globalRegistry
}

// Register 全局注册主题函数（供主题包的 init() 调用）
func Register(theme Theme) error {
	return GetRegistry().Register(theme)
}

// GetActive 获取当前激活主题
func GetActive() (Theme, error) {
	return GetRegistry().GetActive()
}

// SetActive 设置激活主题
func SetActive(themeID string) error {
	return GetRegistry().SetActive(themeID)
}

// All 列出所有主题
func All() []Theme {
	return GetRegistry().All()
}

// ThemeRegistry 主题管理接口
type ThemeRegistry interface {
	// Register 注册主题
	Register(theme Theme) error
	// GetActive 获取当前激活主题
	GetActive() (Theme, error)
	// SetActive 切换主题
	SetActive(themeID string) error
	// All 列出所有主题
	All() []Theme
}

// themeRegistry 主题注册表实现
type themeRegistry struct {
	mu           sync.RWMutex
	themes       map[string]Theme
	activeThemeID string
}

// NewThemeRegistry 创建主题注册表
func NewThemeRegistry() ThemeRegistry {
	return &themeRegistry{
		themes: make(map[string]Theme),
	}
}

// Register 注册主题
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

// GetActive 获取当前激活主题
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

// SetActive 切换主题
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

// All 列出所有主题
func (r *themeRegistry) All() []Theme {
	r.mu.RLock()
	defer r.mu.RUnlock()

	themes := make([]Theme, 0, len(r.themes))
	for _, theme := range r.themes {
		themes = append(themes, theme)
	}
	return themes
}