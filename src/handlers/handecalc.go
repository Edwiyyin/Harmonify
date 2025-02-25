package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func init() {
	loadPlaylistFromFile()
}

func loadPlaylistFromFile() {
	data, err := ioutil.ReadFile(PlaylistFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error reading playlist file: %v", err)
		}
		return
	}

	if err := json.Unmarshal(data, &Playlist); err != nil {
		log.Printf("Error parsing playlist file: %v", err)
	}
}

func savePlaylistToFile() error {
    data, err := json.MarshalIndent(Playlist, "", "  ")
    if err != nil {
        log.Printf("Error marshaling playlist: %v", err)
        return fmt.Errorf("failed to encode playlist: %v", err)
    }

    err = ioutil.WriteFile(PlaylistFile, data, 0644)
    if err != nil {
        log.Printf("Error writing playlist file: %v", err)
        return fmt.Errorf("failed to write playlist file: %v", err)
    }

    log.Println("Playlist updated successfully")
    return nil
}