/**
 * 媒体上传模块
 * 支持拖拽上传、进度条显示、多文件上传
 */

interface UploadProgress {
  id: string;
  file: File;
  progress: number;
  status: "pending" | "uploading" | "success" | "error";
  error?: string;
}

interface UploadResponse {
  id: string;
  filename: string;
  url: string;
  mime_type: string;
  size: number;
}

export class MediaUploader {
  private dropzone: HTMLElement;
  private fileInput: HTMLInputElement;
  private progressContainer: HTMLElement;
  private uploadList: HTMLElement;
  private uploadBtn: HTMLButtonElement;
  private selectFileBtn: HTMLButtonElement;
  private uploads: Map<string, UploadProgress> = new Map();
  private onUploadComplete?: (response: UploadResponse) => void;

  constructor(options: {
    dropzone: HTMLElement;
    fileInput: HTMLInputElement;
    progressContainer: HTMLElement;
    uploadList: HTMLElement;
    uploadBtn: HTMLButtonElement;
    selectFileBtn: HTMLButtonElement;
    onUploadComplete?: (response: UploadResponse) => void;
  }) {
    this.dropzone = options.dropzone;
    this.fileInput = options.fileInput;
    this.progressContainer = options.progressContainer;
    this.uploadList = options.uploadList;
    this.uploadBtn = options.uploadBtn;
    this.selectFileBtn = options.selectFileBtn;
    this.onUploadComplete = options.onUploadComplete;

    this.init();
  }

  private init(): void {
    // 上传按钮点击 - 显示/隐藏上传区域
    this.uploadBtn.addEventListener("click", () => {
      this.dropzone.classList.toggle("hidden");
    });

    // 选择文件按钮
    this.selectFileBtn.addEventListener("click", () => {
      this.fileInput.click();
    });

    // 文件选择
    this.fileInput.addEventListener("change", (e) => {
      const files = (e.target as HTMLInputElement).files;
      if (files && files.length > 0) {
        this.handleFiles(Array.from(files));
      }
    });

    // 拖拽事件
    this.dropzone.addEventListener("dragover", (e) => {
      e.preventDefault();
      this.dropzone.classList.add("border-blue-500", "bg-blue-50");
    });

    this.dropzone.addEventListener("dragleave", (e) => {
      e.preventDefault();
      this.dropzone.classList.remove("border-blue-500", "bg-blue-50");
    });

    this.dropzone.addEventListener("drop", (e) => {
      e.preventDefault();
      this.dropzone.classList.remove("border-blue-500", "bg-blue-50");

      const files = e.dataTransfer?.files;
      if (files && files.length > 0) {
        this.handleFiles(Array.from(files));
      }
    });

    // 点击上传区域触发文件选择
    this.dropzone.addEventListener("click", (e) => {
      if (
        e.target !== this.selectFileBtn &&
        !(e.target as HTMLElement).closest("#select-file-btn")
      ) {
        this.fileInput.click();
      }
    });
  }

  /**
   * 处理文件上传
   */
  private async handleFiles(files: File[]): Promise<void> {
    // 验证文件
    const validFiles = files.filter((file) => this.validateFile(file));

    if (validFiles.length === 0) {
      return;
    }

    // 显示进度区域
    this.progressContainer.classList.remove("hidden");

    // 创建上传任务
    for (const file of validFiles) {
      const uploadId = this.generateId();
      const upload: UploadProgress = {
        id: uploadId,
        file,
        progress: 0,
        status: "pending",
      };
      this.uploads.set(uploadId, upload);
      this.renderUploadItem(upload);
    }

    // 并行上传
    const uploadPromises = validFiles.map((file) => this.uploadFile(file));
    await Promise.allSettled(uploadPromises);
  }

  /**
   * 验证文件
   */
  private validateFile(file: File): boolean {
    // 检查文件大小（10MB 限制）
    const maxSize = 10 * 1024 * 1024;
    if (file.size > maxSize) {
      alert(`文件 ${file.name} 超过 10MB 限制`);
      return false;
    }

    // 检查文件类型
    const allowedTypes = [
      "image/jpeg",
      "image/png",
      "image/gif",
      "image/webp",
      "image/svg+xml",
      "application/pdf",
      "application/msword",
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      "application/zip",
      "text/plain",
    ];

    if (!allowedTypes.includes(file.type)) {
      alert(`文件类型 ${file.type} 不支持`);
      return false;
    }

    return true;
  }

  /**
   * 上传单个文件
   */
  private async uploadFile(file: File): Promise<UploadResponse | null> {
    const uploadId = this.findUploadId(file);
    if (!uploadId) return null;

    try {
      this.updateUploadStatus(uploadId, "uploading", 0);

      const formData = new FormData();
      formData.append("file", file);

      // 使用 XMLHttpRequest 以支持进度
      const response = await this.uploadWithProgress(uploadId, formData);

      this.updateUploadStatus(uploadId, "success", 100);

      // 触发完成回调
      if (this.onUploadComplete) {
        this.onUploadComplete(response);
      }

      return response;
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : "上传失败";
      this.updateUploadStatus(uploadId, "error", 0, errorMsg);
      return null;
    }
  }

  /**
   * 带进度的上传
   */
  private uploadWithProgress(uploadId: string, formData: FormData): Promise<UploadResponse> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();

      xhr.upload.addEventListener("progress", (e) => {
        if (e.lengthComputable) {
          const progress = Math.round((e.loaded / e.total) * 100);
          this.updateUploadStatus(uploadId, "uploading", progress);
        }
      });

      xhr.addEventListener("load", () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const response = JSON.parse(xhr.responseText);
            resolve(response);
          } catch {
            reject(new Error("解析响应失败"));
          }
        } else {
          try {
            const error = JSON.parse(xhr.responseText);
            reject(new Error(error.message || "上传失败"));
          } catch {
            reject(new Error("上传失败"));
          }
        }
      });

      xhr.addEventListener("error", () => {
        reject(new Error("网络错误"));
      });

      xhr.addEventListener("abort", () => {
        reject(new Error("上传已取消"));
      });

      xhr.open("POST", "/api/v1/media/upload");

      // 添加认证 token
      const token = this.getAuthToken();
      if (token) {
        xhr.setRequestHeader("Authorization", `Bearer ${token}`);
      }

      xhr.send(formData);
    });
  }

  /**
   * 渲染上传项
   */
  private renderUploadItem(upload: UploadProgress): void {
    const item = document.createElement("div");
    item.id = `upload-item-${upload.id}`;
    item.className = "flex items-center gap-3 p-2 bg-gray-50 rounded";
    item.innerHTML = `
      <div class="flex-1 min-w-0">
        <div class="flex items-center justify-between mb-1">
          <span class="text-sm text-gray-700 truncate">${upload.file.name}</span>
          <span class="upload-status text-xs text-gray-500">等待中</span>
        </div>
        <div class="w-full bg-gray-200 rounded-full h-2">
          <div class="upload-progress bg-blue-600 h-2 rounded-full transition-all duration-200" style="width: 0%"></div>
        </div>
      </div>
    `;

    this.uploadList.appendChild(item);
  }

  /**
   * 更新上传状态
   */
  private updateUploadStatus(
    uploadId: string,
    status: UploadProgress["status"],
    progress: number,
    error?: string
  ): void {
    const upload = this.uploads.get(uploadId);
    if (!upload) return;

    upload.status = status;
    upload.progress = progress;
    upload.error = error;

    const item = document.getElementById(`upload-item-${uploadId}`);
    if (!item) return;

    const progressBar = item.querySelector(".upload-progress") as HTMLElement;
    const statusEl = item.querySelector(".upload-status") as HTMLElement;

    if (progressBar) {
      progressBar.style.width = `${progress}%`;
    }

    if (statusEl) {
      switch (status) {
        case "uploading":
          statusEl.textContent = `${progress}%`;
          statusEl.className = "upload-status text-xs text-blue-600";
          break;
        case "success":
          statusEl.textContent = "完成";
          statusEl.className = "upload-status text-xs text-green-600";
          if (progressBar) {
            progressBar.classList.remove("bg-blue-600");
            progressBar.classList.add("bg-green-600");
          }
          break;
        case "error":
          statusEl.textContent = error || "失败";
          statusEl.className = "upload-status text-xs text-red-600";
          if (progressBar) {
            progressBar.classList.remove("bg-blue-600");
            progressBar.classList.add("bg-red-600");
          }
          break;
      }
    }
  }

  /**
   * 查找文件对应的上传 ID
   */
  private findUploadId(file: File): string | null {
    for (const [id, upload] of this.uploads) {
      if (upload.file === file) {
        return id;
      }
    }
    return null;
  }

  /**
   * 生成唯一 ID
   */
  private generateId(): string {
    return Math.random().toString(36).substring(2, 15);
  }

  /**
   * 获取认证 token
   */
  private getAuthToken(): string | null {
    // 从 localStorage 获取 token
    return localStorage.getItem("auth_token");
  }

  /**
   * 清除已完成的上传
   */
  public clearCompleted(): void {
    const completedIds: string[] = [];
    for (const [id, upload] of this.uploads) {
      if (upload.status === "success" || upload.status === "error") {
        completedIds.push(id);
      }
    }

    for (const id of completedIds) {
      const item = document.getElementById(`upload-item-${id}`);
      if (item) {
        item.remove();
      }
      this.uploads.delete(id);
    }

    if (this.uploads.size === 0) {
      this.progressContainer.classList.add("hidden");
    }
  }
}

// 导出工厂函数
export function createUploader(options: {
  dropzone: HTMLElement;
  fileInput: HTMLInputElement;
  progressContainer: HTMLElement;
  uploadList: HTMLElement;
  uploadBtn: HTMLButtonElement;
  selectFileBtn: HTMLButtonElement;
  onUploadComplete?: (response: UploadResponse) => void;
}): MediaUploader {
  return new MediaUploader(options);
}
