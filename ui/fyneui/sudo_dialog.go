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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"cryoutils-ng/core"
	"cryoutils-ng/i18n"
)

// askPassword shows a modal sudo password popup. The password is verified
// with TestAuth (run off the UI goroutine); on success it is stored in
// e.Password — memory only, never written to disk. Returns true if the user
// authenticated, false if they skipped.
func askPassword(e *core.Engine, w fyne.Window, lang *i18n.Lang) bool {
	entry := widget.NewPasswordEntry()
	entry.Resize(fyne.NewSize(380, 0)) // widen the popup via content min size
	errLabel := widget.NewLabel("")
	errLabel.Hide()
	errLabel.Importance = widget.HighImportance

	result := make(chan bool, 1)
	var pop *widget.PopUp
	done := func(ok bool) { result <- ok }

	skip := widget.NewButton(lang.T("native.sudoDialogSkip"), func() {
		pop.Hide()
		done(false)
	})
	var enter *widget.Button
	enter = widget.NewButton(lang.T("native.sudoDialogEnter"), func() {
		pw := entry.Text
		if pw == "" {
			return
		}
		errLabel.Hide()
		enter.Disable()
		go func() {
			err := e.TestAuth(pw)
			fyne.Do(func() {
				enter.Enable()
				if err != nil {
					errLabel.SetText(lang.T("native.sudoAuthFailed"))
					errLabel.Show()
					entry.SetText("")
					return
				}
				e.Password = pw
				pop.Hide()
				done(true)
			})
		}()
	})

	title := widget.NewLabel(lang.T("native.sudoDialogTitle"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	pop = widget.NewModalPopUp(container.NewVBox(
		title,
		widget.NewLabel(lang.T("native.sudoDialogPrompt")),
		entry,
		errLabel,
		container.NewHBox(layout.NewSpacer(), skip, enter),
	), w.Canvas())
	pop.Show()

	ok := <-result
	if ok {
		pop.Hide()
	}
	return ok
}

// withAuth ensures sudo authentication is current before a privileged
// operation: it refreshes the cached sudo timestamp and re-prompts the user
// interactively when the stored password is missing or no longer valid.
// Returns false when the user declined (read-only mode). Must be called on
// the UI goroutine before launching the privileged work.
func withAuth(e *core.Engine, w fyne.Window, lang *i18n.Lang) bool {
	if os.Geteuid() == 0 {
		return true
	}
	if e.Password == "" {
		return askPassword(e, w, lang) // TestAuth leaves a fresh timestamp
	}
	if err := e.RenewAuth(); err != nil {
		e.Password = "" // stored password is no longer valid
		return askPassword(e, w, lang)
	}
	return true
}
