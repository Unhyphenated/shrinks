package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParsePeriodToTime(period string) (time.Time, error) {
	period = strings.ToLower(strings.TrimSpace(period))
	var duration time.Duration
	var err error

	if strings.HasSuffix(period, "d") {
		daysStr := strings.TrimSuffix(period, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid days format")
		}
		duration = time.Hour * 24 * time.Duration(days)
	} else {
		duration, err = time.ParseDuration(period)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid duration format")
		}
	}

	// Always use UTC for database consistency
	return time.Now().UTC().Add(-duration), nil
}
