package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"harmonify/src/api"
	"harmonify/src/auth"
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

    inPlaylist := false
    for _, song := range Playlist {
        if song.ID == songID {
            inPlaylist = true
            break
        }
    }

    data := struct {
        ID            string
        Title         string
        Artist        string
        Lyrics        string
        PreviewURL    string
        SpotifyURL    string
        InPlaylist    bool
        ActionMessage string
        Query         string
        Page          int
    }{
        ID:            songID,
        Title:         songTitle,
        Artist:        artist,
        Lyrics:        lyrics,
        PreviewURL:    previewURL,
        SpotifyURL:    spotifyURL,
        InPlaylist:    inPlaylist,
        ActionMessage: actionMessage,
        Query:         query,
        Page:          pageNum,
    }

    if err := LyricsTemplate.Execute(w, data); err != nil {
        log.Printf("Error rendering lyrics template: %v", err)
        http.Error(w, "Error rendering lyrics", http.StatusInternalServerError)
        return
    }
}

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

    inPlaylist := false
    for _, song := range Playlist {
        if song.ID == songID {
            inPlaylist = true;
            break;
        }
    }

    data := struct {
        ID            string
        Title         string
        Artist        string
        Lyrics        string
        SpotifyURL    string
        InPlaylist    bool
        Query         string
        Page          int
    }{
        ID:            songID,
        Title:         songTitle,
        Artist:        artist,
        Lyrics:        lyrics,
        SpotifyURL:    spotifyURL,
        InPlaylist:    inPlaylist,
        Query:         query,
        Page:          pageNum,
    }

    if err := PlaylistLyricsTemplate.Execute(w, data); err != nil {
        log.Printf("Error rendering playlist-lyrics template: %v", err)
        http.Error(w, "Error rendering playlist-lyrics", http.StatusInternalServerError)
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


func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		if err := LoginTemplate.Execute(w, nil); err != nil {
			log.Printf("Error rendering login template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	
	if r.Method == http.MethodPost {

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		
		username := r.FormValue("username")
		password := r.FormValue("password")

		if auth.AuthenticateUser(username, password) {

			sessionID := generateSessionID()

			activeSessions[sessionID] = Session{
				Username:  username,
				LoggedIn:  true,
				CreatedAt: time.Now(),
			}
			cookie := http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Path:     "/",
				MaxAge:   3600 * 24,
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)
		
			playlist, err := auth.LoadUserPlaylist(username)
			if err != nil {
				log.Printf("Error loading user playlist: %v", err)
			}

			Playlist = playlist

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		data := struct {
			Error string
		}{
			Error: "Invalid username or password",
		}
		
		if err := LoginTemplate.Execute(w, data); err != nil {
			log.Printf("Error rendering login template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	
	// Method not allowed
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleRegister processes registration requests
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Display registration form
		if err := RegisterTemplate.Execute(w, nil); err != nil {
			log.Printf("Error rendering register template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	
	if r.Method == http.MethodPost {
		// Process registration form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")
		
		// Validate input
		if username == "" || password == "" {
			data := struct {
				Error string
			}{
				Error: "Username and password are required",
			}
			
			if err := RegisterTemplate.Execute(w, data); err != nil {
				log.Printf("Error rendering register template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		
		if password != confirmPassword {
			data := struct {
				Error string
			}{
				Error: "Passwords do not match",
			}
			
			if err := RegisterTemplate.Execute(w, data); err != nil {
				log.Printf("Error rendering register template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		
		// Register user
		if err := auth.RegisterUser(username, password); err != nil {
			data := struct {
				Error string
			}{
				Error: err.Error(),
			}
			
			if err := RegisterTemplate.Execute(w, data); err != nil {
				log.Printf("Error rendering register template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		
		// Redirect to login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	// Method not allowed
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
    // Get current user and session
    sessionID, username, loggedIn := getSessionInfo(r)
    
    if loggedIn {
        // Save user's playlist
        if err := auth.SaveUserPlaylist(username, Playlist); err != nil {
            log.Printf("Error saving user playlist: %v", err)
        }
        
        // Clear session
        delete(activeSessions, sessionID)
        
        // Clear session cookie
        cookie := http.Cookie{
            Name:     "session_id",
            Value:    "",
            Path:     "/",
            MaxAge:   -1,
            HttpOnly: true,
        }
        http.SetCookie(w, &cookie)
    }
    
    // Reset global playlist
    Playlist = []api.Song{}
    
    // Redirect to home page
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

// generateSessionID creates a unique session ID


// AuthMiddleware checks if a user is logged in
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session info
		_, _, loggedIn := getSessionInfo(r)
		
		if !loggedIn {
			// Redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}