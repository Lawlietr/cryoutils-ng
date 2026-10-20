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

	zramNote := widget.NewLabel(lang.T("swap.zramNote"))
	zramNote.TextStyle = fyne.TextStyle{Italic: true}

	// Unlike the status card, this card is built once, so the dropdown used to keep
	// the "Current Size" entry that was read at launch: after a resize the label
	// still showed the old size until the app was restarted. The options list is
	// also filtered by the free space in /home, so it goes stale for the same
	// reason. Refresh it in place, keeping an explicit choice the user made from
	// the list (see nextSwapSelection).
	c.addRefresher(func() {
		fresh := c.e.GetAvailableSwapSizesStr()
		keep := nextSwapSelection(selectSize.Selected, fresh)
		selectSize.SetOptions(fresh)
		selectSize.SetSelected(keep)

		// The swappiness placeholder mirrors the live value only while the user has
		// not typed anything.
		if strings.TrimSpace(swappinessEntry.Text) == "" && c.lastStatus != nil {
			swappinessEntry.SetPlaceHolder(c.lastStatus["Swappiness"])
		}
	})

	return widget.NewCard(lang.T("swap.title"), "", container.NewVBox(
		widget.NewLabel(lang.T("swap.swapSizeGB")),
		selectSize,
		btnResize,
		zramNote,
		container.NewHBox(widget.NewLabel(lang.T("status.swappiness")), swappinessEntry, btnSwappiness),
	))
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

// nextSwapSelection decides what the size dropdown shows after a refresh. An
// explicit entry the user chose from the list is kept as long as it is still
// offered; the "N - Current Size" pseudo entry must not be kept verbatim, because
// its whole meaning is the size that was current when the list was built.
func nextSwapSelection(previous string, options []string) string {
	if previous != "" && !strings.Contains(previous, "- Current Size") {
		for _, option := range options {
			if option == previous {
				return previous
			}
		}
	}
	return pickCurrentSwapSize(options)
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
