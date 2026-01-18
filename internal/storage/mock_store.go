package storage

import (
	"context"
	"time"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
)


type MockStore struct {
	SaveLinkFn    func(ctx context.Context, longURL string, userID *uint64) (string, error)
	GetLinkByCodeFn func(ctx context.Context, shortURL string) (*model.Link, error)
	GetUserLinksFn func(ctx context.Context, userID uint64, limit int, offset int) ([]model.Link, int, error)
	DeleteLinkFn func(ctx context.Context, shortCode string, userID uint64) error
	GetAnalyticsEventsFn func(ctx context.Context, linkID uint64, startDate, endDate time.Time) ([]*model.AnalyticsEvent, error)
	CloseFn       func()
}

var _ LinkStore = (*MockStore)(nil)

func (m *MockStore) SaveLink(ctx context.Context, longURL string, userID *uint64) (string, error) {
	return m.SaveLinkFn(ctx, longURL, userID) 
}

func (m *MockStore) GetLinkByCode(ctx context.Context, shortURL string) (*model.Link, error) {
	return m.GetLinkByCodeFn(ctx, shortURL)
}

func (m *MockStore) GetUserLinks(ctx context.Context, userID uint64, limit int, offset int) ([]model.Link, int, error) {
	return m.GetUserLinksFn(ctx, userID, limit, offset)
}

func (m *MockStore) DeleteLink(ctx context.Context, shortCode string, userID uint64) error {
	return m.DeleteLinkFn(ctx, shortCode, userID)
}

func (m *MockStore) GetAnalyticsEvents(ctx context.Context, linkID uint64, startDate, endDate time.Time) ([]*model.AnalyticsEvent, error) {
	return m.GetAnalyticsEventsFn(ctx, linkID, startDate, endDate)
}

func (m *MockStore) Close() {
	m.CloseFn()
}