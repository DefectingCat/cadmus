<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# utils - 前端工具函数目录

## Purpose

本目录包含前端通用的工具函数，提供 DOM 操作、事件处理、表单数据收集等基础能力。所有工具函数均为纯 TypeScript 实现，零外部依赖，可在任何前端模块中复用。

## Key Files

| File | Description |
|------|-------------|
| `dom.ts` | DOM 操作工具函数集 |

## Directory Details

### dom.ts

DOM 操作工具函数，提供以下功能：

#### 元素选择

| Function | Description |
|----------|-------------|
| `$` | 查询单个元素 (`querySelector`) |
| `$$` | 查询多个元素 (`querySelectorAll`)，返回数组 |

#### 元素创建

| Function | Description |
|----------|-------------|
| `createElement<T>` | 创建 HTML 元素，支持属性设置和子节点添加 |

#### 显示控制

| Function | Description |
|----------|-------------|
| `show` | 显示元素（移除 `hidden` 类） |
| `hide` | 隐藏元素（添加 `hidden` 类） |
| `toggle` | 切换显示状态 |

#### 类名操作

| Function | Description |
|----------|-------------|
| `addClass` | 添加 CSS 类 |
| `removeClass` | 移除 CSS 类 |
| `hasClass` | 检查是否包含 CSS 类 |

#### 事件处理

| Function | Description |
|----------|-------------|
| `on` | 绑定事件监听器 |
| `off` | 移除事件监听器 |
| `delegate` | 事件委托（支持选择器匹配） |

#### 表单工具

| Function | Description |
|----------|-------------|
| `getFormData` | 收集表单数据为对象 |
| `setFormData` | 设置表单字段值 |

#### UI 组件

| Function | Description |
|----------|-------------|
| `showMessage` | 显示消息提示（支持 success/error/warning/info） |
| `confirm` | 自定义确认对话框（返回 Promise） |

## For AI Agents

### 使用示例

```typescript
import { 
  $, $$, createElement, on, off, delegate,
  getFormData, showMessage, confirm
} from './dom';

// 元素选择
const btn = $('#submit-btn');
const items = $$('.list-item');

// 创建元素
const el = createElement<HTMLDivElement>('div', { 
  className: 'container',
  id: 'main'
}, 'Hello World');

// 事件绑定
on(btn, 'click', (e) => {
  console.log('clicked');
});

// 事件委托
delegate(document, '.delete-btn', 'click', (e, target) => {
  // 处理删除按钮点击
});

// 表单操作
const formData = getFormData(formEl);
setFormData(formEl, { name: 'test', active: true });

// 消息提示
showMessage('操作成功', 'success');

// 确认对话框
const result = await confirm('确定要删除吗？');
```

### 开发指南

1. **纯函数优先** - 工具函数应保持纯函数特性，避免副作用
2. **类型安全** - 所有函数必须有完整的 TypeScript 类型定义
3. **零依赖** - 不引入外部库，保持工具函数的轻量性
4. **可复用性** - 工具函数应通用，不绑定特定业务逻辑
5. **文档完整** - 新增函数必须在本文档中说明

### 添加工具函数

```typescript
// 1. 在 dom.ts 中导出新函数
export const newUtil = (param: string): void => {
  // 实现
};

// 2. 在本文档的 Key Files 表格中更新说明

// 3. 运行类型检查
bun run typecheck
```

### 注意事项

- `createElement` 的 `dataset` 属性需要特殊处理（当前未实现）
- `confirm` 对话框会阻塞 UI，返回 Promise 而非同步阻塞
- `showMessage` 默认 3 秒后自动消失
- 所有选择器函数支持传入自定义 parent 元素

<!-- MANUAL: -->
