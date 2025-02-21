package main

import (
	"log"
	"net/http"
	"os"
	"html/template"
	"time"

	"harmonify/src/handlers"
	"harmonify/src/api"
	"harmonify/src/calc"
)

func init() {
    if err := api.LoadConfig(); err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    funcMap := template.FuncMap{
        "minus":           calc.Minus,
        "plus":            calc.Plus,
        "urlencodeTitle":  calc.UrlencodeTitle,
        "durationMinutes": calc.DurationMinutes,
        "durationSeconds": calc.DurationSeconds,
    }

    var err error
    handlers.HomeTemplate, err = template.New("home.html").Funcs(funcMap).ParseFiles("templates/home.html")
    if err != nil {
        log.Fatalf("Failed to parse home template: %v", err)
    }

    handlers.SearchResultsTemplate, err = template.New("search.html").Funcs(funcMap).ParseFiles("templates/search.html")
    if err != nil {
        log.Fatalf("Failed to parse search template: %v", err)
    }

    handlers.LyricsTemplate, err = template.New("lyrics.html").Funcs(funcMap).ParseFiles("templates/lyrics.html")
    if err != nil {
        log.Fatalf("Failed to parse lyrics template: %v", err)
    }

    handlers.PlaylistTemplate, err = template.New("playlist.html").Funcs(funcMap).ParseFiles("templates/playlist.html")
    if err != nil {
        log.Fatalf("Failed to parse playlist template: %v", err)
    }

    handlers.PlaylistLyricsTemplate, err = template.New("playlist-lyrics.html").Funcs(funcMap).ParseFiles("templates/playlist-lyrics.html")
    if err != nil {
        log.Fatalf("Failed to parse playlist-lyrics template: %v", err)
    }

    handlers.ErrorTemplate, err = template.New("error.html").Funcs(funcMap).ParseFiles("templates/error.html")
    if err != nil {
        log.Fatalf("Failed to parse error template: %v", err)
    }

	handlers.FAQTemplate, err = template.New("faq.html").Funcs(funcMap).ParseFiles("templates/faq.html")
	if err != nil {
		log.Fatalf("Failed to parse FAQ template: %v", err)
	}
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", handlers.HandleHome)
	http.HandleFunc("/search", handlers.HandleSearch)
	http.HandleFunc("/lyrics", handlers.HandleLyrics)
	http.HandleFunc("/playlist", handlers.HandlePlaylist)
	http.HandleFunc("/playlist-lyrics", handlers.HandlePlaylistLyrics)
	http.HandleFunc("/add-to-playlist", handlers.HandleAddToPlaylist)
	http.HandleFunc("/remove-from-playlist", handlers.HandleRemoveFromPlaylist)
	http.HandleFunc("/get-lyrics-text", handlers.HandleGetLyricsText)
	http.HandleFunc("/error", handlers.HandleError)
	http.HandleFunc("/faq", handlers.HandleFAQ)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
