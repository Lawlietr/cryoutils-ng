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

import "testing"

// The swap card is built once, so the dropdown used to keep the "Current Size"
// entry that was computed when the page was created: after a resize it still
// showed the previous size until the app was restarted. nextSwapSelection is the
// decision the in-place refresher makes.
func TestNextSwapSelectionTracksCurrentSize(t *testing.T) {
	atLaunch := []string{"1 - Default", "2", "4", "8", "12 - Current Size", "16"}
	afterResizeTo16 := []string{"1 - Default", "2", "4", "8", "12", "16 - Current Size"}

	if got := nextSwapSelection(pickCurrentSwapSize(atLaunch), afterResizeTo16); got != "16 - Current Size" {
		t.Errorf("after resize the dropdown shows %q, want the new current size entry", got)
	}

	// An explicit choice from the list survives a refresh.
	if got := nextSwapSelection("8", afterResizeTo16); got != "8" {
		t.Errorf("explicit choice = %q, want it preserved", got)
	}

	// An explicit choice that is no longer offered (free space in /home shrank)
	// falls back to the current size instead of leaving a missing selection.
	narrow := []string{"1 - Default", "2", "4 - Current Size"}
	if got := nextSwapSelection("16", narrow); got != "4 - Current Size" {
		t.Errorf("stale explicit choice = %q, want the current size entry", got)
	}

	// The default entry is a real option, so it is kept like any other choice.
	if got := nextSwapSelection("1 - Default", atLaunch); got != "1 - Default" {
		t.Errorf("default entry = %q, want it preserved", got)
	}

	// An empty previous selection must not stick.
	if got := nextSwapSelection("", afterResizeTo16); got != "16 - Current Size" {
		t.Errorf("empty previous = %q, want the current size entry", got)
	}
}

// TestPickCurrentSwapSizeSkipsWhenAbsent documents the helper the fallback uses.
func TestPickCurrentSwapSizeSkipsWhenAbsent(t *testing.T) {
	if got := pickCurrentSwapSize([]string{"1 - Default", "2", "8"}); got != "1 - Default" {
		t.Errorf("no current-size entry: got %q, want the default entry as fallback", got)
	}
	if got := pickCurrentSwapSize([]string{"2", "6 - Current Size"}); got != "6 - Current Size" {
		t.Errorf("got %q, want the current-size entry", got)
	}
	if got := pickCurrentSwapSize(nil); got != "" {
		t.Errorf("empty options: got %q, want empty", got)
	}
}
