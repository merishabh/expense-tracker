package utils

import (
	"time"
)

// ResolvePeriod converts a Period enum string to start and end UTC timestamps
func ResolvePeriod(period string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	var start, end time.Time

	switch period {
	case "TODAY":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)

	case "YESTERDAY":
		yesterday := now.AddDate(0, 0, -1)
		start = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
		end = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999999999, time.UTC)

	case "THIS_WEEK":
		// Week starts on Monday (ISO 8601)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday becomes 7
		}
		daysFromMonday := weekday - 1
		startOfWeek := now.AddDate(0, 0, -daysFromMonday)
		start = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, time.UTC)
		end = now

	case "LAST_WEEK":
		// Week starts on Monday (ISO 8601)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday becomes 7
		}
		daysFromMonday := weekday - 1
		startOfThisWeek := now.AddDate(0, 0, -daysFromMonday)
		startOfLastWeek := startOfThisWeek.AddDate(0, 0, -7)
		endOfLastWeek := startOfThisWeek.AddDate(0, 0, -1)
		start = time.Date(startOfLastWeek.Year(), startOfLastWeek.Month(), startOfLastWeek.Day(), 0, 0, 0, 0, time.UTC)
		end = time.Date(endOfLastWeek.Year(), endOfLastWeek.Month(), endOfLastWeek.Day(), 23, 59, 59, 999999999, time.UTC)

	case "THIS_MONTH":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = now

	case "LAST_MONTH":
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		lastOfLastMonth := firstOfThisMonth.AddDate(0, 0, -1)
		firstOfLastMonth := time.Date(lastOfLastMonth.Year(), lastOfLastMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		start = firstOfLastMonth
		end = time.Date(lastOfLastMonth.Year(), lastOfLastMonth.Month(), lastOfLastMonth.Day(), 23, 59, 59, 999999999, time.UTC)

	default:
		return time.Time{}, time.Time{}, &InvalidPeriodError{Period: period}
	}

	return start, end, nil
}

// InvalidPeriodError represents an invalid period error
type InvalidPeriodError struct {
	Period string
}

func (e *InvalidPeriodError) Error() string {
	return "invalid period: " + e.Period
}
