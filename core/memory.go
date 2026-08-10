// CryoUtilities
// Copyright (C) 2023 CryoByte33 and contributors to the CryoUtilities project
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

package core

// ─── Status helpers ─────────────────────────────────────────────────────────

func (e *Engine) getHugePagesStatus() bool {
	status, err := e.getUnitStatus("hugepages")
	if err != nil {
		e.ErrorLog.Println("Unable to get current hugepages value")
		return false
	}
	return status == RecommendedHugePages
}

func (e *Engine) getCompactionProactivenessStatus() bool {
	status, err := e.getUnitStatus("compaction_proactiveness")
	if err != nil {
		e.ErrorLog.Println("Unable to get current compaction_proactiveness")
		return false
	}
	return status == RecommendedCompactionProactiveness
}

func (e *Engine) getPageLockUnfairnessStatus() bool {
	status, err := e.getUnitStatus("page_lock_unfairness")
	if err != nil {
		e.ErrorLog.Println("Unable to get current page_lock_unfairness")
		return false
	}
	return status == RecommendedPageLockUnfairness
}

func (e *Engine) getShMemStatus() bool {
	status, err := e.getUnitStatus("shmem_enabled")
	if err != nil {
		e.ErrorLog.Println("Unable to get current shmem_enabled")
		return false
	}
	return status == RecommendedShMem
}

func (e *Engine) getDefragStatus() bool {
	status, err := e.getUnitStatus("defrag")
	if err != nil {
		e.ErrorLog.Println("Unable to get current defrag")
		return false
	}
	return status == RecommendedHugePageDefrag
}

// ─── Toggle helpers ─────────────────────────────────────────────────────────

// ToggleHugePages toggles hugepages between recommended and default.
func (e *Engine) ToggleHugePages() error {
	if e.getHugePagesStatus() {
		return e.RevertHugePages()
	}
	return e.SetHugePages()
}

// ToggleShMem toggles shmem between recommended and default.
func (e *Engine) ToggleShMem() error {
	if e.getShMemStatus() {
		return e.RevertShMem()
	}
	return e.SetShMem()
}

// ToggleCompactionProactiveness toggles between recommended and default.
func (e *Engine) ToggleCompactionProactiveness() error {
	if e.getCompactionProactivenessStatus() {
		return e.RevertCompactionProactiveness()
	}
	return e.SetCompactionProactiveness()
}

// ToggleDefrag toggles defrag between recommended and default.
func (e *Engine) ToggleDefrag() error {
	if e.getDefragStatus() {
		return e.RevertDefrag()
	}
	return e.SetDefrag()
}

// TogglePageLockUnfairness toggles between recommended and default.
func (e *Engine) TogglePageLockUnfairness() error {
	if e.getPageLockUnfairnessStatus() {
		return e.RevertPageLockUnfairness()
	}
	return e.SetPageLockUnfairness()
}

// ─── Set helpers (recommended) ──────────────────────────────────────────────

func (e *Engine) SetHugePages() error {
	e.InfoLog.Println("Enabling hugepages...")
	_ = e.removeFile(NHPTestingFile)
	if err := e.setUnitValue("hugepages", RecommendedHugePages); err != nil {
		return err
	}
	return e.writeUnitFile("hugepages", RecommendedHugePages)
}

func (e *Engine) SetCompactionProactiveness() error {
	e.InfoLog.Println("Setting compaction_proactiveness...")
	if err := e.setUnitValue("compaction_proactiveness", RecommendedCompactionProactiveness); err != nil {
		return err
	}
	return e.writeUnitFile("compaction_proactiveness", RecommendedCompactionProactiveness)
}

func (e *Engine) SetPageLockUnfairness() error {
	e.InfoLog.Println("Enabling page_lock_unfairness...")
	if err := e.setUnitValue("page_lock_unfairness", RecommendedPageLockUnfairness); err != nil {
		return err
	}
	return e.writeUnitFile("page_lock_unfairness", RecommendedPageLockUnfairness)
}

func (e *Engine) SetShMem() error {
	e.InfoLog.Println("Enabling shmem_enabled...")
	if err := e.setUnitValue("shmem_enabled", RecommendedShMem); err != nil {
		return err
	}
	return e.writeUnitFile("shmem_enabled", RecommendedShMem)
}

func (e *Engine) SetDefrag() error {
	e.InfoLog.Println("Disabling hugepage defragmentation...")
	if err := e.setUnitValue("defrag", RecommendedHugePageDefrag); err != nil {
		return err
	}
	return e.writeUnitFile("defrag", RecommendedHugePageDefrag)
}

// ─── Revert helpers (default / stock) ───────────────────────────────────────

func (e *Engine) RevertHugePages() error {
	e.InfoLog.Println("Disabling hugepages...")
	if err := e.setUnitValue("hugepages", DefaultHugePages); err != nil {
		return err
	}
	return e.removeUnitFile("hugepages")
}

func (e *Engine) RevertCompactionProactiveness() error {
	e.InfoLog.Println("Disabling compaction_proactiveness...")
	if err := e.setUnitValue("compaction_proactiveness", DefaultCompactionProactiveness); err != nil {
		return err
	}
	return e.removeUnitFile("compaction_proactiveness")
}

func (e *Engine) RevertPageLockUnfairness() error {
	e.InfoLog.Println("Disabling page_lock_unfairness...")
	if err := e.setUnitValue("page_lock_unfairness", DefaultPageLockUnfairness); err != nil {
		return err
	}
	return e.removeUnitFile("page_lock_unfairness")
}

func (e *Engine) RevertShMem() error {
	e.InfoLog.Println("Disabling shmem_enabled...")
	if err := e.setUnitValue("shmem_enabled", DefaultShMem); err != nil {
		return err
	}
	return e.removeUnitFile("shmem_enabled")
}

func (e *Engine) RevertDefrag() error {
	e.InfoLog.Println("Enabling hugepage defragmentation...")
	if err := e.setUnitValue("defrag", DefaultHugePageDefrag); err != nil {
		return err
	}
	return e.removeUnitFile("defrag")
}
