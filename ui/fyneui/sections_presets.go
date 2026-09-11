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
)

func sectionsPresets(c *uiCtx) fyne.CanvasObject {
	lang := c.lang

	var btnRecommended, btnStock *widget.Button
	btnRecommended = widget.NewButton(lang.T("presets.recommended"), func() {
		dialog.ShowConfirm(
			lang.T("native.presetConfirmTitle"),
			lang.T("native.presetConfirmMsg"),
			func(agree bool) {
				if !agree {
					return
				}
				btnRecommended.Disable()
				btnStock.Disable()
				c.runTask(lang.T("presets.applying"), c.e.UseRecommendedSettings, func(error) {
					btnRecommended.Enable()
					btnStock.Enable()
					c.refreshStatus()
				})
			},
			c.win,
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
				btnRecommended.Disable()
				btnStock.Disable()
				c.runTask(lang.T("presets.applying"), c.e.UseStockSettings, func(error) {
					btnRecommended.Enable()
					btnStock.Enable()
					c.refreshStatus()
				})
			},
			c.win,
		)
	})

	return widget.NewCard(lang.T("presets.title"), "", container.NewHBox(btnRecommended, btnStock))
}
