package main

import (
	"context"
	"fmt"
	"log"
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
	"rua.plus/cadmus/internal/services"
	_ "rua.plus/cadmus/plugins/mermaid-block" // 启用 Mermaid 图表块插件（blank import 触发 init）
	_ "rua.plus/cadmus/themes/default"        // 启用默认主题（blank import 触发 init()）
)

// 版本信息（通过 -ldflags 在编译时注入）
var (
	version       = "dev"
	gitCommit     = "unknown"
	gitBranch     = "unknown"
	buildTime     = "unknown"
	goVersion     = "unknown"
	buildPlatform = "unknown"
)

// printVersionInfo 打印版本信息
func printVersionInfo() {
	fmt.Println("Cadmus Blog Platform")
	fmt.Printf("  Version:    %s\n", version)
	fmt.Printf("  Git Commit: %s\n", gitCommit)
	fmt.Printf("  Git Branch: %s\n", gitBranch)
	fmt.Printf("  Build Time: %s\n", buildTime)
	fmt.Printf("  Go Version: %s\n", goVersion)
	fmt.Printf("  Platform:   %s\n", buildPlatform)
}

func main() {
	// 打印版本信息
	printVersionInfo()

	// 加载配置
	cfg := loadConfig()

	// 初始化基础设施
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Database connection pool initialized")

	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connection pool initialized")

	cacheService := cache.NewService(redisClient)
	log.Println("Cache service initialized")

	// 初始化 repositories
	userRepo := database.NewUserRepository(pool)
	permRepo := database.NewPermissionRepository(pool)
	txManager := database.NewTransactionManager(pool)
	log.Println("Transaction manager initialized")

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
	log.Println("Repositories initialized")

	// 初始化 JWT 服务
	jwtService := auth.NewJWTService(cfg.JWT)
	log.Println("JWT service initialized")

	// 初始化 token 黑名单和权限缓存
	tokenBlacklist := auth.NewRedisTokenBlacklist(redisClient)
	log.Println("Token blacklist initialized")

	permCache := auth.NewPermissionCache(cacheService, permRepo, redisClient.Client())
	log.Println("Permission cache initialized")

	// 初始化限流器
	loginLimiter := middleware.NewRateLimiter(redisClient.Client(), middleware.LoginLimit, middleware.LoginWindow)
	publicLimiter := middleware.NewRateLimiter(redisClient.Client(), middleware.PublicAPILimit, middleware.PublicAPIWindow)
	userLimiter := middleware.NewRateLimiter(redisClient.Client(), middleware.UserActionLimit, middleware.UserActionWindow)
	log.Println("Rate limiters initialized")

	// 初始化 Service 容器
	serviceContainer := services.NewContainerWithMedia(
		userRepo, roleRepo, jwtService, tokenBlacklist,
		postRepo, categoryRepo, tagRepo, seriesRepo,
		commentRepo, commentLikeRepo,
		mediaRepo, cfg.Upload.Dir, cfg.Upload.BaseURL, postLikeRepo, searchRepo,
	)
	log.Println("Service container initialized")

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
	log.Println("Handlers initialized")

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