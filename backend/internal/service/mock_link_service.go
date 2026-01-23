package service

import (
	"context"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
)

type MockLinkService struct {
	ShortenFn               func(ctx context.Context, longURL string, userID *uint64) (string, error)
	RedirectFn              func(ctx context.Context, shortCode string, event *model.AnalyticsEvent) (string, error)
	GetLinkByCodeFn         func(ctx context.Context, shortCode string) (*model.Link, error)
	GetUserLinksFn          func(ctx context.Context, userID uint64, limit int, offset int) ([]model.Link, int, error)
	DeleteLinkFn            func(ctx context.Context, shortCode string, userID uint64) error
	GetGlobalStatsFn        func(ctx context.Context) (*model.GlobalStatsResponse, error)
	RecordEventBackgroundFn func(shortCode string, link *model.Link, e *model.AnalyticsEvent)
}

func (m *MockLinkService) Shorten(ctx context.Context, longURL string, userID *uint64) (string, error) {
	if m.ShortenFn != nil {
		return m.ShortenFn(ctx, longURL, userID)
	}
	return "", nil
}

func (m *MockLinkService) Redirect(ctx context.Context, shortCode string, event *model.AnalyticsEvent) (string, error) {
	if m.RedirectFn != nil {
		return m.RedirectFn(ctx, shortCode, event)
	}
	return "", nil
}

func (m *MockLinkService) GetLinkByCode(ctx context.Context, shortCode string) (*model.Link, error) {
	if m.GetLinkByCodeFn != nil {
		return m.GetLinkByCodeFn(ctx, shortCode)
	}
	return nil, nil
}

func (m *MockLinkService) GetUserLinks(ctx context.Context, userID uint64, limit int, offset int) ([]model.Link, int, error) {
	if m.GetUserLinksFn != nil {
		return m.GetUserLinksFn(ctx, userID, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockLinkService) DeleteLink(ctx context.Context, shortCode string, userID uint64) error {
	if m.DeleteLinkFn != nil {
		return m.DeleteLinkFn(ctx, shortCode, userID)
	}
	return nil
}

func (m *MockLinkService) GetGlobalStats(ctx context.Context) (*model.GlobalStatsResponse, error) {
	if m.GetGlobalStatsFn != nil {
		return m.GetGlobalStatsFn(ctx)
	}
	return &model.GlobalStatsResponse{}, nil
}

func (m *MockLinkService) RecordEventBackground(shortCode string, link *model.Link, e *model.AnalyticsEvent) {
	if m.RecordEventBackgroundFn != nil {
		m.RecordEventBackgroundFn(shortCode, link, e)
	}
}
