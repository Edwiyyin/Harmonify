package api

import (
	"encoding/json"
	"fmt"
	"os"
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


func PassesFilters(song Song, filters SearchFilters) bool {
    if filters.StartDate != "" {
        startDate, err := time.Parse("2006-01-02", filters.StartDate)
        if err == nil && song.ReleaseDate.Before(startDate) {
            return false
        }
    }
    if filters.EndDate != "" {
        endDate, err := time.Parse("2006-01-02", filters.EndDate)
        if err == nil && song.ReleaseDate.After(endDate) {
            return false
        }
    }

    if filters.MinDuration > 0 && song.Duration < filters.MinDuration*1000 {
        return false
    }
    if filters.MaxDuration > 0 && song.Duration > filters.MaxDuration*1000 {
        return false
    }

    return true
}

func LoadConfig() error {
	configFile, err := os.Open("config.json")
	if err != nil {
		return fmt.Errorf("error opening config file: %v", err)
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("error parsing config file: %v", err)
	}

	return nil
}