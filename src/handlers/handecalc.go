package handlers

import (
	"encoding/json"
	"fmt"
	"harmonify/src/api"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	HomeTemplate          *template.Template
	SearchResultsTemplate *template.Template
	LyricsTemplate        *template.Template
	PlaylistTemplate      *template.Template
    PlaylistLyricsTemplate *template.Template
    ErrorTemplate          *template.Template
    FAQTemplate            *template.Template
    Playlist               []api.Song
    PlaylistFile           = "playlists/playlist.json"
)

var (
	LoginTemplate  *template.Template
	RegisterTemplate *template.Template
)

type Session struct {
	Username  string
	LoggedIn  bool
	CreatedAt time.Time
}

var activeSessions = make(map[string]Session)

func init() {
	LoadPlaylistFromFile()
}

func LoadPlaylistFromFile() ([]api.Song, error) {

    _, username, loggedIn := getSessionInfo(&http.Request{})
    playlistFile := "default.json"

    if loggedIn {
        playlistFile = fmt.Sprintf("%s_playlist.json", username)
    }
    playlistPath := filepath.Join("data", "playlists", playlistFile)

    if _, err := os.Stat(playlistPath); os.IsNotExist(err) {
        return []api.Song{}, nil
    }
    data, err := ioutil.ReadFile(playlistPath)

    if err != nil {
        return nil, err
    }
    var playlist []api.Song

    if err := json.Unmarshal(data, &playlist); err != nil {
        return nil, err
    }
    return playlist, nil
}

func SavePlaylistToFile() error {
    _, username, loggedIn := getSessionInfo(&http.Request{})

    playlistFile := "default.json"
    if loggedIn {
        playlistFile = fmt.Sprintf("%s_playlist.json", username)
    }

    if err := os.MkdirAll("data/playlists", 0755); err != nil {
        return err
    }

    playlistPath := filepath.Join("data", "playlists", playlistFile)
    data, err := json.MarshalIndent(Playlist, "", "  ")
    if err != nil {
        return err
    }

    return ioutil.WriteFile(playlistPath, data, 0644)
}

func getSessionInfo(r *http.Request) (string, string, bool) {

	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", "", false
	}

	session, exists := activeSessions[cookie.Value]
	if !exists || !session.LoggedIn {
		return "", "", false
	}
	
	return cookie.Value, session.Username, true
}

func generateSessionID() string {

	return fmt.Sprintf("%d", time.Now().UnixNano())
}