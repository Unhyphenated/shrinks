package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/analytics"
	"github.com/Unhyphenated/shrinks-backend/internal/cache"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)

var (
	ErrLinkNotFound = errors.New("link not found")
)

type LinkService struct {
	Store storage.LinkStore // The Store interface is the dependency
	Cache cache.Cache
	Analytics analytics.AnalyticsProvider
	mu sync.RWMutex
}

func NewLinkService(s storage.LinkStore, c cache.Cache, a analytics.AnalyticsProvider) *LinkService {
	return &LinkService{
		Store: s, 
		Cache: c, 
		Analytics: a, 
	}
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
	link, err := ls.Cache.Get(ctx, shortCode)
	if err != nil {
		log.Printf("cache error (falling back to DB): %v", err)
	}
    
	// Cache hit
    if link != nil {
        go ls.RecordEventBackground(shortCode, link, event)
        return link.LongURL, nil
    }

	// Check if link is in DB
	link, err = ls.Store.GetLinkByCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to get link by code: %w", err)
	}

	if link == nil {
		return "", ErrLinkNotFound
	}
	
	go ls.RecordEventBackground(shortCode, link, event)

	return link.LongURL, nil
}

func (ls *LinkService) RecordEventBackground(shortCode string, link *model.Link, e *model.AnalyticsEvent) {
	bgCtx := context.Background()

	if ls.Cache != nil {
		_ = ls.Cache.Set(bgCtx, shortCode, link, 24*time.Hour)
	}

	if ls.Analytics != nil  && e != nil {
		eCopy := *e
		eCopy.LinkID = link.ID
		eCopy.ClickedAt = time.Now()

		ls.mu.Lock()
		defer ls.mu.Unlock()

		if err := ls.Analytics.RecordEvent(bgCtx, &eCopy); err != nil {
			log.Printf("failed to record analytics event: %v", err)
		}
	}
}
