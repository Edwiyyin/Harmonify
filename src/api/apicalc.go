package api

import (
	"fmt"
	"time"
	
)

func (s Song) FormattedDuration() string {
    seconds := s.Duration / 1000
    minutes := seconds / 60
    remainingSeconds := seconds % 60
    return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
}

func FormatReleaseDate(dateStr string) time.Time {
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return parsedDate
}

func (s Song) FormattedReleaseDate() string {
	if s.ReleaseDate.IsZero() {
		return "Unknown"
	}
	return s.ReleaseDate.Format("2 January 2006")
}
