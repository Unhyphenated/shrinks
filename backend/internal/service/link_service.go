package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/analytics"
	"github.com/Unhyphenated/shrinks-backend/internal/cache"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)

var (
	ErrLinkNotFound = errors.New("link not found")
	ErrNotOwner     = errors.New("not owner")
	ErrInvalidURL   = errors.New("invalid URL")
	ErrURLScheme    = errors.New("invalid URL scheme")
	ErrURLHost      = errors.New("invalid URL host")
)

type LinkProvider interface {
	Shorten(ctx context.Context, longURL string, userID *uint64) (string, error)
	Redirect(ctx context.Context, shortCode string, event *model.AnalyticsEvent) (string, error)
	GetLinkByCode(ctx context.Context, shortCode string) (*model.Link, error)
	GetUserLinks(ctx context.Context, userID uint64, limit int, offset int) ([]model.Link, int, error)
	DeleteLink(ctx context.Context, shortCode string, userID uint64) error
	GetGlobalStats(ctx context.Context) (*model.GlobalStatsResponse, error)
	RecordEventBackground(shortCode string, link *model.Link, e *model.AnalyticsEvent)
}

type LinkService struct {
	Store     storage.LinkStore // The Store interface is the dependency
	Cache     cache.Cache
	Analytics analytics.AnalyticsProvider
}

func NewLinkService(s storage.LinkStore, c cache.Cache, a analytics.AnalyticsProvider) *LinkService {
	return &LinkService{
		Store:     s,
		Cache:     c,
		Analytics: a,
	}
}

func (ls *LinkService) Shorten(ctx context.Context, longURL string, userID *uint64) (string, error) {
	err := validateURL(longURL)
	if err != nil {
		return "", err
	}

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

func (ls *LinkService) GetLinkByCode(ctx context.Context, shortCode string) (*model.Link, error) {
	return ls.Store.GetLinkByCode(ctx, shortCode)
}

func (ls *LinkService) GetUserLinks(ctx context.Context, userID uint64, limit int, offset int) ([]model.Link, int, error) {
	links, total, err := ls.Store.GetUserLinks(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user links: %w", err)
	}
	return links, total, nil
}

func (ls *LinkService) DeleteLink(ctx context.Context, shortCode string, userID uint64) error {
	err := ls.Store.DeleteLink(ctx, shortCode, userID)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	if ls.Cache != nil {
		err = ls.Cache.Delete(ctx, shortCode)
		if err != nil {
			return fmt.Errorf("failed to delete link from cache: %w", err)
		}
	}
	return nil
}

func (ls *LinkService) GetGlobalStats(ctx context.Context) (*model.GlobalStatsResponse, error) {
	totalLinks, err := ls.Store.GetTotalLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total links: %w", err)
	}

	totalRequests, err := ls.Store.GetTotalRequests(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total requests: %w", err)
	}

	return &model.GlobalStatsResponse{
		TotalLinks:    totalLinks,
		TotalRequests: totalRequests,
	}, nil
}

func (ls *LinkService) RecordEventBackground(shortCode string, link *model.Link, e *model.AnalyticsEvent) {
	bgCtx := context.Background()

	if ls.Cache != nil {
		_ = ls.Cache.Set(bgCtx, shortCode, link, 24*time.Hour)
	}

	if ls.Analytics != nil && e != nil {
		eCopy := *e
		eCopy.LinkID = link.ID
		eCopy.ClickedAt = time.Now()

		if err := ls.Analytics.RecordEvent(bgCtx, &eCopy); err != nil {
			log.Printf("failed to record analytics event: %v", err)
		}
	}
}

func validateURL(rawUrl string) error {
	if rawUrl == "" {
		return ErrInvalidURL
	}

	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrURLScheme
	}

	if parsed.Host == "" {
		return ErrURLHost
	}

	return nil
}
