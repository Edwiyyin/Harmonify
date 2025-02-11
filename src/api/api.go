package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"os"

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

func SearchSpotifySongs(query string, page int, filters SearchFilters) ([]Song, int, error) {
    accessToken, err := GetSpotifyAccessToken()
    if err != nil {
        return nil, 0, fmt.Errorf("spotify token error: %v", err)
    }

    sanitizedQuery := calc.SanitizeSearchQuery(query)
    encodedQuery := url.QueryEscape(sanitizedQuery)

    req, err := http.NewRequest("GET", 
        fmt.Sprintf("https://api.spotify.com/v1/search?q=%s&type=track&limit=50&offset=%d", 
        encodedQuery, (page-1)*50), nil)
    if err != nil {
        return nil, 0, err
    }

    req.Header.Add("Authorization", "Bearer "+accessToken)
    req.Header.Add("Content-Type", "application/json")

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, 0, err
    }
    defer resp.Body.Close()

    var spotifyResp struct {
        Tracks struct {
            Total int `json:"total"`
            Items []struct {
                ID     string `json:"id"`
                Name   string `json:"name"`
                Artists []struct {
                    Name string `json:"name"`
                } `json:"artists"`
                Album struct {
                    Images []struct {
                        URL string `json:"url"`
                    } `json:"images"`
                    ReleaseDate string `json:"release_date"`
                } `json:"album"`
                Duration int `json:"duration_ms"`
                PreviewURL string `json:"preview_url"`
            } `json:"items"`
        } `json:"tracks"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&spotifyResp); err != nil {
        return nil, 0, err
    }

    var songs []Song
    startDate, _ := time.Parse("2006-01-02", filters.StartDate)
    endDate, _ := time.Parse("2006-01-02", filters.EndDate)

    for _, track := range spotifyResp.Tracks.Items {
        var coverURL string
        var releaseDate time.Time

        if len(track.Album.Images) > 0 {
            coverURL = track.Album.Images[0].URL
        }

        if track.Album.ReleaseDate != "" {
            releaseDate = FormatReleaseDate(track.Album.ReleaseDate)
        }

        if !startDate.IsZero() && releaseDate.Before(startDate) {
            continue
        }
        if !endDate.IsZero() && releaseDate.After(endDate) {
            continue
        }

        if filters.MinDuration > 0 && track.Duration < filters.MinDuration*1000 {
            continue
        }
        if filters.MaxDuration > 0 && track.Duration > filters.MaxDuration*1000 {
            continue
        }

        songs = append(songs, Song{
            ID:          track.ID,
            Title:       track.Name,
            Artist:      track.Artists[0].Name,
            CoverURL:    coverURL,
            ReleaseDate: releaseDate,
            Duration:    track.Duration,
            PreviewURL:  track.PreviewURL,
        })
    }

    switch filters.SortBy {
    case "title":
        sort.Slice(songs, func(i, j int) bool {
            if filters.SortOrder == "desc" {
                return songs[i].Title > songs[j].Title
            }
            return songs[i].Title < songs[j].Title
        })
    case "artist":
        sort.Slice(songs, func(i, j int) bool {
            if filters.SortOrder == "desc" {
                return songs[i].Artist > songs[j].Artist
            }
            return songs[i].Artist < songs[j].Artist
        })
    case "date":
        sort.Slice(songs, func(i, j int) bool {
            if filters.SortOrder == "desc" {
                return songs[i].ReleaseDate.After(songs[j].ReleaseDate)
            }
            return songs[i].ReleaseDate.Before(songs[j].ReleaseDate)
        })
    }

    start := (page - 1) * 10
    end := start + 10
    if end > len(songs) {
        end = len(songs)
    }
    if start > len(songs) {
        return []Song{}, len(songs), nil
    }

    return songs[start:end], len(songs), nil
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
