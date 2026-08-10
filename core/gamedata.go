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
	"path/filepath"
	"strconv"

	cp "github.com/otiai10/copy"
)

// GameStatus holds metadata about a single game found on disk.
type GameStatus struct {
	GameName    string
	IsInstalled bool
}

// StorageStatus holds directory listings for two storage locations.
type StorageStatus struct {
	LeftCompatDirectories  []string
	LeftShaderDirectories  []string
	RightCompatDirectories []string
	RightShaderDirectories []string
}

// DataToMove holds the queue of game directories to move between locations.
type DataToMove struct {
	right     []string
	left      []string
	rightSize int64
	leftSize  int64
}

// GetRight returns the list of directories to remove from the right location.
func (d DataToMove) GetRight() []string { return d.right }

// GetLeft returns the list of directories to remove from the left location.
func (d DataToMove) GetLeft() []string { return d.left }

// GetDirectoryList returns the names of directories inside path.
func GetDirectoryList(path string, includeSymlinks bool) ([]string, error) {
	var folderList []string
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		fullPath := filepath.Join(path, file.Name())
		if !includeSymlinks {
			if file.IsDir() && !isSymbolicLink(fullPath) {
				folderList = append(folderList, file.Name())
			}
		} else {
			if file.IsDir() {
				folderList = append(folderList, file.Name())
			}
		}
	}
	return folderList, nil
}

func (s *StorageStatus) getStorageStatus(e *Engine, left string, right string) error {
	var err error
	if left == SteamDataRoot {
		s.LeftCompatDirectories, err = GetDirectoryList(SteamCompatRoot, false)
		if err != nil {
			return err
		}
		s.LeftShaderDirectories, err = GetDirectoryList(SteamShaderRoot, false)
		if err != nil {
			return err
		}
	} else {
		compat := filepath.Join(left, ExternalCompatRoot)
		shader := filepath.Join(left, ExternalShaderRoot)
		_ = os.MkdirAll(compat, 0777)
		_ = os.MkdirAll(shader, 0777)
		s.LeftCompatDirectories, err = GetDirectoryList(compat, false)
		if err != nil {
			return err
		}
		s.LeftShaderDirectories, err = GetDirectoryList(shader, false)
		if err != nil {
			return err
		}
	}
	if right == SteamDataRoot {
		s.RightCompatDirectories, err = GetDirectoryList(SteamCompatRoot, false)
		if err != nil {
			return err
		}
		s.RightShaderDirectories, err = GetDirectoryList(SteamShaderRoot, false)
		if err != nil {
			return err
		}
	} else {
		compat := filepath.Join(right, ExternalCompatRoot)
		shader := filepath.Join(right, ExternalShaderRoot)
		_ = os.MkdirAll(compat, 0777)
		_ = os.MkdirAll(shader, 0777)
		s.RightCompatDirectories, err = GetDirectoryList(compat, false)
		if err != nil {
			return err
		}
		s.RightShaderDirectories, err = GetDirectoryList(shader, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DataToMove) getSpaceNeeded(left string, right string) {
	var leftCompat, rightCompat, leftShader, rightShader string
	if left == SteamDataRoot {
		leftCompat = SteamCompatRoot
		leftShader = SteamShaderRoot
	} else {
		leftCompat = filepath.Join(left, ExternalCompatRoot)
		leftShader = filepath.Join(left, ExternalShaderRoot)
	}
	if right == SteamDataRoot {
		rightCompat = SteamCompatRoot
		rightShader = SteamShaderRoot
	} else {
		rightCompat = filepath.Join(right, ExternalCompatRoot)
		rightShader = filepath.Join(right, ExternalShaderRoot)
	}
	for x := range d.left {
		d.leftSize += getDirectorySize(filepath.Join(leftCompat, d.left[x]))
		d.leftSize += getDirectorySize(filepath.Join(leftShader, d.left[x]))
	}
	for x := range d.right {
		d.rightSize += getDirectorySize(filepath.Join(rightCompat, d.right[x]))
		d.rightSize += getDirectorySize(filepath.Join(rightShader, d.right[x]))
	}
}

func (d *DataToMove) getDataToMove(e *Engine, left string, right string) error {
	libraries, err := e.FindDataFolders()
	if err != nil {
		return err
	}
	storage := new(StorageStatus)
	if err := storage.getStorageStatus(e, left, right); err != nil {
		return err
	}
	for i := range libraries {
		if isSubPath(left, libraries[i].Path) {
			e.InfoLog.Println("Library location selected as left:", libraries[i].Path)
			for _, game := range libraries[i].InstalledGames {
				gameString := strconv.Itoa(game)
				e.InfoLog.Println("Library contains:", gameString)
				if contains(storage.RightCompatDirectories, gameString) && contains(storage.RightShaderDirectories, gameString) {
					d.right = append(d.right, gameString)
				}
			}
		} else if isSubPath(right, libraries[i].Path) {
			e.InfoLog.Println("Library location selected as right:", libraries[i].Path)
			for _, game := range libraries[i].InstalledGames {
				gameString := strconv.Itoa(game)
				e.InfoLog.Println("Library contains:", gameString)
				if contains(storage.LeftCompatDirectories, gameString) && contains(storage.LeftShaderDirectories, gameString) {
					d.left = append(d.left, gameString)
				}
			}
		}
	}
	return nil
}

// GetStorageStatus returns a populated StorageStatus for left and right paths.
func (e *Engine) GetStorageStatus(left string, right string) (StorageStatus, error) {
	storage := new(StorageStatus)
	if err := storage.getStorageStatus(e, left, right); err != nil {
		return StorageStatus{}, err
	}
	return *storage, nil
}

// GetDataToMove returns a DataToMove with the queue of directories to move.
func (e *Engine) GetDataToMove(left string, right string) (DataToMove, error) {
	d := new(DataToMove)
	if err := d.getDataToMove(e, left, right); err != nil {
		return DataToMove{}, err
	}
	d.getSpaceNeeded(left, right)
	return *d, nil
}

// MoveGameData moves game data from left to right (or vice versa).
func (e *Engine) MoveGameData(left string, right string, data DataToMove) error {
	e.InfoLog.Println("Moving the following data:")
	for x := range data.right {
		e.InfoLog.Println(data.right[x])
	}
	for x := range data.left {
		e.InfoLog.Println(data.left[x])
	}
	for x := range data.right {
		leftPath := filepath.Join(left, ExternalCompatRoot, data.right[x])
		rightPath := filepath.Join(right, ExternalCompatRoot, data.right[x])
		_ = os.MkdirAll(rightPath, 0777)
		if err := cp.Copy(leftPath, rightPath); err != nil {
			return err
		}
		leftPath = filepath.Join(left, ExternalShaderRoot, data.right[x])
		rightPath = filepath.Join(right, ExternalShaderRoot, data.right[x])
		if err := cp.Copy(leftPath, rightPath); err != nil {
			return err
		}
	}
	for x := range data.left {
		leftPath := filepath.Join(left, ExternalCompatRoot, data.left[x])
		rightPath := filepath.Join(right, ExternalCompatRoot, data.left[x])
		_ = os.MkdirAll(rightPath, 0777)
		if err := cp.Copy(leftPath, rightPath); err != nil {
			return err
		}
		leftPath = filepath.Join(left, ExternalShaderRoot, data.left[x])
		rightPath = filepath.Join(right, ExternalShaderRoot, data.left[x])
		if err := cp.Copy(leftPath, rightPath); err != nil {
			return err
		}
	}
	return nil
}

// ConfirmDirectoryStatus returns a descriptive string for the directory status.
func (e *Engine) ConfirmDirectoryStatus(left string, right string) string {
	var leftSpace, rightSpace int64
	leftFree, _ := e.getFreeSpace(left)
	rightFree, _ := e.getFreeSpace(right)
	leftSpace, _ = e.getFreeSpace(filepath.Join(left, ExternalCompatRoot))
	rightSpace, _ = e.getFreeSpace(filepath.Join(right, ExternalCompatRoot))
	leftShaderSpace, _ := e.getFreeSpace(filepath.Join(left, ExternalShaderRoot))
	rightShaderSpace, _ := e.getFreeSpace(filepath.Join(right, ExternalShaderRoot))
	return fmt.Sprintf(`
Left  Total: %s | Compat: %s | Shader: %s
Right Total: %s | Compat: %s | Shader: %s`,
		GetHumanVRAMSize(int(leftFree / 1024 / 1024)),
		GetHumanVRAMSize(int(leftSpace / 1024 / 1024)),
		GetHumanVRAMSize(int(leftShaderSpace / 1024 / 1024)),
		GetHumanVRAMSize(int(rightFree / 1024 / 1024)),
		GetHumanVRAMSize(int(rightSpace / 1024 / 1024)),
		GetHumanVRAMSize(int(rightShaderSpace / 1024 / 1024)),
	)
}

// GetUninstalledGamesData returns a DataToMove for uninstalled games.
func (e *Engine) GetUninstalledGamesData(left string, right string) (DataToMove, error) {
	libraries, err := e.FindDataFolders()
	if err != nil {
		return DataToMove{}, err
	}
	storage := new(StorageStatus)
	if err := storage.getStorageStatus(e, left, right); err != nil {
		return DataToMove{}, err
	}
	d := new(DataToMove)
	for i := range libraries {
		if isSubPath(left, libraries[i].Path) {
			e.InfoLog.Println("Library location selected as left:", libraries[i].Path)
			for _, game := range storage.LeftCompatDirectories {
				if contains(storage.RightCompatDirectories, game) && contains(storage.RightShaderDirectories, game) {
					d.right = append(d.right, game)
				}
			}
			for _, game := range storage.LeftShaderDirectories {
				if contains(storage.RightCompatDirectories, game) && contains(storage.RightShaderDirectories, game) {
					d.right = append(d.right, game)
				}
			}
		} else if isSubPath(right, libraries[i].Path) {
			e.InfoLog.Println("Library location selected as right:", libraries[i].Path)
			for _, game := range storage.RightCompatDirectories {
				if contains(storage.LeftCompatDirectories, game) && contains(storage.LeftShaderDirectories, game) {
					d.left = append(d.left, game)
				}
			}
			for _, game := range storage.RightShaderDirectories {
				if contains(storage.LeftCompatDirectories, game) && contains(storage.LeftShaderDirectories, game) {
					d.left = append(d.left, game)
				}
			}
		}
	}
	d.getSpaceNeeded(left, right)
	return *d, nil
}
