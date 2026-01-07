package api

import (
	"encoding/json"
	"net/http"

	"groupie-tracker/models"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

func FetchArtists() ([]models.Artist, error) {
	resp, err := http.Get(baseURL + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var artists []models.Artist
	err = json.NewDecoder(resp.Body).Decode(&artists)
	if err != nil {
		return nil, err
	}

	return artists, nil
}
