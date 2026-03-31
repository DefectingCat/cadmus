// Package post 文章、分类、标签管理模块
package post

import (
	"time"

	"github.com/google/uuid"
)

// PostStatus 文章状态枚举
type PostStatus string

const (
	StatusDraft     PostStatus = "draft"     // 草稿
	StatusPublished PostStatus = "published" // 已发布
	StatusScheduled PostStatus = "scheduled" // 定时发布
	StatusPrivate   PostStatus = "private"   // 私密文章
)

// IsValid 检查文章状态是否有效
func (s PostStatus) IsValid() bool {
	switch s {
	case StatusDraft, StatusPublished, StatusScheduled, StatusPrivate:
		return true
	default:
		return false
	}
}

// SEOMeta SEO 元数据结构
type SEOMeta struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

// Post 文章实体
type Post struct {
	ID           uuid.UUID  `json:"id"`
	AuthorID     uuid.UUID  `json:"author_id"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	Content      []byte     `json:"content"`       // BlockDocument JSON (JSONB)
	ContentText  string     `json:"content_text"`  // 纯文本用于搜索
	Excerpt      string     `json:"excerpt"`       // 文章摘要
	CategoryID   uuid.UUID  `json:"category_id"`   // 可为空（零值表示无分类）
	Tags         []*Tag     `json:"tags,omitempty"` // 文章标签（不存储在表中）
	Status       PostStatus `json:"status"`
	PublishAt    *time.Time `json:"publish_at"`    // 定时发布时间
	FeaturedImage string    `json:"featured_image,omitempty"`
	SEOMeta      SEOMeta    `json:"seo_meta"`
	ViewCount    int        `json:"view_count"`
	LikeCount    int        `json:"like_count"`
	CommentCount int        `json:"comment_count"`
	SeriesID     *uuid.UUID `json:"series_id"`     // 可为空（不属于任何系列）
	SeriesOrder  int        `json:"series_order"`  // 系列内排序
	IsPaid       bool       `json:"is_paid"`       // 是否付费文章
	Price        *float64   `json:"price"`         // 付费价格（可为空）
	Version      int        `json:"version"`       // 版本号
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Category 分类实体
type Category struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id"`      // 父分类ID（可为空）
	SortOrder   int        `json:"sort_order"`     // 排序权重
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Tag 标签实体
type Tag struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

// Series 文章系列实体
type Series struct {
	ID          uuid.UUID `json:"id"`
	AuthorID    uuid.UUID `json:"author_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CoverImage  string    `json:"cover_image,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PostVersion 文章版本历史实体
type PostVersion struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"post_id"`
	Version   int       `json:"version"`
	Content   []byte    `json:"content"`    // BlockDocument JSON (JSONB)
	CreatorID uuid.UUID `json:"creator_id"` // 创建此版本的用户
	Note      string    `json:"note"`       // 版本说明
	CreatedAt time.Time `json:"created_at"`
}

// PostListFilters 文章列表筛选条件
type PostListFilters struct {
	Status     PostStatus
	AuthorID   uuid.UUID
	CategoryID uuid.UUID
	SeriesID   uuid.UUID
	TagID      uuid.UUID
	Search     string
}

// 常见错误定义
var (
	ErrPostNotFound     = &PostError{Code: "post_not_found", Message: "文章不存在"}
	ErrPostAlreadyExists = &PostError{Code: "post_already_exists", Message: "文章已存在"}
	ErrInvalidStatus    = &PostError{Code: "invalid_status", Message: "无效的文章状态"}
	ErrCategoryNotFound = &PostError{Code: "category_not_found", Message: "分类不存在"}
	ErrTagNotFound      = &PostError{Code: "tag_not_found", Message: "标签不存在"}
	ErrSeriesNotFound   = &PostError{Code: "series_not_found", Message: "文章系列不存在"}
	ErrVersionNotFound  = &PostError{Code: "version_not_found", Message: "版本不存在"}
	ErrPermissionDenied = &PostError{Code: "permission_denied", Message: "权限不足"}
	ErrPaidContent      = &PostError{Code: "paid_content", Message: "此为付费内容，请先购买"}
)

// PostError 文章模块自定义错误
type PostError struct {
	Code    string
	Message string
}

func (e *PostError) Error() string {
	return e.Message
}

// Is 实现 errors.Is 接口
func (e *PostError) Is(target error) bool {
	t, ok := target.(*PostError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}