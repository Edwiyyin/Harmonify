package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"harmonify/src/api"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserDB struct {
	Users []User `json:"users"`
}

var (
	UsersFile = "data/users.json"
	userDB    UserDB
)

func InitUserSystem() error {

	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	if err := os.MkdirAll("data/playlists", 0755); err != nil {
		return err
	}

	if _, err := os.Stat(UsersFile); os.IsNotExist(err) {

		userDB = UserDB{Users: []User{}}
		return SaveUserDB()
	} else {

		data, err := ioutil.ReadFile(UsersFile)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, &userDB)
	}
}

func SaveUserDB() error {
	data, err := json.MarshalIndent(userDB, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(UsersFile, data, 0644)
}

func RegisterUser(username, password string) error {
	for _, user := range userDB.Users {
		if user.Username == username {
			return errors.New("username already exists")
		}
	}

	userDB.Users = append(userDB.Users, User{
		Username: username,
		Password: password,
	})

	if err := SaveUserDB(); err != nil {
		return err
	}

	return SaveUserPlaylist(username, []api.Song{})
}

func AuthenticateUser(username, password string) bool {
	for _, user := range userDB.Users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

func GetUserPlaylistPath(username string) string {
	return filepath.Join("data/playlists", fmt.Sprintf("%s_playlist.json", username))
}

func LoadUserPlaylist(username string) ([]api.Song, error) {
	playlistPath := GetUserPlaylistPath(username)

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

func SaveUserPlaylist(username string, playlist []api.Song) error {
	playlistPath := GetUserPlaylistPath(username)
	
	data, err := json.MarshalIndent(playlist, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(playlistPath, data, 0644)
}