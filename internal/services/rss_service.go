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

// RSSService RSS 订阅服务接口
type RSSService interface {
	// GenerateFeed 生成 RSS 订阅源 XML
	GenerateFeed(ctx context.Context, config rss.FeedConfig, categorySlug string) (string, error)

	// GenerateFeedForCategory 生成指定分类的 RSS 订阅源
	GenerateFeedForCategory(ctx context.Context, config rss.FeedConfig, categoryID string) (string, error)
}

// rssServiceImpl RSS 服务实现
type rssServiceImpl struct {
	postRepo     post.PostRepository
	categoryRepo post.CategoryRepository
}

// NewRSSService 创建 RSS 服务
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