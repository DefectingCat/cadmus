package cache

import "fmt"

// 缓存 key 常量定义
const (
	CacheNamespace = "cadmus"

	// 文章缓存 key 格式
	PostDetailKey = "cadmus:post:detail:%s:v%d" // {slug}:{version}
	PostListKey   = "cadmus:post:list:%s:%d:%s" // {category}:{page}:{sort}

	// 用户缓存 key 格式
	UserInfoKey  = "cadmus:user:info:%s"    // {user_id}
	UserPermsKey = "cadmus:user:perms:%s:%s" // {user_id}:{permission}

	// 角色缓存 key 格式
	RoleInfoKey  = "cadmus:role:info:%s"  // {role_id}
	RolePermsKey = "cadmus:role:perms:%s" // {role_id}

	// 配置缓存 key 格式
	SiteConfigKey  = "cadmus:site:config"
	ThemeConfigKey = "cadmus:theme:config:%s" // {theme_id}
)

// BuildPostDetailKey 构建文章详情缓存 key
func BuildPostDetailKey(slug string, version int) string {
	return fmt.Sprintf(PostDetailKey, slug, version)
}

// BuildPostListKey 构建文章列表缓存 key
func BuildPostListKey(category string, page int, sort string) string {
	return fmt.Sprintf(PostListKey, category, page, sort)
}

// BuildUserInfoKey 构建用户信息缓存 key
func BuildUserInfoKey(userID string) string {
	return fmt.Sprintf(UserInfoKey, userID)
}

// BuildUserPermsKey 构建用户权限缓存 key
func BuildUserPermsKey(userID, permission string) string {
	return fmt.Sprintf(UserPermsKey, userID, permission)
}

// BuildRoleInfoKey 构建角色信息缓存 key
func BuildRoleInfoKey(roleID string) string {
	return fmt.Sprintf(RoleInfoKey, roleID)
}

// BuildRolePermsKey 构建角色权限缓存 key
func BuildRolePermsKey(roleID string) string {
	return fmt.Sprintf(RolePermsKey, roleID)
}

// BuildThemeConfigKey 构建主题配置缓存 key
func BuildThemeConfigKey(themeID string) string {
	return fmt.Sprintf(ThemeConfigKey, themeID)
}