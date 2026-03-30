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

// ctxKey 类型安全的 context key
type ctxKey string

const (
	ctxUserID   ctxKey = "user_id"
	ctxUserRole ctxKey = "user_role_id"
)

// AuthMiddleware JWT 认证中间件
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

// GetUserID 从 context 获取用户 ID
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(ctxUserID).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user not authenticated")
	}
	return id, nil
}

// GetUserRoleID 从 context 获取用户角色 ID
func GetUserRoleID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(ctxUserRole).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user role not found")
	}
	return id, nil
}

// PermissionMiddleware 权限检查中间件
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

	// 从 query 参数提取
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
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