<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# default

## Purpose
Cadmus 默认主题，使用 Tailwind CSS 样式。

## Key Files
| File | Description |
|------|-------------|
| `theme.go` | 主题实现：ThemeComponents 接口 |
| `theme.yaml` | 主题元信息配置文件 |
| `README.md` | 主题说明文档 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `assets/` | 主题静态资源（图片、字体等） |

## For AI Agents

### Working In This Directory
- 默认主题，通过 blank import 自动注册
- 组件使用 `web/templates/` 下的 templ 组件
- 自定义主题可参考此实现

### Theme Metadata
| Field | Value |
|-------|-------|
| ID | `default` |
| Name | Cadmus Default |
| Version | 1.0.0 |
| Author | Cadmus Team |

### Component Mapping
| Method | Templ Component |
|--------|-----------------|
| `Layout()` | `layouts.BaseLayout` |
| `Header()` | `layouts.Header` |
| `Footer()` | `layouts.Footer` |
| `PostList()` | `pages.PostListPage` |
| `PostDetail()` | `pages.PostDetailPage` |
| `ErrorPage()` | 内置 HTML 模板 |

### Style System
- Tailwind CSS 实用类
- 响应式设计
- 深色模式支持（计划）

<!-- MANUAL: -->