package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"harmonify/src/calc"
)

type SpotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type SpotifyTrackResponse struct {
	Tracks struct {
		Items []struct {
			PreviewURL string `json:"preview_url"`
            Duration int `json:"duration_ms"`
        } `json:"items"`
    } `json:"tracks"`
}

type SearchFilters struct {
    StartDate   string `json:"startDate"`
    EndDate     string `json:"endDate"`
    SortBy      string `json:"sortBy"`
    SortOrder   string `json:"sortOrder"`
    MinDuration int    `json:"minDuration"`
    MaxDuration int    `json:"maxDuration"`
}

type Song struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Lyrics   string `json:"lyrics,omitempty"`
	CoverURL string `json:"cover_url,omitempty"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	PreviewURL string `json:"preview_url,omitempty"`
    Duration    int       `json:"duration"`
}

type Config struct {
	SpotifyClientID     string `json:"spotify_client_id"`
	SpotifyClientSecret string `json:"spotify_client_secret"`
}

var (
    config             Config
    spotifyAccessToken string
    spotifyTokenExpiry time.Time
)

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

func GetSpotifyAccessToken() (string, error) {
	if spotifyAccessToken != "" && time.Now().Before(spotifyTokenExpiry) {
		return spotifyAccessToken, nil
	}

	clientID := config.SpotifyClientID
	clientSecret := config.SpotifyClientSecret

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	spotifyAccessToken = tokenResp.AccessToken
	spotifyTokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return spotifyAccessToken, nil
}
const (
    MaxTotalResults = 1000
    ResultsPerPage  = 8
)

func SearchSpotifySongs(query string, page int, filters SearchFilters) ([]Song, int, error) {
    if query == "" {
        return nil, 0, fmt.Errorf("empty search query")
    }

    accessToken, err := GetSpotifyAccessToken()
    if err != nil {
        return nil, 0, fmt.Errorf("failed to get access token: %v", err)
    }

    offset := (page - 1) * ResultsPerPage 
    baseURL := "https://api.spotify.com/v1/search"
    params := url.Values{}
    params.Add("q", query)
    params.Add("type", "track")
    params.Add("limit", fmt.Sprintf("%d", ResultsPerPage))
    params.Add("offset", fmt.Sprintf("%d", offset))

    client := &http.Client{Timeout: 10 * time.Second}

    req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to create request: %v", err)
    }

    req.Header.Add("Authorization", "Bearer "+accessToken)
    resp, err := client.Do(req)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to execute request: %v", err)
    }
    defer resp.Body.Close()

    var searchResp struct {
        Tracks struct {
            Total int `json:"total"`
            Items []struct {
                ID          string    `json:"id"`
                Name        string    `json:"name"`
                Duration    int       `json:"duration_ms"`
                Artists     []struct {
                    Name string `json:"name"`
                } `json:"artists"`
                Album struct {
                    Images []struct {
                        URL string `json:"url"`
                    } `json:"images"`
                    ReleaseDate string `json:"release_date"`
                } `json:"album"`
            } `json:"items"`
        } `json:"tracks"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
        return nil, 0, fmt.Errorf("failed to decode response: %v", err)
    }

    var songs []Song
    for _, item := range searchResp.Tracks.Items {
        artist := ""
        if len(item.Artists) > 0 {
            artist = item.Artists[0].Name
        }

        coverURL := ""
        if len(item.Album.Images) > 0 {
            coverURL = item.Album.Images[0].URL
        }

        song := Song{
            ID:          item.ID,
            Title:       item.Name,
            Artist:      artist,
            Duration:    item.Duration,
            CoverURL:    coverURL,
            ReleaseDate: FormatReleaseDate(item.Album.ReleaseDate),
        }

        if passesFilters(song, filters) {
            songs = append(songs, song)
        }
    }

    switch filters.SortBy {
    case "date":
        if filters.SortOrder == "asc" {
            sort.Slice(songs, func(i, j int) bool {
                return songs[i].ReleaseDate.Before(songs[j].ReleaseDate)
            })
        } else {
            sort.Slice(songs, func(i, j int) bool {
                return songs[i].ReleaseDate.After(songs[j].ReleaseDate)
            })
        }
    case "title":
        if filters.SortOrder == "asc" {
            sort.Slice(songs, func(i, j int) bool {
                return songs[i].Title < songs[j].Title
            })
        } else {
            sort.Slice(songs, func(i, j int) bool {
                return songs[i].Title > songs[j].Title
            })
        }
    case "artist":
        if filters.SortOrder == "asc" {
            sort.Slice(songs, func(i, j int) bool {
                return songs[i].Artist < songs[j].Artist
            })
        } else {
            sort.Slice(songs, func(i, j int) bool {
                return songs[i].Artist > songs[j].Artist
            })
        }
    }

    totalResults := searchResp.Tracks.Total
    if totalResults > MaxTotalResults {
        totalResults = MaxTotalResults
    }

    return songs, totalResults, nil
}

func passesFilters(song Song, filters SearchFilters) bool {
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

func FetchLyricsOvh(title, artist string) (string, error) {
    sanitizedTitle := calc.SanitizeSearchQuery(title)
    
    encodedTitle := url.QueryEscape(sanitizedTitle)
    encodedArtist := url.QueryEscape(artist)

    apiURL := fmt.Sprintf("https://api.lyrics.ovh/v1/%s/%s", encodedArtist, encodedTitle)

    resp, err := http.Get(apiURL)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("no lyrics found")
    }

    var lyricsResp struct {
        Lyrics string `json:"lyrics"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&lyricsResp); err != nil {
        return "", err
    }

    lyrics := strings.TrimSpace(lyricsResp.Lyrics)
    if lyrics == "" {
        return "", fmt.Errorf("empty lyrics")
    }

    if len(lyrics) > 5000 {
        lyrics = lyrics[:5000] + "... (lyrics truncated)"
    }

    return lyrics, nil
}

func SearchSpotifyMusicSource(title, artist string) (string, error) {
    accessToken, err := GetSpotifyAccessToken()
    if err != nil {
        return "", fmt.Errorf("spotify token error: %v", err)
    }

    sanitizedTitle := calc.SanitizeSearchQuery(title)
    firstArtist := strings.Split(artist, ",")[0]
    firstArtist = strings.TrimSpace(firstArtist)

    query := fmt.Sprintf("track:%s artist:%s", sanitizedTitle, firstArtist)
    encodedQuery := url.QueryEscape(query)

    req, err := http.NewRequest("GET",
        fmt.Sprintf("https://api.spotify.com/v1/search?q=%s&type=track&limit=1", encodedQuery),
        nil)
    if err != nil {
        return "", err
    }

    req.Header.Add("Authorization", "Bearer "+accessToken)
    req.Header.Add("Content-Type", "application/json")

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var trackResp SpotifyTrackResponse
    if err := json.NewDecoder(resp.Body).Decode(&trackResp); err != nil {
        return "", fmt.Errorf("failed to decode Spotify response: %v", err)
    }

    if len(trackResp.Tracks.Items) > 0 {
        return trackResp.Tracks.Items[0].PreviewURL, nil
    }

    return "", fmt.Errorf("no preview URL found")
}
