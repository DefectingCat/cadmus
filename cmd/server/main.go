// Package main 提供 Cadmus 博客平台的服务器入口点。
//
// 该文件包含服务器启动和生命周期管理的核心逻辑，包括：
//   - 版本信息的显示和注入
//   - 配置的加载和验证
//   - 数据库和缓存连接的初始化
//   - 服务容器和处理器依赖链的构建
//   - HTTP 服务器的配置、启动和优雅关闭
//
// 主要用途：
//
//	作为 Cadmus 博客平台的唯一入口点，负责协调所有基础设施组件的初始化，
//	并提供 HTTP API 服务。
//
// 注意事项：
//   - 版本信息通过 -ldflags 在编译时注入，运行时默认值为 "dev"
//   - 服务器支持优雅关闭，最大等待时间为 10 秒
//   - 所有 blank import 用于触发插件的 init() 函数
//
// 作者：xfy
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rua.plus/cadmus/internal/api/handlers"
	"rua.plus/cadmus/internal/api/middleware"
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/cache"
	"rua.plus/cadmus/internal/core/rss"
	"rua.plus/cadmus/internal/database"
	"rua.plus/cadmus/internal/logger"
	"rua.plus/cadmus/internal/services"
	_ "rua.plus/cadmus/plugins/mermaid-block" // 启用 Mermaid 图表块插件（blank import 触发 init）
	_ "rua.plus/cadmus/themes/default"        // 启用默认主题（blank import 触发 init()）
)

// 版本信息变量（通过 -ldflags 在编译时注入）
//
// 这些变量在编译时通过 -ldflags 参数注入实际值，用于追踪
// 构建版本、Git 提交、构建时间等信息。未注入时使用默认值。
var (
	// version 应用版本号，格式如 "v1.0.0"
	version = "dev"

	// gitCommit Git 提交哈希，用于追踪具体代码版本
	gitCommit = "unknown"

	// gitBranch Git 分支名，用于识别构建来源
	gitBranch = "unknown"

	// buildTime 构建时间，格式为 RFC3339
	buildTime = "unknown"

	// goVersion 编译使用的 Go 版本
	goVersion = "unknown"

	// buildPlatform 目标平台，格式如 "darwin/amd64"
	buildPlatform = "unknown"
)

// printVersionInfo 打印应用版本信息到标准输出。
//
// 该函数用于在服务器启动时显示构建信息，便于运维人员
// 确认当前运行版本。输出格式为多行文本，包含所有版本变量。
//
// 输出格式：
//   Cadmus Blog Platform
//     Version:    {version}
//     Git Commit: {gitCommit}
//     Git Branch: {gitBranch}
//     Build Time: {buildTime}
//     Go Version: {goVersion}
//     Platform:   {buildPlatform}
func printVersionInfo() {
	fmt.Println("Cadmus Blog Platform")
	fmt.Printf("  Version:    %s\n", version)
	fmt.Printf("  Git Commit: %s\n", gitCommit)
	fmt.Printf("  Git Branch: %s\n", gitBranch)
	fmt.Printf("  Build Time: %s\n", buildTime)
	fmt.Printf("  Go Version: %s\n", goVersion)
	fmt.Printf("  Platform:   %s\n", buildPlatform)
}

// main 是 Cadmus 博客平台的程序入口。
//
// 该函数负责协调所有组件的初始化和启动，执行以下步骤：
//   1. 显示版本信息
//   2. 加载配置（数据库、Redis、JWT、服务器等）
//   3. 初始化基础设施（数据库连接池、Redis 客户端）
//   4. 创建数据访问层（repositories）
//   5. 初始化认证和安全组件（JWT、黑名单、权限缓存、限流器）
//   6. 构建服务容器和处理器
//   7. 配置 HTTP 路由
//   8. 启动 HTTP 服务器
//   9. 等待中断信号并优雅关闭
//
// 注意事项：
//   - 数据库或 Redis 连接失败会导致程序退出
//   - 服务器在独立的 goroutine 中运行
//   - 优雅关闭等待最大 10 秒，超时后强制关闭
func main() {
	// 打印版本信息
	printVersionInfo()

	// 加载配置
	cfg := loadConfig()

	// 初始化基础设施
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	logger.Println("Database connection pool initialized")

	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	logger.Println("Redis connection pool initialized")

	cacheService := cache.NewService(redisClient)
	logger.Println("Cache service initialized")

	// 初始化 repositories
	userRepo := database.NewUserRepository(pool)
	permRepo := database.NewPermissionRepository(pool)
	txManager := database.NewTransactionManager(pool)
	logger.Println("Transaction manager initialized")

	roleRepo := database.NewRoleRepositoryWithTxManager(pool, txManager)
	postRepo := database.NewPostRepository(pool)
	categoryRepo := database.NewCategoryRepository(pool)
	tagRepo := database.NewTagRepository(pool)
	seriesRepo := database.NewSeriesRepository(pool)
	commentRepo := database.NewCommentRepository(pool)
	commentLikeRepo := database.NewCommentLikeRepository(pool)
	postLikeRepo := database.NewPostLikeRepository(pool)
	mediaRepo := database.NewMediaRepository(pool)
	searchRepo := database.NewSearchRepository(pool)
	logger.Println("Repositories initialized")

	// 初始化 JWT 服务
	jwtService := auth.NewJWTService(cfg.JWT)
	logger.Println("JWT service initialized")

	// 初始化 token 黑名单和权限缓存
	tokenBlacklist := auth.NewRedisTokenBlacklist(redisClient)
	logger.Println("Token blacklist initialized")

	permCache := auth.NewPermissionCache(cacheService, permRepo, redisClient.Client())
	logger.Println("Permission cache initialized")

	// 初始化限流器
	loginLimiter := middleware.NewRateLimiter(redisClient.Client(), middleware.LoginLimit, middleware.LoginWindow)
	publicLimiter := middleware.NewRateLimiter(redisClient.Client(), middleware.PublicAPILimit, middleware.PublicAPIWindow)
	userLimiter := middleware.NewRateLimiter(redisClient.Client(), middleware.UserActionLimit, middleware.UserActionWindow)
	logger.Println("Rate limiters initialized")

	// 初始化 Service 容器
	serviceContainer := services.NewContainerWithMedia(
		userRepo, roleRepo, jwtService, tokenBlacklist,
		postRepo, categoryRepo, tagRepo, seriesRepo,
		commentRepo, commentLikeRepo,
		mediaRepo, cfg.Upload.Dir, cfg.Upload.BaseURL, postLikeRepo, searchRepo,
	)
	logger.Println("Service container initialized")

	// 初始化 handlers
	authHandler := handlers.NewAuthHandlerWithServices(
		serviceContainer.AuthService,
		serviceContainer.UserService,
		serviceContainer.JWTService(),
		roleRepo,
	)
	postHandler := handlers.NewPostHandler(serviceContainer.PostService)
	categoryHandler := handlers.NewCategoryHandler(serviceContainer.CategoryService)
	tagHandler := handlers.NewTagHandler(serviceContainer.TagService)
	commentHandler := handlers.NewCommentHandler(serviceContainer.CommentService)
	mediaHandler := handlers.NewMediaHandler(serviceContainer.MediaService)
	searchHandler := handlers.NewSearchHandler(serviceContainer.SearchService)

	rssConfig := rss.DefaultFeedConfig()
	rssConfig.BaseURL = cfg.Upload.BaseURL + "/posts"
	rssConfig.Link = cfg.Upload.BaseURL
	rssHandler := handlers.NewRSSHandler(serviceContainer.RSSService, rssConfig)

	adminHandler := handlers.NewAdminHandler(
		userRepo, roleRepo, permRepo,
		postRepo, categoryRepo, tagRepo, commentRepo,
		serviceContainer.UserService,
		jwtService,
		permCache,
	)
	logger.Println("Handlers initialized")

	// 配置路由
	mux := http.NewServeMux()
	routeDeps := &RouteDeps{
		AuthHandler:     authHandler,
		PostHandler:     postHandler,
		CategoryHandler: categoryHandler,
		TagHandler:      tagHandler,
		CommentHandler:  commentHandler,
		MediaHandler:    mediaHandler,
		RSSHandler:      rssHandler,
		SearchHandler:   searchHandler,
		AdminHandler:    adminHandler,
		LoginLimiter:    loginLimiter,
		PublicLimiter:   publicLimiter,
		UserLimiter:     userLimiter,
		JWTService:      jwtService,
		TokenBlacklist:  tokenBlacklist,
		PermCache:       permCache,
		PostRepo:        postRepo,
		CommentRepo:     commentRepo,
		Services:        serviceContainer,
		UploadDir:       cfg.Upload.Dir,
	}
	setupRoutes(mux, routeDeps)

	// 添加首页路由（不在 routes.go 中，因为它需要 templ import）
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Cadmus - 博客平台</title></head><body><h1>Welcome to Cadmus</h1></body></html>`))
	}))

	// 创建 HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 启动服务器
	go func() {
		logger.Printf("Cadmus server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down server...")

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("Server shutdown error: %v", err)
	}
	logger.Println("Server stopped")
}