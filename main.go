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

	handlers.HomeTemplate = template.Must(template.New("home.html").Funcs(funcMap).ParseFiles("templates/home.html"))
	handlers.SearchResultsTemplate = template.Must(template.New("search.html").Funcs(funcMap).ParseFiles("templates/search.html"))
	handlers.LyricsTemplate = template.Must(template.New("lyrics.html").Funcs(funcMap).ParseFiles("templates/lyrics.html"))
	handlers.PlaylistTemplate = template.Must(template.New("playlist.html").Funcs(funcMap).ParseFiles("templates/playlist.html"))
	handlers.PlaylistLyricsTemplate = template.Must(template.New("playlist-lyrics.html").Funcs(funcMap).ParseFiles("templates/playlist-lyrics.html"))
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
	
	
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
