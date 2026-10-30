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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"cryoutils-ng/core"

	"cryoutils-ng/i18n"
)

func sectionsStatus(c *uiCtx) fyne.CanvasObject {
	sum := c.e.GetStatusSummary()
	c.lastStatus = sum
	lang := c.lang

	swapFile := widget.NewLabel(fmt.Sprintf("%s: %s", lang.T("status.swapFile"), sum["SwapFile"]))
	swapSize := widget.NewLabel(fmt.Sprintf("%s: %s GB%s", lang.T("status.swapSize"), sum["SwapSizeGB"], lang.T("swap.current")))
	zramSize := widget.NewLabel(fmt.Sprintf("%s: %s GB", lang.T("status.zramSize"), sum["ZramSizeGB"]))
	zramActive := widget.NewLabel(fmt.Sprintf("%s: %s", lang.T("status.zramActive"), formatEnabled(sum["ZramActive"], lang)))
	totalSwap := widget.NewLabel(fmt.Sprintf("%s: %s GB", lang.T("status.totalSwap"), sum["TotalSwapGB"]))
	swappiness := widget.NewLabel(fmt.Sprintf("%s: %s", lang.T("status.swappiness"), sum["Swappiness"]))
	vram := widget.NewLabel(fmt.Sprintf("%s: %s", lang.T("status.vram"), sum["VRAM"]))

	// The five memory parameters are shown as the raw value the kernel holds,
	// next to the checkboxes in the memory card that say whether that value is
	// the recommended one. Showing the value is what makes "recommended",
	// "stock default" and "could not read" distinguishable on screen.
	unitRows := []struct{ labelKey, valueKey string }{
		{"memoryParams.hugepages", "HugePagesValue"},
		{"memoryParams.shmem", "ShMemValue"},
		{"memoryParams.compaction_proactiveness", "CompactionProactivenessValue"},
		{"memoryParams.defrag", "DefragValue"},
		{"memoryParams.page_lock_unfairness", "PageLockUnfairnessValue"},
	}
	extra := make([]fyne.CanvasObject, 0, len(unitRows))
	for _, u := range unitRows {
		value := sum[u.valueKey]
		if value == core.ValueUnknown || value == "" {
			value = lang.T("status.unknown")
		}
		extra = append(extra, widget.NewLabel(fmt.Sprintf("%s: %s", lang.T(u.labelKey), value)))
	}

	// Two-column grid: key/value rows collapse into half as many lines (4K readability).
	grid := container.NewGridWithColumns(2,
		append([]fyne.CanvasObject{
			swapFile, swapSize,
			zramSize, zramActive, totalSwap,
			swappiness, vram,
		}, extra...)...,
	)

	return widget.NewCard(lang.T("status.title"), "", grid)
}

func formatEnabled(val string, lang *i18n.Lang) string {
	if val == "true" {
		return lang.T("status.enabled")
	}
	return lang.T("status.disabled")
}
