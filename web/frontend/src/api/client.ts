// API 基础请求客户端
const API_BASE = "/api/v1";

export interface APIError {
  code: string;
  message: string;
  details?: string[];
  request_id?: string;
}

export interface APIResponse<T> {
  data?: T;
  error?: APIError;
}

export class APIClient {
  private baseUrl: string;
  private token: string | null = null;

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl;
    this.loadToken();
  }

  private loadToken(): void {
    // 从 cookie 加载 token
    const cookies = document.cookie.split(";");
    for (const cookie of cookies) {
      const [name, value] = cookie.trim().split("=");
      if (name === "auth_token") {
        this.token = value;
        break;
      }
    }
    // 也尝试从 localStorage 加载
    const stored = localStorage.getItem("auth_token");
    if (stored) {
      this.token = stored;
    }
  }

  setToken(token: string): void {
    this.token = token;
    localStorage.setItem("auth_token", token);
  }

  clearToken(): void {
    this.token = null;
    localStorage.removeItem("auth_token");
  }

  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      "Content-Type": "application/json",
    };
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }
    return headers;
  }

  async get<T>(path: string, params?: Record<string, string>): Promise<APIResponse<T>> {
    let url = `${this.baseUrl}${path}`;
    if (params) {
      const query = new URLSearchParams(params);
      url += `?${query}`;
    }

    try {
      const response = await fetch(url, {
        method: "GET",
        headers: this.getHeaders(),
      });

      if (!response.ok) {
        const error = (await response.json()) as APIError;
        return { error };
      }

      const data = (await response.json()) as T;
      return { data };
    } catch (e) {
      return {
        error: {
          code: "NETWORK_ERROR",
          message: "网络请求失败",
        },
      };
    }
  }

  async post<T>(path: string, body?: unknown): Promise<APIResponse<T>> {
    try {
      const response = await fetch(`${this.baseUrl}${path}`, {
        method: "POST",
        headers: this.getHeaders(),
        body: body ? JSON.stringify(body) : undefined,
      });

      if (!response.ok) {
        const error = (await response.json()) as APIError;
        return { error };
      }

      const data = (await response.json()) as T;
      return { data };
    } catch (e) {
      return {
        error: {
          code: "NETWORK_ERROR",
          message: "网络请求失败",
        },
      };
    }
  }

  async put<T>(path: string, body: unknown): Promise<APIResponse<T>> {
    try {
      const response = await fetch(`${this.baseUrl}${path}`, {
        method: "PUT",
        headers: this.getHeaders(),
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const error = (await response.json()) as APIError;
        return { error };
      }

      const data = (await response.json()) as T;
      return { data };
    } catch (e) {
      return {
        error: {
          code: "NETWORK_ERROR",
          message: "网络请求失败",
        },
      };
    }
  }

  async delete<T>(path: string): Promise<APIResponse<T>> {
    try {
      const response = await fetch(`${this.baseUrl}${path}`, {
        method: "DELETE",
        headers: this.getHeaders(),
      });

      if (!response.ok) {
        const error = (await response.json()) as APIError;
        return { error };
      }

      if (response.status === 204) {
        return { data: undefined as T };
      }

      const data = (await response.json()) as T;
      return { data };
    } catch (e) {
      return {
        error: {
          code: "NETWORK_ERROR",
          message: "网络请求失败",
        },
      };
    }
  }
}

// 全局 API 客户端实例
export const apiClient = new APIClient();
