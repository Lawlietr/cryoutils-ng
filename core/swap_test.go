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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// steamOSSwapTable is a real /proc/swaps snapshot from a Steam Deck (SteamOS 3.6+):
// zram at priority 100 coexisting with the swap file at priority -2.
const steamOSSwapTable = `Filename				Type		Size		Used		Priority
/dev/zram0                              partition	7579644		7577844		100
/home/swapfile                          file		16777212	8996248		-2
`

// idleZramTable is the same device layout with both swap areas empty.
const idleZramTable = `Filename				Type		Size		Used		Priority
/dev/zram0                              partition	7579644		0		100
/home/swapfile                          file		16777212	0		-2
`

func TestParseSwapEntries(t *testing.T) {
	entries := parseSwapEntries(steamOSSwapTable)
	if len(entries) != 2 {
		t.Fatalf("parseSwapEntries() returned %d entries, want 2", len(entries))
	}

	want := []swapEntry{
		{Filename: "/dev/zram0", Type: "partition", SizeKB: 7579644, UsedKB: 7577844, Priority: 100},
		{Filename: "/home/swapfile", Type: "file", SizeKB: 16777212, UsedKB: 8996248, Priority: -2},
	}
	for i, w := range want {
		if entries[i] != w {
			t.Errorf("entry[%d] = %+v, want %+v", i, entries[i], w)
		}
	}
}

func TestParseSwapEntriesIgnoresMalformedRows(t *testing.T) {
	table := `Filename				Type		Size		Used		Priority
/dev/zram0                              partition	7579644		7577844		100
garbage row
/home/swapfile                          file		notanumber	0		-2
`
	entries := parseSwapEntries(table)
	if len(entries) != 1 {
		t.Fatalf("parseSwapEntries() returned %d entries, want 1", len(entries))
	}
	if entries[0].Filename != "/dev/zram0" {
		t.Errorf("entry[0].Filename = %q, want %q", entries[0].Filename, "/dev/zram0")
	}
}

// TestParseSwapFileLocationSkipsZram is the regression test for the original
// bug where the /dev/ row was mishandled: /dev/zram0 must be skipped, not
// returned as the swap file and not an error.
func TestParseSwapFileLocationSkipsZram(t *testing.T) {
	tests := []struct {
		name  string
		table string
		want  string
	}{
		{"SteamOS zram + swapfile", steamOSSwapTable, "/home/swapfile"},
		{"zram only", `Filename				Type		Size		Used		Priority
/dev/zram0                              partition	7579644		0		100
`, ""},
		{"swapfile only", `Filename				Type		Size		Used		Priority
/swapfile                               file		2097148		0		-2
`, "/swapfile"},
		{"header only", "Filename\t\t\t\tType\t\tSize\t\tUsed\t\tPriority\n", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSwapFileLocation(parseSwapEntries(tt.table)); got != tt.want {
				t.Errorf("parseSwapFileLocation() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestZramActiveIn(t *testing.T) {
	if !zramActiveIn(parseSwapEntries(steamOSSwapTable)) {
		t.Errorf("zramActiveIn(SteamOS table) = false, want true")
	}
	withoutZram := `Filename				Type		Size		Used		Priority
/home/swapfile                          file		16777212	0		-2
`
	if zramActiveIn(parseSwapEntries(withoutZram)) {
		t.Errorf("zramActiveIn(no zram) = true, want false")
	}
}

func TestParseMemAvailableKB(t *testing.T) {
	meminfo := `MemTotal:       15159816 kB
MemFree:         6725916 kB
MemAvailable:    6932000 kB
SwapTotal:      24356856 kB
SwapFree:       24356856 kB
`
	if got := parseMemAvailableKB(meminfo); got != 6932000 {
		t.Errorf("parseMemAvailableKB() = %d, want 6932000", got)
	}
	if got := parseMemAvailableKB("MemTotal: 1 kB\n"); got != -1 {
		t.Errorf("parseMemAvailableKB(missing) = %d, want -1", got)
	}
}

func TestSwapMigrationFit(t *testing.T) {
	const margin = int64(512 * 1024) // 512MB, matches SwapMigrationMarginKB

	tests := []struct {
		name           string
		table          string
		target         string
		memAvailableKB int64
		wantFits       bool
	}{
		{
			// The real situation measured on the Deck before a resize:
			// ~8.6 GB held by the swap file, zram nearly full, ~6.6 GB of RAM
			// available. Turning the swap file off here would have pushed the
			// machine into earlyoom territory.
			name:           "SteamOS tight - zram full and little RAM",
			table:          steamOSSwapTable,
			target:         "/home/swapfile",
			memAvailableKB: 6932000,
			wantFits:       false,
		},
		{
			// Idle swap: the file holds nothing, so nothing has to migrate.
			name:           "idle swap areas",
			table:          idleZramTable,
			target:         "/home/swapfile",
			memAvailableKB: 100000,
			wantFits:       true,
		},
		{
			// 1 GB held by the swap file, 7.2 GB of free zram to absorb it.
			name: "zram free absorbs the file",
			table: `Filename				Type		Size		Used		Priority
/dev/zram0                              partition	7579644		0		100
/home/swapfile                          file		16777212	1048576		-2
`,
			target:         "/home/swapfile",
			memAvailableKB: 1048576,
			wantFits:       true,
		},
		{
			// zram gone (the P0.7 symptom): 6 GB in the file, 500 MB of RAM.
			name: "zram missing and little RAM must refuse",
			table: `Filename				Type		Size		Used		Priority
/home/swapfile                          file		16777212	6291456		-2
`,
			target:         "/home/swapfile",
			memAvailableKB: 512000,
			wantFits:       false,
		},
		{
			// Unknown MemAvailable: destination is the other swap devices only.
			name:           "unknown MemAvailable falls back to swap room",
			table:          idleZramTable,
			target:         "/home/swapfile",
			memAvailableKB: -1,
			wantFits:       true,
		},
		{
			name:           "target not in table needs no migration",
			table:          idleZramTable,
			target:         "/home/does-not-exist",
			memAvailableKB: 0,
			wantFits:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			needed, dest, fits := swapMigrationFit(parseSwapEntries(tt.table), tt.target, tt.memAvailableKB, margin)
			if fits != tt.wantFits {
				t.Errorf("swapMigrationFit() fits = %v, want %v (needed=%d kB, dest=%d kB)", fits, tt.wantFits, needed, dest)
			}
			if needed < 0 || dest < 0 {
				t.Errorf("swapMigrationFit() returned negative values: needed=%d dest=%d", needed, dest)
			}
		})
	}
}

// TestSwapMigrationFitSteamOSNumbers pins the arithmetic for the snapshot above,
// documenting why the resize must be refused before swapoff, not after.
func TestSwapMigrationFitSteamOSNumbers(t *testing.T) {
	const margin = int64(512 * 1024)
	needed, dest, fits := swapMigrationFit(parseSwapEntries(steamOSSwapTable), "/home/swapfile", 6932000, margin)

	if needed != 8996248 {
		t.Errorf("needed = %d kB, want 8996248 kB", needed)
	}
	// 7579644 - 7577844 free on zram (1800 kB) + 6932000 kB of RAM.
	if dest != 6933800 {
		t.Errorf("dest = %d kB, want 6933800 kB", dest)
	}
	if fits {
		t.Errorf("fits = true, want false: %d kB needed + %d kB margin > %d kB destination", needed, margin, dest)
	}
}

// swapTools are the external commands a swap resize runs.
var swapTools = []string{"swapoff", "swapon", "mkswap", "dd", "chmod"}

// installFakeSwapTools puts stand-in swap tools at the front of PATH and records
// every invocation ("<tool> <args>") into recordPath. When failOn names a tool,
// that tool writes to stderr and exits non-zero, which lets the tests assert that
// command output reaches the returned error.
func installFakeSwapTools(t *testing.T, recordPath string, failOn string) {
	t.Helper()
	binDir := t.TempDir()
	for _, tool := range swapTools {
		body := "#!/bin/sh\necho \"" + tool + " $*\" >> " + recordPath + "\n"
		if tool == failOn {
			body += "echo \"simulated failure of " + tool + "\" >&2\nexit 1\n"
		} else {
			body += "exit 0\n"
		}
		path := filepath.Join(binDir, tool)
		if err := os.WriteFile(path, []byte(body), 0755); err != nil {
			t.Fatalf("writing fake %s: %v", tool, err)
		}
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func recordedCalls(t *testing.T, recordPath string) []string {
	t.Helper()
	data, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("reading recorded calls: %v", err)
	}
	var calls []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line != "" {
			calls = append(calls, line)
		}
	}
	return calls
}

func indexOfCall(calls []string, prefix string) int {
	for i, call := range calls {
		if strings.HasPrefix(call, prefix) {
			return i
		}
	}
	return -1
}

// TestStatusSummaryReportsSwapFileLocation guards the display regression where
// `status` printed an empty SwapFile: the map literal read SwapFileLocation in
// source order, before the SwapSizeGB entry resolved it.
func TestStatusSummaryReportsSwapFileLocation(t *testing.T) {
	if entries, err := readSwapEntries(); err == nil && parseSwapFileLocation(entries) != "" {
		t.Skipf("this machine has a swap file (%s); the default-location path is not reachable here", parseSwapFileLocation(entries))
	}

	swapFile := filepath.Join(t.TempDir(), "swapfile")
	if err := os.WriteFile(swapFile, []byte("x"), 0600); err != nil {
		t.Fatalf("creating stand-in swap file: %v", err)
	}

	originalDefault := DefaultSwapFileLocation
	DefaultSwapFileLocation = swapFile
	defer func() { DefaultSwapFileLocation = originalDefault }()

	e := NewEngine(io.Discard, io.Discard)
	summary := e.GetStatusSummary()
	if summary["SwapFile"] != swapFile {
		t.Errorf("status SwapFile = %q, want %q", summary["SwapFile"], swapFile)
	}
}

// TestChangeSwapSizeNeverTurnsOffOtherSwap is the regression test for the bug in
// docs/zram-swapoff-reset.md: the resize may only switch off its own swap file.
// "swapoff -a" makes systemd tear the zram device down outright
// (zram-generator --reset-device), which is why an unscoped swapoff is not
// merely inelegant but destructive.
func TestChangeSwapSizeNeverTurnsOffOtherSwap(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("needs root: the non-root path authenticates through sudo")
	}
	recordPath := filepath.Join(t.TempDir(), "calls.txt")
	installFakeSwapTools(t, recordPath, "")

	swapFile := filepath.Join(t.TempDir(), "swapfile")
	if err := os.WriteFile(swapFile, []byte("x"), 0600); err != nil {
		t.Fatalf("creating stand-in swap file: %v", err)
	}

	e := NewEngine(io.Discard, io.Discard)
	e.SwapFileLocation = swapFile

	if err := e.ChangeSwapSize(3); err != nil {
		t.Fatalf("ChangeSwapSize() returned %v, want success", err)
	}

	calls := recordedCalls(t, recordPath)

	for _, call := range calls {
		if strings.Contains(call, "-a") {
			t.Errorf("recorded %q: an unscoped swapoff would destroy zram on SteamOS", call)
		}
		if strings.Contains(call, "zram") {
			t.Errorf("recorded %q: zram must never be touched by a resize", call)
		}
	}

	swapoffs := 0
	for _, call := range calls {
		if strings.HasPrefix(call, "swapoff ") {
			swapoffs++
			if call != "swapoff "+swapFile {
				t.Errorf("swapoff call = %q, want %q", call, "swapoff "+swapFile)
			}
		}
	}
	if swapoffs != 1 {
		t.Errorf("swapoff invoked %d times, want exactly 1", swapoffs)
	}

	if got := indexOfCall(calls, "dd "); got < 0 {
		t.Errorf("no dd call recorded; calls = %v", calls)
	} else if !strings.Contains(calls[got], "of="+swapFile) || !strings.Contains(calls[got], "count=3") {
		t.Errorf("dd call = %q, want of=%s and count=3", calls[got], swapFile)
	}

	if offAt, onAt := indexOfCall(calls, "swapoff "), indexOfCall(calls, "swapon "); offAt < 0 || onAt < 0 || offAt > onAt {
		t.Errorf("expected swapoff before swapon, calls = %v", calls)
	}
	if mkswapAt, onAt := indexOfCall(calls, "mkswap "), indexOfCall(calls, "swapon "); mkswapAt < 0 || onAt < 0 || mkswapAt > onAt {
		t.Errorf("expected mkswap before swapon, calls = %v", calls)
	}
}

// TestChangeSwapSizeResolvesLocationItself covers the CLI path, which never
// queries status before resizing: SwapFileLocation is empty there, and an empty
// location previously produced "dd of=" against nothing. DefaultSwapFileLocation
// is redirected to a temp file so the test never touches a real swap file.
func TestChangeSwapSizeResolvesLocationItself(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("needs root: the non-root path authenticates through sudo")
	}
	recordPath := filepath.Join(t.TempDir(), "calls.txt")
	installFakeSwapTools(t, recordPath, "")

	swapFile := filepath.Join(t.TempDir(), "swapfile")
	if err := os.WriteFile(swapFile, []byte("x"), 0600); err != nil {
		t.Fatalf("creating stand-in swap file: %v", err)
	}

	// On a machine that actually has a swap file, discovery finds that one first
	// and the default-location fallback is not the path under test.
	if entries, err := readSwapEntries(); err == nil && parseSwapFileLocation(entries) != "" {
		t.Skipf("this machine has a swap file (%s); the default-location fallback is not reachable here", parseSwapFileLocation(entries))
	}

	originalDefault := DefaultSwapFileLocation
	DefaultSwapFileLocation = swapFile
	defer func() { DefaultSwapFileLocation = originalDefault }()

	e := NewEngine(io.Discard, io.Discard)
	if err := e.ChangeSwapSize(2); err != nil {
		t.Fatalf("ChangeSwapSize() returned %v, want success via the default location fallback", err)
	}
	if e.SwapFileLocation != swapFile {
		t.Errorf("SwapFileLocation = %q, want %q", e.SwapFileLocation, swapFile)
	}

	calls := recordedCalls(t, recordPath)
	ddAt := indexOfCall(calls, "dd ")
	if ddAt < 0 {
		t.Fatalf("no dd call recorded; calls = %v", calls)
	}
	if !strings.Contains(calls[ddAt], "of="+swapFile) {
		t.Errorf("dd = %q, want of=%s (an empty of= is the bug this guards)", calls[ddAt], swapFile)
	}
}

// TestChangeSwapSizeSurfacesCommandStderr guards the diagnosability regression:
// every swap failure used to be flattened into a fixed string.
func TestChangeSwapSizeSurfacesCommandStderr(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("needs root: the non-root path authenticates through sudo")
	}
	recordPath := filepath.Join(t.TempDir(), "calls.txt")
	installFakeSwapTools(t, recordPath, "swapoff")

	swapFile := filepath.Join(t.TempDir(), "swapfile")
	if err := os.WriteFile(swapFile, []byte("x"), 0600); err != nil {
		t.Fatalf("creating stand-in swap file: %v", err)
	}

	e := NewEngine(io.Discard, io.Discard)
	e.SwapFileLocation = swapFile

	err := e.ChangeSwapSize(2)
	if err == nil {
		t.Fatalf("ChangeSwapSize() succeeded, want failure")
	}
	if !strings.Contains(err.Error(), "simulated failure of swapoff") {
		t.Errorf("error %q does not carry the failing command's stderr", err.Error())
	}
}

// TestFailedResizeTriesToLeaveSwapOn checks that a resize which fails part way
// through does not abandon the machine with no swap at all.
func TestFailedResizeTriesToLeaveSwapOn(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("needs root: the non-root path authenticates through sudo")
	}
	recordPath := filepath.Join(t.TempDir(), "calls.txt")
	installFakeSwapTools(t, recordPath, "dd")

	swapFile := filepath.Join(t.TempDir(), "swapfile")
	if err := os.WriteFile(swapFile, []byte("x"), 0600); err != nil {
		t.Fatalf("creating stand-in swap file: %v", err)
	}

	e := NewEngine(io.Discard, io.Discard)
	e.SwapFileLocation = swapFile

	if err := e.ChangeSwapSize(2); err == nil {
		t.Fatalf("ChangeSwapSize() succeeded, want failure from the simulated dd")
	}

	calls := recordedCalls(t, recordPath)
	ddAt := indexOfCall(calls, "dd ")
	if mkswapAt := indexOfCall(calls, "mkswap "); mkswapAt < 0 || ddAt < 0 || mkswapAt < ddAt {
		t.Errorf("expected a best-effort mkswap after the failed dd, calls = %v", calls)
	}
	if onAt := indexOfCall(calls, "swapon "); onAt < 0 || ddAt < 0 || onAt < ddAt {
		t.Errorf("expected a best-effort swapon after the failed dd, calls = %v", calls)
	}
}
