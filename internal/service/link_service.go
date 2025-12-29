package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/cache"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)

type LinkService struct {
	Store storage.Store // The Store interface is the dependency
	Cache cache.Cache
}

func NewLinkService(s storage.Store, c cache.Cache) *LinkService {
	return &LinkService{Store: s, Cache: c}
}

func (ls *LinkService) Shorten(ctx context.Context, longURL string) (string, error) {
	// Rate Limiting
	// Input Sanitation

	shortURL, err := ls.Store.SaveLink(ctx, longURL)
	if err != nil {
		return "", fmt.Errorf("failed to save link: %w", err)
	}
	return shortURL, nil
}

func (ls *LinkService) Redirect(ctx context.Context, shortURL string) (string, error) {
	// Check if link is in cache
	longURL, err := ls.Cache.Get(ctx, shortURL)
	if err != nil {
		log.Printf("Cache error (falling back to DB): %v", err)
	}

	if longURL != "" {
		return longURL, nil
	}

	link, err := ls.Store.GetLinkByCode(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to get link by code: %w", err)
	}

	if link == nil {
		return "", fmt.Errorf("link not found")
	}

	if ls.Cache != nil {
		go func() {
			if err := ls.Cache.Set(context.Background(), shortURL, link.LongURL, 24*time.Hour); err != nil {
				log.Printf("Failed to cache link: %v", err)
			}
		}()
	}

	ls.Store.UpdateClickCount(ctx, link.ID)
	return link.LongURL, nil
}
