package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/user"
)

// AuthHandler 认证 API 处理器
type AuthHandler struct {
	authService *auth.AuthService
	jwtService  *auth.JWTService
	userRepo    user.UserRepository
	roleRepo    user.RoleRepository
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *auth.AuthService, jwtService *auth.JWTService, userRepo user.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtService:  jwtService,
		userRepo:    userRepo,
	}
}

// NewAuthHandlerWithRole 创建带角色仓库的认证处理器
func NewAuthHandlerWithRole(authService *auth.AuthService, jwtService *auth.JWTService, userRepo user.UserRepository, roleRepo user.RoleRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtService:  jwtService,
		userRepo:    userRepo,
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

	// 获取默认角色
	ctx := r.Context()
	defaultRole, err := h.getDefaultRole(ctx)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取默认角色失败", nil, http.StatusInternalServerError)
		return
	}

	// 创建用户（密码哈希在 service 中处理）
	newUser := &user.User{
		ID:       uuid.New(),
		Username: req.Username,
		Email:    req.Email,
		RoleID:   defaultRole.ID,
		Status:   user.StatusPending,
	}

	// 哈希密码
	if err := newUser.SetPassword(req.Password); err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "密码处理失败", nil, http.StatusInternalServerError)
		return
	}

	if err := h.userRepo.Create(ctx, newUser); err != nil {
		if err == user.ErrUserAlreadyExists {
			WriteAPIError(w, "USER_EXISTS", "用户已存在", nil, http.StatusConflict)
			return
		}
		WriteAPIError(w, "INTERNAL_ERROR", "创建用户失败", nil, http.StatusInternalServerError)
		return
	}

	// 生成 token
	token, err := h.jwtService.Generate(newUser.ID, newUser.RoleID)
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

	ctx := r.Context()
	u, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		WriteAPIError(w, "AUTH_FAILED", "无效的凭证", nil, http.StatusUnauthorized)
		return
	}

	// 验证密码
	if !u.CheckPassword(req.Password) {
		WriteAPIError(w, "AUTH_FAILED", "无效的凭证", nil, http.StatusUnauthorized)
		return
	}

	// 检查用户状态
	if u.Status == user.StatusBanned {
		WriteAPIError(w, "USER_BANNED", "用户已被封禁", nil, http.StatusForbidden)
		return
	}

	// 生成 token
	token, err := h.jwtService.Generate(u.ID, u.RoleID)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "生成令牌失败", nil, http.StatusInternalServerError)
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

	newToken, err := h.jwtService.Refresh(token)
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

	u, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		WriteAPIError(w, "USER_NOT_FOUND", "用户不存在", nil, http.StatusNotFound)
		return
	}

	WriteJSON(w, toUserInfo(u), http.StatusOK)
}

// getDefaultRole 获取默认角色
func (h *AuthHandler) getDefaultRole(_ context.Context) (*user.Role, error) {
	// TODO: 需要注入 RoleRepository
	// 暂时返回硬编码的普通用户角色 ID
	// 实际应该从 role_repo.GetDefault(ctx) 获取
	return &user.Role{
		ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:        "user",
		DisplayName: "普通用户",
		IsDefault:   true,
	}, nil
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