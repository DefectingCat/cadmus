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
	"rua.plus/cadmus/internal/services"
	adminpages "rua.plus/cadmus/web/templates/pages/admin"
)

// RouteDeps 路由依赖聚合
type RouteDeps struct {
	// Handlers
	AuthHandler    *handlers.AuthHandler
	PostHandler    *handlers.PostHandler
	CategoryHandler *handlers.CategoryHandler
	TagHandler     *handlers.TagHandler
	CommentHandler *handlers.CommentHandler
	MediaHandler   *handlers.MediaHandler
	RSSHandler     *handlers.RSSHandler
	SearchHandler  *handlers.SearchHandler
	AdminHandler   *handlers.AdminHandler

	// Middleware
	LoginLimiter  *middleware.RateLimiter
	PublicLimiter *middleware.RateLimiter
	UserLimiter   *middleware.RateLimiter

	// Auth
	JWTService    *auth.JWTService
	TokenBlacklist auth.TokenBlacklist
	PermCache     *auth.PermissionCache

	// Repositories (for admin pages)
	PostRepo    *database.PostRepository
	CommentRepo *database.CommentRepository

	// Services
	Services *services.Container

	// Config
	UploadDir string
}

// setupRoutes 配置所有路由
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

// setupAuthRoutes 认证路由
func setupAuthRoutes(mux *http.ServeMux, deps *RouteDeps) {
	loginRL := middleware.RateLimitMiddleware(deps.LoginLimiter, middleware.IPKeyFunc("login"))
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	mux.HandleFunc("POST /api/v1/auth/register", loginRL(http.HandlerFunc(deps.AuthHandler.Register)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/auth/login", loginRL(http.HandlerFunc(deps.AuthHandler.Login)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/auth/logout", authMW(http.HandlerFunc(deps.AuthHandler.Logout)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/auth/refresh", deps.AuthHandler.Refresh)
	mux.HandleFunc("GET /api/v1/auth/me", authMW(http.HandlerFunc(deps.AuthHandler.Me)).ServeHTTP)
}

// setupPostRoutes 文章路由
func setupPostRoutes(mux *http.ServeMux, deps *RouteDeps) {
	publicRL := middleware.RateLimitMiddleware(deps.PublicLimiter, middleware.IPKeyFunc("public"))
	userRL := middleware.RateLimitMiddleware(deps.UserLimiter, middleware.IPKeyFunc("user"))
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	// 公开路由
	mux.HandleFunc("GET /api/v1/posts", publicRL(http.HandlerFunc(deps.PostHandler.List)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/posts/{slug}", publicRL(http.HandlerFunc(deps.PostHandler.Get)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/posts/{id}/versions", publicRL(http.HandlerFunc(deps.PostHandler.Versions)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/users/{id}/posts", publicRL(http.HandlerFunc(deps.PostHandler.GetUserPosts)).ServeHTTP)

	// 需认证路由
	mux.HandleFunc("POST /api/v1/posts", userRL(authMW(http.HandlerFunc(deps.PostHandler.Create))).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/posts/{id}", userRL(authMW(http.HandlerFunc(deps.PostHandler.Update))).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/posts/{id}", userRL(authMW(http.HandlerFunc(deps.PostHandler.Delete))).ServeHTTP)
	mux.HandleFunc("POST /api/v1/posts/{id}/publish", userRL(authMW(http.HandlerFunc(deps.PostHandler.Publish))).ServeHTTP)
	mux.HandleFunc("POST /api/v1/posts/{id}/rollback", userRL(authMW(http.HandlerFunc(deps.PostHandler.Rollback))).ServeHTTP)
	mux.HandleFunc("POST /api/v1/posts/{id}/like", userRL(authMW(http.HandlerFunc(deps.PostHandler.Like))).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/posts/{id}/like", userRL(authMW(http.HandlerFunc(deps.PostHandler.Unlike))).ServeHTTP)
}

// setupCategoryRoutes 分类路由
func setupCategoryRoutes(mux *http.ServeMux, deps *RouteDeps) {
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	mux.HandleFunc("GET /api/v1/categories", deps.CategoryHandler.List)
	mux.HandleFunc("GET /api/v1/categories/{slug}", deps.CategoryHandler.Get)
	mux.HandleFunc("POST /api/v1/categories", authMW(http.HandlerFunc(deps.CategoryHandler.Create)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/categories/{id}", authMW(http.HandlerFunc(deps.CategoryHandler.Update)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/categories/{id}", authMW(http.HandlerFunc(deps.CategoryHandler.Delete)).ServeHTTP)
}

// setupTagRoutes 标签路由
func setupTagRoutes(mux *http.ServeMux, deps *RouteDeps) {
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	mux.HandleFunc("GET /api/v1/tags", deps.TagHandler.List)
	mux.HandleFunc("GET /api/v1/tags/{slug}", deps.TagHandler.Get)
	mux.HandleFunc("POST /api/v1/tags", authMW(http.HandlerFunc(deps.TagHandler.Create)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/tags/{id}", authMW(http.HandlerFunc(deps.TagHandler.Delete)).ServeHTTP)
}

// setupCommentRoutes 评论路由
func setupCommentRoutes(mux *http.ServeMux, deps *RouteDeps) {
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)
	permMW := handlers.CachedPermissionMiddleware(deps.JWTService, deps.PermCache, "comment:moderate")

	// 公开路由
	mux.HandleFunc("GET /api/v1/comments/post/{postId}", deps.CommentHandler.GetByPost)

	// 需认证路由
	mux.HandleFunc("POST /api/v1/comments", authMW(http.HandlerFunc(deps.CommentHandler.Create)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/comments/{id}", authMW(http.HandlerFunc(deps.CommentHandler.Update)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/comments/{id}", authMW(http.HandlerFunc(deps.CommentHandler.Delete)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/comments/{id}/like", authMW(http.HandlerFunc(deps.CommentHandler.Like)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/comments/{id}/like", authMW(http.HandlerFunc(deps.CommentHandler.Unlike)).ServeHTTP)

	// 评论审核路由
	mux.HandleFunc("PUT /api/v1/comments/{id}/approve", permMW(http.HandlerFunc(deps.CommentHandler.Approve)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/comments/{id}/reject", permMW(http.HandlerFunc(deps.CommentHandler.Reject)).ServeHTTP)

	// 评论管理后台路由
	mux.HandleFunc("GET /api/v1/admin/comments", permMW(http.HandlerFunc(deps.CommentHandler.AdminListComments)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/admin/comments/batch-approve", permMW(http.HandlerFunc(deps.CommentHandler.BatchApprove)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/admin/comments/batch-reject", permMW(http.HandlerFunc(deps.CommentHandler.BatchReject)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/admin/comments/batch-delete", permMW(http.HandlerFunc(deps.CommentHandler.BatchDelete)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/admin/comments/{id}", permMW(http.HandlerFunc(deps.CommentHandler.AdminDeleteComment)).ServeHTTP)
}

// setupMediaRoutes 媒体路由
func setupMediaRoutes(mux *http.ServeMux, deps *RouteDeps) {
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)

	mux.HandleFunc("POST /api/v1/media/upload", authMW(http.HandlerFunc(deps.MediaHandler.Upload)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/media/{id}", authMW(http.HandlerFunc(deps.MediaHandler.Delete)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/media", authMW(http.HandlerFunc(deps.MediaHandler.List)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/media/{id}", authMW(http.HandlerFunc(deps.MediaHandler.Get)).ServeHTTP)
}

// setupRSSRoutes RSS 路由
func setupRSSRoutes(mux *http.ServeMux, deps *RouteDeps) {
	mux.HandleFunc("GET /api/v1/rss", deps.RSSHandler.Feed)
}

// setupSearchRoutes 搜索路由
func setupSearchRoutes(mux *http.ServeMux, deps *RouteDeps) {
	publicRL := middleware.RateLimitMiddleware(deps.PublicLimiter, middleware.IPKeyFunc("public"))

	mux.HandleFunc("GET /api/v1/search", publicRL(http.HandlerFunc(deps.SearchHandler.Search)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/search/suggestions", publicRL(http.HandlerFunc(deps.SearchHandler.Suggestions)).ServeHTTP)
}

// setupAdminRoutes 管理员 API 路由
func setupAdminRoutes(mux *http.ServeMux, deps *RouteDeps) {
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)
	adminMW := handlers.AdminMiddleware(deps.PermCache)

	// 角色管理
	mux.Handle("GET /api/v1/admin/roles", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.ListRoles))))
	mux.Handle("POST /api/v1/admin/roles", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.CreateRole))))
	mux.Handle("PUT /api/v1/admin/roles/{id}", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.UpdateRole))))
	mux.Handle("DELETE /api/v1/admin/roles/{id}", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.DeleteRole))))

	// 用户管理
	mux.Handle("GET /api/v1/admin/users", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.ListUsers))))
	mux.Handle("PUT /api/v1/admin/users/{id}/ban", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.BanUser))))

	// 批量操作
	mux.Handle("POST /api/v1/admin/batch", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.BatchOperation))))

	// 排序更新
	mux.Handle("PUT /api/v1/admin/order", authMW(adminMW(http.HandlerFunc(deps.AdminHandler.UpdateOrder))))
}

// setupAdminPages 管理后台页面路由
func setupAdminPages(mux *http.ServeMux, deps *RouteDeps) {
	authMW := handlers.AuthMiddlewareWithBlacklist(deps.JWTService, deps.TokenBlacklist)
	adminMW := handlers.AdminMiddleware(deps.PermCache)

	// 仪表盘
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

	// 评论管理
	mux.Handle("GET /admin/comments", authMW(adminMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

// setupStaticRoutes 静态文件路由
func setupStaticRoutes(mux *http.ServeMux, deps *RouteDeps) {
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(deps.UploadDir))))
}

// setupHealthRoute 健康检查路由
func setupHealthRoute(mux *http.ServeMux, deps *RouteDeps) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}