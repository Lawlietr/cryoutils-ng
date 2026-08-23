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

func sectionsVRAM(e *core.Engine, lang *i18n.Lang, r *refreshers) fyne.CanvasObject {
	sum := e.GetStatusSummary()
	vram := widget.NewLabel(fmt.Sprintf("%s: %s", lang.T("status.vram"), sum["VRAM"]))
	vram.Importance = widget.LowImportance
	r.add(func() {
		vram.SetText(fmt.Sprintf("%s: %s", lang.T("status.vram"), e.GetStatusSummary()["VRAM"]))
	})

	notes := widget.NewLabel(lang.T("vram.readOnly"))
	notes.TextStyle = fyne.TextStyle{Italic: true}

	title := widget.NewLabel(lang.T("vram.title"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewVBox(
		title,
		vram,
		notes,
		widget.NewSeparator(),
	)
}
