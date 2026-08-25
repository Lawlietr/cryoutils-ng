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
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func sectionsSwap(c *uiCtx) fyne.CanvasObject {
	lang := c.lang

	sizes := c.e.GetAvailableSwapSizesStr()
	selectSize := widget.NewSelect(sizes, func(string) {})
	selectSize.SetSelected(pickCurrentSwapSize(sizes))

	var btnResize *widget.Button
	btnResize = widget.NewButton(lang.T("swap.resizeSwap"), func() {
		dialog.ShowConfirm(
			lang.T("native.resizeConfirmTitle"),
			lang.T("native.resizeConfirmMsg"),
			func(agree bool) {
				if !agree {
					return
				}
				size := parseSwapSize(selectSize.Selected)
				btnResize.Disable()
				c.runTask(lang.T("swap.resizing"), func() error {
					return c.e.ChangeSwapSize(size)
				}, func(error) {
					btnResize.Enable()
					c.refreshStatus()
				})
			},
			c.win,
		)
	})

	swappinessEntry := widget.NewEntry()
	swappinessEntry.SetPlaceHolder(c.e.GetStatusSummary()["Swappiness"])
	var btnSwappiness *widget.Button
	btnSwappiness = widget.NewButton(lang.T("swap.apply"), func() {
		val := strings.TrimSpace(swappinessEntry.Text)
		btnSwappiness.Disable()
		c.runTask(lang.T("swap.setting"), func() error {
			return c.e.ChangeSwappiness(val)
		}, func(error) {
			btnSwappiness.Enable()
			c.refreshStatus()
		})
	})
	btnSwappiness.Disable()
	swappinessEntry.OnChanged = func(s string) {
		if _, err := strconv.Atoi(strings.TrimSpace(s)); err != nil {
			btnSwappiness.Disable()
		} else {
			btnSwappiness.Enable()
		}
	}

	title := widget.NewLabel(lang.T("swap.title"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	zramNote := widget.NewLabel(lang.T("swap.zramNote"))
	zramNote.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		title,
		widget.NewLabel(lang.T("swap.swapSizeGB")),
		selectSize,
		btnResize,
		zramNote,
		container.NewHBox(widget.NewLabel(lang.T("status.swappiness")), swappinessEntry, btnSwappiness),
		widget.NewSeparator(),
	)
}

// pickCurrentSwapSize finds the "Current Size" entry, falling back to Default.
func pickCurrentSwapSize(sizes []string) string {
	var fallback string
	for _, s := range sizes {
		if strings.Contains(s, "- Current Size") {
			return s
		}
		if strings.HasPrefix(s, "1 ") {
			fallback = s
		}
	}
	if fallback != "" {
		return fallback
	}
	if len(sizes) > 0 {
		return sizes[0]
	}
	return ""
}

// parseSwapSize extracts the GB number from entries like "1 - Default",
// "16 - Current Size" or plain "8".
func parseSwapSize(s string) int {
	if i := strings.Index(s, " -"); i >= 0 {
		s = s[:i]
	}
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
