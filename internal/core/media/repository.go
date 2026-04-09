// Package media 提供媒体文件管理的数据访问接口。
//
// 该文件定义媒体系统的 Repository 接口。
//
// 主要用途：
//
//	抽象媒体数据访问层，便于实现不同的存储后端，
//	如本地文件系统、云存储（S3、OSS）等。
//
// 注意事项：
//   - 所有接口方法必须支持 context.Context 进行超时控制
//   - 返回的错误应使用 models.go 中定义的语义化错误类型
//   - 接口实现必须是并发安全的
//
// 作者：xfy
package media

import (
	"context"

	"github.com/google/uuid"
)

// MediaRepository 媒体数据访问接口。
//
// 定义媒体文件的 CRUD 操作和查询方法。
// 实现该接口的类必须保证所有方法的并发安全性。
type MediaRepository interface {
	// Create 创建媒体记录。
	//
	// 文件上传成功后，调用此方法在数据库中创建媒体记录。
	// 返回创建成功的媒体对象，包含生成的 ID。
	//
	// 参数：
	//   - ctx: 上下文，用于控制超时和取消操作
	//   - input: 上传文件输入信息
	//   - filename: 存储文件名（系统生成的唯一名称）
	//   - filepath: 文件存储路径
	//   - url: 文件访问 URL
	//   - width: 图片宽度（可选，仅图片类型）
	//   - height: 图片高度（可选，仅图片类型）
	//
	// 返回值：
	//   - media: 创建成功的媒体对象
	//   - err: 创建失败错误
	//
	// 使用示例：
	//   media, err := repo.Create(ctx, input, "uuid.jpg", "uploads/2024/01/uuid.jpg", "https://cdn.example.com/...", 800, 600)
	Create(ctx context.Context, input *UploadInput, filename, filepath, url string, width, height *int) (*Media, error)

	// GetByID 根据 ID 获取媒体。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 媒体 ID
	//
	// 返回值：
	//   - media: 媒体对象
	//   - err: 媒体不存在时返回 ErrMediaNotFound
	GetByID(ctx context.Context, id uuid.UUID) (*Media, error)

	// GetByUploaderID 获取用户上传的所有媒体。
	//
	// 返回指定用户上传的所有媒体文件列表。
	// 用于用户媒体库管理页面。
	//
	// 参数：
	//   - ctx: 上下文
	//   - uploaderID: 上传者用户 ID
	//
	// 返回值：
	//   - mediaList: 媒体列表
	//   - err: 查询错误
	GetByUploaderID(ctx context.Context, uploaderID uuid.UUID) ([]*Media, error)

	// Delete 删除媒体记录。
	//
	// 删除数据库中的媒体记录。通常应同时删除实际文件。
	// 注意：删除操作不可逆，应确保用户有权限执行删除。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 媒体 ID
	//
	// 返回值：
	//   - err: 媒体不存在返回 ErrMediaNotFound，权限不足返回 ErrPermissionDenied
	Delete(ctx context.Context, id uuid.UUID) error

	// List 分页获取媒体列表。
	//
	// 支持按上传者、类型等条件筛选，用于后台管理媒体列表。
	//
	// 参数：
	//   - ctx: 上下文
	//   - filters: 筛选条件
	//   - offset: 分页偏移量（从 0 开始）
	//   - limit: 每页数量
	//
	// 返回值：
	//   - mediaList: 媒体列表
	//   - err: 查询错误
	List(ctx context.Context, filters *MediaListFilters, offset, limit int) ([]*Media, error)

	// Count 统计媒体数量。
	//
	// 统计符合条件的媒体文件总数，用于计算分页信息。
	//
	// 参数：
	//   - ctx: 上下文
	//   - filters: 筛选条件
	//
	// 返回值：
	//   - count: 符合条件的媒体数量
	//   - err: 查询错误
	Count(ctx context.Context, filters *MediaListFilters) (int, error)
}
