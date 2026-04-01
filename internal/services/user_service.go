// Package services 提供用户服务的实现。
//
// 该文件包含用户管理相关的核心逻辑，包括：
//   - 用户注册（邮箱验证、密码哈希、默认角色）
//   - 用户信息查询（ID、邮箱、用户名）
//   - 用户信息更新和删除
//   - 用户列表分页查询
//
// 主要用途：
//
//	用于处理用户账户的完整生命周期管理。
//
// 安全设计：
//   - 注册时自动获取默认角色，避免权限配置错误
//   - 新用户默认为 Pending 状态，需管理员审核
//   - 删除采用软删除方式，保留数据可恢复
//
// 作者：xfy
package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/user"
)

// UserService 用户业务服务接口。
//
// 该接口定义了用户管理的核心操作，包括注册、查询、更新和删除。
// 所有方法均为并发安全，可在多个 goroutine 中同时调用。
type UserService interface {
	// Register 注册新用户，执行完整的注册流程。
	//
	// 参数：
	//   - ctx: 上下文，用于控制超时和取消操作
	//   - username: 用户名，唯一标识用户
	//   - email: 邮箱地址，用于登录和通知
	//   - password: 用户密码（明文，内部进行哈希处理）
	//
	// 返回值：
	//   - user: 创建成功的用户对象
	//   - err: 可能的错误包括：
	//       - "username, email and password are required": 必填字段缺失
	//       - user.ErrUserAlreadyExists: 邮箱或用户名已存在
	//       - "failed to get default role": 无法获取默认角色
	//       - "failed to hash password": 密码哈希失败
	//
	// 使用示例：
	//   user, err := userService.Register(ctx, "testuser", "test@example.com", "password123")
	Register(ctx context.Context, username, email, password string) (*user.User, error)

	// GetByID 根据用户 ID 获取用户信息。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 用户唯一标识符（UUID）
	//
	// 返回值：
	//   - user: 用户对象
	//   - err: 用户不存在时返回错误
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)

	// GetByEmail 根据邮箱地址获取用户信息。
	//
	// 参数：
	//   - ctx: 上下文
	//   - email: 邮箱地址
	//
	// 返回值：
	//   - user: 用户对象
	//   - err: 用户不存在时返回错误
	GetByEmail(ctx context.Context, email string) (*user.User, error)

	// GetByUsername 根据用户名获取用户信息。
	//
	// 参数：
	//   - ctx: 上下文
	//   - username: 用户名
	//
	// 返回值：
	//   - user: 用户对象
	//   - err: 用户不存在时返回错误
	GetByUsername(ctx context.Context, username string) (*user.User, error)

	// Update 更新用户信息。
	//
	// 参数：
	//   - ctx: 上下文
	//   - u: 用户对象（包含更新后的字段值）
	//
	// 返回值：
	//   - err: 可能的错误包括：
	//       - user.ErrInvalidStatus: 用户状态无效
	//       - 其他数据库错误
	//
	// 注意事项：
	//   - 更新前会验证用户状态有效性
	Update(ctx context.Context, u *user.User) error

	// List 分页获取用户列表。
	//
	// 参数：
	//   - ctx: 上下文
	//   - offset: 偏移量（跳过的记录数）
	//   - limit: 每页最大记录数
	//
	// 返回值：
	//   - users: 用户对象列表
	//   - total: 符合条件的总记录数
	//   - err: 查询错误
	List(ctx context.Context, offset, limit int) ([]*user.User, int, error)

	// Delete 删除用户（软删除）。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 用户唯一标识符
	//
	// 返回值：
	//   - err: 删除错误
	//
	// 注意事项：
	//   - 实际执行软删除，将用户状态设为 banned
	//   - 数据仍保留在数据库中，可恢复
	Delete(ctx context.Context, id uuid.UUID) error
}

// userServiceImpl 用户服务的具体实现。
//
// 该结构体实现了 UserService 接口，依赖 UserRepository 和 RoleRepository。
// 通过依赖注入模式解耦数据访问层和业务逻辑层。
type userServiceImpl struct {
	// userRepo 用户数据仓库，用于用户数据的 CRUD 操作
	userRepo user.UserRepository

	// roleRepo 角色数据仓库，用于获取默认角色
	roleRepo user.RoleRepository
}

// NewUserService 创建用户服务实例。
//
// 参数：
//   - userRepo: 用户数据仓库
//   - roleRepo: 角色数据仓库
//
// 返回值：
//   - UserService: 用户服务接口实例
func NewUserService(userRepo user.UserRepository, roleRepo user.RoleRepository) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Register 注册新用户，执行完整的注册流程。
//
// 该方法执行以下步骤：
//  1. 验证必填字段（用户名、邮箱、密码）
//  2. 检查邮箱唯一性
//  3. 检查用户名唯一性
//  4. 获取默认角色（用于权限初始化）
//  5. 创建用户实体并设置密码哈希
//  6. 持久化到数据库
//
// 参数：
//   - ctx: 上下文，用于控制超时
//   - username: 用户名（必须唯一）
//   - email: 邮箱地址（必须唯一）
//   - password: 用户密码（明文，内部进行 bcrypt 哈希）
//
// 返回值：
//   - user: 创建成功的用户对象（包含生成的 UUID 和默认角色）
//   - err: 注册失败时返回错误
func (s *userServiceImpl) Register(ctx context.Context, username, email, password string) (*user.User, error) {
	// 步骤1: 验证必填字段
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email and password are required")
	}

	// 步骤2: 检查邮箱是否已存在（避免重复注册）
	if existing, _ := s.userRepo.GetByEmail(ctx, email); existing != nil {
		return nil, user.ErrUserAlreadyExists
	}

	// 步骤3: 检查用户名是否已存在（避免重复注册）
	if existing, _ := s.userRepo.GetByUsername(ctx, username); existing != nil {
		return nil, user.ErrUserAlreadyExists
	}

	// 步骤4: 获取默认角色（新用户默认权限）
	defaultRole, err := s.roleRepo.GetDefault(ctx)
	if err != nil {
		return nil, errors.New("failed to get default role")
	}

	// 步骤5: 创建用户实体
	// 新用户默认为 Pending 状态，需管理员审核后才能激活
	newUser := &user.User{
		ID:       uuid.New(),       // 生成唯一 UUID
		Username: username,
		Email:    email,
		RoleID:   defaultRole.ID,   // 关联默认角色
		Status:   user.StatusPending, // 待审核状态
	}

	// 设置密码（使用 bcrypt 进行哈希处理）
	if err := newUser.SetPassword(password); err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 步骤6: 持久化到数据库
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// GetByID 根据 ID 获取用户
func (s *userServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetByEmail 根据邮箱获取用户
func (s *userServiceImpl) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// GetByUsername 根据用户名获取用户
func (s *userServiceImpl) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

// Update 更新用户信息
func (s *userServiceImpl) Update(ctx context.Context, u *user.User) error {
	// 验证用户状态
	if !u.Status.IsValid() {
		return user.ErrInvalidStatus
	}

	return s.userRepo.Update(ctx, u)
}

// List 分页获取用户列表
func (s *userServiceImpl) List(ctx context.Context, offset, limit int) ([]*user.User, int, error) {
	return s.userRepo.List(ctx, offset, limit)
}

// Delete 删除用户（软删除）
func (s *userServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}