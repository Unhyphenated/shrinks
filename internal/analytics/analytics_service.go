package analytics

import (
	"context"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)

type AnalyticsProvider interface {
	RecordEvent(ctx context.Context, event *model.AnalyticsEvent) error
	RetrieveAnalytics(ctx context.Context, linkID uint64, period string) (*model.AnalyticsSummary, error)
	// GetUserLinksWithStats(ctx context.Context, userID uint64) ([]*model.LinkWithStats, error)
}
type AnalyticsService struct {
	Store storage.AnalyticsStore
}

func NewAnalyticsService(store storage.AnalyticsStore) *AnalyticsService {
	return &AnalyticsService{Store: store}
}

func (as *AnalyticsService) RecordEvent(ctx context.Context, event *model.AnalyticsEvent) error {
	err := as.Store.SaveAnalyticsEvent(ctx, event)

	if err != nil {
		return err
	}

	return nil
}

func (as *AnalyticsService) RetrieveAnalytics(ctx context.Context, linkID uint64, period string) (*model.AnalyticsSummary, error) {
	// as.Store.GetAnalyticsEvents
	// sql aggregation for clicks, browser stats, device stats, etc

	// Return &model.AnalyticsSummary
	return nil, nil
}

// func (as *AnalyticsService) RetrieveAnalytics(ctx context.Context, linkID uint64, period string) (*model.AnalyticsSummary, error) {
// 	// TODO: parse period ("24h", "7d", "30d") and calculate date range
// 	// TODO: get raw events from store.GetAnalyticsEvents()
// 	// TODO: aggregate events:
// 	//   - count total clicks
// 	//   - count unique IPs (unique visitors)
// 	//   - group by date (daily buckets)
// 	//   - group by device type
// 	//   - group by browser
// 	//   - group by OS
// 	// TODO: return AnalyticsSummary model
// }

// func (as *AnalyticsService) GetUserLinksWithStats(ctx context.Context, userID uint64) ([]*model.LinkWithStats, error) {
// 	// TODO: call store.GetUserLinksWithStats()
// 	// TODO: return list of links with total_clicks for each
// }
