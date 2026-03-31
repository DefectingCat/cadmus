<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# admin

## Purpose
后台管理模块，包含文章、评论、媒体管理功能。

## Key Files
| File | Description |
|------|-------------|
| `main.ts` | 后台入口，初始化所有管理模块 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `posts/` | 文章管理：列表、编辑器 |
| `comments/` | 评论管理：审核、批量操作 |
| `media/` | 媒体管理：上传、列表 |

## For AI Agents

### Working In This Directory
- 每个子目录独立模块
- 通过 `main.ts` 统一导入

### Module Structure
```typescript
// admin/main.ts
import { initPostList } from './posts/list';
import { initPostEditor } from './posts/editor';
import { initCommentList } from './comments/list';
import { initMediaUpload } from './media/upload';
import { initMediaList } from './media/list';

// 根据页面初始化对应模块
document.addEventListener('DOMContentLoaded', () => {
    const page = document.body.dataset.page;
    switch (page) {
        case 'admin-posts':
            initPostList();
            break;
        case 'admin-comments':
            initCommentList();
            break;
        // ...
    }
});
```

### Common Patterns
| Pattern | Description |
|---------|-------------|
| 批量操作 | 选择行 → 执行动作 → 刷新列表 |
| 分页 | URL 参数 `page` → API 调用 → 更新表格 |
| 筛选 | 表单控件 → 构建查询参数 → API 调用 |

<!-- MANUAL: -->