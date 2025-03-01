# Harmonify 🎵
![alt text](static/img/2.png)
## Overview

Harmonify is a web application that allows users to search for song lyrics, discover music, and manage their favorite songs. The application integrates multiple music APIs to provide a rich music exploration experience.

## Features

- Song Lyrics Search
- Music Preview Integration
- Favorites Management
- Spotify API Support
- Responsive Web Design

## Tech Stack

- Backend: Go (Golang)
- Frontend: HTML, CSS, JavaScript
- APIs:
  - Spotify API (Music Previews)
  - Lyrics.ovh API (Lyrics Retrieval)

## Prerequisites

- Go 1.16+
- Spotify Developer Account

## Setup

1. Clone the repository
2. Create `data/config.json` with API credentials:
```json
{
    "spotify_client_id": "YOUR_SPOTIFY_CLIENT_ID",
    "spotify_client_secret": "YOUR_SPOTIFY_CLIENT_SECRET"
}
```

3. Install dependencies
4. Run the application:
```bash
go run main.go
```

## Endpoints

- `/`: Home page with search functionality
- `/search`: Display search results
- `/lyrics`: Show song lyrics and additional details
- `/playlist`: Manage playlist songs
- `/playlist-lyrics`: Same as /lyrics but for /playlist
- `/register`: To create an account
- `/login`: To log into your account
- `/error`: Indicate error in /search
- `/faq`: For FAQ

## Structure

```
Harmonify/
    ├── README.md
    ├── go.mod
    ├── main.go
    ├── data/
    │   ├──playlists/
    │   │    ├──<username>_playlist.json
    │   │    └──default.json
    │   │   
    │   ├──configure.json
    │   └──users.json
    ├── src/
    │   ├── api/
    │   │   ├── api.go
    │   │   ├── apicalc.go
    │   │   └── struct.go
    │   ├── calc/
    │   │   └── calc.go
    │   └── handlers/
    │       ├── handecalc.go
    │       └── handle.go
    ├── static/
    │   ├── css/
    │   │   ├── faq.css
    │   │   ├── form.css
    │   │   ├── home.css
    │   │   ├── lyrics.css
    │   │   ├── playlist.css
    │   │   └── search.css
    │   ├── img/
    │   └── js/
    │       ├── home.js
    │       ├── lyrics.js
    │       ├── playlist.js
    │       └── search.js
    └── templates/
        ├── error.html
        ├── faq.html
        ├── home.html
        ├──login.html
        ├── lyrics.html
        ├── playlist-lyrics.html
        ├── playlist.html
        ├── register.html
        └── search.html


```

## Configuration

Modify `config.json` to update API credentials and settings.

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit changes
4. Push to the branch
5. Create a Pull Request
