<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-01 | Updated: 2026-04-01 -->

# posts

## Purpose
文章管理模块，提供文章列表和编辑器功能。

## Key Files
| File | Description |
|------|-------------|
| `list.ts` | 文章列表：分页、筛选、批量操作 |
| `editor.ts` | 文章编辑器：表单、自动保存、发布 |

## For AI Agents

### Working In This Directory
- 列表支持搜索、状态筛选、批量操作
- 编辑器支持自动保存（30秒）、快捷键保存
- 使用状态管理模式管理编辑器状态

### List Module

**PostListManager 类**:
```typescript
interface ListState {
  posts: Post[]
  total: number
  page: number
  pageSize: number
  filters: PostListFilters
  selectedIds: Set<string>
  loading: boolean
}
```

**功能**:
- 搜索防抖（300ms）
- 状态筛选（published/draft/scheduled/private）
- 分页（显示页码导航）
- 批量删除、批量发布
- 单行删除、单行发布

### Editor Module

**PostEditorManager 类**:
```typescript
interface EditorState {
  postId: string | null
  post: Post | null
  autoSaveTimer: number | null
  lastContent: string
  isDirty: boolean
  isSaving: boolean
}
```

**功能**:
- 标题 → Slug 自动生成
- 自动保存（30秒间隔）
- 快捷键保存（Ctrl/Cmd + S）
- 发布确认
- 预览功能
- 页面离开提醒
- 字数统计 + 阅读时间估算

### Status Labels
| Status | Label | CSS Class |
|--------|-------|-----------|
| published | 已发布 | `bg-green-100 text-green-800` |
| draft | 草稿 | `bg-gray-100 text-gray-800` |
| scheduled | 定时发布 | `bg-blue-100 text-blue-800` |
| private | 私密 | `bg-yellow-100 text-yellow-800` |

### Dependencies

**Internal**:
- `../../api/posts` - 文章 API 客户端
- `../../utils/dom` - DOM 工具函数

<!-- MANUAL: -->