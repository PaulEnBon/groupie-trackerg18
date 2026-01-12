package ui

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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

// --- PALETTE DE COULEURS CYBERPUNK ---
var (
	ColBackground = color.NRGBA{R: 15, G: 10, B: 25, A: 255}    // Violet très sombre
	ColCard       = color.NRGBA{R: 30, G: 25, B: 45, A: 255}    // Violet/Gris
	ColAccent     = color.NRGBA{R: 0, G: 255, B: 255, A: 255}   // Cyan Fluo
	ColHighlight  = color.NRGBA{R: 255, G: 0, B: 128, A: 255}   // Rose Fluo
	ColText       = color.NRGBA{R: 240, G: 240, B: 255, A: 255} // Blanc bleuté
)

// --- FONCTIONS UTILITAIRES ---

func loadDetailImage(url string, w, h float32) fyne.CanvasObject {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		ph := canvas.NewRectangle(ColCard)
		ph.SetMinSize(fyne.NewSize(w, h))
		return ph
	}
	defer resp.Body.Close()

	imgDecoded, _, err := image.Decode(resp.Body)
	if err != nil {
		ph := canvas.NewRectangle(ColCard)
		ph.SetMinSize(fyne.NewSize(w, h))
		return ph
	}
	img := canvas.NewImageFromImage(imgDecoded)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(w, h))
	return img
}

func createCyberCard(title, value string, icon fyne.Resource) fyne.CanvasObject {
	iconW := widget.NewIcon(icon)

	valText := canvas.NewText(value, ColAccent)
	valText.TextSize = 20
	valText.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	valText.Alignment = fyne.TextAlignCenter

	lblText := canvas.NewText(strings.ToUpper(title), ColText)
	lblText.TextSize = 10
	lblText.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		container.NewCenter(iconW),
		valText,
		lblText,
	)

	bg := canvas.NewRectangle(ColCard)
	border := canvas.NewRectangle(ColHighlight)
	border.SetMinSize(fyne.NewSize(0, 2))

	return container.NewBorder(nil, border, nil, nil,
		container.NewMax(bg, container.NewPadded(content)))
}

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

// --- FONCTION PRINCIPALE ---

func ArtistDetail(app fyne.App, artist models.Artist, onBack func()) fyne.CanvasObject {
	// 1. HEADER
	title := canvas.NewText(strings.ToUpper(artist.Name), ColAccent)
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	title.Alignment = fyne.TextAlignCenter

	decoLine := canvas.NewRectangle(ColHighlight)
	decoLine.SetMinSize(fyne.NewSize(100, 3))

	// 2. STREAMING BAR
	encodedName := url.QueryEscape(artist.Name)
	spotifyUrl, _ := url.Parse("https://open.spotify.com/search/" + encodedName)
	youtubeUrl, _ := url.Parse("https://www.youtube.com/results?search_query=" + encodedName)
	deezerUrl, _ := url.Parse("https://www.deezer.com/search/" + encodedName)

	btnSpotify := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() { app.OpenURL(spotifyUrl) })
	btnYouTube := widget.NewButtonWithIcon("", theme.MediaVideoIcon(), func() { app.OpenURL(youtubeUrl) })
	btnDeezer := widget.NewButtonWithIcon("", theme.VolumeUpIcon(), func() { app.OpenURL(deezerUrl) })

	streamingBar := container.NewGridWithColumns(3, btnSpotify, btnYouTube, btnDeezer)

	// 3. STATS
	relation, err := api.FetchRelation(artist.ID)
	concertCount := 0
	if err == nil && relation != nil {
		for _, dates := range relation.DatesLocations {
			concertCount += len(dates)
		}
	}

	statsGrid := container.NewGridWithColumns(2,
		createCyberCard("Since", fmt.Sprintf("%d", artist.CreationDate), theme.HistoryIcon()),
		createCyberCard("Debut", artist.FirstAlbum, theme.MediaMusicIcon()),
		createCyberCard("Team", fmt.Sprintf("%d", len(artist.Members)), theme.AccountIcon()),
		createCyberCard("Shows", fmt.Sprintf("%d", concertCount), theme.InfoIcon()),
	)

	// 4. MEMBRES
	membersLabel := canvas.NewText("> TEAM_MEMBERS", ColHighlight)
	membersLabel.TextSize = 14
	membersLabel.TextStyle = fyne.TextStyle{Monospace: true}

	membersVBox := container.NewVBox()
	for _, member := range artist.Members {
		txt := canvas.NewText("  "+member, ColText)
		txt.TextSize = 14
		txt.TextStyle = fyne.TextStyle{Monospace: true}
		membersVBox.Add(txt)
	}

	// 5. CONCERTS + MAPS (Panneau Droit)
	concertsTitle := canvas.NewText("> GPS_SATELLITE_LOG", ColHighlight)
	concertsTitle.TextSize = 16
	concertsTitle.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	var concertsContent fyne.CanvasObject
	if err != nil {
		concertsContent = widget.NewLabel("Error loading data...")
	} else if len(relation.DatesLocations) == 0 {
		concertsContent = widget.NewLabel("No dates found.")
	} else {
		cardsContainer := container.NewVBox()

		for location, dates := range relation.DatesLocations {
			locName := toTitle(strings.ReplaceAll(strings.ReplaceAll(location, "-", ", "), "_", " "))

			// --- CARTE GPS ---

			// On utilise SearchIcon car WorldIcon n'existe pas
			mapIcon := widget.NewIcon(theme.SearchIcon())

			// Texte de statut
			statusLbl := widget.NewLabel("SCANNING...")
			statusLbl.Alignment = fyne.TextAlignCenter

			// Bouton Map
			btnMap := widget.NewButtonWithIcon("LOCATE", theme.SearchIcon(), func() {
				u, _ := url.Parse("https://www.openstreetmap.org/search?query=" + url.QueryEscape(locName))
				app.OpenURL(u)
			})

			// --- Lancement téléchargement asynchrone ---
			go func(city string, icon *widget.Icon, status *widget.Label, btn *widget.Button) {
				// Petit délai pour ne pas spammer
				time.Sleep(300 * time.Millisecond)

				// A. GPS (Nominatim)
				latStr, lonStr, err := api.GetCoordinates(city)
				if err != nil {
					fyne.Do(func() { status.SetText("SIGNAL LOST") })
					return
				}

				// B. Calcul Tuile OSM
				lat, _ := strconv.ParseFloat(latStr, 64)
				lon, _ := strconv.ParseFloat(lonStr, 64)
				tileURL := api.GetOSMTileURL(lat, lon, 12)

				// C. Téléchargement Image (AVEC USER-AGENT CORRIGÉ)
				// On crée une requête personnalisée pour éviter le blocage "Access Blocked"
				client := &http.Client{Timeout: 10 * time.Second}
				req, _ := http.NewRequest("GET", tileURL, nil)
				// L'User-Agent est obligatoire pour OpenStreetMap !
				req.Header.Set("User-Agent", "GroupieTracker-StudentProject/1.0")

				resp, errImg := client.Do(req)
				if errImg == nil && resp.StatusCode == 200 {
					data, _ := io.ReadAll(resp.Body)
					resp.Body.Close()

					res := fyne.NewStaticResource("map.png", data)

					// D. Mise à jour UI sécurisée
					fyne.Do(func() {
						icon.SetResource(res)
						status.SetText("")
						status.Hide()

						btn.SetText("GPS LOCK")
						btn.OnTapped = func() {
							u, _ := url.Parse(fmt.Sprintf("https://www.openstreetmap.org/?mlat=%s&mlon=%s#map=12/%s/%s", latStr, lonStr, latStr, lonStr))
							app.OpenURL(u)
						}
					})
				} else {
					// En cas d'erreur de téléchargement de l'image
					fyne.Do(func() { status.SetText("MAP DATA ERR") })
				}
			}(locName, mapIcon, statusLbl, btnMap)

			// --- Correction Taille ---
			// On applique la taille minimale au rectangle de fond (Background)
			bgRect := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 50, A: 255})
			bgRect.SetMinSize(fyne.NewSize(300, 150)) // C'est ici qu'on force la taille

			mapStack := container.NewMax(
				bgRect,
				container.NewPadded(mapIcon),
				container.NewCenter(statusLbl),
			)

			// Assemblage Ligne
			infoBox := container.NewVBox(
				canvas.NewText(":: "+locName, ColAccent),
				mapStack,
				widget.NewLabel(strings.Join(dates, " | ")),
			)

			cardRow := container.NewBorder(nil, nil, nil,
				container.NewVBox(layout.NewSpacer(), btnMap, layout.NewSpacer()),
				container.NewPadded(infoBox))

			bgRow := canvas.NewRectangle(ColCard)
			cardsContainer.Add(container.NewMax(bgRow, container.NewPadded(cardRow)))
			cardsContainer.Add(widget.NewSeparator())
		}
		concertsContent = container.NewVScroll(cardsContainer)
	}

	// --- LAYOUT GLOBAL ---

	avatar := loadDetailImage(artist.Image, 220, 220)
	imgBorder := canvas.NewRectangle(color.Transparent)
	imgBorder.StrokeColor = ColAccent
	imgBorder.StrokeWidth = 2
	avatarFramed := container.NewMax(imgBorder, avatar)

	backButton := widget.NewButtonWithIcon("BACK", theme.NavigateBackIcon(), func() { onBack() })

	leftContent := container.NewVBox(
		container.NewPadded(avatarFramed),
		widget.NewSeparator(),
		title,
		container.NewCenter(decoLine),
		container.NewPadded(streamingBar),
		widget.NewSeparator(),
		container.NewPadded(statsGrid),
		widget.NewSeparator(),
		container.NewPadded(membersLabel),
		container.NewPadded(membersVBox),
	)
	leftScroll := container.NewVScroll(leftContent)

	rightContent := container.NewBorder(
		container.NewVBox(container.NewPadded(concertsTitle), widget.NewSeparator()),
		nil, nil, nil,
		concertsContent,
	)
	rightBg := canvas.NewRectangle(color.NRGBA{R: 20, G: 15, B: 30, A: 255})
	rightPanel := container.NewMax(rightBg, rightContent)

	split := container.NewHSplit(leftScroll, rightPanel)
	split.Offset = 0.40

	header := container.NewVBox(container.NewHBox(backButton, layout.NewSpacer()), widget.NewSeparator())
	mainBg := canvas.NewRectangle(ColBackground)

	return container.NewMax(mainBg, container.NewBorder(header, nil, nil, nil, split))
}
