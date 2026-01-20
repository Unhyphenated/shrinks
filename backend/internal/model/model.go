package model

import "time"

// User Models
type User struct {
	ID           uint64    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID uint64 `json:"user_id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type RefreshToken struct {
	ID        uint64    `db:"id"`
	UserID    uint64    `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// Link Models
type Link struct {
	ID        uint64    `db:"id" redis:"id"`
	UserID    *uint64   `db:"user_id" redis:"user_id,omitempty"`
	ShortCode string    `db:"short_code" redis:"short_code"`
	LongURL   string    `db:"long_url" redis:"long_url"`
	CreatedAt time.Time `db:"created_at" redis:"created_at"`
}

type CreateLinkRequest struct {
	URL string `json:"url"`
}

type CreateLinkResponse struct {
	ShortCode string `json:"short_code"`
	LongURL   string `json:"long_url"`
}

// Analytics Models
type AnalyticsEvent struct {
	ID         uint64    `db:"id"`
	LinkID     uint64    `db:"link_id"`
	IPAddress  string    `db:"ip_address"`
	UserAgent  string    `db:"user_agent"`
	DeviceType string    `db:"device_type"`
	OS         string    `db:"os"`
	Browser    string    `db:"browser"`
	ClickedAt  time.Time `db:"clicked_at"`
}

type AnalyticsSummary struct {
	LinkID          uint64            `db:"link_id"`
	Period          string            `db:"period"`
	TotalClicks     int               `db:"total_clicks"`
	UniqueVisitors  int               `db:"unique_visitors"`
	ClicksByDate    []ClicksByDate    `db:"clicks_by_date"`
	ClicksByDevice  []ClicksByDevice  `db:"clicks_by_device"`
	ClicksByBrowser []ClicksByBrowser `db:"clicks_by_browser"`
	ClicksByOS      []ClicksByOS      `db:"clicks_by_os"`
}

type ClicksByDate struct {
	Date   string `db:"date"`
	Clicks int    `db:"clicks"`
}

type ClicksByDevice struct {
	Device string `db:"device"`
	Clicks int    `db:"clicks"`
}

type ClicksByBrowser struct {
	Browser string `db:"browser"`
	Clicks  int    `db:"clicks"`
}

type ClicksByOS struct {
	OS     string `db:"os"`
	Clicks int    `db:"clicks"`
}
