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
// forceSudoForTest makes the sudo-wrapped branch reachable even when the test
// suite runs as root, so this is covered on every machine.
func TestRunSudoWithoutPasswordNeverSpawnsSudo(t *testing.T) {
	restore := useSudoAlways
	useSudoAlways = true
	defer func() { useSudoAlways = restore }()

	record := filepath.Join(t.TempDir(), "sudo-calls.txt")
	installFakeSudo(t, record, "unused")

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

	// Writing a kernel parameter still needs a password and must fail locally.
	if err := e.setUnitValue("swappiness", "60"); !errors.Is(err, ErrNoSudoPassword) {
		t.Errorf("setUnitValue error = %v, want ErrNoSudoPassword", err)
	}

	// Reading must not need sudo at all: every unit matrix entry is world-readable,
	// and a failed sudo read is what made the memory parameters show up as unknown.
	if _, err := e.getUnitStatus("swappiness"); err != nil {
		t.Errorf("getUnitStatus(swappiness) error = %v, want a value read without root", err)
	}
	if calls := recordedSudoCalls(t, record); len(calls) != 0 {
		t.Errorf("status reads started sudo %d time(s): %v", len(calls), calls)
	}
}

// TestRunSudoWithInputFeedsStdin covers the `sudo tee` path, where the value to
// write is delivered on stdin.
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

// TestRunSudoWithInputNeverLeaksPasswordIntoTheCommand is the regression test for
// the memory parameters (hugepages, shared memory in THP, compaction
// proactiveness, page lock unfairness, defrag) becoming unusable.
//
// sudo only reads the password from stdin when it has to authenticate. Without
// -k the timestamp is already valid (RenewAuth/TestAuth authenticate immediately
// before every operation), so sudo skipped stdin entirely and the command
// inherited the payload starting with the plaintext password line: `tee` wrote
// that line into the parameter file, the kernel rejected it with EINVAL, tee
// exited non-zero, and every toggle was reported as a failure - while `tee`
// echoed the same password back on stdout, straight into the log file.
//
// The fake sudo below reproduces both sudo behaviours: it consumes exactly one
// stdin line when it must authenticate (always, if -k is present; otherwise only
// while the timestamp marker is missing), and otherwise hands the whole payload
// to the command.
func TestRunSudoWithInputNeverLeaksPasswordIntoTheCommand(t *testing.T) {
	restore := useSudoAlways
	useSudoAlways = true
	defer func() { useSudoAlways = restore }()

	dir := t.TempDir()
	record := filepath.Join(dir, "sudo-calls.txt")
	marker := filepath.Join(dir, "sudo-timestamp")
	password := "hunter2-not-a-real-secret"
	installFakeSudo(t, record, password)
	t.Setenv("FAKE_SUDO_MARKER", marker)

	target := filepath.Join(dir, "param")
	if err := os.WriteFile(target, []byte("never"), 0o644); err != nil {
		t.Fatalf("writing target: %v", err)
	}

	var logs bytes.Buffer
	e := NewEngine(&logs, &logs)
	e.Password = password

	// Two calls: the first authenticates and creates the timestamp marker, the
	// second runs with that timestamp cached. Both must deliver only the value.
	for i := 1; i <= 2; i++ {
		if _, err := e.runSudoWithInput("always\n", "tee", target); err != nil {
			t.Fatalf("call %d: runSudoWithInput: %v", i, err)
		}
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("call %d: reading target: %v", i, err)
		}
		if string(data) != "always\n" {
			t.Errorf("call %d: parameter file = %q, want %q (the password line leaked into the write)",
				i, string(data), "always\n")
		}
	}

	calls := recordedSudoCalls(t, record)
	if len(calls) != 2 {
		t.Fatalf("sudo calls = %v, want 2", calls)
	}
	for _, line := range calls {
		if !strings.Contains(line, "-k") {
			t.Errorf("sudo call %q is missing -k, so sudo may skip stdin and hand the password line to the command", line)
		}
		if strings.Contains(line, password) {
			t.Errorf("sudo call %q contains the password", line)
		}
	}
	if strings.Contains(logs.String(), password) {
		t.Errorf("the logs contain the sudo password: %q", logs.String())
	}
}

// TestRunSudoReturnsStdoutOnly locks the separation of the two output streams.
// sudo writes its "[sudo] password for user:" prompt to stderr; when both streams
// were merged, a status read became "[sudo] password for deck: 60", which never
// matched the expected value, so the parameter looked wrong (or unknown) forever.
func TestRunSudoReturnsStdoutOnly(t *testing.T) {
	restore := useSudoAlways
	useSudoAlways = true
	defer func() { useSudoAlways = restore }()

	dir := t.TempDir()
	record := filepath.Join(dir, "sudo-calls.txt")
	installFakeSudo(t, record, "pw")
	t.Setenv("FAKE_SUDO_MARKER", filepath.Join(dir, "sudo-timestamp"))

	var logs bytes.Buffer
	e := NewEngine(&logs, &logs)
	e.Password = "pw"

	out, err := e.runSudo("sh", "-c", "printf 'VALUE_FROM_STDOUT\\n'; printf 'PROMPT_AND_NOISE\\n' >&2")
	if err != nil {
		t.Fatalf("runSudo: %v", err)
	}
	if out != "VALUE_FROM_STDOUT\n" {
		t.Errorf("runSudo output = %q, want only stdout", out)
	}
}

// TestRunSudoFailureCarriesStderrButNotThePassword checks that a failing
// privileged command is still diagnosable (command line + its stderr) without
// exporting the sudo password into the error message and therefore into the log.
func TestRunSudoFailureCarriesStderrButNotThePassword(t *testing.T) {
	restore := useSudoAlways
	useSudoAlways = true
	defer func() { useSudoAlways = restore }()

	dir := t.TempDir()
	record := filepath.Join(dir, "sudo-calls.txt")
	password := "topsecret"
	installFakeSudo(t, record, password)
	t.Setenv("FAKE_SUDO_MARKER", filepath.Join(dir, "sudo-timestamp"))

	var logs bytes.Buffer
	e := NewEngine(&logs, &logs)
	e.Password = password

	_, err := e.runSudoWithInput("0\n", "tee", "/proc/definitely-not-here")
	if err == nil {
		t.Fatal("want an error for writing to a non-existent parameter")
	}
	if !strings.Contains(err.Error(), "No such file") {
		t.Errorf("error = %v, want the command's stderr in it", err)
	}
	if strings.Contains(err.Error(), password) || strings.Contains(logs.String(), password) {
		t.Errorf("the password appeared in the error or the log: err=%v logs=%q", err, logs.String())
	}
}

// installFakeSudo puts a sudo stand-in first on PATH. It records its arguments,
// so a test can prove whether sudo was started at all, and it models sudo's stdin
// contract: the password is taken from stdin as exactly one line, and only when
// authentication is actually required (-k present, or no timestamp marker yet).
func installFakeSudo(t *testing.T, record, password string) {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> \"$FAKE_SUDO_RECORD\"\n" +
		"cmd=''; rest=''; has_k=0\n" +
		"for a in \"$@\"; do\n" +
		"  if [ \"$a\" = \"-k\" ]; then has_k=1; continue; fi\n" +
		"  if [ \"$a\" = \"-S\" ] || [ \"$a\" = \"--\" ]; then continue; fi\n" +
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
	if os.Getenv("FAKE_SUDO_MARKER") == "" {
		t.Setenv("FAKE_SUDO_MARKER", filepath.Join(dir, "sudo-timestamp"))
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

// installRecordingSudo puts a sudo stand-in first on PATH that only records that
// it was called, so a test can assert a code path never reaches for root.
func installRecordingSudo(t *testing.T, record string) {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$FAKE_SUDO_RECORD\"\nexit 0\n"
	path := filepath.Join(dir, "sudo")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("writing fake sudo: %v", err)
	}
	t.Setenv("FAKE_SUDO_RECORD", record)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}
