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

	mainStack := container.NewStack()
	contentContainer := container.NewStack()

	var refreshContent func(string)

	// Navigation vers Détail
	showDetails := func(artist models.Artist) {
		isFav := favorites[artist.ID]
		detailView := ArtistDetail(app, artist, isFav, func() {
			// Callback Retour
			mainStack.Objects = mainStack.Objects[:1]
			mainStack.Refresh()
			refreshContent("") // Rafraîchir pour voir les changements de favoris
		}, func(newState bool) {
			// Callback Changement Favori
			favorites[artist.ID] = newState
			SaveFavorites(favorites)
		})
		mainStack.Add(detailView)
	}

	// Helper Image Robuste
	loadImage := func(url string, s float32) fyne.CanvasObject {
		// Placeholder (Carré gris)
		rect := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 50, A: 255})
		rect.SetMinSize(fyne.NewSize(s, s))

		// Container qui contiendra l'image
		c := container.NewMax(rect)
		// Important : On applique la taille au container via Layout si possible,
		// ou on laisse le layout parent gérer, mais ici on force le min size du contenu interne.
		rect.SetMinSize(fyne.NewSize(s, s))

		// Chargement Async
		go func() {
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode == 200 {
				defer resp.Body.Close()
				imgData, _, errDec := image.Decode(resp.Body)
				if errDec == nil {
					// Mise à jour UI sécurisée
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

	// Moteur de rendu
	refreshContent = func(filter string) {
		var filtered []models.Artist
		filterLow := strings.ToLower(filter)

		for _, a := range artists {
			match := false
			if filter == "" {
				match = true
			} else {
				if strings.Contains(strings.ToLower(a.Name), filterLow) {
					match = true
				}
				if strings.Contains(strconv.Itoa(a.CreationDate), filterLow) {
					match = true
				}
				for _, m := range a.Members {
					if strings.Contains(strings.ToLower(m), filterLow) {
						match = true
						break
					}
				}
			}
			if match {
				filtered = append(filtered, a)
			}
		}

		var listObj fyne.CanvasObject

		if currentMode == ModeList {
			// === MODE LISTE ===
			listBox := container.NewVBox()
			for _, artist := range filtered {
				cardBg := canvas.NewRectangle(color.NRGBA{R: 30, G: 25, B: 45, A: 255})

				// Avatar
				img := loadImage(artist.Image, 70)

				// Textes
				name := canvas.NewText(strings.ToUpper(artist.Name), ColAccent)
				name.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
				name.TextSize = 18

				year := canvas.NewText(fmt.Sprintf("EST. %d", artist.CreationDate), ColText)
				year.TextSize = 12

				// Icone Favori
				var favIcon fyne.CanvasObject
				if favorites[artist.ID] {
					favIcon = widget.NewIcon(theme.ConfirmIcon()) // Check
				} else {
					favIcon = layout.NewSpacer()
				}

				btn := widget.NewButton("VIEW", func() { showDetails(artist) })

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
			// === MODE GRILLE ===
			// GridWithColumns(3) crée 3 colonnes fixes
			gridContainer := container.NewGridWithColumns(3)

			for _, artist := range filtered {
				cardBg := canvas.NewRectangle(color.NRGBA{R: 30, G: 25, B: 45, A: 255})

				// Image (Plus grande)
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

				content := container.NewVBox(
					container.NewPadded(img),
					name,
					favInd,
				)

				// Stack: Fond -> Contenu -> Bouton Transparent (pour le clic)
				card := container.NewMax(cardBg, container.NewPadded(content), btn)
				gridContainer.Add(card)
			}
			listObj = container.NewVScroll(container.NewPadded(gridContainer))
		}

		contentContainer.Objects = []fyne.CanvasObject{listObj}
		contentContainer.Refresh()
	}

	// Header
	title := canvas.NewText("GROUPIE // DATABASE", ColAccent)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Rechercher...")
	searchEntry.OnChanged = func(s string) { refreshContent(s) }

	// Bouton Toggle Grid/List
	btnToggle := widget.NewButtonWithIcon("", theme.GridIcon(), nil)
	btnToggle.OnTapped = func() {
		if currentMode == ModeList {
			currentMode = ModeGrid
			btnToggle.SetIcon(theme.ListIcon())
		} else {
			currentMode = ModeList
			btnToggle.SetIcon(theme.GridIcon())
		}
		refreshContent(searchEntry.Text)
	}

	header := container.NewVBox(
		container.NewBorder(nil, nil, title, btnToggle, nil),
		container.NewPadded(searchEntry),
		widget.NewSeparator(),
	)

	refreshContent("")

	pageLayout := container.NewBorder(header, nil, nil, nil, contentContainer)
	bg := canvas.NewRectangle(ColBackground)

	mainStack.Add(container.NewMax(bg, pageLayout))

	return mainStack
}
