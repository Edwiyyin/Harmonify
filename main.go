package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Song struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Lyrics   string `json:"lyrics,omitempty"`
	CoverURL string `json:"cover_url,omitempty"`
}

type GeniusResponse struct {
	Response struct {
		Hits []struct {
			Result struct {
				ID             int    `json:"id"`
				Title          string `json:"title"`
				PrimaryArtist  struct {
					Name string `json:"name"`
				} `json:"primary_artist"`
				URL         string `json:"url"`
				Song_art_image_url string `json:"song_art_image_url"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}


type Config struct {
	GeniusClientID     string `json:"genius_client_id"`
	GeniusClientSecret string `json:"genius_client_secret"`
	GeniusAccessToken  string `json:"genius_access_token"`
}

var (
	homeTemplate        *template.Template
	searchResultsTemplate *template.Template
	lyricsTemplate      *template.Template
	favoritesTemplate   *template.Template
	favorites           []Song
	config              Config
)

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

func fetchLyrics(title, artist string) (string, error) {
	title = strings.TrimSpace(strings.ToLower(title))
	artist = strings.TrimSpace(strings.ToLower(artist))

	apis := []func(string, string) (string, error){
		fetchLyricsOvh,
		fetchGeniusLyrics,
	}

	for _, apiFunc := range apis {
		lyrics, err := apiFunc(title, artist)
		if err == nil && lyrics != "" {
			return lyrics, nil
		}
	}

	return "Lyrics not found", nil
}

func fetchLyricsOvh(title, artist string) (string, error) {
	encodedTitle := url.QueryEscape(title)
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

func fetchGeniusLyrics(title, artist string) (string, error) {
	apiURL := fmt.Sprintf("https://api.genius.com/search?q=%s%%20%s", 
		url.QueryEscape(title), 
		url.QueryEscape(artist))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+config.GeniusAccessToken)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var geniusResp struct {
		Response struct {
			Hits []struct {
				Result struct {
					URL string `json:"url"`
				} `json:"result"`
			} `json:"hits"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geniusResp); err != nil {
		return "", err
	}

	if len(geniusResp.Response.Hits) == 0 {
		return "", fmt.Errorf("no song found")
	}

	songURL := geniusResp.Response.Hits[0].Result.URL
	
	scrapeResp, err := http.Get(songURL)
	if err != nil {
		return "", err
	}
	defer scrapeResp.Body.Close()

	body, err := io.ReadAll(scrapeResp.Body)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`<div class="lyrics">(.*?)</div>`)
	matches := re.FindStringSubmatch(string(body))
	
	if len(matches) > 1 {
		lyrics := html.UnescapeString(matches[1])
		lyrics = strings.TrimSpace(regexp.MustCompile(`<[^>]*>`).ReplaceAllString(lyrics, ""))
		
		if len(lyrics) > 5000 {
			lyrics = lyrics[:5000] + "... (lyrics truncated)"
		}
		
		return lyrics, nil
	}

	return "", fmt.Errorf("no lyrics found")
}

func searchSongURL(title, artist string) (string, error) {
	encodedTitle := url.QueryEscape(title)
	encodedArtist := url.QueryEscape(artist)
	query := fmt.Sprintf("%s %s", encodedTitle, encodedArtist)

	apiURL := fmt.Sprintf("https://api.genius.com/search?q=%s", query)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+config.GeniusAccessToken)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var geniusResp GeniusResponse
	if err := json.NewDecoder(resp.Body).Decode(&geniusResp); err != nil {
		return "", err
	}

	if len(geniusResp.Response.Hits) > 0 {
		return geniusResp.Response.Hits[0].Result.URL, nil
	}

	return "", fmt.Errorf("no song URL found")
}

func calculateTotalPages(totalResults int) int {
	return (totalResults + 9) / 10
}

func searchGeniusSongs(query string, page int) ([]Song, int, error) {
	if config.GeniusClientID == "" {
		if err := loadConfig(); err != nil {
			return nil, 0, err
		}
	}
	
	if config.GeniusAccessToken == "" {
		return nil, 0, fmt.Errorf("genius API access token is missing")
	}

	encodedQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://api.genius.com/search?q=%s&page=%d", encodedQuery, page)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+config.GeniusAccessToken)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var geniusResp GeniusResponse
	if err := json.NewDecoder(resp.Body).Decode(&geniusResp); err != nil {
		return nil, 0, fmt.Errorf("failed to parse JSON: %v", err)
	}

	var songs []Song
	for _, hit := range geniusResp.Response.Hits {
		songs = append(songs, Song{
			ID:       strconv.Itoa(hit.Result.ID),
			Title:    hit.Result.Title,
			Artist:   hit.Result.PrimaryArtist.Name,
			CoverURL: hit.Result.Song_art_image_url,
		})
	}

	return songs, len(songs), nil
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
	
	pageNum, _ := strconv.Atoi(page)
	if pageNum == 0 {
		pageNum = 1
	}

	songs, totalResults, err := searchGeniusSongs(query, pageNum)
	if err != nil {
		log.Printf("Search error: %v", err)
		http.Error(w, "Error searching songs", http.StatusInternalServerError)
		return
	}

	data := struct {
		Songs         []Song
		Query         string
		CurrentPage   int
		TotalPages    int
		TotalResults  int
	}{
		Songs:         songs,
		Query:         query,
		CurrentPage:   pageNum,
		TotalPages:    calculateTotalPages(totalResults),
		TotalResults:  totalResults,
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
	
	_, err := searchSongURL(songTitle, artist)
	if err != nil {
		log.Printf("Error fetching song URL: %v", err)
	}

	lyrics, err := fetchLyrics(songTitle, artist)
	if err != nil {
		log.Printf("Lyrics fetch error: %v", err)
		lyrics = "Unable to retrieve lyrics"
	}

	data := struct {
		Title   string
		Artist  string
		Lyrics  string
		ID      string
	}{
		Title:  songTitle,
		Artist: artist,
		Lyrics: lyrics,
		ID:     songID,
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
		if existingSong.ID == song.ID {
			json.NewEncoder(w).Encode(map[string]bool{"success": false})
			return
		}
	}

	favorites = append(favorites, song)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
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
	}

	loadConfig()

	homeTemplate = template.Must(template.ParseFiles("templates/home.html"))
	searchResultsTemplate = template.Must(template.New("search_results.html").
		Funcs(funcMap).
		ParseFiles("templates/search_results.html"))
	lyricsTemplate = template.Must(template.ParseFiles("templates/lyrics.html"))
	favoritesTemplate = template.Must(template.ParseFiles("templates/favorites.html"))
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