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

func sectionsPresets(e *core.Engine, w fyne.Window, lang *i18n.Lang, prog *widget.ProgressBar, progLabel *widget.Label) fyne.CanvasObject {
	var btnRecommended, btnStock *widget.Button

	btnRecommended = widget.NewButton(lang.T("presets.recommended"), func() {
		dialog.ShowConfirm(
			lang.T("native.presetConfirmTitle"),
			lang.T("native.presetConfirmMsg"),
			func(agree bool) {
				if !agree {
					return
				}
				if !withAuth(e, w, lang) {
					return
				}
				btnRecommended.Disable()
				btnStock.Disable()
				prog.Show()
				progLabel.Show()
				progLabel.SetText(lang.T("presets.applying"))
				go func() {
					err := e.UseRecommendedSettings()
					fyne.Do(func() {
						prog.Hide()
						progLabel.Hide()
						btnRecommended.Enable()
						btnStock.Enable()
						if err != nil {
							dialog.ShowError(err, w)
						}
					})
				}()
			},
			nil,
		)
	})

	btnStock = widget.NewButton(lang.T("presets.stock"), func() {
		dialog.ShowConfirm(
			lang.T("native.presetConfirmTitle"),
			lang.T("native.presetConfirmMsg"),
			func(agree bool) {
				if !agree {
					return
				}
				if !withAuth(e, w, lang) {
					return
				}
				btnRecommended.Disable()
				btnStock.Disable()
				prog.Show()
				progLabel.Show()
				progLabel.SetText(lang.T("presets.applying"))
				go func() {
					err := e.UseStockSettings()
					fyne.Do(func() {
						prog.Hide()
						progLabel.Hide()
						btnRecommended.Enable()
						btnStock.Enable()
						if err != nil {
							dialog.ShowError(err, w)
						}
					})
				}()
			},
			nil,
		)
	})

	title := widget.NewLabel(lang.T("presets.title"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewVBox(
		title,
		container.NewHBox(btnRecommended, btnStock),
		widget.NewSeparator(),
	)
}
