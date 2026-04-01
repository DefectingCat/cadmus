// Package post 提供文章、分类、标签、系列管理的核心数据模型。
//
// 该文件包含内容管理系统的核心实体定义，包括：
//   - 文章实体及其状态枚举
//   - 分类、标签、系列实体
//   - 文章版本历史记录
//   - 点赞记录及筛选条件
//   - SEO 元数据结构
//
// 主要用途：
//
//	用于博客系统的内容管理，支持文章发布、分类组织、标签聚合、系列连载等功能。
//
// 注意事项：
//   - PostStatus 枚举通过 IsValid() 方法验证状态有效性
//   - 文章内容使用 BlockDocument JSON 格式存储（JSONB）
//   - 所有实体使用 UUID 作为主键
//
// 作者：xfy
package post

import (
	"time"

	"github.com/google/uuid"
)

// PostStatus 文章状态枚举。
//
// 定义文章在系统中的生命周期状态，控制文章的可见性和发布行为。
type PostStatus string

// 文章状态常量定义。
const (
	// StatusDraft 草稿状态，文章未发布，仅作者可见
	StatusDraft PostStatus = "draft"

	// StatusPublished 已发布状态，文章公开可见
	StatusPublished PostStatus = "published"

	// StatusScheduled 定时发布状态，等待设定的发布时间到达后自动发布
	StatusScheduled PostStatus = "scheduled"

	// StatusPrivate 私密状态，文章仅对特定用户或付费用户可见
	StatusPrivate PostStatus = "private"
)

// IsValid 检查文章状态是否有效。
//
// 验证状态值是否为预定义的四种状态之一。
//
// 返回值：
//   - true: 状态有效
//   - false: 状态无效
func (s PostStatus) IsValid() bool {
	switch s {
	case StatusDraft, StatusPublished, StatusScheduled, StatusPrivate:
		return true
	default:
		return false
	}
}

// SEOMeta SEO 元数据结构。
//
// 用于存储文章的搜索引擎优化信息，影响文章在搜索结果中的展示效果。
// 所有字段均为可选，未设置时使用默认值或文章基本信息。
type SEOMeta struct {
	// Title SEO 标题，显示在搜索结果和浏览器标签中
	Title string `json:"title,omitempty"`

	// Description SEO 描述，显示在搜索结果摘要中
	Description string `json:"description,omitempty"`

	// Keywords SEO 关键词，用于搜索引擎索引
	Keywords []string `json:"keywords,omitempty"`
}

// Post 文章实体。
//
// 表示博客系统中的一篇文章，包含内容、元数据、状态等完整信息。
// 支持版本管理、付费内容、系列连载等高级功能。
//
// 注意事项：
//   - ID 由系统自动生成，无需手动设置
//   - Content 使用 BlockDocument JSON 格式存储
//   - ContentText 为纯文本版本，用于全文搜索
//   - CreatedAt/UpdatedAt 使用 UTC 时间
type Post struct {
	// ID 文章的唯一标识符，格式为 UUID
	ID uuid.UUID `json:"id"`

	// AuthorID 作者的用户 ID
	AuthorID uuid.UUID `json:"author_id"`

	// Title 文章标题
	Title string `json:"title"`

	// Slug 文章 URL 别名，用于生成友好链接
	Slug string `json:"slug"`

	// Content 文章正文，存储为 BlockDocument JSON 格式（JSONB）
	Content []byte `json:"content"`

	// ContentText 文章纯文本内容，用于全文搜索索引
	ContentText string `json:"content_text"`

	// Excerpt 文章摘要/简介，显示在列表页和搜索结果中
	Excerpt string `json:"excerpt"`

	// CategoryID 分类 ID，可为空（零值表示未归类）
	CategoryID uuid.UUID `json:"category_id"`

	// Tags 文章标签列表，不直接存储在文章表中，通过关联表查询
	Tags []*Tag `json:"tags,omitempty"`

	// Status 文章状态，控制可见性和发布行为
	Status PostStatus `json:"status"`

	// PublishAt 定时发布时间，用于 StatusScheduled 状态的文章
	PublishAt *time.Time `json:"publish_at"`

	// FeaturedImage 特色图片 URL，显示在文章列表和分享预览中
	FeaturedImage string `json:"featured_image,omitempty"`

	// SEOMeta SEO 元数据，优化搜索引擎展示效果
	SEOMeta SEOMeta `json:"seo_meta"`

	// ViewCount 浏览次数统计
	ViewCount int `json:"view_count"`

	// LikeCount 点赞次数统计
	LikeCount int `json:"like_count"`

	// CommentCount 评论次数统计
	CommentCount int `json:"comment_count"`

	// SeriesID 所属系列 ID，可为空（表示不属于任何系列）
	SeriesID *uuid.UUID `json:"series_id"`

	// SeriesOrder 系列内排序序号，用于系列文章的顺序展示
	SeriesOrder int `json:"series_order"`

	// IsPaid 是否为付费文章
	IsPaid bool `json:"is_paid"`

	// Price 付费价格，仅当 IsPaid 为 true 时有效
	Price *float64 `json:"price"`

	// Version 文章版本号，每次保存时递增
	Version int `json:"version"`

	// CreatedAt 创建时间，使用 UTC 时间戳
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 最后修改时间，使用 UTC 时间戳
	UpdatedAt time.Time `json:"updated_at"`
}

// Category 分类实体。
//
// 用于组织文章的层级分类结构，支持父子关系嵌套。
// 每个分类可包含多个子分类，形成树形结构。
type Category struct {
	// ID 分类的唯一标识符，格式为 UUID
	ID uuid.UUID `json:"id"`

	// Name 分类名称，显示在前端界面
	Name string `json:"name"`

	// Slug 分类 URL 别名，用于生成友好链接
	Slug string `json:"slug"`

	// Description 分类描述，显示在分类页面
	Description string `json:"description,omitempty"`

	// ParentID 父分类 ID，可为空（表示顶级分类）
	ParentID *uuid.UUID `json:"parent_id"`

	// SortOrder 排序权重，用于分类列表的显示顺序
	SortOrder int `json:"sort_order"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 最后修改时间
	UpdatedAt time.Time `json:"updated_at"`
}

// Tag 标签实体。
//
// 用于文章的横向标记和聚合，与分类的层级结构不同，
// 标签是扁平的、可自由组合的文章属性。
type Tag struct {
	// ID 标签的唯一标识符
	ID uuid.UUID `json:"id"`

	// Name 标签名称
	Name string `json:"name"`

	// Slug 标签 URL 别名
	Slug string `json:"slug"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// Series 文章系列实体。
//
// 用于组织连载文章或相关主题文章集合，如教程系列、专题报道等。
// 系列内的文章通过 SeriesOrder 字段排序。
type Series struct {
	// ID 系列的唯一标识符
	ID uuid.UUID `json:"id"`

	// AuthorID 系列创建者的用户 ID
	AuthorID uuid.UUID `json:"author_id"`

	// Title 系列标题
	Title string `json:"title"`

	// Slug 系列 URL 别名
	Slug string `json:"slug"`

	// Description 系列描述，介绍系列内容和目的
	Description string `json:"description,omitempty"`

	// CoverImage 系列封面图片 URL
	CoverImage string `json:"cover_image,omitempty"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 最后修改时间
	UpdatedAt time.Time `json:"updated_at"`
}

// PostVersion 文章版本历史实体。
//
// 用于记录文章的每次修改历史，支持版本回溯和内容恢复。
// 每次文章保存时创建一条新版本记录。
type PostVersion struct {
	// ID 版本记录的唯一标识符
	ID uuid.UUID `json:"id"`

	// PostID 所属文章 ID
	PostID uuid.UUID `json:"post_id"`

	// Version 版本号，与文章的 Version 字段对应
	Version int `json:"version"`

	// Content 该版本的文章内容快照
	Content []byte `json:"content"`

	// CreatorID 创建此版本的用户 ID（通常是文章作者）
	CreatorID uuid.UUID `json:"creator_id"`

	// Note 版本说明，如"修复错别字"、"添加新章节"等
	Note string `json:"note"`

	// CreatedAt 版本创建时间
	CreatedAt time.Time `json:"created_at"`
}

// PostLike 文章点赞实体。
//
// 记录用户对文章的点赞行为，用于统计和查询用户点赞历史。
// 通过唯一约束防止重复点赞。
type PostLike struct {
	// ID 点赞记录的唯一标识符
	ID uuid.UUID `json:"id"`

	// PostID 被点赞文章的 ID
	PostID uuid.UUID `json:"post_id"`

	// UserID 点赞用户的 ID
	UserID uuid.UUID `json:"user_id"`

	// CreatedAt 点赞时间
	CreatedAt time.Time `json:"created_at"`
}

// PostListFilters 文章列表筛选条件。
//
// 用于构建文章查询的过滤条件，支持多条件组合筛选。
// 所有字段均为可选，未设置时忽略该条件。
type PostListFilters struct {
	// Status 按文章状态筛选
	Status PostStatus

	// AuthorID 按作者筛选
	AuthorID uuid.UUID

	// CategoryID 按分类筛选
	CategoryID uuid.UUID

	// SeriesID 按系列筛选
	SeriesID uuid.UUID

	// TagID 按标签筛选
	TagID uuid.UUID

	// Search 搜索关键词，用于全文搜索
	Search string
}

// 常见错误定义。
//
// 使用语义化错误类型，便于调用方进行错误处理和判断。
var (
	// ErrPostNotFound 文章不存在错误，查询文章时 ID 无效时返回
	ErrPostNotFound = &PostError{Code: "post_not_found", Message: "文章不存在"}

	// ErrPostAlreadyExists 文章已存在错误，创建文章时 Slug 冲突时返回
	ErrPostAlreadyExists = &PostError{Code: "post_already_exists", Message: "文章已存在"}

	// ErrInvalidStatus 无效文章状态错误，设置状态值不合法时返回
	ErrInvalidStatus = &PostError{Code: "invalid_status", Message: "无效的文章状态"}

	// ErrCategoryNotFound 分类不存在错误，关联的分类 ID 无效时返回
	ErrCategoryNotFound = &PostError{Code: "category_not_found", Message: "分类不存在"}

	// ErrTagNotFound 标签不存在错误，关联的标签 ID 无效时返回
	ErrTagNotFound = &PostError{Code: "tag_not_found", Message: "标签不存在"}

	// ErrSeriesNotFound 系列不存在错误，关联的系列 ID 无效时返回
	ErrSeriesNotFound = &PostError{Code: "series_not_found", Message: "文章系列不存在"}

	// ErrVersionNotFound 版本不存在错误，查询历史版本时返回
	ErrVersionNotFound = &PostError{Code: "version_not_found", Message: "版本不存在"}

	// ErrPermissionDenied 权限不足错误，用户无权操作文章时返回
	ErrPermissionDenied = &PostError{Code: "permission_denied", Message: "权限不足"}

	// ErrPaidContent 付费内容错误，未购买用户访问付费文章时返回
	ErrPaidContent = &PostError{Code: "paid_content", Message: "此为付费内容，请先购买"}

	// ErrAlreadyLiked 已点赞错误，用户重复点赞同一文章时返回
	ErrAlreadyLiked = &PostError{Code: "already_liked", Message: "已点赞过该文章"}

	// ErrNotLiked 未点赞错误，取消点赞但未点赞过时返回
	ErrNotLiked = &PostError{Code: "not_liked", Message: "未点赞过该文章"}
)

// PostError 文章模块自定义错误类型。
//
// 实现 error 和 errors.Is 接口，支持错误比较和类型判断。
// 通过 Code 字段区分不同错误类型，Message 字段提供人类可读描述。
type PostError struct {
	// Code 错误代码，用于程序化错误判断
	Code string

	// Message 错误消息，用于展示给用户或记录日志
	Message string
}

// Error 实现 error 接口，返回错误消息。
func (e *PostError) Error() string {
	return e.Message
}

// Is 实现 errors.Is 接口，支持错误类型比较。
//
// 通过比较 Code 字段判断是否为同类型错误，
// 便于使用 errors.Is(err, ErrPostNotFound) 进行错误判断。
func (e *PostError) Is(target error) bool {
	t, ok := target.(*PostError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}