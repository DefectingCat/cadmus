// Package services 提供文章服务的实现。
//
// 该文件包含文章管理相关的核心逻辑，包括：
//   - 文章 CRUD 操作（创建、更新、删除、查询）
//   - 文章发布和定时发布
//   - 文章版本管理和回滚
//   - 文章点赞功能
//   - 分类、标签、系列管理
//
// 主要用途：
//
//	用于处理博客文章的完整生命周期管理，支持多版本历史记录。
//
// 设计特点：
//   - 支持草稿、已发布、定时发布等多种状态
//   - 版本控制支持内容回滚
//   - 点赞功能使用原子操作保证并发安全
//
// 作者：xfy
package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/post"
)

// PostService 文章业务服务接口。
//
// 该接口定义了文章管理的核心操作，包括 CRUD、发布控制、版本管理和点赞功能。
// 所有方法均为并发安全，可在多个 goroutine 中同时调用。
type PostService interface {
	// Create 创建文章，设置标签关联。
	//
	// 参数：
	//   - ctx: 上下文
	//   - p: 文章对象（必须包含 Title 和 Slug）
	//   - tagIDs: 关联的标签 ID 列表
	//
	// 返回值：
	//   - err: 可能的错误包括标题或 Slug 缺失、状态无效
	//
	// 注意事项：
	//   - 新文章默认状态为草稿
	//   - 标签关联在文章创建后设置
	Create(ctx context.Context, p *post.Post, tagIDs []uuid.UUID) error

	// Update 更新文章内容和标签。
	//
	// 参数：
	//   - ctx: 上下文
	//   - p: 文章对象（包含更新后的字段）
	//   - tagIDs: 新的标签 ID 列表（会替换原有标签）
	//
	// 返回值：
	//   - err: 更新失败时返回错误
	//
	// 注意事项：
	//   - 标签更新采用先删后加策略
	Update(ctx context.Context, p *post.Post, tagIDs []uuid.UUID) error

	// Delete 删除文章。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 文章唯一标识符
	//
	// 返回值：
	//   - err: 删除失败时返回错误
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取文章详情。
	GetByID(ctx context.Context, id uuid.UUID) (*post.Post, error)

	// GetBySlug 根据 URL 友好名称获取文章。
	//
	// 参数：
	//   - slug: 文章的 URL 标识符
	//
	// 返回值：
	//   - post: 文章对象
	//   - err: 文章不存在时返回错误
	GetBySlug(ctx context.Context, slug string) (*post.Post, error)

	// List 分页获取文章列表。
	//
	// 参数：
	//   - ctx: 上下文
	//   - filters: 筛选条件（状态、分类、作者等）
	//   - page: 页码（从 1 开始）
	//   - pageSize: 每页数量（最大 100）
	//
	// 返回值：
	//   - posts: 文章列表
	//   - total: 符合条件的总数
	//   - err: 查询错误
	List(ctx context.Context, filters post.PostListFilters, page, pageSize int) ([]*post.Post, int, error)

	// GetByAuthor 获取指定作者的文章列表。
	GetByAuthor(ctx context.Context, authorID uuid.UUID, status post.PostStatus, page, pageSize int) ([]*post.Post, int, error)

	// Search 全文搜索文章内容。
	//
	// 参数：
	//   - query: 搜索关键词
	//   - page, pageSize: 分页参数
	//
	// 返回值：
	//   - posts: 匹配的文章列表
	//   - total: 匹配总数
	//   - err: 搜索错误
	Search(ctx context.Context, query string, page, pageSize int) ([]*post.Post, int, error)

	// Publish 立即发布文章。
	//
	// 将文章状态改为已发布，并设置发布时间。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 文章 ID
	//
	// 返回值：
	//   - err: 文章不存在或更新失败
	Publish(ctx context.Context, id uuid.UUID) error

	// Schedule 设置定时发布。
	//
	// 将文章状态改为定时发布，到达指定时间后自动发布。
	//
	// 参数：
	//   - ctx: 上下文
	//   - id: 文章 ID
	//   - publishAt: 计划发布时间
	//
	// 返回值：
	//   - err: 设置失败
	Schedule(ctx context.Context, id uuid.UUID, publishAt time.Time) error

	// CreateVersion 创建文章版本快照。
	//
	// 保存当前文章内容的版本记录，用于历史追踪和回滚。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//   - note: 版本备注（如 "修复错别字"）
	//   - creatorID: 创建版本的用户 ID
	CreateVersion(ctx context.Context, postID uuid.UUID, note string, creatorID uuid.UUID) error

	// GetVersions 获取文章的版本历史列表。
	GetVersions(ctx context.Context, postID uuid.UUID) ([]*post.PostVersion, error)

	// Rollback 回滚到指定版本。
	//
	// 将文章内容恢复到历史版本的内容。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//   - version: 目标版本号
	//
	// 返回值：
	//   - err: 版本不存在或回滚失败
	Rollback(ctx context.Context, postID uuid.UUID, version int) error

	// IncrementViewCount 增加文章浏览量。
	//
	// 每次访问文章详情页时调用。
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	// LikePost 点赞文章。
	//
	// 使用原子操作保证并发安全，同一用户不可重复点赞。
	//
	// 参数：
	//   - ctx: 上下文
	//   - postID: 文章 ID
	//   - userID: 用户 ID
	//
	// 返回值：
	//   - err: 可能的错误包括文章不存在、已点赞
	LikePost(ctx context.Context, postID, userID uuid.UUID) error

	// UnlikePost 取消点赞文章。
	UnlikePost(ctx context.Context, postID, userID uuid.UUID) error

	// IsPostLiked 检查用户是否已点赞文章。
	IsPostLiked(ctx context.Context, postID, userID uuid.UUID) (bool, error)
}

// postServiceImpl 文章服务的具体实现。
//
// 该结构体实现了 PostService 接口，依赖多个 Repository 进行数据操作。
// 可选支持点赞功能，通过 likeRepo 字段配置。
type postServiceImpl struct {
	// postRepo 文章数据仓库
	postRepo post.PostRepository

	// categoryRepo 分类数据仓库
	categoryRepo post.CategoryRepository

	// tagRepo 标签数据仓库
	tagRepo post.TagRepository

	// seriesRepo 系列数据仓库
	seriesRepo post.SeriesRepository

	// likeRepo 文章点赞数据仓库（可选）
	likeRepo post.PostLikeRepository
}

// NewPostService 创建基础文章服务。
//
// 该函数创建不包含点赞功能的文章服务，适用于简单场景。
//
// 参数：
//   - postRepo: 文章数据仓库
//   - categoryRepo: 分类数据仓库
//   - tagRepo: 标签数据仓库
//   - seriesRepo: 系列数据仓库
//
// 返回值：
//   - PostService: 文章服务实例
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

// NewPostServiceWithLikes 创建带点赞功能的文章服务。
//
// 该函数在基础服务上增加点赞功能，适用于需要用户互动的场景。
//
// 参数：
//   - 各仓库参数同 NewPostService
//   - likeRepo: 文章点赞数据仓库
//
// 返回值：
//   - PostService: 包含点赞功能的文章服务
func NewPostServiceWithLikes(
	postRepo post.PostRepository,
	categoryRepo post.CategoryRepository,
	tagRepo post.TagRepository,
	seriesRepo post.SeriesRepository,
	likeRepo post.PostLikeRepository,
) PostService {
	return &postServiceImpl{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		seriesRepo:   seriesRepo,
		likeRepo:     likeRepo,
	}
}

// Create 创建文章，设置标签关联。
//
// 该方法执行以下步骤：
//  1. 验证必填字段（标题、Slug）
//  2. 验证状态有效性
//  3. 设置默认状态为草稿
//  4. 创建文章记录
//  5. 关联标签
func (s *postServiceImpl) Create(ctx context.Context, p *post.Post, tagIDs []uuid.UUID) error {
	// 验证必填字段
	if p.Title == "" || p.Slug == "" {
		return errors.New("title and slug are required")
	}

	// 验证状态有效性
	if p.Status != "" && !p.Status.IsValid() {
		return post.ErrInvalidStatus
	}

	// 默认状态为草稿，新文章需要显式发布
	if p.Status == "" {
		p.Status = post.StatusDraft
	}

	// 创建文章记录
	if err := s.postRepo.Create(ctx, p); err != nil {
		return err
	}

	// 设置标签关联（文章创建后添加标签）
	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			if err := s.tagRepo.AddPostTag(ctx, p.ID, tagID); err != nil {
				return err
			}
		}
	}

	return nil
}

// Update 更新文章内容和标签。
//
// 标签更新采用替换策略：先删除所有旧标签，再添加新标签。
//
// 参数：
//   - ctx: 上下文
//   - p: 文章对象
//   - tagIDs: 新的标签列表（替换原有标签）
func (s *postServiceImpl) Update(ctx context.Context, p *post.Post, tagIDs []uuid.UUID) error {
	// 验证状态有效性
	if p.Status != "" && !p.Status.IsValid() {
		return post.ErrInvalidStatus
	}

	// 更新文章基本信息
	if err := s.postRepo.Update(ctx, p); err != nil {
		return err
	}

	// 更新标签：采用替换策略（先删后加）
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

// Delete 删除文章。
func (s *postServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return s.postRepo.Delete(ctx, id)
}

// GetByID 根据 ID 获取文章详情。
func (s *postServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	return s.postRepo.GetByID(ctx, id)
}

// GetBySlug 根据 URL 友好名称获取文章。
func (s *postServiceImpl) GetBySlug(ctx context.Context, slug string) (*post.Post, error) {
	return s.postRepo.GetBySlug(ctx, slug)
}

// List 分页获取文章列表。
//
// 支持多条件筛选，自动校正分页参数。
func (s *postServiceImpl) List(ctx context.Context, filters post.PostListFilters, page, pageSize int) ([]*post.Post, int, error) {
	// 分页参数校正
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.postRepo.List(ctx, filters, offset, pageSize)
}

// GetByAuthor 获取指定作者的文章列表。
func (s *postServiceImpl) GetByAuthor(ctx context.Context, authorID uuid.UUID, status post.PostStatus, page, pageSize int) ([]*post.Post, int, error) {
	filters := post.PostListFilters{
		AuthorID: authorID,
		Status:   status,
	}
	return s.List(ctx, filters, page, pageSize)
}

// Search 全文搜索文章。
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

// Publish 立即发布文章。
//
// 将状态改为已发布并设置发布时间为当前时间。
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

// Schedule 设置定时发布。
//
// 将状态改为定时发布，需配合定时任务使用。
func (s *postServiceImpl) Schedule(ctx context.Context, id uuid.UUID, publishAt time.Time) error {
	p, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	p.Status = post.StatusScheduled
	p.PublishAt = &publishAt

	return s.postRepo.Update(ctx, p)
}

// CreateVersion 创建文章版本快照。
//
// 保存当前内容用于历史追踪和回滚。
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

// GetVersions 获取文章的版本历史列表。
func (s *postServiceImpl) GetVersions(ctx context.Context, postID uuid.UUID) ([]*post.PostVersion, error) {
	return s.postRepo.GetVersions(ctx, postID)
}

// Rollback 回滚到指定版本。
//
// 将文章内容恢复到历史版本，不改变其他字段。
func (s *postServiceImpl) Rollback(ctx context.Context, postID uuid.UUID, version int) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// 获取目标版本内容
	targetVersion, err := s.postRepo.GetVersionByNumber(ctx, postID, version)
	if err != nil {
		return err
	}

	// 恢复内容到当前文章
	p.Content = targetVersion.Content
	return s.postRepo.Update(ctx, p)
}

// IncrementViewCount 增加文章浏览量。
//
// 每次访问文章详情页时调用，不影响其他字段。
func (s *postServiceImpl) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return s.postRepo.IncrementViewCount(ctx, id)
}

// LikePost 点赞文章。
//
// 使用原子操作保证并发安全，防止重复点赞。
//
// 返回值：
//   - err: 可能的错误包括文章不存在、已点赞
func (s *postServiceImpl) LikePost(ctx context.Context, postID, userID uuid.UUID) error {
	// 检查文章是否存在
	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return post.ErrPostNotFound
	}

	// 使用原子操作创建点赞记录并更新计数
	created, err := s.likeRepo.CreateIfNotExists(ctx, postID, userID)
	if err != nil {
		return err
	}
	if !created {
		return post.ErrAlreadyLiked
	}

	return nil
}

// UnlikePost 取消点赞文章。
//
// 使用原子操作保证并发安全。
func (s *postServiceImpl) UnlikePost(ctx context.Context, postID, userID uuid.UUID) error {
	// 检查文章是否存在
	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return post.ErrPostNotFound
	}

	// 使用原子操作删除点赞记录并更新计数
	deleted, err := s.likeRepo.DeleteIfExists(ctx, postID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return post.ErrNotLiked
	}

	return nil
}

// IsPostLiked 检查用户是否已点赞文章。
func (s *postServiceImpl) IsPostLiked(ctx context.Context, postID, userID uuid.UUID) (bool, error) {
	return s.likeRepo.Exists(ctx, postID, userID)
}

// CategoryService 分类服务接口。
//
// 提供分类的 CRUD 操作和层级关系查询。
type CategoryService interface {
	// Create 创建分类
	Create(ctx context.Context, c *post.Category) error
	// Update 更新分类
	Update(ctx context.Context, c *post.Category) error
	// Delete 删除分类
	Delete(ctx context.Context, id uuid.UUID) error
	// GetByID 根据 ID 获取分类
	GetByID(ctx context.Context, id uuid.UUID) (*post.Category, error)
	// GetBySlug 根据 Slug 获取分类
	GetBySlug(ctx context.Context, slug string) (*post.Category, error)
	// GetAll 获取所有分类
	GetAll(ctx context.Context) ([]*post.Category, error)
	// GetChildren 获取子分类列表
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*post.Category, error)
	// GetPostCount 获取分类下的文章数量
	GetPostCount(ctx context.Context, categoryID uuid.UUID) (int, error)
}

// categoryServiceImpl 分类服务的具体实现。
type categoryServiceImpl struct {
	repo post.CategoryRepository
}

// NewCategoryService 创建分类服务实例。
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

func (s *categoryServiceImpl) GetPostCount(ctx context.Context, categoryID uuid.UUID) (int, error) {
	return s.repo.GetPostCount(ctx, categoryID)
}

// TagService 标签服务接口。
//
// 提供标签的 CRUD 操作和文章计数查询。
type TagService interface {
	// Create 创建标签
	Create(ctx context.Context, t *post.Tag) error
	// Delete 删除标签
	Delete(ctx context.Context, id uuid.UUID) error
	// GetByID 根据 ID 获取标签
	GetByID(ctx context.Context, id uuid.UUID) (*post.Tag, error)
	// GetBySlug 根据 Slug 获取标签
	GetBySlug(ctx context.Context, slug string) (*post.Tag, error)
	// GetByName 根据名称获取标签
	GetByName(ctx context.Context, name string) (*post.Tag, error)
	// GetAll 获取所有标签
	GetAll(ctx context.Context) ([]*post.Tag, error)
	// GetPostCount 获取标签下的文章数量
	GetPostCount(ctx context.Context, tagID uuid.UUID) (int, error)
}

// tagServiceImpl 标签服务的具体实现。
type tagServiceImpl struct {
	repo post.TagRepository
}

// NewTagService 创建标签服务实例。
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

func (s *tagServiceImpl) GetPostCount(ctx context.Context, tagID uuid.UUID) (int, error) {
	return s.repo.GetPostCount(ctx, tagID)
}

// SeriesService 系列服务接口。
//
// 系列用于将相关文章组织在一起，形成连载或专题。
type SeriesService interface {
	// Create 创建系列
	Create(ctx context.Context, s *post.Series) error
	// Update 更新系列
	Update(ctx context.Context, s *post.Series) error
	// Delete 删除系列
	Delete(ctx context.Context, id uuid.UUID) error
	// GetByID 根据 ID 获取系列
	GetByID(ctx context.Context, id uuid.UUID) (*post.Series, error)
	// GetBySlug 根据 Slug 获取系列
	GetBySlug(ctx context.Context, slug string) (*post.Series, error)
	// GetByAuthor 获取作者的系列列表
	GetByAuthor(ctx context.Context, authorID uuid.UUID) ([]*post.Series, error)
}

// seriesServiceImpl 系列服务的具体实现。
type seriesServiceImpl struct {
	repo post.SeriesRepository
}

// NewSeriesService 创建系列服务实例。
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
