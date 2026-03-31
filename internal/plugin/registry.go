// Package plugin 插件注册中心
package plugin

import (
	"fmt"
	"log"
	"sync"
)

// 全局插件注册表
var (
	pluginMap = make(map[string]PluginConstructor)
	mu        sync.RWMutex
)

// RegisterPlugin 注册插件（由插件的 init() 调用）
//
// 采用编译时注册模式，插件通过 blank import 触发 init() 自动注册。
// 注册时会验证依赖关系并初始化插件。
func RegisterPlugin(ctor PluginConstructor) {
	p := ctor()
	info := p.Info()

	mu.Lock()
	defer mu.Unlock()

	// 检查是否已注册
	if _, exists := pluginMap[info.ID]; exists {
		log.Printf("Warning: plugin %s already registered, skipping", info.ID)
		return
	}

	// 验证依赖
	for _, dep := range info.Dependencies {
		if _, ok := pluginMap[dep]; !ok {
			log.Printf("Warning: plugin %s depends on %s which is not registered", info.ID, dep)
		}
	}

	pluginMap[info.ID] = ctor
	log.Printf("Plugin registered: %s (%s)", info.Name, info.Version)
}

// RegisterPluginWithContext 注册插件并立即初始化
//
// 当 PluginContext 可用时，使用此函数在注册时直接初始化插件。
func RegisterPluginWithContext(ctor PluginConstructor, ctx *PluginContext) error {
	p := ctor()
	info := p.Info()

	mu.Lock()
	defer mu.Unlock()

	// 检查是否已注册
	if _, exists := pluginMap[info.ID]; exists {
		return fmt.Errorf("plugin %s already registered", info.ID)
	}

	// 验证依赖
	for _, dep := range info.Dependencies {
		if _, ok := pluginMap[dep]; !ok {
			log.Printf("Warning: plugin %s depends on %s which is not registered", info.ID, dep)
		}
	}

	// 初始化插件
	if err := p.Init(ctx); err != nil {
		log.Printf("Warning: plugin %s failed to initialize: %v", info.ID, err)
		return fmt.Errorf("plugin %s init failed: %w", info.ID, err)
	}

	pluginMap[info.ID] = ctor
	log.Printf("Plugin registered and initialized: %s (%s)", info.Name, info.Version)
	return nil
}

// GetPlugin 获取已注册的插件实例
func GetPlugin(id string) (Plugin, error) {
	mu.RLock()
	defer mu.RUnlock()

	ctor, ok := pluginMap[id]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}
	return ctor(), nil
}

// AllPlugins 返回所有已注册插件的元信息
func AllPlugins() []PluginInfo {
	mu.RLock()
	defer mu.RUnlock()

	infos := make([]PluginInfo, 0, len(pluginMap))
	for _, ctor := range pluginMap {
		infos = append(infos, ctor().Info())
	}
	return infos
}

// PluginIDs 返回所有已注册插件的 ID
func PluginIDs() []string {
	mu.RLock()
	defer mu.RUnlock()

	ids := make([]string, 0, len(pluginMap))
	for id := range pluginMap {
		ids = append(ids, id)
	}
	return ids
}

// HasPlugin 检查插件是否已注册
func HasPlugin(id string) bool {
	mu.RLock()
	defer mu.RUnlock()

	_, ok := pluginMap[id]
	return ok
}

// Count 返回已注册插件数量
func Count() int {
	mu.RLock()
	defer mu.RUnlock()

	return len(pluginMap)
}