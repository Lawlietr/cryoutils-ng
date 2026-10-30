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
