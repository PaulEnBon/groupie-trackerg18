package ui

import (
	"fmt"
	"image/color"
	"strings"

	"groupie-tracker/api"
	"groupie-tracker/models"

	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ArtistDetail(app fyne.App, artist models.Artist, onBack func()) fyne.CanvasObject {
	// Fonction pour capitaliser les mots
	toTitle := func(s string) string {
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

	// Titre avec le nom de l'artiste (texte clair sur fond sombre)
	title := canvas.NewText(artist.Name, color.White)
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	// Informations générales
	infoText := fmt.Sprintf(
		"📅 Année de création: %d\n"+
			"🎵 Premier album: %s\n"+
			"👥 Nombre de membres: %d",
		artist.CreationDate,
		artist.FirstAlbum,
		len(artist.Members),
	)
	info := canvas.NewText(infoText, color.NRGBA{R: 235, G: 235, B: 235, A: 255})
	info.TextSize = 14

	// Liste des membres
	membersTitle := canvas.NewText("Membres du groupe", color.White)
	membersTitle.TextSize = 16
	membersTitle.TextStyle = fyne.TextStyle{Bold: true}

	membersList := widget.NewList(
		func() int { return len(artist.Members) },
		func() fyne.CanvasObject {
			return canvas.NewText("", color.NRGBA{R: 235, G: 235, B: 235, A: 255})
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			ct := o.(*canvas.Text)
			ct.Text = "• " + artist.Members[i]
			ct.Color = color.NRGBA{R: 235, G: 235, B: 235, A: 255}
			ct.Refresh()
		},
	)
	membersList.Resize(fyne.NewSize(400, float32(len(artist.Members)*40))) // Resize to fixed size

	// Récupérer les relations (concerts)
	relation, err := api.FetchRelation(artist.ID)

	// Section des concerts
	concertsTitle := canvas.NewText("🌍 Concerts et Tournées", color.White)
	concertsTitle.TextSize = 18
	concertsTitle.TextStyle = fyne.TextStyle{Bold: true}

	var concertsContent fyne.CanvasObject
	if err != nil {
		msg := canvas.NewText("❌ Impossible de charger les informations de concerts", color.NRGBA{R: 235, G: 235, B: 235, A: 255})
		msg.TextSize = 14
		concertsContent = msg
	} else if len(relation.DatesLocations) == 0 {
		msg := canvas.NewText("Aucun concert programmé", color.NRGBA{R: 235, G: 235, B: 235, A: 255})
		msg.TextSize = 14
		concertsContent = msg
	} else {
		concertsList := container.NewVBox()

		for location, dates := range relation.DatesLocations {
			// Formater le nom de la ville
			locationFormatted := strings.ReplaceAll(location, "-", ", ")
			locationFormatted = strings.ReplaceAll(locationFormatted, "_", " ")
			locationFormatted = toTitle(locationFormatted)

			locationLabel := canvas.NewText("📍 "+locationFormatted, color.RGBA{R: 120, G: 220, B: 255, A: 255})
			locationLabel.TextSize = 14
			locationLabel.TextStyle = fyne.TextStyle{Bold: true}

			concertsList.Add(locationLabel)

			// Ajouter les dates
			for _, date := range dates {
				dateLabel := canvas.NewText("   🗓️  "+date, color.NRGBA{R: 235, G: 235, B: 235, A: 255})
				dateLabel.TextSize = 12
				concertsList.Add(dateLabel)
			}

			// Séparateur
			separator := canvas.NewRectangle(color.Gray{Y: 80})
			separator.SetMinSize(fyne.NewSize(0, 1))
			concertsList.Add(separator)
		}

		concertsContent = container.NewVScroll(concertsList)
	}

	// Statistiques
	statsTitle := canvas.NewText("📊 Statistiques", color.White)
	statsTitle.TextSize = 16
	statsTitle.TextStyle = fyne.TextStyle{Bold: true}

	concertCount := 0
	locationCount := 0
	if err == nil && relation != nil {
		locationCount = len(relation.DatesLocations)
		for _, dates := range relation.DatesLocations {
			concertCount += len(dates)
		}
	}

	statsText := fmt.Sprintf(
		"🎤 Total de concerts: %d\n"+
			"🌍 Pays/Villes visités: %d\n"+
			"🎸 Années actives: %d ans",
		concertCount,
		locationCount,
		2026-artist.CreationDate,
	)
	stats := canvas.NewText(statsText, color.NRGBA{R: 235, G: 235, B: 235, A: 255})
	stats.TextSize = 14

	// Bouton retour
	backButton := widget.NewButton("⬅ Retour à la liste", func() {
		onBack()
	})

	// Layout principal
	// Panneaux opaques et paddés pour la lisibilité
	leftPanel := container.NewVBox(
		title,
		widget.NewSeparator(),
		info,
		widget.NewSeparator(),
		membersTitle,
		membersList,
		widget.NewSeparator(),
		statsTitle,
		stats,
	)
	leftBg := canvas.NewRectangle(color.NRGBA{R: 28, G: 28, B: 28, A: 255})
	leftScroll := container.NewVScroll(container.NewMax(leftBg, container.NewPadded(leftPanel)))

	rightPanel := container.NewVBox(
		concertsTitle,
		widget.NewSeparator(),
		concertsContent,
	)
	rightBg := canvas.NewRectangle(color.NRGBA{R: 28, G: 28, B: 28, A: 255})
	rightScroll := container.NewVScroll(container.NewMax(rightBg, container.NewPadded(rightPanel)))

	split := container.NewHSplit(leftScroll, rightScroll)
	split.Offset = 0.4 // 40% gauche, 60% droite

	content := container.NewBorder(
		container.NewVBox(backButton, widget.NewSeparator()),
		nil,
		nil,
		nil,
		split,
	)

	// Fond opaque pour éviter la transparence sur la liste
	background := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 240})
	return container.NewMax(background, content)
}
