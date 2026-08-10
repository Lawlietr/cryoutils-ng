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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// GetSwapFileLocation reads the current swap file from /proc/swaps.
func (e *Engine) GetSwapFileLocation() (string, error) {
	file, err := os.Open("/proc/swaps")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // skip header

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[0] != "Filename" {
			location := fields[0]
			if strings.HasPrefix(location, "/dev/") {
				return "", fmt.Errorf("no swapfile found")
			}
			return location, nil
		}
	}

	if doesFileExist(DefaultSwapFileLocation) {
		return DefaultSwapFileLocation, nil
	}
	return "", fmt.Errorf("no swapfile found")
}

// GetSwappinessValue reads the current vm.swappiness value.
func (e *Engine) GetSwappinessValue() (int, error) {
	cmd, err := exec.Command("sysctl", "vm.swappiness").Output()
	if err != nil {
		return 100, fmt.Errorf("error getting current swappiness")
	}
	output := strings.Fields(string(cmd))
	e.InfoLog.Println("Found a swappiness of", output[2])
	swappiness, _ := strconv.Atoi(output[2])
	return swappiness, nil
}

// GetSwapFileSize returns the current swap file size in bytes.
func (e *Engine) GetSwapFileSize() (int64, error) {
	location, err := e.GetSwapFileLocation()
	if err != nil {
		return DefaultSwapSizeBytes, fmt.Errorf("error getting swapfile location: %v", err)
	}
	e.SwapFileLocation = location

	info, err := os.Stat(e.SwapFileLocation)
	if err != nil {
		return DefaultSwapSizeBytes, fmt.Errorf("error getting current swap file size")
	}
	e.InfoLog.Println("Found a swap file with a size of", info.Size())
	return info.Size(), nil
}

// GetAvailableSwapSizes returns a list of viable swap sizes based on free space.
func (e *Engine) GetAvailableSwapSizes() ([]string, error) {
	currentSwapSize, _ := e.GetSwapFileSize()
	availableSpace, err := e.getFreeSpace("/home")
	if err != nil {
		return nil, fmt.Errorf("error getting available space in /home")
	}

	validSizes := []string{"1 - Default"}
	for _, size := range AvailableSwapSizes {
		intSize, _ := strconv.Atoi(size)
		byteSize := intSize * GigabyteMultiplier
		if int64(byteSize+SpaceOverhead) < (availableSpace + currentSwapSize) {
			if byteSize == int(currentSwapSize) {
				validSizes = append(validSizes, fmt.Sprintf("%s - Current Size", size))
			} else {
				validSizes = append(validSizes, size)
			}
		}
	}
	e.InfoLog.Println("Available Swap Sizes:", validSizes)
	return validSizes, nil
}

// DisableSwap temporarily disables all swap.
func (e *Engine) DisableSwap() error {
	e.InfoLog.Println("Disabling swap temporarily...")
	_, err := exec.Command("sudo", "swapoff", "-a").Output()
	if err != nil {
		return fmt.Errorf("error disabling swap")
	}
	return nil
}

// ResizeSwapFile resizes the swap file to the given size in GB.
func (e *Engine) ResizeSwapFile(size int) error {
	locationArg := fmt.Sprintf("of=%s", e.SwapFileLocation)
	countArg := fmt.Sprintf("count=%d", size)
	e.InfoLog.Println("Resizing swap to", size, "GB...")
	_, err := exec.Command("sudo", "dd", "if=/dev/zero", locationArg, "bs=1G", countArg, "status=progress").Output()
	if err != nil {
		return fmt.Errorf("error resizing %s", e.SwapFileLocation)
	}
	return nil
}

// SetSwapPermissions sets swap file permissions to 0600.
func (e *Engine) SetSwapPermissions() error {
	e.InfoLog.Println("Setting permissions on", e.SwapFileLocation, "to 0600...")
	_, err := exec.Command("sudo", "chmod", "600", e.SwapFileLocation).Output()
	if err != nil {
		return fmt.Errorf("error setting permissions on %s", e.SwapFileLocation)
	}
	return nil
}

// InitNewSwapFile runs mkswap and swapon on the current SwapFileLocation.
func (e *Engine) InitNewSwapFile() error {
	e.InfoLog.Println("Enabling swap on", e.SwapFileLocation, "...")
	_, err := exec.Command("sudo", "mkswap", e.SwapFileLocation).Output()
	if err != nil {
		return fmt.Errorf("error creating swap on %s", e.SwapFileLocation)
	}
	_, err = exec.Command("sudo", "swapon", e.SwapFileLocation).Output()
	if err != nil {
		return fmt.Errorf("error enabling swap on %s", e.SwapFileLocation)
	}
	return nil
}

// ChangeSwappiness sets swappiness and writes a tmpfiles.d conf if non-default.
func (e *Engine) ChangeSwappiness(value string) error {
	e.InfoLog.Println("Setting swappiness...")
	_ = e.removeFile(OldSwappinessUnitFile)
	err := e.setUnitValue("swappiness", value)
	if err != nil {
		return err
	}
	if value == DefaultSwappiness {
		e.InfoLog.Println("Removing swappiness unit to revert to default behavior...")
		return e.removeUnitFile("swappiness")
	}
	return e.writeUnitFile("swappiness", value)
}

// ChangeSwapSize changes the swap file to the specified size in GB.
func (e *Engine) ChangeSwapSize(size int) error {
	e.renewAuth()
	if err := e.DisableSwap(); err != nil {
		return err
	}
	if err := e.ResizeSwapFile(size); err != nil {
		return err
	}
	e.renewAuth()
	if err := e.SetSwapPermissions(); err != nil {
		return err
	}
	return e.InitNewSwapFile()
}
