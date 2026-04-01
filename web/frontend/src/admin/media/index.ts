/**
 * 媒体管理页面入口
 */

import { createUploader, UploadResponse } from "./upload";
import { createMediaList, MediaItem, createMediaPicker } from "./list";

// 初始化媒体管理页面
function initMediaPage(): void {
  // 获取 DOM 元素
  const dropzone = document.getElementById("upload-dropzone") as HTMLElement;
  const fileInput = document.getElementById("file-input") as HTMLInputElement;
  const progressContainer = document.getElementById("upload-progress") as HTMLElement;
  const uploadList = document.getElementById("upload-list") as HTMLElement;
  const uploadBtn = document.getElementById("upload-btn") as HTMLButtonElement;
  const selectFileBtn = document.getElementById("select-file-btn") as HTMLButtonElement;
  const grid = document.getElementById("media-grid") as HTMLElement;
  const searchInput = document.getElementById("media-search") as HTMLInputElement;
  const typeFilter = document.getElementById("media-type-filter") as HTMLSelectElement;

  if (
    !dropzone ||
    !fileInput ||
    !progressContainer ||
    !uploadList ||
    !uploadBtn ||
    !selectFileBtn
  ) {
    console.error("Required DOM elements not found");
    return;
  }

  // 初始化媒体列表
  const mediaList = createMediaList({
    grid,
    searchInput,
    typeFilter,
    onMediaInsert: (media: MediaItem) => {
      // 复制 Markdown 格式到剪贴板
      const isImage = media.mime_type.startsWith("image/");
      const text = isImage
        ? `![${media.original_name}](${media.url})`
        : `[${media.original_name}](${media.url})`;
      navigator.clipboard
        .writeText(text)
        .then(() => {
          showToast("已复制到剪贴板");
        })
        .catch(() => {
          showToast("复制失败", "error");
        });
    },
  });

  // 初始化上传器
  createUploader({
    dropzone,
    fileInput,
    progressContainer,
    uploadList,
    uploadBtn,
    selectFileBtn,
    onUploadComplete: (response: UploadResponse) => {
      // 上传成功后添加到列表
      const media: MediaItem = {
        id: response.id,
        uploader_id: "",
        filename: response.filename,
        original_name: response.filename,
        url: response.url,
        mime_type: response.mime_type,
        size: response.size,
        created_at: new Date().toISOString(),
      };
      mediaList.addMedia(media);
      showToast(`${response.filename} 上传成功`);
    },
  });
}

// 显示提示消息
function showToast(message: string, type: "success" | "error" = "success"): void {
  const toast = document.createElement("div");
  toast.className = `fixed bottom-4 right-4 px-4 py-2 rounded-lg shadow-lg z-50 ${
    type === "success" ? "bg-green-500 text-white" : "bg-red-500 text-white"
  }`;
  toast.textContent = message;
  document.body.appendChild(toast);

  setTimeout(() => {
    toast.remove();
  }, 3000);
}

// 导出媒体选择器供编辑器使用
export { createMediaPicker, MediaItem };

// 页面加载后初始化
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initMediaPage);
} else {
  initMediaPage();
}
