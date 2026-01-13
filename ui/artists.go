package ui

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"
	"strings"

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

func ArtistList(app fyne.App, artists []models.Artist) fyne.CanvasObject {
	// --- ETAT LOCAL ---
	currentMode := ModeList
	favorites := LoadFavorites()

	// Map des localisations (chargée en arrière-plan)
	artistLocations := make(map[int][]string)

	mainStack := container.NewStack()
	contentContainer := container.NewStack()

	var refreshContent func()

	// --- NAVIGATION ---
	showDetails := func(artist models.Artist) {
		isFav := favorites[artist.ID]
		detailView := ArtistDetail(app, artist, isFav, func() {
			mainStack.Objects = mainStack.Objects[:1]
			mainStack.Refresh()
			refreshContent() // Rafraîchir au retour
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

	// --- WIDGETS FILTRES ---

	// 1. Creation Date (Range Filter)
	minCreationEntry := widget.NewEntry()
	minCreationEntry.SetPlaceHolder("Année Min (e.g 1990)")
	maxCreationEntry := widget.NewEntry()
	maxCreationEntry.SetPlaceHolder("Année Max (e.g 2010)")

	// 2. First Album Date (Range Filter)
	minAlbumEntry := widget.NewEntry()
	minAlbumEntry.SetPlaceHolder("Année De Premier Album Min")
	maxAlbumEntry := widget.NewEntry()
	maxAlbumEntry.SetPlaceHolder("Année De Premier Album Max")

	// 3. Number of Members (Checkbox Filter)
	// On utilise CheckGroup pour la sélection multiple
	membersOptions := []string{"1", "2", "3", "4", "5", "6", "7", "8+"}
	membersCheckGroup := widget.NewCheckGroup(membersOptions, func(s []string) { refreshContent() })
	membersCheckGroup.Horizontal = true

	// 4. Location (Search Filter)
	locationEntry := widget.NewEntry()
	locationEntry.SetPlaceHolder("Ville ou Pays...")

	// Barre de recherche globale (Nom)
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Recherche par nom d'artiste...")

	// Callbacks de mise à jour automatique
	updateFilter := func(s string) { refreshContent() }
	minCreationEntry.OnChanged = updateFilter
	maxCreationEntry.OnChanged = updateFilter
	minAlbumEntry.OnChanged = updateFilter
	maxAlbumEntry.OnChanged = updateFilter
	locationEntry.OnChanged = updateFilter
	searchEntry.OnChanged = updateFilter

	// --- LOGIQUE DE FILTRAGE ---
	refreshContent = func() {
		var filtered []models.Artist

		nameFilter := strings.ToLower(searchEntry.Text)
		locFilter := strings.ToLower(locationEntry.Text)

		// Parsing des dates (0 si vide)
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

		// Map des membres sélectionnés pour accès rapide
		selectedMembers := make(map[int]bool)
		for _, s := range membersCheckGroup.Selected {
			if s == "8+" {
				selectedMembers[8] = true // Cas spécial > 8
			} else {
				n, _ := strconv.Atoi(s)
				selectedMembers[n] = true
			}
		}

		for _, a := range artists {
			// 1. Filtre Nom
			if nameFilter != "" && !strings.Contains(strings.ToLower(a.Name), nameFilter) {
				continue
			}

			// 2. Filtre Creation Date (Range)
			if a.CreationDate < minCreation || a.CreationDate > maxCreation {
				continue
			}

			// 3. Filtre First Album (Range Year)
			// Format "dd-mm-yyyy", on prend les 4 derniers caractères
			if len(a.FirstAlbum) >= 4 {
				yearStr := a.FirstAlbum[len(a.FirstAlbum)-4:]
				year, _ := strconv.Atoi(yearStr)
				if year < minAlbum || year > maxAlbum {
					continue
				}
			}

			// 4. Filtre Members (Check Box)
			if len(membersCheckGroup.Selected) > 0 {
				count := len(a.Members)
				matchMember := false

				// Vérification exacte
				if selectedMembers[count] {
					matchMember = true
				}
				// Gestion du cas "8+"
				if count >= 8 && selectedMembers[8] {
					matchMember = true
				}

				if !matchMember {
					continue
				}
			}

			// 5. Filtre Location
			if locFilter != "" {
				locs, hasLocs := artistLocations[a.ID]
				if !hasLocs {
					// Si les locations ne sont pas encore chargées, on ignore ou on attend
					// Ici on exclut par défaut si pas chargé
					continue
				}
				matchLoc := false
				for _, l := range locs {
					// L'API renvoie souvent "paris-france", on remplace pour la recherche
					cleanLoc := strings.ReplaceAll(l, "-", " ")
					cleanLoc = strings.ReplaceAll(cleanLoc, "_", " ")
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

		// --- CONSTRUCTION UI ---
		var listObj fyne.CanvasObject

		if currentMode == ModeList {
			listBox := container.NewVBox()
			for _, artist := range filtered {
				cardBg := canvas.NewRectangle(color.NRGBA{R: 30, G: 25, B: 45, A: 255})
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
				cardBg := canvas.NewRectangle(color.NRGBA{R: 30, G: 25, B: 45, A: 255})
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

	// --- CHARGEMENT ASYNC DES LOCATIONS ---
	go func() {
		locs, err := api.FetchAllLocationsMap()
		if err == nil {
			fyne.Do(func() {
				artistLocations = locs
				// On rafraîchit si l'utilisateur avait déjà tapé une localisation
				if locationEntry.Text != "" {
					refreshContent()
				}
			})
		}
	}()

	// --- HEADER LAYOUT ---
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

	// Formulaire de filtres (Accordion pour gagner de la place)
	filtersForm := container.NewVBox(
		widget.NewLabel("Date de Création (Plage d'années)"),
		container.NewGridWithColumns(2, minCreationEntry, maxCreationEntry),

		widget.NewLabel("Date du Premier Album (Plage d'années)"),
		container.NewGridWithColumns(2, minAlbumEntry, maxAlbumEntry),

		widget.NewLabel("Nombre de Membres"),
		membersCheckGroup,

		widget.NewLabel("Lieu du Concert"),
		locationEntry,
	)

	accordion := widget.NewAccordion(
		widget.NewAccordionItem("FILTRES AVANCÉS", filtersForm),
	)

	header := container.NewVBox(
		container.NewBorder(nil, nil, title, btnToggle, nil),
		container.NewPadded(searchEntry),
		accordion,
		widget.NewSeparator(),
	)

	refreshContent()

	pageLayout := container.NewBorder(header, nil, nil, nil, contentContainer)
	bg := canvas.NewRectangle(ColBackground)
	mainStack.Add(container.NewMax(bg, pageLayout))

	return mainStack
}
