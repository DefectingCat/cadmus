// 评论管理交互逻辑

import * as commentsApi from "../../api/comments";

// 状态管理
const state = {
  selectedIds: new Set<string>(),
  currentStatus: "pending",
};

// DOM 元素引用
let selectAllCheckbox: HTMLInputElement | null = null;
let batchApproveBtn: HTMLButtonElement | null = null;
let batchRejectBtn: HTMLButtonElement | null = null;
let batchDeleteBtn: HTMLButtonElement | null = null;
let selectedCountSpan: HTMLSpanElement | null = null;
let confirmDialog: HTMLElement | null = null;
let dialogTitle: HTMLElement | null = null;
let dialogMessage: HTMLElement | null = null;
let dialogConfirmBtn: HTMLButtonElement | null = null;
let dialogCancelBtn: HTMLButtonElement | null = null;

// 初始化
document.addEventListener("DOMContentLoaded", () => {
  initElements();
  bindEvents();
  updateBatchToolbar();
});

// 初始化 DOM 元素引用
function initElements(): void {
  selectAllCheckbox = document.getElementById("select-all-checkbox") as HTMLInputElement;
  batchApproveBtn = document.getElementById("batch-approve-btn") as HTMLButtonElement;
  batchRejectBtn = document.getElementById("batch-reject-btn") as HTMLButtonElement;
  batchDeleteBtn = document.getElementById("batch-delete-btn") as HTMLButtonElement;
  selectedCountSpan = document.getElementById("selected-count") as HTMLSpanElement;
  confirmDialog = document.getElementById("confirm-dialog") as HTMLElement;
  dialogTitle = document.getElementById("dialog-title") as HTMLElement;
  dialogMessage = document.getElementById("dialog-message") as HTMLElement;
  dialogConfirmBtn = document.getElementById("dialog-confirm") as HTMLButtonElement;
  dialogCancelBtn = document.getElementById("dialog-cancel") as HTMLButtonElement;

  // 获取当前状态
  const urlParams = new URLSearchParams(window.location.search);
  state.currentStatus = urlParams.get("status") || "pending";
}

// 绑定事件
function bindEvents(): void {
  // 全选按钮
  const selectAllBtn = document.getElementById("select-all-btn");
  if (selectAllBtn) {
    selectAllBtn.addEventListener("click", toggleSelectAll);
  }

  // 全选复选框
  if (selectAllCheckbox) {
    selectAllCheckbox.addEventListener("change", (e) => {
      const checked = (e.target as HTMLInputElement).checked;
      toggleAllCheckboxes(checked);
    });
  }

  // 单个评论复选框
  document.querySelectorAll(".comment-checkbox").forEach((checkbox) => {
    checkbox.addEventListener("change", (e) => {
      const target = e.target as HTMLInputElement;
      const id = target.value;
      if (target.checked) {
        state.selectedIds.add(id);
      } else {
        state.selectedIds.delete(id);
      }
      updateBatchToolbar();
    });
  });

  // 批量操作按钮
  if (batchApproveBtn) {
    batchApproveBtn.addEventListener("click", () => showConfirmDialog("batch-approve"));
  }
  if (batchRejectBtn) {
    batchRejectBtn.addEventListener("click", () => showConfirmDialog("batch-reject"));
  }
  if (batchDeleteBtn) {
    batchDeleteBtn.addEventListener("click", () => showConfirmDialog("batch-delete"));
  }

  // 单个操作按钮
  document.querySelectorAll(".approve-btn").forEach((btn) => {
    btn.addEventListener("click", handleSingleApprove);
  });
  document.querySelectorAll(".reject-btn").forEach((btn) => {
    btn.addEventListener("click", handleSingleReject);
  });
  document.querySelectorAll(".delete-btn").forEach((btn) => {
    btn.addEventListener("click", handleSingleDelete);
  });

  // 对话框按钮
  if (dialogCancelBtn) {
    dialogCancelBtn.addEventListener("click", hideConfirmDialog);
  }
}

// 切换全选
function toggleSelectAll(): void {
  if (selectAllCheckbox) {
    const newChecked = !selectAllCheckbox.checked;
    selectAllCheckbox.checked = newChecked;
    toggleAllCheckboxes(newChecked);
  }
}

// 切换所有复选框
function toggleAllCheckboxes(checked: boolean): void {
  document.querySelectorAll(".comment-checkbox").forEach((checkbox) => {
    (checkbox as HTMLInputElement).checked = checked;
    const id = (checkbox as HTMLInputElement).value;
    if (checked) {
      state.selectedIds.add(id);
    } else {
      state.selectedIds.delete(id);
    }
  });
  updateBatchToolbar();
}

// 更新批量操作工具栏
function updateBatchToolbar(): void {
  const count = state.selectedIds.size;
  const hasSelection = count > 0;

  if (batchApproveBtn) {
    batchApproveBtn.classList.toggle("hidden", !hasSelection || state.currentStatus !== "pending");
  }
  if (batchRejectBtn) {
    batchRejectBtn.classList.toggle("hidden", !hasSelection || state.currentStatus !== "pending");
  }
  if (batchDeleteBtn) {
    batchDeleteBtn.classList.toggle("hidden", !hasSelection);
  }
  if (selectedCountSpan) {
    selectedCountSpan.classList.toggle("hidden", !hasSelection);
    selectedCountSpan.textContent = `已选择 ${count} 条`;
  }
}

// 显示确认对话框
let pendingAction: string | null = null;

function showConfirmDialog(action: string): void {
  const count = state.selectedIds.size;
  pendingAction = action;

  let title = "";
  let message = "";
  let confirmText = "";
  let confirmClass = "";

  switch (action) {
    case "batch-approve":
      title = "批量批准";
      message = `确定要批准选中的 ${count} 条评论吗？`;
      confirmText = "批准";
      confirmClass = "bg-green-500 hover:bg-green-600";
      break;
    case "batch-reject":
      title = "批量拒绝";
      message = `确定要拒绝选中的 ${count} 条评论吗？`;
      confirmText = "拒绝";
      confirmClass = "bg-red-500 hover:bg-red-600";
      break;
    case "batch-delete":
      title = "批量删除";
      message = `确定要删除选中的 ${count} 条评论吗？此操作不可撤销。`;
      confirmText = "删除";
      confirmClass = "bg-red-500 hover:bg-red-600";
      break;
    case "single-delete":
      title = "删除评论";
      message = "确定要删除这条评论吗？此操作不可撤销。";
      confirmText = "删除";
      confirmClass = "bg-red-500 hover:bg-red-600";
      break;
  }

  if (dialogTitle) dialogTitle.textContent = title;
  if (dialogMessage) dialogMessage.textContent = message;
  if (dialogConfirmBtn) {
    dialogConfirmBtn.textContent = confirmText;
    dialogConfirmBtn.className = `px-4 py-2 text-white rounded ${confirmClass}`;
    dialogConfirmBtn.removeEventListener("click", handleConfirmAction);
    dialogConfirmBtn.addEventListener("click", handleConfirmAction);
  }
  if (confirmDialog) {
    confirmDialog.classList.remove("hidden");
  }
}

// 隐藏确认对话框
function hideConfirmDialog(): void {
  if (confirmDialog) {
    confirmDialog.classList.add("hidden");
  }
  pendingAction = null;
}

// 处理确认操作
async function handleConfirmAction(): Promise<void> {
  if (!pendingAction) return;

  try {
    const ids = Array.from(state.selectedIds);

    switch (pendingAction) {
      case "batch-approve":
        await commentsApi.batchApproveComments(ids);
        showSuccessToast("批量批准成功");
        break;
      case "batch-reject":
        await commentsApi.batchRejectComments(ids);
        showSuccessToast("批量拒绝成功");
        break;
      case "batch-delete":
        await commentsApi.batchDeleteComments(ids);
        showSuccessToast("批量删除成功");
        break;
      case "single-delete":
        if (ids.length === 1) {
          await commentsApi.deleteComment(ids[0]);
          showSuccessToast("删除成功");
        }
        break;
    }

    // 刷新页面
    window.location.reload();
  } catch (error) {
    showErrorToast(error instanceof Error ? error.message : "操作失败");
    hideConfirmDialog();
  }
}

// 单个批准
async function handleSingleApprove(e: Event): Promise<void> {
  const btn = e.currentTarget as HTMLButtonElement;
  const id = btn.dataset.id;
  if (!id) return;

  try {
    btn.disabled = true;
    btn.textContent = "处理中...";
    await commentsApi.approveComment(id);
    showSuccessToast("批准成功");
    window.location.reload();
  } catch (error) {
    showErrorToast(error instanceof Error ? error.message : "批准失败");
    btn.disabled = false;
    btn.textContent = "通过";
  }
}

// 单个拒绝
async function handleSingleReject(e: Event): Promise<void> {
  const btn = e.currentTarget as HTMLButtonElement;
  const id = btn.dataset.id;
  if (!id) return;

  try {
    btn.disabled = true;
    btn.textContent = "处理中...";
    await commentsApi.rejectComment(id);
    showSuccessToast("拒绝成功");
    window.location.reload();
  } catch (error) {
    showErrorToast(error instanceof Error ? error.message : "拒绝失败");
    btn.disabled = false;
    btn.textContent = "拒绝";
  }
}

// 单个删除
async function handleSingleDelete(e: Event): Promise<void> {
  const btn = e.currentTarget as HTMLButtonElement;
  const id = btn.dataset.id;
  if (!id) return;

  state.selectedIds.clear();
  state.selectedIds.add(id);
  showConfirmDialog("single-delete");
}

// Toast 提示
function showSuccessToast(message: string): void {
  showToast(message, "success");
}

function showErrorToast(message: string): void {
  showToast(message, "error");
}

function showToast(message: string, type: "success" | "error"): void {
  const toast = document.createElement("div");
  toast.className = `fixed bottom-4 right-4 px-4 py-2 rounded-lg text-white ${
    type === "success" ? "bg-green-500" : "bg-red-500"
  }`;
  toast.textContent = message;
  document.body.appendChild(toast);

  setTimeout(() => {
    toast.remove();
  }, 3000);
}
