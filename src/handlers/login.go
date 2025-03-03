package handlers

import (
	"log"
	"net/http"
	"time"

	"harmonify/src/api"
	"harmonify/src/auth"
)




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
	
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		if err := RegisterTemplate.Execute(w, nil); err != nil {
			log.Printf("Error rendering register template: %v", err)
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
		confirmPassword := r.FormValue("confirm_password")

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
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
    sessionID, username, loggedIn := getSessionInfo(r)
    
    if loggedIn {

        if err := auth.SaveUserPlaylist(username, Playlist); err != nil {
            log.Printf("Error saving user playlist: %v", err)
        }
        
        delete(activeSessions, sessionID)
    
        cookie := http.Cookie{
            Name:     "session_id",
            Value:    "",
            Path:     "/",
            MaxAge:   -1,
            HttpOnly: true,
        }
        http.SetCookie(w, &cookie)
    }
    
    Playlist = []api.Song{}
    
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		_, _, loggedIn := getSessionInfo(r)
		
		if !loggedIn {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}