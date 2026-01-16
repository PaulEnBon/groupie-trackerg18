package ui

import (
	"net/url"
	"strconv"
	"strings"

	"groupie-tracker/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// UserBandForm affiche un formulaire pour créer un groupe personnalisé.
// onSave reçoit l'artiste et la relation (ville -> dates).
func UserBandForm(app fyne.App, win fyne.Window, onBack func(), onSave func(models.Artist, map[string][]string)) fyne.CanvasObject {
	nameEntry := widget.NewEntry()
	imageEntry := widget.NewEntry()
	creationEntry := widget.NewEntry()          // ex: 2010
	firstAlbumEntry := widget.NewEntry()        // ex: 2012-06-01
	membersEntry := widget.NewMultiLineEntry()  // noms séparés par virgule
	concertsEntry := widget.NewMultiLineEntry() // lignes: "Paris - 2024-05-01 | 2024-06-10"
	spotifyEntry := widget.NewEntry()           // URL Spotify optionnel
	youtubeEntry := widget.NewEntry()           // URL YouTube optionnel
	deezerEntry := widget.NewEntry()            // URL Deezer optionnel

	saveBtn := widget.NewButtonWithIcon("Enregistrer", theme.ConfirmIcon(), func() {
		artist := models.Artist{
			Name:         strings.TrimSpace(nameEntry.Text),
			Image:        strings.TrimSpace(imageEntry.Text),
			FirstAlbum:   strings.TrimSpace(firstAlbumEntry.Text),
			CreationDate: parseYearSafe(creationEntry.Text),
			Members:      splitMembersSafe(membersEntry.Text),
			SpotifyLink:  strings.TrimSpace(spotifyEntry.Text),
			YoutubeLink:  strings.TrimSpace(youtubeEntry.Text),
			DeezerLink:   strings.TrimSpace(deezerEntry.Text),
		}
		relations := parseConcertsSafe(concertsEntry.Text)
		onSave(artist, relations)
	})

	openMureka := widget.NewButton("Pas de groupe ? Génère ta musique (Mureka)", func() {
		u, _ := url.Parse("https://www.mureka.ai/fr/?utm_source=google&utm_medium=lote&utm_campaign=search-french101&utm_content=french-english_22638597029_181737480872&utm_term=ai%20music%20generator&utm_source=google&utm_medium=lote&utm_campaign=search-french101&utm_content=french-english_22638597029_181737480872&utm_term=ai%20music%20generator&gad_source=1&gad_campaignid=22638597029&gclid=CjwKCAiA4KfLBhB0EiwAUY7GAUBmu3xm_Pmc-e10YQsRSnaYtbUxAKyks9vf6M44b0Xt4lNTPaJZ6hoCreoQAvD_BwE")
		app.OpenURL(u)
	})

	pickBtn := widget.NewButtonWithIcon("Importer PNG / JPEG", theme.FileImageIcon(), func() {
		fd := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			uri := r.URI().String()
			imageEntry.SetText(uri)
			r.Close()
		}, win)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
		fd.Show()
	})

	form := widget.NewForm(
		widget.NewFormItem("Nom du groupe", nameEntry),
		widget.NewFormItem("Icône (URL ou fichier)", container.NewBorder(nil, nil, nil, pickBtn, imageEntry)),
		widget.NewFormItem("Année de création", creationEntry),
		widget.NewFormItem("Date de début de carrière", firstAlbumEntry),
		widget.NewFormItem("Membres (séparés par des virgules)", membersEntry),
		widget.NewFormItem("Concerts (\"Ville - date1 | date2\")", concertsEntry),
		widget.NewFormItem("Lien Spotify (optionnel)", spotifyEntry),
		widget.NewFormItem("Lien YouTube (optionnel)", youtubeEntry),
		widget.NewFormItem("Lien Deezer (optionnel)", deezerEntry),
	)

	return container.NewBorder(
		container.NewHBox(widget.NewButtonWithIcon("Retour", theme.NavigateBackIcon(), onBack)),
		container.NewVBox(openMureka, saveBtn),
		nil, nil,
		container.NewVScroll(form),
	)
}

func splitMembersSafe(text string) []string {
	parts := strings.Split(text, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseYearSafe(s string) int {
	val, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || val < 0 {
		return 0
	}
	return val
}

func parseConcertsSafe(text string) map[string][]string {
	result := make(map[string][]string)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "-", 2)
		if len(parts) < 2 {
			continue
		}
		city := strings.TrimSpace(parts[0])
		datesRaw := strings.Split(parts[1], "|")
		dates := make([]string, 0, len(datesRaw))
		for _, d := range datesRaw {
			d = strings.TrimSpace(d)
			if d != "" {
				dates = append(dates, d)
			}
		}
		if city != "" && len(dates) > 0 {
			result[city] = dates
		}
	}
	return result
}
