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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"cryoutils-ng/core"
	"cryoutils-ng/i18n"
)

// pointUnitFilesAt redirects the five memory parameters at files this test
// controls, and returns the restore function plus a writer for the value file.
func pointUnitFilesAt(t *testing.T) (write func(param, contents string), restore func()) {
	t.Helper()

	params := []string{"hugepages", "shmem_enabled", "compaction_proactiveness", "defrag", "page_lock_unfairness"}
	dir := t.TempDir()

	previous := make(map[string]string, len(params))
	for _, param := range params {
		previous[param] = core.UnitMatrix[param]
		core.UnitMatrix[param] = filepath.Join(dir, param)
	}
	restore = func() {
		for _, param := range params {
			core.UnitMatrix[param] = previous[param]
		}
	}
	write = func(param, contents string) {
		if err := os.WriteFile(filepath.Join(dir, param), []byte(contents), 0o644); err != nil {
			t.Fatalf("writing %s: %v", param, err)
		}
	}
	return write, restore
}

// TestMemorySectionFollowsTheKernel covers the P0.10 fix on the Deck: the memory
// card is built once, so its checkboxes must be refreshed in place, and every
// row must show the value the kernel actually holds. Without this the tick can
// disagree with the system (external change, or a failed write) until the app is
// restarted - which is how "Huge Pages keeps unchecking itself" looked.
func TestMemorySectionFollowsTheKernel(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	write, restore := pointUnitFilesAt(t)
	defer restore()

	write("hugepages", "[always] madvise never\n")
	write("shmem_enabled", "always within_size [never] deny force\n")
	write("compaction_proactiveness", "20\n")
	write("defrag", "1\n")
	write("page_lock_unfairness", "5\n")

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
	// Any real toggle path would need the password dialog; make sure a stray
	// toggle would be loud instead of segfaulting on a nil callback.
	c.refreshStatus = func() { t.Fatal("refreshStatus must not run in this test") }

	section := sectionsMemory(c)
	card, ok := section.(*widget.Card)
	if !ok {
		t.Fatalf("sectionsMemory returned %T, want *widget.Card", section)
	}
	if len(c.refreshers) != 1 {
		t.Fatalf("memory section registered %d refresher(s), want 1", len(c.refreshers))
	}

	box, ok := card.Content.(*fyne.Container)
	if !ok {
		t.Fatalf("card content is %T, want a container", card.Content)
	}
	var checks []*widget.Check
	for _, o := range box.Objects {
		if cb, ok := o.(*widget.Check); ok {
			checks = append(checks, cb)
		}
	}
	if len(checks) != 5 {
		t.Fatalf("memory card has %d checkbox(es), want 5", len(checks))
	}

	// Initial state comes from the kernel, not from a stored UI state.
	if !checks[0].Checked {
		t.Fatalf("Huge Pages starts unchecked although the parameter reads [always]: label %q", checks[0].Text)
	}
	if checks[1].Checked {
		t.Fatalf("Shared memory starts checked although the parameter reads [never]: label %q", checks[1].Text)
	}
	// The value is visible, so "recommended" and "stock default" cannot be confused.
	if !strings.Contains(checks[0].Text, "always") {
		t.Fatalf("Huge Pages label %q does not show the current value", checks[0].Text)
	}

	// Something else changed the system: the refresher (5s ticker) has to follow.
	write("hugepages", "always [madvise] never\n")
	write("shmem_enabled", "always within_size [advise] never deny force\n")
	for range 3 {
		c.refreshers[0]()
	}
	if checks[0].Checked {
		t.Fatalf("Huge Pages stayed checked after the kernel moved to madvise: label %q", checks[0].Text)
	}
	if !strings.Contains(checks[0].Text, "madvise") {
		t.Fatalf("Huge Pages label %q did not follow the kernel value", checks[0].Text)
	}
	if !checks[1].Checked {
		t.Fatalf("Shared memory stayed unchecked after the kernel moved to advise: label %q", checks[1].Text)
	}

	// An unreadable parameter must read as unknown (unchecked, labelled as such),
	// never as a confident "off".
	if err := os.Remove(core.UnitMatrix["page_lock_unfairness"]); err != nil {
		t.Fatal(err)
	}
	c.refreshers[0]()
	if checks[4].Checked {
		t.Fatalf("unreadable parameter reported as recommended: label %q", checks[4].Text)
	}
	if !strings.Contains(checks[4].Text, "Unknown") {
		t.Fatalf("unreadable parameter label %q should say Unknown", checks[4].Text)
	}
}

// installLocalFakeSudo puts a sudo stand-in first on PATH that records its
// arguments and then really runs the command, so a stray privileged write shows
// up as a rewritten file. It mirrors core's test double, including the rule that
// sudo only consumes the password line when it actually has to authenticate.
func installLocalFakeSudo(t *testing.T, record, password string) {
	t.Helper()

	dir := t.TempDir()
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> \"$FAKE_SUDO_RECORD\"\n" +
		"cmd=''; rest=''; has_k=0\n" +
		"for a in \"$@\"; do\n" +
		"  if [ \"$a\" = \"-k\" ]; then has_k=1; continue; fi\n" +
		"  if [ \"$a\" = \"-S\" ] || [ \"$a\" = \"--\" ] || [ \"$a\" = \"-p\" ]; then continue; fi\n" +
		"  if [ \"$a\" = \"\" ]; then continue; fi\n" +
		"  if [ -z \"$cmd\" ]; then cmd=$a; else rest=\"$rest \\\"$a\\\"\"; fi\n" +
		"done\n" +
		"if [ \"$has_k\" = 1 ] || [ ! -f \"$FAKE_SUDO_MARKER\" ]; then\n" +
		"  IFS= read -r pw\n" +
		"  if [ \"$pw\" != \"$FAKE_SUDO_PASSWORD\" ]; then printf 'sudo: sorry, try again.\\n' >&2; exit 1; fi\n" +
		"  : > \"$FAKE_SUDO_MARKER\"\n" +
		"fi\n" +
		"eval \"\\\"$cmd\\\" $rest\"\n"
	path := filepath.Join(dir, "sudo")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("writing fake sudo: %v", err)
	}
	t.Setenv("FAKE_SUDO_RECORD", record)
	t.Setenv("FAKE_SUDO_PASSWORD", password)
	t.Setenv("FAKE_SUDO_MARKER", filepath.Join(dir, "sudo-timestamp"))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// TestMemorySectionBuildNeverTogglesAnything is the regression test for what the
// Deck actually reported: "Huge Pages unchecks itself every time I restart the
// app". Fyne's Check.SetChecked() calls OnChanged, so pushing the real state
// into a checkbox while building the card looked like a user click and fired
// another ToggleXxx(). Launching the app therefore wrote back every parameter
// that already sat at its recommended value - the Deck's log shows a memory
// action seconds after launch with nobody clicking.
//
// Every parameter here is at its recommended value, and compaction/defrag/
// page_lock have a recommended value different from their default, so a stray
// toggle rewrites those files where the guarded build must leave them alone.
func TestMemorySectionBuildNeverTogglesAnything(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	write, restore := pointUnitFilesAt(t)
	defer restore()
	write("hugepages", "[always] madvise never\n")
	write("shmem_enabled", "always within_size [advise] never deny force\n")
	write("compaction_proactiveness", "0\n")
	write("defrag", "0\n")
	write("page_lock_unfairness", "1\n")

	record := filepath.Join(t.TempDir(), "sudo-calls")
	installLocalFakeSudo(t, record, "hunter2")

	e := core.NewEngine(io.Discard, io.Discard)
	e.Password = "hunter2"
	// Force the sudo wrapper on even though this test process may already be
	// root, so a stray toggle really goes through the recorded sudo.
	restoreSudo := core.ForceSudoWrapper()
	defer restoreSudo()

	win := app.NewWindow("test")
	c := &uiCtx{
		e:             e,
		win:           win,
		lang:          i18n.Load("en"),
		prog:          widget.NewProgressBarInfinite(),
		progLabel:     widget.NewLabel(""),
		refreshStatus: func() {},
	}
	c.auth = newAuthManager(e, win)

	sectionsMemory(c)

	// The unguarded version fired the toggle from inside the build; give any such
	// task a moment to land before deciding.
	time.Sleep(500 * time.Millisecond)

	if b, _ := os.ReadFile(record); len(b) != 0 {
		t.Fatalf("building the memory card ran privileged commands: %s", b)
	}
	for param, want := range map[string]string{
		"hugepages":                "[always] madvise never\n",
		"shmem_enabled":            "always within_size [advise] never deny force\n",
		"compaction_proactiveness": "0\n",
		"defrag":                   "0\n",
		"page_lock_unfairness":     "1\n",
	} {
		got, err := os.ReadFile(core.UnitMatrix[param])
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("%s was rewritten while merely drawing the card: %q -> %q", param, want, got)
		}
	}
}
