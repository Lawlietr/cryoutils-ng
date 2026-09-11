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
	"fyne.io/fyne/v2/widget"
)

func sectionsMemory(c *uiCtx) fyne.CanvasObject {
	lang := c.lang
	sum := c.e.GetStatusSummary()

	type toggleDef struct {
		labelKey string
		stateKey string
		toggle   func() error
	}
	defs := []toggleDef{
		{"memoryParams.hugepages", "HugePages", c.e.ToggleHugePages},
		{"memoryParams.shmem", "ShMem", c.e.ToggleShMem},
		{"memoryParams.compaction_proactiveness", "CompactionProactiveness", c.e.ToggleCompactionProactiveness},
		{"memoryParams.defrag", "Defrag", c.e.ToggleDefrag},
		{"memoryParams.page_lock_unfairness", "PageLockUnfairness", c.e.TogglePageLockUnfairness},
	}

	updating := false
	box := container.NewVBox()
	for _, d := range defs {
		d := d
		var cb *widget.Check
		cb = widget.NewCheck(lang.T(d.labelKey), func(bool) {
			if updating {
				return
			}
			updating = true
			cb.Disable()
			c.runTask("", d.toggle, func(error) {
				fresh := c.e.GetStatusSummary()
				cb.SetChecked(fresh[d.stateKey] == "true")
				cb.Enable()
				updating = false
				c.refreshStatus()
			})
		})
		cb.SetChecked(sum[d.stateKey] == "true")
		box.Add(cb)
	}

	return widget.NewCard(lang.T("memory.title"), "", box)
}
