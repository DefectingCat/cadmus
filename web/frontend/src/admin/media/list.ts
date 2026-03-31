/**
 * 媒体列表模块
 * 支持网格视图、分页、选择、搜索筛选
 */

interface MediaItem {
  id: string;
  uploader_id: string;
  filename: string;
  original_name: string;
  url: string;
  mime_type: string;
  size: number;
  width?: number;
  height?: number;
  alt_text?: string;
  created_at: string;
}

interface MediaListResponse {
  media: MediaItem[];
  total: number;
}

export class MediaList {
  private grid: HTMLElement;
  private searchInput: HTMLInputElement;
  private typeFilter: HTMLSelectElement;
  private selectedItems: Set<string> = new Set();
  private currentPage: number = 1;
  private perPage: number = 20;
  private total: number = 0;
  private onMediaSelect?: (media: MediaItem) => void;
  private onMediaInsert?: (media: MediaItem) => void;

  constructor(options: {
    grid: HTMLElement;
    searchInput: HTMLInputElement;
    typeFilter: HTMLSelectElement;
    onMediaSelect?: (media: MediaItem) => void;
    onMediaInsert?: (media: MediaItem) => void;
  }) {
    this.grid = options.grid;
    this.searchInput = options.searchInput;
    this.typeFilter = options.typeFilter;
    this.onMediaSelect = options.onMediaSelect;
    this.onMediaInsert = options.onMediaInsert;

    this.init();
  }

  private init(): void {
    // 搜索防抖
    let searchTimeout: number | null = null;
    this.searchInput.addEventListener('input', () => {
      if (searchTimeout) clearTimeout(searchTimeout);
      searchTimeout = window.setTimeout(() => {
        this.currentPage = 1;
        this.loadMedia();
      }, 300);
    });

    // 类型筛选
    this.typeFilter.addEventListener('change', () => {
      this.currentPage = 1;
      this.loadMedia();
    });

    // 分页按钮
    document.getElementById('prev-page')?.addEventListener('click', () => {
      if (this.currentPage > 1) {
        this.currentPage--;
        this.loadMedia();
      }
    });

    document.getElementById('next-page')?.addEventListener('click', () => {
      if (this.currentPage * this.perPage < this.total) {
        this.currentPage++;
        this.loadMedia();
      }
    });

    // 清除选择
    document.getElementById('clear-selection-btn')?.addEventListener('click', () => {
      this.clearSelection();
    });

    // 全局点击处理（媒体项点击和复选框）
    this.grid.addEventListener('click', (e) => {
      const target = e.target as HTMLElement;
      const mediaItem = target.closest('.media-item') as HTMLElement;

      if (!mediaItem) return;

      const mediaId = mediaItem.dataset.id;

      // 如果点击的是复选框
      if (target.classList.contains('media-checkbox')) {
        const checkbox = target as HTMLInputElement;
        if (checkbox.checked) {
          this.selectedItems.add(mediaId);
        } else {
          this.selectedItems.delete(mediaId);
        }
        this.updateSelectionUI();
        return;
      }

      // 否则打开详情弹窗
      if (mediaId) {
        this.openMediaDetail(mediaId);
      }
    });

    // 删除选中按钮
    document.getElementById('delete-selected-btn')?.addEventListener('click', () => {
      this.deleteSelected();
    });

    // 初始化页面数据
    if ((window as any).mediaData) {
      this.currentPage = (window as any).mediaData.page || 1;
      this.perPage = (window as any).mediaData.perPage || 20;
      this.total = (window as any).mediaData.total || 0;
    }
  }

  /**
   * 加载媒体列表
   */
  public async loadMedia(): Promise<void> {
    const params = new URLSearchParams();
    const offset = (this.currentPage - 1) * this.perPage;

    params.set('offset', String(offset));
    params.set('limit', String(this.perPage));

    const search = this.searchInput.value.trim();
    if (search) {
      params.set('search', search);
    }

    const type = this.typeFilter.value;
    if (type) {
      params.set('type', type);
    }

    try {
      const response = await fetch(`/api/v1/media?${params}`, {
        headers: this.getAuthHeaders()
      });

      if (!response.ok) {
        throw new Error('加载失败');
      }

      const data: MediaListResponse = await response.json();
      this.total = data.total;
      this.renderMediaList(data.media);
      this.updatePagination();
    } catch (error) {
      console.error('加载媒体列表失败:', error);
    }
  }

  /**
   * 渲染媒体列表
   */
  private renderMediaList(medias: MediaItem[]): void {
    if (medias.length === 0) {
      this.grid.innerHTML = `
        <div class="col-span-full text-center py-12 text-gray-500">
          <svg class="w-16 h-16 mx-auto text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"/>
          </svg>
          <p>暂无媒体文件</p>
          <p class="text-sm mt-1">点击上方"上传文件"按钮添加</p>
        </div>
      `;
      return;
    }

    this.grid.innerHTML = medias.map(m => this.renderMediaItem(m)).join('');
  }

  /**
   * 渲染单个媒体项
   */
  private renderMediaItem(m: MediaItem): string {
    const isImage = m.mime_type.startsWith('image/');
    const sizeStr = this.formatFileSize(m.size);
    const isSelected = this.selectedItems.has(m.id);

    if (isImage) {
      return `
        <div
          class="media-item relative group bg-gray-50 rounded-lg overflow-hidden cursor-pointer hover:ring-2 hover:ring-blue-500 transition-all ${isSelected ? 'ring-2 ring-blue-500' : ''}"
          data-id="${m.id}"
          data-mime="${m.mime_type}"
          data-url="${m.url}"
          data-name="${m.original_name}"
        >
          <div class="absolute top-2 left-2 z-10">
            <input
              type="checkbox"
              class="media-checkbox w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              data-id="${m.id}"
              ${isSelected ? 'checked' : ''}
            />
          </div>
          <div class="aspect-square flex items-center justify-center">
            <img src="${m.url}" alt="${m.original_name}" class="object-cover w-full h-full"/>
          </div>
          <div class="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent p-2 group-hover:from-black/80">
            <p class="text-white text-xs truncate">${m.original_name}</p>
            <p class="text-gray-300 text-xs">${sizeStr}</p>
          </div>
        </div>
      `;
    } else {
      return `
        <div
          class="media-item relative group bg-gray-50 rounded-lg overflow-hidden cursor-pointer hover:ring-2 hover:ring-blue-500 transition-all ${isSelected ? 'ring-2 ring-blue-500' : ''}"
          data-id="${m.id}"
          data-mime="${m.mime_type}"
          data-url="${m.url}"
          data-name="${m.original_name}"
        >
          <div class="absolute top-2 left-2 z-10">
            <input
              type="checkbox"
              class="media-checkbox w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              data-id="${m.id}"
              ${isSelected ? 'checked' : ''}
            />
          </div>
          <div class="aspect-square flex items-center justify-center">
            <div class="flex flex-col items-center justify-center text-gray-400">
              <svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
              </svg>
              <span class="text-xs mt-2 truncate max-w-full px-2">${m.original_name}</span>
            </div>
          </div>
          <div class="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent p-2 group-hover:from-black/80">
            <p class="text-white text-xs truncate">${m.original_name}</p>
            <p class="text-gray-300 text-xs">${sizeStr}</p>
          </div>
        </div>
      `;
    }
  }

  /**
   * 更新分页状态
   */
  private updatePagination(): void {
    const prevBtn = document.getElementById('prev-page') as HTMLButtonElement;
    const nextBtn = document.getElementById('next-page') as HTMLButtonElement;

    if (prevBtn) {
      prevBtn.disabled = this.currentPage <= 1;
    }

    if (nextBtn) {
      nextBtn.disabled = this.currentPage * this.perPage >= this.total;
    }
  }

  /**
   * 更新选择 UI
   */
  private updateSelectionUI(): void {
    const countEl = document.getElementById('selected-count');
    const clearBtn = document.getElementById('clear-selection-btn');
    const deleteBtn = document.getElementById('delete-selected-btn');

    const count = this.selectedItems.size;

    if (countEl) {
      countEl.textContent = `已选择 ${count} 个文件`;
    }

    if (clearBtn) {
      clearBtn.classList.toggle('hidden', count === 0);
    }

    if (deleteBtn) {
      (deleteBtn as HTMLButtonElement).disabled = count === 0;
    }
  }

  /**
   * 清除选择
   */
  private clearSelection(): void {
    this.selectedItems.clear();
    this.grid.querySelectorAll('.media-checkbox').forEach(cb => {
      (cb as HTMLInputElement).checked = false;
    });
    this.grid.querySelectorAll('.media-item').forEach(item => {
      item.classList.remove('ring-2', 'ring-blue-500');
    });
    this.updateSelectionUI();
  }

  /**
   * 打开媒体详情弹窗
   */
  private openMediaDetail(mediaId: string): void {
    const mediaItem = this.grid.querySelector(`[data-id="${mediaId}"]`) as HTMLElement;
    if (!mediaItem) return;

    const modal = document.getElementById('media-modal');
    const modalContent = document.getElementById('modal-content');
    const modalTitle = document.getElementById('modal-title');

    if (!modal || !modalContent) return;

    const url = mediaItem.dataset.url;
    const name = mediaItem.dataset.name;
    const mime = mediaItem.dataset.mime;
    const isImage = mime?.startsWith('image/');

    if (modalTitle) {
      modalTitle.textContent = name || '媒体详情';
    }

    if (isImage && url) {
      modalContent.innerHTML = `
        <div class="text-center">
          <img src="${url}" alt="${name}" class="max-w-full max-h-96 mx-auto rounded"/>
        </div>
        <div class="mt-4 space-y-2">
          <p class="text-sm"><span class="font-medium text-gray-700">文件名：</span>${name}</p>
          <p class="text-sm"><span class="font-medium text-gray-700">类型：</span>${mime}</p>
          <p class="text-sm"><span class="font-medium text-gray-700">URL：</span>
            <code class="bg-gray-100 px-2 py-1 rounded text-xs">${url}</code>
          </p>
        </div>
      `;
    } else {
      modalContent.innerHTML = `
        <div class="space-y-2">
          <p class="text-sm"><span class="font-medium text-gray-700">文件名：</span>${name}</p>
          <p class="text-sm"><span class="font-medium text-gray-700">类型：</span>${mime}</p>
          <p class="text-sm"><span class="font-medium text-gray-700">URL：</span>
            <code class="bg-gray-100 px-2 py-1 rounded text-xs">${url}</code>
          </p>
        </div>
      `;
    }

    modal.classList.remove('hidden');
    modal.dataset.mediaId = mediaId;
    modal.dataset.url = url || '';
    modal.dataset.name = name || '';

    // 关闭按钮
    document.getElementById('close-modal')?.addEventListener('click', () => {
      modal.classList.add('hidden');
    });

    // 插入到文章
    document.getElementById('insert-media-btn')?.addEventListener('click', () => {
      this.insertMedia(mediaId, url || '', name || '', isImage);
      modal.classList.add('hidden');
    });

    // 删除媒体
    document.getElementById('delete-media-btn')?.addEventListener('click', async () => {
      if (confirm('确定要删除这个媒体文件吗？')) {
        await this.deleteMedia(mediaId);
        modal.classList.add('hidden');
      }
    });
  }

  /**
   * 插入媒体到编辑器
   */
  private insertMedia(id: string, url: string, name: string, isImage: boolean): void {
    if (this.onMediaInsert) {
      this.onMediaInsert({
        id,
        url,
        original_name: name,
        mime_type: isImage ? 'image/jpeg' : 'application/octet-stream',
        uploader_id: '',
        filename: '',
        size: 0,
        created_at: ''
      });
    } else {
      // 默认行为：复制 Markdown 格式到剪贴板
      const text = isImage ? `![${name}](${url})` : `[${name}](${url})`;
      navigator.clipboard.writeText(text).then(() => {
        alert('已复制到剪贴板：' + text);
      });
    }
  }

  /**
   * 删除单个媒体
   */
  private async deleteMedia(mediaId: string): Promise<void> {
    try {
      const response = await fetch(`/api/v1/media/${mediaId}`, {
        method: 'DELETE',
        headers: this.getAuthHeaders()
      });

      if (!response.ok) {
        throw new Error('删除失败');
      }

      // 从列表中移除
      const item = this.grid.querySelector(`[data-id="${mediaId}"]`);
      if (item) {
        item.remove();
      }

      // 检查是否需要重新加载
      if (this.grid.querySelectorAll('.media-item').length === 0 && this.currentPage > 1) {
        this.currentPage--;
        this.loadMedia();
      }
    } catch (error) {
      console.error('删除媒体失败:', error);
      alert('删除失败');
    }
  }

  /**
   * 删除选中的媒体
   */
  private async deleteSelected(): Promise<void> {
    if (this.selectedItems.size === 0) return;

    const count = this.selectedItems.size;
    if (!confirm(`确定要删除选中的 ${count} 个文件吗？`)) return;

    const ids = Array.from(this.selectedItems);
    let successCount = 0;

    for (const id of ids) {
      try {
        const response = await fetch(`/api/v1/media/${id}`, {
          method: 'DELETE',
          headers: this.getAuthHeaders()
        });

        if (response.ok) {
          successCount++;
          const item = this.grid.querySelector(`[data-id="${id}"]`);
          if (item) item.remove();
        }
      } catch (error) {
        console.error(`删除 ${id} 失败:`, error);
      }
    }

    this.selectedItems.clear();
    this.updateSelectionUI();

    if (successCount < count) {
      alert(`成功删除 ${successCount}/${count} 个文件`);
    }

    // 检查是否需要重新加载
    if (this.grid.querySelectorAll('.media-item').length === 0) {
      if (this.currentPage > 1) {
        this.currentPage--;
      }
      this.loadMedia();
    }
  }

  /**
   * 格式化文件大小
   */
  private formatFileSize(size: number): string {
    if (size < 1024) {
      return size + ' B';
    } else if (size < 1024 * 1024) {
      return (size / 1024).toFixed(1) + ' KB';
    } else {
      return (size / 1024 / 1024).toFixed(1) + ' MB';
    }
  }

  /**
   * 获取认证请求头
   */
  private getAuthHeaders(): HeadersInit {
    const token = localStorage.getItem('auth_token');
    const headers: HeadersInit = {
      'Content-Type': 'application/json'
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
  }

  /**
   * 添加新媒体到列表
   */
  public addMedia(media: MediaItem): void {
    const currentHtml = this.grid.innerHTML;
    const newItem = this.renderMediaItem(media);

    // 插入到开头
    if (currentHtml.includes('col-span-full')) {
      // 当前是空状态
      this.grid.innerHTML = newItem;
    } else {
      this.grid.insertAdjacentHTML('afterbegin', newItem);
    }
  }
}

// 媒体选择器（用于编辑器中插入媒体）
export class MediaPicker {
  private modal: HTMLElement | null = null;
  private onSelect?: (media: MediaItem) => void;

  constructor(options: { onSelect?: (media: MediaItem) => void } = {}) {
    this.onSelect = options.onSelect;
  }

  public open(): void {
    // 创建模态框
    this.modal = document.createElement('div');
    this.modal.className = 'fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center';
    this.modal.innerHTML = `
      <div class="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[80vh] overflow-hidden flex flex-col">
        <div class="p-4 border-b border-gray-200 flex justify-between items-center">
          <h3 class="text-lg font-medium text-gray-900">选择媒体</h3>
          <button class="close-picker text-gray-400 hover:text-gray-600">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <div class="p-4 border-b border-gray-200 flex gap-3">
          <input type="text" class="picker-search px-4 py-2 border border-gray-300 rounded-lg w-64" placeholder="搜索文件名...">
          <button class="picker-upload px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">上传新文件</button>
        </div>
        <div class="picker-grid p-4 overflow-y-auto flex-1 grid grid-cols-4 gap-4"></div>
      </div>
    `;

    document.body.appendChild(this.modal);

    // 加载媒体
    this.loadMedia();

    // 关闭按钮
    this.modal.querySelector('.close-picker')?.addEventListener('click', () => {
      this.close();
    });

    // 点击背景关闭
    this.modal.addEventListener('click', (e) => {
      if (e.target === this.modal) {
        this.close();
      }
    });
  }

  public close(): void {
    if (this.modal) {
      this.modal.remove();
      this.modal = null;
    }
  }

  private async loadMedia(): Promise<void> {
    if (!this.modal) return;

    const grid = this.modal.querySelector('.picker-grid') as HTMLElement;
    if (!grid) return;

    try {
      const response = await fetch('/api/v1/media?limit=40', {
        headers: this.getAuthHeaders()
      });

      if (!response.ok) throw new Error('加载失败');

      const data: MediaListResponse = await response.json();

      grid.innerHTML = data.media.map(m => `
        <div class="picker-item cursor-pointer rounded-lg overflow-hidden hover:ring-2 hover:ring-blue-500" data-id="${m.id}" data-url="${m.url}" data-name="${m.original_name}" data-mime="${m.mime_type}">
          ${m.mime_type.startsWith('image/')
            ? `<img src="${m.url}" alt="${m.original_name}" class="w-full aspect-square object-cover"/>`
            : `<div class="w-full aspect-square bg-gray-100 flex items-center justify-center text-gray-400">
                <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                </svg>
              </div>`
          }
          <p class="text-xs text-gray-600 truncate p-1">${m.original_name}</p>
        </div>
      `).join('');

      // 点击选择
      grid.querySelectorAll('.picker-item').forEach(item => {
        item.addEventListener('click', () => {
          const media: MediaItem = {
            id: item.dataset.id || '',
            url: item.dataset.url || '',
            original_name: item.dataset.name || '',
            mime_type: item.dataset.mime || '',
            uploader_id: '',
            filename: '',
            size: 0,
            created_at: ''
          };

          if (this.onSelect) {
            this.onSelect(media);
          }
          this.close();
        });
      });
    } catch (error) {
      grid.innerHTML = '<p class="text-center text-gray-500 col-span-4">加载失败</p>';
    }
  }

  private getAuthHeaders(): HeadersInit {
    const token = localStorage.getItem('auth_token');
    const headers: HeadersInit = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
  }
}

// 导出工厂函数
export function createMediaList(options: {
  grid: HTMLElement;
  searchInput: HTMLInputElement;
  typeFilter: HTMLSelectElement;
  onMediaSelect?: (media: MediaItem) => void;
  onMediaInsert?: (media: MediaItem) => void;
}): MediaList {
  return new MediaList(options);
}

export function createMediaPicker(options?: { onSelect?: (media: MediaItem) => void }): MediaPicker {
  return new MediaPicker(options);
}