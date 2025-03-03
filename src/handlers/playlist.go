package handlers

import (
	"encoding/json"	
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"harmonify/src/api"
)


func HandlePlaylist(w http.ResponseWriter, r *http.Request) {
    playlist, err := LoadPlaylistFromFile()
    if err != nil {
        log.Printf("Error loading playlist: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    data := struct {
        Playlist []api.Song
    }{
        Playlist: playlist,
    }

    if err := PlaylistTemplate.Execute(w, data); err != nil {
        log.Printf("Error rendering playlist template: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func HandleAddToPlaylist(w http.ResponseWriter, r *http.Request) {
    songId := r.URL.Query().Get("id")
    title := r.URL.Query().Get("title")
    artist := r.URL.Query().Get("artist")
    query := r.URL.Query().Get("query")
    page := r.URL.Query().Get("page")

    for _, existingSong := range Playlist {
        if strings.EqualFold(existingSong.ID, songId) {
            redirectURL := fmt.Sprintf("/lyrics?title=%s&artist=%s&id=%s&query=%s&page=%s&action=already_exists",
                url.QueryEscape(title),
                url.QueryEscape(artist),
                url.QueryEscape(songId),
                url.QueryEscape(query),
                url.QueryEscape(page))
            http.Redirect(w, r, redirectURL, http.StatusSeeOther)
            return
        }
    }

    accessToken, err := api.GetSpotifyAccessToken()
    if err != nil {
        log.Printf("Failed to get Spotify access token: %v", err)
        http.Redirect(w, r, fmt.Sprintf("/search?query=%s&page=%s&action=failed", query, page), http.StatusSeeOther)
        return
    }

    req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", songId), nil)
    if err != nil {
        log.Printf("Failed to create Spotify request: %v", err)
        http.Redirect(w, r, fmt.Sprintf("/search?query=%s&page=%s&action=failed", query, page), http.StatusSeeOther)
        return
    }

    req.Header.Add("Authorization", "Bearer "+accessToken)
    req.Header.Add("Content-Type", "application/json")

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Failed to fetch song details from Spotify: %v", err)
        http.Redirect(w, r, fmt.Sprintf("/search?query=%s&page=%s&action=failed", query, page), http.StatusSeeOther)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Printf("Spotify API returned non-200 status: %v", resp.StatusCode)
        http.Redirect(w, r, fmt.Sprintf("/search?query=%s&page=%s&action=failed", query, page), http.StatusSeeOther)
        return
    }

    var trackDetails struct {
        Name     string `json:"name"`
        Duration int    `json:"duration_ms"`
        Artists  []struct {
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
        log.Printf("Failed to decode Spotify response: %v", err)
        http.Redirect(w, r, fmt.Sprintf("/search?query=%s&page=%s&action=failed", query, page), http.StatusSeeOther)
        return
    }

    fullSong := api.Song{
        ID:          songId,
        Title:       title,
        Artist:      artist,
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
    Playlist = append(Playlist, fullSong)
    if err := SavePlaylistToFile(); err != nil {
        log.Printf("Failed to save playlist: %v", err)
        http.Redirect(w, r, fmt.Sprintf("/search?query=%s&page=%s&action=failed", query, page), http.StatusSeeOther)
        return
    }
    http.Redirect(w, r, fmt.Sprintf("/search?query=%s&page=%s&action=added", query, page), http.StatusSeeOther)
}

func HandleRemoveFromPlaylist(w http.ResponseWriter, r *http.Request) {
    songId := r.URL.Query().Get("id")

    for i, song := range Playlist {
        if song.ID == songId {
            Playlist = append(Playlist[:i], Playlist[i+1:]...)
            if err := SavePlaylistToFile(); err != nil {
                log.Printf("Failed to save playlist: %v", err)
                http.Redirect(w, r, "/playlist?action=failed", http.StatusSeeOther)
                return
            }

            http.Redirect(w, r, "/playlist?action=removed", http.StatusSeeOther)
            return
        }
    }

    http.Redirect(w, r, "/playlist?action=not_found", http.StatusSeeOther)
}