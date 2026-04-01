// Package handlers 提供了 Cadmus API 的 HTTP 处理器实现。
//
// 该文件包含 HTTP 中间件相关的核心逻辑，包括：
//   - JWT 认证中间件
//   - 权限检查中间件
//   - API 错误响应处理
//   - Context 工具函数
//
// 主要用途：
//
//	提供 HTTP 请求的认证、授权和错误处理功能。
//
// 注意事项：
//   - 中间件链顺序：RateLimit -> Auth -> Permission -> Handler
//   - 所有认证信息通过 context 传递
//
// 作者：xfy
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/user"
)

// ctxKey 类型安全的 context key。
//
// 使用自定义类型避免 context key 冲突。
type ctxKey string

const (
	// ctxUserID 用户 ID 的 context key
	ctxUserID ctxKey = "user_id"

	// ctxUserRole 用户角色 ID 的 context key
	ctxUserRole ctxKey = "user_role_id"
)

// AuthMiddleware JWT 认证中间件。
//
// 验证请求中的 JWT 令牌，并将用户 ID 和角色 ID 注入到 context 中。
// 如果令牌无效或缺失，返回 401 错误。
//
// 参数：
//   - jwtService: JWT 服务，用于令牌验证
//
// 返回值：
//   - 中间件函数
//
// 使用示例：
//   router.Use(AuthMiddleware(jwtService))
func AuthMiddleware(jwtService *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ExtractToken(r)
			if token == "" {
				WriteAPIError(w, "AUTH_FAILED", "未授权访问", nil, http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.Validate(token)
			if err != nil {
				WriteAPIError(w, "AUTH_FAILED", "无效的令牌", nil, http.StatusUnauthorized)
				return
			}

			// 使用类型安全的 context key
			ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxUserRole, claims.RoleID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthMiddlewareWithBlacklist JWT 认证中间件（带黑名单检查）。
//
// 验证请求中的 JWT 令牌，并检查令牌是否已被撤销（在黑名单中）。
// 适用于需要即时令牌撤销的场景，如用户登出。
//
// 参数：
//   - jwtService: JWT 服务
//   - blacklist: 令牌黑名单服务
//
// 返回值：
//   - 中间件函数
func AuthMiddlewareWithBlacklist(jwtService *auth.JWTService, blacklist auth.TokenBlacklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ExtractToken(r)
			if token == "" {
				WriteAPIError(w, "AUTH_FAILED", "未授权访问", nil, http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.Validate(token)
			if err != nil {
				WriteAPIError(w, "AUTH_FAILED", "无效的令牌", nil, http.StatusUnauthorized)
				return
			}

			// 检查黑名单（使用 jti）
			ctx := r.Context()
			jti := claims.GetJWTID()
			if jti != "" && blacklist.IsBlacklisted(ctx, jti) {
				WriteAPIError(w, "TOKEN_REVOKED", "令牌已被撤销", nil, http.StatusUnauthorized)
				return
			}

			// 使用类型安全的 context key
			ctx = context.WithValue(ctx, ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxUserRole, claims.RoleID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID 从 context 获取用户 ID。
//
// 从请求 context 中提取认证用户的 ID。
// 必须在 AuthMiddleware 之后调用。
//
// 参数：
//   - ctx: 请求 context
//
// 返回值：
//   - uuid.UUID: 用户 ID
//   - error: 用户未认证时返回错误
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(ctxUserID).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user not authenticated")
	}
	return id, nil
}

// GetUserRoleID 从 context 获取用户角色 ID。
//
// 从请求 context 中提取用户的角色 ID。
// 必须在 AuthMiddleware 之后调用。
//
// 参数：
//   - ctx: 请求 context
//
// 返回值：
//   - uuid.UUID: 角色ID
//   - error: 角色信息未找到时返回错误
func GetUserRoleID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(ctxUserRole).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user role not found")
	}
	return id, nil
}

// PermissionMiddleware 权限检查中间件（直接查库）
func PermissionMiddleware(jwtService *auth.JWTService, permRepo user.PermissionRepository, requiredPerm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 先验证认证
			token := ExtractToken(r)
			if token == "" {
				WriteAPIError(w, "AUTH_FAILED", "未授权访问", nil, http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.Validate(token)
			if err != nil {
				WriteAPIError(w, "AUTH_FAILED", "无效的令牌", nil, http.StatusUnauthorized)
				return
			}

			// 检查权限
			ctx := r.Context()
			hasPerm, err := permRepo.CheckPermission(ctx, claims.RoleID, requiredPerm)
			if err != nil {
				WriteAPIError(w, "INTERNAL_ERROR", "权限检查失败", nil, http.StatusInternalServerError)
				return
			}

			if !hasPerm {
				WriteAPIError(w, "PERMISSION_DENIED", "权限不足", nil, http.StatusForbidden)
				return
			}

			// 设置 context
			ctx = context.WithValue(ctx, ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxUserRole, claims.RoleID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CachedPermissionMiddleware 权限检查中间件（带缓存）
// 命中缓存直接返回，未命中查库并缓存
func CachedPermissionMiddleware(jwtService *auth.JWTService, permCache *auth.PermissionCache, requiredPerm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 先验证认证
			token := ExtractToken(r)
			if token == "" {
				WriteAPIError(w, "AUTH_FAILED", "未授权访问", nil, http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.Validate(token)
			if err != nil {
				WriteAPIError(w, "AUTH_FAILED", "无效的令牌", nil, http.StatusUnauthorized)
				return
			}

			// 使用缓存检查权限
			ctx := r.Context()
			hasPerm, err := permCache.GetPermission(ctx, claims.RoleID, requiredPerm)
			if err != nil {
				WriteAPIError(w, "INTERNAL_ERROR", "权限检查失败", nil, http.StatusInternalServerError)
				return
			}

			if !hasPerm {
				WriteAPIError(w, "PERMISSION_DENIED", "权限不足", nil, http.StatusForbidden)
				return
			}

			// 设置 context
			ctx = context.WithValue(ctx, ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxUserRole, claims.RoleID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ExtractToken 从请求中提取 token
func ExtractToken(r *http.Request) string {
	// 从 Authorization header 提取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		token, found := strings.CutPrefix(authHeader, "Bearer ")
		if found {
			return token
		}
		return authHeader
	}

	// 从 cookie 提取
	cookie, err := r.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// APIError 统一的 API 错误响应格式
type APIError struct {
	Code      string   `json:"code"`       // "AUTH_FAILED", "VALIDATION_ERROR"
	Message   string   `json:"message"`
	Details   []string `json:"details"`    // 详细错误列表
	RequestID string   `json:"request_id"`
}

// WriteAPIError 写入 API 错误响应
func WriteAPIError(w http.ResponseWriter, code string, message string, details []string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: GetRequestID(),
	})
}

// WriteJSON 写入 JSON 响应
func WriteJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// GetRequestID 获取请求 ID（用于追踪）
func GetRequestID() string {
	// TODO: 实现实际的请求 ID 生成逻辑
	return uuid.New().String()
}

// AdminMiddleware 管理员权限检查中间件
// 需要先通过 AuthMiddleware 或 AuthMiddlewareWithBlacklist 验证用户身份
func AdminMiddleware(permCache *auth.PermissionCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 获取用户角色 ID
			roleID, err := GetUserRoleID(ctx)
			if err != nil {
				WriteAPIError(w, "AUTH_FAILED", "未授权访问", nil, http.StatusUnauthorized)
				return
			}

			// 检查是否有管理员权限（admin:access）
			hasPerm, err := permCache.GetPermission(ctx, roleID, "admin:access")
			if err != nil {
				WriteAPIError(w, "INTERNAL_ERROR", "权限检查失败", nil, http.StatusInternalServerError)
				return
			}

			if !hasPerm {
				WriteAPIError(w, "PERMISSION_DENIED", "需要管理员权限", nil, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermissionMiddleware 通用权限检查中间件
// 检查用户是否拥有指定权限
func RequirePermissionMiddleware(permCache *auth.PermissionCache, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 获取用户角色 ID
			roleID, err := GetUserRoleID(ctx)
			if err != nil {
				WriteAPIError(w, "AUTH_FAILED", "未授权访问", nil, http.StatusUnauthorized)
				return
			}

			// 检查权限
			hasPerm, err := permCache.GetPermission(ctx, roleID, permission)
			if err != nil {
				WriteAPIError(w, "INTERNAL_ERROR", "权限检查失败", nil, http.StatusInternalServerError)
				return
			}

			if !hasPerm {
				WriteAPIError(w, "PERMISSION_DENIED", "权限不足", nil, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}