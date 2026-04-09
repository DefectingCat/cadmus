<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# Docs

## 用途
`docs/` 目录存放项目的所有文档，包括设计文档、分析报告和代码规范。这些文档为 AI Agent 和开发者提供完整的项目理解和开发指南。

## 关键文件

| 文件 | 描述 |
|------|------|
| `design.md` | 完整设计文档：系统架构、技术栈、模块设计、API 规范、数据库结构、Docker 部署方案 |
| `analysis-architecture.md` | 架构分析报告：分层架构、模块划分、依赖注入、Service Container、配置管理 |
| `analysis-api.md` | API 层设计分析：路由组织、Handler 实现模式、中间件设计、请求响应处理 |
| `analysis-auth.md` | 认证与权限系统分析：JWT 实现、RBAC 模型、权限缓存、中间件流程 |
| `analysis-data.md` | 数据层分析报告：数据模型、数据库迁移、Repository 实现、UUID 主键使用 |
| `analysis-services.md` | 服务层分析报告：服务容器模式、业务逻辑、错误处理、服务间依赖协作 |
| `analysis-frontend.md` | 前端构建系统分析：esbuild + Tailwind、templ 模板、主题引擎、插件系统 |
| `go-comment-style.md` | Go 代码注释规范：文件头注释、函数注释、步骤注释、最佳实践 |

## 无子目录

所有文档均在 `docs/` 根目录下，无子目录结构。

## For AI Agents

### 工作指南

| 场景 | 参考文档 |
|------|----------|
| 理解项目整体架构 | `analysis-architecture.md` + `design.md` |
| 开发 API Handler | `analysis-api.md` |
| 实现认证/权限功能 | `analysis-auth.md` |
| 数据层开发 | `analysis-data.md` |
| 业务逻辑开发 | `analysis-services.md` |
| 前端/模板开发 | `analysis-frontend.md` |
| 编写代码注释 | `go-comment-style.md` |
| 设计决策参考 | `design.md` |

### 文档更新原则

1. 修改代码时同步更新相关文档
2. 新增功能时在 `design.md` 中补充设计
3. 重大重构时更新 `analysis-*.md` 分析报告
4. 保持文档与代码的一致性
