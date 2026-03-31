// Package services 业务服务层
//
// Service 层负责业务逻辑处理，将 Handler 和 Repository 解耦：
// - Handler 仅负责 HTTP 请求解析/响应序列化
// - Service 包含业务逻辑（验证、密码哈希、默认角色等）
// - Repository 仅负责数据持久化
package services

import (
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/core/user"
)

// Container 服务容器，聚合所有业务服务
type Container struct {
	UserService     UserService
	AuthService     AuthService
	PostService     PostService
	CategoryService CategoryService
	TagService      TagService
	SeriesService   SeriesService
	jwtService      *auth.JWTService
}

// NewContainer 创建服务容器
func NewContainer(
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	jwtService *auth.JWTService,
) *Container {
	userService := NewUserService(userRepo, roleRepo)
	authService := NewAuthService(userRepo, jwtService)

	return &Container{
		UserService:  userService,
		AuthService:  authService,
		jwtService:   jwtService,
	}
}

// NewContainerWithBlacklist 创建带黑名单的服务容器
func NewContainerWithBlacklist(
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	jwtService *auth.JWTService,
	blacklist TokenBlacklist,
) *Container {
	userService := NewUserService(userRepo, roleRepo)
	authService := NewAuthServiceWithBlacklist(userRepo, jwtService, blacklist)

	return &Container{
		UserService:  userService,
		AuthService:  authService,
		jwtService:   jwtService,
	}
}

// NewContainerWithPosts 创建带文章服务的完整容器
func NewContainerWithPosts(
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	jwtService *auth.JWTService,
	blacklist TokenBlacklist,
	postRepo post.PostRepository,
	categoryRepo post.CategoryRepository,
	tagRepo post.TagRepository,
	seriesRepo post.SeriesRepository,
) *Container {
	userService := NewUserService(userRepo, roleRepo)
	authService := NewAuthServiceWithBlacklist(userRepo, jwtService, blacklist)
	postService := NewPostService(postRepo, categoryRepo, tagRepo, seriesRepo)
	categoryService := NewCategoryService(categoryRepo)
	tagService := NewTagService(tagRepo)
	seriesService := NewSeriesService(seriesRepo)

	return &Container{
		UserService:     userService,
		AuthService:     authService,
		PostService:     postService,
		CategoryService: categoryService,
		TagService:      tagService,
		SeriesService:   seriesService,
		jwtService:      jwtService,
	}
}

// JWTService 获取 JWT 服务（供 Handler 直接使用）
func (c *Container) JWTService() *auth.JWTService {
	return c.jwtService
}