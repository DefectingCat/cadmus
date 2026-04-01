<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-01 | Updated: 2026-04-01 -->

# comments

## Purpose
评论管理模块，提供评论审核、批量操作功能。

## Key Files
| File | Description |
|------|-------------|
| `list.ts` | 评论列表：选择、批量操作、状态更新 |

## For AI Agents

### Working In This Directory
- 评论管理支持批量操作：批准、拒绝、删除
- 使用确认对话框防止误操作
- Toast 提示操作结果

### State Management
```typescript
const state = {
  selectedIds: new Set<string>(),
  currentStatus: 'pending',  // 当前筛选状态
}
```

### Operations

| Action | Function | API Call |
|--------|----------|----------|
| 单个批准 | `handleSingleApprove` | `commentsApi.approveComment(id)` |
| 单个拒绝 | `handleSingleReject` | `commentsApi.rejectComment(id)` |
| 单个删除 | `handleSingleDelete` | `commentsApi.deleteComment(id)` |
| 批量批准 | `handleConfirmAction('batch-approve')` | `commentsApi.batchApproveComments(ids)` |
| 批量拒绝 | `handleConfirmAction('batch-reject')` | `commentsApi.batchRejectComments(ids)` |
| 批量删除 | `handleConfirmAction('batch-delete')` | `commentsApi.batchDeleteComments(ids)` |

### UI Components
- 全选复选框 + 单行复选框
- 批量操作工具栏（根据选择数量显示/隐藏）
- 确认对话框（模态框）
- Toast 提示

### Dependencies

**Internal**:
- `../../api/comments` - 评论 API 客户端

<!-- MANUAL: -->