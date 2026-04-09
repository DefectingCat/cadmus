// Package services 提供 Cadmus 的核心业务服务层实现。
//
// 该包包含以下主要服务：
//   - UserService: 用户信息管理服务
//   - AuthService: 用户认证和授权服务
//   - PostService: 文章管理服务
//   - CommentService: 评论管理服务
//   - MediaService: 媒体文件管理服务
//   - NotificationService: 通知推送服务
//   - RSSService: RSS 订阅源生成服务
//   - SearchService: 全文搜索服务
//
// 服务设计原则：
//  1. 所有服务通过接口定义，便于测试和替换实现
//  2. 服务间通过依赖注入解耦，使用 Container 聚合
//  3. 所有公开方法均为并发安全
//  4. 错误使用语义化错误类型，便于调用方处理
//
// Service 层负责业务逻辑处理，将 Handler 和 Repository 解耦：
//   - Handler 仅负责 HTTP 请求解析/响应序列化
//   - Service 包含业务逻辑（验证、密码哈希、默认角色等）
//   - Repository 仅负责数据持久化
//
// 作者：xfy
package services

import (
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/internal/core/media"
	"rua.plus/cadmus/internal/core/notify"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/core/search"
	"rua.plus/cadmus/internal/core/user"
)

// Container 服务容器，聚合所有业务服务。
//
// Container 采用依赖注入模式，统一管理所有服务的创建和依赖关系。
// 通过不同的构造函数（NewContainer、NewContainerWithPosts 等），
// 可以按需创建不同功能级别的服务容器，实现渐进式功能启用。
//
// 使用示例：
//
//	// 创建基础容器（仅用户和认证服务）
//	container := services.NewContainer(userRepo, roleRepo, jwtService)
//
//	// 创建完整容器（包含文章、评论、媒体等所有服务）
//	container := services.NewContainerWithNotifications(...)
//
// 注意事项：
//   - 所有字段在创建后不可修改（不可变设计）
//   - 使用前需确保依赖的 Repository 已正确初始化
type Container struct {
	// UserService 用户信息管理服务，处理用户注册、查询、更新等操作
	UserService UserService

	// AuthService 认证服务，处理登录、登出、Token 验证等操作
	AuthService AuthService

	// PostService 文章服务，处理文章创建、发布、版本管理等操作
	PostService PostService

	// CategoryService 分类服务，处理分类的 CRUD 操作
	CategoryService CategoryService

	// TagService 标签服务，处理标签的 CRUD 操作
	TagService TagService

	// SeriesService 系列服务，处理文章系列的 CRUD 操作
	SeriesService SeriesService

	// CommentService 评论服务，处理评论创建、审核、点赞等操作
	CommentService CommentService

	// MediaService 媒体服务，处理文件上传、管理、删除等操作
	MediaService MediaService

	// NotificationService 通知服务，处理评论通知、回复通知等推送
	NotificationService NotificationService

	// RSSService RSS 服务，生成 RSS/Atom 订阅源
	RSSService RSSService

	// SearchService 搜索服务，提供全文搜索功能
	SearchService SearchService

	// jwtService JWT 服务（内部字段，供 Handler 直接使用）
	jwtService *auth.JWTService
}

// NewContainer 创建基础服务容器。
//
// 该函数创建包含用户服务和认证服务的最小容器，适用于仅需用户管理功能的场景。
// 其他服务字段为零值，需要后续通过其他构造函数补充。
//
// 参数：
//   - userRepo: 用户数据仓库，用于用户数据的持久化操作
//   - roleRepo: 角色数据仓库，用于获取默认角色等信息
//   - jwtService: JWT 服务，用于 Token 生成和验证
//
// 返回值：
//   - *Container: 初始化完成的服务容器
//
// 使用示例：
//
//	container := services.NewContainer(userRepo, roleRepo, jwtService)
//	user, err := container.UserService.Register(ctx, "test", "test@example.com", "password")
func NewContainer(
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	jwtService *auth.JWTService,
) *Container {
	userService := NewUserService(userRepo, roleRepo)
	authService := NewAuthService(userRepo, jwtService)

	return &Container{
		UserService: userService,
		AuthService: authService,
		jwtService:  jwtService,
	}
}

// NewContainerWithBlacklist 创建带 Token 黑名单的服务容器。
//
// 该函数在 NewContainer 基础上增加 Token 黑名单功能，用于支持用户登出后
// 立即失效 Token 的安全需求。适用于需要更严格安全控制的场景。
//
// 参数：
//   - userRepo: 用户数据仓库
//   - roleRepo: 角色数据仓库
//   - jwtService: JWT 服务
//   - blacklist: Token 黑名单实现（通常基于 Redis 或内存）
//
// 返回值：
//   - *Container: 包含黑名单功能的认证服务容器
//
// 注意事项：
//   - 黑名单会增加 Token 验证的额外开销
//   - 需确保黑名单存储的可靠性，否则可能导致 Token 无法失效
func NewContainerWithBlacklist(
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	jwtService *auth.JWTService,
	blacklist TokenBlacklist,
) *Container {
	userService := NewUserService(userRepo, roleRepo)
	authService := NewAuthServiceWithBlacklist(userRepo, jwtService, blacklist)

	return &Container{
		UserService: userService,
		AuthService: authService,
		jwtService:  jwtService,
	}
}

// NewContainerWithPosts 创建带文章服务的完整容器。
//
// 该函数在基础容器上增加文章管理相关服务，包括分类、标签、系列等。
// 适用于博客系统核心功能的场景。
//
// 参数：
//   - userRepo, roleRepo: 用户和角色数据仓库
//   - jwtService: JWT 服务
//   - blacklist: Token 黑名单（可选，传 nil 则不启用）
//   - postRepo: 文章数据仓库
//   - categoryRepo: 分类数据仓库
//   - tagRepo: 标签数据仓库
//   - seriesRepo: 系列数据仓库
//
// 返回值：
//   - *Container: 包含用户、认证、文章及其相关服务的容器
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

// JWTService 获取 JWT 服务实例。
//
// 该方法提供对内部 JWT 服务的访问，供 Handler 直接使用。
// 某些场景下 Handler 需要直接操作 JWT（如解析 Token 获取 Claims），
// 而不经过 AuthService 的完整验证流程。
//
// 返回值：
//   - *auth.JWTService: JWT 服务实例
//
// 注意事项：
//   - 该方法返回的服务实例不可修改
func (c *Container) JWTService() *auth.JWTService {
	return c.jwtService
}

// NewContainerWithComments 创建带评论服务的完整容器。
//
// 该函数在文章服务基础上增加评论功能，包括评论审核、点赞等。
// 适用于需要完整博客交互功能的场景。
//
// 参数：
//   - userRepo, roleRepo: 用户和角色数据仓库
//   - jwtService: JWT 服务
//   - blacklist: Token 黑名单
//   - postRepo, categoryRepo, tagRepo, seriesRepo: 文章相关仓库
//   - commentRepo: 评论数据仓库
//   - commentLikeRepo: 评论点赞数据仓库
//
// 返回值：
//   - *Container: 包含用户、认证、文章和评论服务的容器
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

// NewContainerWithMedia 创建带媒体服务的完整容器。
//
// 该函数在评论服务基础上增加媒体文件管理和 RSS、搜索功能。
// 适用于需要文件上传和内容搜索的完整博客系统。
//
// 参数：
//   - 各用户、文章、评论相关仓库（见前述函数说明）
//   - mediaRepo: 媒体数据仓库
//   - uploadDir: 文件上传目录路径（物理存储位置）
//   - baseURL: 基础 URL（用于生成媒体文件的访问链接）
//   - postLikeRepo: 文章点赞数据仓库
//   - searchRepo: 搜索数据仓库
//
// 返回值：
//   - *Container: 包含所有核心业务服务的容器
//
// 注意事项：
//   - uploadDir 目录需确保有写入权限
//   - baseURL 应为完整的外部访问地址（如 https://example.com）
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
	searchRepo search.SearchRepository,
) *Container {
	userService := NewUserService(userRepo, roleRepo)
	authService := NewAuthServiceWithBlacklist(userRepo, jwtService, blacklist)
	postService := NewPostServiceWithLikes(postRepo, categoryRepo, tagRepo, seriesRepo, postLikeRepo)
	categoryService := NewCategoryService(categoryRepo)
	tagService := NewTagService(tagRepo)
	seriesService := NewSeriesService(seriesRepo)
	commentService := NewCommentService(commentRepo, commentLikeRepo)
	mediaService := NewMediaService(mediaRepo, uploadDir, baseURL)
	rssService := NewRSSService(postRepo, categoryRepo)
	searchService := NewSearchService(searchRepo)

	return &Container{
		UserService:     userService,
		AuthService:     authService,
		PostService:     postService,
		CategoryService: categoryService,
		TagService:      tagService,
		SeriesService:   seriesService,
		CommentService:  commentService,
		MediaService:    mediaService,
		RSSService:      rssService,
		SearchService:   searchService,
		jwtService:      jwtService,
	}
}

// NotificationChannel 通知渠道别名，用于简化容器创建。
//
// 该别名将 notify 包的 NotificationChannel 接口映射到 services 包，
// 便于在构造函数中直接使用 NotificationChannel 类型，避免导入 notify 包。
type NotificationChannel = notify.NotificationChannel

// NewContainerWithNotifications 创建带通知服务的完整容器。
//
// 该函数是最高级别的容器构造函数，包含所有业务服务。
// 适用于需要完整博客系统功能的场景，包括通知推送。
//
// 参数：
//   - 各仓库和服务配置（见前述函数说明）
//   - notificationChannel: 通知渠道实现（如 EmailChannel）
//
// 返回值：
//   - *Container: 包含所有业务服务的完整容器
//
// 使用示例：
//
//	emailChannel := services.NewEmailChannel(smtpConfig)
//	container := services.NewContainerWithNotifications(
//	    userRepo, roleRepo, jwtService, blacklist,
//	    postRepo, categoryRepo, tagRepo, seriesRepo,
//	    commentRepo, commentLikeRepo,
//	    mediaRepo, "/uploads", "https://example.com",
//	    postLikeRepo, searchRepo,
//	    emailChannel,
//	)
func NewContainerWithNotifications(
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
	searchRepo search.SearchRepository,
	notificationChannel NotificationChannel,
) *Container {
	userService := NewUserService(userRepo, roleRepo)
	authService := NewAuthServiceWithBlacklist(userRepo, jwtService, blacklist)
	postService := NewPostServiceWithLikes(postRepo, categoryRepo, tagRepo, seriesRepo, postLikeRepo)
	categoryService := NewCategoryService(categoryRepo)
	tagService := NewTagService(tagRepo)
	seriesService := NewSeriesService(seriesRepo)
	commentService := NewCommentService(commentRepo, commentLikeRepo)
	mediaService := NewMediaService(mediaRepo, uploadDir, baseURL)
	rssService := NewRSSService(postRepo, categoryRepo)
	searchService := NewSearchService(searchRepo)
	notificationService := NewNotificationService(notificationChannel)

	return &Container{
		UserService:         userService,
		AuthService:         authService,
		PostService:         postService,
		CategoryService:     categoryService,
		TagService:          tagService,
		SeriesService:       seriesService,
		CommentService:      commentService,
		MediaService:        mediaService,
		NotificationService: notificationService,
		RSSService:          rssService,
		SearchService:       searchService,
		jwtService:          jwtService,
	}
}
