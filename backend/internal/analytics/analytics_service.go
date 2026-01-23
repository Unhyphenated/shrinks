package analytics

import (
	"context"
	"fmt"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
	"github.com/Unhyphenated/shrinks-backend/internal/util"
)

type AnalyticsProvider interface {
	RecordEvent(ctx context.Context, event *model.AnalyticsEvent) error
	RetrieveAnalytics(ctx context.Context, linkID uint64, periodString string) (*model.AnalyticsSummary, error)
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

func (as *AnalyticsService) RetrieveAnalytics(ctx context.Context, linkID uint64, periodString string) (*model.AnalyticsSummary, error) {
	period, err := util.ParsePeriodToTime(periodString)
	if err != nil {
		return nil, fmt.Errorf("failure to parse time: %w", err)
	}

	events, err := as.Store.GetAnalyticsEvents(ctx, linkID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics events: %w", err)
	}

	dateMap := make(map[string]int)
	deviceMap := make(map[string]int)
	browserMap := make(map[string]int)
	osMap := make(map[string]int)
	uniqueIPs := make(map[string]struct{})

	for _, event := range events {
		uniqueIPs[event.IPAddress] = struct{}{}
		dateMap[event.ClickedAt.Format("2006-01-02")]++
		deviceMap[event.DeviceType]++
		browserMap[event.Browser]++
		osMap[event.OS]++
	}

	summary := &model.AnalyticsSummary{
		LinkID:         linkID,
		Period:         periodString,
		TotalClicks:    len(events),
		UniqueVisitors: len(uniqueIPs),
	}

	for date, clicks := range dateMap {
		summary.ClicksByDate = append(summary.ClicksByDate, model.ClicksByDate{Date: date, Clicks: clicks})
	}
	for device, clicks := range deviceMap {
		summary.ClicksByDevice = append(summary.ClicksByDevice, model.ClicksByDevice{Device: device, Clicks: clicks})
	}
	for browser, clicks := range browserMap {
		summary.ClicksByBrowser = append(summary.ClicksByBrowser, model.ClicksByBrowser{Browser: browser, Clicks: clicks})
	}
	for os, clicks := range osMap {
		summary.ClicksByOS = append(summary.ClicksByOS, model.ClicksByOS{OS: os, Clicks: clicks})
	}

	return summary, nil
}
