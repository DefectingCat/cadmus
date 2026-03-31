package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a-h/templ"
	"rua.plus/cadmus/internal/api/handlers"
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/cache"
	"rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/internal/core/rss"
	"rua.plus/cadmus/internal/database"
	"rua.plus/cadmus/internal/services"
	_ "rua.plus/cadmus/plugins/mermaid-block" // 启用 Mermaid 图表块插件（blank import 触发 init）
	"rua.plus/cadmus/web/templates/pages"
	adminpages "rua.plus/cadmus/web/templates/pages/admin"

	// 启用默认主题（blank import 触发 init()）
	_ "rua.plus/cadmus/themes/default"
)

func main() {
	// 获取端口配置
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 初始化数据库连接池
	dbCfg := database.DefaultConfig()
	dbCfg.Host = getEnvOrDefault("DB_HOST", "localhost")
	dbCfg.Port = atoi(getEnvOrDefault("DB_PORT", "5432"))
	dbCfg.Name = getEnvOrDefault("DB_NAME", "cadmus")
	dbCfg.User = getEnvOrDefault("DB_USER", "cadmus")
	dbCfg.Password = getEnvOrDefault("DB_PASSWORD", "")
	dbCfg.SSLMode = getEnvOrDefault("DB_SSLMODE", "disable")

	ctx := context.Background()
	pool, err := database.NewPool(ctx, dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Database connection pool initialized")

	// 初始化 Redis 连接池
	redisCfg := cache.DefaultConfig()
	redisCfg.Host = getEnvOrDefault("REDIS_HOST", "localhost")
	redisCfg.Port = atoi(getEnvOrDefault("REDIS_PORT", "6379"))
	redisCfg.Password = getEnvOrDefault("REDIS_PASSWORD", "")

	redisClient, err := cache.NewRedisClient(redisCfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connection pool initialized")

	// 初始化缓存服务
	cacheService := cache.NewService(redisClient)
	log.Println("Cache service initialized")

	// 初始化 repositories
	userRepo := database.NewUserRepository(pool)
	roleRepo := database.NewRoleRepository(pool)
	permRepo := database.NewPermissionRepository(pool)

	// 文章相关 repositories
	postRepo := database.NewPostRepository(pool)
	categoryRepo := database.NewCategoryRepository(pool)
	tagRepo := database.NewTagRepository(pool)
	seriesRepo := database.NewSeriesRepository(pool)

	// 评论相关 repositories
	commentRepo := database.NewCommentRepository(pool)
	commentLikeRepo := database.NewCommentLikeRepository(pool)

	// 文章点赞 repository
	postLikeRepo := database.NewPostLikeRepository(pool)

	// 媒体相关 repositories
	mediaRepo := database.NewMediaRepository(pool)

	// 搜索相关 repositories
	searchRepo := database.NewSearchRepository(pool)
	log.Println("Repositories initialized")

	// 初始化 JWT 服务
	jwtCfg, err := auth.DefaultJWTConfig()
	if err != nil {
		log.Fatalf("JWT config error: %v", err)
	}
	jwtService := auth.NewJWTService(jwtCfg)
	log.Println("JWT service initialized")

	// 初始化 token 黑名单
	tokenBlacklist := auth.NewRedisTokenBlacklist(redisClient)
	log.Println("Token blacklist initialized")

	// 初始化权限缓存
	permCache := auth.NewPermissionCache(cacheService, permRepo, redisClient.Client())
	log.Println("Permission cache initialized")

	// 初始化 Service 容器（带媒体服务）
	uploadDir := getEnvOrDefault("UPLOAD_DIR", "./uploads")
	baseURL := getEnvOrDefault("BASE_URL", "http://localhost:"+port)
	serviceContainer := services.NewContainerWithMedia(
		userRepo, roleRepo, jwtService, tokenBlacklist,
		postRepo, categoryRepo, tagRepo, seriesRepo,
		commentRepo, commentLikeRepo,
		mediaRepo, uploadDir, baseURL, postLikeRepo, searchRepo,
	)
	log.Println("Service container initialized")

	// 初始化认证处理器（通过 Service 层）
	authHandler := handlers.NewAuthHandlerWithServices(
		serviceContainer.AuthService,
		serviceContainer.UserService,
		serviceContainer.JWTService(),
		roleRepo,
	)
	log.Println("Auth handlers initialized")

	// 初始化文章相关处理器
	postHandler := handlers.NewPostHandler(serviceContainer.PostService)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	tagHandler := handlers.NewTagHandler(tagRepo)
	log.Println("Post handlers initialized")

	// 初始化评论处理器
	commentHandler := handlers.NewCommentHandler(serviceContainer.CommentService)
	log.Println("Comment handlers initialized")

	// 初始化媒体处理器
	mediaHandler := handlers.NewMediaHandler(serviceContainer.MediaService)
	log.Println("Media handlers initialized")

	// 初始化 RSS 处理器
	rssConfig := rss.DefaultFeedConfig()
	rssConfig.BaseURL = baseURL + "/posts"
	rssConfig.Link = baseURL
	rssHandler := handlers.NewRSSHandler(serviceContainer.RSSService, rssConfig)
	log.Println("RSS handlers initialized")

	// 初始化搜索处理器
	searchHandler := handlers.NewSearchHandler(serviceContainer.SearchService)
	log.Println("Search handlers initialized")

	// 初始化管理员处理器
	adminHandler := handlers.NewAdminHandler(
		userRepo, roleRepo, permRepo,
		postRepo, categoryRepo, tagRepo, commentRepo,
		serviceContainer.UserService,
		jwtService,
		permCache,
	)
	log.Println("Admin handlers initialized")

	// 创建路由
	mux := http.NewServeMux()

	// API 路由组
	// 认证 API
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(authHandler.Logout)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("GET /api/v1/auth/me", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(authHandler.Me)).ServeHTTP)

	// 文章 API（公开）
	mux.HandleFunc("GET /api/v1/posts", postHandler.List)
	mux.HandleFunc("GET /api/v1/posts/{slug}", postHandler.Get)
	mux.HandleFunc("GET /api/v1/posts/{id}/versions", postHandler.Versions)
	mux.HandleFunc("GET /api/v1/users/{id}/posts", postHandler.GetUserPosts)

	// 搜索 API（公开）
	mux.HandleFunc("GET /api/v1/search", searchHandler.Search)
	mux.HandleFunc("GET /api/v1/search/suggestions", searchHandler.Suggestions)

	// 文章 API（需认证）
	mux.HandleFunc("POST /api/v1/posts", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(postHandler.Create)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/posts/{id}", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(postHandler.Update)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/posts/{id}", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(postHandler.Delete)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/posts/{id}/publish", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(postHandler.Publish)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/posts/{id}/rollback", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(postHandler.Rollback)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/posts/{id}/like", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(postHandler.Like)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/posts/{id}/like", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(postHandler.Unlike)).ServeHTTP)

	// 分类 API
	mux.HandleFunc("GET /api/v1/categories", categoryHandler.List)
	mux.HandleFunc("GET /api/v1/categories/{slug}", categoryHandler.Get)
	mux.HandleFunc("POST /api/v1/categories", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(categoryHandler.Create)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/categories/{id}", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(categoryHandler.Update)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/categories/{id}", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(categoryHandler.Delete)).ServeHTTP)

	// 标签 API
	mux.HandleFunc("GET /api/v1/tags", tagHandler.List)
	mux.HandleFunc("GET /api/v1/tags/{slug}", tagHandler.Get)
	mux.HandleFunc("POST /api/v1/tags", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(tagHandler.Create)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/tags/{id}", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(tagHandler.Delete)).ServeHTTP)

	// 评论 API（公开）
	mux.HandleFunc("GET /api/v1/comments/post/{postId}", commentHandler.GetByPost)

	// 评论 API（需认证）
	mux.HandleFunc("POST /api/v1/comments", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(commentHandler.Create)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/comments/{id}", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(commentHandler.Update)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/comments/{id}", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(commentHandler.Delete)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/comments/{id}/like", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(commentHandler.Like)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/comments/{id}/like", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(commentHandler.Unlike)).ServeHTTP)

	// 评论审核 API（需认证，需 comment:moderate 权限）
	mux.HandleFunc("PUT /api/v1/comments/{id}/approve", handlers.CachedPermissionMiddleware(jwtService, permCache, "comment:moderate")(http.HandlerFunc(commentHandler.Approve)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/comments/{id}/reject", handlers.CachedPermissionMiddleware(jwtService, permCache, "comment:moderate")(http.HandlerFunc(commentHandler.Reject)).ServeHTTP)

	// 评论管理后台 API（需 comment:moderate 权限）
	mux.HandleFunc("GET /api/v1/admin/comments", handlers.CachedPermissionMiddleware(jwtService, permCache, "comment:moderate")(http.HandlerFunc(commentHandler.AdminListComments)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/admin/comments/batch-approve", handlers.CachedPermissionMiddleware(jwtService, permCache, "comment:moderate")(http.HandlerFunc(commentHandler.BatchApprove)).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/admin/comments/batch-reject", handlers.CachedPermissionMiddleware(jwtService, permCache, "comment:moderate")(http.HandlerFunc(commentHandler.BatchReject)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/admin/comments/batch-delete", handlers.CachedPermissionMiddleware(jwtService, permCache, "comment:moderate")(http.HandlerFunc(commentHandler.BatchDelete)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/admin/comments/{id}", handlers.CachedPermissionMiddleware(jwtService, permCache, "comment:moderate")(http.HandlerFunc(commentHandler.AdminDeleteComment)).ServeHTTP)

	// 媒体 API（需认证）
	mux.HandleFunc("POST /api/v1/media/upload", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(mediaHandler.Upload)).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/media/{id}", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(mediaHandler.Delete)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/media", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(mediaHandler.List)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/media/{id}", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(mediaHandler.Get)).ServeHTTP)

	// RSS 订阅 API（公开）
	mux.HandleFunc("GET /api/v1/rss", rssHandler.Feed)

	// 管理员 API（需认证和管理员权限）
	adminAuth := handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)
	adminPerm := handlers.AdminMiddleware(permCache)

	// 角色管理
	mux.HandleFunc("GET /api/v1/admin/roles", adminAuth(adminPerm(http.HandlerFunc(adminHandler.ListRoles))).ServeHTTP)
	mux.HandleFunc("POST /api/v1/admin/roles", adminAuth(adminPerm(http.HandlerFunc(adminHandler.CreateRole))).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/admin/roles/{id}", adminAuth(adminPerm(http.HandlerFunc(adminHandler.UpdateRole))).ServeHTTP)
	mux.HandleFunc("DELETE /api/v1/admin/roles/{id}", adminAuth(adminPerm(http.HandlerFunc(adminHandler.DeleteRole))).ServeHTTP)

	// 用户管理
	mux.HandleFunc("GET /api/v1/admin/users", adminAuth(adminPerm(http.HandlerFunc(adminHandler.ListUsers))).ServeHTTP)
	mux.HandleFunc("PUT /api/v1/admin/users/{id}/ban", adminAuth(adminPerm(http.HandlerFunc(adminHandler.BanUser))).ServeHTTP)

	// 批量操作
	mux.HandleFunc("POST /api/v1/admin/batch", adminAuth(adminPerm(http.HandlerFunc(adminHandler.BatchOperation))).ServeHTTP)

	// 排序更新
	mux.HandleFunc("PUT /api/v1/admin/order", adminAuth(adminPerm(http.HandlerFunc(adminHandler.UpdateOrder))).ServeHTTP)

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 首页
	mux.Handle("/", templ.Handler(pages.HomePage("Cadmus - 博客平台")))

	// 管理后台页面（需要认证和管理员权限）
	adminPageAuth := handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)
	adminPagePerm := handlers.AdminMiddleware(permCache)
	mux.Handle("GET /admin/comments", adminPageAuth(adminPagePerm(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取查询参数
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

		// 获取评论列表
		comments, total, err := serviceContainer.CommentService.GetCommentsByStatus(r.Context(), comment.CommentStatus(status), (page-1)*perPage, perPage)
		if err != nil {
			http.Error(w, "获取评论失败", http.StatusInternalServerError)
			return
		}

		templ.Handler(adminpages.AdminCommentsPage("评论管理 - Cadmus", status, comments, page, perPage, total)).ServeHTTP(w, r)
	}))))

	// 静态文件服务
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// 上传文件静态服务
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	// 创建 HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器（在 goroutine 中）
	go func() {
		log.Printf("Cadmus server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// atoi 字符串转整数
func atoi(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

// parseIntStr 解析整数字符串
func parseIntStr(s string) (int, error) {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return 0, nil
		}
	}
	return n, nil
}

