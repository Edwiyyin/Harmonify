package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"harmonify/src/api"
    "harmonify/src/calc"
)

var (
	HomeTemplate          *template.Template
	SearchResultsTemplate *template.Template
	LyricsTemplate        *template.Template
	FavoritesTemplate     *template.Template
	Favorites             []api.Song
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := HomeTemplate.Execute(w, nil); err != nil {
		log.Printf("Error rendering home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func HandleLyrics(w http.ResponseWriter, r *http.Request) {
    songTitle := r.URL.Query().Get("title")
    artist := r.URL.Query().Get("artist")
    songID := r.URL.Query().Get("id")

    lyrics, err := api.FetchLyricsOvh(songTitle, artist)
    if err != nil {
        log.Printf("Lyrics fetch error: %v", err)
        lyrics = "Lyrics not available for this song"
    }

    previewURL, _ := api.SearchSpotifyMusicSource(songTitle, artist)
    spotifyURL := fmt.Sprintf("https://open.spotify.com/track/%s", songID)

    data := struct {
        Title      string
        Artist     string
        Lyrics     string
        PreviewURL string
        SpotifyURL string
    }{
        Title:      songTitle,
        Artist:     artist,
        Lyrics:     lyrics,
        PreviewURL: previewURL,
        SpotifyURL: spotifyURL,
    }

    if err := LyricsTemplate.Execute(w, data); err != nil {
        log.Printf("Error rendering lyrics template: %v", err)
        http.Error(w, "Error rendering lyrics", http.StatusInternalServerError)
        return
    }
}

func HandleFavorites(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Favorites []api.Song
	}{
		Favorites: Favorites,
	}

	if err := FavoritesTemplate.Execute(w, data); err != nil {
		log.Printf("Error rendering favorites template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")
    page := r.URL.Query().Get("page")
    pageNum, _ := strconv.Atoi(page)
    if pageNum == 0 {
        pageNum = 1
    }

    filters := api.SearchFilters{
        StartDate:   r.URL.Query().Get("startDate"),
        EndDate:     r.URL.Query().Get("endDate"),
        SortBy:      r.URL.Query().Get("sortBy"),
        SortOrder:   r.URL.Query().Get("sortOrder"),
        MinDuration: calc.ParseDuration(r.URL.Query().Get("minDuration")),
        MaxDuration: calc.ParseDuration(r.URL.Query().Get("maxDuration")),
    }

    songs, totalResults, err := api.SearchSpotifySongs(query, pageNum, filters)
    if err != nil {
        log.Printf("Search error: %v", err)
        http.Error(w, "Error searching songs", http.StatusInternalServerError)
        return
    }

    totalPages := (totalResults + 9) / 10 

    data := struct {
        Songs        []api.Song
        Query        string
        CurrentPage  int
        TotalPages   int
        TotalResults int
        Filters      api.SearchFilters
        DurationMinutes func(int) int
        DurationSeconds func(int) int
    }{
        Songs:        songs,
        Query:        query,
        CurrentPage:  pageNum,
        TotalPages:   totalPages,
        TotalResults: totalResults,
        Filters:      filters,
        DurationMinutes: calc.DurationMinutes,
        DurationSeconds: calc.DurationSeconds,
    }

    if err := SearchResultsTemplate.Execute(w, data); err != nil {
        log.Printf("Template execution error: %v", err)
        http.Error(w, "Error rendering results", http.StatusInternalServerError)
        return
    }
}

func HandleAddFavorite(w http.ResponseWriter, r *http.Request) {
    var song api.Song
    if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    for _, existingSong := range Favorites {
        if strings.EqualFold(existingSong.ID, song.ID) {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]interface{}{
                "success": false,
                "message": "Song already in favorites",
            })
            return
        }
    }

    accessToken, err := api.GetSpotifyAccessToken()
    if err != nil {
        http.Error(w, "Spotify token error", http.StatusInternalServerError)
        return
    }

    req, err := http.NewRequest("GET", 
        fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", song.ID), nil)
    if err != nil {
        http.Error(w, "Failed to create request", http.StatusInternalServerError)
        return
    }

    req.Header.Add("Authorization", "Bearer "+accessToken)
    req.Header.Add("Content-Type", "application/json")

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        http.Error(w, "Failed to fetch track details", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    var trackDetails struct {
        Name   string `json:"name"`
        Duration int    `json:"duration_ms"`
        Artists []struct {
            Name string `json:"name"`
        } `json:"artists"`
        Album struct {
            Images []struct {
                URL string `json:"url"`
            } `json:"images"`
            ReleaseDate string `json:"release_date"`
        } `json:"album"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&trackDetails); err != nil {
        http.Error(w, "Failed to parse track details", http.StatusInternalServerError)
        return
    }

    fullSong := api.Song{
        ID:          song.ID,
        Title:       trackDetails.Name,
        Artist:      trackDetails.Artists[0].Name,
        CoverURL:    "",
        ReleaseDate: time.Time{},
        Duration:    trackDetails.Duration,
    }

    if len(trackDetails.Album.Images) > 0 {
        fullSong.CoverURL = trackDetails.Album.Images[0].URL
    }

    if trackDetails.Album.ReleaseDate != "" {
        fullSong.ReleaseDate = api.FormatReleaseDate(trackDetails.Album.ReleaseDate)
    }

    Favorites = append(Favorites, fullSong)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Song added to favorites",
    })
}

func HandleGetFavorites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    favorites := []api.Song{}
	json.NewEncoder(w).Encode(favorites)
}

func HandleRemoveFavorite(w http.ResponseWriter, r *http.Request) {
    var songToRemove struct {
        ID string `json:"id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&songToRemove); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    for i, song := range Favorites {
        if song.ID == songToRemove.ID {
            Favorites = append(Favorites[:i], Favorites[i+1:]...)
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]bool{"success": true})
            return
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"success": false})
}

