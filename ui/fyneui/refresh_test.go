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
	"io"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"cryoutils-ng/core"
	"cryoutils-ng/i18n"
)

// TestSwapSectionRegistersRefresher covers the wiring behind the stale dropdown:
// the swap card is built once, so it has to register an in-place refresher, and
// that refresher must be safe to run repeatedly from the 5s ticker.
func TestSwapSectionRegistersRefresher(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	e := core.NewEngine(io.Discard, io.Discard)
	win := app.NewWindow("test")
	c := &uiCtx{
		e:         e,
		win:       win,
		lang:      i18n.Load("en"),
		prog:      widget.NewProgressBarInfinite(),
		progLabel: widget.NewLabel(""),
	}
	c.auth = newAuthManager(e, win)

	swapSection := sectionsSwap(c)
	card, ok := swapSection.(*widget.Card)
	if !ok {
		t.Fatalf("sectionsSwap returned %T, want *widget.Card", swapSection)
	}
	if len(c.refreshers) != 1 {
		t.Fatalf("swap section registered %d refresher(s), want 1", len(c.refreshers))
	}

	box, ok := card.Content.(*fyne.Container)
	if !ok {
		t.Fatalf("card content is %T, want a container", card.Content)
	}
	var selectSize *widget.Select
	for _, o := range box.Objects {
		if s, ok := o.(*widget.Select); ok {
			selectSize = s
		}
	}
	if selectSize == nil {
		t.Fatalf("no select found in the swap card: %d object(s)", len(box.Objects))
	}
	if selectSize.Selected == "" {
		t.Fatalf("the size dropdown starts without a selection")
	}

	// Running the refresher must keep the widget consistent, and the selected
	// entry must always be one that is actually offered.
	for range 3 {
		c.refreshers[0]()
		if !containsOption(selectSize.Options, selectSize.Selected) {
			t.Fatalf("after refresh the selection %q is not among %v", selectSize.Selected, selectSize.Options)
		}
	}
}

func containsOption(options []string, want string) bool {
	for _, o := range options {
		if o == want {
			return true
		}
	}
	return false
}
