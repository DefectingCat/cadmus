-- 004_create_media.up.sql
-- 创建媒体文件表

CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,           -- 存储文件名（唯一）
    original_name VARCHAR(255) NOT NULL,      -- 原始文件名
    filepath TEXT NOT NULL,                   -- 文件存储路径
    url TEXT NOT NULL,                        -- 访问 URL
    mime_type VARCHAR(100) NOT NULL,          -- MIME 类型
    size BIGINT NOT NULL,                     -- 文件大小（字节）
    width INT,                                -- 图片宽度（可选）
    height INT,                               -- 图片高度（可选）
    alt_text TEXT,                            -- 替代文本（可选）
    metadata JSONB,                           -- 其他元数据
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_media_uploader ON media(uploader_id);
CREATE INDEX idx_media_mime ON media(mime_type);
CREATE INDEX idx_media_created ON media(created_at DESC);