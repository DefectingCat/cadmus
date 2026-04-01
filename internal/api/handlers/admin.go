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

// AdminHandler 后台管理 API 处理器
type AdminHandler struct {
	userRepo     user.UserRepository
	roleRepo     user.RoleRepository
	permRepo     user.PermissionRepository
	postRepo     post.PostRepository
	categoryRepo post.CategoryRepository
	tagRepo      post.TagRepository
	commentRepo  comment.CommentRepository
	userService  services.UserService
	jwtService   *auth.JWTService
	permCache    *auth.PermissionCache
}

// NewAdminHandler 创建后台管理处理器
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
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		permRepo:    permRepo,
		postRepo:    postRepo,
		categoryRepo: categoryRepo,
		tagRepo:     tagRepo,
		commentRepo: commentRepo,
		userService: userService,
		jwtService:  jwtService,
		permCache:   permCache,
	}
}

// === 角色管理 ===

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Roles []RoleDetail `json:"roles"`
	Total int          `json:"total"`
}

// RoleDetail 角色详情
type RoleDetail struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	IsDefault   bool            `json:"is_default"`
	Permissions []PermissionInfo `json:"permissions,omitempty"`
	CreatedAt   string          `json:"created_at"`
}

// PermissionInfo 权限信息
type PermissionInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	IsDefault   bool            `json:"is_default,omitempty"`
	Permissions []uuid.UUID     `json:"permissions,omitempty"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	DisplayName string          `json:"display_name"`
	Permissions []uuid.UUID     `json:"permissions"`
}

// ListRoles 角色列表
// GET /api/v1/admin/roles
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

// CreateRole 创建角色
// POST /api/v1/admin/roles
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

// UpdateRole 更新角色权限
// PUT /api/v1/admin/roles/{id}
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

// DeleteRole 删除角色
// DELETE /api/v1/admin/roles/{id}
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

// AdminUserListResponse 用户管理列表响应
type AdminUserListResponse struct {
	Users []AdminUserInfo `json:"users"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
}

// AdminUserInfo 用户管理信息
type AdminUserInfo struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Bio          string    `json:"bio,omitempty"`
	RoleID       uuid.UUID `json:"role_id"`
	RoleName     string    `json:"role_name,omitempty"`
	Status       string    `json:"status"`
	PostCount    int       `json:"post_count,omitempty"`
	CreatedAt    string    `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
}

// BanUserRequest 封禁用户请求
type BanUserRequest struct {
	Reason string `json:"reason,omitempty"` // 封禁原因
}

// ListUsers 用户管理列表
// GET /api/v1/admin/users
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

// BanUser 封禁/解封用户
// PUT /api/v1/admin/users/{id}/ban
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

// BatchRequest 批量操作请求
type BatchRequest struct {
	Action string      `json:"action"`           // delete, move_category, change_status
	IDs    []uuid.UUID `json:"ids"`              // 目标 ID 列表
	Params map[string]any `json:"params,omitempty"` // 操作参数
}

// BatchResponse 批量操作响应
type BatchResponse struct {
	Success  int      `json:"success"`           // 成功数量
	Failed   int      `json:"failed"`            // 失败数量
	Errors   []string `json:"errors,omitempty"`  // 错误信息
	Message  string   `json:"message"`
}

// BatchOperation 批量操作
// POST /api/v1/admin/batch
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

// OrderRequest 排序更新请求
type OrderRequest struct {
	Type   string      `json:"type"`   // category, tag
	Order  []uuid.UUID `json:"order"`  // 新的排序顺序（ID 列表）
}

// UpdateOrder 更新排序
// PUT /api/v1/admin/order
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

func (h *AdminHandler) getPostCount(ctx context.Context, userID uuid.UUID) (int, error) {
	// 使用 postRepo 获取用户文章数
	dbPostRepo, ok := h.postRepo.(*database.PostRepository)
	if ok {
		return dbPostRepo.CountByAuthor(ctx, userID)
	}
	return 0, nil
}

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

func (h *AdminHandler) updateCategoryOrder(ctx context.Context, order []uuid.UUID) error {
	dbCategoryRepo, ok := h.categoryRepo.(*database.CategoryRepository)
	if !ok {
		return nil
	}
	return dbCategoryRepo.UpdateOrder(ctx, order)
}

func (h *AdminHandler) updateTagOrder(ctx context.Context, order []uuid.UUID) error {
	dbTagRepo, ok := h.tagRepo.(*database.TagRepository)
	if !ok {
		return nil
	}
	return dbTagRepo.UpdateOrder(ctx, order)
}

// parseIntStr 解析整数字符串
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