package model

import "time"

type Link struct {
	ID uint64 `db:"id"`
	ShortURL string `db:"short_url"`
	LongURL string `db:"long_url"`
	CreatedAt time.Time `db:"created_at"`
	Clicks uint64 `db:"clicks"`
}