// Package services 提供媒体服务的实现。
//
// 该文件包含媒体文件管理相关的核心逻辑，包括：
//   - 文件上传（支持多种 MIME 类型）
//   - 文件大小限制验证
//   - 唯一文件名生成
//   - 媒体元数据管理
//   - 文件删除（同时删除物理文件）
//
// 主要用途：
//
//	用于处理用户上传媒体文件的完整生命周期管理。
//
// 安全设计：
//   - 文件大小限制（默认 10MB）
//   - MIME 类型白名单验证
//   - 权限检查（仅上传者可删除）
//
// 作者：xfy
package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/media"
	"rua.plus/cadmus/internal/logger"
)

// MaxFileSize 最大文件大小限制。
//
// 设为 10MB，平衡服务器存储成本和用户上传需求。
// 超过此大小的文件将被拒绝上传。
const MaxFileSize = 10 * 1024 * 1024

// MediaService 媒体业务服务接口。
//
// 该接口定义了媒体文件管理的核心操作，包括上传、查询和删除。
// 所有方法均为并发安全。
type MediaService interface {
	// Upload 上传文件，保存到存储后端并创建数据库记录。
	//
	// 参数：
	//   - ctx: 上下文
	//   - userID: 上传用户 ID（用于权限验证）
	//   - file: 上传的文件对象（来自 multipart 表单）
	//   - altText: 替代文本（用于无障碍访问）
	//
	// 返回值：
	//   - media: 创建的媒体对象
	//   - err: 可能的错误包括文件过大、MIME 类型不支持
	//
	// 使用示例：
	//   m, err := mediaService.Upload(ctx, userID, fileHeader, &altText)
	Upload(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader, altText *string) (*media.Media, error)

	// GetByID 根据 ID 获取媒体信息。
	GetByID(ctx context.Context, id uuid.UUID) (*media.Media, error)

	// GetByUser 获取用户上传的所有媒体。
	GetByUser(ctx context.Context, userID uuid.UUID) ([]*media.Media, error)

	// Delete 删除媒体，需验证权限。
	//
	// 仅上传者可以删除自己上传的媒体。
	// 删除操作会同时删除数据库记录和物理文件。
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// List 分页获取媒体列表，支持筛选。
	List(ctx context.Context, filters *media.MediaListFilters, offset, limit int) ([]*media.Media, int, error)
}

// mediaServiceImpl 媒体服务的具体实现。
type mediaServiceImpl struct {
	// mediaRepo 媒体数据仓库
	mediaRepo media.MediaRepository

	// uploadDir 文件上传目录（物理存储路径）
	uploadDir string

	// baseURL 基础 URL（用于生成文件访问链接）
	baseURL string
}

// NewMediaService 创建媒体服务实例。
//
// 参数：
//   - mediaRepo: 媒体数据仓库
//   - uploadDir: 文件存储目录路径
//   - baseURL: 外部访问的基础 URL
//
// 返回值：
//   - MediaService: 媒体服务实例
func NewMediaService(mediaRepo media.MediaRepository, uploadDir, baseURL string) MediaService {
	return &mediaServiceImpl{
		mediaRepo: mediaRepo,
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// Upload 上传文件
func (s *mediaServiceImpl) Upload(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader, altText *string) (*media.Media, error) {
	// 验证文件大小
	if file.Size > MaxFileSize {
		return nil, media.ErrFileSizeTooLarge
	}

	// 打开文件
	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Warn(fmt.Sprintf("failed to close uploaded file: %v", err))
		}
	}()

	// 验证 MIME 类型
	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		// 尝试从文件名推断
		ext := filepath.Ext(file.Filename)
		mimeType = mimeTypeFromExt(ext)
	}

	if !media.AllowedMimeTypes[mimeType] {
		return nil, media.ErrInvalidMimeType
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := generateUniqueFilename(ext)

	// 确保上传目录存在
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// 创建目标文件路径
	filepath := filepath.Join(s.uploadDir, filename)
	dst, err := os.Create(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if err := dst.Close(); err != nil {
			logger.Warn(fmt.Sprintf("failed to close destination file: %v", err))
		}
	}()

	// 复制文件内容
	if _, err := io.Copy(dst, f); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 获取图片尺寸（如果是图片）
	var width, height *int
	if media.IsImageMimeType(mimeType) {
		w, h, err := getImageDimensions(filepath)
		if err == nil {
			width = &w
			height = &h
		}
	}

	// 生成 URL
	url := fmt.Sprintf("%s/uploads/%s", s.baseURL, filename)

	// 创建媒体记录
	input := &media.UploadInput{
		UploaderID:   userID,
		OriginalName: file.Filename,
		MimeType:     mimeType,
		Size:         file.Size,
		AltText:      altText,
	}

	m, err := s.mediaRepo.Create(ctx, input, filename, filepath, url, width, height)
	if err != nil {
		// 删除已保存的文件
		if removeErr := os.Remove(filepath); removeErr != nil {
			logger.Warn(fmt.Sprintf("failed to remove file %s: %v", filepath, removeErr))
		}
		return nil, err
	}

	return m, nil
}

// GetByID 根据 ID 获取媒体
func (s *mediaServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*media.Media, error) {
	return s.mediaRepo.GetByID(ctx, id)
}

// GetByUser 获取用户上传的媒体
func (s *mediaServiceImpl) GetByUser(ctx context.Context, userID uuid.UUID) ([]*media.Media, error) {
	return s.mediaRepo.GetByUploaderID(ctx, userID)
}

// Delete 删除媒体
func (s *mediaServiceImpl) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// 获取媒体信息
	m, err := s.mediaRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查权限：只有上传者可以删除
	if m.UploaderID != userID {
		return media.ErrPermissionDenied
	}

	// 删除数据库记录
	if err := s.mediaRepo.Delete(ctx, id); err != nil {
		return err
	}

	// 删除物理文件
	if err := os.Remove(m.FilePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		// 文件删除失败但数据库记录已删除，记录日志但不返回错误
		logger.Warn(fmt.Sprintf("warning: failed to delete file %s: %v", m.FilePath, err))
	}

	return nil
}

// List 分页获取媒体列表
func (s *mediaServiceImpl) List(ctx context.Context, filters *media.MediaListFilters, offset, limit int) ([]*media.Media, int, error) {
	medias, err := s.mediaRepo.List(ctx, filters, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.mediaRepo.Count(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	return medias, count, nil
}

// generateUniqueFilename 生成唯一文件名
func generateUniqueFilename(ext string) string {
	return fmt.Sprintf("%s%s", uuid.New().String(), ext)
}

// mimeTypeFromExt 根据扩展名推断 MIME 类型
func mimeTypeFromExt(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".zip":
		return "application/zip"
	case ".txt":
		return "text/plain"
	default:
		return ""
	}
}

// getImageDimensions 获取图片尺寸
func getImageDimensions(filepath string) (int, int, error) {
	// 这里简化实现，实际项目中应使用 image 库解码图片获取尺寸
	// 由于不引入额外的 image 解码库，返回零值
	// 实际实现示例：
	// file, err := os.Open(filepath)
	// if err != nil { return 0, 0, err }
	// defer file.Close()
	// img, _, err := image.DecodeConfig(file)
	// if err != nil { return 0, 0, err }
	// return img.Width, img.Height, nil
	return 0, 0, errors.New("image dimensions extraction not implemented")
}
