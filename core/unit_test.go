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

package core

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// TestGetUnitStatusReadsWithoutRoot covers the "hugepages / shared memory in THP
// show as unknown and cannot be changed" report. Every unit matrix parameter is
// world-readable, so reading a status must never depend on sudo: in web mode
// there is no way to supply a password at all, and before this fix the read
// failed with ErrNoSudoPassword, which the UI displayed as "off"/unknown.
func TestGetUnitStatusReadsWithoutRoot(t *testing.T) {
	dir := t.TempDir()
	listFile := filepath.Join(dir, "enabled")
	if err := os.WriteFile(listFile, []byte("always madvise [never]\n"), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	scalarFile := filepath.Join(dir, "compaction")
	if err := os.WriteFile(scalarFile, []byte("0\n"), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	restore := UnitMatrix
	UnitMatrix = map[string]string{"probe_list": listFile, "probe_scalar": scalarFile}
	defer func() { UnitMatrix = restore }()

	// forceSudoForTest makes any privileged fallback a real (recorded) sudo call,
	// so "the read never reaches for root" is asserted even when the suite runs as
	// root, where runSudo would otherwise exec the command directly.
	restoreSudo := useSudoAlways
	useSudoAlways = true
	defer func() { useSudoAlways = restoreSudo }()

	record := filepath.Join(dir, "sudo-calls.txt")
	installRecordingSudo(t, record)

	e := NewEngine(io.Discard, io.Discard)
	e.Password = ""

	if got, err := e.getUnitStatus("probe_list"); err != nil || got != "never" {
		t.Errorf("getUnitStatus(list) = %q, %v; want %q with no error", got, err, "never")
	}
	if got, err := e.getUnitStatus("probe_scalar"); err != nil || got != "0" {
		t.Errorf("getUnitStatus(scalar) = %q, %v; want %q with no error", got, err, "0")
	}
	if calls := recordedSudoCalls(t, record); len(calls) != 0 {
		t.Errorf("reading status started sudo %d time(s): %v", len(calls), calls)
	}

	if _, err := e.getUnitStatus("no_such_parameter"); err == nil {
		t.Error("want an error for an unknown parameter name")
	}
}

// TestGetUnitStatusFallsBackToSudoWhenDenied keeps the sudo fallback for
// parameters that are not readable by an ordinary user.
func TestGetUnitStatusFallsBackToSudoWhenDenied(t *testing.T) {
	dir := t.TempDir()
	private := filepath.Join(dir, "private")
	if err := os.WriteFile(private, []byte("ignored\n"), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	restoreReader := unitFileReader
	unitFileReader = func(string) ([]byte, error) {
		return nil, &os.PathError{Op: "open", Path: private, Err: syscall.EACCES}
	}
	defer func() { unitFileReader = restoreReader }()

	restoreMatrix := UnitMatrix
	UnitMatrix = map[string]string{"probe": private}
	defer func() { UnitMatrix = restoreMatrix }()

	record := filepath.Join(dir, "sudo-calls.txt")
	installFakeSudo(t, record, "pw")

	restoreSudo := useSudoAlways
	useSudoAlways = true
	defer func() { useSudoAlways = restoreSudo }()

	e := NewEngine(io.Discard, io.Discard)
	e.Password = "pw"
	got, err := e.getUnitStatus("probe")
	if err != nil {
		t.Fatalf("getUnitStatus: %v", err)
	}
	if got != "ignored" {
		t.Errorf("getUnitStatus = %q, want %q (the sudo fallback output was not parsed)", got, "ignored")
	}
	calls := recordedSudoCalls(t, record)
	if len(calls) != 1 || !strings.Contains(calls[0], "cat") {
		t.Errorf("sudo calls = %v, want one `cat`", calls)
	}
}

// TestParseUnitValue locks the bracket parsing that picks the active value out of
// a list-style sysfs file.
func TestParseUnitValue(t *testing.T) {
	cases := map[string]string{
		"always madvise [never]\n": "never",
		"[always] madvise never\n": "always",
		"advise [never]\n":         "never",
		"20\n":                     "20",
		"  1  \n":                  "1",
		"never\n":                  "never",
	}
	for raw, want := range cases {
		if got := parseUnitValue(raw); got != want {
			t.Errorf("parseUnitValue(%q) = %q, want %q", raw, got, want)
		}
	}
}

// TestRemoveFileToleratesAMissingPath covers the revert path: Revert* removes a
// tmpfiles.d conf that is often simply not present, and `rm` without -f turned
// that into a hard error, so a revert that actually worked was reported as one
// that failed.
func TestRemoveFileToleratesAMissingPath(t *testing.T) {
	var logs bytes.Buffer
	e := NewEngine(&logs, &logs)
	if err := e.removeFile(filepath.Join(t.TempDir(), "definitely-missing.conf")); err != nil {
		t.Errorf("removeFile(missing) = %v, want nil (rm -f semantics)", err)
	}
}

// TestSetUnitValueReportsFailures covers the other half of the report: a write
// that really fails must be visible, and a write that succeeds must not be
// reported as a failure.
func TestSetUnitValueReportsFailures(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "param")
	if err := os.WriteFile(target, []byte("0"), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	restore := UnitMatrix
	UnitMatrix = map[string]string{"probe": target}
	defer func() { UnitMatrix = restore }()

	var logs bytes.Buffer
	e := NewEngine(&logs, &logs)
	if err := e.setUnitValue("probe", "1"); err != nil {
		t.Fatalf("setUnitValue on a writable parameter: %v", err)
	}
	if data, _ := os.ReadFile(target); string(data) != "1" {
		t.Errorf("parameter = %q, want %q", string(data), "1")
	}

	// A directory that does not exist: tee cannot create the file, which is the
	// same class of failure as the kernel rejecting the write (EINVAL).
	missing := filepath.Join(dir, "no-such-dir", "param")
	UnitMatrix["probe"] = missing
	if err := e.setUnitValue("probe", "1"); err == nil || !strings.Contains(err.Error(), "No such file") {
		t.Errorf("setUnitValue on a missing parameter = %v, want the tee error", err)
	}
}

// TestGetUnitStatesReadsWithoutRootAndFlagsRecommended covers the single-pass
// read the UI refreshes with: no sudo, the bracketed sysfs answer, and the
// "is this the recommended value" decision made from the value itself.
func TestGetUnitStatesReadsWithoutRootAndFlagsRecommended(t *testing.T) {
	var plainReads int
	restoreReader, restoreSudo := unitFileReader, useSudoAlways
	defer func() { unitFileReader, useSudoAlways = restoreReader, restoreSudo }()

	// A sudo stand-in that fails the test the moment anything tries to use it.
	sudoRecord := filepath.Join(t.TempDir(), "sudo-calls")
	installRecordingSudo(t, sudoRecord)

	useSudoAlways = false
	unitFileReader = func(path string) ([]byte, error) {
		plainReads++
		switch path {
		case "/sys/kernel/mm/transparent_hugepage/enabled":
			return []byte("always madvise [always]\n"), nil
		case "/sys/kernel/mm/transparent_hugepage/shmem_enabled":
			return []byte("always within_size [advise] never deny force\n"), nil
		case "/proc/sys/vm/compaction_proactiveness":
			return []byte("20\n"), nil
		case "/sys/kernel/mm/transparent_hugepage/khugepaged/defrag":
			return []byte("1\n"), nil
		}
		return nil, os.ErrNotExist // page_lock_unfairness: not on this kernel
	}

	e := NewEngine(io.Discard, io.Discard)
	states := e.GetUnitStates()

	if plainReads != len(unitTunableParams()) {
		t.Fatalf("reading unit states issued %d plain read(s), want one per parameter (%d)",
			plainReads, len(unitTunableParams()))
	}
	if b, _ := os.ReadFile(sudoRecord); len(b) != 0 {
		t.Fatalf("reading unit states went through sudo: %s", b)
	}
	if got := states["hugepages"]; got.Value != "always" || !got.IsRecommended {
		t.Fatalf("hugepages = %+v, want value \"always\" and recommended", got)
	}
	if got := states["shmem_enabled"]; got.Value != "advise" || !got.IsRecommended {
		t.Fatalf("shmem_enabled = %+v, want value \"advise\" and recommended", got)
	}
	if got := states["compaction_proactiveness"]; got.Value != "20" || got.IsRecommended {
		t.Fatalf("compaction_proactiveness = %+v, want value \"20\" and NOT recommended", got)
	}
	if got := states["page_lock_unfairness"]; got.Value != ValueUnknown || got.IsRecommended {
		t.Fatalf("unreadable parameter = %+v, want value %q and NOT recommended", got, ValueUnknown)
	}
}

// TestGetStatusSummaryPublishesRawUnitValues is what the UI needs to show the
// actual value instead of only a checkmark.
func TestGetStatusSummaryPublishesRawUnitValues(t *testing.T) {
	restoreReader := unitFileReader
	defer func() { unitFileReader = restoreReader }()
	unitFileReader = func(path string) ([]byte, error) {
		switch path {
		case "/sys/kernel/mm/transparent_hugepage/enabled":
			return []byte("[always] madvise never\n"), nil
		case "/sys/kernel/mm/transparent_hugepage/shmem_enabled":
			return []byte("always within_size [advise] never deny force\n"), nil
		case "/proc/sys/vm/compaction_proactiveness":
			return []byte("20\n"), nil
		case "/sys/kernel/mm/transparent_hugepage/khugepaged/defrag":
			return []byte("1\n"), nil
		case "/proc/sys/vm/page_lock_unfairness":
			return []byte("1\n"), nil
		}
		return nil, os.ErrNotExist
	}

	e := NewEngine(io.Discard, io.Discard)
	sum := e.GetStatusSummary()
	for key, want := range map[string]string{
		"HugePagesValue":               "always",
		"ShMemValue":                   "advise",
		"CompactionProactivenessValue": "20",
		"DefragValue":                  "1",
		"PageLockUnfairnessValue":      "1",
	} {
		if sum[key] != want {
			t.Errorf("summary[%q] = %q, want %q", key, sum[key], want)
		}
	}
	// The boolean answer must agree with the value it sits next to.
	if sum["HugePages"] != "true" || sum["CompactionProactiveness"] != "false" || sum["PageLockUnfairness"] != "true" {
		t.Errorf("booleans disagree with the published values: %v", sum)
	}
}

// TestRecommendedValuesMatchWhatSteamOSShips guards the other half of P0.10:
// the verdict the UI shows must match the machine. Measured on a Deck booted
// 2026-10-27 with no tmpfiles entries of ours and no transparent_hugepage= on
// the kernel command line, the five parameters came up as always / advise / 20 /
// 5 / 1. Two of them are already at what this app calls recommended, so a
// freshly booted Deck must show those two rows as recommended - and must not
// offer a "fix" for something Valve already configured.
func TestRecommendedValuesMatchWhatSteamOSShips(t *testing.T) {
	if RecommendedHugePages != "always" {
		t.Fatalf("RecommendedHugePages = %q, want \"always\" (the value SteamOS boots with)", RecommendedHugePages)
	}
	if RecommendedShMem != "advise" {
		t.Fatalf("RecommendedShMem = %q, want \"advise\" (the value SteamOS boots with)", RecommendedShMem)
	}
	if RecommendedCompactionProactiveness != "0" || RecommendedHugePageDefrag != "0" || RecommendedPageLockUnfairness != "1" {
		t.Fatalf("recommended trio = %q/%q/%q, want 0/0/1",
			RecommendedCompactionProactiveness, RecommendedHugePageDefrag, RecommendedPageLockUnfairness)
	}
	// The stock/default side deliberately stays at the upstream kernel numbers.
	// If that ever changes, the CLI's `disable`/`stock` commands stop doing
	// anything for those parameters, which is why the pair is pinned here.
	if DefaultHugePages != "madvise" || DefaultShMem != "never" {
		t.Fatalf("stock values = %q/%q, want madvise/never (upstream kernel defaults; see core/config.go)",
			DefaultHugePages, DefaultShMem)
	}
}

// TestRevertHugePagesEndsOnTheDefaultNoMatterTheOrder runs the revert path
// against a stand-in parameter file: the value left in the file must be the boot
// default (so nothing has to be re-applied after a reboot), and the persistence
// conf must be dropped with `rm -f` through sudo with -k.
func TestRevertHugePagesWritesTheDefaultAndDropsTheConf(t *testing.T) {
	dir := t.TempDir()
	paramFile := filepath.Join(dir, "enabled")
	// Start from THP being on, which is what a Deck boots with. Seeding the file
	// with the value revert is supposed to write would make the assertion pass even
	// if revert never wrote anything.
	if err := os.WriteFile(paramFile, []byte("always\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	confDir := filepath.Join(dir, "tmpfiles.d")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatal(err)
	}

	record := filepath.Join(dir, "sudo-calls")
	installFakeSudo(t, record, "pw")
	restore := useSudoAlways
	useSudoAlways = true
	defer func() { useSudoAlways = restore }()

	restoreMatrix, restoreTmpFiles := UnitMatrix["hugepages"], TmpFilesRoot
	UnitMatrix["hugepages"] = paramFile
	TmpFilesRoot = confDir
	defer func() {
		UnitMatrix["hugepages"] = restoreMatrix
		TmpFilesRoot = restoreTmpFiles
	}()

	e := NewEngine(io.Discard, io.Discard)
	e.Password = "pw"
	if err := e.RevertHugePages(); err != nil {
		t.Fatalf("RevertHugePages: %v", err)
	}

	got, err := os.ReadFile(paramFile)
	if err != nil {
		t.Fatal(err)
	}
	// The fake sudo hands `tee` exactly what sudo left on stdin, which is how the
	// real tee stores it; compare through the same parser the app reads with.
	if parseUnitValue(string(got)) != DefaultHugePages {
		t.Fatalf("parameter file = %q, want %q (the stock value revert must write)", string(got), DefaultHugePages)
	}
	calls := recordedSudoCalls(t, record)
	var removed, sawKey bool
	for _, line := range calls {
		if strings.Contains(line, "rm -f "+filepath.Join(confDir, "hugepages.conf")) {
			removed = true
		}
		if strings.Contains(line, " -k ") {
			sawKey = true
		}
	}
	if !removed {
		t.Fatalf("revert did not remove the persistence conf with rm -f; sudo calls: %v", calls)
	}
	if !sawKey {
		t.Fatalf("sudo was called without -k: %v", calls)
	}
}
