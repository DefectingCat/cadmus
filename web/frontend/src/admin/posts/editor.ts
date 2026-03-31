// 文章编辑器增强功能
import { postsAPI, Post, UpdatePostRequest } from '../../api/posts'
import { $, $$, on, showMessage, confirm, getFormData, setFormData } from '../../utils/dom'

interface EditorState {
  postId: string | null
  post: Post | null
  autoSaveTimer: number | null
  lastContent: string
  isDirty: boolean
  isSaving: boolean
}

class PostEditorManager {
  private state: EditorState = {
    postId: null,
    post: null,
    autoSaveTimer: null,
    lastContent: '',
    isDirty: false,
    isSaving: false,
  }

  private elements: {
    form: HTMLFormElement | null
    titleInput: HTMLInputElement | null
    slugInput: HTMLInputElement | null
    contentEditor: Element | null
    excerptInput: HTMLTextAreaElement | null
    statusSelect: HTMLSelectElement | null
    categorySelect: HTMLSelectElement | null
    saveBtn: Element | null
    publishBtn: Element | null
    previewBtn: Element | null
    autoSaveIndicator: Element | null
    wordCount: Element | null
  } = {
    form: null,
    titleInput: null,
    slugInput: null,
    contentEditor: null,
    excerptInput: null,
    statusSelect: null,
    categorySelect: null,
    saveBtn: null,
    publishBtn: null,
    previewBtn: null,
    autoSaveIndicator: null,
    wordCount: null,
  }

  private readonly AUTO_SAVE_DELAY = 30000 // 30 秒自动保存

  constructor(postId?: string) {
    this.state.postId = postId || null
    this.init()
  }

  private init(): void {
    this.cacheElements()
    this.bindEvents()
    this.setupAutoSave()

    if (this.state.postId) {
      this.loadPost()
    }
  }

  private cacheElements(): void {
    this.elements.form = $('#post-form') as HTMLFormElement
    this.elements.titleInput = $('#title-input') as HTMLInputElement
    this.elements.slugInput = $('#slug-input') as HTMLInputElement
    this.elements.contentEditor = $('#content-editor')
    this.elements.excerptInput = $('#excerpt-input') as HTMLTextAreaElement
    this.elements.statusSelect = $('#status-select') as HTMLSelectElement
    this.elements.categorySelect = $('#category-select') as HTMLSelectElement
    this.elements.saveBtn = $('#save-btn')
    this.elements.publishBtn = $('#publish-btn')
    this.elements.previewBtn = $('#preview-btn')
    this.elements.autoSaveIndicator = $('#autosave-indicator')
    this.elements.wordCount = $('#word-count')
  }

  private bindEvents(): void {
    // 标题变化时自动生成 slug
    if (this.elements.titleInput && this.elements.slugInput) {
      on(this.elements.titleInput, 'input', () => {
        if (!this.state.postId) { // 仅新建时自动生成 slug
          const slug = this.generateSlug(this.elements.titleInput!.value)
          this.elements.slugInput!.value = slug
        }
      })
    }

    // 内容变化时标记 dirty 和更新字数
    if (this.elements.contentEditor) {
      // 监听输入事件（对于普通 textarea）
      if (this.elements.contentEditor instanceof HTMLTextAreaElement) {
        on(this.elements.contentEditor, 'input', () => {
          this.markDirty()
          this.updateWordCount()
        })
      }

      // 对于富文本编辑器，监听自定义事件
      on(this.elements.contentEditor, 'editor-change', () => {
        this.markDirty()
        this.updateWordCount()
      })
    }

    // 保存按钮
    if (this.elements.saveBtn) {
      on(this.elements.saveBtn, 'click', (e) => {
        e.preventDefault()
        this.savePost()
      })
    }

    // 发布按钮
    if (this.elements.publishBtn) {
      on(this.elements.publishBtn, 'click', (e) => {
        e.preventDefault()
        this.publishPost()
      })
    }

    // 预览按钮
    if (this.elements.previewBtn) {
      on(this.elements.previewBtn, 'click', (e) => {
        e.preventDefault()
        this.previewPost()
      })
    }

    // 页面离开提醒
    on(window, 'beforeunload', (e: Event) => {
      if (this.state.isDirty) {
        e.preventDefault()
        // returnValue 需要设置字符串值以触发提示
        ;(e as BeforeUnloadEvent).returnValue = '有未保存的内容，确定要离开吗？'
      }
    })

    // 快捷键保存 Ctrl+S / Cmd+S
    on(document, 'keydown', (e: Event) => {
      const keyEvent = e as KeyboardEvent
      if ((keyEvent.ctrlKey || keyEvent.metaKey) && keyEvent.key === 's') {
        e.preventDefault()
        this.savePost()
      }
    })
  }

  private setupAutoSave(): void {
    this.state.autoSaveTimer = window.setInterval(() => {
      if (this.state.isDirty && !this.state.isSaving) {
        this.autoSave()
      }
    }, this.AUTO_SAVE_DELAY)
  }

  private async loadPost(): Promise<void> {
    if (!this.state.postId) return

    const result = await postsAPI.get(this.state.postId)
    if (result.error) {
      showMessage('加载文章失败', 'error')
      return
    }

    this.state.post = result.data!
    this.populateForm()
    this.state.lastContent = this.getFormContent()
  }

  private populateForm(): void {
    if (!this.state.post || !this.elements.form) return

    const post = this.state.post
    setFormData(this.elements.form, {
      title: post.title,
      slug: post.slug,
      excerpt: post.excerpt || '',
      status: post.status,
      category_id: post.category_id || '',
      featured_image: post.featured_image || '',
      seo_title: post.seo_meta.title || '',
      seo_description: post.seo_meta.description || '',
      is_paid: String(post.is_paid),
      price: post.price?.toString() || '',
    })

    // 设置内容
    if (this.elements.contentEditor instanceof HTMLTextAreaElement) {
      this.elements.contentEditor.value = post.content
    }

    this.updateWordCount()
  }

  private async savePost(): Promise<void> {
    if (this.state.isSaving) return
    this.state.isSaving = true
    this.updateSaveIndicator('saving')

    const data = this.getFormData()

    try {
      if (this.state.postId) {
        // 更新现有文章
        const result = await postsAPI.update(this.state.postId, data)
        if (result.error) {
          showMessage(result.error.message, 'error')
          this.updateSaveIndicator('error')
          return
        }
        this.state.post = result.data!
        showMessage('文章已保存', 'success')
      } else {
        // 创建新文章
        const result = await postsAPI.create(data)
        if (result.error) {
          showMessage(result.error.message, 'error')
          this.updateSaveIndicator('error')
          return
        }
        this.state.post = result.data!
        this.state.postId = result.data!.id
        showMessage('文章已创建', 'success')
        // 更新 URL（不刷新页面）
        window.history.replaceState(null, '', `/admin/posts/${this.state.postId}/edit`)
      }

      this.state.lastContent = this.getFormContent()
      this.state.isDirty = false
      this.updateSaveIndicator('saved')
    } finally {
      this.state.isSaving = false
    }
  }

  private async autoSave(): Promise<void> {
    this.updateSaveIndicator('saving')
    await this.savePost()
  }

  private async publishPost(): Promise<void> {
    // 先保存
    if (this.state.isDirty) {
      await this.savePost()
    }

    if (!this.state.postId) return

    const confirmed = await confirm('确定要发布这篇文章吗？')
    if (!confirmed) return

    const result = await postsAPI.publish(this.state.postId)
    if (result.error) {
      showMessage(result.error.message, 'error')
      return
    }

    showMessage('文章已发布', 'success')
    if (this.elements.statusSelect) {
      this.elements.statusSelect.value = 'published'
    }
    this.state.isDirty = false
  }

  private previewPost(): void {
    // 打开预览窗口
    if (this.state.postId) {
      const previewUrl = `/posts/${this.state.post?.slug || this.state.postId}?preview=1`
      window.open(previewUrl, '_blank')
    }
  }

  private getFormData(): UpdatePostRequest {
    const form = this.elements.form
    if (!form) {
      return {
        title: '',
        slug: '',
        content: '',
        is_paid: false,
      }
    }

    const data = getFormData(form)

    let content = ''
    if (this.elements.contentEditor instanceof HTMLTextAreaElement) {
      content = this.elements.contentEditor.value
    }

    return {
      title: data.title || '',
      slug: data.slug || '',
      content,
      content_text: content.replace(/<[^>]*>/g, ''), // 提取纯文本
      excerpt: data.excerpt,
      category_id: data.category_id || undefined,
      status: data.status || 'draft',
      featured_image: data.featured_image || undefined,
      seo_title: data.seo_title || undefined,
      seo_description: data.seo_description || undefined,
      is_paid: data.is_paid === 'true',
      price: data.price ? parseFloat(data.price) : undefined,
    }
  }

  private getFormContent(): string {
    if (this.elements.contentEditor instanceof HTMLTextAreaElement) {
      return this.elements.contentEditor.value
    }
    return ''
  }

  private markDirty(): void {
    const currentContent = this.getFormContent()
    if (currentContent !== this.state.lastContent) {
      this.state.isDirty = true
      this.updateSaveIndicator('dirty')
    }
  }

  private updateSaveIndicator(status: 'saving' | 'saved' | 'error' | 'dirty'): void {
    if (!this.elements.autoSaveIndicator) return

    const messages = {
      saving: '正在保存...',
      saved: '已保存',
      error: '保存失败',
      dirty: '有未保存的更改',
    }

    const classes = {
      saving: 'text-yellow-600',
      saved: 'text-green-600',
      error: 'text-red-600',
      dirty: 'text-orange-600',
    }

    this.elements.autoSaveIndicator.textContent = messages[status]
    this.elements.autoSaveIndicator.className = `text-sm ${classes[status]}`
  }

  private updateWordCount(): void {
    if (!this.elements.wordCount || !this.elements.contentEditor) return

    let content = ''
    if (this.elements.contentEditor instanceof HTMLTextAreaElement) {
      content = this.elements.contentEditor.value
    }

    const text = content.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim()
    const wordCount = text.length
    const readTime = Math.ceil(wordCount / 300) // 假设每分钟阅读 300 字

    this.elements.wordCount.textContent = `${wordCount} 字 · 预计阅读 ${readTime} 分钟`
  }

  private generateSlug(title: string): string {
    return title
      .toLowerCase()
      .replace(/[^\w\s-]/g, '') // 移除特殊字符
      .replace(/\s+/g, '-') // 空格转为连字符
      .replace(/-+/g, '-') // 多个连字符合并
      .replace(/^-|-$/g, '') // 移除首尾连字符
      .slice(0, 100) // 截断
  }

  destroy(): void {
    if (this.state.autoSaveTimer) {
      clearInterval(this.state.autoSaveTimer)
    }
  }
}

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', () => {
  if ($('#post-form')) {
    // 从 URL 获取文章 ID
    const match = window.location.pathname.match(/\/admin\/posts\/([^/]+)\/edit/)
    const postId = match ? match[1] : undefined
    new PostEditorManager(postId)
  }
})

export { PostEditorManager }