package ui

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"sort"
	"strings"

	"groupie-tracker/api"
	"groupie-tracker/models"

	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// loadDetailImage fetches artist image for detail view
func loadDetailImage(url string, w, h float32) fyne.CanvasObject {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		ph := canvas.NewRectangle(color.NRGBA{R: 60, G: 60, B: 60, A: 255})
		ph.SetMinSize(fyne.NewSize(w, h))
		return ph
	}
	defer resp.Body.Close()

	imgDecoded, _, err := image.Decode(resp.Body)
	if err != nil {
		ph := canvas.NewRectangle(color.NRGBA{R: 60, G: 60, B: 60, A: 255})
		ph.SetMinSize(fyne.NewSize(w, h))
		return ph
	}
	img := canvas.NewImageFromImage(imgDecoded)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(w, h))
	return img
}

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

	// CORRECTION 1 : Utilisation de Label avec Wrapping pour les infos
	info := widget.NewLabel(infoText)
	info.Wrapping = fyne.TextWrapWord

	// Liste des membres
	membersTitle := canvas.NewText("Membres du groupe", color.White)
	membersTitle.TextSize = 20
	membersTitle.TextStyle = fyne.TextStyle{Bold: true}

	// CORRECTION 2 : Remplacement de widget.NewList par un VBox simple
	// Cela force l'affichage de tout le texte sans couper
	membersVBox := container.NewVBox()
	for _, member := range artist.Members {
		// Création d'un label pour chaque membre
		lbl := widget.NewLabel("• " + member)
		lbl.Wrapping = fyne.TextWrapWord             // Important pour les longs noms
		lbl.TextStyle = fyne.TextStyle{Italic: true} // Style optionnel
		membersVBox.Add(lbl)
	}

	// Récupérer les relations (concerts)
	relation, err := api.FetchRelation(artist.ID)

	// Section des concerts
	concertsTitle := canvas.NewText("🌍 Concerts et Tournées", color.White)
	concertsTitle.TextSize = 20
	concertsTitle.TextStyle = fyne.TextStyle{Bold: true}

	var concertsContent fyne.CanvasObject
	if err != nil {
		msg := widget.NewLabel("❌ Impossible de charger les informations de concerts")
		msg.Wrapping = fyne.TextWrapWord
		concertsContent = msg
	} else if len(relation.DatesLocations) == 0 {
		msg := widget.NewLabel("Aucun concert programmé")
		msg.Wrapping = fyne.TextWrapWord
		concertsContent = msg
	} else {
		// Ordonner les lieux pour un affichage stable
		locs := make([]string, 0, len(relation.DatesLocations))
		totalDates := 0
		for loc, dates := range relation.DatesLocations {
			locs = append(locs, loc)
			totalDates += len(dates)
		}
		sort.Strings(locs)

		header := widget.NewLabel(fmt.Sprintf("%d lieux • %d dates", len(locs), totalDates))
		header.TextStyle = fyne.TextStyle{Bold: true}

		concertCards := container.NewVBox(header, widget.NewSeparator())

		for _, location := range locs {
			dates := relation.DatesLocations[location]
			locationFormatted := strings.ReplaceAll(location, "-", ", ")
			locationFormatted = strings.ReplaceAll(locationFormatted, "_", " ")
			locationFormatted = toTitle(locationFormatted)

			locLabel := widget.NewLabel("📍 " + locationFormatted)
			locLabel.TextStyle = fyne.TextStyle{Bold: true}
			locLabel.Wrapping = fyne.TextWrapWord

			datesLabel := widget.NewLabel("🗓️  " + strings.Join(dates, "  •  "))
			datesLabel.Wrapping = fyne.TextWrapWord

			cardBg := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 40, A: 255})
			cardBg.SetMinSize(fyne.NewSize(0, 100))

			cardContent := container.NewVBox(
				locLabel,
				widget.NewSeparator(),
				datesLabel,
			)
			card := container.NewMax(cardBg, container.NewPadded(cardContent))

			concertCards.Add(card)
			concertCards.Add(widget.NewSeparator())
		}

		concertsContent = container.NewVScroll(concertCards)
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

	// CORRECTION 4 : Label avec Wrapping pour les stats
	stats := widget.NewLabel(statsText)
	stats.Wrapping = fyne.TextWrapWord

	// Bouton retour
	backButton := widget.NewButton("⬅ Retour à la liste", func() {
		onBack()
	})

	// ========== PANNEAU GAUCHE - INFOS GÉNÉRALES ==========
	// Avatar en grand au-dessus du titre
	avatar := loadDetailImage(artist.Image, 180, 180)

	infoBox := container.NewVBox(
		container.NewCenter(avatar),
		widget.NewSeparator(),
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
	// On utilise membersVBox ici au lieu de la List
	membersBox := container.NewVBox(
		membersTitle,
		widget.NewSeparator(),
		membersVBox,
	)
	membersBg := canvas.NewRectangle(color.NRGBA{R: 32, G: 32, B: 32, A: 255})
	// On augmente un peu la taille min pour être sûr
	membersBg.SetMinSize(fyne.NewSize(300, 300))
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

	// Scroll global pour la gauche
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

	// Scroll global pour la droite
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

	// Fond opaque
	background := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 240})
	return container.NewMax(background, content)
}
