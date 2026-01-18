package analytics

import (
	"context"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
)

type MockAnalytics struct {
	RecordEventFn func(ctx context.Context, event *model.AnalyticsEvent) error
	RetrieveAnalyticsFn func(ctx context.Context, linkID uint64, period string) (*model.AnalyticsSummary, error)
	
}

// Ensure MockAnalytics implements AnalyticsProvider interface
var _ AnalyticsProvider = (*MockAnalytics)(nil)

func (m *MockAnalytics) RecordEvent(ctx context.Context, event *model.AnalyticsEvent) error {
	if m.RecordEventFn != nil {
		return m.RecordEventFn(ctx, event)
	}
	return nil // Default: no-op
}

func (m *MockAnalytics) RetrieveAnalytics(ctx context.Context, linkID uint64, period string) (*model.AnalyticsSummary, error) {
	if m.RetrieveAnalyticsFn != nil {
		return m.RetrieveAnalyticsFn(ctx, linkID, period)
	}
	return nil, nil // Default: no-op
}