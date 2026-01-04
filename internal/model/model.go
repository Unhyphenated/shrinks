package model

import "time"

type User struct {
	ID uint64 `db:"id"`
	Email string `db:"email"`
	PasswordHash string `db:"password_hash"`
	CreatedAt time.Time `db:"created_at"`
}

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
	ShortURL string `json:"short_url"`
	LongURL string `json:"long_url"`
}