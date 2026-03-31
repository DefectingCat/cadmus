<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# migrations

## Purpose
PostgreSQL 数据库迁移脚本目录，管理数据库版本演进。

## Key Files
| File | Description |
|------|-------------|
| `001_init.up.sql` | 初始化：用户、角色、权限表 |
| `001_init.down.sql` | 初始化回滚脚本 |
| `002_create_posts.up.sql` | 文章、分类、标签、系列表 |
| `003_create_comments.up.sql` | 评论、评论点赞表 |
| `004_create_media.up.sql` | 媒体文件表 |
| `004_create_post_likes.up.sql` | 文章点赞表 |
| `README.md` | 迁移说明文档 |

## For AI Agents

### Working In This Directory
- 迁移文件命名格式：`{序号}_{描述}.up.sql` / `.down.sql`
- 新增表需创建对应迁移文件
- 迁移脚本需同时提供 up 和 down 版本

### Migration Order
| Order | Tables Created |
|-------|----------------|
| 001 | users, roles, permissions, role_permissions, user_roles |
| 002 | categories, tags, posts, post_tags, post_versions, series |
| 003 | comments, comment_likes |
| 004 | media, post_likes |

### Running Migrations
```bash
# 使用 migrate 工具
migrate -path ./migrations -database "postgres://cadmus:@localhost:5432/cadmus?sslmode=disable" up

# 或使用 goose
goose postgres "host=localhost port=5432 user=cadmus dbname=cadmus sslmode=disable" up
```

<!-- MANUAL: -->