// 文章列表页面交互
import { postsAPI, Post, PostListFilters } from '../../api/posts'
import { $, $$, on, delegate, showMessage, confirm } from '../../utils/dom'

interface ListState {
  posts: Post[]
  total: number
  page: number
  pageSize: number
  filters: PostListFilters
  selectedIds: Set<string>
  loading: boolean
}

class PostListManager {
  private state: ListState = {
    posts: [],
    total: 0,
    page: 1,
    pageSize: 20,
    filters: {},
    selectedIds: new Set(),
    loading: false,
  }

  private elements: {
    tableBody: Element | null
    pagination: Element | null
    searchInput: HTMLInputElement | null
    statusFilter: HTMLSelectElement | null
    bulkActions: Element | null
    selectAllCheckbox: HTMLInputElement | null
    emptyState: Element | null
    loadingOverlay: Element | null
  } = {
    tableBody: null,
    pagination: null,
    searchInput: null,
    statusFilter: null,
    bulkActions: null,
    selectAllCheckbox: null,
    emptyState: null,
    loadingOverlay: null,
  }

  constructor() {
    this.init()
  }

  private init(): void {
    this.cacheElements()
    this.bindEvents()
    this.loadPosts()
  }

  private cacheElements(): void {
    this.elements.tableBody = $('#posts-table-body')
    this.elements.pagination = $('#pagination')
    this.elements.searchInput = $('#search-input') as HTMLInputElement
    this.elements.statusFilter = $('#status-filter') as HTMLSelectElement
    this.elements.bulkActions = $('#bulk-actions')
    this.elements.selectAllCheckbox = $('#select-all') as HTMLInputElement
    this.elements.emptyState = $('#empty-state')
    this.elements.loadingOverlay = $('#loading-overlay')
  }

  private bindEvents(): void {
    // 搜索事件
    if (this.elements.searchInput) {
      let searchTimeout: number
      on(this.elements.searchInput, 'input', () => {
        clearTimeout(searchTimeout)
        searchTimeout = window.setTimeout(() => {
          this.state.filters.search = this.elements.searchInput!.value
          this.state.page = 1
          this.loadPosts()
        }, 300)
      })
    }

    // 状态筛选
    if (this.elements.statusFilter) {
      on(this.elements.statusFilter, 'change', () => {
        this.state.filters.status = this.elements.statusFilter!.value || undefined
        this.state.page = 1
        this.loadPosts()
      })
    }

    // 分页事件
    if (this.elements.pagination) {
      delegate(this.elements.pagination, '.page-link', 'click', (e, target) => {
        e.preventDefault()
        const page = target.getAttribute('data-page')
        if (page) {
          this.state.page = parseInt(page, 10)
          this.loadPosts()
        }
      })
    }

    // 全选事件
    if (this.elements.selectAllCheckbox) {
      on(this.elements.selectAllCheckbox, 'change', () => {
        const checked = this.elements.selectAllCheckbox!.checked
        this.state.posts.forEach(post => {
          if (checked) {
            this.state.selectedIds.add(post.id)
          } else {
            this.state.selectedIds.delete(post.id)
          }
        })
        this.updateCheckboxStates()
        this.updateBulkActions()
      })
    }

    // 单行选择事件
    if (this.elements.tableBody) {
      delegate(this.elements.tableBody, '.post-checkbox', 'change', (e, target) => {
        const checkbox = target as HTMLInputElement
        const postId = checkbox.getAttribute('data-id')
        if (postId) {
          if (checkbox.checked) {
            this.state.selectedIds.add(postId)
          } else {
            this.state.selectedIds.delete(postId)
          }
          this.updateBulkActions()
          this.updateSelectAllState()
        }
      })
    }

    // 批量删除
    const bulkDeleteBtn = $('#bulk-delete')
    if (bulkDeleteBtn) {
      on(bulkDeleteBtn, 'click', () => this.bulkDelete())
    }

    // 批量发布
    const bulkPublishBtn = $('#bulk-publish')
    if (bulkPublishBtn) {
      on(bulkPublishBtn, 'click', () => this.bulkPublish())
    }

    // 单行删除
    if (this.elements.tableBody) {
      delegate(this.elements.tableBody, '.delete-btn', 'click', async (e, target) => {
        e.preventDefault()
        const postId = target.getAttribute('data-id')
        if (postId) {
          await this.deletePost(postId)
        }
      })
    }

    // 单行发布
    if (this.elements.tableBody) {
      delegate(this.elements.tableBody, '.publish-btn', 'click', async (e, target) => {
        e.preventDefault()
        const postId = target.getAttribute('data-id')
        if (postId) {
          await this.publishPost(postId)
        }
      })
    }
  }

  private async loadPosts(): Promise<void> {
    if (this.state.loading) return
    this.state.loading = true
    this.showLoading()

    const result = await postsAPI.list({
      ...this.state.filters,
      page: this.state.page,
      page_size: this.state.pageSize,
    })

    this.state.loading = false
    this.hideLoading()

    if (result.error) {
      showMessage(result.error.message, 'error')
      return
    }

    this.state.posts = result.data!.posts
    this.state.total = result.data!.total
    this.renderPosts()
    this.renderPagination()
    this.updateBulkActions()
  }

  private renderPosts(): void {
    if (!this.elements.tableBody) return

    if (this.state.posts.length === 0) {
      this.elements.tableBody.innerHTML = ''
      if (this.elements.emptyState) {
        this.elements.emptyState.classList.remove('hidden')
      }
      return
    }

    if (this.elements.emptyState) {
      this.elements.emptyState.classList.add('hidden')
    }

    this.elements.tableBody.innerHTML = this.state.posts.map(post => `
      <tr class="hover:bg-gray-50 ${this.state.selectedIds.has(post.id) ? 'bg-blue-50' : ''}">
        <td class="px-4 py-3">
          <input type="checkbox" class="post-checkbox rounded" data-id="${post.id}"
            ${this.state.selectedIds.has(post.id) ? 'checked' : ''}>
        </td>
        <td class="px-4 py-3">
          <a href="/admin/posts/${post.id}/edit" class="text-blue-600 hover:text-blue-800 font-medium">
            ${this.escapeHtml(post.title)}
          </a>
          ${post.is_paid ? '<span class="ml-2 text-xs text-orange-600 bg-orange-100 px-1 rounded">付费</span>' : ''}
        </td>
        <td class="px-4 py-3 text-sm text-gray-500">${post.slug}</td>
        <td class="px-4 py-3">
          <span class="px-2 py-1 text-xs rounded-full ${this.getStatusClass(post.status)}">
            ${this.getStatusLabel(post.status)}
          </span>
        </td>
        <td class="px-4 py-3 text-sm text-gray-500">${post.view_count}</td>
        <td class="px-4 py-3 text-sm text-gray-500">${post.comment_count}</td>
        <td class="px-4 py-3 text-sm text-gray-500">${this.formatDate(post.created_at)}</td>
        <td class="px-4 py-3">
          <div class="flex gap-2">
            ${post.status !== 'published' ? `
              <button class="publish-btn text-green-600 hover:text-green-800 text-sm" data-id="${post.id}">
                发布
              </button>
            ` : ''}
            <a href="/posts/${post.slug}" target="_blank" class="text-blue-600 hover:text-blue-800 text-sm">
              查看
            </a>
            <button class="delete-btn text-red-600 hover:text-red-800 text-sm" data-id="${post.id}">
              删除
            </button>
          </div>
        </td>
      </tr>
    `).join('')
  }

  private renderPagination(): void {
    if (!this.elements.pagination) return

    const totalPages = Math.ceil(this.state.total / this.state.pageSize)
    if (totalPages <= 1) {
      this.elements.pagination.innerHTML = ''
      return
    }

    const pages: (number | string)[] = []
    const current = this.state.page

    // 生成页码列表
    if (totalPages <= 7) {
      for (let i = 1; i <= totalPages; i++) pages.push(i)
    } else {
      if (current <= 3) {
        pages.push(1, 2, 3, 4, '...', totalPages)
      } else if (current >= totalPages - 3) {
        pages.push(1, '...', totalPages - 3, totalPages - 2, totalPages - 1, totalPages)
      } else {
        pages.push(1, '...', current - 1, current, current + 1, '...', totalPages)
      }
    }

    this.elements.pagination.innerHTML = `
      <nav class="flex justify-center items-center gap-1">
        ${current > 1 ? `
          <a href="#" class="page-link px-3 py-1 rounded border hover:bg-gray-100" data-page="${current - 1}">
            上一页
          </a>
        ` : `
          <span class="px-3 py-1 rounded border text-gray-400 cursor-not-allowed">上一页</span>
        `}
        ${pages.map(p => {
          if (p === '...') {
            return '<span class="px-3 py-1">...</span>'
          }
          if (p === current) {
            return `<span class="px-3 py-1 rounded bg-blue-600 text-white">${p}</span>`
          }
          return `<a href="#" class="page-link px-3 py-1 rounded border hover:bg-gray-100" data-page="${p}">${p}</a>`
        }).join('')}
        ${current < totalPages ? `
          <a href="#" class="page-link px-3 py-1 rounded border hover:bg-gray-100" data-page="${current + 1}">
            下一页
          </a>
        ` : `
          <span class="px-3 py-1 rounded border text-gray-400 cursor-not-allowed">下一页</span>
        `}
      </nav>
    `
  }

  private updateCheckboxStates(): void {
    $$('.post-checkbox').forEach(checkbox => {
      const id = checkbox.getAttribute('data-id')
      if (id) {
        (checkbox as HTMLInputElement).checked = this.state.selectedIds.has(id)
      }
    })
    // 更新行样式
    $$('tr', this.elements.tableBody!).forEach(row => {
      const checkbox = $('.post-checkbox', row)
      const id = checkbox?.getAttribute('data-id')
      if (id && this.state.selectedIds.has(id)) {
        row.classList.add('bg-blue-50')
      } else {
        row.classList.remove('bg-blue-50')
      }
    })
  }

  private updateSelectAllState(): void {
    if (!this.elements.selectAllCheckbox) return
    const allSelected = this.state.posts.length > 0 &&
      this.state.posts.every(p => this.state.selectedIds.has(p.id))
    this.elements.selectAllCheckbox.checked = allSelected
    this.elements.selectAllCheckbox.indeterminate =
      this.state.selectedIds.size > 0 && !allSelected
  }

  private updateBulkActions(): void {
    if (!this.elements.bulkActions) return
    const count = this.state.selectedIds.size
    if (count > 0) {
      this.elements.bulkActions.classList.remove('hidden')
      const countEl = $('#selected-count')
      if (countEl) countEl.textContent = String(count)
    } else {
      this.elements.bulkActions.classList.add('hidden')
    }
  }

  private async deletePost(id: string): Promise<void> {
    const confirmed = await confirm('确定要删除这篇文章吗？')
    if (!confirmed) return

    const result = await postsAPI.delete(id)
    if (result.error) {
      showMessage(result.error.message, 'error')
      return
    }

    showMessage('文章已删除', 'success')
    this.loadPosts()
  }

  private async publishPost(id: string): Promise<void> {
    const result = await postsAPI.publish(id)
    if (result.error) {
      showMessage(result.error.message, 'error')
      return
    }

    showMessage('文章已发布', 'success')
    this.loadPosts()
  }

  private async bulkDelete(): Promise<void> {
    const ids = Array.from(this.state.selectedIds)
    if (ids.length === 0) return

    const confirmed = await confirm(`确定要删除 ${ids.length} 篇文章吗？`)
    if (!confirmed) return

    const result = await postsAPI.batchDelete(ids)
    if (result.error) {
      showMessage(result.error.message, 'error')
      return
    }

    showMessage(result.data!.message, 'success')
    this.state.selectedIds.clear()
    this.loadPosts()
  }

  private async bulkPublish(): Promise<void> {
    const ids = Array.from(this.state.selectedIds)
    if (ids.length === 0) return

    const result = await postsAPI.batchPublish(ids)
    if (result.error) {
      showMessage(result.error.message, 'error')
      return
    }

    showMessage(result.data!.message, 'success')
    this.state.selectedIds.clear()
    this.loadPosts()
  }

  private showLoading(): void {
    if (this.elements.loadingOverlay) {
      this.elements.loadingOverlay.classList.remove('hidden')
    }
  }

  private hideLoading(): void {
    if (this.elements.loadingOverlay) {
      this.elements.loadingOverlay.classList.add('hidden')
    }
  }

  private getStatusClass(status: string): string {
    switch (status) {
      case 'published': return 'bg-green-100 text-green-800'
      case 'draft': return 'bg-gray-100 text-gray-800'
      case 'scheduled': return 'bg-blue-100 text-blue-800'
      case 'private': return 'bg-yellow-100 text-yellow-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  private getStatusLabel(status: string): string {
    switch (status) {
      case 'published': return '已发布'
      case 'draft': return '草稿'
      case 'scheduled': return '定时发布'
      case 'private': return '私密'
      default: return status
    }
  }

  private formatDate(dateStr: string): string {
    const date = new Date(dateStr)
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    })
  }

  private escapeHtml(str: string): string {
    const div = document.createElement('div')
    div.textContent = str
    return div.innerHTML
  }
}

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', () => {
  if ($('#posts-table-body')) {
    new PostListManager()
  }
})

export { PostListManager }