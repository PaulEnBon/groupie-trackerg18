package ui

import (
	"encoding/json"
	"os"
	"sync"
)

const favFileName = "favorites.json"

// Gestionnaire de favoris avec Mutex pour éviter les conflits
var favLock sync.Mutex

// LoadFavorites charge la liste des IDs favoris depuis le disque
func LoadFavorites() map[int]bool {
	favLock.Lock()
	defer favLock.Unlock()

	favs := make(map[int]bool)

	file, err := os.Open(favFileName)
	if err != nil {
		return favs // Retourne vide si le fichier n'existe pas
	}
	defer file.Close()

	var ids []int
	json.NewDecoder(file).Decode(&ids)

	for _, id := range ids {
		favs[id] = true
	}
	return favs
}

// SaveFavorites sauvegarde la map des favoris sur le disque
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
