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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

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

// runSudo runs a privileged command and returns its combined stdout+stderr.
//
// When the process is already root (e.g. the CLI launched via sudo) the command
// runs directly, without a sudo wrapper. Otherwise sudo is called with -S and
// the stored password is written to stdin; sudo with a cached timestamp never
// reads stdin, so a write failure on the pipe is not treated as an error.
//
// Unlike the previous per-call exec.Command("sudo", ...).Output() sites, the
// command line, the wait error and the command's own output are all reported.
// Swap failures used to be reduced to fixed strings such as "error disabling
// swap", which is why the zram teardown in docs/zram-swapoff-reset.md stayed
// invisible for so long.
func (e *Engine) runSudo(args ...string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("runSudo: no command given")
	}

	if os.Geteuid() == 0 {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			e.ErrorLog.Println(commandError(args, err, string(out)))
		}
		return string(out), commandError(args, err, string(out))
	}

	cmd := exec.Command("sudo", append([]string{"-S"}, args...)...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	stdin, err := cmd.StdinPipe()
	if err != nil {
		e.ErrorLog.Println(err)
		return "", err
	}
	if err := cmd.Start(); err != nil {
		e.ErrorLog.Println(err)
		return "", commandError(args, err, "")
	}
	// A write error here just means sudo never asked (cached timestamp).
	if _, err := stdin.Write([]byte(e.Password + "\n")); err != nil {
		_ = err
	}
	_ = stdin.Close()

	err = cmd.Wait()
	output := out.String()
	if err != nil {
		e.ErrorLog.Println(commandError(args, err, output))
		return output, commandError(args, err, output)
	}
	return output, nil
}

// commandError builds an error that carries the command line and (the tail of)
// its output, so a failure can actually be diagnosed. It returns nil when the
// command succeeded.
func commandError(args []string, err error, output string) error {
	if err == nil {
		return nil
	}
	detail := strings.TrimSpace(output)
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
