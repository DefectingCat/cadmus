package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/post"
)

// PostService 文章业务服务接口
type PostService interface {
	// Create 创建文章
	Create(ctx context.Context, p *post.Post, tagIDs []uuid.UUID) error

	// Update 更新文章
	Update(ctx context.Context, p *post.Post, tagIDs []uuid.UUID) error

	// Delete 删除文章
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取文章
	GetByID(ctx context.Context, id uuid.UUID) (*post.Post, error)

	// GetBySlug 根据 Slug 获取文章
	GetBySlug(ctx context.Context, slug string) (*post.Post, error)

	// List 分页获取文章列表
	List(ctx context.Context, filters post.PostListFilters, page, pageSize int) ([]*post.Post, int, error)

	// Search 全文搜索
	Search(ctx context.Context, query string, page, pageSize int) ([]*post.Post, int, error)

	// Publish 发布文章
	Publish(ctx context.Context, id uuid.UUID) error

	// Schedule 定时发布
	Schedule(ctx context.Context, id uuid.UUID, publishAt time.Time) error

	// CreateVersion 创建版本
	CreateVersion(ctx context.Context, postID uuid.UUID, note string, creatorID uuid.UUID) error

	// GetVersions 获取版本历史
	GetVersions(ctx context.Context, postID uuid.UUID) ([]*post.PostVersion, error)

	// Rollback 回滚版本
	Rollback(ctx context.Context, postID uuid.UUID, version int) error

	// IncrementViewCount 增加浏览量
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
}

// postServiceImpl 文章服务实现
type postServiceImpl struct {
	postRepo     post.PostRepository
	categoryRepo post.CategoryRepository
	tagRepo      post.TagRepository
	seriesRepo   post.SeriesRepository
}

// NewPostService 创建文章服务
func NewPostService(
	postRepo post.PostRepository,
	categoryRepo post.CategoryRepository,
	tagRepo post.TagRepository,
	seriesRepo post.SeriesRepository,
) PostService {
	return &postServiceImpl{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		seriesRepo:   seriesRepo,
	}
}

// Create 创建文章
func (s *postServiceImpl) Create(ctx context.Context, p *post.Post, tagIDs []uuid.UUID) error {
	// 验证必填字段
	if p.Title == "" || p.Slug == "" {
		return errors.New("title and slug are required")
	}

	// 验证状态
	if p.Status != "" && !p.Status.IsValid() {
		return post.ErrInvalidStatus
	}

	// 默认状态为草稿
	if p.Status == "" {
		p.Status = post.StatusDraft
	}

	// 创建文章
	if err := s.postRepo.Create(ctx, p); err != nil {
		return err
	}

	// 设置标签
	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			if err := s.tagRepo.AddPostTag(ctx, p.ID, tagID); err != nil {
				return err
			}
		}
	}

	return nil
}

// Update 更新文章
func (s *postServiceImpl) Update(ctx context.Context, p *post.Post, tagIDs []uuid.UUID) error {
	// 验证状态
	if p.Status != "" && !p.Status.IsValid() {
		return post.ErrInvalidStatus
	}

	// 更新文章
	if err := s.postRepo.Update(ctx, p); err != nil {
		return err
	}

	// 更新标签：先删除旧的，再添加新的
	tags, err := s.tagRepo.GetPostTags(ctx, p.ID)
	if err == nil {
		for _, t := range tags {
			s.tagRepo.RemovePostTag(ctx, p.ID, t.ID)
		}
	}
	for _, tagID := range tagIDs {
		s.tagRepo.AddPostTag(ctx, p.ID, tagID)
	}

	return nil
}

// Delete 删除文章
func (s *postServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return s.postRepo.Delete(ctx, id)
}

// GetByID 根据 ID 获取文章
func (s *postServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	return s.postRepo.GetByID(ctx, id)
}

// GetBySlug 根据 Slug 获取文章
func (s *postServiceImpl) GetBySlug(ctx context.Context, slug string) (*post.Post, error) {
	return s.postRepo.GetBySlug(ctx, slug)
}

// List 分页获取文章列表
func (s *postServiceImpl) List(ctx context.Context, filters post.PostListFilters, page, pageSize int) ([]*post.Post, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.postRepo.List(ctx, filters, offset, pageSize)
}

// Search 全文搜索
func (s *postServiceImpl) Search(ctx context.Context, query string, page, pageSize int) ([]*post.Post, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.postRepo.Search(ctx, query, offset, pageSize)
}

// Publish 发布文章
func (s *postServiceImpl) Publish(ctx context.Context, id uuid.UUID) error {
	p, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	p.Status = post.StatusPublished
	now := time.Now()
	p.PublishAt = &now

	return s.postRepo.Update(ctx, p)
}

// Schedule 定时发布
func (s *postServiceImpl) Schedule(ctx context.Context, id uuid.UUID, publishAt time.Time) error {
	p, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	p.Status = post.StatusScheduled
	p.PublishAt = &publishAt

	return s.postRepo.Update(ctx, p)
}

// CreateVersion 创建版本
func (s *postServiceImpl) CreateVersion(ctx context.Context, postID uuid.UUID, note string, creatorID uuid.UUID) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	v := &post.PostVersion{
		PostID:    postID,
		Version:   p.Version,
		Content:   p.Content,
		CreatorID: creatorID,
		Note:      note,
	}

	return s.postRepo.CreateVersion(ctx, v)
}

// GetVersions 获取版本历史
func (s *postServiceImpl) GetVersions(ctx context.Context, postID uuid.UUID) ([]*post.PostVersion, error) {
	return s.postRepo.GetVersions(ctx, postID)
}

// Rollback 回滚版本
func (s *postServiceImpl) Rollback(ctx context.Context, postID uuid.UUID, version int) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// 获取指定版本
	versions, err := s.postRepo.GetVersions(ctx, postID)
	if err != nil {
		return err
	}

	var targetVersion *post.PostVersion
	for _, v := range versions {
		if v.Version == version {
			targetVersion = v
			break
		}
	}

	if targetVersion == nil {
		return post.ErrVersionNotFound
	}

	// 恢复内容
	p.Content = targetVersion.Content
	return s.postRepo.Update(ctx, p)
}

// IncrementViewCount 增加浏览量
func (s *postServiceImpl) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return s.postRepo.IncrementViewCount(ctx, id)
}

// CategoryService 分类服务接口
type CategoryService interface {
	Create(ctx context.Context, c *post.Category) error
	Update(ctx context.Context, c *post.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*post.Category, error)
	GetBySlug(ctx context.Context, slug string) (*post.Category, error)
	GetAll(ctx context.Context) ([]*post.Category, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*post.Category, error)
}

type categoryServiceImpl struct {
	repo post.CategoryRepository
}

func NewCategoryService(repo post.CategoryRepository) CategoryService {
	return &categoryServiceImpl{repo: repo}
}

func (s *categoryServiceImpl) Create(ctx context.Context, c *post.Category) error {
	return s.repo.Create(ctx, c)
}

func (s *categoryServiceImpl) Update(ctx context.Context, c *post.Category) error {
	return s.repo.Update(ctx, c)
}

func (s *categoryServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *categoryServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*post.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *categoryServiceImpl) GetBySlug(ctx context.Context, slug string) (*post.Category, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *categoryServiceImpl) GetAll(ctx context.Context) ([]*post.Category, error) {
	return s.repo.GetAll(ctx)
}

func (s *categoryServiceImpl) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*post.Category, error) {
	return s.repo.GetChildren(ctx, parentID)
}

// TagService 标签服务接口
type TagService interface {
	Create(ctx context.Context, t *post.Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*post.Tag, error)
	GetBySlug(ctx context.Context, slug string) (*post.Tag, error)
	GetByName(ctx context.Context, name string) (*post.Tag, error)
	GetAll(ctx context.Context) ([]*post.Tag, error)
}

type tagServiceImpl struct {
	repo post.TagRepository
}

func NewTagService(repo post.TagRepository) TagService {
	return &tagServiceImpl{repo: repo}
}

func (s *tagServiceImpl) Create(ctx context.Context, t *post.Tag) error {
	return s.repo.Create(ctx, t)
}

func (s *tagServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *tagServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*post.Tag, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *tagServiceImpl) GetBySlug(ctx context.Context, slug string) (*post.Tag, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *tagServiceImpl) GetByName(ctx context.Context, name string) (*post.Tag, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *tagServiceImpl) GetAll(ctx context.Context) ([]*post.Tag, error) {
	return s.repo.GetAll(ctx)
}

// SeriesService 系列服务接口
type SeriesService interface {
	Create(ctx context.Context, s *post.Series) error
	Update(ctx context.Context, s *post.Series) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*post.Series, error)
	GetBySlug(ctx context.Context, slug string) (*post.Series, error)
	GetByAuthor(ctx context.Context, authorID uuid.UUID) ([]*post.Series, error)
}

type seriesServiceImpl struct {
	repo post.SeriesRepository
}

func NewSeriesService(repo post.SeriesRepository) SeriesService {
	return &seriesServiceImpl{repo: repo}
}

func (s *seriesServiceImpl) Create(ctx context.Context, series *post.Series) error {
	return s.repo.Create(ctx, series)
}

func (s *seriesServiceImpl) Update(ctx context.Context, series *post.Series) error {
	return s.repo.Update(ctx, series)
}

func (s *seriesServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *seriesServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*post.Series, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *seriesServiceImpl) GetBySlug(ctx context.Context, slug string) (*post.Series, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *seriesServiceImpl) GetByAuthor(ctx context.Context, authorID uuid.UUID) ([]*post.Series, error) {
	return s.repo.GetByAuthor(ctx, authorID)
}
