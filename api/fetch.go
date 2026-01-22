package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"groupie-tracker/models"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

var cacheRelation = make(map[int]*models.Relation)

func FetchArtists() ([]models.Artist, error) {
	resp, err := http.Get(baseURL + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API: status %d", resp.StatusCode)
	}

	var artists []models.Artist
	err = json.NewDecoder(resp.Body).Decode(&artists)
	if err != nil {
		return nil, err
	}

	return artists, nil
}

func FetchRelation(id int) (*models.Relation, error) {
    if donnee, existe := cacheRelation[id]; existe {
        return donnee, nil
    }

    resp, err := http.Get(baseURL + "/relation/" + strconv.Itoa(id))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var relation models.Relation
    err = json.NewDecoder(resp.Body).Decode(&relation)
    if err != nil {
        return nil, err
    }

    cacheRelation[id] = &relation
    return &relation, nil
}

func FetchLocations(id int) (*models.Location, error) {
	resp, err := http.Get(baseURL + "/locations/" + strconv.Itoa(id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var location models.Location
	err = json.NewDecoder(resp.Body).Decode(&location)
	if err != nil {
		return nil, err
	}

	return &location, nil
}

// --- NOUVEAU : Récupère toutes les localisations pour le filtrage ---
func FetchAllLocationsMap() (map[int][]string, error) {
	resp, err := http.Get(baseURL + "/locations")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Index []struct {
			ID        int      `json:"id"`
			Locations []string `json:"locations"`
		} `json:"index"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	locMap := make(map[int][]string)
	for _, item := range result.Index {
		locMap[item.ID] = item.Locations
	}
	return locMap, nil
}
