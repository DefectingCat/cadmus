<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# admin/media/ - 媒体资源管理模块

## Purpose

本目录包含 Cadmus 管理后台的媒体资源管理功能，提供图片、PDF、文档等文件的上传、浏览、搜索、选择和删除功能。

核心功能：
- **拖拽上传** - 支持拖拽和点击上传，带实时进度条
- **媒体列表** - 网格视图展示，支持分页、搜索、类型筛选
- **媒体选择器** - 模态框选择媒体，插入到文章编辑器
- **批量操作** - 支持多选批量删除

技术栈：
- **原生 TypeScript** - 零框架依赖，模块化设计
- **TailwindCSS v4** - 响应式网格布局
- **XMLHttpRequest** - 带进度回调的文件上传

## Key Files

| File | Description |
|------|-------------|
| `index.ts` | 媒体管理页面入口，初始化上传器和媒体列表 |
| `upload.ts` | 媒体上传模块 - 拖拽上传、进度显示、多文件并行上传 |
| `list.ts` | 媒体列表模块 - 网格渲染、分页、搜索、媒体选择器 |

### index.ts 功能说明

```typescript
// 核心功能
- initMediaPage() - 初始化媒体管理页面，协调上传器和列表模块
- createMediaPicker() - 创建媒体选择器，供编辑器插入媒体使用
- showToast() - 显示上传成功/失败提示消息
```

### upload.ts 功能说明

```typescript
// MediaUploader 类
- handleFiles() - 处理拖拽/选择的多文件上传
- validateFile() - 验证文件大小 (10MB) 和类型
- uploadWithProgress() - XMLHttpRequest 上传，实时显示进度
- clearCompleted() - 清除已完成/失败的上传项

// 支持的文件类型
- 图片：image/jpeg, image/png, image/gif, image/webp, image/svg+xml
- 文档：application/pdf, application/msword, .docx
- 其他：application/zip, text/plain
```

### list.ts 功能说明

```typescript
// MediaList 类
- loadMedia() - 从 /api/v1/media 加载媒体列表
- renderMediaList() - 渲染网格视图，图片显示缩略图，文件显示图标
- openMediaDetail() - 打开详情弹窗，显示 URL、文件名、类型
- deleteMedia() / deleteSelected() - 删除单个或批量删除

// MediaPicker 类
- open() - 打开媒体选择器模态框
- loadMedia() - 加载最新 40 个媒体项
- onSelect() - 回调函数，返回选中的媒体信息

// MediaItem 接口
interface MediaItem {
  id: string;
  filename: string;
  original_name: string;
  url: string;
  mime_type: string;
  size: number;
  created_at: string;
}
```

## Subdirectories

无子目录。所有功能模块均在当前目录以独立文件形式组织。

## For AI Agents

### 开发指南

1. **模块初始化** - `index.ts` 自动检测 DOM 元素并初始化，无需手动调用

2. **API 端点** - 所有请求发送至 `/api/v1/media` 系列端点
   - `GET /api/v1/media` - 获取媒体列表（支持分页、搜索、类型筛选）
   - `POST /api/v1/media/upload` - 上传文件
   - `DELETE /api/v1/media/{id}` - 删除媒体

3. **认证方式** - 从 `localStorage` 获取 `auth_token`，通过 `Bearer` 头发送

4. **响应式布局** - 使用 `grid-cols-4` 网格，移动端需调整列数

### 添加新功能

```bash
# 1. 创建新模块文件
touch src/admin/media/gallery.ts  # 例如：画廊视图

# 2. 在 index.ts 中导入并初始化
import { createGallery } from './gallery';

# 3. 类型检查
bun run typecheck

# 4. 格式化
bun run fmt
```

### 构建输出

| Entry Point | Output Bundle |
|-------------|---------------|
| `admin/media/index.ts` | `web/frontend/static/dist/admin-media.js` |

### 模块依赖图

```
admin/media/index.ts
├── upload.ts (MediaUploader)
│   ├── 拖拽上传
│   ├── 进度显示
│   └── 文件验证
└── list.ts
    ├── MediaList (媒体列表)
    │   ├── 网格渲染
    │   ├── 搜索筛选
    │   ├── 分页
    │   └── 批量删除
    └── MediaPicker (媒体选择器)
        ├── 模态框选择
        └── 回调插入
```

### 编码规范

```typescript
// 1. 使用工厂函数创建实例
const uploader = createUploader({ dropzone, fileInput, ... });
const mediaList = createMediaList({ grid, searchInput, ... });

// 2. 回调函数优先
createUploader({
  onUploadComplete: (response: UploadResponse) => {
    // 上传完成后的处理
  }
});

// 3. 错误处理
try {
  const response = await fetch('/api/v1/media', {
    headers: getAuthHeaders(),
  });
  if (!response.ok) throw new Error('加载失败');
  const data: MediaListResponse = await response.json();
} catch (error) {
  console.error('加载媒体列表失败:', error);
}

// 4. 防抖搜索
let searchTimeout: number | null = null;
searchInput.addEventListener('input', () => {
  if (searchTimeout) clearTimeout(searchTimeout);
  searchTimeout = window.setTimeout(() => {
    loadMedia();
  }, 300);
});
```

### 常见问题

| 问题 | 解决方案 |
|------|----------|
| 上传失败 | 检查 `Authorization` header 是否包含有效 token |
| 列表不刷新 | 删除后手动调用 `loadMedia()` 或调整 `currentPage` |
| 选择器不关闭 | 确保调用 `picker.close()` 或点击关闭按钮 |
| 进度条不动 | 使用 XMLHttpRequest 而非 fetch 以支持进度回调 |
