// 评论管理 API

const API_BASE = '/api/v1';

interface Comment {
  id: string;
  post_id: string;
  user_id: string;
  parent_id?: string;
  depth: number;
  content: string;
  status: string;
  like_count: number;
  created_at: string;
  updated_at: string;
}

interface AdminCommentListResponse {
  comments: Comment[];
  total: number;
  page: number;
  per_page: number;
}

interface ApiError {
  code: string;
  message: string;
  details?: Record<string, string>;
}

// 获取存储的 token
function getAuthToken(): string | null {
  return localStorage.getItem('auth_token');
}

// 创建带认证的请求头
function createHeaders(): HeadersInit {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  const token = getAuthToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

// 处理 API 错误
async function handleApiError(response: Response): Promise<ApiError> {
  try {
    const data = await response.json();
    return data as ApiError;
  } catch {
    return {
      code: 'UNKNOWN_ERROR',
      message: `请求失败: ${response.status} ${response.statusText}`,
    };
  }
}

// 获取评论列表
export async function getAdminComments(
  status: string,
  page: number = 1,
  perPage: number = 20
): Promise<AdminCommentListResponse> {
  const url = `${API_BASE}/admin/comments?status=${status}&page=${page}&per_page=${perPage}`;
  const response = await fetch(url, {
    method: 'GET',
    headers: createHeaders(),
  });

  if (!response.ok) {
    const error = await handleApiError(response);
    throw new Error(error.message);
  }

  return response.json();
}

// 批准评论
export async function approveComment(id: string): Promise<void> {
  const url = `${API_BASE}/comments/${id}/approve`;
  const response = await fetch(url, {
    method: 'PUT',
    headers: createHeaders(),
  });

  if (!response.ok) {
    const error = await handleApiError(response);
    throw new Error(error.message);
  }
}

// 拒绝评论
export async function rejectComment(id: string): Promise<void> {
  const url = `${API_BASE}/comments/${id}/reject`;
  const response = await fetch(url, {
    method: 'PUT',
    headers: createHeaders(),
  });

  if (!response.ok) {
    const error = await handleApiError(response);
    throw new Error(error.message);
  }
}

// 管理员删除评论
export async function deleteComment(id: string): Promise<void> {
  const url = `${API_BASE}/admin/comments/${id}`;
  const response = await fetch(url, {
    method: 'DELETE',
    headers: createHeaders(),
  });

  if (!response.ok && response.status !== 204) {
    const error = await handleApiError(response);
    throw new Error(error.message);
  }
}

// 批量批准评论
export async function batchApproveComments(ids: string[]): Promise<void> {
  const url = `${API_BASE}/admin/comments/batch-approve`;
  const response = await fetch(url, {
    method: 'PUT',
    headers: createHeaders(),
    body: JSON.stringify({ ids }),
  });

  if (!response.ok) {
    const error = await handleApiError(response);
    throw new Error(error.message);
  }
}

// 批量拒绝评论
export async function batchRejectComments(ids: string[]): Promise<void> {
  const url = `${API_BASE}/admin/comments/batch-reject`;
  const response = await fetch(url, {
    method: 'PUT',
    headers: createHeaders(),
    body: JSON.stringify({ ids }),
  });

  if (!response.ok) {
    const error = await handleApiError(response);
    throw new Error(error.message);
  }
}

// 批量删除评论
export async function batchDeleteComments(ids: string[]): Promise<void> {
  const url = `${API_BASE}/admin/comments/batch-delete`;
  const response = await fetch(url, {
    method: 'DELETE',
    headers: createHeaders(),
    body: JSON.stringify({ ids }),
  });

  if (!response.ok) {
    const error = await handleApiError(response);
    throw new Error(error.message);
  }
}