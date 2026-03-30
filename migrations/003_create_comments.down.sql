-- 003_create_comments.down.sql
-- 回滚评论相关表结构

-- 删除 comment_likes 表
DROP TABLE IF EXISTS comment_likes;

-- 删除评论深度触发器和函数
DROP TRIGGER IF EXISTS trg_comment_depth ON comments;
DROP FUNCTION IF EXISTS calculate_comment_depth();

-- 删除 comments 表
DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;
DROP TABLE IF EXISTS comments;