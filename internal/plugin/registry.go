// Package plugin 提供插件注册中心实现。
//
// 该文件包含插件系统的全局注册表和相关操作函数，包括：
//   - 全局插件映射表
//   - 插件注册函数
//   - 插件查询函数
//   - 插件列表函数
//
// 主要用途：
//
//	管理已注册插件的生命周期，支持插件的注册、查询和遍历。
//
// 注意事项：
//   - 所有注册表操作都是并发安全的
//   - 插件通过 init() 函数在编译时自动注册
//   - 重复注册同一插件 ID 会被忽略并记录警告
//
// 作者：xfy
package plugin

import (
	"fmt"
	"sync"

	"rua.plus/cadmus/internal/logger"
)

// 全局插件注册表。
var (
	// pluginMap 存储已注册插件的构造函数
	// 键为插件 ID，值为构造函数
	pluginMap = make(map[string]PluginConstructor)

	// mu 保护 pluginMap 的并发访问
	mu sync.RWMutex
)

// RegisterPlugin 注册插件。
//
// 由插件的 init() 函数调用，将插件注册到全局注册表。
// 采用编译时注册模式，插件通过 blank import 触发 init() 自动注册。
//
// 注册流程：
//  1. 调用构造函数创建插件实例
//  2. 获取插件元信息
//  3. 检查是否已注册（重复注册会被忽略）
//  4. 验证依赖关系（缺失依赖会记录警告）
//  5. 存储构造函数到注册表
//
// 参数：
//   - ctor: 插件构造函数
//
// 注意事项：
//   - 此方法不会初始化插件，仅完成注册
//   - 如需在注册时初始化，请使用 RegisterPluginWithContext
func RegisterPlugin(ctor PluginConstructor) {
	p := ctor()
	info := p.Info()

	mu.Lock()
	defer mu.Unlock()

	// 检查是否已注册
	if _, exists := pluginMap[info.ID]; exists {
		logger.Warn(fmt.Sprintf("plugin %s already registered, skipping", info.ID))
		return
	}

	// 验证依赖
	for _, dep := range info.Dependencies {
		if _, ok := pluginMap[dep]; !ok {
			logger.Warn(fmt.Sprintf("plugin %s depends on %s which is not registered", info.ID, dep))
		}
	}

	pluginMap[info.ID] = ctor
	logger.Info(fmt.Sprintf("Plugin registered: %s (%s)", info.Name, info.Version))
}

// RegisterPluginWithContext 注册插件并立即初始化。
//
// 当 PluginContext 可用时，使用此函数在注册时直接初始化插件。
// 适用于应用启动阶段，此时系统资源已准备就绪。
//
// 参数：
//   - ctor: 插件构造函数
//   - ctx: 插件运行时上下文
//
// 返回值：
//   - err: 注册或初始化失败时返回错误
//
// 错误情况：
//   - 插件已注册：返回 "plugin {id} already registered"
//   - 初始化失败：返回 "plugin {id} init failed: {error}"
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
			logger.Warn(fmt.Sprintf("plugin %s depends on %s which is not registered", info.ID, dep))
		}
	}

	// 初始化插件
	if err := p.Init(ctx); err != nil {
		logger.Warn(fmt.Sprintf("plugin %s failed to initialize: %v", info.ID, err))
		return fmt.Errorf("plugin %s init failed: %w", info.ID, err)
	}

	pluginMap[info.ID] = ctor
	logger.Info(fmt.Sprintf("Plugin registered and initialized: %s (%s)", info.Name, info.Version))
	return nil
}

// GetPlugin 获取已注册的插件实例。
//
// 根据插件 ID 获取插件的新实例。每次调用都会创建新实例。
//
// 参数：
//   - id: 插件 ID
//
// 返回值：
//   - plugin: 插件实例
//   - err: 插件不存在时返回错误
//
// 使用示例：
//   p, err := plugin.GetPlugin("github-auth")
//   if err != nil {
//       log.Fatal(err)
//   }
//   info := p.Info()
func GetPlugin(id string) (Plugin, error) {
	mu.RLock()
	defer mu.RUnlock()

	ctor, ok := pluginMap[id]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}
	return ctor(), nil
}

// AllPlugins 返回所有已注册插件的元信息。
//
// 遍历注册表，返回所有插件的 PluginInfo 列表。
// 用于管理界面展示插件列表。
//
// 返回值：
//   - []PluginInfo: 插件元信息列表
func AllPlugins() []PluginInfo {
	mu.RLock()
	defer mu.RUnlock()

	infos := make([]PluginInfo, 0, len(pluginMap))
	for _, ctor := range pluginMap {
		infos = append(infos, ctor().Info())
	}
	return infos
}

// PluginIDs 返回所有已注册插件的 ID。
//
// 用于快速获取插件 ID 列表，便于遍历或检查。
//
// 返回值：
//   - []string: 插件 ID 列表
func PluginIDs() []string {
	mu.RLock()
	defer mu.RUnlock()

	ids := make([]string, 0, len(pluginMap))
	for id := range pluginMap {
		ids = append(ids, id)
	}
	return ids
}

// HasPlugin 检查插件是否已注册。
//
// 快速判断指定 ID 的插件是否存在于注册表中。
//
// 参数：
//   - id: 插件 ID
//
// 返回值：
//   - true: 插件已注册
//   - false: 插件未注册
func HasPlugin(id string) bool {
	mu.RLock()
	defer mu.RUnlock()

	_, ok := pluginMap[id]
	return ok
}

// Count 返回已注册插件数量。
//
// 返回值：
//   - int: 已注册插件总数
func Count() int {
	mu.RLock()
	defer mu.RUnlock()

	return len(pluginMap)
}