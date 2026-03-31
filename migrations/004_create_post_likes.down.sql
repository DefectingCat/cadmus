-- 004_create_post_likes.down.sql
-- 删除文章点赞表

DROP INDEX IF EXISTS idx_post_likes_user;
DROP INDEX IF EXISTS idx_post_likes_post;
DROP TABLE IF EXISTS post_likes;