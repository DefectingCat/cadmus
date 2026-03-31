<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# api

## Purpose
API 调用封装，统一管理 HTTP 请求。

## Key Files
| File | Description |
|------|-------------|
| `client.ts` | 基础请求客户端：headers、错误处理 |
| `posts.ts` | 文章 API：CRUD、发布、版本管理 |
| `comments.ts` | 评论 API：CRUD、审核 |

## For AI Agents

### Working In This Directory
- 所有 API 调用通过 `client.ts` 统一处理
- JWT Token 自动从 localStorage 读取
- 错误统一处理

### Client Pattern
```typescript
// api/client.ts
const BASE_URL = '/api/v1';

interface RequestOptions {
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
    body?: any;
    headers?: Record<string, string>;
}

export async function apiClient<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const token = localStorage.getItem('token');
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...options.headers,
    };
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${BASE_URL}${path}`, {
        method: options.method || 'GET',
        headers,
        body: options.body ? JSON.stringify(options.body) : undefined,
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Request failed');
    }

    return response.json();
}
```

### API Modules
```typescript
// api/posts.ts
export const postsApi = {
    list: (params?: PostListParams) =>
        apiClient<PostListResponse>(`/posts?${new URLSearchParams(params)}`),

    get: (slug: string) =>
        apiClient<Post>(`/posts/${slug}`),

    create: (data: CreatePostInput) =>
        apiClient<Post>('/posts', { method: 'POST', body: data }),

    update: (id: string, data: UpdatePostInput) =>
        apiClient<Post>(`/posts/${id}`, { method: 'PUT', body: data }),

    delete: (id: string) =>
        apiClient<void>(`/posts/${id}`, { method: 'DELETE' }),

    publish: (id: string) =>
        apiClient<Post>(`/posts/${id}/publish`, { method: 'POST' }),
};
```

### Error Types
| HTTP Status | Error Type |
|-------------|------------|
| 401 | 认证失败，跳转登录 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 422 | 参数验证失败 |
| 500 | 服务器错误 |

<!-- MANUAL: -->