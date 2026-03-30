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
	"github.com/google/uuid"
	"rua.plus/cadmus/internal/api/handlers"
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/database"
	"rua.plus/cadmus/internal/core/user"
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

	// 初始化 repositories
	userRepo := database.NewUserRepository(pool)
	roleRepo := database.NewRoleRepository(pool)
	_ = database.NewPermissionRepository(pool) // 保留用于后续权限检查
	log.Println("Repositories initialized")

	// 初始化 JWT 服务
	jwtCfg := auth.DefaultJWTConfig()
	jwtCfg.Secret = getEnvOrDefault("JWT_SECRET", jwtCfg.Secret)
	jwtService := auth.NewJWTService(jwtCfg)
	log.Println("JWT service initialized")

	// 创建认证服务适配器（连接 database repository 和 auth service）
	authUserRepo := &authUserRepoAdapter{userRepo: userRepo}
	authService := auth.NewAuthService(jwtService, authUserRepo)
	log.Println("Auth service initialized")

	// 初始化认证处理器
	authHandler := handlers.NewAuthHandlerWithRole(authService, jwtService, userRepo, roleRepo)
	log.Println("Auth handlers initialized")

	// 创建路由
	mux := http.NewServeMux()

	// API 路由组
	// 认证 API
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", handlers.AuthMiddleware(jwtService)(http.HandlerFunc(authHandler.Logout)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("GET /api/v1/auth/me", handlers.AuthMiddleware(jwtService)(http.HandlerFunc(authHandler.Me)).ServeHTTP)

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

// authUserRepoAdapter 适配 database.UserRepository 到 auth.UserRepository
type authUserRepoAdapter struct {
	userRepo *database.UserRepository
}

func (a *authUserRepoAdapter) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	u, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &auth.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		RoleID:       u.RoleID,
		Status:       string(u.Status),
	}, nil
}

func (a *authUserRepoAdapter) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	u, err := a.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &auth.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		RoleID:       u.RoleID,
		Status:       string(u.Status),
	}, nil
}

func (a *authUserRepoAdapter) GetByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	u, err := a.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &auth.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		RoleID:       u.RoleID,
		Status:       string(u.Status),
	}, nil
}

func (a *authUserRepoAdapter) Create(ctx context.Context, u *auth.User) (*auth.User, error) {
	coreUser := &user.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		RoleID:       u.RoleID,
		Status:       user.UserStatus(u.Status),
	}
	if err := a.userRepo.Create(ctx, coreUser); err != nil {
		return nil, err
	}
	return &auth.User{
		ID:           coreUser.ID,
		Username:     coreUser.Username,
		Email:        coreUser.Email,
		PasswordHash: coreUser.PasswordHash,
		RoleID:       coreUser.RoleID,
		Status:       string(coreUser.Status),
	}, nil
}

func (a *authUserRepoAdapter) Update(ctx context.Context, u *auth.User) error {
	coreUser := &user.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		RoleID:       u.RoleID,
		Status:       user.UserStatus(u.Status),
	}
	return a.userRepo.Update(ctx, coreUser)
}