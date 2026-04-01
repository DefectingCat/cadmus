<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-01 | Updated: 2026-04-01 -->

# styles

## Purpose
前端样式文件，包含管理后台和前台主题样式。

## Key Files
| File | Description |
|------|-------------|
| `admin.css` | 后台管理样式：CSS 变量、布局、组件、暗色主题 |
| `main.css` | 前台基础样式：Tailwind 导入 |

## For AI Agents

### Working In This Directory
- 使用 CSS 变量实现主题系统
- `admin.css` 包含完整的设计系统
- `main.css` 仅导入 Tailwind，前台样式由主题提供

### Design System (admin.css)

**CSS 变量分类**:
| Category | Variables |
|----------|-----------|
| Colors | `--bg-*`, `--accent-*`, `--text-*` |
| Spacing | `--space-xs` to `--space-2xl` |
| Shadows | `--shadow-xs` to `--shadow-xl` |
| Radius | `--radius-sm` to `--radius-xl` |
| Typography | `--font-display`, `--font-body` |
| Transitions | `--transition-fast/base/slow` |

**主题切换**:
```css
:root { /* Light theme */ }
[data-theme="dark"] { /* Dark theme */ }
```

### Component Classes
| Component | Classes |
|-----------|---------|
| Layout | `.admin-layout`, `.admin-sidebar`, `.admin-main`, `.admin-topbar` |
| Cards | `.card`, `.stat-card` |
| Tables | `.data-table` |
| Buttons | `.btn-primary`, `.btn-secondary`, `.btn-ghost`, `.btn-danger` |
| Forms | `.form-input`, `.form-select`, `.search-bar` |
| Status | `.status-badge.published/draft/scheduled/private` |

### Responsive Breakpoints
- `1280px`: Sidebar 缩小到 240px
- `1024px`: Sidebar 隐藏，内容全宽
- `640px`: 单列布局

<!-- MANUAL: -->