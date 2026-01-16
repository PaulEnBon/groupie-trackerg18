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
	// --- ETAT LOCAL ---
	currentMode := ModeList
	favorites := LoadFavorites()
	isDarkTheme := true

	// Copie modifiable pour intégrer les groupes créés par l'utilisateur
	localArtists := append([]models.Artist(nil), artists...)

	// Map des localisations
	artistLocations := make(map[int][]string)

	mainStack := container.NewStack()
	contentContainer := container.NewStack()

	bgRectangle := canvas.NewRectangle(ColBackground)

	var refreshContent func()

	// --- NAVIGATION ---
	showDetails := func(artist models.Artist) {
		isFav := favorites[artist.ID]
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
			// Supporte URL distante et fichiers locaux (file:// ou chemin brut)
			if strings.HasPrefix(url, "file://") || (!strings.Contains(url, "://") && url != "") {
				path := strings.TrimPrefix(url, "file://")
				if len(path) > 2 && path[0] == '/' && path[2] == ':' { // windows file URI -> drop leading slash
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

	// --- WIDGETS CONTROLE ---

	// Bouton Thème
	btnTheme := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), nil)
	btnTheme.OnTapped = func() {
		isDarkTheme = !isDarkTheme
		if isDarkTheme {
			ColBackground = color.NRGBA{R: 15, G: 10, B: 25, A: 255}
			ColCard = color.NRGBA{R: 30, G: 25, B: 45, A: 255}
			ColAccent = color.NRGBA{R: 0, G: 255, B: 255, A: 255}
			ColText = color.NRGBA{R: 240, G: 240, B: 255, A: 255}
			app.Settings().SetTheme(theme.DarkTheme())
		} else {
			ColBackground = color.NRGBA{R: 240, G: 240, B: 245, A: 255}
			ColCard = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			ColAccent = color.NRGBA{R: 0, G: 100, B: 200, A: 255}
			ColText = color.NRGBA{R: 40, G: 40, B: 50, A: 255}
			app.Settings().SetTheme(theme.LightTheme())
		}
		bgRectangle.FillColor = ColBackground
		bgRectangle.Refresh()
		refreshContent()
	}

	// Bouton ajout de groupe utilisateur
	btnAdd := widget.NewButtonWithIcon("Créer un groupe", theme.ContentAddIcon(), func() {
		form := UserBandForm(app, win,
			func() { // onBack
				mainStack.Objects = mainStack.Objects[:1]
				mainStack.Refresh()
			},
			func(a models.Artist, rel map[string][]string) { // onSave
				// Assigner un ID unique
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

				// Retour à la liste et rafraîchissement
				mainStack.Objects = mainStack.Objects[:1]
				mainStack.Refresh()
				refreshContent()
			},
		)
		mainStack.Add(container.NewMax(canvas.NewRectangle(ColBackground), form))
	})
	btnAdd.Importance = widget.HighImportance

	// Menu de Tri
	sortOptions := []string{
		"Nom (A-Z)", "Nom (Z-A)",
		"Année Création (Récent)", "Année Création (Ancien)",
		"Premier Album (Récent)", "Premier Album (Ancien)",
	}
	sortSelect := widget.NewSelect(sortOptions, func(s string) { refreshContent() })
	sortSelect.PlaceHolder = "Trier par..."
	sortSelect.Selected = "Nom (A-Z)"

	// --- FILTRES ---
	minCreationEntry := widget.NewEntry()
	minCreationEntry.SetPlaceHolder("Année Min")
	maxCreationEntry := widget.NewEntry()
	maxCreationEntry.SetPlaceHolder("Année Max")

	minAlbumEntry := widget.NewEntry()
	minAlbumEntry.SetPlaceHolder("Min Album")
	maxAlbumEntry := widget.NewEntry()
	maxAlbumEntry.SetPlaceHolder("Max Album")

	membersOptions := []string{"1", "2", "3", "4", "5", "6", "7", "8+"}
	membersCheckGroup := widget.NewCheckGroup(membersOptions, func(s []string) { refreshContent() })
	membersCheckGroup.Horizontal = true

	locationEntry := widget.NewEntry()
	locationEntry.SetPlaceHolder("Ville ou Pays...")

	favOnlyCheck := widget.NewCheck("Afficher seulement les favoris", func(b bool) { refreshContent() })

	searchEntry := widget.NewEntry()
	// Mise à jour du placeholder pour indiquer la nouvelle fonctionnalité
	searchEntry.SetPlaceHolder("Recherche (Groupe ou Membre)...")

	// Callbacks Auto-refresh
	updateFilter := func(s string) { refreshContent() }
	minCreationEntry.OnChanged = updateFilter
	maxCreationEntry.OnChanged = updateFilter
	minAlbumEntry.OnChanged = updateFilter
	maxAlbumEntry.OnChanged = updateFilter
	locationEntry.OnChanged = updateFilter
	searchEntry.OnChanged = updateFilter

	// --- LOGIQUE DE FILTRAGE ET TRI ---
	refreshContent = func() {
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

		// 1. FILTRAGE
		for _, a := range localArtists {
			if favOnlyCheck.Checked && !favorites[a.ID] {
				continue
			}

			// --- MODIFICATION ICI : RECHERCHE PAR NOM OU MEMBRE ---
			if searchFilter != "" {
				matchName := strings.Contains(strings.ToLower(a.Name), searchFilter)
				matchMember := false
				// On boucle sur chaque membre pour voir si le texte correspond
				for _, m := range a.Members {
					if strings.Contains(strings.ToLower(m), searchFilter) {
						matchMember = true
						break
					}
				}
				// Si ni le nom du groupe ni aucun membre ne correspond, on passe
				if !matchName && !matchMember {
					continue
				}
			}
			// -----------------------------------------------------

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

		// 2. TRI (SORTING)
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
			default: // "Nom (A-Z)"
				return strings.ToLower(a.Name) < strings.ToLower(b.Name)
			}
		})

		// 3. CONSTRUCTION UI
		var listObj fyne.CanvasObject

		if currentMode == ModeList {
			listBox := container.NewVBox()
			for _, artist := range filtered {
				cardBg := canvas.NewRectangle(ColCard)

				img := loadImage(artist.Image, 70)
				name := canvas.NewText(strings.ToUpper(artist.Name), ColAccent)
				name.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
				name.TextSize = 18

				year := canvas.NewText(fmt.Sprintf("%d | %d Membres", artist.CreationDate, len(artist.Members)), ColText)
				year.TextSize = 12

				var favIcon fyne.CanvasObject
				if favorites[artist.ID] {
					favIcon = widget.NewIcon(theme.ConfirmIcon())
				} else {
					favIcon = layout.NewSpacer()
				}
				btn := widget.NewButton("VOIR", func() { showDetails(artist) })

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

		contentContainer.Objects = []fyne.CanvasObject{listObj}
		contentContainer.Refresh()
	}

	// --- ASYNC DATA ---
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

	// --- HEADER UI ---
	title := canvas.NewText("GROUPIE // DATABASE", ColAccent)
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

	topControl := container.NewBorder(nil, nil, title, container.NewHBox(btnAdd, btnTheme, btnToggle), nil)

	filtersForm := container.NewVBox(
		widget.NewLabel("Favoris"),
		favOnlyCheck,
		widget.NewLabel("Date de Création"),
		container.NewGridWithColumns(2, minCreationEntry, maxCreationEntry),
		widget.NewLabel("Premier Album"),
		container.NewGridWithColumns(2, minAlbumEntry, maxAlbumEntry),
		widget.NewLabel("Membres"),
		membersCheckGroup,
		widget.NewLabel("Lieu"),
		locationEntry,
	)
	accordion := widget.NewAccordion(widget.NewAccordionItem("FILTRES AVANCÉS", filtersForm))

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
