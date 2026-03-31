// Package media 媒体文件管理模块
package media

import (
	"context"

	"github.com/google/uuid"
)

// MediaRepository 媒体仓库接口
type MediaRepository interface {
	// Create 创建媒体记录
	Create(ctx context.Context, input *UploadInput, filename, filepath, url string, width, height *int) (*Media, error)

	// GetByID 根据 ID 获取媒体
	GetByID(ctx context.Context, id uuid.UUID) (*Media, error)

	// GetByUploaderID 获取用户上传的所有媒体
	GetByUploaderID(ctx context.Context, uploaderID uuid.UUID) ([]*Media, error)

	// Delete 删除媒体记录
	Delete(ctx context.Context, id uuid.UUID) error

	// List 分页获取媒体列表
	List(ctx context.Context, filters *MediaListFilters, offset, limit int) ([]*Media, error)

	// Count 统计媒体数量
	Count(ctx context.Context, filters *MediaListFilters) (int, error)
}