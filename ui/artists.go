package ui

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"sort"
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

// Mode d'affichage
type ViewMode int

const (
	ModeList ViewMode = iota
	ModeGrid
)

func ArtistList(app fyne.App, win fyne.Window, artists []models.Artist) fyne.CanvasObject {
	currentMode := ModeList

	localArtists := append([]models.Artist(nil), artists...)

	artistLocations := make(map[int][]string)

	mainStack := container.NewStack()
	contentContainer := container.NewStack()

	bgRectangle := canvas.NewRectangle(ColBackground)

	var refreshContent func()

	// --- WIDGETS DYNAMIQUES ---
	title := canvas.NewText("", ColAccent) // Sera mis à jour par refreshContent
	btnAdd := widget.NewButtonWithIcon("", theme.ContentAddIcon(), nil)
	minCreationEntry := widget.NewEntry()
	maxCreationEntry := widget.NewEntry()
	minAlbumEntry := widget.NewEntry()
	maxAlbumEntry := widget.NewEntry()
	locationEntry := widget.NewEntry()
	searchEntry := widget.NewEntry()
	favOnlyCheck := widget.NewCheck("", nil)

	lblFav := widget.NewLabel("")
	lblCrea := widget.NewLabel("")
	lblAlbum := widget.NewLabel("")
	lblMembers := widget.NewLabel("")
	lblLoc := widget.NewLabel("")
	accordionItem := widget.NewAccordionItem("", nil)

	showDetails := func(artist models.Artist) {
		favorites := LoadFavorites()
		isFav := favorites[artist.ID]
		// NOTE : ArtistDetail reste partiellement en français car non traduit ici
		detailView := ArtistDetail(app, artist, isFav, func() {
			mainStack.Objects = mainStack.Objects[:1]
			mainStack.Refresh()
			refreshContent()
		}, func(newState bool) {
			favorites[artist.ID] = newState
			SaveFavorites(favorites)
			refreshContent()
		})
		mainStack.Add(detailView)
	}

	// --- HELPER IMAGE ---
	loadImage := func(url string, s float32) fyne.CanvasObject {
		rect := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 50, A: 255})
		rect.SetMinSize(fyne.NewSize(s, s))
		c := container.NewMax(rect)
		rect.SetMinSize(fyne.NewSize(s, s))

		go func() {
			if strings.HasPrefix(url, "file://") || (!strings.Contains(url, "://") && url != "") {
				path := strings.TrimPrefix(url, "file://")
				if len(path) > 2 && path[0] == '/' && path[2] == ':' {
					path = path[1:]
				}
				f, err := os.Open(path)
				if err == nil {
					defer f.Close()
					imgData, _, errDec := image.Decode(f)
					if errDec == nil {
						fyne.Do(func() {
							img := canvas.NewImageFromImage(imgData)
							img.FillMode = canvas.ImageFillContain
							img.SetMinSize(fyne.NewSize(s, s))
							c.Objects = []fyne.CanvasObject{img}
							c.Refresh()
						})
					}
				}
				return
			}
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode == 200 {
				defer resp.Body.Close()
				imgData, _, errDec := image.Decode(resp.Body)
				if errDec == nil {
					fyne.Do(func() {
						img := canvas.NewImageFromImage(imgData)
						img.FillMode = canvas.ImageFillContain
						img.SetMinSize(fyne.NewSize(s, s))
						c.Objects = []fyne.CanvasObject{img}
						c.Refresh()
					})
				}
			}
		}()
		return c
	}

	// --- BOUTON PARAMETRES ---
	btnSettings := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		ShowSettingsModal(app, win, func() {
			refreshContent()
		})
	})

	// Setup Bouton Ajouter
	btnAdd.OnTapped = func() {
		form := UserBandForm(app, win,
			func() {
				mainStack.Objects = mainStack.Objects[:1]
				mainStack.Refresh()
			},
			func(a models.Artist, rel map[string][]string) {
				maxID := 0
				for _, ar := range localArtists {
					if ar.ID > maxID {
						maxID = ar.ID
					}
				}
				a.ID = maxID + 1
				if strings.TrimSpace(a.Image) == "" {
					a.Image = "https://via.placeholder.com/300x300.png?text=Band"
				}
				if strings.TrimSpace(a.FirstAlbum) == "" {
					a.FirstAlbum = "01-01-2000"
				}
				localArtists = append(localArtists, a)
				locs := make([]string, 0, len(rel))
				for city := range rel {
					locs = append(locs, city)
				}
				artistLocations[a.ID] = locs
				mainStack.Objects = mainStack.Objects[:1]
				mainStack.Refresh()
				refreshContent()
			},
		)
		mainStack.Add(container.NewMax(canvas.NewRectangle(ColBackground), form))
	}
	btnAdd.Importance = widget.HighImportance

	sortOptions := []string{
		"Nom (A-Z)", "Nom (Z-A)",
		"Année Création (Récent)", "Année Création (Ancien)",
		"Premier Album (Récent)", "Premier Album (Ancien)",
	}
	sortSelect := widget.NewSelect(sortOptions, func(s string) { refreshContent() })
	sortSelect.Selected = "Nom (A-Z)"

	membersOptions := []string{"1", "2", "3", "4", "5", "6", "7", "8+"}
	membersCheckGroup := widget.NewCheckGroup(membersOptions, func(s []string) { refreshContent() })
	membersCheckGroup.Horizontal = true

	updateFilter := func(s string) { refreshContent() }
	minCreationEntry.OnChanged = updateFilter
	maxCreationEntry.OnChanged = updateFilter
	minAlbumEntry.OnChanged = updateFilter
	maxAlbumEntry.OnChanged = updateFilter
	locationEntry.OnChanged = updateFilter
	searchEntry.OnChanged = updateFilter
	favOnlyCheck.OnChanged = func(b bool) { refreshContent() }

	refreshContent = func() {
		title.Text = TR("app_title")
		title.Refresh()
		btnAdd.SetText(TR("btn_create"))
		searchEntry.SetPlaceHolder(TR("search_place"))
		sortSelect.PlaceHolder = TR("sort_place")
		favOnlyCheck.Text = TR("fav_only")
		favOnlyCheck.Refresh()
		lblFav.SetText(TR("fav_only"))
		lblCrea.SetText(TR("creation_date"))
		lblAlbum.SetText(TR("first_album"))
		lblMembers.SetText(TR("members"))
		lblLoc.SetText(TR("location"))
		accordionItem.Title = TR("filters")
		if accordionItem.Detail != nil {
			accordionItem.Detail.Refresh()
		}

		favorites := LoadFavorites()

		var filtered []models.Artist
		searchFilter := strings.ToLower(searchEntry.Text)
		locFilter := strings.ToLower(locationEntry.Text)
		minCreation, _ := strconv.Atoi(minCreationEntry.Text)
		maxCreation, _ := strconv.Atoi(maxCreationEntry.Text)
		if maxCreation == 0 {
			maxCreation = 3000
		}
		minAlbum, _ := strconv.Atoi(minAlbumEntry.Text)
		maxAlbum, _ := strconv.Atoi(maxAlbumEntry.Text)
		if maxAlbum == 0 {
			maxAlbum = 3000
		}

		selectedMembers := make(map[int]bool)
		for _, s := range membersCheckGroup.Selected {
			if s == "8+" {
				selectedMembers[8] = true
			} else {
				n, _ := strconv.Atoi(s)
				selectedMembers[n] = true
			}
		}

		for _, a := range localArtists {
			if favOnlyCheck.Checked && !favorites[a.ID] {
				continue
			}
			if searchFilter != "" {
				matchName := strings.Contains(strings.ToLower(a.Name), searchFilter)
				matchMember := false
				for _, m := range a.Members {
					if strings.Contains(strings.ToLower(m), searchFilter) {
						matchMember = true
						break
					}
				}
				if !matchName && !matchMember {
					continue
				}
			}
			if a.CreationDate < minCreation || a.CreationDate > maxCreation {
				continue
			}
			var albumYear int
			if len(a.FirstAlbum) >= 4 {
				yearStr := a.FirstAlbum[len(a.FirstAlbum)-4:]
				albumYear, _ = strconv.Atoi(yearStr)
			}
			if albumYear != 0 && (albumYear < minAlbum || albumYear > maxAlbum) {
				continue
			}
			if len(membersCheckGroup.Selected) > 0 {
				count := len(a.Members)
				match := false
				if selectedMembers[count] || (count >= 8 && selectedMembers[8]) {
					match = true
				}
				if !match {
					continue
				}
			}
			if locFilter != "" {
				locs, hasLocs := artistLocations[a.ID]
				if !hasLocs {
					continue
				}
				matchLoc := false
				for _, l := range locs {
					cleanLoc := strings.ReplaceAll(strings.ReplaceAll(l, "-", " "), "_", " ")
					if strings.Contains(strings.ToLower(cleanLoc), locFilter) {
						matchLoc = true
						break
					}
				}
				if !matchLoc {
					continue
				}
			}
			filtered = append(filtered, a)
		}

		sort.Slice(filtered, func(i, j int) bool {
			a, b := filtered[i], filtered[j]
			switch sortSelect.Selected {
			case "Nom (Z-A)":
				return strings.ToLower(a.Name) > strings.ToLower(b.Name)
			case "Année Création (Ancien)":
				return a.CreationDate < b.CreationDate
			case "Année Création (Récent)":
				return a.CreationDate > b.CreationDate
			case "Premier Album (Ancien)":
				da, _ := time.Parse("02-01-2006", a.FirstAlbum)
				db, _ := time.Parse("02-01-2006", b.FirstAlbum)
				return da.Before(db)
			case "Premier Album (Récent)":
				da, _ := time.Parse("02-01-2006", a.FirstAlbum)
				db, _ := time.Parse("02-01-2006", b.FirstAlbum)
				return da.After(db)
			default:
				return strings.ToLower(a.Name) < strings.ToLower(b.Name)
			}
		})

		var listObj fyne.CanvasObject

		if currentMode == ModeList {
			listBox := container.NewVBox()
			for _, artist := range filtered {
				cardBg := canvas.NewRectangle(ColCard)
				img := loadImage(artist.Image, 70)
				name := canvas.NewText(strings.ToUpper(artist.Name), ColAccent)
				name.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
				name.TextSize = 18

				txtMembres := TR("members")
				year := canvas.NewText(fmt.Sprintf("%d | %d %s", artist.CreationDate, len(artist.Members), txtMembres), ColText)
				year.TextSize = 12

				var favIcon fyne.CanvasObject
				if favorites[artist.ID] {
					favIcon = widget.NewIcon(theme.ConfirmIcon())
				} else {
					favIcon = layout.NewSpacer()
				}

				btn := widget.NewButton(TR("see_btn"), func() { showDetails(artist) })

				row := container.NewBorder(nil, nil,
					container.NewPadded(img),
					container.NewHBox(favIcon, btn),
					container.NewVBox(layout.NewSpacer(), name, year, layout.NewSpacer()),
				)
				wrapper := container.NewMax(cardBg, container.NewPadded(row))
				listBox.Add(wrapper)
				listBox.Add(widget.NewSeparator())
			}
			listObj = container.NewVScroll(container.NewPadded(listBox))
		} else {
			gridContainer := container.NewGridWithColumns(3)
			for _, artist := range filtered {
				cardBg := canvas.NewRectangle(ColCard)
				img := loadImage(artist.Image, 120)
				name := widget.NewLabel(strings.ToUpper(artist.Name))
				name.Alignment = fyne.TextAlignCenter
				name.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

				var favInd fyne.CanvasObject
				if favorites[artist.ID] {
					txt := canvas.NewText("★", ColHighlight)
					txt.TextSize = 20
					txt.Alignment = fyne.TextAlignCenter
					favInd = txt
				} else {
					favInd = layout.NewSpacer()
				}
				btn := widget.NewButton("", func() { showDetails(artist) })
				btn.Importance = widget.LowImportance

				content := container.NewVBox(container.NewPadded(img), name, favInd)
				card := container.NewMax(cardBg, container.NewPadded(content), btn)
				gridContainer.Add(card)
			}
			listObj = container.NewVScroll(container.NewPadded(gridContainer))
		}

		bgRectangle.FillColor = ColBackground
		bgRectangle.Refresh()
		contentContainer.Objects = []fyne.CanvasObject{listObj}
		contentContainer.Refresh()
	}

	go func() {
		locs, err := api.FetchAllLocationsMap()
		if err == nil {
			fyne.Do(func() {
				artistLocations = locs
				if locationEntry.Text != "" {
					refreshContent()
				}
			})
		}
	}()

	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	btnToggle := widget.NewButtonWithIcon("", theme.GridIcon(), nil)
	btnToggle.OnTapped = func() {
		if currentMode == ModeList {
			currentMode = ModeGrid
			btnToggle.SetIcon(theme.ListIcon())
		} else {
			currentMode = ModeList
			btnToggle.SetIcon(theme.GridIcon())
		}
		refreshContent()
	}

	topControl := container.NewBorder(nil, nil, title, container.NewHBox(btnAdd, btnSettings, btnToggle), nil)

	filtersForm := container.NewVBox(
		lblFav, favOnlyCheck,
		lblCrea, container.NewGridWithColumns(2, minCreationEntry, maxCreationEntry),
		lblAlbum, container.NewGridWithColumns(2, minAlbumEntry, maxAlbumEntry),
		lblMembers, membersCheckGroup,
		lblLoc, locationEntry,
	)

	accordionItem.Detail = filtersForm
	accordionItem.Title = "FILTRES"
	accordion := widget.NewAccordion(accordionItem)

	header := container.NewVBox(
		topControl,
		container.NewGridWithColumns(2, searchEntry, sortSelect),
		accordion,
		widget.NewSeparator(),
	)

	refreshContent()

	pageLayout := container.NewBorder(header, nil, nil, nil, contentContainer)
	mainStack.Add(container.NewMax(bgRectangle, pageLayout))

	return mainStack
}
