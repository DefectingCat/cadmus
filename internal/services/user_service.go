package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/user"
)

// UserService 用户业务服务接口
type UserService interface {
	// Register 注册新用户
	Register(ctx context.Context, username, email, password string) (*user.User, error)

	// GetByID 根据 ID 获取用户
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*user.User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*user.User, error)

	// Update 更新用户信息
	Update(ctx context.Context, u *user.User) error

	// List 分页获取用户列表
	List(ctx context.Context, offset, limit int) ([]*user.User, int, error)

	// Delete 删除用户（软删除，设为 banned 状态）
	Delete(ctx context.Context, id uuid.UUID) error
}

// userServiceImpl 用户服务实现
type userServiceImpl struct {
	userRepo user.UserRepository
	roleRepo user.RoleRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo user.UserRepository, roleRepo user.RoleRepository) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Register 注册新用户
func (s *userServiceImpl) Register(ctx context.Context, username, email, password string) (*user.User, error) {
	// 验证必填字段
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email and password are required")
	}

	// 检查邮箱是否已存在
	if existing, _ := s.userRepo.GetByEmail(ctx, email); existing != nil {
		return nil, user.ErrUserAlreadyExists
	}

	// 检查用户名是否已存在
	if existing, _ := s.userRepo.GetByUsername(ctx, username); existing != nil {
		return nil, user.ErrUserAlreadyExists
	}

	// 获取默认角色
	defaultRole, err := s.roleRepo.GetDefault(ctx)
	if err != nil {
		return nil, errors.New("failed to get default role")
	}

	// 创建用户实体
	newUser := &user.User{
		ID:       uuid.New(),
		Username: username,
		Email:    email,
		RoleID:   defaultRole.ID,
		Status:   user.StatusPending,
	}

	// 设置密码（哈希处理）
	if err := newUser.SetPassword(password); err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 持久化
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