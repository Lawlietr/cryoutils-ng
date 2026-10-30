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
)

func sectionsMemory(c *uiCtx) fyne.CanvasObject {
	lang := c.lang
	states := c.e.GetUnitStates()

	type toggleDef struct {
		labelKey string
		param    string // key in core.UnitMatrix, so the row follows the same source of truth the toggle writes
		toggle   func() error
	}
	defs := []toggleDef{
		{"memoryParams.hugepages", "hugepages", c.e.ToggleHugePages},
		{"memoryParams.shmem", "shmem_enabled", c.e.ToggleShMem},
		{"memoryParams.compaction_proactiveness", "compaction_proactiveness", c.e.ToggleCompactionProactiveness},
		{"memoryParams.defrag", "defrag", c.e.ToggleDefrag},
		{"memoryParams.page_lock_unfairness", "page_lock_unfairness", c.e.TogglePageLockUnfairness},
	}

	// The label carries the value the kernel actually holds. A bare checkmark
	// could not tell "recommended", "stock default" and "could not read" apart,
	// which is what made Huge Pages look broken on the Deck (P0.10).
	valueText := func(v string) string {
		if v == core.ValueUnknown {
			return lang.T("status.unknown")
		}
		return v
	}
	labelFor := func(d toggleDef, st core.UnitState) string {
		return fmt.Sprintf("%s  (%s)", lang.T(d.labelKey), valueText(st.Value))
	}

	// Two guards, and the reason for the second one is a Fyne behaviour trap:
	// Check.SetChecked() invokes OnChanged. So the state we push *into* a checkbox
	// (at build time, after a toggle, and from the 5s refresher) used to look like
	// a user click and fired another ToggleXxx() - which is why launching the app
	// silently flipped every parameter that was already at its recommended value
	// (Huge Pages always -> madvise). The Deck's own log showed the very first
	// memory action seconds after launch, with nobody clicking.
	var updating, syncing bool

	checks := make([]*widget.Check, 0, len(defs))
	box := container.NewVBox()

	// applyRow pushes the kernel's answer into a checkbox without pretending it
	// came from the user.
	applyRow := func(cb *widget.Check, d toggleDef, states map[string]core.UnitState) {
		st := states[d.param]
		syncing = true
		defer func() { syncing = false }()
		cb.SetText(labelFor(d, st))
		cb.SetChecked(st.IsRecommended)
	}

	for _, d := range defs {
		d := d
		var cb *widget.Check
		cb = widget.NewCheck(lang.T(d.labelKey), func(bool) {
			if syncing || updating {
				return
			}
			updating = true
			cb.Disable()
			c.runTask("", d.toggle, func(error) {
				applyRow(cb, d, c.e.GetUnitStates())
				cb.Enable()
				updating = false
				c.refreshStatus()
			})
		})
		box.Add(cb)
		checks = append(checks, cb)
		// Initial state comes from the kernel, guarded so it cannot toggle anything.
		applyRow(cb, d, states)
	}

	// The card is built once, so the checkboxes have to be refreshed in place:
	// otherwise an external change (CLI, another app, a failed write) leaves the
	// tick disagreeing with the kernel until the app is restarted.
	c.addRefresher(func() {
		if updating {
			return
		}
		fresh := c.e.GetUnitStates()
		for i, d := range defs {
			applyRow(checks[i], d, fresh)
		}
	})

	return widget.NewCard(lang.T("memory.title"), "", box)
}
