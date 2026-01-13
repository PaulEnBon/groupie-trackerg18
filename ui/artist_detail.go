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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ArtistDetail affiche les détails d'un artiste
func ArtistDetail(app fyne.App, artist models.Artist, isFavorite bool, onBack func(), onToggleFavorite func(bool)) fyne.CanvasObject {

	// --- HEADER ---
	title := canvas.NewText(strings.ToUpper(artist.Name), ColAccent)
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	title.Alignment = fyne.TextAlignCenter

	// --- BOUTON FAVORIS ---
	var favBtn *widget.Button
	updateFavBtn := func(state bool) {
		if state {
			favBtn.SetText("FAVORIS")
			favBtn.SetIcon(theme.ConfirmIcon())
			favBtn.Importance = widget.HighImportance
		} else {
			favBtn.SetText("Ajouter aux favoris")
			favBtn.SetIcon(theme.ContentAddIcon())
			favBtn.Importance = widget.MediumImportance
		}
	}

	favBtn = widget.NewButton("FAV", func() {
		isFavorite = !isFavorite
		updateFavBtn(isFavorite)
		onToggleFavorite(isFavorite)
	})
	updateFavBtn(isFavorite)

	// Header container
	headerTop := container.NewBorder(nil, nil,
		widget.NewButtonWithIcon("RETOUR", theme.NavigateBackIcon(), onBack),
		favBtn,
		nil,
	)

	// --- BARRE DE STREAMING ---
	encodedName := url.QueryEscape(artist.Name)
	spotifyUrl, _ := url.Parse("https://open.spotify.com/search/" + encodedName)
	youtubeUrl, _ := url.Parse("https://www.youtube.com/results?search_query=" + encodedName)
	deezerUrl, _ := url.Parse("https://www.deezer.com/search/" + encodedName)

	streamingBar := container.NewGridWithColumns(3,
		widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() { app.OpenURL(spotifyUrl) }),
		widget.NewButtonWithIcon("", theme.MediaVideoIcon(), func() { app.OpenURL(youtubeUrl) }),
		widget.NewButtonWithIcon("", theme.VolumeUpIcon(), func() { app.OpenURL(deezerUrl) }),
	)

	// --- STATS ---
	relation, err := api.FetchRelation(artist.ID)
	concertCount := 0
	if err == nil && relation != nil {
		for _, dates := range relation.DatesLocations {
			concertCount += len(dates)
		}
	}

	statsGrid := container.NewGridWithColumns(2,
		createCyberCard("Depuis", fmt.Sprintf("%d", artist.CreationDate), theme.HistoryIcon()),
		createCyberCard("Début", artist.FirstAlbum, theme.MediaMusicIcon()),
		createCyberCard("Équipe", fmt.Sprintf("%d", len(artist.Members)), theme.AccountIcon()),
		createCyberCard("Concerts", fmt.Sprintf("%d", concertCount), theme.InfoIcon()),
	)

	// --- LISTE MEMBRES ---
	membersVBox := container.NewVBox()
	for _, m := range artist.Members {
		membersVBox.Add(canvas.NewText(" "+m, ColText))
	}

	// --- CONCERTS + MAPS ---
	concertsTitle := canvas.NewText("> Vue Satellite", ColHighlight)
	concertsTitle.TextSize = 16
	concertsTitle.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	cardsContainer := container.NewVBox()
	if err == nil && len(relation.DatesLocations) > 0 {
		for location, dates := range relation.DatesLocations {
			locName := toTitle(strings.ReplaceAll(strings.ReplaceAll(location, "-", ", "), "_", " "))

			mapIcon := widget.NewIcon(theme.SearchIcon())
			statusLbl := widget.NewLabel("SCANNING...")
			statusLbl.Alignment = fyne.TextAlignCenter

			btnMap := widget.NewButtonWithIcon("LOCATE", theme.SearchIcon(), func() {
				u, _ := url.Parse("https://www.openstreetmap.org/search?query=" + url.QueryEscape(locName))
				app.OpenURL(u)
			})

			// Goroutine GPS & Tuiles
			go func(city string, icon *widget.Icon, status *widget.Label, btn *widget.Button) {
				time.Sleep(200 * time.Millisecond)
				latStr, lonStr, err := api.GetCoordinates(city)
				if err != nil {
					fyne.Do(func() { status.SetText("SIGNAL LOST") })
					return
				}

				lat, _ := strconv.ParseFloat(latStr, 64)
				lon, _ := strconv.ParseFloat(lonStr, 64)
				tileURL := api.GetOSMTileURL(lat, lon, 12)

				client := &http.Client{Timeout: 10 * time.Second}
				req, _ := http.NewRequest("GET", tileURL, nil)
				req.Header.Set("User-Agent", "GroupieTracker/1.0")

				resp, errImg := client.Do(req)
				if errImg == nil && resp.StatusCode == 200 {
					data, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					res := fyne.NewStaticResource("map.png", data)

					fyne.Do(func() {
						icon.SetResource(res)
						status.Hide()
						btn.SetText("PLAN")
						btn.OnTapped = func() {
							u, _ := url.Parse(fmt.Sprintf("https://www.openstreetmap.org/?mlat=%s&mlon=%s", latStr, lonStr))
							app.OpenURL(u)
						}
					})
				}
			}(locName, mapIcon, statusLbl, btnMap)

			// --- MODIFICATION TAILLE CARTE ---
			// C'est ici qu'on change la taille ! (600x350 au lieu de 300x150)
			bgRect := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 50, A: 255})
			bgRect.SetMinSize(fyne.NewSize(600, 350)) // <--- TAILLE AGRANDIE

			mapStack := container.NewMax(bgRect, container.NewPadded(mapIcon), container.NewCenter(statusLbl))

			infoBox := container.NewVBox(
				canvas.NewText(":: "+locName, ColAccent),
				mapStack,
				widget.NewLabel(strings.Join(dates, " | ")),
			)

			row := container.NewBorder(nil, nil, nil,
				container.NewVBox(layout.NewSpacer(), btnMap, layout.NewSpacer()),
				container.NewPadded(infoBox),
			)

			bgRow := canvas.NewRectangle(ColCard)
			cardsContainer.Add(container.NewMax(bgRow, container.NewPadded(row)))
			cardsContainer.Add(widget.NewSeparator())
		}
	} else {
		cardsContainer.Add(widget.NewLabel("No Data Available"))
	}

	// --- LAYOUT GLOBAL ---
	avatar := loadDetailImage(artist.Image, 220, 220)
	imgBorder := canvas.NewRectangle(color.Transparent)
	imgBorder.StrokeColor = ColAccent
	imgBorder.StrokeWidth = 2

	left := container.NewVBox(
		container.NewMax(imgBorder, avatar),
		widget.NewSeparator(),
		title,
		container.NewPadded(streamingBar),
		widget.NewSeparator(),
		container.NewPadded(statsGrid),
		widget.NewSeparator(),
		container.NewPadded(membersVBox),
	)

	right := container.NewBorder(
		container.NewPadded(concertsTitle),
		nil, nil, nil,
		container.NewVScroll(cardsContainer),
	)

	rightBg := canvas.NewRectangle(color.NRGBA{R: 20, G: 15, B: 30, A: 255})

	split := container.NewHSplit(
		container.NewVScroll(left),
		container.NewMax(rightBg, right),
	)
	split.Offset = 0.35 // Ajusté un peu pour laisser plus de place aux cartes à droite

	mainBg := canvas.NewRectangle(ColBackground)

	page := container.NewBorder(
		container.NewVBox(headerTop, widget.NewSeparator()),
		nil, nil, nil,
		split,
	)

	return container.NewMax(mainBg, page)
}

// --- FONCTIONS UTILITAIRES ---

func createCyberCard(title, value string, icon fyne.Resource) fyne.CanvasObject {
	iconW := widget.NewIcon(icon)
	valText := canvas.NewText(value, ColAccent)
	valText.TextSize = 20
	valText.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	valText.Alignment = fyne.TextAlignCenter
	lblText := canvas.NewText(strings.ToUpper(title), ColText)
	lblText.TextSize = 10
	lblText.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(container.NewCenter(iconW), valText, lblText)
	bg := canvas.NewRectangle(ColCard)
	border := canvas.NewRectangle(ColHighlight)
	border.SetMinSize(fyne.NewSize(0, 2))

	return container.NewBorder(nil, border, nil, nil, container.NewMax(bg, container.NewPadded(content)))
}

func loadDetailImage(url string, w, h float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(ColCard)
	rect.SetMinSize(fyne.NewSize(w, h))
	c := container.NewMax(rect)

	go func() {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			imgData, _, errDec := image.Decode(resp.Body)
			if errDec == nil {
				fyne.Do(func() {
					img := canvas.NewImageFromImage(imgData)
					img.FillMode = canvas.ImageFillContain
					img.SetMinSize(fyne.NewSize(w, h))
					c.Objects = []fyne.CanvasObject{img}
					c.Refresh()
				})
			}
		}
	}()
	return c
}
