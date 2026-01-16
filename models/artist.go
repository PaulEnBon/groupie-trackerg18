package models

type Artist struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	SpotifyLink  string   `json:"spotifyLink,omitempty"`
	YoutubeLink  string   `json:"youtubeLink,omitempty"`
	DeezerLink   string   `json:"deezerLink,omitempty"`
}
