package articlelisting

import (
	"fmt"
	"time"
)

var frenchMonthsLong = [12]string{
	"janvier", "février", "mars", "avril", "mai", "juin",
	"juillet", "août", "septembre", "octobre", "novembre", "décembre",
}

// formatLong renders an ISO date (2006-01-02) as "21 décembre 2025".
// It returns an error if date is empty or malformed.
func formatLong(date string) (string, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", fmt.Errorf("formatLong(%q): %w", date, err)
	}
	return fmt.Sprintf("%d %s %d", t.Day(), frenchMonthsLong[t.Month()-1], t.Year()), nil
}
