// Package services 业务服务层
//
// Service 层负责业务逻辑处理，将 Handler 和 Repository 解耦：
// - Handler 仅负责 HTTP 请求解析/响应序列化
// - Service 包含业务逻辑（验证、密码哈希、默认角色等）
// - Repository 仅负责数据持久化
package services

import (
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/internal/core/media"
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
	CommentService  CommentService
	MediaService    MediaService
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

// NewContainerWithComments 创建带评论服务的完整容器
func NewContainerWithComments(
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	jwtService *auth.JWTService,
	blacklist TokenBlacklist,
	postRepo post.PostRepository,
	categoryRepo post.CategoryRepository,
	tagRepo post.TagRepository,
	seriesRepo post.SeriesRepository,
	commentRepo comment.CommentRepository,
	commentLikeRepo comment.CommentLikeRepository,
) *Container {
	userService := NewUserService(userRepo, roleRepo)
	authService := NewAuthServiceWithBlacklist(userRepo, jwtService, blacklist)
	postService := NewPostService(postRepo, categoryRepo, tagRepo, seriesRepo)
	categoryService := NewCategoryService(categoryRepo)
	tagService := NewTagService(tagRepo)
	seriesService := NewSeriesService(seriesRepo)
	commentService := NewCommentService(commentRepo, commentLikeRepo)

	return &Container{
		UserService:     userService,
		AuthService:     authService,
		PostService:     postService,
		CategoryService: categoryService,
		TagService:      tagService,
		SeriesService:   seriesService,
		CommentService:  commentService,
		jwtService:      jwtService,
	}
}

// NewContainerWithMedia 创建带媒体服务的完整容器
func NewContainerWithMedia(
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	jwtService *auth.JWTService,
	blacklist TokenBlacklist,
	postRepo post.PostRepository,
	categoryRepo post.CategoryRepository,
	tagRepo post.TagRepository,
	seriesRepo post.SeriesRepository,
	commentRepo comment.CommentRepository,
	commentLikeRepo comment.CommentLikeRepository,
	mediaRepo media.MediaRepository,
	uploadDir string,
	baseURL string,
	postLikeRepo post.PostLikeRepository,
) *Container {
	userService := NewUserService(userRepo, roleRepo)
	authService := NewAuthServiceWithBlacklist(userRepo, jwtService, blacklist)
	postService := NewPostServiceWithLikes(postRepo, categoryRepo, tagRepo, seriesRepo, postLikeRepo)
	categoryService := NewCategoryService(categoryRepo)
	tagService := NewTagService(tagRepo)
	seriesService := NewSeriesService(seriesRepo)
	commentService := NewCommentService(commentRepo, commentLikeRepo)
	mediaService := NewMediaService(mediaRepo, uploadDir, baseURL)

	return &Container{
		UserService:     userService,
		AuthService:     authService,
		PostService:     postService,
		CategoryService: categoryService,
		TagService:      tagService,
		SeriesService:   seriesService,
		CommentService:  commentService,
		MediaService:    mediaService,
		jwtService:      jwtService,
	}
}