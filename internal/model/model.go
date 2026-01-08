package model

import "time"

// User Models
type User struct {
	ID uint64 `db:"id"`
	Email string `db:"email"`
	PasswordHash string `db:"password_hash"`
	CreatedAt time.Time `db:"created_at"`
}

type RegisterRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID uint64 `json:"user_id"`
}

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User User `json:"user"`
}

// Link Models
type Link struct {
	ID uint64 `db:"id"`
	ShortCode string `db:"short_code"`
	LongURL string `db:"long_url"`
	CreatedAt time.Time `db:"created_at"`
}

type CreateLinkRequest struct {
	URL string `json:"url"`
	CustomCode string `json:"custom_code"`
}

type CreateLinkResponse struct {
	ShortCode string `json:"short_code"`
	LongURL string `json:"long_url"`
}

// Analytics Models
type AnalyticsEvent struct {
	ID uint64 `db:"id"`
	LinkID uint64 `db:"link_id"`
	IPAddress string `db:"ip_address"`
	UserAgent string `db:"user_agent"`
	DeviceType string `db:"device_type"`
	OS string `db:"os"`
	Browser string `db:"browser"`
	ClickedAt time.Time `db:"clicked_at"`
}