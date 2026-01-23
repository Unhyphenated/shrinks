// User types
export interface User {
  id: number;
  email: string;
  created_at?: string;
}

// Auth types
export interface RegisterRequest {
  email: string;
  password: string;
}

export interface RegisterResponse {
  user_id: number;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  access_token: string;
}

export interface LogoutRequest {
  refresh_token: string;
}

// Link types
export interface Link {
  id: number;
  short_code: string;
  long_url: string;
  created_at: string;
  user_id?: number;
}

export interface CreateLinkRequest {
  url: string;
}

export interface CreateLinkResponse {
  short_code: string;
  long_url: string;
}

export interface LinksResponse {
  links: Link[];
  total: number;
}

// Analytics types
export interface ClicksByDate {
  date: string;
  clicks: number;
}

export interface ClicksByDevice {
  device: string;
  clicks: number;
}

export interface ClicksByBrowser {
  browser: string;
  clicks: number;
}

export interface ClicksByOS {
  os: string;
  clicks: number;
}

export interface AnalyticsSummary {
  link_id: number;
  period: string;
  total_clicks: number;
  unique_visitors: number;
  clicks_by_date: ClicksByDate[];
  clicks_by_device: ClicksByDevice[];
  clicks_by_browser: ClicksByBrowser[];
  clicks_by_os: ClicksByOS[];
}

// API Error type
export interface ApiError {
  error: string;
  message?: string;
}

// View state
export type ViewState = 'home' | 'analytics' | 'login' | 'forgot-password' | 'email-preview' | 'links';
