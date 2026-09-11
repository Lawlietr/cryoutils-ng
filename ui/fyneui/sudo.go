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
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"cryoutils-ng/core"
	"cryoutils-ng/i18n"
)

// authManager owns the sudo password lifecycle for the native UI.
// The password is kept only in memory (core.Engine.Password).
type authManager struct {
	e       *core.Engine
	win     fyne.Window
	lang    *i18n.Lang
	banner  fyne.CanvasObject
	skipped bool
	// dialogOpen is true while a password dialog is on screen. It guarantees
	// only one dialog at a time: the startup prompt is shown 200ms after
	// launch, and if the user clicks an operation before answering, the
	// request is queued instead of opening a second dialog.
	dialogOpen  bool
	pending     func()
	pendingSkip func()
}

func newAuthManager(e *core.Engine, win fyne.Window) *authManager {
	return &authManager{e: e, win: win}
}

// attach registers the warning banner for the current language page.
func (a *authManager) attach(lang *i18n.Lang, banner fyne.CanvasObject) {
	a.lang = lang
	a.banner = banner
	if !isRoot() && a.skipped && a.e.Password == "" {
		banner.Show()
	} else {
		banner.Hide()
	}
}

func isRoot() bool {
	return os.Geteuid() == 0
}

// ensure runs run on the UI thread once privilege is available.
// If sudo auth is missing or expired it prompts for the password;
// when the user skips, it shows the warning banner and calls onSkip.
func (a *authManager) ensure(run func(), onSkip func()) {
	if isRoot() {
		if run != nil {
			run()
		}
		return
	}
	if err := a.e.RenewAuth(); err == nil {
		if run != nil {
			run()
		}
		return
	}
	a.prompt(run, onSkip)
}

// flushPending re-enters ensure for a request that arrived while the
// password dialog was open. Must be called after the dialog closed.
func (a *authManager) flushPending() {
	if a.pending == nil && a.pendingSkip == nil {
		return
	}
	run, onSkip := a.pending, a.pendingSkip
	a.pending, a.pendingSkip = nil, nil
	a.ensure(run, onSkip)
}

// prompt shows the password dialog. A wrong password re-prompts.
// If a dialog is already open, the request is queued (flushed after the
// current dialog resolves) so the user is never asked twice at once.
func (a *authManager) prompt(run func(), onSkip func()) {
	if a.dialogOpen {
		a.pending = run
		a.pendingSkip = onSkip
		return
	}
	a.dialogOpen = true

	lang := a.lang
	if lang == nil {
		lang = i18n.Load("en")
	}
	entry := widget.NewPasswordEntry()

	var dlg dialog.Dialog

	verify := func() {
		if err := a.e.TestAuth(entry.Text); err != nil {
			a.dialogOpen = false
			dialog.ShowError(err, a.win)
			dlg.Hide()
			a.prompt(run, onSkip)
			return
		}
		a.dialogOpen = false
		a.e.Password = entry.Text
		a.skipped = false
		if a.banner != nil {
			a.banner.Hide()
		}
		dlg.Hide()
		if run != nil {
			run()
		}
		a.flushPending()
	}

	skip := func() {
		a.dialogOpen = false
		a.skipped = true
		dlg.Hide()
		if a.banner != nil {
			a.banner.Show()
		}
		if onSkip != nil {
			onSkip()
		}
		a.flushPending()
	}

	entry.OnSubmitted = func(string) { verify() }
	enterBtn := widget.NewButton(lang.T("native.sudoDialogEnter"), verify)
	skipBtn := widget.NewButton(lang.T("native.sudoDialogSkip"), skip)

	content := container.NewVBox(
		widget.NewLabel(lang.T("native.sudoDialogPrompt")),
		entry,
		container.NewHBox(enterBtn, skipBtn),
	)

	dlg = dialog.NewCustom(lang.T("native.sudoDialogTitle"), "", content, a.win)
	dlg.Show()
}
