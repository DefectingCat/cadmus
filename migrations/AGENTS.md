<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# 数据库迁移目录

## 用途

此目录存放 PostgreSQL 数据库迁移脚本，用于管理数据库结构的版本控制和演进。每个迁移文件包含向上迁移（`.up.sql`）和向下回滚（`.down.sql`）两个脚本。

## 迁移文件列表

| 编号 | 文件名 | 描述 |
|------|--------|------|
| 001 | `001_init` | 初始化数据库：启用 UUID 扩展、创建通用函数、权限系统（permissions/roles/role_permissions）、用户表（users）及默认角色 |
| 002 | `002_create_posts` | 文章系统：分类（categories）、标签（tags）、文章系列（series）、文章表（posts）、文章标签关联（post_tags）、版本历史（post_versions）、全文搜索支持 |
| 003 | `003_create_comments` | 评论系统：评论表（comments，支持嵌套回复）、评论点赞（comment_likes）、自动计算评论深度的触发器 |
| 004 | `004_create_post_likes` | 文章点赞表（post_likes），记录用户对文章的点赞 |
| 005 | `005_create_media` | 媒体库：媒体文件表（media），存储图片/文件等上传资源 |

## 子目录

无子目录，所有迁移文件均位于此目录。

## AI Agent 迁移执行指南

### 执行迁移

使用 `psql` 按顺序执行 `.up.sql` 文件：

```bash
# 设置数据库连接
export DATABASE_URL="postgresql://user:password@localhost:5432/cadmus"

# 按顺序执行迁移
psql $DATABASE_URL -f migrations/001_init.up.sql
psql $DATABASE_URL -f migrations/002_create_posts.up.sql
psql $DATABASE_URL -f migrations/003_create_comments.up.sql
psql $DATABASE_URL -f migrations/004_create_post_likes.up.sql
psql $DATABASE_URL -f migrations/005_create_media.up.sql
```

### 回滚迁移

如需回滚，执行对应的 `.down.sql` 文件（从最新到最旧）：

```bash
psql $DATABASE_URL -f migrations/005_create_media.down.sql
psql $DATABASE_URL -f migrations/004_create_post_likes.down.sql
psql $DATABASE_URL -f migrations/003_create_comments.down.sql
psql $DATABASE_URL -f migrations/002_create_posts.down.sql
psql $DATABASE_URL -f migrations/001_init.down.sql
```

### 注意事项

- 迁移必须按数字顺序执行（001 → 005）
- 迁移 001 创建基础架构（UUID 扩展、函数、权限系统），必须先执行
- 迁移 002 依赖 001（引用 users 表）
- 迁移 003 依赖 002（引用 posts 表）
- 迁移 004 依赖 002（引用 posts 表）
- 迁移 005 依赖 001（引用 users 表）