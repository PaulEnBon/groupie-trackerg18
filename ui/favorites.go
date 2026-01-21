package ui

import (
	"encoding/json"
	"os"
	"sync"
)

const favFileName = "favorites.json"

// Gestionnaire de favoris avec Mutex pour éviter les conflits
var favLock sync.Mutex

func LoadFavorites() map[int]bool {
	favLock.Lock()
	defer favLock.Unlock()

	favs := make(map[int]bool)

	file, err := os.Open(favFileName)
	if err != nil {
		return favs
	}
	defer file.Close()

	var ids []int
	json.NewDecoder(file).Decode(&ids)

	for _, id := range ids {
		favs[id] = true
	}
	return favs
}

// SaveFavorites sauvegarde la map des favoris sur le .json
func SaveFavorites(favs map[int]bool) {
	favLock.Lock()
	defer favLock.Unlock()

	var ids []int
	for id, isFav := range favs {
		if isFav {
			ids = append(ids, id)
		}
	}

	file, err := os.Create(favFileName)
	if err == nil {
		defer file.Close()
		json.NewEncoder(file).Encode(ids)
	}
}
