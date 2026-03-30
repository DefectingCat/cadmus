-- 001_init.up.sql
-- 初始化数据库表结构

-- 启用 UUID 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建 updated_at 自动更新函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ============================================
-- Permissions 表：细粒度权限定义
-- ============================================
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_permissions_category ON permissions(category);
CREATE INDEX idx_permissions_name ON permissions(name);

-- ============================================
-- Roles 表：用户角色定义
-- ============================================
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_is_default ON roles(is_default);

-- ============================================
-- Role_Permissions 表：角色-权限关联
-- ============================================
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);

-- ============================================
-- Users 表：用户信息
-- ============================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    bio TEXT,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role_id);
CREATE INDEX idx_users_status ON users(status);

-- 创建 users 表的 updated_at 触发器
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 插入默认权限
-- ============================================
INSERT INTO permissions (name, description, category) VALUES
    -- 文章权限
    ('post.create', '创建文章', 'post'),
    ('post.edit', '编辑文章', 'post'),
    ('post.delete', '删除文章', 'post'),
    ('post.publish', '发布文章', 'post'),
    ('post.view', '查看文章', 'post'),
    -- 评论权限
    ('comment.create', '创建评论', 'comment'),
    ('comment.edit', '编辑评论', 'comment'),
    ('comment.delete', '删除评论', 'comment'),
    ('comment.view', '查看评论', 'comment'),
    -- 用户权限
    ('user.view', '查看用户信息', 'user'),
    ('user.edit', '编辑用户信息', 'user'),
    ('user.delete', '删除用户', 'user'),
    ('user.manage', '管理用户', 'user'),
    -- 主题权限
    ('theme.view', '查看主题', 'theme'),
    ('theme.edit', '编辑主题', 'theme'),
    ('theme.install', '安装主题', 'theme'),
    -- 插件权限
    ('plugin.view', '查看插件', 'plugin'),
    ('plugin.edit', '编辑插件', 'plugin'),
    ('plugin.install', '安装插件', 'plugin'),
    -- 系统权限
    ('system.admin', '系统管理员', 'system'),
    ('system.settings', '系统设置', 'system');

-- ============================================
-- 插入默认角色
-- ============================================
INSERT INTO roles (name, display_name, is_default) VALUES
    ('admin', '管理员', FALSE),
    ('editor', '编辑', FALSE),
    ('author', '作者', FALSE),
    ('user', '普通用户', TRUE),
    ('guest', '访客', FALSE);

-- ============================================
-- 配置默认角色权限
-- ============================================
-- admin: 所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin';

-- editor: 文章和评论相关权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'editor'
    AND p.category IN ('post', 'comment');

-- author: 创建和编辑自己的内容
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'author'
    AND p.name IN ('post.create', 'post.edit', 'post.view', 'comment.create', 'comment.edit', 'comment.view', 'user.view', 'user.edit');

-- user: 基础权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'user'
    AND p.name IN ('post.view', 'comment.view', 'comment.create', 'user.view', 'user.edit');

-- guest: 只读权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'guest'
    AND p.name IN ('post.view', 'comment.view', 'user.view');