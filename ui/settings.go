package ui

import (
	"encoding/json"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ShowSettingsModal affiche une modale de configuration
func ShowSettingsModal(app fyne.App, win fyne.Window, onRefresh func()) {

	// 1. SELECTEUR DE LANGUE
	langSelect := widget.NewSelect([]string{"Français", "English", "Español", "Deutsch"}, func(s string) {
		switch s {
		case "English":
			CurrentLang = "EN"
		case "Español":
			CurrentLang = "ES"
		case "Deutsch":
			CurrentLang = "DE"
		default:
			CurrentLang = "FR"
		}
	})

	// Sélection par défaut
	switch CurrentLang {
	case "EN":
		langSelect.Selected = "English"
	case "ES":
		langSelect.Selected = "Español"
	case "DE":
		langSelect.Selected = "Deutsch"
	default:
		langSelect.Selected = "Français"
	}

	// 2. SELECTEUR DE THEME
	themeSelect := widget.NewSelect([]string{TR("theme_dark"), TR("theme_light")}, func(s string) {
		if s == TR("theme_dark") {
			app.Settings().SetTheme(theme.DarkTheme())
		} else {
			app.Settings().SetTheme(theme.LightTheme())
		}
	})
	themeSelect.PlaceHolder = TR("theme_label")

	// 3. ACTIONS DE DONNÉES (IMPORT / EXPORT)

	// EXPORT
	btnExport := widget.NewButtonWithIcon(TR("btn_export"), theme.DownloadIcon(), func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			defer writer.Close()

			// On charge les favoris actuels
			favs := LoadFavorites()
			var ids []int
			for id, isFav := range favs {
				if isFav {
					ids = append(ids, id)
				}
			}

			// On écrit le JSON
			if json.NewEncoder(writer).Encode(ids) == nil {
				dialog.ShowInformation(TR("success_title"), TR("export_msg"), win)
			}
		}, win)
	})
	// Suggestion de nom de fichier par défaut
	btnExport.OnTapped = func() {
		d := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			defer writer.Close()
			favs := LoadFavorites()
			var ids []int
			for id, isFav := range favs {
				if isFav {
					ids = append(ids, id)
				}
			}
			if json.NewEncoder(writer).Encode(ids) == nil {
				dialog.ShowInformation(TR("success_title"), TR("export_msg"), win)
			}
		}, win)
		d.SetFileName("favorites_backup.json")
		d.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		d.Show()
	}

	// IMPORT
	btnImport := widget.NewButtonWithIcon(TR("btn_import"), theme.UploadIcon(), func() {
		d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			var ids []int
			if json.NewDecoder(reader).Decode(&ids) == nil {
				newFavs := make(map[int]bool)
				for _, id := range ids {
					newFavs[id] = true
				}
				SaveFavorites(newFavs)
				dialog.ShowInformation(TR("success_title"), TR("import_msg"), win)
				onRefresh() // Rafraîchir l'interface derrière
			}
		}, win)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		d.Show()
	})

	// 4. RESET (Zone Danger)
	btnResetFav := widget.NewButtonWithIcon(TR("bonus_clean"), theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Confirmation", TR("bonus_clean")+" ?", func(ok bool) {
			if ok {
				emptyFav := make(map[int]bool)
				SaveFavorites(emptyFav)
				dialog.ShowInformation(TR("success_title"), TR("bonus_clean_msg"), win)
				onRefresh()
			}
		}, win)
	})
	btnResetFav.Importance = widget.DangerImportance

	// 5. A PROPOS
	btnAbout := widget.NewButtonWithIcon(TR("btn_about"), theme.InfoIcon(), func() {
		dialog.ShowInformation(TR("btn_about"), TR("about_text"), win)
	})

	// CONSTRUCTION DU FORMULAIRE
	form := widget.NewForm(
		widget.NewFormItem(TR("lang_label"), langSelect),
		widget.NewFormItem(TR("theme_label"), themeSelect),
	)

	// GROUPE DONNÉES
	dataGroup := widget.NewCard("Data", "", container.NewVBox(
		btnExport,
		btnImport,
		widget.NewSeparator(),
		btnResetFav,
	))

	// CONTENEUR DE LA POPUP
	content := container.NewVBox(
		widget.NewLabelWithStyle(TR("settings_title"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		dataGroup,
		widget.NewSeparator(),
		btnAbout,
		widget.NewButton(TR("btn_close"), func() {
			// Fermeture auto via dialog
		}),
	)

	// Affichage via une CustomDialog
	d := dialog.NewCustom(TR("settings_title"), TR("btn_close"), content, win)
	d.SetOnClosed(func() {
		onRefresh()
	})
	d.Resize(fyne.NewSize(400, 500))
	d.Show()
}
