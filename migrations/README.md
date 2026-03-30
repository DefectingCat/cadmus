# 数据库迁移目录

此目录存放 PostgreSQL 数据库迁移脚本。

## 迁移文件命名

使用数字前缀命名迁移文件：
```
001_init_schema.up.sql
001_init_schema.down.sql
002_add_user_table.up.sql
002_add_user_table.down.sql
```