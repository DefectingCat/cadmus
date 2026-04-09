<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# 评论管理模块 - AI Agent 开发指南

## 模块用途

`comments` 目录是后台管理系统中的评论管理模块，负责处理用户评论的审核、批量操作和状态管理。

## 关键文件

### list.ts

评论管理交互逻辑的核心文件。

**主要功能：**

- **状态管理**
  - `selectedIds`: 选中的评论 ID 集合
  - `currentStatus`: 当前筛选状态（pending/approved/rejected）

- **DOM 元素引用**
  - 全选复选框、批量操作按钮（批准/拒绝/删除）
  - 确认对话框元素（标题、消息、确认/取消按钮）
  - 已选择数量显示

- **事件绑定**
  - 全选/取消全选
  - 单个评论复选框选择
  - 批量操作按钮（批准/拒绝/删除）
  - 单个评论操作按钮
  - 对话框交互

- **批量操作**
  - `toggleSelectAll()`: 切换全选状态
  - `toggleAllCheckboxes()`: 切换所有复选框
  - `updateBatchToolbar()`: 更新批量工具栏状态
  - `showConfirmDialog()`: 显示确认对话框
  - `handleConfirmAction()`: 处理确认操作（调用 API）

- **单个操作**
  - `handleSingleApprove()`: 批准单条评论
  - `handleSingleReject()`: 拒绝单条评论
  - `handleSingleDelete()`: 删除单条评论

- **API 调用**
  - 导入 `../../api/comments` 模块
  - 调用 `approveComment`, `rejectComment`, `deleteComment`
  - 调用 `batchApproveComments`, `batchRejectComments`, `batchDeleteComments`

- **Toast 提示**
  - `showSuccessToast()`: 成功提示
  - `showErrorToast()`: 错误提示

**开发注意事项：**

1. 所有操作按钮通过 `data-id` 属性传递评论 ID
2. 批量操作前显示确认对话框，防止误操作
3. 操作完成后自动刷新页面以反映最新状态
4. 错误处理通过 try-catch 捕获并显示 Toast 提示

## 子目录

无子目录。

## 相关 API

- `../../api/comments` - 评论相关 API 调用模块
