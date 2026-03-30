---
title: Cadmus 博客平台设计方案
date: 2026-03-30
---

## Context

Cadmus 是一个多用户博客平台，支持完整的主题和插件系统。目标是提供一个类似 WordPress 的灵活平台，但使用 Go + templ 的现代技术栈，编译部署模式。

---

## 技术栈总览

| 层级     | 技术选型                        | 说明                       |
| -------- | ------------------------------- | -------------------------- |
| 后端框架 | Go 1.26 + templ                 | templ 编译型模板，类型安全 |
| 前台渲染 | templ + CSS                     | 服务端渲染，SEO 友好       |
| 前端构建 | esbuild + TypeScript + Tailwind | 原生 JS，零框架依赖        |
| 后台管理 | templ + 增强型 JS               | 表格、拖拽等复杂交互       |
| 数据库   | PostgreSQL                   | 先用 PG，后续可扩展 SQLite |
| 认证     | JWT                          | 无状态，适合分布式         |
| 缓存     | Redis                        | 热点数据缓存               |

---

## 核心模块架构

```
cadmus/
├── cmd/                    # 入口
│   └── server/
│       └── main.go
├── internal/
│   ├── core/               # 核心业务
│   │   ├── user/           # 用户、角色、权限
│   │   ├── post/           # 文章、分类、标签
│   │   ├── comment/        # 评论系统
│   │   ├── block/          # 块编辑器核心
│   │   ├── media/          # 多媒体附件
│   │   ├── search/         # 全文搜索
│   │   ├── rss/            # RSS 订阅
│   │   └── notify/         # 邮件通知
│   ├── api/                 # REST API
│   │   └── handlers/
│   ├── auth/                # JWT 认证
│   ├── cache/               # Redis 缓存
│   └── database/            # PostgreSQL 连接
│   ├── theme/               # 主题引擎
│   ├── plugin/              # 插件引擎
├── pkg/                     # 公共库
│   ├── interfaces/          # 插件接口定义
│   └── utils/
├── web/
│   ├── frontend/            # Vite + TS + Tailwind
│   │   ├── src/
│   │   │   ├── components/  # 前台交互组件
│   │   │   ├── admin/       # 后台管理组件
│   │   │   ├── editor/      # 块编辑器
│   │   │   └── styles/      # Tailwind 入口
│   │   ├── package.json
│   │   ├── vite.config.ts
│   │   ├── tsconfig.json
│   │   └── tailwind.config.js
│   ├── templates/           # templ 组件
│   │   ├── layouts/
│   │   ├── pages/
│   │   ├── partials/
│   ├── static/              # Vite 构建输出
├── themes/                  # 默认主题（示例）
├── plugins/                 # 默认插件（示例）
├── migrations/              # 数据库迁移
├── configs/                 # 配置文件
└── docs/
```

---

## 1. 用户与权限系统

### 1.1 数据模型

```go
// User
type User struct {
    ID          uuid.UUID
    Username    string
    Email       string
    PasswordHash string
    AvatarURL   string
    Bio         string
    RoleID      uuid.UUID
    Status      UserStatus // active/banned/pending
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Role - 自定义角色
type Role struct {
    ID          uuid.UUID
    Name        string
    DisplayName string
    Permissions []Permission
    IsDefault   bool
    CreatedAt   time.Time
}

// Permission - 细粒度权限
type Permission struct {
    ID          uuid.UUID
    Name        string    // post.create, post.edit, comment.delete 等
    Description string
    Category    string    // post/comment/user/theme/plugin 等
}
```

### 1.2 权限检查流程

> **P0 修复**：原设计每次权限检查都查询数据库，存在严重性能瓶颈。现增加 Redis 缓存机制。

```go
// 类型安全的 context key
type ctxKey string
const ctxUserID ctxKey = "user_id"

// 权限缓存服务
type PermissionCache struct {
    cache redis.Client
    ttl   time.Duration  // 默认 1 小时
}

// 带缓存的权限检查
func (s *UserService) HasPermission(ctx context.Context, user *User, perm string) bool {
    cacheKey := fmt.Sprintf("cadmus:user:perms:%s:%s", user.ID, perm)

    // 1. 尝试缓存命中
    cached, err := s.cache.Get(ctx, cacheKey).Bool()
    if err == nil {
        return cached
    }

    // 2. 缓存未命中，查询数据库
    role := s.GetRoleCached(ctx, user.RoleID)
    for _, p := range role.Permissions {
        if p.Name == perm {
            s.cache.Set(ctx, cacheKey, true, s.ttl)
            return true
        }
    }

    // 3. 缓存否定结果（防止缓存穿透）
    s.cache.Set(ctx, cacheKey, false, s.ttl)
    return false
}

// 获取角色（带缓存）
func (s *UserService) GetRoleCached(ctx context.Context, roleID uuid.UUID) *Role {
    cacheKey := fmt.Sprintf("cadmus:role:info:%s", roleID)

    var role Role
    if err := s.cache.Get(ctx, cacheKey).Scan(&role); err == nil {
        return &role
    }

    // 查询数据库并缓存
    role = s.repo.GetRoleByID(roleID)
    s.cache.Set(ctx, cacheKey, role, time.Hour)
    return &role
}

// 权限变更时清除缓存
func (s *UserService) InvalidateUserPermissions(userID uuid.UUID) {
    pattern := fmt.Sprintf("cadmus:user:perms:%s:*", userID)
    s.cache.DeleteByPattern(context.Background(), pattern)
}
```

### 1.3 默认角色

| 角色          | 权限范围                    |
| ------------- | --------------------------- |
| Administrator | 全部权限 + 角色管理         |
| Editor        | 文章管理 + 评论审核         |
| Author        | 发布自己文章 + 管理自己评论 |
| Moderator     | 评论审核                    |
| Subscriber    | 评论发表                    |

---

## 2. 文章系统

### 2.1 数据模型

```go
type Post struct {
    ID          uuid.UUID
    AuthorID    uuid.UUID
    Title       string
    Slug        string
    Content     BlockDocument // JSON 存储
    Excerpt     string
    CategoryID  uuid.UUID
    Tags        []Tag
    Status      PostStatus // draft/published/scheduled/private
    PublishAt   *time.Time
    FeaturedImage string
    SEO         SEOMeta
    ViewCount   int
    LikeCount   int
    SeriesID    *uuid.UUID  // 文章系列
    SeriesOrder int
    IsPaid      bool        // 付费文章
    Price       *float64
    Version     int         // 版本号
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type SEOMeta struct {
    Title       string
    Description string
    Keywords    []string
}

type PostVersion struct {
    ID        uuid.UUID
    PostID    uuid.UUID
    Version   int
    Content   BlockDocument
    CreatedAt time.Time
    CreatorID uuid.UUID
    Note      string // 版本说明
}
```

### 2.2 核心功能

- **草稿/发布/定时发布**：状态流转 + 定时任务
- **版本历史**：每次保存生成新版本，可回滚
- **付费文章**：IsPaid 标记 + 权限检查
- **文章系列**：SeriesID + SeriesOrder 关联
- **SEO 元数据**：独立字段，模板渲染时注入

---

## 3. 块编辑器

### 3.1 数据结构

```go
// BlockDocument - 文章内容存储格式
type BlockDocument struct {
    Version  int
    Blocks   []Block
}

type Block struct {
    ID       string
    Type     string      // paragraph/heading/image/code/quote/list/custom...
    Data     BlockData   // 类型特定的数据
    Children []Block     // 嵌套块（如嵌套列表）
    Meta     BlockMeta   // 对齐、样式等
}

// BlockData - 接口，不同块类型实现
type BlockData interface {
    Validate() error
    Render() string // HTML 渲染逻辑
}
```

### 3.2 默认块类型

| 类型      | 数据结构                                         |
| --------- | ------------------------------------------------ |
| paragraph | `{ text: string }`                               |
| heading   | `{ text: string, level: 1-6 }`                   |
| image     | `{ url: string, alt: string, caption?: string }` |
| code      | `{ code: string, language: string }`             |
| quote     | `{ text: string, source?: string }`              |
| list      | `{ items: string[], style: ordered/unordered }`  |
| divider   | `{}`                                             |
| table     | `{ rows: [][]string }`                           |
| embed     | `{ url: string, type: video/link }`              |
| file      | `{ url: string, name: string, size: number }`    |
| callout   | `{ text: string, type: tip/warning/error }`      |

### 3.3 插件扩展机制

```go
// BlockTypeRegistry - 块类型注册器
type BlockTypeRegistry interface {
    Register(blockType BlockType) error
    Get(typeName string) (BlockType, error)
    All() []BlockType
}

// BlockType - 插件定义的块类型
type BlockType interface {
    Name() string
    ParseData(data json.RawMessage) (BlockData, error)
    Render(data BlockData) string
    EditorComponent() string // 编辑器 UI 组件名（前端 JS 注册）
}
```

---

## 4. 评论系统

### 4.1 数据模型

```go
type Comment struct {
    ID        uuid.UUID
    PostID    uuid.UUID
    UserID    uuid.UUID
    ParentID  *uuid.UUID  // 嵌套回复
    Depth     int         // 嵌套深度（0-5）
    Content   string
    Status    CommentStatus // pending/approved/spam/deleted
    LikeCount int
    CreatedAt time.Time
    UpdatedAt time.Time
}

type CommentLike struct {
    ID        uuid.UUID
    CommentID uuid.UUID
    UserID    uuid.UUID
    CreatedAt time.Time
}
```

### 4.2 功能

- **嵌套回复**：ParentID 关联，支持多级嵌套
- **审核系统**：pending → approved 流程，Moderator 权限
- **垃圾评论检测**：接口扩展点，插件可接入检测逻辑

### 4.3 嵌套深度限制

> **P0 设计决策**：限制评论嵌套深度为 5 层，防止无限递归和性能问题。

#### 为什么需要深度限制

| 问题 | 无限制 | 限制后 |
|------|--------|--------|
| 数据库查询复杂度 | O(n) 递归 | 固定 5 层 JOIN |
| 前端渲染性能 | 无限嵌套 DOM | 可控层级 |
| 用户体验 | 嵌套过深难以阅读 | 结构清晰 |
| SQL 注入风险 | 递归 CTE 攻击 | 固定深度安全 |

#### Go 服务层实现

```go
package comment

const MaxCommentDepth = 5

type CommentService struct {
    db    *sql.DB
    cache CacheService
}

// CreateComment 创建评论（带深度检查）
func (s *CommentService) CreateComment(ctx context.Context, input CreateCommentInput) (*Comment, error) {
    // 计算深度
    depth := 0
    if input.ParentID != nil {
        parent, err := s.GetByID(ctx, *input.ParentID)
        if err != nil {
            return nil, fmt.Errorf("parent comment not found: %w", err)
        }
        
        // 检查是否超过最大深度
        if parent.Depth >= MaxCommentDepth {
            return nil, ErrMaxDepthExceeded
        }
        depth = parent.Depth + 1
    }
    
    // 创建评论
    comment := &Comment{
        ID:        uuid.New(),
        PostID:    input.PostID,
        UserID:    input.UserID,
        ParentID:  input.ParentID,
        Depth:     depth,
        Content:   input.Content,
        Status:    CommentStatusPending,
        CreatedAt: time.Now(),
    }
    
    // 数据库插入（数据库也有 CHECK 约束作为最后一道防线）
    if err := s.insert(ctx, comment); err != nil {
        return nil, err
    }
    
    return comment, nil
}

// GetCommentsByPost 获取文章评论（扁平化 + 树形构建）
func (s *CommentService) GetCommentsByPost(ctx context.Context, postID uuid.UUID) ([]*Comment, error) {
    // 一次查询获取所有评论
    rows, err := s.db.Query(ctx, `
        SELECT id, post_id, user_id, parent_id, depth, content, status, like_count, created_at
        FROM comments
        WHERE post_id = $1 AND status = 'approved'
        ORDER BY created_at ASC
    `, postID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var comments []*Comment
    for rows.Next() {
        c := &Comment{}
        if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.ParentID, &c.Depth, 
            &c.Content, &c.Status, &c.LikeCount, &c.CreatedAt); err != nil {
            return nil, err
        }
        comments = append(comments, c)
    }
    
    // 在内存中构建树形结构（避免递归查询）
    return buildCommentTree(comments), nil
}

// buildCommentTree 将扁平评论列表构建为树形结构
func buildCommentTree(comments []*Comment) []*Comment {
    // 按深度分组
    byDepth := make(map[int][]*Comment)
    for _, c := range comments {
        byDepth[c.Depth] = append(byDepth[c.Depth], c)
    }
    
    // 建立父子关系映射
    byID := make(map[uuid.UUID]*Comment)
    for _, c := range comments {
        byID[c.ID] = c
    }
    
    // 找出根评论（depth=0）
    var roots []*Comment
    for _, c := range byDepth[0] {
        roots = append(roots, c)
    }
    
    // 构建子评论列表
    for depth := 1; depth <= MaxCommentDepth; depth++ {
        for _, c := range byDepth[depth] {
            if parent, ok := byID[*c.ParentID]; ok {
                parent.Children = append(parent.Children, c)
            }
        }
    }
    
    return roots
}

// 错误定义
var ErrMaxDepthExceeded = errors.New("comment depth exceeds maximum limit of 5")
```

#### 前端渲染优化

```typescript
// web/frontend/src/components/comment.ts

const MAX_DISPLAY_DEPTH = 5;

interface Comment {
  id: string;
  parentId?: string;
  depth: number;
  content: string;
  children: Comment[];
}

function renderComment(comment: Comment, depth: number = 0): string {
  if (depth > MAX_DISPLAY_DEPTH) {
    return ''; // 超过最大深度不渲染
  }
  
  const indent = depth * 20; // 每层缩进 20px
  const children = comment.children
    .map(c => renderComment(c, depth + 1))
    .join('');
  
  return `
    <div class="comment" style="margin-left: ${indent}px" data-depth="${depth}">
      <div class="comment-content">${escapeHtml(comment.content)}</div>
      ${depth < MAX_DISPLAY_DEPTH ? `
        <button class="reply-btn" data-comment-id="${comment.id}">回复</button>
        <div class="comment-children">${children}</div>
      ` : `
        <span class="depth-limit-hint">回复层级已达上限</span>
      `}
    </div>
  `;
}
```

#### API 响应示例

```json
{
  "comments": [
    {
      "id": "uuid-1",
      "depth": 0,
      "content": "这篇文章写得很好！",
      "children": [
        {
          "id": "uuid-2",
          "depth": 1,
          "content": "同意！",
          "children": []
        }
      ]
    }
  ]
}
```

---

## 5. 主题系统

### 5.1 架构设计

主题是 Go 模块，实现核心定义的组件接口：

```go
// Theme - 主题元信息
type Theme struct {
    ID          string
    Name        string
    Version     string
    Author      string
    Description string
    Components  ThemeComponents
}

// ThemeComponents - 主题必须实现的组件
type ThemeComponents interface {
    Layout() templ.Component       // 整体布局框架
    Header() templ.Component       // 头部
    Footer() templ.Component       // 底部
    PostList() templ.Component     // 文章列表页
    PostDetail() templ.Component   // 文章详情页
    CategoryPage() templ.Component // 分类页
    TagPage() templ.Component      // 标签页
    Sidebar() templ.Component      // 侧边栏（可选）
    ErrorPage() templ.Component    // 错误页
}
```

### 5.2 主题注册

```go
// ThemeRegistry - 主题管理
type ThemeRegistry interface {
    Register(theme Theme) error
    GetActive() Theme
    SetActive(themeID string) error
    All() []Theme
}
```

### 5.3 主题目录结构

```
themes/
└── default/
    ├── theme.go          # Theme 实现
    ├── components/
    │   ├── layout_templ.go
    │   ├── header_templ.go
    │   ├── post_list_templ.go
    │   └── ...
    ├── static/
    │   ├── css/
    │   ├── images/
    └── theme.json        # 元信息（可选，也可在代码中定义）
```

---

## 6. 插件系统

### 6.1 接口定义

```go
// Plugin - 插件元信息 + 入口
type Plugin interface {
    Info() PluginInfo
    Init(ctx *PluginContext) error
}

type PluginInfo struct {
    ID          string
    Name        string
    Version     string
    Author      string
    Description string
    Dependencies []string
}

type PluginContext struct {
    DB         *sql.DB
    Cache      CacheService
    Registry   *BlockTypeRegistry
    Services   *ServicesRegistry // 核心服务访问
    Config     map[string]any
}
```

### 6.2 功能扩展接口

```go
// CommentFilter - 评论处理扩展
type CommentFilter interface {
    BeforeSave(comment *Comment) (*Comment, error)
    AfterSave(comment *Comment) error
}

// PostRenderer - 文章渲染扩展
type PostRenderer interface {
    BeforeRender(post *Post) (*Post, error)
    AfterRender(html string) (string, error)
}

// AuthProvider - 第三方登录扩展
type AuthProvider interface {
    Name() string
    Authenticate(token string) (*UserInfo, error)
    Config() AuthProviderConfig
}

// BlockType - 自定义块类型（见块编辑器部分）

// NotificationChannel - 通知渠道扩展
type NotificationChannel interface {
    Name() string
    Send(to string, subject string, body string) error
}
```

### 6.3 插件目录结构

```
plugins/
├── github-auth/          # GitHub OAuth 登录示例
│   ├── plugin.go
│   ├── auth.go
│   └── plugin.json
├── mermaid-block/        # Mermaid 图表块示例
│   ├── plugin.go
│   ├── block.go
│   └── plugin.json
├── discord-notify/       # Discord 通知示例
│   ├── plugin.go
│   └── notify.go
│   └── plugin.json
```

### 6.4 编译与加载机制

> **P0 设计决策**：采用编译时接口注册模式，而非 Go 原生 `plugin` 包的动态加载。

#### 为什么不使用 Go 原生 plugin 包

Go 原生 `plugin` 包存在以下限制：

| 限制 | 影响 |
|------|------|
| 仅支持 Linux/FreeBSD/macOS | Windows 用户无法使用 |
| 主程序与插件必须使用完全相同的 Go 版本 | 部署复杂度增加 |
| 需要 CGO_ENABLED=1 | 纯 Go 编译不可用 |
| 插件加载后无法卸载 | 内存无法释放 |
| Race detector 支持不完善 | 调试困难 |

#### 采用编译时注册模式

参考 [Alist](https://github.com/AlistGo/alist)（40K+ stars）的生产验证模式：

```go
// internal/plugin/registry.go
package plugin

var pluginMap = map[string]PluginConstructor{}

type PluginConstructor func() Plugin

// RegisterPlugin 注册插件（由插件的 init() 调用）
func RegisterPlugin(ctor PluginConstructor) {
    p := ctor()
    info := p.Info()
    
    // 验证依赖
    for _, dep := range info.Dependencies {
        if _, ok := pluginMap[dep]; !ok {
            log.Printf("Warning: plugin %s depends on %s which is not registered", info.ID, dep)
        }
    }
    
    // 初始化插件
    ctx := &PluginContext{
        DB:       GetDB(),
        Cache:    GetCache(),
        Registry: GetBlockRegistry(),
        Services: GetServices(),
    }
    if err := p.Init(ctx); err != nil {
        log.Printf("Warning: plugin %s failed to initialize: %v", info.ID, err)
        return
    }
    
    pluginMap[info.ID] = ctor
    log.Printf("Plugin registered: %s (%s)", info.Name, info.Version)
}

// GetPlugin 获取已注册的插件
func GetPlugin(id string) (Plugin, error) {
    ctor, ok := pluginMap[id]
    if !ok {
        return nil, fmt.Errorf("plugin not found: %s", id)
    }
    return ctor(), nil
}

// AllPlugins 返回所有已注册插件
func AllPlugins() []PluginInfo {
    var infos []PluginInfo
    for _, ctor := range pluginMap {
        infos = append(infos, ctor().Info())
    }
    return infos
}
```

#### 插件实现示例

```go
// plugins/github-auth/plugin.go
package github_auth

import (
    "rua.plus/cadmus/internal/plugin"
    "rua.plus/cadmus/pkg/interfaces"
)

type GitHubAuthPlugin struct {
    config interfaces.AuthProviderConfig
}

func (p *GitHubAuthPlugin) Info() plugin.PluginInfo {
    return plugin.PluginInfo{
        ID:          "github-auth",
        Name:        "GitHub OAuth 登录",
        Version:     "1.0.0",
        Author:      "Cadmus Team",
        Description: "支持 GitHub 账号登录",
        Dependencies: []string{},
    }
}

func (p *GitHubAuthPlugin) Init(ctx *plugin.PluginContext) error {
    // 从配置初始化
    p.config = ctx.Config["github"].(interfaces.AuthProviderConfig)
    return nil
}

// 实现 AuthProvider 接口
func (p *GitHubAuthPlugin) Name() string {
    return "github"
}

func (p *GitHubAuthPlugin) Authenticate(token string) (*interfaces.UserInfo, error) {
    // GitHub OAuth 认证逻辑...
}

func (p *GitHubAuthPlugin) Config() interfaces.AuthProviderConfig {
    return p.config
}

// init() 自动注册
func init() {
    plugin.RegisterPlugin(func() plugin.Plugin {
        return &GitHubAuthPlugin{}
    })
}
```

#### 主程序启用插件

通过 blank import 触发插件 init()：

```go
// cmd/server/main.go
package main

import (
    // ... 其他 import
    
    // 启用内置插件（blank import 触发 init()）
    _ "rua.plus/cadmus/plugins/github-auth"
    _ "rua.plus/cadmus/plugins/mermaid-block"
    _ "rua.plus/cadmus/plugins/discord-notify"
)

func main() {
    // 插件已自动注册，无需额外代码
    plugins := plugin.AllPlugins()
    log.Printf("Loaded %d plugins", len(plugins))
    
    // ... 启动服务器
}
```

#### 用户自定义插件

用户创建自己的插件模块后，在编译时 import：

```go
// 用户项目的 main.go
import (
    _ "rua.plus/cadmus/plugins/builtin"      // 内置插件
    _ "github.com/user/my-cadmus-plugin"      // 用户自定义插件
)
```

**优势**：
- ✅ 跨平台支持（Windows/Linux/macOS）
- ✅ 类型安全（编译时验证接口实现）
- ✅ 无版本兼容问题
- ✅ 简单可靠（无 subprocess 管理）
- ✅ 生产验证（Alist 等大型项目采用）

---

## 7. REST API 设计

### 7.1 路由结构

```
/api/v1/
├── auth/
│   ├── POST   /login          # 登录
│   ├── POST   /register       # 注册
│   ├── POST   /logout         # 登出
│   ├── POST   /refresh        # 刷新 token
│   ├── GET    /me             # 当前用户信息
├── posts/
│   ├── GET    /               # 文章列表
│   ├── GET    /:slug          # 文章详情
│   ├── POST   /               # 创建文章（需权限）
│   ├── PUT    /:id            # 更新文章
│   ├── DELETE /:id            # 删除文章
│   ├── POST   /:id/like       # 点赞
│   ├── GET    /:id/versions   # 版本历史
│   ├── POST   /:id/rollback   # 回滚版本
├── comments/
│   ├── GET    /post/:postId   # 文章评论列表
│   ├── POST   /               # 发表评论
│   ├── PUT    /:id            # 编辑评论
│   ├── DELETE /:id            # 删除评论
│   ├── POST   /:id/like       # 点赞
│   ├── PUT    /:id/approve    # 审核（需权限）
├── users/
│   ├── GET    /:id            # 用户信息
│   ├── PUT    /:id            # 更新用户信息
│   ├── GET    /:id/posts      # 用户文章列表
├── categories/
│   ├── GET    /               # 分类列表
│   ├── POST   /               # 创建分类（需权限）
│   ├── PUT    /:id            # 更新分类
│   ├── DELETE /:id            # 删除分类
├── tags/
│   ├── GET    /               # 标签列表
│   ├── POST   /               # 创建标签（需权限）
│   ├── DELETE /:id            # 删除标签
├── admin/
│   ├── roles/
│   │   ├── GET    /           # 角色列表
│   │   ├── POST   /           # 创建角色
│   │   ├── PUT    /:id        # 更新角色权限
│   │   ├── DELETE /:id        # 删除角色
│   ├── users/
│   │   ├── GET    /           # 用户管理列表
│   │   ├── PUT    /:id/ban    # 封禁用户
│   ├── comments/
│   │   ├── GET    /pending    # 待审核评论
│   │   ├── POST   /batch      # 批量操作
│   ├── plugins/
│   │   ├── GET    /           # 插件列表
│   │   ├── PUT    /:id/config # 配置插件
│   ├── themes/
│   │   ├── GET    /           # 主题列表
│   │   ├── PUT    /active     # 切换主题
│   ├── batch/
│   │   ├── POST   /           # 批量操作
│   │   Body: { action: string, ids: string[], params?: object }
│   ├── order/
│   │   ├── PUT    /           # 更新排序
│   │   Body: { order: string[] }
├── search/
│   ├── GET    /               # 全文搜索
├── rss/
│   ├── GET    /               # RSS feed
├── media/
│   ├── POST   /upload         # 上传文件
│   ├── DELETE /:id            # 删除文件
```

### 7.2 认证中间件

> **P0 修复**：原设计使用字符串作为 context key，存在类型安全问题。现改为类型安全的 context key。

```go
// 类型安全的 context key
type ctxKey string

const (
    ctxUserID   ctxKey = "user_id"
    ctxUserRole ctxKey = "user_role_id"
)

func AuthMiddleware(jwtService *JWTService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := ExtractToken(r)
            if token == "" {
                WriteAPIError(w, "AUTH_FAILED", "未授权访问", nil, http.StatusUnauthorized)
                return
            }
            claims, err := jwtService.Validate(token)
            if err != nil {
                WriteAPIError(w, "AUTH_FAILED", "无效的令牌", nil, http.StatusUnauthorized)
                return
            }
            // 使用类型安全的 context key
            ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
            ctx = context.WithValue(ctx, ctxUserRole, claims.RoleID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// 类型安全的取值函数
func GetUserID(ctx context.Context) (uuid.UUID, error) {
    id, ok := ctx.Value(ctxUserID).(uuid.UUID)
    if !ok {
        return uuid.Nil, errors.New("user not authenticated")
    }
    return id, nil
}

func GetUserRoleID(ctx context.Context) (uuid.UUID, error) {
    id, ok := ctx.Value(ctxUserRole).(uuid.UUID)
    if !ok {
        return uuid.Nil, errors.New("user role not found")
    }
    return id, nil
}

// 统一的 API 错误响应格式
type APIError struct {
    Code      string   `json:"code"`       // "AUTH_FAILED", "VALIDATION_ERROR"
    Message   string   `json:"message"`
    Details   []string `json:"details"`    // 详细错误列表
    RequestID string   `json:"request_id"`
}

func WriteAPIError(w http.ResponseWriter, code string, message string, details []string, status int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(APIError{
        Code:      code,
        Message:   message,
        Details:   details,
        RequestID: GetRequestID(),
    })
}
```

---

## 8. 缓存策略

### 8.1 缓存层级

| 数据类型 | 缓存策略        | TTL     |
| -------- | --------------- | ------- |
| 文章详情 | 缓存渲染后 HTML | 10 分钟 |
| 文章列表 | 缓存分页数据    | 5 分钟  |
| 用户信息 | 缓存基本信息    | 30 分钟 |
| 角色权限 | 缓存权限列表    | 1 小时  |
| 配置数据 | 缓存站点配置    | 1 小时  |
| 搜索结果 | 不缓存          | -       |

### 8.2 Redis Key 命名规范

> **P0 修复**：原设计缺少 Redis key 命名规范，可能导致 key 冲突和管理混乱。现定义统一的命名规范。

```go
// 缓存 key 常量定义
const (
    CacheNamespace = "cadmus"

    // 文章缓存 key 格式
    PostDetailKey = "cadmus:post:detail:%s:v%d"    // {slug}:{version}
    PostListKey   = "cadmus:post:list:%s:%d:%s"    // {category}:{page}:{sort}

    // 用户缓存 key 格式
    UserInfoKey  = "cadmus:user:info:%s"           // {user_id}
    UserPermsKey = "cadmus:user:perms:%s:%s"       // {user_id}:{permission}

    // 角色缓存 key 格式
    RoleInfoKey  = "cadmus:role:info:%s"           // {role_id}
    RolePermsKey = "cadmus:role:perms:%s"          // {role_id}

    // 配置缓存 key 格式
    SiteConfigKey = "cadmus:site:config"
    ThemeConfigKey = "cadmus:theme:config:%s"      // {theme_id}
)

// 缓存 key 构建函数
func BuildPostDetailKey(slug string, version int) string {
    return fmt.Sprintf(PostDetailKey, slug, version)
}

func BuildUserPermsKey(userID uuid.UUID, permission string) string {
    return fmt.Sprintf(UserPermsKey, userID, permission)
}
```

**Key 命名规则：**

- 格式：`{namespace}:{entity}:{id}:{variant}`
- 使用冒号 `:` 分隔层级
- 包含版本号实现自然失效（如 `v{version}`）
- 单个 key 长度不超过 200 字符

### 8.3 缓存失效

- 文章更新 → 删除文章详情缓存 + 删除相关列表缓存
- 用户更新 → 删除用户缓存
- 角色权限变更 → 删除所有用户权限缓存

### 8.4 缓存穿透/击穿防护

> **P0 修复**：原设计缺少缓存防护机制，可能导致数据库压力。

```go
// 缓存穿透防护：缓存空值
func (c *CacheService) GetWithNullProtection(ctx context.Context, key string, dest interface{}, fetchFn func() (interface{}, error)) error {
    // 1. 尝试缓存命中
    err := c.Get(ctx, key).Scan(dest)
    if err == nil {
        // 检查是否为空值标记
        if c.IsNullMarker(dest) {
            return ErrNotFound
        }
        return nil
    }

    // 2. 查询数据
    data, err := fetchFn()
    if err != nil {
        return err
    }

    // 3. 缓存结果（包括空值）
    if data == nil {
        c.Set(ctx, key, NullMarker, 5*time.Minute) // 空值缓存短 TTL
        return ErrNotFound
    }

    c.Set(ctx, key, data, c.ttl)
    return nil
}

// 缓存击穿防护：互斥锁
func (c *CacheService) GetWithMutex(ctx context.Context, key string, dest interface{}, fetchFn func() (interface{}, error)) error {
    err := c.Get(ctx, key).Scan(dest)
    if err == nil {
        return nil
    }

    // 获取分布式锁
    mutexKey := key + ":lock"
    if c.SetNX(ctx, mutexKey, "1", 10*time.Second) {
        defer c.Del(ctx, mutexKey)
        // 双重检查
        err = c.Get(ctx, key).Scan(dest)
        if err == nil {
            return nil
        }
        // 查询并缓存
        data, err := fetchFn()
        if err != nil {
            return err
        }
        c.Set(ctx, key, data, c.ttl)
        return nil
    }

    // 等待其他 goroutine 完成
    time.Sleep(100 * time.Millisecond)
    return c.GetWithMutex(ctx, key, dest, fetchFn)
}
```

---

## 9. 全文搜索

### 9.1 方案选择

使用 PostgreSQL 全文搜索（pg_trgm 扩展）：

```sql
-- 启用扩展
CREATE EXTENSION pg_trgm;

-- 文章表添加搜索索引
ALTER TABLE posts ADD COLUMN search_vector tsvector;

CREATE INDEX posts_search_idx ON posts USING GIN(search_vector);

-- 更新触发器
CREATE FUNCTION update_search_vector() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', NEW.title), 'A') ||
        setweight(to_tsvector('english', NEW.excerpt), 'B') ||
        setweight(to_tsvector('english', NEW.content_text), 'C');
    RETURN NEW;
END
$$ LANGUAGE plpgsql;
```

### 9.2 搜索 API

```go
func (s *SearchService) Search(query string, filters SearchFilters) (*SearchResult, error) {
    sql := `
        SELECT id, title, excerpt, ts_rank(search_vector, $1) as rank
        FROM posts
        WHERE search_vector @@ to_tsquery($1)
        ORDER BY rank DESC
        LIMIT 20
    `
    // 执行查询...
}
```

---

## 10. 邮件通知

### 10.1 通知场景

| 事件     | 接收者     | 内容                         |
| -------- | ---------- | ---------------------------- |
| 评论发表 | 文章作者   | 新评论通知                   |
| 评论回复 | 被回复用户 | 回复通知                     |
| 文章发布 | 订阅用户   | 新文章通知（RSS + 邓件可选） |
| 用户注册 | 管理员     | 新用户通知                   |
| 系统警告 | 管理员     | 异常通知                     |

### 10.2 接口设计

```go
type NotificationService interface {
    SendCommentNotification(comment *Comment, post *Post) error
    SendReplyNotification(reply *Comment, parentAuthor *User) error
    SendPostNotification(post *Post, subscribers []string) error
}

// 雊件渠道实现
type EmailChannel struct {
    SMTP     SMTPConfig
    Template TemplateEngine
}
```

---

## 11. 前端架构

### 11.1 目录结构

```
web/frontend/
├── src/
│   ├── main.ts              # 入口
│   ├── admin/               # 后台管理
│   │   ├── posts/
│   │   │   ├── list.ts      # 文章列表（表格、分页）
│   │   │   ├── editor.ts    # 文章编辑器入口
│   │   ├── users/
│   │   ├── roles/
│   │   ├── comments/
│   │   ├── media/
│   │   ├── settings/
│   ├── editor/              # 块编辑器
│   │   ├── blocks/
│   │   │   ├── paragraph.ts
│   │   │   ├── heading.ts
│   │   │   ├── image.ts
│   │   │   ├── code.ts
│   │   │   └── ...
│   │   ├── core/
│   │   │   ├── document.ts  # BlockDocument 管理
│   │   │   ├── registry.ts  # 前端块注册
│   │   │   ├── renderer.ts  # 前端渲染预览
│   │   ├── ui/
│   │   │   ├── toolbar.ts   # 工具栏
│   │   │   ├── block-menu.ts
│   │   ├── index.ts         # 编辑器入口
│   ├── components/          # 前台交互组件
│   │   ├── comment.ts       # 评论提交、嵌套显示
│   │   ├── like.ts          # 点赞交互
│   │   ├── share.ts         # 分享按钮
│   │   ├── search.ts        # 搜索框
│   ├── api/                 # API 调用
│   │   ├── client.ts        # 基础请求
│   │   ├── auth.ts
│   │   ├── posts.ts
│   │   ├── comments.ts
│   ├── styles/
│   │   ├── main.css         # Tailwind 入口
│   │   ├── editor.css       # 编辑器样式
│   │   ├── admin.css        # 后台样式
│   ├── utils/
│   │   ├── dom.ts
│   │   ├── form.ts
│   │   ├── storage.ts       # LocalStorage
│   ├── types/               # TypeScript 类型定义
│   │   ├── block.d.ts
│   │   ├── api.d.ts
│   │   ├── user.d.ts
├── vite.config.ts
├── tailwind.config.js
├── tsconfig.json
├── package.json
```

### 11.2 构建流程

> **P0 设计决策**：采用 esbuild 而非 Vite。templ 官方推荐 esbuild 用于 SSR 场景，Vite 的 HMR 设计针对客户端 SPA。

#### 为什么选择 esbuild

| 特性 | Vite | esbuild | 推荐 |
|------|------|---------|------|
| SSR 支持 | 针对 SPA 设计 | 通用打包，适合 SSR | ✅ esbuild |
| HMR | 客户端热更新 | 配合 `templ generate --watch` | ✅ 组合方案 |
| 构建速度 | 快 | 极快（Go 编写） | ✅ esbuild |
| 与 templ 集成 | 需额外配置 | 官方文档推荐 | ✅ esbuild |

#### 开发模式 Makefile

```makefile
# Makefile - 多进程并行开发模式

.PHONY: live live/templ live/server live/esbuild live/tailwind

# 同时启动所有开发进程
live:
	@make -j5 live/templ live/server live/esbuild live/tailwind

# templ watch 模式（生成 Go 代码）
live/templ:
	templ generate --watch --proxy="http://localhost:8080" --open-browser=false

# Go 服务器（使用 air 热重载）
live/server:
	air -c .air.toml

# esbuild 打包 TypeScript
live/esbuild:
	npx esbuild src/main.ts \
		--bundle \
		--outdir=../static/dist \
		--watch \
		--sourcemap=inline \
		--format=esm

# Tailwind CSS 编译
live/tailwind:
	npx @tailwindcss/cli \
		-i src/styles/main.css \
		-o ../static/dist/styles.css \
		--watch

# 单独构建（生产环境）
build/frontend:
	cd web/frontend && npx esbuild src/main.ts \
		--bundle \
		--outdir=../static/dist \
		--minify \
		--format=esm
	npx @tailwindcss/cli \
		-i src/styles/main.css \
		-o ../static/dist/styles.css \
		--minify
```

#### Air 配置（Go 服务器热重载）

```toml
# .air.toml
root = "."
tmp_dir = "tmp"

[build]
bin = "./tmp/main"
cmd = "go build -o ./tmp/main ./cmd/server"
delay = 1000
exclude_dir = ["tmp", "vendor", "web/frontend/node_modules"]
exclude_file = ["*_templ.go"]  # templ 生成的文件
include_ext = ["go", "templ"]
kill_delay = "2s"

[log]
time = false

[color]
main = "magenta"
watcher = "cyan"
build = "yellow"
runner = "green"
```

#### 静态文件服务配置

```go
// cmd/server/main.go
import (
    "embed"
    "net/http"
)

// 开发模式：从文件系统读取（无缓存）
// 生产模式：从 embed.FS 读取

//go:embed static/dist/*
var staticAssets embed.FS

func setupStaticAssets(mux *http.ServeMux, devMode bool) {
    if devMode {
        // 开发模式：实时读取，无缓存
        mux.Handle("/static/", 
            http.StripPrefix("/static/", 
                http.FileServer(http.Dir("web/static"))))
    } else {
        // 生产模式：嵌入二进制，带缓存控制
        mux.Handle("/static/", 
            http.StripPrefix("/static/", 
                withCacheControl(http.FileServer(http.FS(staticAssets)))))
    }
}

func withCacheControl(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 静态资源缓存 1 年
        w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
        next.ServeHTTP(w, r)
    })
}

// 开发模式禁用缓存
func disableCacheInDevMode(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Cache-Control", "no-store, must-revalidate")
        next.ServeHTTP(w, r)
    })
}
```

#### 多入口构建（编辑器分离）

```makefile
# 编辑器单独入口（仅编辑器页面加载）
live/esbuild-editor:
	npx esbuild src/editor/index.ts \
		--bundle \
		--outdir=../static/dist/editor.js \
		--watch \
		--sourcemap=inline

# 生产构建
build/editor:
	npx esbuild src/editor/index.ts \
		--bundle \
		--outdir=../static/dist/editor.js \
		--minify
```

### 11.3 templ 集成

```go
// 页面模板引用前端资源
templ BaseLayout(title string) {
    <html>
        <head>
            <title>{ title }</title>
            <link rel="stylesheet" href="/static/dist/main.css">
        </head>
        <body>
            { children... }
            <script src="/static/dist/main.js"></script>
            <script src="/static/dist/editor.js"></script> // 仅编辑器页面
        </body>
    </html>
}
```

---

## 12. 配置管理

### 12.1 配置文件结构

```yaml
# configs/config.yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 5432
  name: "cadmus"
  user: "cadmus"
  password: "${DB_PASSWORD}"

redis:
  host: "localhost"
  port: 6379
  password: "${REDIS_PASSWORD}"

jwt:
  secret: "${JWT_SECRET}"
  expiry: "24h"

email:
  smtp_host: "smtp.example.com"
  smtp_port: 587
  smtp_user: "noreply@example.com"
  smtp_password: "${EMAIL_PASSWORD}"

site:
  name: "Cadmus Blog"
  description: "A modern blog platform"
  active_theme: "default"
  active_plugins:
    - "github-auth"
    - "mermaid-block"
```

---

## 13. 依赖管理原则

> **核心原则**：所有依赖使用最新稳定版本。

### 13.1 运行环境版本

| 组件 | 推荐版本 | 说明 | 来源 |
|------|---------|------|------|
| Go | 1.26.1 | 当前最新稳定版 | go.dev/dl |
| PostgreSQL | 18.3 | 最新稳定版 | postgresql.org |
| Redis | 8.6.2 | 最新稳定版 | redis.io |
| Node.js | 24.14.1 LTS | LTS 版本 | nodejs.org |
| Alpine | 3.23 | Docker 镜像最新版 | hub.docker.com |

### 13.2 Go 模块依赖

| 模块 | 版本 | 说明 |
|------|------|------|
| github.com/a-h/templ | v0.3.1001 | templ 模板引擎 |
| github.com/jackc/pgx/v5 | v5.9.1 | PostgreSQL 驱动 |

### 13.3 npm 前端依赖

> **P0 更新**：使用 esbuild 替代 Vite，符合 templ SSR 最佳实践。

| 包名 | 版本 | 说明 |
|------|------|------|
| esbuild | 0.25.4 | 构建工具（Go 编写，极快） |
| tailwindcss | 4.2.2 | CSS 框架 |
| typescript | 6.0.2 | TypeScript 编译器 |
| @tailwindcss/cli | 4.2.2 | Tailwind CLI 工具 |

**package.json 示例**：

```json
{
  "name": "cadmus-frontend",
  "version": "0.0.1",
  "type": "module",
  "scripts": {
    "build": "esbuild src/main.ts --bundle --outdir=../static/dist --minify --format=esm",
    "build:editor": "esbuild src/editor/index.ts --bundle --outdir=../static/dist/editor.js --minify",
    "build:css": "@tailwindcss/cli -i src/styles/main.css -o ../static/dist/styles.css --minify"
  },
  "devDependencies": {
    "esbuild": "^0.25.4",
    "tailwindcss": "^4.2.2",
    "typescript": "^6.0.2",
    "@tailwindcss/cli": "^4.2.2"
  }
}
```

### 13.4 更新检查流程

```bash
# Go 依赖更新
go list -m -u all  # 检查可更新模块
go get -u ./...    # 更新所有依赖

# npm 依赖更新
npm outdated       # 检查过期包
npm update         # 更新兼容版本

# Docker 镜像更新
docker-compose pull  # 拉取最新镜像
```

### 13.5 安全考量

- 关注依赖安全公告（`go list -m -u` / `npm audit`）
- 及时修复 CVE 漏洞
- 生产环境锁定版本（`go.sum` / `package-lock.json`）

---

## 14. 数据库迁移

使用迁移工具（如 migrate 或 goose）：

```
migrations/
├── 001_init.up.sql          # 初始化表结构
├── 001_init.down.sql
├── 002_add_post_version.up.sql
├── 002_add_post_version.down.sql
├── 003_add_series.up.sql
├── 003_add_series.down.sql
├── ...
```

### 14.1 完整数据库表结构

> **P0 补充**：设计文档提及但缺少迁移脚本的表结构定义。

#### 14.1.1 文章表 (posts)

```sql
-- 002_create_posts.up.sql
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_categories_slug ON categories(slug);
CREATE INDEX idx_categories_parent ON categories(parent_id);

CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tags_slug ON tags(slug);

CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    content JSONB NOT NULL,           -- BlockDocument JSON
    content_text TEXT,                 -- 纯文本用于搜索
    excerpt TEXT,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft/published/scheduled/private
    publish_at TIMESTAMP WITH TIME ZONE,
    featured_image TEXT,
    seo_title VARCHAR(255),
    seo_description TEXT,
    seo_keywords TEXT[],
    view_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    series_id UUID REFERENCES series(id) ON DELETE SET NULL,
    series_order INT DEFAULT 0,
    is_paid BOOLEAN DEFAULT FALSE,
    price DECIMAL(10, 2),
    version INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_slug ON posts(slug);
CREATE INDEX idx_posts_category ON posts(category_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_publish_at ON posts(publish_at) WHERE status = 'published';
CREATE INDEX idx_posts_series ON posts(series_id);

-- 全文搜索索引
ALTER TABLE posts ADD COLUMN search_vector tsvector;
CREATE INDEX posts_search_idx ON posts USING GIN(search_vector);

-- 标签关联表
CREATE TABLE post_tags (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

CREATE INDEX idx_post_tags_post ON post_tags(post_id);
CREATE INDEX idx_post_tags_tag ON post_tags(tag_id);

-- 文章版本历史
CREATE TABLE post_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    version INT NOT NULL,
    content JSONB NOT NULL,
    creator_id UUID NOT NULL REFERENCES users(id),
    note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(post_id, version)
);

CREATE INDEX idx_post_versions_post ON post_versions(post_id);

-- 文章系列
CREATE TABLE series (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    cover_image TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_series_author ON series(author_id);
CREATE INDEX idx_series_slug ON series(slug);
```

#### 14.1.2 评论表 (comments)

```sql
-- 003_create_comments.up.sql
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    depth INT NOT NULL DEFAULT 0,        -- 嵌套深度（P0：防止无限递归）
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending/approved/spam/deleted
    like_count INT DEFAULT 0,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 约束：最大嵌套深度为 5
    CONSTRAINT chk_depth CHECK (depth >= 0 AND depth <= 5)
);

CREATE INDEX idx_comments_post ON comments(post_id);
CREATE INDEX idx_comments_user ON comments(user_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);
CREATE INDEX idx_comments_status ON comments(status);
CREATE INDEX idx_comments_depth ON comments(depth);

-- 评论点赞
CREATE TABLE comment_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(comment_id, user_id)
);

CREATE INDEX idx_comment_likes_comment ON comment_likes(comment_id);
CREATE INDEX idx_comment_likes_user ON comment_likes(user_id);

-- 触发器：自动计算 depth
CREATE OR REPLACE FUNCTION calculate_comment_depth()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.depth := 0;
    ELSE
        SELECT depth + 1 INTO NEW.depth
        FROM comments
        WHERE id = NEW.parent_id;
        
        -- 检查是否超过最大深度
        IF NEW.depth > 5 THEN
            RAISE EXCEPTION 'Comment depth exceeds maximum limit of 5';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_comment_depth
    BEFORE INSERT ON comments
    FOR EACH ROW
    EXECUTE FUNCTION calculate_comment_depth();
```

#### 14.1.3 媒体表 (media)

```sql
-- 004_create_media.up.sql
CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    filepath TEXT NOT NULL,
    url TEXT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,              -- 字节数
    width INT,                          -- 图片宽度
    height INT,                         -- 图片高度
    alt_text TEXT,
    metadata JSONB,                     -- 其他元数据
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_media_uploader ON media(uploader_id);
CREATE INDEX idx_media_mime ON media(mime_type);
CREATE INDEX idx_media_created ON media(created_at DESC);
```

#### 14.1.4 其他辅助表

```sql
-- 005_create_auxiliary.up.sql

-- 文章点赞
CREATE TABLE post_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(post_id, user_id)
);

CREATE INDEX idx_post_likes_post ON post_likes(post_id);
CREATE INDEX idx_post_likes_user ON post_likes(user_id);

-- 站点配置
CREATE TABLE site_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 用户会话（用于 JWT 黑名单/刷新令牌）
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_sessions_token ON user_sessions(refresh_token_hash);
CREATE INDEX idx_sessions_expires ON user_sessions(expires_at);

-- 审计日志
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at DESC);
```

### 14.2 迁移文件规划

| 文件 | 内容 | 依赖 |
|------|------|------|
| `001_init.up.sql` | Users, Roles, Permissions | 无 |
| `002_create_posts.up.sql` | Categories, Tags, Posts, PostTags, PostVersions, Series | 001 |
| `003_create_comments.up.sql` | Comments, CommentLikes | 002 |
| `004_create_media.up.sql` | Media | 001 |
| `005_create_auxiliary.up.sql` | PostLikes, SiteSettings, UserSessions, AuditLogs | 002 |

---

## 15. 开发阶段规划

### Phase 1: 核心骨架

- 项目初始化 + 基础目录结构
- 数据库连接 + 迁移系统
- templ 配置 + 基础布局
- esbuild + TypeScript + Tailwind 配置

### Phase 2: 用户系统

- User + Role + Permission 模型
- JWT 认证
- 登录/注册 API
- 用户管理后台

### Phase 3: 文章系统

- Post 模型 + 数据库表
- REST API（文章 CRUD）
- 文章列表/详情 templ 组件
- 基础块编辑器

### Phase 4: 评论系统

- Comment 模型 + 嵌套结构
- 评论 API + templ 组件
- 审核流程

### Phase 5: 主题/插件引擎

- ThemeRegistry + PluginRegistry
- 默认主题实现
- 示例插件

### Phase 6: 增强功能

- 全文搜索
- RSS
- 邓件通知
- Redis 缓存

### Phase 7: 后台完善

- 文章管理后台（丰富 JS 交互）
- 评论审核后台
- 媒体管理
- 插件/主题配置

---

## 16. Docker 部署

### 16.1 Dockerfile

> **P0 修复**：原设计缺少安全加固（non-root 用户）和健康检查，现已改进。

```dockerfile
# 前端构建阶段
FROM node:24-alpine AS frontend-builder
WORKDIR /app/web/frontend
COPY web/frontend/package*.json ./
RUN npm ci --only=production
COPY web/frontend/ .
RUN npm run build

# Go 构建阶段
FROM golang:1.26-alpine AS go-builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/static /app/web/static
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cadmus ./cmd/server

# 运行阶段
FROM alpine:3.23
WORKDIR /app

# 安装依赖 + 创建非 root 用户
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 cadmus && \
    adduser -u 1000 -G cadmus -D cadmus

# 复制构建产物
COPY --from=go-builder /cadmus /app/cadmus
COPY --from=go-builder /app/configs /app/configs
COPY --from=go-builder /app/web/static /app/web/static

# 设置权限
RUN chown -R cadmus:cadmus /app

# 切换非 root 用户
USER cadmus

ENV TZ=Asia/Shanghai
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -q --spider http://localhost:8080/api/v1/health || exit 1

CMD ["/app/cadmus"]
```

### 16.2 docker-compose.yml

> **P0 修复**：原设计使用环境变量传递敏感信息，存在安全风险。现使用 Docker secrets。

```yaml
version: "3.8"

services:
  cadmus:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_PASSWORD_FILE=/run/secrets/db_password
      - REDIS_PASSWORD_FILE=/run/secrets/redis_password
      - JWT_SECRET_FILE=/run/secrets/jwt_secret
      - EMAIL_PASSWORD_FILE=/run/secrets/email_password
    secrets:
      - db_password
      - redis_password
      - jwt_secret
      - email_password
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./configs:/app/configs
      - ./uploads:/app/uploads
    healthcheck:
      test:
        ["CMD", "wget", "-q", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 3s
      retries: 3

  postgres:
    image: postgres:18-alpine
    environment:
      - POSTGRES_USER=cadmus
      - POSTGRES_PASSWORD_FILE=/run/secrets/postgres_password
      - POSTGRES_DB=cadmus
    secrets:
      - postgres_password
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U cadmus"]
      interval: 10s
      timeout: 3s
      retries: 5
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    ports:
      - "5432:5432"

  redis:
    image: redis:8-alpine
    command: redis-server --requirepass-file /run/secrets/redis_password
    secrets:
      - redis_password
    healthcheck:
      test:
        ["CMD", "redis-cli", "-a", "$(cat /run/secrets/redis_password)", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

secrets:
  db_password:
    file: ./secrets/db_password.txt
  redis_password:
    file: ./secrets/redis_password.txt
  jwt_secret:
    file: ./secrets/jwt_secret.txt
  email_password:
    file: ./secrets/email_password.txt
  postgres_password:
    file: ./secrets/postgres_password.txt

volumes:
  postgres_data:
  redis_data:
```

### 16.3 开发环境 Docker

```yaml
# docker-compose.dev.yml
version: "3.8"

services:
  postgres:
    image: postgres:18-alpine
    environment:
      - POSTGRES_USER=cadmus
      - POSTGRES_PASSWORD=cadmus
      - POSTGRES_DB=cadmus
    ports:
      - "5432:5432"
    volumes:
      - postgres_dev_data:/var/lib/postgresql/data

  redis:
    image: redis:8-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_dev_data:/data

volumes:
  postgres_dev_data:
  redis_dev_data:
```

---

## 17. 后台管理交互功能

### 17.1 表格排序

```typescript
// web/frontend/src/admin/table-sort.ts
interface SortConfig {
  field: string;
  direction: "asc" | "desc";
}

class TableSorter {
  private table: HTMLElement;
  private config: SortConfig;

  constructor(table: HTMLElement) {
    this.table = table;
    this.config = { field: "created_at", direction: "desc" };
    this.init();
  }

  private init() {
    const headers = this.table.querySelectorAll("th[data-sortable]");
    headers.forEach((header) => {
      header.addEventListener("click", () => this.sort(header.dataset.field!));
    });
  }

  private sort(field: string) {
    if (this.config.field === field) {
      this.config.direction = this.config.direction === "asc" ? "desc" : "asc";
    } else {
      this.config.field = field;
      this.config.direction = "asc";
    }
    this.updateUI();
    this.fetchData();
  }

  private fetchData() {
    const params = new URLSearchParams({
      sort: this.config.field,
      order: this.config.direction,
    });
    // API 调用...
  }
}
```

### 17.2 批量操作

```typescript
// web/frontend/src/admin/batch-actions.ts
interface BatchAction {
  action: "delete" | "approve" | "reject" | "move_category";
  ids: string[];
  params?: Record<string, any>;
}

class BatchOperator {
  private selectedIds: Set<string> = new Set();
  private checkboxes: NodeListOf<HTMLInputElement>;

  constructor(container: HTMLElement) {
    this.checkboxes = container.querySelectorAll("input[data-id]");
    this.init();
  }

  private init() {
    // 全选/取消全选
    document.querySelector("#select-all")?.addEventListener("change", (e) => {
      const checked = (e.target as HTMLInputElement).checked;
      this.checkboxes.forEach((cb) => {
        cb.checked = checked;
        this.toggleId(cb.dataset.id!, checked);
      });
      this.updateBatchButton();
    });

    // 单选
    this.checkboxes.forEach((cb) => {
      cb.addEventListener("change", () => {
        this.toggleId(cb.dataset.id!, cb.checked);
        this.updateBatchButton();
      });
    });
  }

  async executeBatch(action: BatchAction) {
    if (this.selectedIds.size === 0) return;

    const confirmMsg = `确认要${this.getActionText(action.action)} ${this.selectedIds.size} 项吗？`;
    if (!confirm(confirmMsg)) return;

    const response = await fetch("/api/v1/admin/batch", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        action: action.action,
        ids: Array.from(this.selectedIds),
        ...action.params,
      }),
    });

    if (response.ok) {
      this.refresh();
    }
  }

  private getActionText(action: string): string {
    const texts = {
      delete: "删除",
      approve: "批准",
      reject: "拒绝",
      move_category: "移动分类",
    };
    return texts[action] || action;
  }
}
```

### 17.3 拖拽排序

```typescript
// web/frontend/src/admin/drag-sort.ts
class DragSorter {
  private container: HTMLElement;
  private items: HTMLElement[];
  private dragItem: HTMLElement | null = null;
  private dragIndex: number = -1;

  constructor(container: HTMLElement) {
    this.container = container;
    this.items = Array.from(container.children) as HTMLElement[];
    this.init();
  }

  private init() {
    this.items.forEach((item) => {
      item.draggable = true;
      item.addEventListener("dragstart", (e) => this.onDragStart(e, item));
      item.addEventListener("dragover", (e) => this.onDragOver(e));
      item.addEventListener("drop", (e) => this.onDrop(e));
      item.addEventListener("dragend", () => this.onDragEnd());
    });
  }

  private onDragStart(e: DragEvent, item: HTMLElement) {
    this.dragItem = item;
    this.dragIndex = this.items.indexOf(item);
    e.dataTransfer!.effectAllowed = "move";
  }

  private onDragOver(e: DragEvent) {
    e.preventDefault();
    const target = e.target as HTMLElement;
    const targetIndex = this.items.indexOf(target);

    if (targetIndex !== this.dragIndex && this.dragItem) {
      // 视觉 reorder
      if (targetIndex > this.dragIndex) {
        target.after(this.dragItem);
      } else {
        target.before(this.dragItem);
      }
      this.items = Array.from(this.container.children) as HTMLElement[];
    }
  }

  private onDrop(e: DragEvent) {
    e.preventDefault();
    this.saveOrder();
  }

  private onDragEnd() {
    this.dragItem = null;
  }

  private async saveOrder() {
    const order = this.items.map((item) => item.dataset.id);
    await fetch("/api/v1/admin/order", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ order }),
    });
  }
}
```

### 17.4 筛选过滤

```typescript
// web/frontend/src/admin/filter.ts
interface FilterState {
  search?: string;
  category?: string;
  tag?: string;
  status?: string;
  author?: string;
  dateRange?: { start: string; end: string };
}

class FilterPanel {
  private filters: FilterState = {};
  private debounceTimer: number | null = null;

  constructor(panel: HTMLElement) {
    this.init(panel);
  }

  private init(panel: HTMLElement) {
    // 搜索框（防抖）
    const searchInput = panel.querySelector(
      "#filter-search",
    ) as HTMLInputElement;
    searchInput?.addEventListener("input", (e) => {
      this.debounce(() => {
        this.filters.search = (e.target as HTMLInputElement).value;
        this.applyFilters();
      }, 300);
    });

    // 下拉筛选
    const selects = panel.querySelectorAll("select[data-filter]");
    selects.forEach((select) => {
      select.addEventListener("change", (e) => {
        const key = select.dataset.filter!;
        this.filters[key] = (e.target as HTMLSelectElement).value;
        this.applyFilters();
      });
    });

    // 日期范围
    const startDate = panel.querySelector(
      "#filter-date-start",
    ) as HTMLInputElement;
    const endDate = panel.querySelector("#filter-date-end") as HTMLInputElement;
    startDate?.addEventListener("change", () => this.updateDateRange());
    endDate?.addEventListener("change", () => this.updateDateRange());

    // 清除筛选
    panel.querySelector("#filter-clear")?.addEventListener("click", () => {
      this.filters = {};
      this.resetUI();
      this.applyFilters();
    });
  }

  private debounce(fn: () => void, delay: number) {
    if (this.debounceTimer) clearTimeout(this.debounceTimer);
    this.debounceTimer = setTimeout(fn, delay);
  }

  private async applyFilters() {
    const params = new URLSearchParams();
    Object.entries(this.filters).forEach(([key, value]) => {
      if (value) params.set(key, String(value));
    });

    const response = await fetch(`/api/v1/admin/posts?${params}`);
    // 更新表格...
  }
}
```

---

## 18. 验证方案

1. **单元测试**：核心业务逻辑测试覆盖率 > 80%
2. **API 测试**：使用 httptest 测试所有 REST API
3. **集成测试**：完整流程测试（注册 → 发布文章 → 评论 → 审核）
4. **手动验证**：
   - 本地运行服务
   - 测试前台渲染、编辑器交互
   - 测试后台管理功能
   - 测试主题切换、插件启用
