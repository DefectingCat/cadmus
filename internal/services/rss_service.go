// Package services 提供 RSS 订阅服务的实现。
//
// 该文件包含 RSS 订阅源生成相关的核心逻辑，包括：
//   - RSS 2.0 格式 XML 生成
//   - 全站订阅源生成
//   - 分类订阅源生成
//   - 批量查询优化（避免 N+1 问题）
//
// 主要用途：
//
//	用于生成标准的 RSS 订阅源，供用户订阅博客更新。
//
// 设计特点：
//   - 符合 RSS 2.0 规范
//   - 支持分类过滤订阅
//   - 批量查询分类信息优化性能
//
// 作者：xfy
package services

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/google/uuid"
	"rua.plus/cadmus/internal/core/post"
	"rua.plus/cadmus/internal/core/rss"
	"rua.plus/cadmus/internal/database"
)

// RSSService RSS 订阅服务接口。
//
// 该接口定义了 RSS 订阅源生成的操作，符合 RSS 2.0 规范。
type RSSService interface {
	// GenerateFeed 生成 RSS 订阅源 XML。
	//
	// 参数：
	//   - ctx: 上下文
	//   - config: RSS 配置（标题、链接、描述等）
	//   - categorySlug: 分类 Slug（可选，为空则生成全站订阅）
	//
	// 返回值：
	//   - string: 格式化的 RSS XML 字符串
	//   - error: 生成失败时返回错误
	GenerateFeed(ctx context.Context, config rss.FeedConfig, categorySlug string) (string, error)

	// GenerateFeedForCategory 生成指定分类的 RSS 订阅源。
	//
	// 参数：
	//   - ctx: 上下文
	//   - config: RSS 配置
	//   - categoryID: 分类 ID（字符串格式）
	//
	// 返回值：
	//   - string: 分类专属的 RSS XML
	//   - error: 分类不存在时返回全站订阅
	GenerateFeedForCategory(ctx context.Context, config rss.FeedConfig, categoryID string) (string, error)
}

// rssServiceImpl RSS 服务的具体实现。
type rssServiceImpl struct {
	// postRepo 文章数据仓库
	postRepo post.PostRepository

	// categoryRepo 分类数据仓库
	categoryRepo post.CategoryRepository
}

// NewRSSService 创建 RSS 服务实例。
//
// 参数：
//   - postRepo: 文章数据仓库
//   - categoryRepo: 分类数据仓库
//
// 返回值：
//   - RSSService: RSS 服务实例
func NewRSSService(
	postRepo post.PostRepository,
	categoryRepo post.CategoryRepository,
) RSSService {
	return &rssServiceImpl{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
	}
}

// GenerateFeed 生成完整的 RSS 订阅源 XML
func (s *rssServiceImpl) GenerateFeed(ctx context.Context, config rss.FeedConfig, categorySlug string) (string, error) {
	// 获取已发布的文章列表
	filters := post.PostListFilters{
		Status: post.StatusPublished,
	}

	// 如果指定了分类，添加分类筛选
	if categorySlug != "" {
		category, err := s.categoryRepo.GetBySlug(ctx, categorySlug)
		if err == nil && category != nil {
			filters.CategoryID = category.ID
		}
	}

	// 获取最近 20 篇文章
	posts, _, err := s.postRepo.List(ctx, filters, 0, 20)
	if err != nil {
		return "", err
	}

	// 构建 RSS Feed
	now := time.Now()
	feed := &rss.RSSFeed{
		Version: "2.0",
		Channel: &rss.RSSChannel{
			Title:         config.Title,
			Link:          config.Link,
			Description:   config.Description,
			Language:      config.Language,
			LastBuildDate: rss.FormatRFC822Time(now),
			Generator:     config.Generator,
			Items:         make([]*rss.RSSItem, 0, len(posts)),
		},
	}

	// 添加文章项
	// 收集所有分类 ID 用于批量查询
	categoryIDs := make([]uuid.UUID, 0)
	for _, p := range posts {
		if p.CategoryID != uuid.Nil {
			categoryIDs = append(categoryIDs, p.CategoryID)
		}
	}

	// 批量查询分类（避免 N+1）
	categoriesMap := make(map[uuid.UUID]*post.Category)
	dbCategoryRepo, ok := s.categoryRepo.(*database.CategoryRepository)
	if ok && len(categoryIDs) > 0 {
		categoriesMap, _ = dbCategoryRepo.GetByIDs(ctx, categoryIDs)
	}

	for _, p := range posts {
		// 获取发布时间
		pubTime := p.CreatedAt
		if p.PublishAt != nil {
			pubTime = *p.PublishAt
		}

		item := &rss.RSSItem{
			Title:       p.Title,
			Link:        config.BaseURL + "/" + p.Slug,
			Description: p.Excerpt,
			PubDate:     rss.FormatRFC822Time(pubTime),
			GUID:        p.Slug,
		}

		// 如果有分类，添加分类信息（使用预查询的结果）
		if p.CategoryID != uuid.Nil {
			category := categoriesMap[p.CategoryID]
			if category != nil {
				item.Category = category.Name
			}
		}

		feed.Channel.Items = append(feed.Channel.Items, item)
	}

	// 序列化为 XML
	xmlData, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", err
	}

	return xml.Header + string(xmlData), nil
}

// GenerateFeedForCategory 生成指定分类的 RSS 订阅源
func (s *rssServiceImpl) GenerateFeedForCategory(ctx context.Context, config rss.FeedConfig, categoryID string) (string, error) {
	// 解析分类 ID
	id, err := uuid.Parse(categoryID)
	if err != nil {
		return s.GenerateFeed(ctx, config, "")
	}

	// 获取分类信息
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return s.GenerateFeed(ctx, config, "")
	}

	// 更新配置以反映分类
	categoryConfig := config
	if category != nil {
		categoryConfig.Title = config.Title + " - " + category.Name
		categoryConfig.Description = category.Description
	}

	return s.GenerateFeed(ctx, categoryConfig, category.Slug)
}