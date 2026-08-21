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
	"fmt"
	"os"
	"os/exec"
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
