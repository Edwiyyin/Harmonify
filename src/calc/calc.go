package calc

import (
    "net/url"
    "regexp"
    "strconv"
    "strings"
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
    if durationStr == "" {
        return 0
    }
    duration, _ := strconv.Atoi(durationStr)
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
    return duration / 60
}

func DurationSeconds(duration int) int {
    return duration % 60
}
