// CryoUtils NG
// Copyright (C) 2025 CryoUtils NG contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"cryoutils-ng/core"
	"cryoutils-ng/i18n"
)

func sectionsGameData(e *core.Engine, w fyne.Window, lang *i18n.Lang, prog *widget.ProgressBar, progLabel *widget.Label, r *refreshers) fyne.CanvasObject {
	libraries, _ := e.FindDataFolders()
	options := append([]string{lang.T("gamedata.select")}, libraryPaths(libraries)...)

	var ssdIdx, extIdx int
	if len(options) > 1 {
		ssdIdx = 1
	}
	if len(options) > 2 {
		extIdx = 2
	}

	ssdSelect := widget.NewSelect(options, func(string) {})
	ssdSelect.SetSelected(options[ssdIdx])
	extSelect := widget.NewSelect(options, func(string) {})
	extSelect.SetSelected(options[extIdx])

	// Re-read libraries in place; keep the user's current picks if they
	// still exist in the new list.
	r.add(func() {
		libs, _ := e.FindDataFolders()
		newOpts := append([]string{lang.T("gamedata.select")}, libraryPaths(libs)...)
		if equalOptions(ssdSelect.Options, newOpts) {
			return
		}
		keepSSD, keepExt := ssdSelect.Selected, extSelect.Selected
		ssdSelect.SetOptions(newOpts)
		extSelect.SetOptions(newOpts)
		if containsString(newOpts, keepSSD) {
			ssdSelect.SetSelected(keepSSD)
		}
		if containsString(newOpts, keepExt) {
			extSelect.SetSelected(keepExt)
		}
	})

	var btnSync, btnCleanup *widget.Button

	btnSync = widget.NewButton(lang.T("gamedata.syncGameData"), func() {
		dialog.ShowConfirm(
			lang.T("native.syncConfirmTitle"),
			lang.T("native.syncConfirmMsg"),
			func(agree bool) {
				if !agree {
					return
				}
				if !withAuth(e, w, lang) {
					return
				}
				btnSync.Disable()
				btnCleanup.Disable()
				prog.Show()
				progLabel.Show()
				progLabel.SetText(lang.T("gamedata.syncing"))
				go func() {
					ssd := ssdSelect.Selected
					ext := extSelect.Selected
					if ssd == lang.T("gamedata.select") || ext == lang.T("gamedata.select") {
						fyne.Do(func() {
							prog.Hide()
							progLabel.Hide()
							btnSync.Enable()
							btnCleanup.Enable()
						})
						return
					}
					data, err := e.GetDataToMove(ssd, ext)
					if err != nil {
						fyne.Do(func() {
							prog.Hide()
							progLabel.Hide()
							btnSync.Enable()
							btnCleanup.Enable()
							dialog.ShowError(err, w)
						})
						return
					}
					if err := e.MoveGameData(ssd, ext, data); err != nil {
						fyne.Do(func() {
							prog.Hide()
							progLabel.Hide()
							btnSync.Enable()
							btnCleanup.Enable()
							dialog.ShowError(err, w)
						})
						return
					}
					fyne.Do(func() {
						prog.Hide()
						progLabel.Hide()
						btnSync.Enable()
						btnCleanup.Enable()
					})
				}()
			},
			nil,
		)
	})

	btnCleanup = widget.NewButton(lang.T("gamedata.cleanupOrphanedData"), func() {
		dialog.ShowConfirm(
			lang.T("native.cleanupConfirmTitle"),
			lang.T("native.cleanupConfirmMsg"),
			func(agree bool) {
				if !agree {
					return
				}
				if !withAuth(e, w, lang) {
					return
				}
				btnSync.Disable()
				btnCleanup.Disable()
				prog.Show()
				progLabel.Show()
				progLabel.SetText(lang.T("gamedata.cleaning"))
				go func() {
					ssd := ssdSelect.Selected
					ext := extSelect.Selected
					if ssd == lang.T("gamedata.select") || ext == lang.T("gamedata.select") {
						fyne.Do(func() {
							prog.Hide()
							progLabel.Hide()
							btnSync.Enable()
							btnCleanup.Enable()
						})
						return
					}
					data, err := e.GetUninstalledGamesData(ssd, ext)
					if err != nil {
						fyne.Do(func() {
							prog.Hide()
							progLabel.Hide()
							btnSync.Enable()
							btnCleanup.Enable()
							dialog.ShowError(err, w)
						})
						return
					}
					e.RemoveGameData(data.GetRight(), []string{ext})
					fyne.Do(func() {
						prog.Hide()
						progLabel.Hide()
						btnSync.Enable()
						btnCleanup.Enable()
					})
				}()
			},
			nil,
		)
	})

	title := widget.NewLabel(lang.T("gamedata.title"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewVBox(
		title,
		widget.NewLabel(lang.T("gamedata.ssdLibrary")),
		ssdSelect,
		widget.NewLabel(lang.T("gamedata.externalLibrary")),
		extSelect,
		container.NewHBox(btnSync, btnCleanup),
	)
}

func libraryPaths(libs []core.Library) []string {
	var paths []string
	for _, l := range libs {
		paths = append(paths, l.Path)
	}
	return paths
}

func equalOptions(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
