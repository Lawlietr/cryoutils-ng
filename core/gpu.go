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
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// GetVRAMValue reads the current VRAM from glxinfo.
// Note: actual VRAM change is via BIOS; this is read-only.
func (e *Engine) GetVRAMValue() (int, error) {
	cmd, err := exec.Command("glxinfo", "-B").Output()
	if err != nil {
		return 100, fmt.Errorf("error getting current VRAM")
	}

	re := regexp.MustCompile(`Video memory: [0-9]+`)
	match := re.FindStringSubmatch(string(cmd))
	if match == nil {
		return 100, fmt.Errorf("error getting current VRAM")
	}

	output := strings.Split(match[0], " ")[2]
	e.InfoLog.Println("Found a VRAM of", output)
	vram, _ := strconv.Atoi(output)
	return vram, nil
}
