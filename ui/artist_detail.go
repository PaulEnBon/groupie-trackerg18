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

func ArtistDetail(app fyne.App, artist models.Artist, isFavorite bool, onBack func(), onToggleFavorite func(bool)) fyne.CanvasObject {

	title := canvas.NewText(strings.ToUpper(artist.Name), ColAccent)
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	title.Alignment = fyne.TextAlignCenter

	var favBtn *widget.Button
	updateFavBtn := func(state bool) {
		if state {
			favBtn.SetText(TR("fav_yes"))
			favBtn.SetIcon(theme.ConfirmIcon())
			favBtn.Importance = widget.HighImportance
		} else {
			favBtn.SetText(TR("fav_add"))
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

	headerTop := container.NewBorder(nil, nil,
		widget.NewButtonWithIcon(TR("back_btn"), theme.NavigateBackIcon(), onBack),
		favBtn,
		nil,
	)

	buttons := make([]fyne.CanvasObject, 0)

	wikiUrl, _ := url.Parse("https://www.wikipedia.org/w/index.php?search=" + url.QueryEscape(artist.Name))
	btnWiki := widget.NewButtonWithIcon(TR("wiki_btn"), theme.SearchIcon(), func() {
		app.OpenURL(wikiUrl)
	})
	buttons = append(buttons, btnWiki)

	if artist.SpotifyLink != "" {
		spotifyUrl, _ := url.Parse(artist.SpotifyLink)
		buttons = append(buttons, widget.NewButton("SPOTIFY", func() { app.OpenURL(spotifyUrl) }))
	}
	if artist.YoutubeLink != "" {
		youtubeUrl, _ := url.Parse(artist.YoutubeLink)
		buttons = append(buttons, widget.NewButton("YOUTUBE", func() { app.OpenURL(youtubeUrl) }))
	}
	if artist.DeezerLink != "" {
		deezerUrl, _ := url.Parse(artist.DeezerLink)
		buttons = append(buttons, widget.NewButton("DEEZER", func() { app.OpenURL(deezerUrl) }))
	}

	streamingBar := container.NewGridWithColumns(len(buttons), buttons...)

	// --- STATS (Traduit) ---
	relation, err := api.FetchRelation(artist.ID)
	concertCount := 0
	if err == nil && relation != nil {
		for _, dates := range relation.DatesLocations {
			concertCount += len(dates)
		}
	}

	statsGrid := container.NewGridWithColumns(2,
		createCyberCard(TR("since"), fmt.Sprintf("%d", artist.CreationDate), theme.HistoryIcon()),
		createCyberCard(TR("start"), artist.FirstAlbum, theme.MediaMusicIcon()),
		createCyberCard(TR("team"), fmt.Sprintf("%d", len(artist.Members)), theme.AccountIcon()),
		createCyberCard(TR("concerts_cnt"), fmt.Sprintf("%d", concertCount), theme.InfoIcon()),
	)

	membersVBox := container.NewVBox()
	for _, m := range artist.Members {
		membersVBox.Add(canvas.NewText(" "+m, ColText))
	}

	concertsTitle := canvas.NewText(TR("sat_view"), ColHighlight)
	concertsTitle.TextSize = 16
	concertsTitle.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	cardsContainer := container.NewVBox()
	if err == nil && len(relation.DatesLocations) > 0 {

		requestIndex := 0

		for location, dates := range relation.DatesLocations {
			locName := toTitle(strings.ReplaceAll(strings.ReplaceAll(location, "-", ", "), "_", " "))

			mapIcon := widget.NewIcon(theme.SearchIcon())
			statusLbl := widget.NewLabel(TR("scanning"))
			statusLbl.Alignment = fyne.TextAlignCenter

			btnMap := widget.NewButtonWithIcon(TR("loc_proc"), theme.SearchIcon(), func() {
				u, _ := url.Parse("https://www.openstreetmap.org/search?query=" + url.QueryEscape(locName))
				app.OpenURL(u)
			})

			pin := canvas.NewCircle(color.NRGBA{R: 255, G: 0, B: 50, A: 255})
			pin.Hide()
			pinWrapper := container.NewGridWrap(fyne.NewSize(15, 15), pin)

			go func(city string, icon *widget.Icon, status *widget.Label, btn *widget.Button, p *canvas.Circle, delayIdx int) {

				time.Sleep(time.Duration(delayIdx) * 1500 * time.Millisecond)

				latStr, lonStr, err := api.GetCoordinates(city)
				if err != nil {
					fyne.Do(func() { status.SetText(TR("loc_err")) })
					return
				}

				lat, _ := strconv.ParseFloat(latStr, 64)
				lon, _ := strconv.ParseFloat(lonStr, 64)
				tileURL := api.GetOSMTileURL(lat, lon, 12)

				client := &http.Client{Timeout: 10 * time.Second}
				req, _ := http.NewRequest("GET", tileURL, nil)
				req.Header.Set("User-Agent", "GroupieTracker-StudentProject/2.0 (education)")

				resp, errImg := client.Do(req)

				if errImg == nil && resp.StatusCode == 200 {
					data, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					res := fyne.NewStaticResource("map.png", data)

					fyne.Do(func() {
						icon.SetResource(res)
						status.Hide()
						p.Show()
						btn.SetText(TR("plan_btn"))
						btn.OnTapped = func() {
							u, _ := url.Parse(fmt.Sprintf("https://www.openstreetmap.org/?mlat=%s&mlon=%s", latStr, lonStr))
							app.OpenURL(u)
						}
					})
				} else {
					fyne.Do(func() { status.SetText(TR("map_err")) })
					if resp != nil {
						resp.Body.Close()
					}
				}
			}(locName, mapIcon, statusLbl, btnMap, pin, requestIndex)

			requestIndex++

			bgRect := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 50, A: 255})

			mapStack := container.NewMax(
				bgRect,
				container.NewPadded(mapIcon),
				container.NewCenter(pinWrapper),
				container.NewCenter(statusLbl),
			)

			mapContainer := container.NewGridWrap(fyne.NewSize(600, 350), mapStack)

			infoBox := container.NewVBox(
				canvas.NewText(":: "+locName, ColAccent),
				mapContainer,
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
		cardsContainer.Add(widget.NewLabel(TR("no_data")))
	}

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
	split.Offset = 0.35

	mainBg := canvas.NewRectangle(ColBackground)

	page := container.NewBorder(
		container.NewVBox(headerTop, widget.NewSeparator()),
		nil, nil, nil,
		split,
	)

	return container.NewMax(mainBg, page)
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

	content := container.NewVBox(container.NewCenter(iconW), valText, lblText)
	bg := canvas.NewRectangle(ColCard)

	return container.NewMax(bg, container.NewPadded(content))
}

func loadDetailImage(url string, w, h float32) fyne.CanvasObject {
	rect := canvas.NewRectangle(ColCard)
	placeholderWrapper := container.NewGridWrap(fyne.NewSize(w, h), rect)

	c := container.NewStack(placeholderWrapper)

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
