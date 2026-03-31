// DOM 工具函数
export const $ = (selector: string, parent: Element | Document = document): Element | null => {
  return parent.querySelector(selector)
}

export const $$ = (selector: string, parent: Element | Document = document): Element[] => {
  return Array.from(parent.querySelectorAll(selector))
}

export const createElement = <T extends HTMLElement>(
  tag: string,
  attrs?: Record<string, string>,
  children?: string | Element | Element[]
): T => {
  const el = document.createElement(tag) as T
  if (attrs) {
    for (const [key, value] of Object.entries(attrs)) {
      if (key === 'className') {
        el.className = value
      } else if (key === 'dataset') {
        // dataset 需要特殊处理
      } else {
        el.setAttribute(key, value)
      }
    }
  }
  if (children) {
    if (typeof children === 'string') {
      el.textContent = children
    } else if (Array.isArray(children)) {
      children.forEach(child => el.appendChild(child))
    } else {
      el.appendChild(children)
    }
  }
  return el
}

export const show = (el: Element): void => {
  el.classList.remove('hidden')
}

export const hide = (el: Element): void => {
  el.classList.add('hidden')
}

export const toggle = (el: Element): void => {
  el.classList.toggle('hidden')
}

export const addClass = (el: Element, className: string): void => {
  el.classList.add(className)
}

export const removeClass = (el: Element, className: string): void => {
  el.classList.remove(className)
}

export const hasClass = (el: Element, className: string): boolean => {
  return el.classList.contains(className)
}

// 事件绑定辅助函数
export const on = (
  el: Element | Window | Document,
  event: string,
  handler: EventListener,
  options?: AddEventListenerOptions
): void => {
  el.addEventListener(event, handler, options)
}

export const off = (
  el: Element | Window | Document,
  event: string,
  handler: EventListener
): void => {
  el.removeEventListener(event, handler)
}

export const delegate = (
  parent: Element,
  selector: string,
  event: string,
  handler: (e: Event, target: Element) => void
): void => {
  on(parent, event, (e) => {
    const target = (e.target as Element).closest(selector)
    if (target && parent.contains(target)) {
      handler(e, target)
    }
  })
}

// 表单数据收集
export const getFormData = (form: HTMLFormElement): Record<string, string> => {
  const data: Record<string, string> = {}
  const inputs = form.querySelectorAll<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>(
    'input[name], select[name], textarea[name]'
  )
  inputs.forEach(input => {
    if (input.type === 'checkbox') {
      data[input.name] = (input as HTMLInputElement).checked ? 'true' : 'false'
    } else {
      data[input.name] = input.value
    }
  })
  return data
}

export const setFormData = (form: HTMLFormElement, data: Record<string, string | boolean>): void => {
  const inputs = form.querySelectorAll<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>(
    'input[name], select[name], textarea[name]'
  )
  inputs.forEach(input => {
    const value = data[input.name]
    if (value !== undefined) {
      if (input.type === 'checkbox') {
        (input as HTMLInputElement).checked = value === true || value === 'true'
      } else {
        input.value = String(value)
      }
    }
  })
}

// 显示消息提示
export const showMessage = (message: string, type: 'success' | 'error' | 'warning' | 'info' = 'info'): void => {
  // 创建消息元素
  const msgEl = createElement('div', {
    className: `fixed top-4 right-4 px-4 py-3 rounded-lg shadow-lg z-50 transition-all duration-300 ${
      type === 'success' ? 'bg-green-500 text-white' :
      type === 'error' ? 'bg-red-500 text-white' :
      type === 'warning' ? 'bg-yellow-500 text-white' :
      'bg-blue-500 text-white'
    }`,
  }, message)

  document.body.appendChild(msgEl)

  // 3 秒后消失
  setTimeout(() => {
    msgEl.style.opacity = '0'
    setTimeout(() => msgEl.remove(), 300)
  }, 3000)
}

// 确认对话框
export const confirm = (message: string): Promise<boolean> => {
  return new Promise((resolve) => {
    const dialog = createElement('div', {
      className: 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50',
    })

    const box = createElement('div', {
      className: 'bg-white rounded-lg shadow-xl p-6 max-w-sm mx-4',
    })

    const text = createElement('p', { className: 'text-gray-800 mb-4' }, message)
    const btnGroup = createElement('div', { className: 'flex justify-end gap-2' })

    const cancelBtn = createElement('button', {
      className: 'px-4 py-2 text-gray-600 hover:text-gray-800 rounded',
    }, '取消')
    const okBtn = createElement('button', {
      className: 'px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700',
    }, '确认')

    on(cancelBtn, 'click', () => {
      dialog.remove()
      resolve(false)
    })
    on(okBtn, 'click', () => {
      dialog.remove()
      resolve(true)
    })

    btnGroup.appendChild(cancelBtn)
    btnGroup.appendChild(okBtn)
    box.appendChild(text)
    box.appendChild(btnGroup)
    dialog.appendChild(box)
    document.body.appendChild(dialog)
  })
}