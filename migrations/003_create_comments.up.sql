-- 003_create_comments.up.sql
-- 创建评论相关表结构

-- ============================================
-- Comments 表：评论内容
-- ============================================
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    depth INT NOT NULL DEFAULT 0,        -- 嵌套深度（防止无限递归）
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending/approved/spam/deleted
    like_count INT DEFAULT 0,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- 约束：最大嵌套深度为 5
    CONSTRAINT chk_depth CHECK (depth >= 0 AND depth <= 5)
);

CREATE INDEX idx_comments_post ON comments(post_id);
CREATE INDEX idx_comments_user ON comments(user_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);
CREATE INDEX idx_comments_status ON comments(status);
CREATE INDEX idx_comments_depth ON comments(depth);

-- comments 表的 updated_at 触发器
CREATE TRIGGER update_comments_updated_at
    BEFORE UPDATE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 触发器：自动计算评论深度
-- ============================================
CREATE OR REPLACE FUNCTION calculate_comment_depth()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.depth := 0;
    ELSE
        SELECT depth + 1 INTO NEW.depth
        FROM comments
        WHERE id = NEW.parent_id;

        -- 检查是否超过最大深度
        IF NEW.depth > 5 THEN
            RAISE EXCEPTION 'Comment depth exceeds maximum limit of 5';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_comment_depth
    BEFORE INSERT ON comments
    FOR EACH ROW
    EXECUTE FUNCTION calculate_comment_depth();

-- ============================================
-- Comment_Likes 表：评论点赞
-- ============================================
CREATE TABLE comment_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(comment_id, user_id)
);

CREATE INDEX idx_comment_likes_comment ON comment_likes(comment_id);
CREATE INDEX idx_comment_likes_user ON comment_likes(user_id);