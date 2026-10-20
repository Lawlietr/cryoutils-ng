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
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunSudoWithoutPasswordNeverSpawnsSudo locks the fix for the native UI being
// asked twice: a sudo started without -S inherits the terminal and prompts on the
// tty, so the user typed the password in the CLI and then again in the UI dialog.
// With no password stored, runSudo must fail locally instead of starting sudo.
func TestRunSudoWithoutPasswordNeverSpawnsSudo(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: runSudo does not wrap the command in sudo")
	}
	record := filepath.Join(t.TempDir(), "sudo-calls.txt")
	installFakeSudo(t, record)

	e := NewEngine(io.Discard, io.Discard)
	e.Password = ""

	out, err := e.runSudo("cat", "/proc/swaps")
	if !errors.Is(err, ErrNoSudoPassword) {
		t.Fatalf("runSudo error = %v, want ErrNoSudoPassword", err)
	}
	if out != "" {
		t.Errorf("runSudo output = %q, want empty", out)
	}
	if calls := recordedSudoCalls(t, record); len(calls) != 0 {
		t.Errorf("sudo was started %d time(s) without a password: %v", len(calls), calls)
	}

	// The status readers must fail the same way rather than prompting.
	if _, err := e.getUnitStatus("swappiness"); !errors.Is(err, ErrNoSudoPassword) {
		t.Errorf("getUnitStatus error = %v, want ErrNoSudoPassword", err)
	}
	if err := e.setUnitValue("swappiness", "60"); !errors.Is(err, ErrNoSudoPassword) {
		t.Errorf("setUnitValue error = %v, want ErrNoSudoPassword", err)
	}
	if calls := recordedSudoCalls(t, record); len(calls) != 0 {
		t.Errorf("status reads started sudo %d time(s): %v", len(calls), calls)
	}
}

// TestRunSudoWithInputFeedsStdin covers the `sudo tee` path, where the value to
// write is delivered on stdin after the password line.
func TestRunSudoWithInputFeedsStdin(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("needs root: the non-root path authenticates through sudo")
	}
	target := filepath.Join(t.TempDir(), "param")
	if err := os.WriteFile(target, []byte("0"), 0o644); err != nil {
		t.Fatalf("writing target: %v", err)
	}

	e := NewEngine(io.Discard, io.Discard)
	if _, err := e.runSudoWithInput("140\n", "tee", target); err != nil {
		t.Fatalf("runSudoWithInput: %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading target: %v", err)
	}
	if string(data) != "140\n" {
		t.Errorf("target = %q, want %q (stdin was not delivered)", string(data), "140\n")
	}
}

// installFakeSudo puts a sudo stand-in first on PATH that records its arguments,
// so a test can prove whether sudo was started at all.
func installFakeSudo(t *testing.T, record string) {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + record + "\nexit 0\n"
	path := filepath.Join(dir, "sudo")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("writing fake sudo: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func recordedSudoCalls(t *testing.T, record string) []string {
	t.Helper()
	data, err := os.ReadFile(record)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("reading %s: %v", record, err)
	}
	var calls []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line != "" {
			calls = append(calls, line)
		}
	}
	return calls
}
