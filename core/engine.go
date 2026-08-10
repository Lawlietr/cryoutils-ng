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

import (
	"fmt"
	"io"
	"log"
)

// OnProgressCallback is called with a progress message during long-running operations.
type OnProgressCallback func(string)

// Engine replaces the old global CryoUtils. It holds all state needed for
// tuning operations and replaces the previous global variable.
type Engine struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger

	Password string
	OnProgress OnProgressCallback

	SteamAPIResponse map[int]string
	SwapFileLocation string
}

// UseRecommendedSettings applies all recommended tuning parameters.
func (e *Engine) UseRecommendedSettings() error {
	if err := e.ChangeSwappiness(RecommendedSwappiness); err != nil {
		return err
	}
	if err := e.SetHugePages(); err != nil {
		return err
	}
	if err := e.SetCompactionProactiveness(); err != nil {
		return err
	}
	if err := e.SetDefrag(); err != nil {
		return err
	}
	if err := e.SetPageLockUnfairness(); err != nil {
		return err
	}
	if err := e.SetShMem(); err != nil {
		return err
	}
	return nil
}

// UseStockSettings reverts all tuning parameters to their default values.
func (e *Engine) UseStockSettings() error {
	if err := e.ChangeSwappiness(DefaultSwappiness); err != nil {
		return err
	}
	if err := e.RevertHugePages(); err != nil {
		return err
	}
	if err := e.RevertCompactionProactiveness(); err != nil {
		return err
	}
	if err := e.RevertDefrag(); err != nil {
		return err
	}
	if err := e.RevertPageLockUnfairness(); err != nil {
		return err
	}
	if err := e.RevertShMem(); err != nil {
		return err
	}
	return nil
}

// NewEngine creates a new Engine with the given log writers.
func NewEngine(infoWriter, errWriter io.Writer) *Engine {
	return &Engine{
		InfoLog:  log.New(infoWriter, "INFO  ", log.LstdFlags|log.Lmicroseconds),
		ErrorLog: log.New(errWriter, "ERROR ", log.LstdFlags|log.Lmicroseconds),
	}
}

// GetStatusSummary returns a map of current tuning statuses for CLI output.
func (e *Engine) GetStatusSummary() map[string]string {
	return map[string]string{
		"SwapFile":                e.SwapFileLocation,
		"SwapSizeGB":              fmt.Sprintf("%d", e.getSwapSizeGB()),
		"Swappiness":              fmt.Sprintf("%d", e.getSwappinessValueCLI()),
		"VRAM":                    fmt.Sprintf("%s", GetHumanVRAMSize(e.getVRAMValueCLI())),
		"HugePages":               fmt.Sprintf("%v", e.getHugePagesStatus()),
		"ShMem":                   fmt.Sprintf("%v", e.getShMemStatus()),
		"CompactionProactiveness": fmt.Sprintf("%v", e.getCompactionProactivenessStatus()),
		"Defrag":                  fmt.Sprintf("%v", e.getDefragStatus()),
		"PageLockUnfairness":      fmt.Sprintf("%v", e.getPageLockUnfairnessStatus()),
	}
}

func (e *Engine) getSwapSizeGB() int {
	size, _ := e.GetSwapFileSize()
	return int(size / int64(GigabyteMultiplier))
}

func (e *Engine) getSwappinessValueCLI() int {
	v, _ := e.GetSwappinessValue()
	return v
}

func (e *Engine) getVRAMValueCLI() int {
	v, _ := e.GetVRAMValue()
	return v
}
