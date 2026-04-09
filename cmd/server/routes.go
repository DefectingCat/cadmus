// Package main 提供 Cadmus 博客平台的 HTTP 路由配置功能。
//
// 该文件包含路由注册相关的核心逻辑，包括：
//   - 路由依赖聚合结构体的定义
//   - 各模块路由的分组和注册（认证、文章、评论、媒体等）
//   - 中间件的组合和应用（限流、认证、权限检查）
//   - 管理后台页面和静态文件的路由配置
//
// 主要用途：
//
//	为 Cadmus 服务器提供完整的 HTTP 路由映射，将请求正确分发到
//	对应的处理器，并应用必要的安全和限流中间件。
//
// 注意事项：
//   - 公开路由使用公开限流器，认证路由使用用户限流器
//   - 管理后台路由需要同时通过认证和权限中间件
//   - 部分路由使用 templ 组件直接渲染页面
//
// 作者：xfy
package main

import (
	"net/http"

	"github.com/a-h/templ"
	"rua.plus/cadmus/internal/api/handlers"
	"rua.plus/cadmus/internal/api/middleware"
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/database"
	"rua.plus/cadmus/internal/logger"
	"rua.plus/cadmus/internal/services"
	adminpages "rua.plus/cadmus/web/templates/pages/admin"
)

// RouteDeps 路由依赖聚合结构体。
//
// 该结构体收集了所有路由处理器和中间件所需的依赖项，
// 包括业务处理器、限流器、认证组件、数据访问对象等。
// 通过依赖注入方式传递，避免全局变量耦合。
//
// 使用示例：
//
//	deps := &RouteDeps{
//	    AuthHandler: authHandler,
//	    PostHandler: postHandler,
//	    ...
//	}
//	setupRoutes(mux, deps)
//
// 注意事项：
//   - 所有字段必须在使用前正确初始化
//   - 中间件字段为可选，部分路由不使用限流器
type RouteDeps struct {
	// === 业务处理器 ===

	// AuthHandler 认证处理器，处理登录、注册、注销等请求
	AuthHandler *handlers.AuthHandler

	// PostHandler 文章处理器，处理文章 CRUD、发布、版本管理
	PostHandler *handlers.PostHandler

	// CategoryHandler 分类处理器，处理分类的增删改查
	CategoryHandler *handlers.CategoryHandler

	// TagHandler 标签处理器，处理标签的创建和删除
	TagHandler *handlers.TagHandler

	// CommentHandler 评论处理器，处理评论的 CRUD 和审核
	CommentHandler *handlers.CommentHandler

	// MediaHandler 媒体处理器，处理文件上传和管理
	MediaHandler *handlers.MediaHandler

	// RSSHandler RSS 订阅处理器，生成 RSS/Atom Feed
	RSSHandler *handlers.RSSHandler

	// SearchHandler 搜索处理器，处理全文搜索和自动补全
	SearchHandler *handlers.SearchHandler

	// AdminHandler 管理处理器，处理角色、用户管理等后台操作
	AdminHandler *handlers.AdminHandler

	// === 限流中间件 ===

	// LoginLimiter 登录限流器，防止暴力破解攻击
	LoginLimiter *middleware.RateLimiter

	// PublicLimiter 公开 API 限流器，控制匿名请求频率
	PublicLimiter *middleware.RateLimiter

	// UserLimiter 用户操作限流器，控制已认证用户的请求频率
	UserLimiter *middleware.RateLimiter

	// === 认证组件 ===

	// JWTService JWT 服务，用于令牌签发和验证
	JWTService *auth.JWTService

	// TokenBlacklist 令牌黑名单，用于注销令牌的失效管理
	TokenBlacklist auth.TokenBlacklist

	// PermCache 权限缓存，加速权限检查查询
	PermCache *auth.PermissionCache

	// === 数据访问对象（用于管理后台页面）===

	// PostRepo 文章仓库，用于仪表盘统计查询
	PostRepo *database.PostRepository

	// CommentRepo 评论仓库，用于评论管理页面查询
	CommentRepo *database.CommentRepository

	// === 服务容器 ===

	// Services 服务容器，聚合所有业务服务
	Services *services.Container

	// === 配置参数 ===

	// UploadDir 上传文件存储目录路径
	UploadDir string
}

// setupRoutes 配置所有 HTTP 路由。
//
// 该函数是路由注册的入口点，依次调用各模块的路由设置函数，
// 完成全部 API 和页面路由的注册。
//
// 参数：
//   - mux: HTTP 路由映射器，用于注册所有路由
//   - deps: 路由依赖聚合对象，包含所有处理器和中间件
//
// 注册的路由模块：
//   - 认证路由 (auth): 登录、注册、注销、令牌刷新
//   - 文章路由 (posts): 文章 CRUD、发布、版本、点赞
//   - 分类路由 (categories): 分类管理
//   - 标签路由 (tags): 标签管理
//   - 评论路由 (comments): 评论 CRUD、审核、批量操作
//   - 媒体路由 (media): 文件上传、管理
//   - RSS 路由 (rss): RSS/Atom Feed
//   - 搜索路由 (search): 全文搜索、自动补全
//   - 管理路由 (admin): 角色、用户管理、批量操作
//   - 管理页面 (admin pages): 仪表盘、评论管理页面
//   - 静态文件 (static): CSS、JS、图片等
//   - 健康检查 (health): 服务状态探针
func setupRoutes(mux *http.ServeMux, deps *RouteDeps) {
	setupAuthRoutes(mux, deps)
	setupPostRoutes(mux, deps)
	setupCategoryRoutes(mux, deps)
	setupTagRoutes(mux, deps)
	setupCommentRoutes(mux, deps)
	setupMediaRoutes(mux, deps)
	setupRSSRoutes(mux, deps)
	setupSearchRoutes(mux, deps)
	setupAdminRoutes(mux, deps)
	setupAdminPages(mux, deps)
	setupStaticRoutes(mux, deps)
	setupHealthRoute(mux, deps)
}

// setupAuthRoutes 配置认证相关路由。
//
// 注册用户认证相关的 API 路由，包括注册、登录、注销和令牌管理。
// 登录相关路由应用限流中间件防止暴力破解。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 路由列表：
//   - POST /api/v1/auth/register: 用户注册（限流）
//   - POST /api/v1/auth/login: 用户登录（限流）
//   - POST /api/v1/auth/logout: 用户注销（需认证）
//   - POST /api/v1/auth/refresh: 刷新令牌
//   - GET /api/v1/auth/me: 获取当前用户信息（需认证）
func setupAuthRoutes(mux *http.ServeMux, deps *RouteDeps) {
	// 创建登录限流中间件，基于 IP 地址限流
	loginRL := middleware.RateLimitMiddleware(deps.LoginLimiter, middleware.IPKeyFunc("login"))

	// 创建认证中间件，包含令牌黑名单检查
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	// 注册路由：应用限流保护
	mux.HandleFunc("POST /api/v1/auth/register", loginRL(http.HandlerFunc(deps.AuthHandler.Register)).ServeHTTP)

	// 登录路由：应用限流保护，防止暴力破解
	mux.HandleFunc("POST /api/v1/auth/login", loginRL(http.HandlerFunc(deps.AuthHandler.Login)).ServeHTTP)

	// 注销路由：需认证，将令牌加入黑名单
	mux.HandleFunc("POST /api/v1/auth/logout", authMW(http.HandlerFunc(deps.AuthHandler.Logout)).ServeHTTP)

	// 令牌刷新路由：无需限流，令牌本身提供保护
	mux.HandleFunc("POST /api/v1/auth/refresh", deps.AuthHandler.Refresh)

	// 当前用户路由：需认证，返回用户信息
	mux.HandleFunc("GET /api/v1/auth/me", authMW(http.HandlerFunc(deps.AuthHandler.Me)).ServeHTTP)
}

// setupPostRoutes 配置文章相关路由。
//
// 注册文章 CRUD、发布管理、版本控制和点赞等 API 路由。
// 公开路由使用公开限流器，认证路由使用用户限流器。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 公开路由列表：
//   - GET /api/v1/posts: 文章列表（限流）
//   - GET /api/v1/posts/{slug}: 获取单篇文章（限流）
//   - GET /api/v1/posts/{id}/versions: 获取文章版本历史（限流）
//   - GET /api/v1/users/{id}/posts: 获取用户文章列表（限流）
//
// 认证路由列表：
//   - POST /api/v1/posts: 创建文章
//   - PUT /api/v1/posts/{id}: 更新文章
//   - DELETE /api/v1/posts/{id}: 删除文章
//   - POST /api/v1/posts/{id}/publish: 发布文章
//   - POST /api/v1/posts/{id}/rollback: 回滚文章版本
//   - POST /api/v1/posts/{id}/like: 点赞文章
//   - DELETE /api/v1/posts/{id}/like: 取消点赞
func setupPostRoutes(mux *http.ServeMux, deps *RouteDeps) {
	// 创建公开限流中间件，控制匿名请求频率
	publicRL := middleware.RateLimitMiddleware(deps.PublicLimiter, middleware.IPKeyFunc("public"))

	// 创建用户限流中间件，控制已认证用户请求频率
	userRL := middleware.RateLimitMiddleware(deps.UserLimiter, middleware.IPKeyFunc("user"))

	// 创建认证中间件
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	// === 公开路由 ===
	// 文章列表：支持分页和过滤
	mux.HandleFunc("GET /api/v1/posts", publicRL(http.HandlerFunc(deps.PostHandler.List)).ServeHTTP)

	// 获取文章详情：通过 slug 查询
	mux.HandleFunc("GET /api/v1/posts/{slug}", publicRL(http.HandlerFunc(deps.PostHandler.Get)).ServeHTTP)

	// 文章版本历史：查看编辑记录
	mux.HandleFunc("GET /api/v1/posts/{id}/versions", publicRL(http.HandlerFunc(deps.PostHandler.Versions)).ServeHTTP)

	// 用户文章列表：查看特定用户的文章
	mux.HandleFunc("GET /api/v1/users/{id}/posts", publicRL(http.HandlerFunc(deps.PostHandler.GetUserPosts)).ServeHTTP)

	// === 需认证路由 ===
	// 创建新文章
	mux.HandleFunc("POST /api/v1/posts", userRL(authMW(http.HandlerFunc(deps.PostHandler.Create))).ServeHTTP)

	// 更新文章内容
	mux.HandleFunc("PUT /api/v1/posts/{id}", userRL(authMW(http.HandlerFunc(deps.PostHandler.Update))).ServeHTTP)

	// 删除文章
	mux.HandleFunc("DELETE /api/v1/posts/{id}", userRL(authMW(http.HandlerFunc(deps.PostHandler.Delete))).ServeHTTP)

	// 发布文章：将草稿转为已发布状态
	mux.HandleFunc("POST /api/v1/posts/{id}/publish", userRL(authMW(http.HandlerFunc(deps.PostHandler.Publish))).ServeHTTP)

	// 回滚文章版本：恢复到历史版本
	mux.HandleFunc("POST /api/v1/posts/{id}/rollback", userRL(authMW(http.HandlerFunc(deps.PostHandler.Rollback))).ServeHTTP)

	// 点赞文章
	mux.HandleFunc("POST /api/v1/posts/{id}/like", userRL(authMW(http.HandlerFunc(deps.PostHandler.Like))).ServeHTTP)

	// 取消点赞
	mux.HandleFunc("DELETE /api/v1/posts/{id}/like", userRL(authMW(http.HandlerFunc(deps.PostHandler.Unlike))).ServeHTTP)
}

// setupCategoryRoutes 配置分类相关路由。
//
// 注册分类的增删改查 API 路由。
// 公开路由无中间件保护，修改操作需要认证。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 路由列表：
//   - GET /api/v1/categories: 分类列表（公开）
//   - GET /api/v1/categories/{slug}: 获取单个分类（公开）
//   - POST /api/v1/categories: 创建分类（需认证）
//   - PUT /api/v1/categories/{id}: 更新分类（需认证）
//   - DELETE /api/v1/categories/{id}: 删除分类（需认证）
func setupCategoryRoutes(mux *http.ServeMux, deps *RouteDeps) {
	// 创建认证中间件
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	// 分类列表：公开访问
	mux.HandleFunc("GET /api/v1/categories", deps.CategoryHandler.List)

	// 获取分类详情：通过 slug 查询
	mux.HandleFunc("GET /api/v1/categories/{slug}", deps.CategoryHandler.Get)

	// 创建分类：需认证
	mux.HandleFunc("POST /api/v1/categories", authMW(http.HandlerFunc(deps.CategoryHandler.Create)).ServeHTTP)

	// 更新分类：需认证
	mux.HandleFunc("PUT /api/v1/categories/{id}", authMW(http.HandlerFunc(deps.CategoryHandler.Update)).ServeHTTP)

	// 删除分类：需认证
	mux.HandleFunc("DELETE /api/v1/categories/{id}", authMW(http.HandlerFunc(deps.CategoryHandler.Delete)).ServeHTTP)
}

// setupTagRoutes 配置标签相关路由。
//
// 注册标签的增删改查 API 路由。
// 公开路由无中间件保护，修改操作需要认证。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 路由列表：
//   - GET /api/v1/tags: 标签列表（公开）
//   - GET /api/v1/tags/{slug}: 获取单个标签（公开）
//   - POST /api/v1/tags: 创建标签（需认证）
//   - DELETE /api/v1/tags/{id}: 删除标签（需认证）
func setupTagRoutes(mux *http.ServeMux, deps *RouteDeps) {
	// 创建认证中间件
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	// 标签列表：公开访问
	mux.HandleFunc("GET /api/v1/tags", deps.TagHandler.List)

	// 获取标签详情：通过 slug 查询
	mux.HandleFunc("GET /api/v1/tags/{slug}", deps.TagHandler.Get)

	// 创建标签：需认证
	mux.HandleFunc("POST /api/v1/tags", authMW(http.HandlerFunc(deps.TagHandler.Create)).ServeHTTP)

	// 删除标签：需认证
	mux.HandleFunc("DELETE /api/v1/tags/{id}", authMW(http.HandlerFunc(deps.TagHandler.Delete)).ServeHTTP)
}

// setupCommentRoutes 配置评论相关路由。
//
// 注册评论的 CRUD、审核和批量操作 API 路由。
// 审核相关路由需要 comment:moderate 权限。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 公开路由列表：
//   - GET /api/v1/comments/post/{postId}: 获取文章的评论列表
//
// 认证路由列表：
//   - POST /api/v1/comments: 创建评论
//   - PUT /api/v1/comments/{id}: 更新评论
//   - DELETE /api/v1/comments/{id}: 删除评论
//   - POST /api/v1/comments/{id}/like: 点赞评论
//   - DELETE /api/v1/comments/{id}/like: 取消点赞
//
// 审核路由列表（需 comment:moderate 权限）：
//   - PUT /api/v1/comments/{id}/approve: 批准评论
//   - PUT /api/v1/comments/{id}/reject: 拒绝评论
//   - GET /api/v1/admin/comments: 管理后台评论列表
//   - PUT /api/v1/admin/comments/batch-approve: 批量批准
//   - PUT /api/v1/admin/comments/batch-reject: 批量拒绝
//   - DELETE /api/v1/admin/comments/batch-delete: 批量删除
//   - DELETE /api/v1/admin/comments/{id}: 管理后台删除评论
func setupCommentRoutes(mux *http.ServeMux, deps *RouteDeps) {
	// 创建认证中间件
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	// 创建权限中间件：检查评论审核权限
	permMW := handlers.CachedPermissionMiddleware(deps.JWTService, deps.PermCache, "comment:moderate")

	// === 公开路由 ===
	// 获取文章评论列表：按文章 ID 查询
	mux.HandleFunc("GET /api/v1/comments/post/{postId}", deps.CommentHandler.GetByPost)

	// === 需认证路由 ===
	// 创建评论
	mux.HandleFunc("POST /api/v1/comments", authMW(http.HandlerFunc(deps.CommentHandler.Create)).ServeHTTP)

	// 更新评论
	mux.HandleFunc("PUT /api/v1/comments/{id}", authMW(http.HandlerFunc(deps.CommentHandler.Update)).ServeHTTP)

	// 删除评论
	mux.HandleFunc("DELETE /api/v1/comments/{id}", authMW(http.HandlerFunc(deps.CommentHandler.Delete)).ServeHTTP)

	// 点赞评论
	mux.HandleFunc("POST /api/v1/comments/{id}/like", authMW(http.HandlerFunc(deps.CommentHandler.Like)).ServeHTTP)

	// 取消点赞评论
	mux.HandleFunc("DELETE /api/v1/comments/{id}/like", authMW(http.HandlerFunc(deps.CommentHandler.Unlike)).ServeHTTP)

	// === 评论审核路由（需权限）===
	// 批准单条评论
	mux.HandleFunc("PUT /api/v1/comments/{id}/approve", permMW(http.HandlerFunc(deps.CommentHandler.Approve)).ServeHTTP)

	// 拒绝单条评论
	mux.HandleFunc("PUT /api/v1/comments/{id}/reject", permMW(http.HandlerFunc(deps.CommentHandler.Reject)).ServeHTTP)

	// === 评论管理后台路由 ===
	// 管理后台评论列表
	mux.HandleFunc("GET /api/v1/admin/comments", permMW(http.HandlerFunc(deps.CommentHandler.AdminListComments)).ServeHTTP)

	// 批量批准评论
	mux.HandleFunc("PUT /api/v1/admin/comments/batch-approve", permMW(http.HandlerFunc(deps.CommentHandler.BatchApprove)).ServeHTTP)

	// 批量拒绝评论
	mux.HandleFunc("PUT /api/v1/admin/comments/batch-reject", permMW(http.HandlerFunc(deps.CommentHandler.BatchReject)).ServeHTTP)

	// 批量删除评论
	mux.HandleFunc("DELETE /api/v1/admin/comments/batch-delete", permMW(http.HandlerFunc(deps.CommentHandler.BatchDelete)).ServeHTTP)

	// 管理后台删除单条评论
	mux.HandleFunc("DELETE /api/v1/admin/comments/{id}", permMW(http.HandlerFunc(deps.CommentHandler.AdminDeleteComment)).ServeHTTP)
}

// setupMediaRoutes 配置媒体文件相关路由。
//
// 注册媒体文件的上传、查询和删除 API 路由。
// 所有媒体操作都需要用户认证。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 路由列表（均需认证）：
//   - POST /api/v1/media/upload: 上传媒体文件
//   - DELETE /api/v1/media/{id}: 删除媒体文件
//   - GET /api/v1/media: 获取媒体文件列表
//   - GET /api/v1/media/{id}: 获取单个媒体文件信息
func setupMediaRoutes(mux *http.ServeMux, deps *RouteDeps) {
	// 创建认证中间件
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	// 上传媒体文件
	mux.HandleFunc("POST /api/v1/media/upload", authMW(http.HandlerFunc(deps.MediaHandler.Upload)).ServeHTTP)

	// 删除媒体文件
	mux.HandleFunc("DELETE /api/v1/media/{id}", authMW(http.HandlerFunc(deps.MediaHandler.Delete)).ServeHTTP)

	// 获取媒体文件列表
	mux.HandleFunc("GET /api/v1/media", authMW(http.HandlerFunc(deps.MediaHandler.List)).ServeHTTP)

	// 获取单个媒体文件信息
	mux.HandleFunc("GET /api/v1/media/{id}", authMW(http.HandlerFunc(deps.MediaHandler.Get)).ServeHTTP)
}

// setupRSSRoutes 配置 RSS 订阅路由。
//
// 注册 RSS/Atom Feed 生成 API 路由。
// 该路由为公开访问，无需认证。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 路由列表：
//   - GET /api/v1/rss: 获取 RSS Feed
func setupRSSRoutes(mux *http.ServeMux, deps *RouteDeps) {
	// RSS Feed：公开访问
	mux.HandleFunc("GET /api/v1/rss", deps.RSSHandler.Feed)
}

// setupSearchRoutes 配置搜索相关路由。
//
// 注册全文搜索和自动补全 API 路由。
// 应用公开限流器防止滥用。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 路由列表（限流保护）：
//   - GET /api/v1/search: 全文搜索
//   - GET /api/v1/search/suggestions: 搜索自动补全
func setupSearchRoutes(mux *http.ServeMux, deps *RouteDeps) {
	// 创建公开限流中间件
	publicRL := middleware.RateLimitMiddleware(deps.PublicLimiter, middleware.IPKeyFunc("public"))

	// 全文搜索
	mux.HandleFunc("GET /api/v1/search", publicRL(http.HandlerFunc(deps.SearchHandler.Search)).ServeHTTP)

	// 搜索自动补全
	mux.HandleFunc("GET /api/v1/search/suggestions", publicRL(http.HandlerFunc(deps.SearchHandler.Suggestions)).ServeHTTP)
}

// setupAdminRoutes 配置管理员 API 路由。
//
// 注册角色管理、用户管理和批量操作等后台管理 API。
// 所有路由需同时通过认证和管理员权限中间件。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 角色管理路由列表：
//   - GET /api/v1/admin/roles: 角色列表
//   - POST /api/v1/admin/roles: 创建角色
//   - PUT /api/v1/admin/roles/{id}: 更新角色
//   - DELETE /api/v1/admin/roles/{id}: 删除角色
//
// 用户管理路由列表：
//   - GET /api/v1/admin/users: 用户列表
//   - PUT /api/v1/admin/users/{id}/ban: 封禁用户
//
// 其他管理路由：
//   - POST /api/v1/admin/batch: 批量操作
//   - PUT /api/v1/admin/order: 排序更新
func setupAdminRoutes(mux *http.ServeMux, deps *RouteDeps) {
	// 创建认证中间件
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	// 创建管理员权限中间件
	adminMW := handlers.AdminMiddleware(deps.PermCache)

	// === 角色管理路由 ===
	// 角色列表
	mux.Handle("GET /api/v1/admin/roles", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.ListRoles))))

	// 创建角色
	mux.Handle("POST /api/v1/admin/roles", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.CreateRole))))

	// 更新角色
	mux.Handle("PUT /api/v1/admin/roles/{id}", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.UpdateRole))))

	// 删除角色
	mux.Handle("DELETE /api/v1/admin/roles/{id}", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.DeleteRole))))

	// === 用户管理路由 ===
	// 用户列表
	mux.Handle("GET /api/v1/admin/users", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.ListUsers))))

	// 封禁用户
	mux.Handle("PUT /api/v1/admin/users/{id}/ban", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.BanUser))))

	// === 批量操作路由 ===
	// 执行批量操作
	mux.Handle("POST /api/v1/admin/batch", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.BatchOperation))))

	// === 排序更新路由 ===
	// 更新资源排序顺序
	mux.Handle("PUT /api/v1/admin/order", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.UpdateOrder))))
}

// setupAdminPages 配置管理后台页面路由。
//
// 注册管理后台的 Web 页面路由，使用 templ 组件渲染。
// 所有路由需同时通过认证和管理员权限中间件。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 页面路由列表：
//   - GET /admin: 仪表盘页面，显示统计数据和最近内容
//   - GET /admin/comments: 评论管理页面，支持状态过滤和分页
func setupAdminPages(mux *http.ServeMux, deps *RouteDeps) {
	// 创建认证中间件
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	// 创建管理员权限中间件
	adminMW := handlers.AdminMiddleware(deps.PermCache)

	// === 仪表盘页面 ===
	// 显示统计数据：文章数、用户数、待审核评论数等
	mux.Handle("GET /admin", authMW(adminMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 获取统计数据
		filters := post.PostListFilters{}
		recentPosts, postCount, _ := deps.PostRepo.List(ctx, filters, 0, 5)

		commentFilters := &comment.CommentListFilters{}
		recentComments, _ := deps.CommentRepo.List(ctx, commentFilters, 0, 5)

		pendingFilters := &comment.CommentListFilters{Status: comment.StatusPending}
		pendingComments, _ := deps.CommentRepo.List(ctx, pendingFilters, 0, 100)
		pendingCount := len(pendingComments)

		// 获取用户数
		_, userCount, _ := deps.Services.UserService.List(ctx, 0, 1)

		stats := adminpages.DashboardStats{
			TotalPosts:    postCount,
			TotalViews:    0,
			TotalComments: 0,
			TotalUsers:    userCount,
			ActiveTheme:   "default",
		}

		templ.Handler(adminpages.DashboardPage(stats, recentPosts, recentComments, pendingCount)).ServeHTTP(w, r)
	}))))

	// === 评论管理页面 ===
	// 支持按状态过滤（pending/approved/rejected）和分页
	mux.Handle("GET /admin/comments", authMW(adminMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取状态过滤参数，默认为 pending
		status := r.URL.Query().Get("status")
		if status == "" {
			status = "pending"
		}
		pageStr := r.URL.Query().Get("page")
		page := 1
		if pageStr != "" {
			if p, err := parseIntStr(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		perPage := 20

		comments, total, err := deps.Services.CommentService.GetCommentsByStatus(r.Context(), comment.CommentStatus(status), (page-1)*perPage, perPage)
		if err != nil {
			http.Error(w, "获取评论失败", http.StatusInternalServerError)
			return
		}

		templ.Handler(adminpages.AdminCommentsPage("评论管理 - Cadmus", status, comments, page, perPage, total)).ServeHTTP(w, r)
	}))))
}

// setupStaticRoutes 配置静态文件服务路由。
//
// 注册静态资源和上传文件的文件服务路由。
// 使用 http.FileServer 提供文件服务，支持目录浏览。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象
//
// 路由列表：
//   - /static/: 静态资源目录（CSS、JS、图片等），映射到 web/static
//   - /uploads/: 用户上传文件目录，映射到配置的 UploadDir
//
// 注意事项：
//   - 静态文件路由为公开访问，无认证保护
//   - 使用 http.StripPrefix 去除 URL 前缀以正确映射文件路径
func setupStaticRoutes(mux *http.ServeMux, deps *RouteDeps) {
	// 静态资源服务：CSS、JS、图片等
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// 上传文件服务：用户上传的媒体文件
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(deps.UploadDir))))
}

// setupHealthRoute 配置健康检查路由。
//
// 注册服务健康检查端点，用于负载均衡器或监控系统探测服务状态。
// 该路由为公开访问，无需认证，返回简单的 OK 响应。
//
// 参数：
//   - mux: HTTP 路由映射器
//   - deps: 路由依赖聚合对象（未使用，保留用于未来扩展）
//
// 路由列表：
//   - GET /health: 健康检查端点，返回 200 OK
//
// 返回格式：
//
//	HTTP 200 状态码，响应体为 "OK"
//
// 注意事项：
//   - 该路由不检查数据库或缓存连接状态
//   - 未来可扩展为返回详细组件状态
func setupHealthRoute(mux *http.ServeMux, deps *RouteDeps) {
	// 健康检查：返回简单的 OK 响应
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Printf("Failed to write health check response: %v", err)
		}
	})
}
