package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func SanitizeSearchQuery(query string) string {

    re := regexp.MustCompile(`\s*\([^)]*\)`)
    query = re.ReplaceAllString(query, "")

    parts := strings.Split(query, "-")
    query = strings.TrimSpace(parts[0])

    return strings.TrimSpace(query)
}

func CalculateTotalPages(totalResults int) int {
    if totalResults <= 0 {
        return 1
    }
    const resultsPerPage = 8
    return (totalResults + resultsPerPage - 1) / resultsPerPage
}

func ParseDuration(durationStr string) int {
    duration, err := strconv.Atoi(durationStr)
    if err != nil {
        return 0
    }
    return duration
}

func Minus(a, b int) int {
    return a - b
}

func Plus(a, b int) int {
    return a + b
}

func UrlencodeTitle(s string) string {
    return url.QueryEscape(s)
}

func DurationMinutes(duration int) int {
    return (duration / 1000) / 60
}

func DurationSeconds(duration int) int {
    return (duration / 1000) % 60
}


func (s Song) FormattedDuration() string {
    seconds := s.Duration / 1000
    minutes := seconds / 60
    remainingSeconds := seconds % 60
    return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
}

func FormatReleaseDate(dateStr string) time.Time {
    if dateStr == "" {
        return time.Time{}
    }
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

func LoadConfig() error {
	configFile, err := os.Open("data/config.json")
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
