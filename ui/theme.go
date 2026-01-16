package ui

import (
	"image/color"
	"strings"
	"unicode"
)

// --- COULEURS CYBERPUNK ---
var (
	ColBackground = color.NRGBA{R: 28, G: 30, B: 48, A: 255}    // Fond plus lisible
	ColCard       = color.NRGBA{R: 45, G: 52, B: 78, A: 255}    // Cartes contrastées
	ColAccent     = color.NRGBA{R: 0, G: 255, B: 255, A: 255}   // Cyan Fluo
	ColHighlight  = color.NRGBA{R: 255, G: 0, B: 128, A: 255}   // Rose Fluo
	ColText       = color.NRGBA{R: 240, G: 240, B: 255, A: 255} // Blanc bleuté
)

// toTitle : Fonction utilitaire pour le formatage de texte
func toTitle(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}
