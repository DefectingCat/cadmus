// Package handlers 提供了 Cadmus API 的 HTTP 处理器实现。
//
// 该文件包含认证相关的核心逻辑，包括：
//   - 用户注册和登录
//   - JWT 令牌的生成和刷新
//   - 用户登出和会话管理
//   - 当前用户信息获取
//
// 主要用途：
//
//	用于处理用户认证流程的完整生命周期，从注册到登出。
//
// 注意事项：
//   - 所有认证接口都需要验证请求格式的完整性
//   - 登录接口会检查用户封禁状态
//   - 令牌刷新需要有效的 JWT
//
// 作者：xfy
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/auth"
	"rua.plus/cadmus/internal/core/user"
	"rua.plus/cadmus/internal/services"
)

// AuthHandler 认证 API 处理器。
//
// 该处理器负责处理所有认证相关的 HTTP 请求，包括用户注册、登录、登出和令牌刷新。
// 它依赖 AuthService 处理业务逻辑，UserService 管理用户数据，JWTService 处理令牌操作。
//
// 注意事项：
//   - 该处理器不包含权限检查，权限验证由中间件负责
//   - 登录时会自动生成 JWT 令牌
type AuthHandler struct {
	// authService 认证服务，处理登录登出逻辑
	authService services.AuthService

	// userService 用户服务，处理用户数据操作
	userService services.UserService

	// jwtService JWT 服务，处理令牌生成和验证
	jwtService *auth.JWTService

	// roleRepo 角色仓库，用于获取角色信息（可选）
	roleRepo user.RoleRepository
}

// NewAuthHandler 创建认证处理器。
//
// 创建一个基础的认证处理器，适用于不需要角色信息查询的场景。
//
// 参数：
//   - authService: 认证服务，处理登录登出业务逻辑
//   - userService: 用户服务，处理用户数据操作
//   - jwtService: JWT 服务，处理令牌生成和验证
//
// 返回值：
//   - *AuthHandler: 新创建的认证处理器实例
//
// 使用示例：
//
//	handler := NewAuthHandler(authSvc, userSvc, jwtSvc)
func NewAuthHandler(authService services.AuthService, userService services.UserService, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		jwtService:  jwtService,
	}
}

// NewAuthHandlerWithServices 创建完整功能的认证处理器。
//
// 创建一个包含所有依赖的认证处理器，支持角色信息查询功能。
// 当需要获取用户角色详细信息时，应使用此构造函数。
//
// 参数：
//   - authService: 认证服务，处理登录登出业务逻辑
//   - userService: 用户服务，处理用户数据操作
//   - jwtService: JWT 服务，处理令牌生成和验证
//   - roleRepo: 角色仓库，用于获取角色信息
//
// 返回值：
//   - *AuthHandler: 新创建的完整功能认证处理器实例
//
// 使用示例：
//
//	handler := NewAuthHandlerWithServices(authSvc, userSvc, jwtSvc, roleRepo)
func NewAuthHandlerWithServices(authService services.AuthService, userService services.UserService, jwtService *auth.JWTService, roleRepo user.RoleRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		jwtService:  jwtService,
		roleRepo:    roleRepo,
	}
}

// RegisterRequest 用户注册请求结构体。
//
// 包含用户注册所需的所有必填字段。
type RegisterRequest struct {
	// Username 用户名，必须唯一
	Username string `json:"username"`

	// Email 用户邮箱，必须唯一且格式正确
	Email string `json:"email"`

	// Password 用户密码，需要满足最小长度要求
	Password string `json:"password"`
}

// LoginRequest 用户登录请求结构体。
//
// 包含用户登录所需的凭证信息。
type LoginRequest struct {
	// Email 用户邮箱
	Email string `json:"email"`

	// Password 用户密码
	Password string `json:"password"`
}

// AuthResponse 认证成功响应结构体。
//
// 包含 JWT 令牌和用户基本信息。
type AuthResponse struct {
	// Token JWT 认证令牌，用于后续请求认证
	Token string `json:"token"`

	// User 用户基本信息
	User *UserInfo `json:"user"`
}

// UserInfo 用户信息响应结构体。
//
// 包含用户的公开信息，用于 API 响应。
type UserInfo struct {
	// ID 用户唯一标识符
	ID uuid.UUID `json:"id"`

	// Username 用户名
	Username string `json:"username"`

	// Email 用户邮箱
	Email string `json:"email"`

	// AvatarURL 用户头像 URL（可选）
	AvatarURL string `json:"avatar_url,omitempty"`

	// Bio 用户简介（可选）
	Bio string `json:"bio,omitempty"`

	// RoleID 用户角色 ID
	RoleID uuid.UUID `json:"role_id"`

	// Status 用户状态（active, banned 等）
	Status string `json:"status"`
}

// Register 用户注册。
//
// 处理用户注册请求，创建新用户账户并返回认证令牌。
// 注册成功后自动生成 JWT 令牌，用户无需再次登录。
//
// 路由：POST /api/v1/auth/register
//
// 参数（通过请求体）：
//   - username: 用户名，必填，必须唯一
//   - email: 用户邮箱，必填，必须唯一且格式正确
//   - password: 用户密码，必填
//
// 返回值（通过响应体）：
//   - token: JWT 认证令牌
//   - user: 新创建的用户信息
//
// 可能的错误：
//   - INVALID_REQUEST: 请求格式错误
//   - VALIDATION_ERROR: 必填字段缺失
//   - USER_EXISTS: 用户已存在（邮箱或用户名重复）
//   - INTERNAL_ERROR: 创建用户或生成令牌失败
//
// 使用示例：
//
//	POST /api/v1/auth/register
//	Body: {"username": "test", "email": "test@example.com", "password": "password123"}
//	Response: {"token": "jwt...", "user": {...}}
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
		return
	}

	// 验证必填字段：用户名、邮箱和密码都不能为空
	if req.Username == "" || req.Email == "" || req.Password == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "用户名、邮箱和密码为必填项", nil, http.StatusBadRequest)
		return
	}

	// 调用 Service 层处理注册逻辑，包括密码加密和用户创建
	ctx := r.Context()
	newUser, err := h.userService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		// 用户已存在错误（邮箱或用户名重复）
		if err == user.ErrUserAlreadyExists {
			WriteAPIError(w, "USER_EXISTS", "用户已存在", nil, http.StatusConflict)
			return
		}
		// 其他内部错误
		WriteAPIError(w, "INTERNAL_ERROR", "创建用户失败", nil, http.StatusInternalServerError)
		return
	}

	// 注册成功后自动生成 JWT 令牌，方便用户直接使用
	// 注意：忽略 jti，Handler 层不需要关心令牌唯一标识
	token, _, err := h.jwtService.Generate(newUser.ID, newUser.RoleID)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "生成令牌失败", nil, http.StatusInternalServerError)
		return
	}

	// 返回认证成功响应，包含令牌和用户信息
	WriteJSON(w, AuthResponse{
		Token: token,
		User:  toUserInfo(newUser),
	}, http.StatusCreated)
}

// Login 用户登录。
//
// 处理用户登录请求，验证用户凭证并返回认证令牌。
// 登录成功后返回 JWT 令牌，用于后续请求认证。
//
// 路由：POST /api/v1/auth/login
//
// 参数（通过请求体）：
//   - email: 用户邮箱，必填
//   - password: 用户密码，必填
//
// 返回值（通过响应体）：
//   - token: JWT 认证令牌
//   - user: 用户信息
//
// 可能的错误：
//   - INVALID_REQUEST: 请求格式错误
//   - VALIDATION_ERROR: 必填字段缺失
//   - USER_BANNED: 用户已被封禁
//   - AUTH_FAILED: 凭证无效（邮箱或密码错误）
//
// 使用示例：
//
//	POST /api/v1/auth/login
//	Body: {"email": "test@example.com", "password": "password123"}
//	Response: {"token": "jwt...", "user": {...}}
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
