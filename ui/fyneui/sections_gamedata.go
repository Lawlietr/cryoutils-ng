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
)

func sectionsGameData(c *uiCtx) fyne.CanvasObject {
	lang := c.lang

	libraries, _ := c.e.FindDataFolders()
	options := append([]string{lang.T("gamedata.select")}, libraryPaths(libraries)...)

	ssdSelect := widget.NewSelect(options, func(string) {})
	extSelect := widget.NewSelect(options, func(string) {})
	if len(options) > 1 {
		ssdSelect.SetSelected(options[1])
	}
	if len(options) > 2 {
		extSelect.SetSelected(options[2])
	}

	selectionOK := func() (string, string, bool) {
		placeholder := lang.T("gamedata.select")
		ssd, ext := ssdSelect.Selected, extSelect.Selected
		if ssd == "" || ssd == placeholder || ext == "" || ext == placeholder {
			return "", "", false
		}
		return ssd, ext, true
	}

	var btnSync, btnCleanup *widget.Button
	btnSync = widget.NewButton(lang.T("gamedata.syncGameData"), func() {
		dialog.ShowConfirm(
			lang.T("native.syncConfirmTitle"),
			lang.T("native.syncConfirmMsg"),
			func(agree bool) {
				if !agree {
					return
				}
				ssd, ext, ok := selectionOK()
				if !ok {
					dialog.ShowInformation(lang.T("gamedata.title"), lang.T("gamedata.select"), c.win)
					return
				}
				btnSync.Disable()
				btnCleanup.Disable()
				c.runTask(lang.T("gamedata.syncing"), func() error {
					data, err := c.e.GetDataToMove(ssd, ext)
					if err != nil {
						return err
					}
					return c.e.MoveGameData(ssd, ext, data)
				}, func(error) {
					btnSync.Enable()
					btnCleanup.Enable()
				})
			},
			c.win,
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
				ssd, ext, ok := selectionOK()
				if !ok {
					dialog.ShowInformation(lang.T("gamedata.title"), lang.T("gamedata.select"), c.win)
					return
				}
				btnSync.Disable()
				btnCleanup.Disable()
				c.runTask(lang.T("gamedata.cleaning"), func() error {
					data, err := c.e.GetUninstalledGamesData(ssd, ext)
					if err != nil {
						return err
					}
					c.e.RemoveGameData(data.GetRight(), []string{ext})
					return nil
				}, func(error) {
					btnSync.Enable()
					btnCleanup.Enable()
				})
			},
			c.win,
		)
	})

	return widget.NewCard(lang.T("gamedata.title"), "", container.NewVBox(
		widget.NewLabel(lang.T("gamedata.ssdLibrary")),
		ssdSelect,
		widget.NewLabel(lang.T("gamedata.externalLibrary")),
		extSelect,
		container.NewHBox(btnSync, btnCleanup),
	))
}

func libraryPaths(libs []core.Library) []string {
	var paths []string
	for _, l := range libs {
		paths = append(paths, l.Path)
	}
	return paths
}
