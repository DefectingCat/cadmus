-- 001_init.down.sql
-- 回滚初始化数据库表结构

-- 删除 users 表
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TABLE IF EXISTS users;

-- 删除 role_permissions 表
DROP TABLE IF EXISTS role_permissions;

-- 删除 roles 表
DROP TABLE IF EXISTS roles;

-- 删除 permissions 表
DROP TABLE IF EXISTS permissions;

-- 删除 updated_at 函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除 UUID 扩展（可选，保留以备后用）
-- DROP EXTENSION IF EXISTS "uuid-ossp";