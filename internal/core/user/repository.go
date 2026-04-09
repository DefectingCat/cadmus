// Package user 提供用户、角色、权限管理的数据访问接口。
//
// 该文件定义用户系统的 Repository 接口，包括：
//   - UserRepository: 用户数据访问接口
//   - RoleRepository: 角色数据访问接口
//   - PermissionRepository: 权限数据访问接口
//
// 主要用途：
//
//	抽象用户数据访问层，便于实现不同的存储后端，
//	并支持单元测试时使用 mock 实现。
//
// 注意事项：
//   - 所有接口方法必须支持 context.Context 进行超时控制
//   - 返回的错误应使用 models.go 中定义的语义化错误类型
//   - 接口实现必须是并发安全的
//
// 作者：xfy
package user

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository 用户数据访问接口。
//
// 定义用户实体的 CRUD 操作和查询方法。
// 实现该接口的类必须保证所有方法的并发安全性。
type UserRepository interface {
	// Create 创建新用户。
	//
	// 参数：
	//   - ctx: 上下文，用于控制超时和取消操作
	//   - user: 用户对象，必填字段包括 Username、Email、RoleID
	//
	// 返回值：
	//   - err: 可能的错误包括 ErrUserAlreadyExists（用户名或邮箱冲突）
	Create(ctx context.Context, user *User) error

	// GetByID 根据 ID 获取用户。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 用户 ID
	//
	// 返回值：
	//   - user: 用户对象
	//   - err: 用户不存在时返回 ErrUserNotFound
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail 根据邮箱获取用户。
	//
	// 用于登录验证和邮箱相关的查询。
	//
	// 参数：
	//   - ctx: 上下文
	//   - email: 用户邮箱
	//
	// 返回值：
	//   - user: 用户对象
	//   - err: 用户不存在时返回 ErrUserNotFound
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername 根据用户名获取用户。
	//
	// 用于登录验证和用户名相关的查询。
	//
	// 参数：
	//   - ctx: 上下文
	//   - username: 用户名
	//
	// 返回值：
	//   - user: 用户对象
	//   - err: 用户不存在时返回 ErrUserNotFound
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Update 更新用户信息。
	//
	// 参数：
	//   - ctx: 上下文
	//   - user: 用户对象，ID 字段必须有效
	//
	// 返回值：
	//   - err: 用户不存在时返回 ErrUserNotFound
	Update(ctx context.Context, user *User) error

	// Delete 删除用户。
	//
	// 注意：通常建议使用软删除（设置状态为 banned）而非物理删除，
	// 以保留用户数据用于审计和历史记录。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 用户 ID
	//
	// 返回值：
	//   - err: 用户不存在时返回 ErrUserNotFound
	Delete(ctx context.Context, id uuid.UUID) error

	// List 分页获取用户列表。
	//
	// 用于后台管理用户列表展示。
	//
	// 参数：
	//   - ctx: 上下文
	//   - offset: 分页偏移量（从 0 开始）
	//   - limit: 每页数量
	//
	// 返回值：
	//   - users: 用户列表
	//   - total: 用户总数
	//   - err: 查询错误
	List(ctx context.Context, offset, limit int) ([]*User, int, error)
}

// RoleRepository 角色数据访问接口。
//
// 定义角色实体的查询方法。
// 角色通常由系统初始化时创建，运行时主要是查询操作。
type RoleRepository interface {
	// GetByID 根据 ID 获取角色（不含权限列表）。
	//
	// 仅返回角色的基本信息，不加载关联的权限列表。
	// 适用于仅需角色元信息的场景。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 角色 ID
	//
	// 返回值：
	//   - role: 角色对象（不含 Permissions 字段）
	//   - err: 角色不存在时返回 ErrRoleNotFound
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)

	// GetByName 根据名称获取角色（含权限列表）。
	//
	// 返回角色的完整信息，包括关联的权限列表。
	// 适用于需要判断用户权限的场景。
	//
	// 参数：
	//   - ctx: 上下文
	//   - name: 角色内部名称，如 "admin"、"editor"
	//
	// 返回值：
	//   - role: 角色对象（含 Permissions 字段）
	//   - err: 角色不存在时返回 ErrRoleNotFound
	GetByName(ctx context.Context, name string) (*Role, error)

	// GetAll 获取所有角色。
	//
	// 返回系统中所有角色的基本信息，用于角色选择列表等场景。
	//
	// 参数：
	//   - ctx: 上下文
	//
	// 返回值：
	//   - roles: 角色列表
	//   - err: 查询错误
	GetAll(ctx context.Context) ([]*Role, error)

	// GetDefault 获取默认角色。
	//
	// 返回系统设置的默认角色，新注册用户自动分配此角色。
	// 每个系统应有一个且仅有一个默认角色。
	//
	// 参数：
	//   - ctx: 上下文
	//
	// 返回值：
	//   - role: 默认角色对象
	//   - err: 无默认角色时返回 ErrRoleNotFound
	GetDefault(ctx context.Context) (*Role, error)

	// GetWithPermissions 根据 ID 获取角色及其权限。
	//
	// 返回角色的完整信息，包括关联的权限列表。
	// 功能与 GetByName 类似，但通过 ID 查询。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 角色 ID
	//
	// 返回值：
	//   - role: 角色对象（含 Permissions 字段）
	//   - err: 角色不存在时返回 ErrRoleNotFound
	GetWithPermissions(ctx context.Context, id uuid.UUID) (*Role, error)
}

// PermissionRepository 权限数据访问接口。
//
// 定义权限实体的查询方法。
// 权限由系统定义，运行时主要是查询操作，不支持动态增删。
type PermissionRepository interface {
	// GetByRoleID 获取角色的所有权限。
	//
	// 查询指定角色拥有的所有权限列表。
	// 用于判断用户权限或展示角色权限详情。
	//
	// 参数：
	//   - ctx: 上下文
	//   - roleID: 角色 ID
	//
	// 返回值：
	//   - permissions: 权限列表
	//   - err: 查询错误
	GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]Permission, error)

	// GetAll 获取所有权限。
	//
	// 返回系统中定义的所有权限，用于权限管理界面展示。
	//
	// 参数：
	//   - ctx: 上下文
	//
	// 返回值：
	//   - permissions: 权限列表
	//   - err: 查询错误
	GetAll(ctx context.Context) ([]Permission, error)

	// GetByCategory 获取指定类别的权限。
	//
	// 权限按类别组织，如 "post"、"comment"、"user" 等。
	// 用于分类展示权限列表。
	//
	// 参数：
	//   - ctx: 上下文
	//   - category: 权限类别
	//
	// 返回值：
	//   - permissions: 该类别下的权限列表
	//   - err: 查询错误
	GetByCategory(ctx context.Context, category string) ([]Permission, error)

	// CheckPermission 检查角色是否拥有指定权限。
	//
	// 快速判断角色是否拥有某个具体权限，用于权限验证。
	//
	// 参数：
	//   - ctx: 上下文
	//   - roleID: 角色 ID
	//   - permissionName: 权限名称，如 "post.create"
	//
	// 返回值：
	//   - has: true 表示拥有该权限
	//   - err: 查询错误
	CheckPermission(ctx context.Context, roleID uuid.UUID, permissionName string) (bool, error)
}
