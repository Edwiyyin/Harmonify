package main

import (
	"log"
	"net/http"
	"os"
	"html/template"
	"time"

	"harmonify/src/handlers"
	"harmonify/src/api"
)


func init() {
	if err := api.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	funcMap := template.FuncMap{
        "minus":           handlers.Minus,
        "plus":            handlers.Plus,
        "urlencodeTitle":  handlers.UrlencodeTitle,
        "durationMinutes": handlers.DurationMinutes,
        "durationSeconds": handlers.DurationSeconds,
	}

	handlers.HomeTemplate = template.Must(template.New("home.html").Funcs(funcMap).ParseFiles("templates/home.html"))
	handlers.SearchResultsTemplate = template.Must(template.New("search.html").Funcs(funcMap).ParseFiles("templates/search.html"))
	handlers.LyricsTemplate = template.Must(template.New("lyrics.html").Funcs(funcMap).ParseFiles("templates/lyrics.html"))
	handlers.FavoritesTemplate = template.Must(template.New("favorites.html").Funcs(funcMap).ParseFiles("templates/favorites.html"))
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", handlers.HandleHome)
	http.HandleFunc("/search", handlers.HandleSearch)
	http.HandleFunc("/lyrics", handlers.HandleLyrics)
	http.HandleFunc("/favorites", handlers.HandleFavorites)
	http.HandleFunc("/add-favorite", handlers.HandleAddFavorite)
	http.HandleFunc("/remove-favorite", handlers.HandleRemoveFavorite)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
