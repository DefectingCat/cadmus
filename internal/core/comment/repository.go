// Package comment 提供评论管理的数据访问接口。
//
// 该文件定义评论系统的 Repository 接口，包括：
//   - CommentRepository: 评论数据访问接口
//   - CommentLikeRepository: 评论点赞数据访问接口
//
// 主要用途：
//
//	抽象评论数据访问层，便于实现不同的存储后端，
//	并支持单元测试时使用 mock 实现。
//
// 注意事项：
//   - 所有接口方法必须支持 context.Context 进行超时控制
//   - 返回的错误应使用 models.go 中定义的语义化错误类型
//   - 接口实现必须是并发安全的
//
// 作者：xfy
package comment

import (
	"context"

	"github.com/google/uuid"
)

// CommentRepository 评论数据访问接口。
//
// 定义评论实体的 CRUD 操作和查询方法，支持嵌套回复结构和状态管理。
// 实现该接口的类必须保证所有方法的并发安全性。
type CommentRepository interface {
	// Create 创建新评论。
	//
	// 参数：
	//   - ctx: 上下文，用于控制超时和取消操作
	//   - input: 创建评论输入，包含文章ID、用户ID、内容等
	//
	// 返回值：
	//   - comment: 创建成功的评论对象
	//   - err: 可能的错误包括 ErrPostNotFound、ErrParentNotFound、ErrMaxDepthExceeded
	Create(ctx context.Context, input *CreateCommentInput) (*Comment, error)

	// GetByID 根据 ID 获取评论。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 评论 ID
	//
	// 返回值：
	//   - comment: 评论对象
	//   - err: 评论不存在时返回 ErrCommentNotFound
	GetByID(ctx context.Context, id uuid.UUID) (*Comment, error)

	// GetByPostID 获取文章的所有评论。
	//
	// 支持按条件筛选，如状态、深度、父评论等。
	// 返回的评论按创建时间排序。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//   - filters: 筛选条件，可选
	//
	// 返回值：
	//   - comments: 评论列表
	//   - err: 查询错误
	GetByPostID(ctx context.Context, postID uuid.UUID, filters *CommentListFilters) ([]*Comment, error)

	// GetByUserID 获取用户的所有评论。
	//
	// 用于展示用户的评论历史或分析用户活动。
	//
	// 参数：
	//   - ctx: 上下文
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - comments: 评论列表
	//   - err: 查询错误
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Comment, error)

	// GetChildren 获取评论的子评论（回复）。
	//
	// 查询指定评论的直接回复列表，用于递归加载嵌套结构。
	//
	// 参数：
	//   - ctx: 上下文
	//   - parentID: 父评论 ID
	//
	// 返回值：
	//   - comments: 子评论列表
	//   - err: 查询错误
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*Comment, error)

	// Update 更新评论。
	//
	// 用于修改评论内容，更新时间会自动刷新。
	//
	// 参数：
	//   - ctx: 上下文
	//   - comment: 评论对象，ID 字段必须有效
	//
	// 返回值：
	//   - err: 评论不存在时返回 ErrCommentNotFound，权限不足返回 ErrPermissionDenied
	Update(ctx context.Context, comment *Comment) error

	// UpdateStatus 更新评论状态。
	//
	// 用于审核评论（批准/标记垃圾）或软删除评论。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 评论 ID
	//   - status: 新状态值
	//
	// 返回值：
	//   - err: 评论不存在返回 ErrCommentNotFound，状态无效返回 ErrInvalidStatus
	UpdateStatus(ctx context.Context, id uuid.UUID, status CommentStatus) error

	// Delete 删除评论。
	//
	// 执行软删除操作，将评论状态标记为 deleted，内容保留用于审计。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 评论 ID
	//
	// 返回值：
	//   - err: 评论不存在返回 ErrCommentNotFound
	Delete(ctx context.Context, id uuid.UUID) error

	// CountByPostID 统计文章的评论数量。
	//
	// 用于展示文章的评论数，默认统计已批准状态的评论。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//
	// 返回值：
	//   - count: 评论数量
	//   - err: 查询错误
	CountByPostID(ctx context.Context, postID uuid.UUID) (int, error)

	// CountByUserID 统计用户的评论数量。
	//
	// 用于展示用户的活跃度统计。
	//
	// 参数：
	//   - ctx: 上下文
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - count: 评论数量
	//   - err: 查询错误
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// List 分页获取评论列表。
	//
	// 支持多条件筛选和分页，用于后台管理评论列表。
	//
	// 参数：
	//   - ctx: 上下文
	//   - filters: 筛选条件
	//   - offset: 分页偏移量（从 0 开始）
	//   - limit: 每页数量
	//
	// 返回值：
	//   - comments: 评论列表
	//   - err: 查询错误
	List(ctx context.Context, filters *CommentListFilters, offset, limit int) ([]*Comment, error)
}

// CommentLikeRepository 评论点赞数据访问接口。
//
// 定义评论点赞记录的创建、删除和查询方法。
// 点赞记录用于防止重复点赞和查询用户点赞历史。
type CommentLikeRepository interface {
	// Create 创建点赞记录。
	//
	// 创建新的点赞记录并更新评论的点赞计数。
	//
	// 参数：
	//   - ctx: 上下文
	//   - commentID: 评论 ID
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - like: 创建成功的点赞记录
	//   - err: 已点赞返回 ErrAlreadyLiked
	Create(ctx context.Context, commentID, userID uuid.UUID) (*CommentLike, error)

	// CreateIfNotExists 创建点赞记录（使用 ON CONFLICT DO NOTHING）。
	//
	// 使用原子操作确保不会重复点赞，同时更新评论的点赞计数。
	// 适用于高并发场景，避免竞态条件。
	//
	// 参数：
	//   - ctx: 上下文
	//   - commentID: 评论 ID
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - created: true 表示实际创建了记录（新点赞），false 表示已存在
	//   - err: 操作错误
	CreateIfNotExists(ctx context.Context, commentID, userID uuid.UUID) (created bool, err error)

	// GetByCommentAndUser 获取用户对评论的点赞记录。
	//
	// 用于查询特定的点赞记录详情。
	//
	// 参数：
	//   - ctx: 上下文
	//   - commentID: 评论 ID
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - like: 点赞记录
	//   - err: 未点赞返回 ErrNotLiked
	GetByCommentAndUser(ctx context.Context, commentID, userID uuid.UUID) (*CommentLike, error)

	// Delete 删除点赞记录（取消点赞）。
	//
	// 删除点赞记录并更新评论的点赞计数。
	//
	// 参数：
	//   - ctx: 上下文
	//   - commentID: 评论 ID
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - err: 未点赞返回 ErrNotLiked
	Delete(ctx context.Context, commentID, userID uuid.UUID) error

	// DeleteIfExists 删除点赞记录（返回是否实际删除）。
	//
	// 使用原子操作确保取消点赞的准确性，同时更新评论的点赞计数。
	// 适用于高并发场景，避免竞态条件。
	//
	// 参数：
	//   - ctx: 上下文
	//   - commentID: 评论 ID
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - deleted: true 表示实际删除了记录，false 表示不存在
	//   - err: 操作错误
	DeleteIfExists(ctx context.Context, commentID, userID uuid.UUID) (deleted bool, err error)

	// Exists 检查用户是否已点赞评论。
	//
	// 用于前端展示点赞状态。
	//
	// 参数：
	//   - ctx: 上下文
	//   - commentID: 评论 ID
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - exists: true 表示已点赞
	//   - err: 查询错误
	Exists(ctx context.Context, commentID, userID uuid.UUID) (bool, error)

	// CountByCommentID 统计评论的点赞数量。
	//
	// 用于展示评论的点赞数，与 Comment.LikeCount 字段对应。
	//
	// 参数：
	//   - ctx: 上下文
	//   - commentID: 评论 ID
	//
	// 返回值：
	//   - count: 点赞数量
	//   - err: 查询错误
	CountByCommentID(ctx context.Context, commentID uuid.UUID) (int, error)

	// GetByUserID 获取用户的所有点赞记录。
	//
	// 用于展示用户的点赞历史或分析用户兴趣偏好。
	//
	// 参数：
	//   - ctx: 上下文
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - likes: 点赞记录列表
	//   - err: 查询错误
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*CommentLike, error)
}
