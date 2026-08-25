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
	"fmt"

	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"cryoutils-ng/core"
	"cryoutils-ng/i18n"
)

const uiConfigFile = "ui.json"

type uiConfig struct {
	Language string `json:"language"`
}

// uiCtx bundles everything a section builder needs.
type uiCtx struct {
	e             *core.Engine
	win           fyne.Window
	lang          *i18n.Lang
	auth          *authManager
	prog          *widget.ProgressBarInfinite
	progLabel     *widget.Label
	refreshStatus func()
}

// runTask executes a privileged long task with progress UI.
// It must be called from the UI thread; done runs inside fyne.Do on completion.
func (c *uiCtx) runTask(msg string, task func() error, done func(error)) {
	c.auth.ensure(func() {
		if msg != "" {
			c.progLabel.SetText(msg)
			c.progLabel.Show()
		}
		c.prog.Show()
		go func() {
			err := task()
			fyne.Do(func() {
				c.prog.Hide()
				c.progLabel.Hide()
				if err != nil {
					dialog.ShowError(err, c.win)
				}
				if done != nil {
					done(err)
				}
			})
		}()
	}, nil)
}

// Run starts the Fyne native UI. It blocks until the window is closed.
func Run(e *core.Engine) {
	a := app.NewWithID("io.cryoutils-ng")
	a.Settings().SetTheme(&customTheme{inner: theme.DarkTheme()})

	cfg := loadUIConfig()
	code := cfg.Language
	if i18n.Load(code) == nil {
		code = "en"
	}

	w := a.NewWindow("CryoUtils NG")
	w.Resize(fyne.NewSize(1280, 800))

	prog := widget.NewProgressBarInfinite()
	prog.Hide()
	progLabel := widget.NewLabel("")
	progLabel.Hide()
	e.SetProgressCallback(func(msg string) {
		fyne.Do(func() {
			progLabel.SetText(msg)
			progLabel.Show()
			prog.Show()
		})
	})

	auth := newAuthManager(e, w)
	var current *uiCtx

	var buildPage func()
	buildPage = func() {
		lang := i18n.Load(code)
		c := &uiCtx{e: e, win: w, lang: lang, auth: auth, prog: prog, progLabel: progLabel}
		current = c

		statusBox := container.NewVBox(sectionsStatus(c))
		c.refreshStatus = func() {
			statusBox.Objects = []fyne.CanvasObject{sectionsStatus(c)}
			statusBox.Refresh()
		}

		banner := widget.NewRichText(&widget.TextSegment{
			Text:  lang.T("native.sudoBanner"),
			Style: widget.RichTextStyle{ColorName: theme.ColorNameError},
		})
		auth.attach(lang, banner)

		title := widget.NewLabel("CryoUtils NG " + core.CurrentVersionNumber)
		title.TextStyle = fyne.TextStyle{Bold: true}
		refreshBtn := widget.NewButton(lang.T("native.refreshBtn"), c.refreshStatus)

		langSelect := widget.NewSelect(i18n.Available(), func(s string) {
			if s == "" || s == code {
				return
			}
			code = s
			cfg.Language = s
			saveUIConfig(cfg)
			buildPage()
		})
		langSelect.SetSelected(code)

		header := container.NewHBox(title, refreshBtn, widget.NewLabel(lang.T("native.langMenu")), langSelect)

		content := container.NewVBox(
			banner,
			statusBox,
			sectionsSwap(c),
			sectionsMemory(c),
			sectionsVRAM(c),
			sectionsPresets(c),
			sectionsGameData(c),
		)
		scroll := container.NewVScroll(content)
		w.SetContent(container.NewBorder(header, nil, nil, nil, scroll))

		if os.Getenv("CRYOUTILS_UI_DEBUG") != "" {
			go func() {
				time.Sleep(4 * time.Second)
				fyne.Do(func() {
					fmt.Println("DEBUG scroll.Size():", scroll.Size())
					fmt.Println("DEBUG content.MinSize():", content.MinSize())
					for i, o := range content.Objects {
						fmt.Printf("DEBUG obj[%d] %T visible=%v size=%v min=%v\n", i, o, o.Visible(), o.Size(), o.MinSize())
					}
				})
			}()
		}
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			fyne.Do(func() {
				if current != nil && current.refreshStatus != nil {
					current.refreshStatus()
				}
			})
		}
	}()

	buildPage()

	if shot := os.Getenv("CRYOUTILS_UI_SHOT"); shot != "" {
		delay := 3 * time.Second
		if d := os.Getenv("CRYOUTILS_UI_SHOT_DELAY_MS"); d != "" {
			if ms, err := strconv.Atoi(d); err == nil {
				delay = time.Duration(ms) * time.Millisecond
			}
		}
		go func() {
			time.Sleep(delay)
			fyne.Do(func() {
				img := w.Canvas().Capture()
				f, err := os.Create(shot)
				if err == nil {
					_ = png.Encode(f, img)
					_ = f.Close()
				} else {
					fyne.LogError("screenshot save failed", err)
				}
				os.Exit(0)
			})
		}()
	}

	if os.Geteuid() != 0 && e.Password == "" {
		go func() {
			time.Sleep(200 * time.Millisecond)
			fyne.Do(func() { auth.ensure(nil, nil) })
		}()
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
