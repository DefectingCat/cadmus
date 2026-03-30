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
	"rua.plus/cadmus/internal/database"
	"rua.plus/cadmus/internal/services"
	"rua.plus/cadmus/web/templates/pages"
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

	// 初始化缓存服务（保留用于后续任务）
	_ = cache.NewService(redisClient)
	log.Println("Cache service initialized")

	// 初始化 repositories
	userRepo := database.NewUserRepository(pool)
	roleRepo := database.NewRoleRepository(pool)
	_ = database.NewPermissionRepository(pool) // 保留用于后续权限检查
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

	// 初始化 Service 容器（带黑名单）
	serviceContainer := services.NewContainerWithBlacklist(userRepo, roleRepo, jwtService, tokenBlacklist)
	log.Println("Service container initialized")

	// 初始化认证处理器（通过 Service 层）
	authHandler := handlers.NewAuthHandlerWithServices(
		serviceContainer.AuthService,
		serviceContainer.UserService,
		serviceContainer.JWTService(),
		roleRepo,
	)
	log.Println("Auth handlers initialized")

	// 创建路由
	mux := http.NewServeMux()

	// API 路由组
	// 认证 API
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(authHandler.Logout)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("GET /api/v1/auth/me", handlers.AuthMiddlewareWithBlacklist(jwtService, tokenBlacklist)(http.HandlerFunc(authHandler.Me)).ServeHTTP)

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 首页
	mux.Handle("/", templ.Handler(pages.HomePage("Cadmus - 博客平台")))

	// 静态文件服务
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

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

