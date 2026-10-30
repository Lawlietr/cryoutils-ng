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
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ErrNoSudoPassword is returned instead of starting sudo when no password is
// available. Starting sudo in that situation is exactly what made the native UI
// ask twice: a sudo without -S inherits the terminal and prompts on the tty, and
// the UI dialog then asks again through RenewAuth (which uses sudo -k on
// purpose). The short circuit also keeps the 5s status refresh from writing one
// failed sudo line per parameter into the log.
var ErrNoSudoPassword = errors.New("sudo password not provided")

// useSudoAlways makes runSudo wrap its command in sudo even when the test
// process is already root. It exists so the password/payload semantics (the -k
// flag, stdin line ownership, stdout vs stderr) can be locked by unit tests on
// any machine, instead of only on a host where the suite happens to run
// unprivileged. Never set it outside a test.
var useSudoAlways = false

// RenewAuth sends the stored sudo password to a dummy sudo command to refresh
// the sudo timestamp cache. Returns error if password is empty or authentication fails.
// When already running as root (e.g. CLI launched via sudo), authentication is skipped.
func (e *Engine) RenewAuth() error {
	if os.Geteuid() == 0 {
		return nil
	}
	if e.Password == "" {
		return fmt.Errorf("no stored password")
	}
	// -k forces sudo to forget cached timestamps, so stdin password is always required
	cmd := exec.Command("sudo", "-S", "-k", "--", "echo")
	cmd.WaitDelay = 500 * time.Millisecond
	stdin, err := cmd.StdinPipe()
	if err != nil {
		e.ErrorLog.Println(err)
		return err
	}
	if err := cmd.Start(); err != nil {
		e.ErrorLog.Println(err)
		return err
	}
	if _, err := stdin.Write([]byte(e.Password + "\n")); err != nil {
		cmd.Process.Kill()
		e.ErrorLog.Println(err)
		return err
	}
	stdin.Close()
	if err := cmd.Wait(); err != nil {
		e.ErrorLog.Println("sudo auth failed")
		return fmt.Errorf("sudo authentication failed")
	}
	return nil
}

// runSudo runs a privileged command and returns its standard output.
//
// When the process is already root (e.g. the CLI launched via sudo) the command
// runs directly, without a sudo wrapper. Otherwise sudo is called with -S -k and
// the stored password is written to stdin (see runSudoWithInput for why -k is
// mandatory).
//
// Unlike the previous per-call exec.Command("sudo", ...).Output() sites, the
// command line, the wait error and the command's own stderr are all reported.
// Swap failures used to be reduced to fixed strings such as "error disabling
// swap", which is why the zram teardown in docs/zram-swapoff-reset.md stayed
// invisible for so long.
func (e *Engine) runSudo(args ...string) (string, error) {
	return e.runSudoWithInput("", args...)
}

// runSudoWithInput is runSudo with data delivered to the command's stdin. It
// exists for `sudo tee`, where the value to write is fed on stdin.
//
// sudo is always called with -k, and that is not cosmetic. sudo reads the
// password from stdin only when it actually has to authenticate; when its
// timestamp is still cached it does not read stdin at all and the child
// inherits the whole payload, password line included. Since RenewAuth/TestAuth
// run right before every operation, the timestamp is cached in practice, so the
// previous code (`sudo -S`, no -k) wrote the plaintext sudo password as the
// first line of every parameter file: `tee /sys/kernel/mm/transparent_hugepage/
// enabled` rejected that line with EINVAL, tee exited non-zero, and every
// memory parameter toggle was reported as a failure. `tee` also echoes its
// input back on stdout, so the same password line ended up in the log file.
// With -k sudo authenticates on every call and therefore always consumes
// exactly one line, leaving only the payload for the command.
//
// stdout and stderr are kept in separate buffers: sudo writes its password
// prompt to stderr, and mixing the two used to turn a read into something like
// "[sudo] password for deck: 60", which never matched the expected value.
func (e *Engine) runSudoWithInput(input string, args ...string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("runSudo: no command given")
	}

	if os.Geteuid() == 0 && !useSudoAlways {
		cmd := exec.Command(args[0], args[1:]...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if input != "" {
			cmd.Stdin = strings.NewReader(input)
		}
		err := cmd.Run()
		if err != nil {
			failure := commandError(args, err, stderr.String())
			e.ErrorLog.Println(e.redact(failure.Error()))
			return stdout.String(), failure
		}
		return stdout.String(), nil
	}

	if e.Password == "" {
		return "", ErrNoSudoPassword
	}

	cmd := exec.Command("sudo", append([]string{"-S", "-k"}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		e.ErrorLog.Println(err)
		return "", err
	}
	if err := cmd.Start(); err != nil {
		e.ErrorLog.Println(err)
		return "", commandError(args, err, "")
	}
	payload := e.Password + "\n"
	if input != "" {
		payload += input
	}
	// With -k sudo always reads the password line, so the write itself is not
	// expected to fail; a failure here is still not fatal to the command.
	if _, err := stdin.Write([]byte(payload)); err != nil {
		_ = err
	}
	_ = stdin.Close()

	err = cmd.Wait()
	if err != nil {
		failure := commandError(args, err, stderr.String())
		e.ErrorLog.Println(e.redact(failure.Error()))
		return stdout.String(), failure
	}
	return stdout.String(), nil
}

// redact removes the in-memory sudo password from any text before it is logged
// or returned to the UI. `tee` echoes its stdin, and that stdin used to start
// with the password, so error strings (and therefore the log file) could
// contain it in plain text.
func (e *Engine) redact(text string) string {
	if e.Password == "" {
		return text
	}
	return strings.ReplaceAll(text, e.Password, "[redacted]")
}

// commandError builds an error that carries the command line and (the tail of)
// the command's stderr, so a failure can actually be diagnosed. It returns nil
// when the command succeeded. stdout is deliberately not included: on reads it
// is the value itself, and on `tee` it is an echo of what was written.
func commandError(args []string, err error, stderr string) error {
	if err == nil {
		return nil
	}
	detail := strings.TrimSpace(stderr)
	if len(detail) > 400 {
		detail = detail[len(detail)-400:]
	}
	if detail == "" {
		return fmt.Errorf("%s: %v", strings.Join(args, " "), err)
	}
	return fmt.Errorf("%s: %v: %s", strings.Join(args, " "), err, detail)
}

// TestAuth verifies that the provided password is correct.
// -k forces sudo to forget cached timestamps, so stdin password is always required.
func (e *Engine) TestAuth(password string) error {
	cmd := exec.Command("sudo", "-S", "-k", "--", "echo")
	cmd.WaitDelay = 500 * time.Millisecond
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := stdin.Write([]byte(password + "\n")); err != nil {
		cmd.Process.Kill()
		return err
	}
	stdin.Close()
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("authentication failed")
	}
	return nil
}
