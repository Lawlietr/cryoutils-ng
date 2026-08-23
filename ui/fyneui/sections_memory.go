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

	"cryoutils-ng/core"
	"cryoutils-ng/i18n"
)

func sectionsMemory(e *core.Engine, lang *i18n.Lang, r *refreshers) fyne.CanvasObject {
	sum := e.GetStatusSummary()

	hugePages := widget.NewCheck(lang.T("memoryParams.hugepages"), func(bool) {})
	hugePages.SetChecked(sum["HugePages"] == "true")
	hugePages.Disabled()

	shmem := widget.NewCheck(lang.T("memoryParams.shmem"), func(bool) {})
	shmem.SetChecked(sum["ShMem"] == "true")
	shmem.Disabled()

	compaction := widget.NewCheck(lang.T("memoryParams.compaction_proactiveness"), func(bool) {})
	compaction.SetChecked(sum["CompactionProactiveness"] == "true")
	compaction.Disabled()

	defrag := widget.NewCheck(lang.T("memoryParams.defrag"), func(bool) {})
	defrag.SetChecked(sum["Defrag"] == "true")
	defrag.Disabled()

	pageLock := widget.NewCheck(lang.T("memoryParams.page_lock_unfairness"), func(bool) {})
	pageLock.SetChecked(sum["PageLockUnfairness"] == "true")
	pageLock.Disabled()

	r.add(func() {
		sum = e.GetStatusSummary()
		hugePages.SetChecked(sum["HugePages"] == "true")
		shmem.SetChecked(sum["ShMem"] == "true")
		compaction.SetChecked(sum["CompactionProactiveness"] == "true")
		defrag.SetChecked(sum["Defrag"] == "true")
		pageLock.SetChecked(sum["PageLockUnfairness"] == "true")
	})

	title := widget.NewLabel(lang.T("memory.title"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewVBox(
		title,
		hugePages, shmem, compaction, defrag, pageLock,
		widget.NewSeparator(),
	)
}
