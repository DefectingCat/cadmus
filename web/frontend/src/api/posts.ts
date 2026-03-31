// 文章 API 调用
import { apiClient, APIResponse, APIError } from './client'

// 文章类型定义
export interface Post {
  id: string
  author_id: string
  title: string
  slug: string
  content: string
  excerpt?: string
  category_id?: string
  status: 'draft' | 'published' | 'scheduled' | 'private'
  publish_at?: string
  featured_image?: string
  seo_meta: {
    title?: string
    description?: string
    keywords?: string[]
  }
  view_count: number
  like_count: number
  comment_count: number
  series_id?: string
  series_order: number
  is_paid: boolean
  price?: number
  version: number
  created_at: string
  updated_at: string
}

export interface PostListResponse {
  posts: Post[]
  total: number
  page: number
  page_size: number
}

export interface CreatePostRequest {
  title: string
  slug: string
  content: string
  content_text?: string
  excerpt?: string
  category_id?: string
  status?: string
  featured_image?: string
  seo_title?: string
  seo_description?: string
  seo_keywords?: string[]
  series_id?: string
  series_order?: number
  is_paid?: boolean
  price?: number
  tag_ids?: string[]
}

export interface UpdatePostRequest {
  title: string
  slug: string
  content: string
  content_text?: string
  excerpt?: string
  category_id?: string
  status?: string
  featured_image?: string
  seo_title?: string
  seo_description?: string
  seo_keywords?: string[]
  series_id?: string
  series_order?: number
  is_paid: boolean
  price?: number
  tag_ids?: string[]
}

export interface PostListFilters {
  status?: string
  author_id?: string
  category_id?: string
  search?: string
  page?: number
  page_size?: number
}

// 文章 API 服务
export const postsAPI = {
  // 获取文章列表
  list: async (filters?: PostListFilters): Promise<APIResponse<PostListResponse>> => {
    const params: Record<string, string> = {}
    if (filters) {
      if (filters.status) params.status = filters.status
      if (filters.author_id) params.author_id = filters.author_id
      if (filters.category_id) params.category_id = filters.category_id
      if (filters.search) params.search = filters.search
      if (filters.page) params.page = String(filters.page)
      if (filters.page_size) params.page_size = String(filters.page_size)
    }
    return apiClient.get<PostListResponse>('/posts', params)
  },

  // 获取单篇文章
  get: async (id: string): Promise<APIResponse<Post>> => {
    return apiClient.get<Post>(`/posts/${id}`)
  },

  // 通过 slug 获取文章
  getBySlug: async (slug: string): Promise<APIResponse<Post>> => {
    return apiClient.get<Post>(`/posts/${slug}`)
  },

  // 创建文章
  create: async (data: CreatePostRequest): Promise<APIResponse<Post>> => {
    return apiClient.post<Post>('/posts', data)
  },

  // 更新文章
  update: async (id: string, data: UpdatePostRequest): Promise<APIResponse<Post>> => {
    return apiClient.put<Post>(`/posts/${id}`, data)
  },

  // 删除文章
  delete: async (id: string): Promise<APIResponse<void>> => {
    return apiClient.delete<void>(`/posts/${id}`)
  },

  // 发布文章
  publish: async (id: string): Promise<APIResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`/posts/${id}/publish`)
  },

  // 批量删除文章
  batchDelete: async (ids: string[]): Promise<APIResponse<{ message: string }>> => {
    // 目前没有专门的批量删除 API，逐个删除
    const results = await Promise.all(ids.map(id => postsAPI.delete(id)))
    const failed = results.filter((r: APIResponse<void>) => r.error)
    if (failed.length > 0) {
      return {
        error: {
          code: 'BATCH_DELETE_ERROR',
          message: `删除失败 ${failed.length}/${ids.length} 篇文章`,
        },
      }
    }
    return { data: { message: `成功删除 ${ids.length} 篇文章` } }
  },

  // 批量发布文章
  batchPublish: async (ids: string[]): Promise<APIResponse<{ message: string }>> => {
    const results = await Promise.all(ids.map(id => postsAPI.publish(id)))
    const failed = results.filter((r: APIResponse<{ message: string }>) => r.error)
    if (failed.length > 0) {
      return {
        error: {
          code: 'BATCH_PUBLISH_ERROR',
          message: `发布失败 ${failed.length}/${ids.length} 篇文章`,
        },
      }
    }
    return { data: { message: `成功发布 ${ids.length} 篇文章` } }
  },

  // 获取版本历史
  versions: async (id: string): Promise<APIResponse<unknown[]>> => {
    return apiClient.get<unknown[]>(`/posts/${id}/versions`)
  },

  // 回滚到指定版本
  rollback: async (id: string, version: number): Promise<APIResponse<{ message: string }>> => {
    return apiClient.post<{ message: string }>(`/posts/${id}/rollback`, { version })
  },
}