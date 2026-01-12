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

	// Titre avec le nom de l'artiste
	title := canvas.NewText(artist.Name, color.White)
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	// Informations générales
	infoText := fmt.Sprintf(
		"📅 Année de création: %d\n\n"+
			"🎵 Premier album: %s\n\n"+
			"👥 Nombre de membres: %d",
		artist.CreationDate,
		artist.FirstAlbum,
		len(artist.Members),
	)
	info := canvas.NewText(infoText, color.NRGBA{R: 235, G: 235, B: 235, A: 255})
	info.TextSize = 18

	// Liste des membres
	membersTitle := canvas.NewText("Membres du groupe", color.White)
	membersTitle.TextSize = 20
	membersTitle.TextStyle = fyne.TextStyle{Bold: true}

	membersList := widget.NewList(
		func() int { return len(artist.Members) },
		func() fyne.CanvasObject {
			return canvas.NewText("", color.NRGBA{R: 235, G: 235, B: 235, A: 255})
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			ct := o.(*canvas.Text)
			ct.Text = "• " + artist.Members[i]
			ct.TextSize = 16
			ct.Color = color.NRGBA{R: 235, G: 235, B: 235, A: 255}
			ct.Refresh()
		},
	)
	membersList.Resize(fyne.NewSize(0, float32(len(artist.Members)*50)))

	// Récupérer les relations (concerts)
	relation, err := api.FetchRelation(artist.ID)

	// Section des concerts
	concertsTitle := canvas.NewText("🌍 Concerts et Tournées", color.White)
	concertsTitle.TextSize = 20
	concertsTitle.TextStyle = fyne.TextStyle{Bold: true}

	var concertsContent fyne.CanvasObject
	if err != nil {
		msg := canvas.NewText("❌ Impossible de charger les informations de concerts", color.NRGBA{R: 235, G: 235, B: 235, A: 255})
		msg.TextSize = 16
		concertsContent = msg
	} else if len(relation.DatesLocations) == 0 {
		msg := canvas.NewText("Aucun concert programmé", color.NRGBA{R: 235, G: 235, B: 235, A: 255})
		msg.TextSize = 16
		concertsContent = msg
	} else {
		concertsList := container.NewVBox()

		for location, dates := range relation.DatesLocations {
			// Formater le nom de la ville
			locationFormatted := strings.ReplaceAll(location, "-", ", ")
			locationFormatted = strings.ReplaceAll(locationFormatted, "_", " ")
			locationFormatted = toTitle(locationFormatted)

			locationLabel := canvas.NewText("📍 "+locationFormatted, color.RGBA{R: 120, G: 220, B: 255, A: 255})
			locationLabel.TextSize = 16
			locationLabel.TextStyle = fyne.TextStyle{Bold: true}

			concertsList.Add(locationLabel)

			// Ajouter les dates
			for _, date := range dates {
				dateLabel := canvas.NewText("   🗓️  "+date, color.NRGBA{R: 235, G: 235, B: 235, A: 255})
				dateLabel.TextSize = 14
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
	statsTitle.TextSize = 20
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
		"🎤 Total de concerts: %d\n\n"+
			"🌍 Pays/Villes visités: %d\n\n"+
			"🎸 Années actives: %d ans",
		concertCount,
		locationCount,
		2026-artist.CreationDate,
	)
	stats := canvas.NewText(statsText, color.NRGBA{R: 235, G: 235, B: 235, A: 255})
	stats.TextSize = 18

	// Bouton retour
	backButton := widget.NewButton("⬅ Retour à la liste", func() {
		onBack()
	})

	// ========== PANNEAU GAUCHE - INFOS GÉNÉRALES ==========
	infoBox := container.NewVBox(
		title,
		widget.NewSeparator(),
		info,
	)
	infoBg := canvas.NewRectangle(color.NRGBA{R: 32, G: 32, B: 32, A: 255})
	infoBg.SetMinSize(fyne.NewSize(300, 250))
	infoPadded := container.NewVBox(
		container.NewCenter(container.NewMax(infoBg, container.NewPadded(infoBox))),
	)

	// ========== PANNEAU GAUCHE - MEMBRES ==========
	membersBox := container.NewVBox(
		membersTitle,
		widget.NewSeparator(),
		membersList,
	)
	membersBg := canvas.NewRectangle(color.NRGBA{R: 32, G: 32, B: 32, A: 255})
	membersBg.SetMinSize(fyne.NewSize(300, 400))
	membersPadded := container.NewVBox(
		container.NewCenter(container.NewMax(membersBg, container.NewPadded(membersBox))),
	)

	// ========== PANNEAU GAUCHE - STATS ==========
	statsBox := container.NewVBox(
		statsTitle,
		widget.NewSeparator(),
		stats,
	)
	statsBg := canvas.NewRectangle(color.NRGBA{R: 32, G: 32, B: 32, A: 255})
	statsBg.SetMinSize(fyne.NewSize(300, 250))
	statsPadded := container.NewVBox(
		container.NewCenter(container.NewMax(statsBg, container.NewPadded(statsBox))),
	)

	// ========== PANNEAU GAUCHE COMPLET ==========
	leftPanelContent := container.NewVBox(
		infoPadded,
		widget.NewSeparator(),
		membersPadded,
		widget.NewSeparator(),
		statsPadded,
	)
	leftScroll := container.NewVScroll(leftPanelContent)

	// ========== PANNEAU DROIT - CONCERTS ==========
	rightPanel := container.NewVBox(
		concertsTitle,
		widget.NewSeparator(),
		concertsContent,
	)
	rightBg := canvas.NewRectangle(color.NRGBA{R: 32, G: 32, B: 32, A: 255})
	rightBg.SetMinSize(fyne.NewSize(400, 900))
	rightPadded := container.NewMax(rightBg, container.NewPadded(rightPanel))
	rightScroll := container.NewVScroll(rightPadded)
	rightScroll.SetMinSize(fyne.NewSize(400, 900))

	split := container.NewHSplit(leftScroll, rightScroll)
	split.Offset = 0.35 // 35% gauche, 65% droite

	splitContainer := container.NewVBox(split)
	splitContainer.Resize(fyne.NewSize(0, 950))

	content := container.NewBorder(
		container.NewVBox(backButton, widget.NewSeparator()),
		nil,
		nil,
		nil,
		splitContainer,
	)

	// Fond opaque pour éviter la transparence sur la liste
	background := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 240})
	return container.NewMax(background, content)
}
