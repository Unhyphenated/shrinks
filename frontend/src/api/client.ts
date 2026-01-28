import type {
  User,
  AuthResponse,
  RegisterResponse,
  RefreshTokenResponse,
  CreateLinkResponse,
  LinksResponse,
  AnalyticsSummary,
  GlobalStats,
  ApiError,
} from "../types";

export const SHORT_DOMAIN = import.meta.env.VITE_SHORT_DOMAIN || window.location.host;
const API_BASE_URL = import.meta.env.VITE_API_URL || "";

// Token storage keys
const ACCESS_TOKEN_KEY = "access_token";
const REFRESH_TOKEN_KEY = "refresh_token";

// Token storage with localStorage for persistence
let accessToken: string | null = localStorage.getItem(ACCESS_TOKEN_KEY);
let refreshToken: string | null = localStorage.getItem(REFRESH_TOKEN_KEY);
let onUnauthorized: (() => void) | null = null;

export const setTokens = (access: string, refresh: string) => {
  accessToken = access;
  refreshToken = refresh;
  localStorage.setItem(ACCESS_TOKEN_KEY, access);
  localStorage.setItem(REFRESH_TOKEN_KEY, refresh);
};

export const getAccessToken = () => accessToken;
export const getRefreshToken = () => refreshToken;

export const clearTokens = () => {
  accessToken = null;
  refreshToken = null;
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
};

export const setOnUnauthorized = (callback: () => void) => {
  onUnauthorized = callback;
};

class ApiClient {
  private baseUrl: string;
  private isRefreshing = false;
  private refreshPromise: Promise<boolean> | null = null;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    requiresAuth = false,
    retry = true,
  ): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    // Merge existing headers
    if (options.headers) {
      const existingHeaders = options.headers as Record<string, string>;
      Object.assign(headers, existingHeaders);
    }

    if (accessToken && (requiresAuth || accessToken)) {
      headers["Authorization"] = `Bearer ${accessToken}`;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers,
    });

    // Handle 401 Unauthorized
    if (response.status === 401 && retry && refreshToken) {
      const refreshed = await this.attemptRefresh();
      if (refreshed) {
        return this.request<T>(endpoint, options, requiresAuth, false);
      } else {
        onUnauthorized?.();
        throw new Error("Session expired. Please log in again.");
      }
    }

    if (!response.ok) {
      let errorMessage = `HTTP error ${response.status}`;
      try {
        const errorData: ApiError = await response.json();
        errorMessage = errorData.error || errorData.message || errorMessage;
      } catch {
        // Response is not JSON
      }
      throw new Error(errorMessage);
    }

    // Handle 204 No Content
    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  }

  private async attemptRefresh(): Promise<boolean> {
    if (this.isRefreshing) {
      return this.refreshPromise!;
    }

    this.isRefreshing = true;
    this.refreshPromise = this.doRefresh();

    try {
      return await this.refreshPromise;
    } finally {
      this.isRefreshing = false;
      this.refreshPromise = null;
    }
  }

  private async doRefresh(): Promise<boolean> {
    if (!refreshToken) return false;

    try {
      const response = await fetch(`${this.baseUrl}/api/v1/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) {
        clearTokens();
        return false;
      }

      const data: RefreshTokenResponse = await response.json();
      accessToken = data.access_token;
      return true;
    } catch {
      clearTokens();
      return false;
    }
  }

  // Auth endpoints
  async register(email: string, password: string): Promise<RegisterResponse> {
    return this.request<RegisterResponse>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  }

  async login(email: string, password: string): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    setTokens(response.access_token, response.refresh_token);
    return response;
  }

  async logout(): Promise<void> {
    if (refreshToken) {
      try {
        await this.request("/api/v1/auth/logout", {
          method: "POST",
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
      } catch {
        // Ignore logout errors
      }
    }
    clearTokens();
  }

  async refreshAccessToken(): Promise<RefreshTokenResponse> {
    if (!refreshToken) {
      throw new Error("No refresh token available");
    }
    const response = await this.request<RefreshTokenResponse>(
      "/api/v1/auth/refresh",
      {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
      },
    );
    accessToken = response.access_token;
    return response;
  }

  // Link endpoints
  async shortenUrl(url: string): Promise<CreateLinkResponse> {
    return this.request<CreateLinkResponse>("/api/v1/links/shorten", {
      method: "POST",
      body: JSON.stringify({ url }),
    });
  }

  async getLinks(limit = 10, offset = 0): Promise<LinksResponse> {
    return this.request<LinksResponse>(
      `/api/v1/links?limit=${limit}&offset=${offset}`,
      { method: "GET" },
      true,
    );
  }

  async deleteLink(shortCode: string): Promise<void> {
    return this.request<void>(
      `/api/v1/links/${shortCode}`,
      { method: "DELETE" },
      true,
    );
  }

  // Analytics endpoint
  async getAnalytics(
    shortCode: string,
    period = "7d",
  ): Promise<AnalyticsSummary> {
    return this.request<AnalyticsSummary>(
      `/api/v1/links/${shortCode}/analytics?period=${period}`,
      { method: "GET" },
      true,
    );
  }

  // Global stats endpoint (public, no auth required)
  async getGlobalStats(): Promise<GlobalStats> {
    return this.request<GlobalStats>("/api/v1/links/stats", {
      method: "GET",
    });
  }

  async getCurrentUser(): Promise<User> {
    return this.request<User>(
      "/api/v1/auth/me",
      {
        method: "GET",
      },
      true,
    );
  }
}

export const apiClient = new ApiClient(API_BASE_URL);
