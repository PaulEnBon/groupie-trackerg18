package main

import (
	"log"

	"groupie-tracker/api"
	"groupie-tracker/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	w := a.NewWindow("Groupie Tracker")

	artists, err := api.FetchArtists()
	if err != nil {
		log.Fatal(err)
	}

	content := ui.ArtistList(a, artists)
	w.SetContent(content)
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
