package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/user"
	"rua.plus/cadmus/internal/services"
)

// AuthHandler 认证 API 处理器
type AuthHandler struct {
	authService services.AuthService
	userService services.UserService
	jwtService  *auth.JWTService
	roleRepo    user.RoleRepository // 保留用于获取角色信息
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService services.AuthService, userService services.UserService, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		jwtService:  jwtService,
	}
}

// NewAuthHandlerWithServices 创建完整功能的认证处理器
func NewAuthHandlerWithServices(authService services.AuthService, userService services.UserService, jwtService *auth.JWTService, roleRepo user.RoleRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		jwtService:  jwtService,
		roleRepo:    roleRepo,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token string    `json:"token"`
	User  *UserInfo `json:"user"`
}

// UserInfo 用户信息响应
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Bio       string    `json:"bio,omitempty"`
	RoleID    uuid.UUID `json:"role_id"`
	Status    string    `json:"status"`
}

// Register 用户注册
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Username == "" || req.Email == "" || req.Password == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "用户名、邮箱和密码为必填项", nil, http.StatusBadRequest)
		return
	}

	// 调用 Service 层处理注册
	ctx := r.Context()
	newUser, err := h.userService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		if err == user.ErrUserAlreadyExists {
			WriteAPIError(w, "USER_EXISTS", "用户已存在", nil, http.StatusConflict)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "创建用户失败", nil, http.StatusInternalServerError)
		return
	}

	// 生成 token（忽略 jti，Handler 不需要）
	token, _, err := h.jwtService.Generate(newUser.ID, newUser.RoleID)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "生成令牌失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, AuthResponse{
		Token: token,
		User:  toUserInfo(newUser),
	}, http.StatusCreated)
}

// Login 用户登录
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "邮箱和密码为必填项", nil, http.StatusBadRequest)
		return
	}

	// 调用 Service 层处理登录
	ctx := r.Context()
	token, u, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err.Error() == "user is banned" {
			WriteAPIError(w, "USER_BANNED", "用户已被封禁", nil, http.StatusForbidden)
			return
		}
		WriteAPIError(w, "AUTH_FAILED", "无效的凭证", nil, http.StatusUnauthorized)
		return
	}

	WriteJSON(w, AuthResponse{
		Token: token,
		User:  toUserInfo(u),
	}, http.StatusOK)
}

// Logout 用户登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := ExtractToken(r)
	if token == "" {
		WriteAPIError(w, "AUTH_FAILED", "未授权访问", nil, http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	if err := h.authService.Logout(ctx, token); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "登出失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "登出成功"}, http.StatusOK)
}

// Refresh 刷新 token
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	token := ExtractToken(r)
	if token == "" {
		WriteAPIError(w, "AUTH_FAILED", "未授权访问", nil, http.StatusUnauthorized)
		return
	}

	newToken, _, err := h.jwtService.Refresh(token)
	if err != nil {
		WriteAPIError(w, "AUTH_FAILED", "令牌刷新失败", nil, http.StatusUnauthorized)
		return
	}

	WriteJSON(w, map[string]string{"token": newToken}, http.StatusOK)
}

// Me 获取当前用户信息
// GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r.Context())
	if err != nil {
		WriteAPIError(w, "AUTH_FAILED", "未授权访问", nil, http.StatusUnauthorized)
		return
	}

	u, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		WriteAPIError(w, "USER_NOT_FOUND", "用户不存在", nil, http.StatusNotFound)
		return
	}

	WriteJSON(w, toUserInfo(u), http.StatusOK)
}

// toUserInfo 转换 User 到 UserInfo
func toUserInfo(u *user.User) *UserInfo {
	return &UserInfo{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
		RoleID:    u.RoleID,
		Status:    string(u.Status),
	}
}