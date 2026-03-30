-- 002_create_posts.down.sql
-- 回滚文章相关表结构

-- 删除 post_versions 表
DROP TABLE IF EXISTS post_versions;

-- 删除 post_tags 表
DROP TABLE IF EXISTS post_tags;

-- 删除全文搜索相关
DROP TRIGGER IF EXISTS update_posts_search_vector ON posts;
DROP FUNCTION IF EXISTS update_posts_search_vector();
DROP INDEX IF EXISTS posts_search_idx;
ALTER TABLE posts DROP COLUMN IF EXISTS search_vector;

-- 删除 posts 表
DROP TRIGGER IF EXISTS update_posts_updated_at ON posts;
DROP TABLE IF EXISTS posts;

-- 删除 series 表
DROP TRIGGER IF EXISTS update_series_updated_at ON series;
DROP TABLE IF EXISTS series;

-- 删除 tags 表
DROP TABLE IF EXISTS tags;

-- 删除 categories 表
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TABLE IF EXISTS categories;