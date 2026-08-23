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

// Package fyneui provides the native Fyne UI for CryoUtils NG.
package fyneui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"cryoutils-ng/core"
	"cryoutils-ng/i18n"
)

const uiConfigFile = "ui.json"

type uiConfig struct {
	Language string `json:"language"`
}

// refreshers collects in-place value updaters registered by the sections.
// The 5-second ticker and the manual Refresh button run these without
// rebuilding the widget tree, so user selections and scroll position are
// preserved. Only a language switch triggers a full rebuild.
type refreshers struct {
	funcs []func()
}

func (r *refreshers) add(f func()) {
	r.funcs = append(r.funcs, f)
}

// run executes all registered updaters. Must be called on the UI goroutine.
func (r *refreshers) run() {
	for _, f := range r.funcs {
		f()
	}
}

// Run starts the Fyne native UI. It blocks until the window is closed.
func Run(e *core.Engine) {
	a := app.NewWithID("io.cryoutils-ng")
	a.Settings().SetTheme(&customTheme{inner: theme.DarkTheme()})

	// Load saved language
	cfg := loadUIConfig()
	langCode := cfg.Language
	if i18n.Load(langCode) == nil {
		langCode = "en"
	}
	lang := i18n.Load(langCode)

	w := a.NewWindow("CryoUtils NG")

	// Progress bar (hidden by default)
	prog := widget.NewProgressBar()
	prog.Hide()
	progLabel := widget.NewLabel("")
	progLabel.Hide()
	e.SetProgressCallback(func(msg string) {
		fyne.Do(func() {
			prog.Show()
			progLabel.Show()
			progLabel.SetText(msg)
			prog.SetValue(0) // indeterminate via zero value
		})
	})

	// In-place updaters; recreated on every full rebuild so closures from
	// the previous widget tree (and their duplicate updates) are dropped.
	var refreshers_ *refreshers

	// Sudo warning banner (shown while non-root and unauthenticated)
	banner := widget.NewLabel("")
	banner.Importance = widget.HighImportance
	updateBanner := func() {
		banner.SetText(lang.T("native.sudoBanner"))
		if os.Geteuid() != 0 && e.Password == "" {
			banner.Show()
		} else {
			banner.Hide()
		}
	}
	updateBanner()

	// Build content (scrollable body); a fresh refreshers registry per build.
	buildContent := func() fyne.CanvasObject {
		refreshers_ = &refreshers{}
		refreshers_.add(updateBanner)
		return container.NewVBox(
			sectionsStatus(e, lang, refreshers_),
			sectionsSwap(e, w, lang, prog, progLabel, refreshers_),
			sectionsMemory(e, lang, refreshers_),
			sectionsVRAM(e, lang, refreshers_),
			sectionsPresets(e, w, lang, prog, progLabel),
			sectionsGameData(e, w, lang, prog, progLabel, refreshers_),
		)
	}

	// In-place refresh (ticker + Refresh button): updates values only.
	refreshValues := func() {
		if refreshers_ != nil {
			refreshers_.run()
		}
	}

	var rebuild func() // bound after buildRoot is defined below

	// Build top bar (title + refresh button + language select)
	buildTop := func() fyne.CanvasObject {
		title := widget.NewLabel("CryoUtils NG v" + core.CurrentVersionNumber)
		title.TextStyle = fyne.TextStyle{Bold: true}
		refreshBtn := widget.NewButton(lang.T("native.refreshBtn"), refreshValues)
		langSelect := widget.NewSelect(
			i18n.Available(),
			func(s string) {
				lang = i18n.Load(s)
				cfg.Language = s
				saveUIConfig(cfg)
				// Language change requires a full rebuild (labels capture lang).
				if rebuild != nil {
					rebuild()
				}
			},
		)
		langSelect.SetSelected(lang.Label())
		return container.NewBorder(nil, nil, langSelect, nil, container.NewHBox(title, refreshBtn))
	}

	// Root layout: fixed top bar + scroll filling remaining space.
	// (A VBox would resize the Scroll to its tiny MinSize height.)
	buildRoot := func() fyne.CanvasObject {
		return container.NewBorder(container.NewVBox(buildTop(), banner, prog, progLabel), nil, nil, nil,
			container.NewVScroll(buildContent()))
	}
	rebuild = func() { w.SetContent(buildRoot()) }

	// 5-second auto-refresh ticker (in-place value updates only)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			fyne.Do(refreshValues)
		}
	}()

	w.SetContent(buildRoot())
	w.Show()
	w.Resize(fyne.NewSize(1280, 800))
	w.CenterOnScreen()

	// First-launch sudo prompt (non-root only; root execution skips core auth)
	if os.Geteuid() != 0 && e.Password == "" {
		if askPassword(e, w, lang) {
			updateBanner()
		}
	}

	w.ShowAndRun()
}

func loadUIConfig() uiConfig {
	var cfg uiConfig
	data, err := os.ReadFile(filepath.Join(core.InstallDirectory, uiConfigFile))
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	return cfg
}

func saveUIConfig(cfg uiConfig) {
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(filepath.Join(core.InstallDirectory, uiConfigFile), data, 0644)
}
