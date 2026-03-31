<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# docs

## Purpose
项目文档目录，包含设计文档和架构说明。

## Key Files
| File | Description |
|------|-------------|
| `design.md` | 完整设计文档：架构、数据模型、API、部署方案 |

## For AI Agents

### Working In This Directory
- `design.md` 是权威设计文档，实现前应参考此文件
- 设计变更需同步更新文档

### Document Structure
| Section | Content |
|---------|---------|
| 技术栈总览 | Go + templ + esbuild + PostgreSQL + Redis |
| 用户权限系统 | User/Role/Permission 模型 + Redis 缓存 |
| 文章系统 | Post 模型 + 版本历史 + SEO 元数据 |
| 块编辑器 | BlockDocument 结构 + 插件扩展机制 |
| 评论系统 | 嵌套回复 + 审核流程 + 深度限制 |
| 主题系统 | ThemeComponents 接口 + 主题注册 |
| 插件系统 | 编译时注册模式 + 扩展接口 |
| REST API | 路由结构 + 认证中间件 |
| 缓存策略 | Redis Key 规范 + 穿透/击穿防护 |
| Docker 部署 | 多阶段构建 + secrets 管理 |

<!-- MANUAL: -->