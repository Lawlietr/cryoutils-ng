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

func sectionsStatus(e *core.Engine, lang *i18n.Lang, r *refreshers) fyne.CanvasObject {
	sum := e.GetStatusSummary()

	swapFile := widget.NewLabel(fmt.Sprintf("%s: %s", lang.T("status.swapFile"), sum["SwapFile"]))
	swapSize := widget.NewLabel(fmt.Sprintf("%s: %s GB%s", lang.T("status.swapSize"), sum["SwapSizeGB"], lang.T("swap.current")))
	zramSize := widget.NewLabel(fmt.Sprintf("%s: %s GB", lang.T("status.zramSize"), sum["ZramSizeGB"]))
	zramActive := widget.NewLabel(fmt.Sprintf("%s: %s", lang.T("status.zramActive"), formatZramActive(sum["ZramActive"], lang)))
	totalSwap := widget.NewLabel(fmt.Sprintf("%s: %s GB", lang.T("status.totalSwap"), sum["TotalSwapGB"]))
	swappiness := widget.NewLabel(fmt.Sprintf("%s: %s", lang.T("status.swappiness"), formatSwappiness(sum["Swappiness"], lang)))
	vram := widget.NewLabel(fmt.Sprintf("%s: %s", lang.T("status.vram"), sum["VRAM"]))

	update := func() {
		sum = e.GetStatusSummary()
		swapFile.SetText(fmt.Sprintf("%s: %s", lang.T("status.swapFile"), sum["SwapFile"]))
		swapSize.SetText(fmt.Sprintf("%s: %s GB%s", lang.T("status.swapSize"), sum["SwapSizeGB"], lang.T("swap.current")))
		zramSize.SetText(fmt.Sprintf("%s: %s GB", lang.T("status.zramSize"), sum["ZramSizeGB"]))
		zramActive.SetText(fmt.Sprintf("%s: %s", lang.T("status.zramActive"), formatZramActive(sum["ZramActive"], lang)))
		totalSwap.SetText(fmt.Sprintf("%s: %s GB", lang.T("status.totalSwap"), sum["TotalSwapGB"]))
		swappiness.SetText(fmt.Sprintf("%s: %s", lang.T("status.swappiness"), formatSwappiness(sum["Swappiness"], lang)))
		vram.SetText(fmt.Sprintf("%s: %s", lang.T("status.vram"), sum["VRAM"]))
	}
	r.add(update)

	title := widget.NewLabel(lang.T("status.title"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewVBox(
		title,
		swapFile, swapSize,
		zramSize, zramActive, totalSwap,
		swappiness, vram,
		widget.NewSeparator(),
	)
}

func formatZramActive(val string, lang *i18n.Lang) string {
	if val == "true" {
		return lang.T("status.enabled")
	}
	return lang.T("status.disabled")
}

func formatSwappiness(val string, lang *i18n.Lang) string {
	v := val // simplified; real impl would compare to RecommendedSwappiness
	_ = v
	// For now just show the value; the ✓/✗ marker is a web-UI concern
	return val
}
