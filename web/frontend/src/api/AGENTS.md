<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# api

## Purpose

`api/` 目录包含前端 API 客户端层，负责与后端 `/api/v1` 接口进行通信。提供统一的请求封装、认证处理和类型安全的响应处理。

## Key Files

| File | Description |
|------|-------------|
| `client.ts` | 基础 HTTP 请求客户端，提供 GET/POST/PUT/DELETE 方法，统一处理认证 token、错误响应和网络异常 |
| `posts.ts` | 文章管理 API 服务，使用 `apiClient` 封装文章 CRUD、发布、批量操作、版本回滚等功能 |
| `comments.ts` | 评论管理 API 服务，独立实现评论列表、审核（批准/拒绝）、删除、批量操作等功能 |

## Architecture

### 基础客户端 (client.ts)

`APIClient` 类提供统一的 API 请求封装：

- **认证**: 从 cookie 和 localStorage 自动加载 `auth_token`，请求时添加 `Authorization: Bearer <token>` 头
- **类型安全**: 返回 `APIResponse<T>` 泛型，包含 `data` 或 `error` 字段
- **错误处理**: 捕获网络异常，返回统一的 `APIError` 结构
- **实例导出**: `apiClient` 作为全局单例供其他模块使用

```typescript
// 使用示例
import { apiClient, APIResponse } from "./client";

const result = await apiClient.get<Post>("/posts/123");
if (result.error) {
  console.error(result.error.message);
} else {
  console.log(result.data);
}
```

### 文章 API (posts.ts)

`postsAPI` 服务对象提供完整的文章管理功能：

| Method | Endpoint | Description |
|--------|----------|-------------|
| `list(filters)` | `GET /posts` | 获取文章列表，支持分页、状态、作者、分类、搜索过滤 |
| `get(id)` | `GET /posts/{id}` | 获取单篇文章详情 |
| `getBySlug(slug)` | `GET /posts/{slug}` | 通过 slug 获取文章 |
| `create(data)` | `POST /posts` | 创建新文章 |
| `update(id, data)` | `PUT /posts/{id}` | 更新文章 |
| `delete(id)` | `DELETE /posts/{id}` | 删除文章 |
| `publish(id)` | `POST /posts/{id}/publish` | 发布文章 |
| `batchDelete(ids)` | 多次 `DELETE` | 批量删除（逐个调用） |
| `batchPublish(ids)` | 多次 `POST` | 批量发布（逐个调用） |
| `versions(id)` | `GET /posts/{id}/versions` | 获取版本历史 |
| `rollback(id, version)` | `POST /posts/{id}/rollback` | 回滚到指定版本 |

### 评论 API (comments.ts)

独立函数导出，用于评论管理：

| Function | Endpoint | Description |
|----------|----------|-------------|
| `getAdminComments(status, page, perPage)` | `GET /admin/comments` | 获取管理员评论列表 |
| `approveComment(id)` | `PUT /comments/{id}/approve` | 批准评论 |
| `rejectComment(id)` | `PUT /comments/{id}/reject` | 拒绝评论 |
| `deleteComment(id)` | `DELETE /admin/comments/{id}` | 删除评论 |
| `batchApproveComments(ids)` | `PUT /admin/comments/batch-approve` | 批量批准 |
| `batchRejectComments(ids)` | `PUT /admin/comments/batch-reject` | 批量拒绝 |
| `batchDeleteComments(ids)` | `DELETE /admin/comments/batch-delete` | 批量删除 |

## For AI Agents

### 添加新 API 服务

1. 在 `api/` 目录下创建新文件（如 `media.ts`）
2. 导入 `apiClient` 和类型：
   ```typescript
   import { apiClient, APIResponse } from "./client";
   ```
3. 定义请求/响应类型接口
4. 创建服务函数或对象，使用 `apiClient` 方法：
   ```typescript
   export const mediaAPI = {
     list: async () => apiClient.get<MediaList>("/media"),
     upload: async (data: UploadRequest) => apiClient.post<Media>("/media", data),
   };
   ```
5. 运行 `bun run typecheck` 验证类型

### 错误处理模式

所有 API 调用都应处理可能的错误：

```typescript
const result = await postsAPI.get(id);
if (result.error) {
  // 处理错误：result.error.code, result.error.message
  return;
}
// 使用 result.data
```

### 认证说明

- Token 自动管理：`APIClient` 在构造时从 cookie 和 localStorage 加载
- 手动设置：`apiClient.setToken(token)` 用于登录后保存
- 清除认证：`apiClient.clearToken()` 用于登出

### 请求路径规范

- 基础路径：`/api/v1`
- 所有路径参数拼接至基础路径后
- 查询参数通过 `URLSearchParams` 自动编码
