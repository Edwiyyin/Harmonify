package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

    "harmonify/src/api"
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }

    playlist, err := LoadPlaylistFromFile()
    if err != nil {
        log.Printf("Error loading playlist: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    Playlist = playlist
    _, _, loggedIn := getSessionInfo(r)

    data := struct {
        LoggedIn bool
    }{
        LoggedIn: loggedIn,
    }

    if err := HomeTemplate.Execute(w, data); err != nil {
        log.Printf("Error rendering home template: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func HandleError(w http.ResponseWriter, r *http.Request) {
    
    query := r.URL.Query().Get("query")
    requestedPageStr := r.URL.Query().Get("page")
    totalPagesStr := r.URL.Query().Get("totalPages")

    requestedPage, err := strconv.Atoi(requestedPageStr)
    if err != nil || requestedPage < 1 {
        requestedPage = 1
    }

    totalPages, err := strconv.Atoi(totalPagesStr)
    if err != nil || totalPages < 1 {
        totalPages = 1
    }

    data := struct {
        RequestedPage int
        TotalPages    int
        Query         string
    }{
        RequestedPage: requestedPage,
        TotalPages:    totalPages,
        Query:         query,
    }

    if err := ErrorTemplate.Execute(w, data); err != nil {
        log.Printf("Failed to render error page: %v", err)
        http.Error(w, "Failed to render error page", http.StatusInternalServerError)
    }
}

func HandleFAQ(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/faq" {
        http.NotFound(w, r)
        return
    }

    if err := FAQTemplate.Execute(w, nil); err != nil {
        log.Printf("Error rendering FAQ template: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")
    page := r.URL.Query().Get("page")
    pageNum, err := strconv.Atoi(page)
    if err != nil || pageNum < 1 {
        pageNum = 1
    }

    filters := api.SearchFilters{
        StartDate:   r.URL.Query().Get("startDate"),
        EndDate:     r.URL.Query().Get("endDate"),
        SortBy:      r.URL.Query().Get("sortBy"),
        SortOrder:   r.URL.Query().Get("sortOrder"),
        MinDuration: api.ParseDuration(r.URL.Query().Get("minDuration")),
        MaxDuration: api.ParseDuration(r.URL.Query().Get("maxDuration")),
        LyricsFilter: r.URL.Query().Get("lyricsFilter"),
        PlaylistFilter: r.URL.Query().Get("playlistFilter"),
    }

    songs, totalResults, err := api.SearchSpotifySongs(query, pageNum, filters)
    if err != nil {
        log.Printf("Search error: %v", err)
        http.Error(w, "Error searching songs", http.StatusInternalServerError)
        return
    }

    if filters.PlaylistFilter == "in_playlist" || filters.PlaylistFilter == "not_in_playlist" {
        var filteredSongs []api.Song
        for _, song := range songs {
            inPlaylist := false
            for _, playlistSong := range Playlist {
                if playlistSong.ID == song.ID {
                    inPlaylist = true
                    break
                }
            }
            
            if (filters.PlaylistFilter == "in_playlist" && inPlaylist) || 
               (filters.PlaylistFilter == "not_in_playlist" && !inPlaylist) {
                filteredSongs = append(filteredSongs, song)
            }
        }
        songs = filteredSongs
        totalResults = len(filteredSongs)

    }

    resultsPerPage := 15
    totalPages := totalResults / resultsPerPage
    if totalResults%resultsPerPage > 0 {
        totalPages++
    }

    if totalPages < 1 {
        totalPages = 1
    }

    if pageNum > totalPages {
        redirectURL := fmt.Sprintf("/error?query=%s&page=%d", query, pageNum)
        log.Printf("Redirecting to error page: %s", redirectURL)
        http.Redirect(w, r, redirectURL, http.StatusFound)
        return
    }

    data := struct {
        Songs        []api.Song
        Query        string
        CurrentPage  int
        TotalPages   int
        TotalResults int
        ResultsPerPage int
        Filters      api.SearchFilters
        DurationMinutes func(int) int
        DurationSeconds func(int) int
    }{
        Songs:        songs,
        Query:        query,
        CurrentPage:  pageNum,
        TotalPages:   totalPages,
        TotalResults: totalResults,
        ResultsPerPage: resultsPerPage,
        Filters:      filters,
        DurationMinutes: api.DurationMinutes,
        DurationSeconds: api.DurationSeconds,
    }

    if err := SearchResultsTemplate.Execute(w, data); err != nil {
        log.Printf("Template execution error: %v", err)
        http.Error(w, "Error rendering results", http.StatusInternalServerError)
        return
    }
}