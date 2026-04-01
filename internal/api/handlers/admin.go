// Package handlers 提供了 Cadmus API 的 HTTP 处理器实现。
//
// 该文件包含后台管理相关的核心逻辑，包括：
//   - 角色管理（CRUD）
//   - 用户管理（列表、封禁）
//   - 批量操作（文章、评论）
//   - 排序管理
//
// 主要用途：
//
//	用于实现管理后台的各种管理功能。
//
// 注意事项：
//   - 所有接口都需要管理员权限
//   - 批量操作有事务保护
//   - 删除操作有保护机制（如不能删除默认角色）
//
// 作者：xfy
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/auth"
	comment "rua.plus/cadmus/internal/core/comment"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/core/user"
	"rua.plus/cadmus/internal/database"
	"rua.plus/cadmus/internal/services"
)

// AdminHandler 后台管理 API 处理器。
//
// 该处理器负责处理所有后台管理相关的 HTTP 请求。
// 所有方法都需要管理员权限验证。
type AdminHandler struct {
	// userRepo 用户仓库
	userRepo user.UserRepository

	// roleRepo 角色仓库
	roleRepo user.RoleRepository

	// permRepo 权限仓库
	permRepo user.PermissionRepository

	// postRepo 文章仓库
	postRepo post.PostRepository

	// categoryRepo 分类仓库
	categoryRepo post.CategoryRepository

	// tagRepo 标签仓库
	tagRepo post.TagRepository

	// commentRepo 评论仓库
	commentRepo comment.CommentRepository

	// userService 用户服务
	userService services.UserService

	// jwtService JWT 服务
	jwtService *auth.JWTService

	// permCache 权限缓存
	permCache *auth.PermissionCache
}

// NewAdminHandler 创建后台管理处理器。
//
// 创建一个完整的后台管理处理器，包含所有必要的管理功能。
//
// 参数：
//   - userRepo: 用户仓库
//   - roleRepo: 角色仓库
//   - permRepo: 权限仓库
//   - postRepo: 文章仓库
//   - categoryRepo: 分类仓库
//   - tagRepo: 标签仓库
//   - commentRepo: 评论仓库
//   - userService: 用户服务
//   - jwtService: JWT 服务
//   - permCache: 权限缓存
//
// 返回值：
//   - *AdminHandler: 新创建的后台管理处理器实例
func NewAdminHandler(
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	permRepo user.PermissionRepository,
	postRepo post.PostRepository,
	categoryRepo post.CategoryRepository,
	tagRepo post.TagRepository,
	commentRepo comment.CommentRepository,
	userService services.UserService,
	jwtService *auth.JWTService,
	permCache *auth.PermissionCache,
) *AdminHandler {
	return &AdminHandler{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		permRepo:     permRepo,
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		commentRepo:  commentRepo,
		userService:  userService,
		jwtService:   jwtService,
		permCache:    permCache,
	}
}

// === 角色管理 ===

// RoleListResponse 角色列表响应结构体。
type RoleListResponse struct {
	// Roles 角色列表
	Roles []RoleDetail `json:"roles"`

	// Total 角色总数
	Total int `json:"total"`
}

// RoleDetail 角色详情结构体。
type RoleDetail struct {
	// ID 角色唯一标识符
	ID uuid.UUID `json:"id"`

	// Name 角色名称（系统内部使用）
	Name string `json:"name"`

	// DisplayName 角色显示名称
	DisplayName string `json:"display_name"`

	// IsDefault 是否为默认角色
	IsDefault bool `json:"is_default"`

	// Permissions 角色拥有的权限列表
	Permissions []PermissionInfo `json:"permissions,omitempty"`

	// CreatedAt 创建时间
	CreatedAt string `json:"created_at"`
}

// PermissionInfo 权限信息结构体。
type PermissionInfo struct {
	// ID 权限唯一标识符
	ID uuid.UUID `json:"id"`

	// Name 权限名称
	Name string `json:"name"`

	// Description 权限描述
	Description string `json:"description"`

	// Category 权限分类
	Category string `json:"category"`
}

// CreateRoleRequest 创建角色请求结构体。
type CreateRoleRequest struct {
	// Name 角色名称（必填）
	Name string `json:"name"`

	// DisplayName 角色显示名称（必填）
	DisplayName string `json:"display_name"`

	// IsDefault 是否为默认角色
	IsDefault bool `json:"is_default,omitempty"`

	// Permissions 角色权限 ID 列表
	Permissions []uuid.UUID `json:"permissions,omitempty"`
}

// UpdateRoleRequest 更新角色请求结构体。
type UpdateRoleRequest struct {
	// DisplayName 角色显示名称
	DisplayName string `json:"display_name"`

	// Permissions 角色权限 ID 列表
	Permissions []uuid.UUID `json:"permissions"`
}

// ListRoles 角色列表。
//
// 获取系统中所有角色的列表，包含每个角色的权限信息。
// 需要管理员权限。
//
// 路由：GET /api/v1/admin/roles
//
// 返回值（通过响应体）：
//   - roles: 角色列表
//   - total: 角色总数
func (h *AdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roles, err := h.roleRepo.GetAll(ctx)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取角色列表失败", nil, http.StatusInternalServerError)
		return
	}

	details := make([]RoleDetail, 0, len(roles))
	for _, role := range roles {
		perms, _ := h.permRepo.GetByRoleID(ctx, role.ID)
		details = append(details, toRoleDetail(role, perms))
	}

	WriteJSON(w, RoleListResponse{
		Roles: details,
		Total: len(details),
	}, http.StatusOK)
}

// CreateRole 创建角色。
//
// 创建新的角色并设置其权限。需要管理员权限。
//
// 路由：POST /api/v1/admin/roles
//
// 参数（通过请求体）：
//   - name: 角色名称（必填）
//   - display_name: 显示名称（必填）
//   - is_default: 是否为默认角色
//   - permissions: 权限 ID 列表
//
// 返回值（通过响应体）：
//   - 新创建的角色信息
func (h *AdminHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.DisplayName == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "角色名称和显示名称为必填项", nil, http.StatusBadRequest)
		return
	}

	// 通过 database 包的角色仓库创建角色（需要扩展）
	dbRoleRepo, ok := h.roleRepo.(*database.RoleRepository)
	if !ok {
		WriteAPIError(w, "INTERNAL_ERROR", "角色仓库类型不支持创建操作", nil, http.StatusInternalServerError)
		return
	}

	roleID, err := dbRoleRepo.Create(ctx, req.Name, req.DisplayName, req.IsDefault)
	if err != nil {
		WriteAPIError(w, "CREATE_FAILED", "创建角色失败", []string{err.Error()}, http.StatusInternalServerError)
		return
	}

	// 设置权限
	if len(req.Permissions) > 0 {
		if err := dbRoleRepo.SetPermissions(ctx, roleID, req.Permissions); err != nil {
			WriteAPIError(w, "PERMISSION_SET_FAILED", "设置权限失败", []string{err.Error()}, http.StatusInternalServerError)
			return
		}
	}

	// 获取完整角色信息
	role, err := h.roleRepo.GetWithPermissions(ctx, roleID)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取角色信息失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toRoleDetail(role, role.Permissions), http.StatusCreated)
}

// UpdateRole 更新角色权限。
//
// 更新指定角色的显示名称和权限。需要管理员权限。
// 更新后会自动清除该角色的权限缓存。
//
// 路由：PUT /api/v1/admin/roles/{id}
//
// 参数：
//   - id: 角色 ID（路径参数）
//   - display_name: 新的显示名称
//   - permissions: 新的权限 ID 列表
//
// 返回值（通过响应体）：
//   - 更新后的角色信息
func (h *AdminHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "缺少角色 ID", nil, http.StatusBadRequest)
		return
	}

	roleID, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "VALIDATION_ERROR", "无效的角色 ID", nil, http.StatusBadRequest)
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
		return
	}

	// 检查角色是否存在
	_, err = h.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		WriteAPIError(w, "ROLE_NOT_FOUND", "角色不存在", nil, http.StatusNotFound)
		return
	}

	// 更新显示名称
	if req.DisplayName != "" {
		dbRoleRepo, ok := h.roleRepo.(*database.RoleRepository)
		if ok {
			if err := dbRoleRepo.UpdateDisplayName(ctx, roleID, req.DisplayName); err != nil {
				WriteAPIError(w, "UPDATE_FAILED", "更新角色失败", []string{err.Error()}, http.StatusInternalServerError)
				return
			}
		}
	}

	// 更新权限
	if len(req.Permissions) >= 0 { // 允许清空权限
		dbRoleRepo, ok := h.roleRepo.(*database.RoleRepository)
		if ok {
			if err := dbRoleRepo.SetPermissions(ctx, roleID, req.Permissions); err != nil {
				WriteAPIError(w, "PERMISSION_SET_FAILED", "设置权限失败", []string{err.Error()}, http.StatusInternalServerError)
				return
			}
		}
	}

	// 清除权限缓存
	if h.permCache != nil {
		h.permCache.InvalidateRolePermissions(ctx, roleID)
	}

	// 获取更新后的角色
	updatedRole, err := h.roleRepo.GetWithPermissions(ctx, roleID)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取角色信息失败", nil, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, toRoleDetail(updatedRole, updatedRole.Permissions), http.StatusOK)
}

// DeleteRole 删除角色。
//
// 删除指定的角色。需要管理员权限。
// 不能删除默认角色或有用户关联的角色。
//
// 路由：DELETE /api/v1/admin/roles/{id}
//
// 参数：
//   - id: 角色 ID（路径参数）
//
// 可能的错误：
//   - ROLE_NOT_FOUND: 角色不存在
//   - DELETE_FAILED: 不能删除默认角色或角色下存在用户
func (h *AdminHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "缺少角色 ID", nil, http.StatusBadRequest)
		return
	}

	roleID, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "VALIDATION_ERROR", "无效的角色 ID", nil, http.StatusBadRequest)
		return
	}

	// 检查角色是否存在
	role, err := h.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		WriteAPIError(w, "ROLE_NOT_FOUND", "角色不存在", nil, http.StatusNotFound)
		return
	}

	// 不能删除默认角色
	if role.IsDefault {
		WriteAPIError(w, "DELETE_FAILED", "不能删除默认角色", nil, http.StatusBadRequest)
		return
	}

	// 检查是否有用户使用该角色
	dbRoleRepo, ok := h.roleRepo.(*database.RoleRepository)
	if ok {
		count, err := dbRoleRepo.GetUserCount(ctx, roleID)
		if err != nil {
			WriteAPIError(w, "INTERNAL_ERROR", "检查用户数量失败", nil, http.StatusInternalServerError)
			return
		}
		if count > 0 {
			WriteAPIError(w, "DELETE_FAILED", "该角色下存在用户，无法删除", nil, http.StatusBadRequest)
			return
		}

		if err := dbRoleRepo.Delete(ctx, roleID); err != nil {
			WriteAPIError(w, "DELETE_FAILED", "删除角色失败", []string{err.Error()}, http.StatusInternalServerError)
			return
		}
	}

	WriteJSON(w, map[string]string{"message": "角色已删除"}, http.StatusOK)
}

// === 用户管理 ===

// AdminUserListResponse 用户管理列表响应结构体。
type AdminUserListResponse struct {
	// Users 用户列表
	Users []AdminUserInfo `json:"users"`

	// Total 用户总数
	Total int `json:"total"`

	// Page 当前页码
	Page int `json:"page"`
}

// AdminUserInfo 用户管理信息结构体。
type AdminUserInfo struct {
	// ID 用户唯一标识符
	ID uuid.UUID `json:"id"`

	// Username 用户名
	Username string `json:"username"`

	// Email 用户邮箱
	Email string `json:"email"`

	// AvatarURL 用户头像 URL
	AvatarURL string `json:"avatar_url,omitempty"`

	// Bio 用户简介
	Bio string `json:"bio,omitempty"`

	// RoleID 角色ID
	RoleID uuid.UUID `json:"role_id"`

	// RoleName 角色名称
	RoleName string `json:"role_name,omitempty"`

	// Status 用户状态
	Status string `json:"status"`

	// PostCount 文章数量
	PostCount int `json:"post_count,omitempty"`

	// CreatedAt 创建时间
	CreatedAt string `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt string `json:"updated_at"`
}

// BanUserRequest 封禁用户请求结构体。
type BanUserRequest struct {
	// Reason 封禁原因（可选）
	Reason string `json:"reason,omitempty"`
}

// ListUsers 用户管理列表。
//
// 获取系统中所有用户的列表，包含角色信息和文章数量。
// 需要管理员权限。使用批量查询避免 N+1 问题。
//
// 路由：GET /api/v1/admin/users
//
// 查询参数：
//   - page: 页码，默认 1
//   - limit: 每页数量，默认 20，最大 100
//
// 返回值（通过响应体）：
//   - users: 用户列表
//   - total: 用户总数
//   - page: 当前页码
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 分页参数
	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := parseIntStr(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := parseIntStr(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := (page - 1) * limit

	users, total, err := h.userService.List(ctx, offset, limit)
	if err != nil {
		WriteAPIError(w, "INTERNAL_ERROR", "获取用户列表失败", nil, http.StatusInternalServerError)
		return
	}

	// 收集所有角色 ID 和用户 ID 用于批量查询
	roleIDs := make([]uuid.UUID, 0, len(users))
	userIDs := make([]uuid.UUID, 0, len(users))
	for _, u := range users {
		roleIDs = append(roleIDs, u.RoleID)
		userIDs = append(userIDs, u.ID)
	}

	// 批量查询角色（避免 N+1）
	rolesMap := make(map[uuid.UUID]*user.Role)
	dbRoleRepo, ok := h.roleRepo.(*database.RoleRepository)
	if ok && len(roleIDs) > 0 {
		rolesMap, _ = dbRoleRepo.GetByIDs(ctx, roleIDs)
	}

	// 批量查询文章数（避免 N+1）
	postCountsMap := make(map[uuid.UUID]int)
	dbPostRepo, ok := h.postRepo.(*database.PostRepository)
	if ok && len(userIDs) > 0 {
		postCountsMap, _ = dbPostRepo.CountByAuthors(ctx, userIDs)
	}

	userInfos := make([]AdminUserInfo, 0, len(users))
	for _, u := range users {
		role := rolesMap[u.RoleID]
		postCount := postCountsMap[u.ID]
		userInfos = append(userInfos, toAdminUserInfo(u, role, postCount))
	}

	WriteJSON(w, AdminUserListResponse{
		Users: userInfos,
		Total: total,
		Page:  page,
	}, http.StatusOK)
}

// BanUser 封禁/解封用户。
//
// 切换用户的封禁状态。需要管理员权限。
// 不能封禁自己。
//
// 路由：PUT /api/v1/admin/users/{id}/ban
//
// 参数：
//   - id: 用户 ID（路径参数）
//   - reason: 封禁原因（可选，请求体）
//
// 返回值（通过响应体）：
//   - 更新后的用户信息
func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		WriteAPIError(w, "VALIDATION_ERROR", "缺少用户 ID", nil, http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		WriteAPIError(w, "VALIDATION_ERROR", "无效的用户 ID", nil, http.StatusBadRequest)
		return
	}

	// 获取用户
	u, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		WriteAPIError(w, "USER_NOT_FOUND", "用户不存在", nil, http.StatusNotFound)
		return
	}

	// 获取当前用户 ID（检查不能封禁自己）
	currentUserID, err := GetUserID(ctx)
	if err == nil && currentUserID == userID {
		WriteAPIError(w, "BAN_FAILED", "不能封禁自己", nil, http.StatusBadRequest)
		return
	}

	var req BanUserRequest
	json.NewDecoder(r.Body).Decode(&req)

	// 切换封禁状态
	if u.Status == user.StatusBanned {
		u.Status = user.StatusActive
	} else {
		u.Status = user.StatusBanned
	}

	if err := h.userRepo.Update(ctx, u); err != nil {
		WriteAPIError(w, "UPDATE_FAILED", "更新用户状态失败", []string{err.Error()}, http.StatusInternalServerError)
		return
	}

	role, _ := h.roleRepo.GetByID(ctx, u.RoleID)
	postCount, _ := h.getPostCount(ctx, u.ID)

	WriteJSON(w, toAdminUserInfo(u, role, postCount), http.StatusOK)
}

// === 批量操作 ===

// BatchRequest 批量操作请求结构体。
type BatchRequest struct {
	// Action 操作类型：delete_posts, delete_comments, move_category, change_status
	Action string `json:"action"`

	// IDs 目标 ID 列表
	IDs []uuid.UUID `json:"ids"`

	// Params 操作参数（可选）
	Params map[string]any `json:"params,omitempty"`
}

// BatchResponse 批量操作响应结构体。
type BatchResponse struct {
	// Success 成功数量
	Success int `json:"success"`

	// Failed 失败数量
	Failed int `json:"failed"`

	// Errors 错误信息列表
	Errors []string `json:"errors,omitempty"`

	// Message 操作结果消息
	Message string `json:"message"`
}

// BatchOperation 批量操作。
//
// 执行批量操作，支持删除文章/评论、移动分类、更改状态等。
// 需要管理员权限。
//
// 路由：POST /api/v1/admin/batch
//
// 参数（通过请求体）：
//   - action: 操作类型
//   - ids: 目标 ID 列表
//   - params: 操作参数
//
// 返回值（通过响应体）：
//   - success: 成功数量
//   - failed: 失败数量
//   - errors: 错误信息
func (h *AdminHandler) BatchOperation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
		return
	}

	if req.Action == "" || len(req.IDs) == 0 {
		WriteAPIError(w, "VALIDATION_ERROR", "操作类型和 ID 列表为必填项", nil, http.StatusBadRequest)
		return
	}

	var success, failed int
	var errors []string

	switch req.Action {
	case "delete_posts":
		success, failed, errors = h.batchDeletePosts(ctx, req.IDs)
	case "delete_comments":
		success, failed, errors = h.batchDeleteComments(ctx, req.IDs)
	case "move_category":
		categoryID, ok := getUUIDFromParams(req.Params, "category_id")
		if !ok {
			WriteAPIError(w, "VALIDATION_ERROR", "缺少目标分类 ID", nil, http.StatusBadRequest)
			return
		}
		success, failed, errors = h.batchMoveCategory(ctx, req.IDs, categoryID)
	case "change_status":
		status, ok := getStringFromParams(req.Params, "status")
		if !ok {
			WriteAPIError(w, "VALIDATION_ERROR", "缺少目标状态", nil, http.StatusBadRequest)
			return
		}
		success, failed, errors = h.batchChangePostStatus(ctx, req.IDs, status)
	default:
		WriteAPIError(w, "INVALID_ACTION", "不支持的操作类型", nil, http.StatusBadRequest)
		return
	}

	WriteJSON(w, BatchResponse{
		Success: success,
		Failed:  failed,
		Errors:  errors,
		Message: "批量操作完成",
	}, http.StatusOK)
}

// === 排序更新 ===

// OrderRequest 排序更新请求结构体。
type OrderRequest struct {
	// Type 排序类型：category, tag
	Type string `json:"type"`

	// Order 新的排序顺序（ID 列表）
	Order []uuid.UUID `json:"order"`
}

// UpdateOrder 更新排序。
//
// 更新分类或标签的排序顺序。需要管理员权限。
//
// 路由：PUT /api/v1/admin/order
//
// 参数（通过请求体）：
//   - type: 排序类型（category 或 tag）
//   - order: 新的排序顺序（ID 列表）
//
// 返回值（通过响应体）：
//   - message: 更新成功提示
func (h *AdminHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteAPIError(w, "INVALID_REQUEST", "无效的请求格式", nil, http.StatusBadRequest)
		return
	}

	if req.Type == "" || len(req.Order) == 0 {
		WriteAPIError(w, "VALIDATION_ERROR", "类型和排序列表为必填项", nil, http.StatusBadRequest)
		return
	}

	switch req.Type {
	case "category":
		if err := h.updateCategoryOrder(ctx, req.Order); err != nil {
			WriteAPIError(w, "UPDATE_FAILED", "更新分类排序失败", []string{err.Error()}, http.StatusInternalServerError)
			return
		}
	case "tag":
		if err := h.updateTagOrder(ctx, req.Order); err != nil {
			WriteAPIError(w, "UPDATE_FAILED", "更新标签排序失败", []string{err.Error()}, http.StatusInternalServerError)
			return
		}
	default:
		WriteAPIError(w, "INVALID_TYPE", "不支持排序类型", nil, http.StatusBadRequest)
		return
	}

	WriteJSON(w, map[string]string{"message": "排序已更新"}, http.StatusOK)
}

// === 辅助方法 ===

// toRoleDetail 转换 Role 实体到 RoleDetail 响应结构。
func toRoleDetail(role *user.Role, perms []user.Permission) RoleDetail {
	permInfos := make([]PermissionInfo, 0, len(perms))
	for _, p := range perms {
		permInfos = append(permInfos, PermissionInfo{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
		})
	}
	return RoleDetail{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		IsDefault:   role.IsDefault,
		Permissions: permInfos,
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// toAdminUserInfo 转换 User 实体到 AdminUserInfo 响应结构。
func toAdminUserInfo(u *user.User, role *user.Role, postCount int) AdminUserInfo {
	roleName := ""
	if role != nil {
		roleName = role.DisplayName
	}
	return AdminUserInfo{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
		RoleID:    u.RoleID,
		RoleName:  roleName,
		Status:    string(u.Status),
		PostCount: postCount,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// getPostCount 获取用户的文章数量。
func (h *AdminHandler) getPostCount(ctx context.Context, userID uuid.UUID) (int, error) {
	// 使用 postRepo 获取用户文章数
	dbPostRepo, ok := h.postRepo.(*database.PostRepository)
	if ok {
		return dbPostRepo.CountByAuthor(ctx, userID)
	}
	return 0, nil
}

// batchDeletePosts 批量删除文章。
func (h *AdminHandler) batchDeletePosts(ctx context.Context, ids []uuid.UUID) (success, failed int, errors []string) {
	for _, id := range ids {
		if err := h.postRepo.Delete(ctx, id); err != nil {
			failed++
			errors = append(errors, err.Error())
		} else {
			success++
		}
	}
	return success, failed, errors
}

// batchDeleteComments 批量删除评论。
func (h *AdminHandler) batchDeleteComments(ctx context.Context, ids []uuid.UUID) (success, failed int, errors []string) {
	for _, id := range ids {
		if err := h.commentRepo.Delete(ctx, id); err != nil {
			failed++
			errors = append(errors, err.Error())
		} else {
			success++
		}
	}
	return success, failed, errors
}

// batchMoveCategory 批量移动文章到指定分类。
func (h *AdminHandler) batchMoveCategory(ctx context.Context, ids []uuid.UUID, categoryID uuid.UUID) (success, failed int, errors []string) {
	dbPostRepo, ok := h.postRepo.(*database.PostRepository)
	if !ok {
		return 0, len(ids), []string{"仓库类型不支持批量移动"}
	}

	for _, id := range ids {
		if err := dbPostRepo.MoveCategory(ctx, id, categoryID); err != nil {
			failed++
			errors = append(errors, err.Error())
		} else {
			success++
		}
	}
	return success, failed, errors
}

// batchChangePostStatus 批量更改文章状态。
func (h *AdminHandler) batchChangePostStatus(ctx context.Context, ids []uuid.UUID, status string) (success, failed int, errors []string) {
	dbPostRepo, ok := h.postRepo.(*database.PostRepository)
	if !ok {
		return 0, len(ids), []string{"仓库类型不支持批量状态变更"}
	}

	for _, id := range ids {
		if err := dbPostRepo.ChangeStatus(ctx, id, status); err != nil {
			failed++
			errors = append(errors, err.Error())
		} else {
			success++
		}
	}
	return success, failed, errors
}

// updateCategoryOrder 更新分类排序。
func (h *AdminHandler) updateCategoryOrder(ctx context.Context, order []uuid.UUID) error {
	dbCategoryRepo, ok := h.categoryRepo.(*database.CategoryRepository)
	if !ok {
		return nil
	}
	return dbCategoryRepo.UpdateOrder(ctx, order)
}

// updateTagOrder 更新标签排序。
func (h *AdminHandler) updateTagOrder(ctx context.Context, order []uuid.UUID) error {
	dbTagRepo, ok := h.tagRepo.(*database.TagRepository)
	if !ok {
		return nil
	}
	return dbTagRepo.UpdateOrder(ctx, order)
}

// parseIntStr 解析整数字符串。
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

// getUUIDFromParams 从参数映射中获取 UUID 值。
func getUUIDFromParams(params map[string]any, key string) (uuid.UUID, bool) {
	if params == nil {
		return uuid.Nil, false
	}
	val, ok := params[key]
	if !ok {
		return uuid.Nil, false
	}

	switch v := val.(type) {
	case string:
		id, err := uuid.Parse(v)
		return id, err == nil
	case uuid.UUID:
		return v, true
	default:
		return uuid.Nil, false
	}
}

// getStringFromParams 从参数映射中获取字符串值。
func getStringFromParams(params map[string]any, key string) (string, bool) {
	if params == nil {
		return "", false
	}
	val, ok := params[key]
	if !ok {
		return "", false
	}

	str, ok := val.(string)
	return str, ok
}