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
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"cryoutils-ng/core"
	"cryoutils-ng/i18n"
)

func sectionsSwap(e *core.Engine, w fyne.Window, lang *i18n.Lang, prog *widget.ProgressBar, progLabel *widget.Label, r *refreshers) fyne.CanvasObject {
	sizes := e.GetAvailableSwapSizesStr()
	currentGB := e.GetStatusSummary()["SwapSizeGB"]
	current := fmt.Sprintf("%s GB%s", currentGB, lang.T("swap.current"))

	selectSize := widget.NewSelect(sizes, func(s string) {})
	selectSize.SetSelected(current)

	sizeLabel := widget.NewLabel(fmt.Sprintf("%s: %s", lang.T("swap.swapSizeGB"), current))
	r.add(func() {
		// Only touch the select when the live value actually changed,
		// so the user's pending selection is never stomped.
		cur := fmt.Sprintf("%s GB%s", e.GetStatusSummary()["SwapSizeGB"], lang.T("swap.current"))
		sizeLabel.SetText(fmt.Sprintf("%s: %s", lang.T("swap.swapSizeGB"), cur))
		if cur != selectSize.Selected {
			selectSize.SetOptions(e.GetAvailableSwapSizesStr())
			selectSize.SetSelected(cur)
		}
	})

	var btn *widget.Button

	btn = widget.NewButton(lang.T("swap.resizeSwap"), func() {
		dialog.ShowConfirm(
			lang.T("native.resizeConfirmTitle"),
			lang.T("native.resizeConfirmMsg"),
			func(agree bool) {
				if !agree {
					return
				}
				if !withAuth(e, w, lang) {
					return
				}
				btn.Disable()
				prog.Show()
				progLabel.Show()
				progLabel.SetText(lang.T("swap.resizing"))
				go func() {
					size, _ := strconv.Atoi(strings.TrimSuffix(selectSize.Selected, " GB"))
					err := e.ChangeSwapSize(size)
					fyne.Do(func() {
						prog.Hide()
						progLabel.Hide()
						btn.Enable()
						if err != nil {
							dialog.ShowError(err, w)
						}
					})
				}()
			},
			nil,
		)
	})

	title := widget.NewLabel(lang.T("swap.title"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	zramNote := widget.NewLabel(lang.T("swap.zramNote"))
	zramNote.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		title,
		sizeLabel,
		btn,
		zramNote,
		widget.NewSeparator(),
	)
}
