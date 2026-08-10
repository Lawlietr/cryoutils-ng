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
	"io"
	"os"
	"os/exec"
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

	tempPath := filepath.Join(InstallDirectory, "temp.txt")
	_ = e.removeFile(tempPath)
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

	_, err = exec.Command("sudo", "mv", tempPath, path).Output()
	if err != nil {
		return fmt.Errorf("error moving temp file to final location")
	}
	return nil
}

func (e *Engine) removeFile(path string) error {
	e.InfoLog.Println("Removing", path)
	_, err := exec.Command("sudo", "rm", path).Output()
	if err != nil {
		e.InfoLog.Println("Couldn't delete", path, ", likely missing.")
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
func (e *Engine) getUnitStatus(param string) (string, error) {
	var output string
	cmd, err := exec.Command("sudo", "cat", UnitMatrix[param]).Output()
	if err != nil {
		e.ErrorLog.Println(err)
		return "nil", err
	}
	// This is just to get the actual value in units which present as a list.
	if strings.Contains(string(cmd), "[") {
		slice := strings.Fields(string(cmd))
		for x := range slice {
			if strings.Contains(slice[x], "[") {
				output = strings.ReplaceAll(slice[x], "[", "")
				output = strings.ReplaceAll(output, "]", "")
			}
		}
	} else {
		output = strings.TrimSpace(string(cmd))
	}
	return output, nil
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

// setUnitValue writes a value directly to a kernel parameter via sudo tee.
func (e *Engine) setUnitValue(param string, value string) error {
	e.InfoLog.Println("Writing", value, "for param", param, "to memory.")
	// This mess is the only way I could find to push directly to unit files,
	// without requiring a sudo password on installation to change capabilities.
	echoCmd := exec.Command("echo", value)
	teeCmd := exec.Command("sudo", "tee", UnitMatrix[param])
	reader, writer := io.Pipe()
	var buf bytes.Buffer
	echoCmd.Stdout = writer
	teeCmd.Stdin = reader
	teeCmd.Stdout = &buf
	echoCmd.Start()
	teeCmd.Start()
	echoCmd.Wait()
	writer.Close()
	teeCmd.Wait()
	reader.Close()
	io.Copy(os.Stdout, &buf)
	return nil
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
