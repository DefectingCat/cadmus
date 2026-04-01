<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-01 | Updated: 2026-04-01 -->

# utils

## Purpose
前端工具函数库，提供 DOM 操作、事件处理、表单处理等通用功能。

## Key Files
| File | Description |
|------|-------------|
| `dom.ts` | DOM 工具函数：选择器、事件、表单、消息提示 |

## For AI Agents

### Working In This Directory
- 所有函数为纯函数，无副作用
- 使用 TypeScript 类型标注
- 提供事件委托、表单序列化等常用模式

### Exported Functions

**DOM 选择器**:
```typescript
$(selector, parent?)      // querySelector
$$(selector, parent?)     // querySelectorAll
createElement(tag, attrs?, children?) // 创建元素
```

**DOM 操作**:
```typescript
show(el), hide(el), toggle(el)
addClass(el, cls), removeClass(el, cls), hasClass(el, cls)
```

**事件处理**:
```typescript
on(el, event, handler, options?)    // 绑定事件
off(el, event, handler)             // 解绑事件
delegate(parent, selector, event, handler) // 事件委托
```

**表单处理**:
```typescript
getFormData(form)    // 收集表单数据为对象
setFormData(form, data) // 填充表单数据
```

**用户反馈**:
```typescript
showMessage(message, type?) // Toast 提示 (success/error/warning/info)
confirm(message)            // Promise 化的确认对话框
```

### Usage Pattern
```typescript
import { $, $$, on, delegate, showMessage, confirm } from './utils/dom'

// 简单选择
const form = $('#post-form') as HTMLFormElement

// 事件委托
delegate(tableBody, '.delete-btn', 'click', async (e, target) => {
    const id = target.dataset.id
    if (await confirm('确定删除?')) {
        // 执行删除
    }
})
```

<!-- MANUAL: -->