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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/moby/sys/mountinfo"
	"golang.org/x/sys/unix"
)

var stat unix.Statfs_t

func (e *Engine) getFreeSpace(path string) (int64, error) {
	err := unix.Statfs(path, &stat)
	if err != nil {
		e.ErrorLog.Println(err)
		return 0, fmt.Errorf("error getting free space")
	}
	return int64(stat.Bfree * uint64(stat.Bsize)), nil
}

func getDirectorySize(path string) int64 {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, _ error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func isSymbolicLink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		panic(err)
	}
	return fi.Mode()&os.ModeSymlink != 0
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func waitForDeletion(path string, directory string) {
	for {
		if !doesDirectoryExist(path, directory) {
			break
		}
		time.Sleep(time.Second)
	}
}

// doesDirectoryExist checks whether directory exists inside path.
func doesDirectoryExist(path string, directory string) bool {
	directories, _ := os.ReadDir(path)
	for _, dir := range directories {
		if dir.Name() == directory {
			return true
		}
	}
	return false
}

func doesFileExist(path string) bool {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func isSubPath(parent string, sub string) bool {
	parentElems := strings.Split(strings.TrimRight(parent, "/"), "/")
	subElems := strings.Split(strings.TrimRight(sub, "/"), "/")
	if len(subElems) < len(parentElems) {
		return false
	}
	for i := range parentElems {
		if subElems[i] != parentElems[i] {
			return false
		}
	}
	return true
}

// writeFile writes contents to path via sudo mv from a temp file.
func (e *Engine) writeFile(path string, contents string) error {
	e.InfoLog.Println("Writing", path)

	// The temp file lives in our own install directory, so removing it never needs
	// root. Going through removeFile used to run `sudo rm` for it on every write.
	tempPath := filepath.Join(InstallDirectory, "temp.txt")
	_ = os.Remove(tempPath)
	f, err := os.Create(tempPath)
	if err != nil {
		e.ErrorLog.Println(err)
		return err
	}
	defer f.Close()

	_, err = f.WriteString(contents)
	if err != nil {
		e.ErrorLog.Println(err)
		return err
	}

	_, err = e.runSudo("mv", tempPath, path)
	if err != nil {
		return fmt.Errorf("error moving temp file to final location: %w", err)
	}
	return nil
}

// removeFile deletes a path that may need root to remove.
//
// rm is called with -f: the revert path calls this for a tmpfiles.d conf that is
// simply not there (nothing to revert), and rm without -f turns that into a
// non-zero exit, which used to make a perfectly successful revert look broken.
func (e *Engine) removeFile(path string) error {
	e.InfoLog.Println("Removing", path)
	_, err := e.runSudo("rm", "-f", path)
	if err != nil {
		e.InfoLog.Println("Couldn't delete", path)
	}
	return err
}

// getListOfAttachedDrives returns a list of all attached drives including SteamDataRoot.
func (e *Engine) getListOfAttachedDrives() ([]string, error) {
	drives := []string{SteamDataRoot}
	filter := mountinfo.PrefixFilter(MountDirectory)
	info, err := mountinfo.GetMounts(filter)
	if err != nil {
		e.ErrorLog.Println(err)
		return nil, err
	}
	for x := range info {
		drives = append(drives, info[x].Mountpoint)
	}
	e.InfoLog.Printf("Attached drives: %s", drives)
	return drives, nil
}

// getListOfDataAllDataLocations returns a list of all data locations (compat and shader data).
func (e *Engine) getListOfDataAllDataLocations() ([]string, error) {
	drives, err := e.getListOfAttachedDrives()
	if err != nil {
		return nil, err
	}

	var possibleLocations []string
	for x := range drives {
		if drives[x] == SteamDataRoot {
			possibleLocations = append(possibleLocations, SteamCompatRoot)
			possibleLocations = append(possibleLocations, SteamShaderRoot)
		} else {
			compat := filepath.Join(drives[x], ExternalCompatRoot)
			shader := filepath.Join(drives[x], ExternalShaderRoot)
			possibleLocations = append(possibleLocations, compat)
			possibleLocations = append(possibleLocations, shader)
		}
	}
	return possibleLocations, nil
}

func removeElementFromStringSlice(str string, slice []string) []string {
	var newSlice []string
	for x := range slice {
		if str != slice[x] {
			newSlice = append(newSlice, slice[x])
		}
	}
	return newSlice
}

// getUnitStatus reads the current value of a kernel parameter.
//
// It reads the file directly first. Every entry in the unit matrix is
// world-readable on Linux (root-owned, mode 0644), so reading a status never
// needs root: sending the read through sudo is what made the memory parameters
// report as unknown in the web UI, where no password can be supplied at all, and
// it wrote one failing sudo line per parameter into the log on every 5s refresh.
//
// sudo is kept as a fallback for the parameters that are not readable by an
// ordinary user, and it goes through runSudo rather than a bare `sudo cat`: a
// bare sudo inherits the terminal and prompts on the tty, which is how the native
// UI ended up asking twice.
func (e *Engine) getUnitStatus(param string) (string, error) {
	path := UnitMatrix[param]
	if path == "" {
		return "", fmt.Errorf("unknown parameter %q", param)
	}
	if data, err := unitFileReader(path); err == nil {
		return parseUnitValue(string(data)), nil
	} else if !errors.Is(err, os.ErrPermission) {
		// The parameter does not exist on this kernel; sudo cannot fix that, so
		// report it instead of paying for a privileged attempt.
		return "", err
	}
	out, err := e.runSudo("cat", path)
	if err != nil {
		return "", err
	}
	return parseUnitValue(out), nil
}

// unitFileReader is the reader used for the unprivileged status reads. It is a
// variable so a test can simulate a permission failure and prove the sudo
// fallback still works.
var unitFileReader = os.ReadFile

// parseUnitValue reduces a kernel parameter file to its current value. List
// style files (transparent hugepages, defrag) print every accepted value and
// bracket the active one: "always madvise [never]" -> "never".
func parseUnitValue(raw string) string {
	if strings.Contains(raw, "[") {
		for _, field := range strings.Fields(raw) {
			if strings.HasPrefix(field, "[") {
				return strings.Trim(field, "[]")
			}
		}
	}
	return strings.TrimSpace(raw)
}

// writeUnitFile writes a tmpfiles.d conf to persist a kernel parameter.
func (e *Engine) writeUnitFile(param string, value string) error {
	path := filepath.Join(TmpFilesRoot, param+".conf")
	e.InfoLog.Println("Writing", value, "to", path, "to preserve", param, "setting...")
	contents := strings.ReplaceAll(TemplateUnitFile, "PARAM", UnitMatrix[param])
	contents = strings.ReplaceAll(contents, "VALUE", value)
	err := e.writeFile(path, contents)
	if err != nil {
		e.ErrorLog.Println(err)
		return err
	}
	return nil
}

// removeUnitFile removes a tmpfiles.d conf to revert a kernel parameter.
func (e *Engine) removeUnitFile(param string) error {
	path := filepath.Join(TmpFilesRoot, param+".conf")
	e.InfoLog.Println("Removing", path, "to revert", param, "setting...")
	err := e.removeFile(path)
	if err != nil {
		return err
	}
	return nil
}

// setUnitValue writes a value directly to a kernel parameter.
//
// The previous implementation piped `echo value` into a bare `sudo tee`, ignored
// every error, and dumped the command output to os.Stdout, so a failed write
// looked like a successful one. It now goes through runSudoWithInput, which
// supplies the password on stdin without ever prompting on the terminal, and the
// error is returned.
func (e *Engine) setUnitValue(param string, value string) error {
	e.InfoLog.Println("Writing", value, "for param", param, "to memory.")
	_, err := e.runSudoWithInput(value, "tee", UnitMatrix[param])
	return err
}

// getHumanVRAMSize converts VRAM size in MB to a human-readable string.
func GetHumanVRAMSize(size int) string {
	text := fmt.Sprintf("%dMB", size)
	if size >= 1024 {
		text = fmt.Sprintf("%dGB", size/1024)
	}
	return text
}

// removeGameData removes the specified directories from the given locations.
func (e *Engine) RemoveGameData(removeList []string, locations []string) {
	e.InfoLog.Println("Removing the following content:")
	for i := range removeList {
		for j := range locations {
			path := filepath.Join(locations[j], removeList[i])
			e.InfoLog.Println(path)
			err := os.RemoveAll(path)
			if err != nil {
				e.ErrorLog.Println(err)
			}
		}
	}
}
