package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/analytics"
	"github.com/Unhyphenated/shrinks-backend/internal/cache"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)

type LinkService struct {
	Store storage.LinkStore // The Store interface is the dependency
	Cache cache.Cache
	AnalyticsService analytics.AnalyticsProvider
}

func NewLinkService(s storage.LinkStore, c cache.Cache, a analytics.AnalyticsProvider) *LinkService {
	return &LinkService{Store: s, Cache: c}
}

func (ls *LinkService) Shorten(ctx context.Context, longURL string, userID *uint64) (string, error) {
	// Rate Limiting
	// Input Sanitation

	shortCode, err := ls.Store.SaveLink(ctx, longURL, userID)
	if err != nil {
		return "", fmt.Errorf("failed to save link: %w", err)
	}
	return shortCode, nil
}

func (ls *LinkService) Redirect(ctx context.Context, shortCode string, event *model.AnalyticsEvent) (string, error) {
	// Check if link is in cache
	longURL, err := ls.Cache.Get(ctx, shortCode)
	if err != nil {
		log.Printf("Cache error (falling back to DB): %v", err)
	}

	if longURL != "" {
		return longURL, nil
	}

	link, err := ls.Store.GetLinkByCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to get link by code: %w", err)
	}

	if link == nil {
		return "", fmt.Errorf("link not found")
	}

	if ls.Cache != nil {
		go func() {
			if err := ls.Cache.Set(context.Background(), shortCode, link.LongURL, 24*time.Hour); err != nil {
				log.Printf("Failed to cache link: %v", err)
			}
		}()
	}

	event.LinkID = link.ID
    event.ClickedAt = time.Now()

	go func() {
		if err := ls.AnalyticsService.RecordEvent(context.Background(), event); err != nil {
			log.Printf("Failed to record analytics event: %v", err)
		}
	}()

	return link.LongURL, nil
}
