// Package cache 提供 Redis 缓存键管理功能。
//
// 该文件包含缓存键的常量定义和构建函数，包括：
//   - 缓存命名空间定义
//   - 各类数据的缓存键格式
//   - 键构建辅助函数
//
// 主要用途：
//
//	统一管理缓存键命名规范，避免键名冲突和拼写错误。
//
// 注意事项：
//   - 所有缓存键使用 "cadmus" 作为命名空间前缀
//   - 键名格式采用 "namespace:entity:action:identifier" 模式
//   - 使用格式化函数构建键名，避免手动拼接错误
//
// 作者：xfy
package cache

import "fmt"

// 缓存键常量定义。
//
// 定义系统中使用的所有缓存键格式。
// 使用格式化占位符，运行时通过 Build 函数填充实际值。
const (
	// CacheNamespace 缓存命名空间，作为所有键的前缀
	CacheNamespace = "cadmus"

	// PostDetailKey 文章详情缓存键格式
	// 参数: {slug} 文章 Slug, {version} 文章版本号
	// 示例: cadmus:post:detail:my-article:v1
	PostDetailKey = "cadmus:post:detail:%s:v%d"

	// PostListKey 文章列表缓存键格式
	// 参数: {category} 分类标识, {page} 页码, {sort} 排序方式
	// 示例: cadmus:post:list:tech:1:created_at
	PostListKey = "cadmus:post:list:%s:%d:%s"

	// UserInfoKey 用户信息缓存键格式
	// 参数: {user_id} 用户 ID
	// 示例: cadmus:user:info:550e8400-e29b-41d4-a716-446655440000
	UserInfoKey = "cadmus:user:info:%s"

	// UserPermsKey 用户权限缓存键格式
	// 参数: {user_id} 用户 ID, {permission} 权限名称
	// 示例: cadmus:user:perms:550e8400:post.create
	UserPermsKey = "cadmus:user:perms:%s:%s"

	// RoleInfoKey 角色信息缓存键格式
	// 参数: {role_id} 角色 ID
	// 示例: cadmus:role:info:550e8400-e29b-41d4-a716-446655440000
	RoleInfoKey = "cadmus:role:info:%s"

	// RolePermsKey 角色权限缓存键格式
	// 参数: {role_id} 角色 ID
	// 示例: cadmus:role:perms:550e8400-e29b-41d4-a716-446655440000
	RolePermsKey = "cadmus:role:perms:%s"

	// SiteConfigKey 站点配置缓存键
	// 全局唯一，无需参数
	SiteConfigKey = "cadmus:site:config"

	// ThemeConfigKey 主题配置缓存键格式
	// 参数: {theme_id} 主题 ID
	// 示例: cadmus:theme:config:default
	ThemeConfigKey = "cadmus:theme:config:%s"
)

// BuildPostDetailKey 构建文章详情缓存键。
//
// 根据文章 Slug 和版本号生成唯一的缓存键。
// 用于缓存文章详情页数据。
//
// 参数：
//   - slug: 文章 URL 别名
//   - version: 文章版本号
//
// 返回值：
//   - 格式化后的缓存键字符串
//
// 使用示例：
//
//	key := BuildPostDetailKey("my-article", 1)
//	// 返回: "cadmus:post:detail:my-article:v1"
func BuildPostDetailKey(slug string, version int) string {
	return fmt.Sprintf(PostDetailKey, slug, version)
}

// BuildPostListKey 构建文章列表缓存键。
//
// 根据分类、页码和排序方式生成列表缓存键。
// 用于缓存文章列表页数据。
//
// 参数：
//   - category: 分类标识（空字符串表示全部分类）
//   - page: 页码（从 1 开始）
//   - sort: 排序方式（如 "created_at"、"view_count"）
//
// 返回值：
//   - 格式化后的缓存键字符串
//
// 使用示例：
//
//	key := BuildPostListKey("tech", 1, "created_at")
//	// 返回: "cadmus:post:list:tech:1:created_at"
func BuildPostListKey(category string, page int, sort string) string {
	return fmt.Sprintf(PostListKey, category, page, sort)
}

// BuildUserInfoKey 构建用户信息缓存键。
//
// 根据用户 ID 生成用户信息缓存键。
// 用于缓存用户基本信息。
//
// 参数：
//   - userID: 用户 ID 字符串
//
// 返回值：
//   - 格式化后的缓存键字符串
//
// 使用示例：
//
//	key := BuildUserInfoKey("550e8400-e29b-41d4-a716-446655440000")
func BuildUserInfoKey(userID string) string {
	return fmt.Sprintf(UserInfoKey, userID)
}

// BuildUserPermsKey 构建用户权限缓存键。
//
// 根据用户 ID 和权限名称生成权限缓存键。
// 用于缓存用户的特定权限判断结果。
//
// 参数：
//   - userID: 用户 ID 字符串
//   - permission: 权限名称（如 "post.create"）
//
// 返回值：
//   - 格式化后的缓存键字符串
//
// 使用示例：
//
//	key := BuildUserPermsKey("550e8400", "post.create")
func BuildUserPermsKey(userID, permission string) string {
	return fmt.Sprintf(UserPermsKey, userID, permission)
}

// BuildRoleInfoKey 构建角色信息缓存键。
//
// 根据角色 ID 生成角色信息缓存键。
// 用于缓存角色基本信息。
//
// 参数：
//   - roleID: 角色 ID 字符串
//
// 返回值：
//   - 格式化后的缓存键字符串
func BuildRoleInfoKey(roleID string) string {
	return fmt.Sprintf(RoleInfoKey, roleID)
}

// BuildRolePermsKey 构建角色权限缓存键。
//
// 根据角色 ID 生成角色权限列表缓存键。
// 用于缓存角色拥有的所有权限。
//
// 参数：
//   - roleID: 角色 ID 字符串
//
// 返回值：
//   - 格式化后的缓存键字符串
func BuildRolePermsKey(roleID string) string {
	return fmt.Sprintf(RolePermsKey, roleID)
}

// BuildThemeConfigKey 构建主题配置缓存键。
//
// 根据主题 ID 生成主题配置缓存键。
// 用于缓存主题的配置参数。
//
// 参数：
//   - themeID: 主题 ID（如 "default"）
//
// 返回值：
//   - 格式化后的缓存键字符串
func BuildThemeConfigKey(themeID string) string {
	return fmt.Sprintf(ThemeConfigKey, themeID)
}
