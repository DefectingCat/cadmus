<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# posts/ - 文章管理模块

## Purpose

本目录包含 Cadmus 管理后台的文章管理功能前端实现，负责文章的列表展示、创建、编辑、删除、发布等核心功能。

技术栈：
- **原生 TypeScript** - 零框架依赖
- **模块化设计** - 列表与编辑器分离
- **API 层封装** - 所有请求通过 `src/api/posts.ts` 统一处理

## Key Files

| File | Description |
|------|-------------|
| `list.ts` | 文章列表管理 - `PostListManager` 类，处理列表渲染、分页、筛选、批量操作 |
| `editor.ts` | 文章编辑器管理 - `PostEditorManager` 类，处理表单编辑、自动保存、发布预览 |

### list.ts 功能说明

```typescript
class PostListManager {
  // 核心功能
  - loadPosts()        - 加载文章列表，支持分页和筛选
  - renderPosts()      - 渲染文章列表表格
  - renderPagination() - 渲染分页控件（支持页码省略）
  - deletePost()       - 删除单篇文章
  - publishPost()      - 发布单篇文章
  - bulkDelete()       - 批量删除
  - bulkPublish()      - 批量发布

  // 状态管理
  - state: ListState {
      posts: Post[],
      total: number,
      page: number,
      pageSize: number,
      filters: PostListFilters,
      selectedIds: Set<string>,
      loading: boolean
    }
}
```

### editor.ts 功能说明

```typescript
class PostEditorManager {
  // 核心功能
  - loadPost()         - 加载文章详情
  - savePost()         - 保存文章（新建/更新）
  - publishPost()      - 发布文章
  - previewPost()      - 预览文章
  - autoSave()         - 自动保存（30 秒间隔）

  // 用户体验
  - markDirty()        - 标记内容变更
  - updateWordCount()  - 更新字数统计和阅读时间
  - generateSlug()     - 根据标题自动生成 slug

  // 状态管理
  - state: EditorState {
      postId: string | null,
      post: Post | null,
      autoSaveTimer: number | null,
      isDirty: boolean,
      isSaving: boolean
    }
}
```

## Article Status Values

| Status | Label | CSS Class |
|--------|-------|-----------|
| `published` | 已发布 | `bg-green-100 text-green-800` |
| `draft` | 草稿 | `bg-gray-100 text-gray-800` |
| `scheduled` | 定时发布 | `bg-blue-100 text-blue-800` |
| `private` | 私密 | `bg-yellow-100 text-yellow-800` |

## For AI Agents

### 开发指南

1. **列表页面开发** - 使用 `PostListManager` 处理列表逻辑，DOM 元素通过 ID 缓存
2. **编辑器开发** - 使用 `PostEditorManager` 处理编辑逻辑，支持 Ctrl+S 快捷键保存
3. **API 调用** - 所有后端请求必须通过 `postsAPI`（`src/api/posts.ts`）
4. **类型安全** - 使用 `Post` 和 `UpdatePostRequest` 类型定义

### 添加新字段

```typescript
// 1. 在 API 类型定义中添加字段（src/api/posts.ts）
interface Post {
  // ... 现有字段
  newField: string;
}

// 2. 在 list.ts 中渲染新字段
<td class="px-4 py-3">${post.newField}</td>

// 3. 在 editor.ts 表单中添加输入
<input id="new-field" name="newField">

// 4. 在 getFormData() 中包含新字段
return {
  // ...
  new_field: data.newField,
};
```

### 批量操作扩展

```typescript
// 在 list.ts 中添加新的批量操作
private async bulkAction(): Promise<void> {
  const ids = Array.from(this.state.selectedIds);
  if (ids.length === 0) return;

  const confirmed = await confirm(`确认对 ${ids.length} 篇文章执行操作？`);
  if (!confirmed) return;

  const result = await postsAPI.batchAction(ids);
  if (result.error) {
    showMessage(result.error.message, "error");
    return;
  }

  showMessage(result.data!.message, "success");
  this.state.selectedIds.clear();
  this.loadPosts();
}
```

### 构建输出

| Entry Point | Output Bundle |
|-------------|---------------|
| `admin/posts/list.ts` | `web/frontend/static/dist/admin/posts.js` |
| `admin/posts/editor.ts` | `web/frontend/static/dist/admin/post-edit.js` |

### 调试技巧

```bash
# 查看构建输出
bun run build:admin

# 类型检查
bun run typecheck

# 监听模式开发
bun run dev:admin
```
