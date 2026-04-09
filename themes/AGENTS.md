<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# themes 目录

## 用途

`themes/` 目录存放 Cadmus 博客平台的主题系统。主题系统允许用户通过切换主题来改变博客的外观和页面渲染方式，而无需修改核心代码。

每个主题是一个独立的 Go 模块，通过 `init()` 函数自动注册到全局主题注册表。

## 目录结构

```
themes/
├── AGENTS.md           # 本文件：主题系统说明
└── default/            # 默认主题实现
    ├── AGENTS.md       # 默认主题开发指南
    ├── README.md       # 默认主题说明
    ├── theme.go        # 主题组件实现
    ├── theme.yaml      # 主题元数据配置
    └── assets/         # 主题静态资源
```

## 主题注册机制

主题通过 blank import 触发自动注册：

```go
import _ "rua.plus/cadmus/themes/default"
```

`init()` 函数将主题注册到全局注册表，使系统能够识别和加载该主题。

---

# 主题开发指南

## 主题结构

每个主题目录应包含以下文件：

| 文件 | 用途 |
|------|------|
| `theme.go` | 主题实现：定义 `ThemeComponents` 接口的所有方法 |
| `theme.yaml` | 主题元数据：name, version, description, author |
| `templates/` | templ 模板文件（可选，可覆盖默认模板） |
| `assets/` | 主题专属静态资源（CSS, JS, 图片） |

## 主题接口实现

主题必须实现 `theme.ThemeComponents` 接口：

```go
type ThemeComponents interface {
    Layout() templ.Component      // 整体布局框架
    Header() templ.Component      // 头部导航栏
    Footer() templ.Component      // 底部区域
    PostList() templ.Component    // 文章列表页
    PostDetail() templ.Component  // 文章详情页
    CategoryPage() templ.Component // 分类页
    TagPage() templ.Component     // 标签页
    Sidebar() templ.Component     // 侧边栏
    ErrorPage() templ.Component   // 错误页面 (404, 500)
}
```

## 主题元数据格式

`theme.yaml` 定义主题基本信息：

```yaml
name: default
version: 1.0.0
description: Cadmus 默认主题
author: Cadmus Team
```

## 开发步骤

1. **创建主题目录**：`themes/my-theme/`

2. **编写 theme.yaml**：定义主题元数据

3. **实现 theme.go**：
   - 创建独立 package
   - 实现 `ThemeComponents` 接口所有方法
   - 实现 `New()` 函数返回 `theme.Theme` 对象
   - 在 `init()` 中调用 `theme.Register()` 注册主题

4. **添加模板**（可选）：在 `templates/` 下放置 `.templ` 文件覆盖默认模板

5. **添加静态资源**（可选）：将 CSS、JS、图片放入 `assets/`

## 示例：最小主题实现

```go
package mytheme

import (
    "github.com/a-h/templ"
    "rua.plus/cadmus/internal/theme"
    "rua.plus/cadmus/web/templates/layouts"
    "rua.plus/cadmus/web/templates/pages"
)

type MyTheme struct{}

var _ theme.ThemeComponents = (*MyTheme)(nil)

func New() theme.Theme {
    return theme.Theme{
        ID:          "my-theme",
        Name:        "My Theme",
        Version:     "1.0.0",
        Author:      "Your Name",
        Description: "我的自定义主题",
        Components:  &MyTheme{},
    }
}

func (t *MyTheme) Layout() templ.Component {
    return layouts.BaseLayout("My Site")
}

func (t *MyTheme) Header() templ.Component {
    return layouts.Header()
}

func (t *MyTheme) Footer() templ.Component {
    return layouts.Footer()
}

func (t *MyTheme) PostList() templ.Component {
    return pages.PostListPage("文章列表", nil, pages.Pagination{}, nil, nil)
}

func (t *MyTheme) PostDetail() templ.Component {
    return pages.PostDetailPage(nil, nil, nil, nil)
}

func (t *MyTheme) CategoryPage() templ.Component {
    return nil
}

func (t *MyTheme) TagPage() templ.Component {
    return nil
}

func (t *MyTheme) Sidebar() templ.Component {
    return nil
}

func (t *MyTheme) ErrorPage() templ.Component {
    return nil
}

func init() {
    _ = theme.Register(New())
}
```

## 主题切换

主题通过配置文件或环境变量选择。系统加载指定主题后，所有页面渲染均使用该主题的组件实现。

## 注意事项

- 主题 ID 必须唯一，避免与其他主题冲突
- `init()` 注册失败（如主题已存在）会被忽略
- 未实现的组件方法可返回 `nil`，系统会回退到默认行为
- 使用 `templ` 模板引擎确保类型安全的 HTML 生成
