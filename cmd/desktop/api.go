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

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cryoutils-ng/core"
)

// APIResponse is the standard JSON response shape.
type APIResponse struct {
	Success bool              `json:"success"`
	Error   string            `json:"error,omitempty"`
	Data    json.RawMessage   `json:"data,omitempty"`
}

// setupAPIRoutes registers all API endpoints on the given mux.
func setupAPIRoutes(mux *http.ServeMux, e *core.Engine, token string) {
	mux.HandleFunc("/api/progress", func(w http.ResponseWriter, r *http.Request) {
		handleProgress(w, r)
	})
	mux.HandleFunc("/api/auth", func(w http.ResponseWriter, r *http.Request) {
		handleAuth(w, r, e)
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		handleStatus(w, r, e, token)
	})
	mux.HandleFunc("/api/swap/resize", func(w http.ResponseWriter, r *http.Request) {
		handleSwapResize(w, r, e, token)
	})
	mux.HandleFunc("/api/swap/swappiness", func(w http.ResponseWriter, r *http.Request) {
		handleSwapSwappiness(w, r, e, token)
	})
	mux.HandleFunc("/api/memory/", func(w http.ResponseWriter, r *http.Request) {
		handleMemoryToggle(w, r, e, token)
	})
	mux.HandleFunc("/api/recommended", func(w http.ResponseWriter, r *http.Request) {
		handleRecommended(w, r, e, token)
	})
	mux.HandleFunc("/api/stock", func(w http.ResponseWriter, r *http.Request) {
		handleStock(w, r, e, token)
	})
	mux.HandleFunc("/api/gamedata/sync", func(w http.ResponseWriter, r *http.Request) {
		handleGameDataSync(w, r, e, token)
	})
	mux.HandleFunc("/api/gamedata/cleanup", func(w http.ResponseWriter, r *http.Request) {
		handleGameDataCleanup(w, r, e, token)
	})
	mux.HandleFunc("/api/libraries", func(w http.ResponseWriter, r *http.Request) {
		handleLibraries(w, r, e, token)
	})
}

func respondJSON(w http.ResponseWriter, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func respondError(w http.ResponseWriter, msg string, code int) {
	respondJSON(w, APIResponse{Success: false, Error: msg})
	w.WriteHeader(code)
}

func requireToken(w http.ResponseWriter, r *http.Request, token string) bool {
	if r.URL.Query().Get("token") != token {
		respondJSON(w, APIResponse{Success: false, Error: "unauthorized"})
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}

func emitProgress(msg string) {
	core.ProgressChannel <- core.ProgressEvent{Message: msg}
}

// ─── Auth ────────────────────────────────────────────────────────────────────

func handleAuth(w http.ResponseWriter, r *http.Request, e *core.Engine) {
	if r.Method != http.MethodPost {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := e.TestAuth(body.Password); err != nil {
		respondJSON(w, APIResponse{Success: false, Error: "authentication failed"})
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	e.Password = body.Password
	respondJSON(w, APIResponse{Success: true})
}

// ─── Status ──────────────────────────────────────────────────────────────────

func handleStatus(w http.ResponseWriter, r *http.Request, e *core.Engine, token string) {
	if !requireToken(w, r, token) {
		return
	}
	if r.Method != http.MethodGet {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summary := e.GetStatusSummary()
	data, _ := json.Marshal(summary)
	respondJSON(w, APIResponse{Success: true, Data: data})
}

// ─── Swap ────────────────────────────────────────────────────────────────────

func handleSwapResize(w http.ResponseWriter, r *http.Request, e *core.Engine, token string) {
	if !requireToken(w, r, token) {
		return
	}
	if r.Method != http.MethodPost {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Size int `json:"size"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Size < 1 {
		respondError(w, "invalid size (must be >= 1)", http.StatusBadRequest)
		return
	}

	go func() {
		emitProgress(fmt.Sprintf("Resizing swap to %d GB...", body.Size))
		e.RenewAuth()
		if err := e.DisableSwap(); err != nil {
			emitProgress("Error disabling swap: " + err.Error())
			return
		}
		emitProgress("Swap disabled, resizing...")
		if err := e.ResizeSwapFile(body.Size); err != nil {
			emitProgress("Error resizing: " + err.Error())
			return
		}
		e.RenewAuth()
		emitProgress("Setting permissions...")
		if err := e.SetSwapPermissions(); err != nil {
			emitProgress("Error setting permissions: " + err.Error())
			return
		}
		emitProgress("Enabling swap...")
		if err := e.InitNewSwapFile(); err != nil {
			emitProgress("Error enabling swap: " + err.Error())
			return
		}
		emitProgress("Swap resize complete.")
	}()

	respondJSON(w, APIResponse{Success: true})
}

func handleSwapSwappiness(w http.ResponseWriter, r *http.Request, e *core.Engine, token string) {
	if !requireToken(w, r, token) {
		return
	}
	if r.Method != http.MethodPost {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	val := body.Value
	if v, _ := strconv.Atoi(val); (v < 0 || v > 200) && val != core.RecommendedSwappiness && val != core.DefaultSwappiness {
		respondError(w, "invalid swappiness value (0-200)", http.StatusBadRequest)
		return
	}

	go func() {
		emitProgress(fmt.Sprintf("Setting swappiness to %s...", val))
		if err := e.ChangeSwappiness(val); err != nil {
			emitProgress("Error: " + err.Error())
			return
		}
		emitProgress("Swappiness set.")
	}()

	respondJSON(w, APIResponse{Success: true})
}

// ─── Memory ──────────────────────────────────────────────────────────────────

func handleMemoryToggle(w http.ResponseWriter, r *http.Request, e *core.Engine, token string) {
	if !requireToken(w, r, token) {
		return
	}
	if r.Method != http.MethodPost {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	param := strings.TrimPrefix(r.URL.Path, "/api/memory/")
	if param == "" {
		respondError(w, "missing parameter name", http.StatusBadRequest)
		return
	}

	go func() {
		var err error
		switch param {
		case "hugepages":
			emitProgress("Toggling hugepages...")
			err = e.ToggleHugePages()
		case "shmem":
			emitProgress("Toggling shmem...")
			err = e.ToggleShMem()
		case "compaction_proactiveness":
			emitProgress("Toggling compaction_proactiveness...")
			err = e.ToggleCompactionProactiveness()
		case "defrag":
			emitProgress("Toggling defrag...")
			err = e.ToggleDefrag()
		case "page_lock_unfairness":
			emitProgress("Toggling page_lock_unfairness...")
			err = e.TogglePageLockUnfairness()
		default:
			emitProgress("Unknown parameter: " + param)
			return
		}
		if err != nil {
			emitProgress("Error: " + err.Error())
			return
		}
		emitProgress(param + " toggled.")
	}()

	respondJSON(w, APIResponse{Success: true})
}

// ─── Presets ─────────────────────────────────────────────────────────────────

func handleRecommended(w http.ResponseWriter, r *http.Request, e *core.Engine, token string) {
	if !requireToken(w, r, token) {
		return
	}
	if r.Method != http.MethodPost {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		emitProgress("Applying recommended settings...")
		if err := e.UseRecommendedSettings(); err != nil {
			emitProgress("Error: " + err.Error())
			return
		}
		emitProgress("Recommended settings applied.")
	}()

	respondJSON(w, APIResponse{Success: true})
}

func handleStock(w http.ResponseWriter, r *http.Request, e *core.Engine, token string) {
	if !requireToken(w, r, token) {
		return
	}
	if r.Method != http.MethodPost {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		emitProgress("Reverting to stock settings...")
		if err := e.UseStockSettings(); err != nil {
			emitProgress("Error: " + err.Error())
			return
		}
		emitProgress("Stock settings applied.")
	}()

	respondJSON(w, APIResponse{Success: true})
}

// ─── Game Data ───────────────────────────────────────────────────────────────

func handleGameDataSync(w http.ResponseWriter, r *http.Request, e *core.Engine, token string) {
	if !requireToken(w, r, token) {
		return
	}
	if r.Method != http.MethodPost {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Left  string `json:"left"`
		Right string `json:"right"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Left == "" || body.Right == "" {
		respondError(w, "invalid request body (need left and right paths)", http.StatusBadRequest)
		return
	}

	go func() {
		emitProgress(fmt.Sprintf("Syncing game data from %s to %s...", body.Left, body.Right))
		data, err := e.GetDataToMove(body.Left, body.Right)
		if err != nil {
			emitProgress("Error getting data: " + err.Error())
			return
		}
		emitProgress(fmt.Sprintf("Found %d games to move.", len(data.GetRight())+len(data.GetLeft())))
		if err := e.MoveGameData(body.Left, body.Right, data); err != nil {
			emitProgress("Error moving data: " + err.Error())
			return
		}
		emitProgress("Game data sync complete.")
	}()

	respondJSON(w, APIResponse{Success: true})
}

func handleGameDataCleanup(w http.ResponseWriter, r *http.Request, e *core.Engine, token string) {
	if !requireToken(w, r, token) {
		return
	}
	if r.Method != http.MethodPost {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Left  string `json:"left"`
		Right string `json:"right"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Left == "" || body.Right == "" {
		respondError(w, "invalid request body (need left and right paths)", http.StatusBadRequest)
		return
	}

	go func() {
		emitProgress(fmt.Sprintf("Cleaning up uninstalled game data from %s...", body.Right))
		data, err := e.GetUninstalledGamesData(body.Left, body.Right)
		if err != nil {
			emitProgress("Error: " + err.Error())
			return
		}
		emitProgress(fmt.Sprintf("Found %d items to remove.", len(data.GetRight())))
		e.RemoveGameData(data.GetRight(), []string{body.Right})
		emitProgress("Cleanup complete.")
	}()

	respondJSON(w, APIResponse{Success: true})
}

// ─── Libraries ───────────────────────────────────────────────────────────────

func handleLibraries(w http.ResponseWriter, r *http.Request, e *core.Engine, token string) {
	if !requireToken(w, r, token) {
		return
	}
	if r.Method != http.MethodGet {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	locations, err := e.GetLibraryLocations()
	if err != nil {
		respondError(w, "error reading libraries: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(locations)
	respondJSON(w, APIResponse{Success: true, Data: data})
}
