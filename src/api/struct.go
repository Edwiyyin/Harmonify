package api

import (
	"time"
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
    LyricsFilter string `json:"lyricsFilter"`
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
    InPlaylist  bool      `json:"in_playlist"`
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

const (
    MaxTotalResults = 1000
    ResultsPerPage  = 15
)