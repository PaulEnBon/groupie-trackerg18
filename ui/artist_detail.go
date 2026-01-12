package ui

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"strings"

	"groupie-tracker/api"
	"groupie-tracker/models"

	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// loadDetailImage charge l'image de l'artiste
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
		"📅 Création: %d\n"+
			"🎵 1er Album: %s\n"+
			"👥 Membres: %d",
		artist.CreationDate,
		artist.FirstAlbum,
		len(artist.Members),
	)
	info := widget.NewLabel(infoText)
	info.Wrapping = fyne.TextWrapWord
	info.Alignment = fyne.TextAlignCenter

	// --- NOUVEAU : BOUTONS STREAMING ---
	streamingTitle := canvas.NewText("🎧 Écouter sur", color.NRGBA{R: 180, G: 180, B: 180, A: 255})
	streamingTitle.TextSize = 16
	streamingTitle.Alignment = fyne.TextAlignCenter

	// Création des URLs de recherche
	encodedName := url.QueryEscape(artist.Name)
	spotifyUrl, _ := url.Parse("https://open.spotify.com/search/" + encodedName)
	youtubeUrl, _ := url.Parse("https://www.youtube.com/results?search_query=" + encodedName)
	deezerUrl, _ := url.Parse("https://www.deezer.com/search/" + encodedName)

	// Boutons avec icônes (génériques car Fyne n'a pas les logos de marques)
	btnSpotify := widget.NewButtonWithIcon("Spotify", theme.MediaPlayIcon(), func() {
		app.OpenURL(spotifyUrl)
	})
	btnYouTube := widget.NewButtonWithIcon("YouTube", theme.MediaVideoIcon(), func() {
		app.OpenURL(youtubeUrl)
	})
	btnDeezer := widget.NewButtonWithIcon("Deezer", theme.VolumeUpIcon(), func() {
		app.OpenURL(deezerUrl)
	})

	streamingGrid := container.NewGridWithColumns(1,
		btnSpotify,
		btnYouTube,
		btnDeezer,
	)
	// -----------------------------------

	// Liste des membres
	membersTitle := canvas.NewText("Membres du groupe", color.White)
	membersTitle.TextSize = 20
	membersTitle.TextStyle = fyne.TextStyle{Bold: true}

	membersVBox := container.NewVBox()
	for _, member := range artist.Members {
		lbl := widget.NewLabel("• " + member)
		lbl.Wrapping = fyne.TextWrapWord
		lbl.TextStyle = fyne.TextStyle{Italic: true}
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
		// Conteneur principal pour les cartes de concerts
		cardsContainer := container.NewVBox()

		for location, dates := range relation.DatesLocations {
			// Formater le nom de la ville
			locationRaw := strings.ReplaceAll(location, "-", ", ")
			locationRaw = strings.ReplaceAll(locationRaw, "_", " ")
			locationTitle := toTitle(locationRaw)

			// URL Google Maps
			query := url.QueryEscape(locationTitle)
			mapURL := "https://www.google.com/maps/search/?api=1&query=" + query
			parsedURL, _ := url.Parse(mapURL)

			// Contenu de la carte
			cityLabel := widget.NewLabel(locationTitle)
			cityLabel.TextStyle = fyne.TextStyle{Bold: true}
			cityLabel.Wrapping = fyne.TextWrapWord

			var datesText string
			for _, date := range dates {
				datesText += "🗓️ " + date + "\n"
			}
			datesLabel := widget.NewLabel(datesText)
			datesLabel.Wrapping = fyne.TextWrapWord

			// Bouton Map
			mapButton := widget.NewButtonWithIcon("Voir sur la carte", theme.SearchIcon(), func() {
				app.OpenURL(parsedURL)
			})

			cardContent := container.NewVBox(
				cityLabel,
				datesLabel,
				container.NewHBox(layout.NewSpacer(), mapButton),
			)

			// Fond de carte
			cardBg := canvas.NewRectangle(color.NRGBA{R: 60, G: 60, B: 60, A: 255})

			cardItem := container.NewMax(
				cardBg,
				container.NewPadded(cardContent),
			)

			cardsContainer.Add(cardItem)
			cardsContainer.Add(layout.NewSpacer())
		}

		concertsContent = container.NewVScroll(cardsContainer)
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
	stats := widget.NewLabel(statsText)
	stats.Wrapping = fyne.TextWrapWord

	// Bouton retour
	backButton := widget.NewButtonWithIcon("Retour à la liste", theme.NavigateBackIcon(), func() {
		onBack()
	})

	// ========== STRUCTURE GLOBALE ==========

	// Image de l'artiste
	avatar := loadDetailImage(artist.Image, 200, 200)

	// Panneau Gauche : On ajoute la section Streaming ici
	leftPanel := container.NewVBox(
		container.NewCenter(avatar),
		widget.NewSeparator(),
		container.NewPadded(info),
		widget.NewSeparator(),
		container.NewPadded(streamingTitle), // Titre Streaming
		container.NewPadded(streamingGrid),  // Boutons Streaming
		widget.NewSeparator(),
		container.NewPadded(membersTitle),
		container.NewPadded(membersVBox),
		widget.NewSeparator(),
		container.NewPadded(statsTitle),
		container.NewPadded(stats),
	)
	leftScroll := container.NewVScroll(leftPanel)

	// Panneau Droit (Concerts)
	rightPanel := container.NewVBox(
		container.NewPadded(concertsTitle),
		widget.NewSeparator(),
	)

	rightContainer := container.NewBorder(
		rightPanel, // Haut
		nil, nil, nil,
		concertsContent, // Centre (Scroll)
	)

	// Split Container
	split := container.NewHSplit(leftScroll, rightContainer)
	split.Offset = 0.35

	// Header
	header := container.NewVBox(
		container.NewHBox(backButton, layout.NewSpacer()),
		title,
		widget.NewSeparator(),
	)

	// Layout Final
	content := container.NewBorder(
		header,
		nil, nil, nil,
		split,
	)

	// Fond général
	background := canvas.NewRectangle(color.NRGBA{R: 25, G: 25, B: 25, A: 255})

	return container.NewMax(background, content)
}
