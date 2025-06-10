package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"harmonify/src/api"
)

func HandleLyrics(w http.ResponseWriter, r *http.Request) {
    songTitle, _ := url.QueryUnescape(r.URL.Query().Get("title"))
    artist := r.URL.Query().Get("artist")
    songID := r.URL.Query().Get("id")
    query := r.URL.Query().Get("query")
    page := r.URL.Query().Get("page")
    pageNum, _ := strconv.Atoi(page)
    actionMessage := r.URL.Query().Get("action")

    if pageNum == 0 {
        pageNum = 1
    }

    lyrics, err := api.FetchLyricsOvh(songTitle, artist)
    if err != nil {
        log.Printf("Lyrics fetch error: %v", err)
        lyrics = "Lyrics not available for this song"
    }

    previewURL, _ := api.SearchSpotifyMusicSource(songTitle, artist)
    spotifyURL := fmt.Sprintf("https://open.spotify.com/track/%s", songID)

    spotifyTrack, err := api.FetchSpotifyTrack(songID)
    if err != nil {
        log.Printf("Error fetching Spotify track details: %v", err)
    }

    var coverURL, releaseDate, duration string
if spotifyTrack != nil {
    if len(spotifyTrack.Album.Images) > 0 {
        coverURL = spotifyTrack.Album.Images[0].URL
    }
    releaseDate = spotifyTrack.Album.ReleaseDate
    duration = fmt.Sprintf("%d:%02d", spotifyTrack.DurationMs/60000, (spotifyTrack.DurationMs%60000)/1000)
}

    inPlaylist := false
    for _, song := range Playlist {
        if song.ID == songID {
            inPlaylist = true
            break
        }
    }

    data := struct {
        ID                   string
        Title                string
        Artist               string
        Lyrics               string
        PreviewURL           string
        SpotifyURL           string
        InPlaylist           bool
        ActionMessage        string
        Query                string
        Page                 int
        CoverURL             string
        FormattedReleaseDate string
        FormattedDuration    string
    }{
        ID:                   songID,
        Title:                songTitle,
        Artist:               artist,
        Lyrics:               lyrics,
        PreviewURL:           previewURL,
        SpotifyURL:           spotifyURL,
        InPlaylist:           inPlaylist,
        ActionMessage:        actionMessage,
        Query:                query,
        Page:                 pageNum,
        CoverURL:             coverURL,
        FormattedReleaseDate: releaseDate,
        FormattedDuration:    duration,
    }

    if err := LyricsTemplate.Execute(w, data); err != nil {
        log.Printf("Error rendering lyrics template: %v", err)
        http.Error(w, "Error rendering lyrics", http.StatusInternalServerError)
        return
    }
}

func HandleGetLyricsText(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    lyrics := r.URL.Query().Get("lyrics")
    if lyrics == "" {
        http.Error(w, "No lyrics provided", http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(lyrics))
}

func HandlePlaylistLyrics(w http.ResponseWriter, r *http.Request) {
    songTitle, _ := url.QueryUnescape(r.URL.Query().Get("title"))
    artist := r.URL.Query().Get("artist")
    songID := r.URL.Query().Get("id")
    query := r.URL.Query().Get("query")
    page := r.URL.Query().Get("page")
    pageNum, _ := strconv.Atoi(page)
    if pageNum == 0 {
        pageNum = 1
    }

    lyrics, err := api.FetchLyricsOvh(songTitle, artist)
    if err != nil {
        log.Printf("Lyrics fetch error: %v", err)
        lyrics = "Lyrics not available for this song"
    }

    spotifyURL := fmt.Sprintf("https://open.spotify.com/track/%s", songID)

    spotifyTrack, err := api.FetchSpotifyTrack(songID)
    if err != nil {
        log.Printf("Error fetching Spotify track details: %v", err)
    }

    var coverURL, releaseDate, duration string
    if spotifyTrack != nil {
        coverURL = spotifyTrack.Album.Images[0].URL
        releaseDate = spotifyTrack.Album.ReleaseDate
        duration = fmt.Sprintf("%d:%02d", spotifyTrack.DurationMs/60000, (spotifyTrack.DurationMs%60000)/1000)
    }

    inPlaylist := false
    for _, song := range Playlist {
        if song.ID == songID {
            inPlaylist = true
            break
        }
    }

    data := struct {
        ID                   string
        Title                string
        Artist               string
        Lyrics               string
        SpotifyURL           string
        InPlaylist           bool
        Query                string
        Page                 int
        CoverURL             string
        FormattedReleaseDate string
        FormattedDuration    string
    }{
        ID:                   songID,
        Title:                songTitle,
        Artist:               artist,
        Lyrics:               lyrics,
        SpotifyURL:           spotifyURL,
        InPlaylist:           inPlaylist,
        Query:                query,
        Page:                 pageNum,
        CoverURL:             coverURL,
        FormattedReleaseDate: releaseDate,
        FormattedDuration:    duration,
    }

    if err := PlaylistLyricsTemplate.Execute(w, data); err != nil {
        log.Printf("Error rendering playlist-lyrics template: %v", err)
        http.Error(w, "Error rendering playlist-lyrics", http.StatusInternalServerError)
        return
    }
}