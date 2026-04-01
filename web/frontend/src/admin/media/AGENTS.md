<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-01 | Updated: 2026-04-01 -->

# media

## Purpose
媒体管理模块，提供文件上传、媒体列表、媒体选择器功能。

## Key Files
| File | Description |
|------|-------------|
| `index.ts` | 模块入口：初始化上传器和列表 |
| `upload.ts` | 上传功能：拖拽上传、进度显示、多文件 |
| `list.ts` | 媒体列表：网格视图、分页、选择、删除 |

## For AI Agents

### Working In This Directory
- 上传使用 XMLHttpRequest 支持进度显示
- 列表支持搜索、类型筛选、分页
- MediaPicker 可供编辑器调用插入媒体

### Upload Module

**MediaUploader 类**:
```typescript
interface UploadProgress {
  id: string
  file: File
  progress: number
  status: 'pending' | 'uploading' | 'success' | 'error'
}
```

**功能**:
- 拖拽上传 + 点击选择
- 文件验证（10MB 限制、类型白名单）
- 进度条显示
- 多文件并行上传

### List Module

**MediaList 类**:
- 网格视图（图片缩略图/文件图标）
- 搜索防抖（300ms）
- 类型筛选（image/document/...）
- 分页导航
- 批量选择和删除

**MediaPicker 类**:
- 模态框选择器
- 供编辑器调用插入媒体
- 回调 `onSelect(media)`

### File Validation
```typescript
const allowedTypes = [
  'image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml',
  'application/pdf', 'application/msword', ...
]
const maxSize = 10 * 1024 * 1024  // 10MB
```

### Dependencies

**Internal**:
- 类型定义: `MediaItem`, `UploadResponse`

<!-- MANUAL: -->