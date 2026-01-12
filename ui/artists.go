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

// loadImageFromURL fetches an image from a URL and returns a CanvasObject.
func loadImageFromURL(url string, w, h float32) fyne.CanvasObject {
	// Placeholder couleur sombre
	placeholderColor := color.NRGBA{R: 30, G: 25, B: 45, A: 255}

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		ph := canvas.NewRectangle(placeholderColor)
		ph.SetMinSize(fyne.NewSize(w, h))
		return ph
	}
	defer resp.Body.Close()

	imgDecoded, _, err := image.Decode(resp.Body)
	if err != nil {
		ph := canvas.NewRectangle(placeholderColor)
		ph.SetMinSize(fyne.NewSize(w, h))
		return ph
	}

	img := canvas.NewImageFromImage(imgDecoded)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(w, h))
	return img
}

func ArtistList(app fyne.App, artists []models.Artist) fyne.CanvasObject {
	// --- PALETTE DE COULEURS (Locales pour éviter les conflits) ---
	colBackground := color.NRGBA{R: 15, G: 10, B: 25, A: 255} // Violet très sombre
	colCard := color.NRGBA{R: 30, G: 25, B: 45, A: 255}       // Violet/Gris
	colAccent := color.NRGBA{R: 0, G: 255, B: 255, A: 255}    // Cyan Fluo
	colHighlight := color.NRGBA{R: 255, G: 0, B: 128, A: 255} // Rose Fluo
	colText := color.NRGBA{R: 240, G: 240, B: 255, A: 255}    // Blanc bleuté

	// Container principal (Stack pour gérer la navigation Liste <-> Détail)
	mainContainer := container.NewStack()

	// Container des cartes (VBox dans un Scroll)
	cardsContainer := container.NewVBox()

	// Label Résumé (Style Terminal)
	summaryLabel := canvas.NewText("", colHighlight)
	summaryLabel.TextSize = 12
	summaryLabel.TextStyle = fyne.TextStyle{Monospace: true}
	summaryLabel.Alignment = fyne.TextAlignTrailing

	// --- LOGIQUE DE NAVIGATION ---
	showDetails := func(artist models.Artist) {
		detailView := ArtistDetail(app, artist, func() {
			// Retour à la liste : on enlève la vue détail
			if len(mainContainer.Objects) > 1 {
				mainContainer.Objects = mainContainer.Objects[:1]
				mainContainer.Refresh()
			}
		})
		mainContainer.Add(detailView)
	}

	// --- MISE A JOUR DES CARTES ---
	updateCards := func(searchText string) {
		cardsContainer.Objects = nil
		visible := 0

		for _, artist := range artists {
			// Filtre de recherche
			if searchText != "" {
				searchLower := strings.ToLower(searchText)
				match := strings.Contains(strings.ToLower(artist.Name), searchLower)
				// Recherche aussi dans les membres
				for _, m := range artist.Members {
					if strings.Contains(strings.ToLower(m), searchLower) {
						match = true
						break
					}
				}
				// Recherche par date de création (ex: "1998")
				if strings.Contains(strconv.Itoa(artist.CreationDate), searchLower) {
					match = true
				}
				if !match {
					continue
				}
			}

			// Capture pour la closure du bouton
			currentArtist := artist

			// -- DESIGN DE LA CARTE ARTISTE --

			// 1. Avatar avec bordure Rose
			avatarImg := loadImageFromURL(artist.Image, 70, 70)
			avatarBorder := canvas.NewRectangle(color.Transparent)
			avatarBorder.StrokeColor = colHighlight
			avatarBorder.StrokeWidth = 1
			avatarComp := container.NewMax(avatarBorder, avatarImg)

			// 2. Infos (Nom + Année)
			nameTxt := canvas.NewText(strings.ToUpper(artist.Name), colAccent)
			nameTxt.TextSize = 18
			nameTxt.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

			creationTxt := canvas.NewText(fmt.Sprintf("EST. %d", artist.CreationDate), colText)
			creationTxt.TextSize = 12
			creationTxt.TextStyle = fyne.TextStyle{Monospace: true}

			infoBox := container.NewVBox(nameTxt, creationTxt)

			// 3. Bouton "VIEW" (CORRECTION : Utilisation de SearchIcon qui est standard)
			viewBtn := widget.NewButtonWithIcon("ACCESS", theme.SearchIcon(), func() {
				showDetails(currentArtist)
			})
			viewBtn.Importance = widget.HighImportance

			// Assemblage ligne
			rowContent := container.NewBorder(
				nil, nil,
				container.NewPadded(avatarComp), // Gauche
				viewBtn,                         // Droite
				container.NewVBox(layout.NewSpacer(), infoBox, layout.NewSpacer()), // Centre
			)

			// Fond de la carte
			cardBg := canvas.NewRectangle(colCard)

			// Bordure Néon en bas
			neonLine := canvas.NewRectangle(colAccent)
			neonLine.SetMinSize(fyne.NewSize(0, 1))

			// Carte finale
			cardFinal := container.NewVBox(
				container.NewMax(
					cardBg,
					container.NewPadded(rowContent),
				),
				neonLine, // Ligne séparatrice néon
			)

			cardsContainer.Add(cardFinal)
			cardsContainer.Add(layout.NewSpacer()) // Petit espace
			visible++
		}

		// Mise à jour du compteur
		summaryLabel.Text = fmt.Sprintf("> STATUS: %d UNITS FOUND", visible)
		summaryLabel.Refresh()

		cardsContainer.Refresh()
	}

	// --- BARRE DE RECHERCHE & HEADER ---

	// Titre de l'app
	appTitle := canvas.NewText("GROUPIE_TRACKER // DATABASE", colAccent)
	appTitle.TextSize = 20
	appTitle.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	appTitle.Alignment = fyne.TextAlignCenter

	// Champ de recherche
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("SEARCH_QUERY...")
	searchEntry.OnChanged = func(text string) {
		updateCards(text)
	}

	// Conteneur Header (Titre + Recherche)
	headerBox := container.NewVBox(
		container.NewPadded(appTitle),
		container.NewPadded(searchEntry),
		container.NewPadded(summaryLabel),
		widget.NewSeparator(),
	)

	// Init
	updateCards("")

	// Scroll View
	scrollContent := container.NewVScroll(container.NewPadded(cardsContainer))

	// Layout Liste complète
	listViewInner := container.NewBorder(
		headerBox,
		nil, nil, nil,
		scrollContent,
	)

	// Fond global
	bgRect := canvas.NewRectangle(colBackground)

	listView := container.NewMax(bgRect, listViewInner)

	mainContainer.Add(listView)

	return mainContainer
}
