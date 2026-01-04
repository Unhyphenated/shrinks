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