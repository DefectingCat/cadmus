<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# Cadmus Default 主题 - AI Agent 参考

## 主题概述

default 目录是 Cadmus 博客平台的默认主题，使用 Tailwind CSS 提供现代化、响应式的用户界面。

### 主题元信息

| 属性 | 值 |
|------|-----|
| **主题 ID** | `default` |
| **名称** | Cadmus Default |
| **版本** | 1.0.0 |
| **作者** | Cadmus Team |
| **样式框架** | Tailwind CSS |
| **模板引擎** | templ (类型安全) |

## 关键文件

### theme.go

主题核心实现文件，提供：

- `DefaultTheme` 结构体：实现 `theme.ThemeComponents` 接口
- `New()`：创建主题实例，返回完整的 `Theme` 对象
- 页面组件渲染方法：

| 方法 | 功能 | 状态 |
|------|------|------|
| `Layout()` | 整体布局框架 | 已实现 |
| `Header()` | 页面头部导航 | 已实现 |
| `Footer()` | 页面底部区域 | 已实现 |
| `PostList()` | 文章列表页 | 已实现（占位） |
| `PostDetail()` | 文章详情页 | 已实现（占位） |
| `CategoryPage()` | 分类页 | 未实现（返回 nil） |
| `TagPage()` | 标签页 | 未实现（返回 nil） |
| `Sidebar()` | 侧边栏 | 未实现（返回 nil） |
| `ErrorPage()` | 错误页面（404/500） | 已实现 |

### theme.yaml

主题配置文件，定义主题元数据：

```yaml
name: default
version: 1.0.0
description: Cadmus 默认主题
author: Cadmus Team
```

### README.md

主题说明文档，提供基本结构概览。

## 目录结构

```
default/
├── assets/       # 主题静态资源（CSS、JS、图片等）
├── templates/    # templ 模板覆盖目录
├── theme.go      # 主题组件实现
├── theme.yaml    # 主题配置
└── README.md     # 主题说明
```

### 子目录

#### assets/

存放主题专属静态资源：
- CSS 样式文件
- JavaScript 脚本
- 图片资源
- 字体文件

## 主题注册机制

主题通过 `init()` 函数自动注册到全局主题注册表：

```go
func init() {
    theme.Register(New())
}
```

启用方式（blank import）：

```go
import _ "rua.plus/cadmus/themes/default"
```

## AI Agent 主题开发指南

### 添加新页面组件

1. 在 `theme.go` 中实现对应方法：
```go
func (t *DefaultTheme) NewPage() templ.Component {
    return pages.NewPageComponent(...)
}
```

2. 确保返回类型满足 `templ.Component` 接口

3. 如需自定义模板，在 `templates/` 目录添加覆盖文件

### 修改样式

1. 主题使用 Tailwind CSS，直接修改组件中的 class 名称
2. 新增静态资源放入 `assets/` 目录
3. 更新 `theme.yaml` 中的版本号

### 调试提示

- 组件返回 `nil` 表示该功能暂未实现
- 错误页面使用内联 HTML，不依赖 templ 模板
- 布局组件统一使用 `web/templates/layouts` 包

## 相关文档

- [主题系统架构](../../internal/theme/AGENTS.md)
- [模板组件库](../../web/templates/AGENTS.md)
