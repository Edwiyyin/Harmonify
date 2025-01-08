package main

import (
	"encoding/base64"
	"encoding/json"
    "fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var (
	homeTemplate          *template.Template
	searchResultsTemplate *template.Template
	lyricsTemplate        *template.Template
	favoritesTemplate     *template.Template
	favorites             []Song
	config                Config
	spotifyAccessToken    string
	spotifyTokenExpiry    time.Time
)

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
    MinDuration  int    `json:"minDuration"`
    MaxDuration  int    `json:"maxDuration"`
}

func loadConfig() error {
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

func (s Song) FormattedDuration() string {
    seconds := s.Duration / 1000
    minutes := seconds / 60
    remainingSeconds := seconds % 60
    return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
}


func getSpotifyAccessToken() (string, error) {
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

func sanitizeSearchQuery(query string) string {
   
    re := regexp.MustCompile(`\s*\([^)]*\)`)
    query = re.ReplaceAllString(query, "")

    parts := strings.Split(query, "-")
    query = strings.TrimSpace(parts[0])

    return strings.TrimSpace(query)
}

func searchSpotifySongs(query string, page int, filters SearchFilters) ([]Song, int, error) {
    accessToken, err := getSpotifyAccessToken()
    if err != nil {
        return nil, 0, fmt.Errorf("spotify token error: %v", err)
    }

    sanitizedQuery := sanitizeSearchQuery(query)
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
            releaseDate = formatReleaseDate(track.Album.ReleaseDate)
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

func fetchLyricsOvh(title, artist string) (string, error) {
    sanitizedTitle := sanitizeSearchQuery(title)
    
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

func searchSpotifyMusicSource(title, artist string) (string, error) {

    sanitizedTitle := sanitizeSearchQuery(title)
    
    accessToken, err := getSpotifyAccessToken()
    if err != nil {
        return "", fmt.Errorf("spotify token error: %v", err)
    }
    firstArtist := strings.Split(artist, ",")[0]
    firstArtist = strings.TrimSpace(firstArtist)

    query := fmt.Sprintf("%s %s", sanitizedTitle, firstArtist)
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
        log.Printf("Spotify Response Status: %s, Error: %v", resp.Status, err)
        return "", fmt.Errorf("failed to decode Spotify response: %v", err)
    }

    if len(trackResp.Tracks.Items) > 0 && trackResp.Tracks.Items[0].PreviewURL != "" {
        return trackResp.Tracks.Items[0].PreviewURL, nil
    }

    return "", fmt.Errorf("no preview URL found for %s by %s", sanitizedTitle, artist)
}

func calculateTotalPages(totalResults int) int {
    if totalResults <= 0 {
        return 1
    }
    const resultsPerPage = 10
    return (totalResults + resultsPerPage - 1) / resultsPerPage
}

func formatReleaseDate(dateStr string) time.Time {
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
	return s.ReleaseDate.Format("January 2, 2006")
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if err := homeTemplate.Execute(w, nil); err != nil {
		log.Printf("Error rendering home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")
    page := r.URL.Query().Get("page")
    startDate := r.URL.Query().Get("startDate")
    endDate := r.URL.Query().Get("endDate")
    sortBy := r.URL.Query().Get("sortBy")
    sortOrder := r.URL.Query().Get("sortOrder")
    minDuration, _ := strconv.Atoi(r.URL.Query().Get("minDuration"))
    maxDuration, _ := strconv.Atoi(r.URL.Query().Get("maxDuration"))

    pageNum, _ := strconv.Atoi(page)
    if pageNum == 0 {
        pageNum = 1
    }

    filters := SearchFilters{
        StartDate: startDate,
        EndDate:   endDate,
        SortBy:    sortBy,
        SortOrder: sortOrder,
        MinDuration: minDuration,
        MaxDuration: maxDuration,
    }

    songs, totalResults, err := searchSpotifySongs(query, pageNum, filters)
    if err != nil {
        log.Printf("Search error: %v", err)
        http.Error(w, "Error searching songs", http.StatusInternalServerError)
        return
    }

    data := struct {
        Songs        []Song
        Query        string
        CurrentPage  int
        TotalPages   int
        TotalResults int
        Filters      SearchFilters
    }{
        Songs:        songs,
        Query:        query,
        CurrentPage:  pageNum,
        TotalPages:   calculateTotalPages(totalResults),
        TotalResults: totalResults,
        Filters:      filters,
    }

    if err := searchResultsTemplate.Execute(w, data); err != nil {
        log.Printf("Template execution error: %v", err)
        http.Error(w, "Error rendering results", http.StatusInternalServerError)
    }
}

func handleLyrics(w http.ResponseWriter, r *http.Request) {
	songID := r.URL.Query().Get("id")
	songTitle := r.URL.Query().Get("title")
	artist := r.URL.Query().Get("artist")

	musicURL, err := searchSpotifyMusicSource(songTitle, artist)
	if err != nil {
		log.Printf("Music source error: %v", err)
		musicURL = ""
	}

	lyrics, err := fetchLyricsOvh(songTitle, artist)
	if err != nil {
		log.Printf("Lyrics fetch error: %v", err)
		lyrics = "Unable to retrieve lyrics"
	}

	data := struct {
		Title       string
		Artist      string
		Lyrics      string
		ID          string
		MusicURL    string
		ExternalURL struct {
			Spotify string
		}
		PreviewURL string
	}{
		Title:    songTitle,
		Artist:   artist,
		Lyrics:   lyrics,
		ID:       songID,
		MusicURL: musicURL,
		ExternalURL: struct {
			Spotify string
		}{
			Spotify: "https://open.spotify.com/track/" + songID,
		},
		PreviewURL: musicURL,
	}

	if err := lyricsTemplate.Execute(w, data); err != nil {
		log.Printf("Lyrics template execution error: %v", err)
		http.Error(w, "Error rendering lyrics", http.StatusInternalServerError)
	}
}

func handleAddFavorite(w http.ResponseWriter, r *http.Request) {
    var song Song
    if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    for _, existingSong := range favorites {
        if strings.EqualFold(existingSong.ID, song.ID) {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]interface{}{
                "success": false,
                "message": "Song already in favorites",
            })
            return
        }
    }

    accessToken, err := getSpotifyAccessToken()
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

    fullSong := Song{
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
        fullSong.ReleaseDate = formatReleaseDate(trackDetails.Album.ReleaseDate)
    }

    favorites = append(favorites, fullSong)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Song added to favorites",
    })
}
func handleGetFavorites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

func handleRemoveFavorite(w http.ResponseWriter, r *http.Request) {
	var songToRemove struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&songToRemove); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	for i, song := range favorites {
		if song.ID == songToRemove.ID {
			favorites = append(favorites[:i], favorites[i+1:]...)
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": false})
}

func handleFavorites(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Favorites []Song
	}{
		Favorites: favorites,
	}

	if err := favoritesTemplate.Execute(w, data); err != nil {
		log.Printf("Error rendering favorites template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func init() {
    funcMap := template.FuncMap{
        "plus":  func(a int) int { return a + 1 },
        "minus": func(a int) int { return a - 1 },
        "urlencodeTitle": func(s string) string {
            return url.QueryEscape(s)
        },
        "urlencode": func(s string) string {
            return url.QueryEscape(s)
        },
        "div": func(a, b int) int { return a / b },
        "mod": func(a, b int) int { return a % b },
        "durationMinutes": func(seconds int) int { 
            return seconds / 60 
        },
        "durationSeconds": func(seconds int) int { 
            return seconds % 60 
        },
    }

    loadConfig()

    homeTemplate = template.Must(template.ParseFiles("templates/home.html"))

    searchResultsTemplate = template.Must(template.New("search.html").
        Funcs(funcMap).
        ParseFiles("templates/search.html"))

    lyricsTemplate = template.Must(template.New("lyrics.html").
        Funcs(funcMap).
        ParseFiles("templates/lyrics.html"))

    favoritesTemplate = template.Must(template.New("favorites.html").
        Funcs(funcMap).
        ParseFiles("templates/favorites.html"))
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/lyrics", handleLyrics)
	http.HandleFunc("/favorites", handleFavorites)
	http.HandleFunc("/add-favorite", handleAddFavorite)
	http.HandleFunc("/remove-favorite", handleRemoveFavorite)
	http.HandleFunc("/get-favorites", handleGetFavorites)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
