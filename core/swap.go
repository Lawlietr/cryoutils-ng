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

// swapEntry is one row of the /proc/swaps table.
type swapEntry struct {
	Filename string
	Type     string
	SizeKB   int64
	UsedKB   int64
	Priority int
}

// parseSwapEntries parses /proc/swaps content. The header row is skipped and any
// row that does not have the five expected columns is ignored.
func parseSwapEntries(content string) []swapEntry {
	var entries []swapEntry
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Scan() // skip header
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 {
			continue
		}
		size, sizeErr := strconv.ParseInt(fields[2], 10, 64)
		used, usedErr := strconv.ParseInt(fields[3], 10, 64)
		priority, priorityErr := strconv.Atoi(fields[4])
		if sizeErr != nil || usedErr != nil || priorityErr != nil {
			continue
		}
		entries = append(entries, swapEntry{
			Filename: fields[0],
			Type:     fields[1],
			SizeKB:   size,
			UsedKB:   used,
			Priority: priority,
		})
	}
	return entries
}

// readSwapEntries reads the live /proc/swaps table.
func readSwapEntries() ([]swapEntry, error) {
	data, err := os.ReadFile("/proc/swaps")
	if err != nil {
		return nil, err
	}
	return parseSwapEntries(string(data)), nil
}

// parseSwapFileLocation returns the swap *file* from a /proc/swaps table.
// Block devices (/dev/...) are skipped rather than treated as an error: on
// SteamOS /dev/zram0 is a genuine swap entry and must not be mistaken for the
// swap file, nor abort the scan.
func parseSwapFileLocation(entries []swapEntry) string {
	for _, entry := range entries {
		if strings.HasPrefix(entry.Filename, "/dev/") {
			continue
		}
		return entry.Filename
	}
	return ""
}

// zramActiveIn reports whether the zram device is present as a swap entry.
func zramActiveIn(entries []swapEntry) bool {
	for _, entry := range entries {
		if entry.Filename == ZramDevice {
			return true
		}
	}
	return false
}

// parseMemAvailableKB returns MemAvailable in kB from /proc/meminfo content,
// or -1 when the field is not present.
func parseMemAvailableKB(content string) int64 {
	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(line, "MemAvailable:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			if kb, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
				return kb
			}
		}
	}
	return -1
}

// swapMigrationFit reports whether target can be swapped off without running out
// of somewhere to put its pages: everything currently held by target must fit
// into MemAvailable plus the free room of every other active swap entry, while
// keeping marginKB of memory in reserve.
//
// A target that is not in the table at all needs nothing, so it always fits —
// there is simply nothing to migrate.
func swapMigrationFit(entries []swapEntry, target string, memAvailableKB, marginKB int64) (neededKB int64, destinationKB int64, fits bool) {
	for _, entry := range entries {
		if entry.Filename == target {
			neededKB += entry.UsedKB
			continue
		}
		if free := entry.SizeKB - entry.UsedKB; free > 0 {
			destinationKB += free
		}
	}
	if memAvailableKB >= 0 {
		destinationKB += memAvailableKB
	}
	return neededKB, destinationKB, neededKB+marginKB <= destinationKB
}

// GetSwapFileLocation reads the current swap file from /proc/swaps.
func (e *Engine) GetSwapFileLocation() (string, error) {
	entries, err := readSwapEntries()
	if err != nil {
		return "", err
	}
	if location := parseSwapFileLocation(entries); location != "" {
		return location, nil
	}
	if doesFileExist(DefaultSwapFileLocation) {
		return DefaultSwapFileLocation, nil
	}
	return "", fmt.Errorf("no swapfile found")
}

// resolveSwapFileLocation returns the swap file path, discovering it when a
// caller (notably the CLI, which never queries status first) left it unset.
// Without this, an unset SwapFileLocation makes ResizeSwapFile run
// "dd of=" against nothing and swap operations target an empty path.
func (e *Engine) resolveSwapFileLocation() (string, error) {
	if e.SwapFileLocation != "" {
		return e.SwapFileLocation, nil
	}
	location, err := e.GetSwapFileLocation()
	if err != nil {
		return "", err
	}
	e.SwapFileLocation = location
	return location, nil
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

// DisableSwap turns off swap for the swap file only.
//
// It must not use "swapoff -a". On SteamOS that deactivates the generated
// dev-zram0.swap unit, which stops systemd-zram-setup@zram0.service, whose
// ExecStop runs "zram-generator --reset-device zram0" and writes 0 to
// /sys/block/zram0/disksize — destroying the device outright. Nothing can
// re-enable it afterwards with a plain swapon, and it does not come back on its
// own; only a unit restart or a reboot restores it. Full write-up:
// docs/zram-swapoff-reset.md
func (e *Engine) DisableSwap() error {
	location, err := e.resolveSwapFileLocation()
	if err != nil {
		return err
	}
	e.InfoLog.Println("Disabling swap on", location, "temporarily...")
	if _, err := e.runSudo("swapoff", location); err != nil {
		return fmt.Errorf("error disabling swap on %s: %v", location, err)
	}
	return nil
}

// ResizeSwapFile resizes the swap file to the given size in GB.
func (e *Engine) ResizeSwapFile(size int) error {
	location, err := e.resolveSwapFileLocation()
	if err != nil {
		return err
	}
	e.InfoLog.Println("Resizing swap to", size, "GB...")
	if _, err := e.runSudo("dd", "if=/dev/zero", fmt.Sprintf("of=%s", location), "bs=1G", fmt.Sprintf("count=%d", size), "status=progress"); err != nil {
		return fmt.Errorf("error resizing %s: %v", location, err)
	}
	return nil
}

// SetSwapPermissions sets swap file permissions to 0600.
func (e *Engine) SetSwapPermissions() error {
	location, err := e.resolveSwapFileLocation()
	if err != nil {
		return err
	}
	e.InfoLog.Println("Setting permissions on", location, "to 0600...")
	if _, err := e.runSudo("chmod", "600", location); err != nil {
		return fmt.Errorf("error setting permissions on %s: %v", location, err)
	}
	return nil
}

// InitNewSwapFile runs mkswap and swapon on the current SwapFileLocation.
func (e *Engine) InitNewSwapFile() error {
	location, err := e.resolveSwapFileLocation()
	if err != nil {
		return err
	}
	e.InfoLog.Println("Enabling swap on", location, "...")
	if _, err := e.runSudo("mkswap", location); err != nil {
		return fmt.Errorf("error creating swap on %s: %v", location, err)
	}
	// No explicit priority: the kernel gives the swap file the stock SteamOS
	// value (-2) while zram keeps its own 100, which is exactly the pre-resize
	// state — because this code path never touches zram at all.
	if _, err := e.runSudo("swapon", location); err != nil {
		return fmt.Errorf("error enabling swap on %s: %v", location, err)
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

// GetZramStatus returns true if the zram device is currently active as swap.
func (e *Engine) GetZramStatus() (bool, error) {
	entries, err := readSwapEntries()
	if err != nil {
		return false, err
	}
	return zramActiveIn(entries), nil
}

// GetZramSizeBytes returns the zram disk size in bytes, or 0 if zram is not present.
func (e *Engine) GetZramSizeBytes() (int64, error) {
	data, err := os.ReadFile("/sys/block/zram0/disksize")
	if err != nil {
		return 0, nil // zram not present
	}
	size, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, nil
	}
	return size, nil
}

// checkSwapMigrationSafety refuses to turn the swap file off when the pages it
// currently holds would not fit elsewhere. SteamOS runs earlyoom, which kills
// processes under memory pressure, so an unguarded swapoff is a reliable way to
// lose running applications in the middle of a resize.
func (e *Engine) checkSwapMigrationSafety(location string) error {
	entries, err := readSwapEntries()
	if err != nil {
		e.InfoLog.Println("Could not read /proc/swaps, skipping the migration safety check:", err)
		return nil
	}

	memAvailableKB := int64(-1)
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		memAvailableKB = parseMemAvailableKB(string(data))
	} else {
		e.InfoLog.Println("Could not read /proc/meminfo, skipping the migration safety check:", err)
	}
	if memAvailableKB < 0 {
		e.InfoLog.Println("MemAvailable not reported by this kernel, skipping the migration safety check.")
		return nil
	}

	neededKB, destinationKB, fits := swapMigrationFit(entries, location, memAvailableKB, int64(SwapMigrationMarginKB))
	if neededKB == 0 {
		return nil // nothing is held by this swap file, nothing to migrate
	}
	if !fits {
		return fmt.Errorf(
			"refusing to turn off swap on %s: %d MB in use, but only %d MB can receive it "+
				"(free RAM plus other swap devices, keeping %d MB in reserve). "+
				"Close some applications, or set a swap size at least as large as what is in use, then retry",
			location, neededKB/1024, destinationKB/1024, SwapMigrationMarginKB/1024)
	}
	e.InfoLog.Printf("Swap migration check OK: %d MB in use, %d MB able to receive it", neededKB/1024, destinationKB/1024)
	return nil
}

// restoreSwapFile re-enables swap on a best-effort basis after a step failed, so
// an interrupted resize never leaves the machine with no swap at all.
func (e *Engine) restoreSwapFile(location string) {
	e.InfoLog.Println("Resize failed - attempting to re-enable swap on", location, "...")
	if _, err := e.runSudo("mkswap", location); err != nil {
		e.ErrorLog.Println("Best-effort mkswap failed:", err)
	}
	if _, err := e.runSudo("swapon", location); err != nil {
		e.ErrorLog.Println("Best-effort swapon failed:", err)
	}
}

// ChangeSwapSize changes the swap file to the specified size in GB.
//
// zram is deliberately never touched: only the swap file is turned off, so
// zram keeps its capacity and its stock priority of 100 throughout. The earlier
// design (swapoff -a followed by swapon /dev/zram0) both destroyed the device
// and then tried to re-enable something with zero capacity, and it also handed
// zram the kernel default priority instead of 100, which made even the
// "successful" path useless. See docs/zram-swapoff-reset.md.
func (e *Engine) ChangeSwapSize(size int) error {
	if size < 1 {
		return fmt.Errorf("invalid swap size %d GB", size)
	}
	if err := e.RenewAuth(); err != nil {
		return err
	}

	location, err := e.resolveSwapFileLocation()
	if err != nil {
		return err
	}

	// Recorded only to warn if something else removed zram underneath us.
	zramBefore, _ := e.GetZramStatus()

	if err := e.checkSwapMigrationSafety(location); err != nil {
		return err
	}

	e.emitProgress("Disabling swap on " + location + "...")
	if err := e.DisableSwap(); err != nil {
		return err
	}

	e.emitProgress(fmt.Sprintf("Resizing swap file to %d GB...", size))
	if err := e.ResizeSwapFile(size); err != nil {
		e.restoreSwapFile(location)
		return err
	}

	if err := e.SetSwapPermissions(); err != nil {
		e.restoreSwapFile(location)
		return err
	}

	e.emitProgress("Enabling swap...")
	if err := e.InitNewSwapFile(); err != nil {
		return err
	}

	if zramBefore {
		if zramAfter, _ := e.GetZramStatus(); !zramAfter {
			e.ErrorLog.Printf("zram swap (%s) is gone after the resize; restore it with: systemctl restart %s",
				ZramDevice, ZramSetupUnit)
		}
	}

	e.emitProgress("Swap resize complete.")
	return nil
}
