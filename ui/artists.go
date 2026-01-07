package ui

import (
	"image/color"

	"groupie-tracker/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func ArtistList(app fyne.App, artists []models.Artist) fyne.CanvasObject {
	var cards []fyne.CanvasObject

	for _, artist := range artists {
		name := canvas.NewText(artist.Name, color.White)
		name.TextSize = 16
		name.Alignment = fyne.TextAlignCenter

		year := canvas.NewText(
			"Created: "+string(rune(artist.CreationDate)),
			color.Gray{Y: 200},
		)

		card := container.NewVBox(
			name,
			year,
		)

		cards = append(cards, card)
	}

	return container.NewVScroll(container.NewVBox(cards...))
}
