package subscription

import (
	"fmt"
	"time"
)

const monthYearLayout = "01-2006"

func ParseMonthYear(value string) (time.Time, error) {
	parsed, err := time.Parse(monthYearLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: expected MM-YYYY", ErrInvalidDate)
	}

	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func FormatMonthYear(value time.Time) string {
	return value.Format(monthYearLayout)
}
