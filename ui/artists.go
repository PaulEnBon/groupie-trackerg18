package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"groupie-tracker/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func ArtistList(app fyne.App, artists []models.Artist) fyne.CanvasObject {
	// Container principal qui va changer entre la liste et les détails
	mainContainer := container.NewStack()

	// Container pour les cartes d'artistes
	cardsContainer := container.NewVBox()

	// Fonction pour afficher les détails d'un artiste
	showDetails := func(artist models.Artist) {
		detailView := ArtistDetail(app, artist, func() {
			// Retour à la liste
			mainContainer.Objects = mainContainer.Objects[:1]
			mainContainer.Refresh()
		})
		mainContainer.Add(detailView)
		mainContainer.Refresh()
	}

	// Label de résumé du nombre d'items visibles
	summaryLabel := widget.NewLabel("")

	// Fonction pour créer les cartes d'artistes filtrées
	updateCards := func(searchText string) {
		cardsContainer.Objects = nil
		visible := 0

		for _, artist := range artists {
			currentArtist := artist // Capturer la valeur pour la closure
			// Filtre de recherche
			if searchText != "" {
				searchLower := strings.ToLower(searchText)
				nameMatch := strings.Contains(strings.ToLower(artist.Name), searchLower)
				membersMatch := false
				for _, member := range artist.Members {
					if strings.Contains(strings.ToLower(member), searchLower) {
						membersMatch = true
						break
					}
				}

				if !nameMatch && !membersMatch {
					continue
				}
			}

			// Création de la carte d'artiste (texte clair sur fond sombre)
			name := canvas.NewText(artist.Name, color.White)
			name.TextSize = 18
			name.TextStyle = fyne.TextStyle{Bold: true}
			name.Alignment = fyne.TextAlignCenter

			year := canvas.NewText(
				"Année de création: "+strconv.Itoa(artist.CreationDate),
				color.NRGBA{R: 235, G: 235, B: 235, A: 255},
			)
			year.Alignment = fyne.TextAlignCenter

			members := canvas.NewText(
				"Membres: "+strings.Join(artist.Members, ", "),
				color.NRGBA{R: 220, G: 220, B: 220, A: 255},
			)
			members.TextSize = 12
			members.Alignment = fyne.TextAlignCenter

			firstAlbum := canvas.NewText(
				"Premier album: "+artist.FirstAlbum,
				color.NRGBA{R: 220, G: 220, B: 220, A: 255},
			)
			firstAlbum.TextSize = 12
			firstAlbum.Alignment = fyne.TextAlignCenter

			// Bouton pour voir les détails
			detailsButton := widget.NewButton("👁️ Voir les détails", func() {
				showDetails(currentArtist)
			})
			buttonRow := container.NewHBox(layout.NewSpacer(), detailsButton, layout.NewSpacer())

			// Carte avec fond et padding
			cardBg := canvas.NewRectangle(color.NRGBA{R: 32, G: 32, B: 32, A: 255})
			cardBg.SetMinSize(fyne.NewSize(0, 150))
			cardContent := container.NewVBox(
				name,
				widget.NewSeparator(),
				year,
				members,
				firstAlbum,
				buttonRow,
			)
			card := container.NewMax(cardBg, container.NewPadded(cardContent))

			cardsContainer.Add(card)
			cardsContainer.Add(widget.NewSeparator())
			visible++
		}

		summaryLabel.SetText(fmt.Sprintf("%d groupes visibles", visible))
		cardsContainer.Refresh()
	}

	// Création de la barre de recherche
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("🔍 Rechercher un groupe ou un membre...")
	searchEntry.OnChanged = func(text string) {
		updateCards(text)
	}

	// Initialiser avec tous les artistes
	updateCards("")

	// Container principal avec la barre de recherche en haut
	scrollContent := container.NewVScroll(cardsContainer)
	// Fond opaque pour la liste
	listBackground := canvas.NewRectangle(color.NRGBA{R: 24, G: 24, B: 24, A: 240})
	listViewInner := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Groupie Tracker - Recherche de Groupes de Musique"),
			searchEntry,
			summaryLabel,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		scrollContent,
	)
	listView := container.NewMax(listBackground, listViewInner)

	// Ajouter la vue de liste au container principal
	mainContainer.Add(listView)

	return mainContainer
}
