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
	"os"
	"strconv"
	"strings"

	"github.com/andygrunwald/vdf"
)

// Library represents a Steam library location and its installed games.
type Library struct {
	Path           string
	InstalledGames []int
}

// FindDataFolders parses the VDF library folders and returns Library slices.
func (e *Engine) FindDataFolders() ([]Library, error) {
	libraries, err := ParseVDF(LibraryVDFLocation)
	if err != nil {
		e.ErrorLog.Println("Error parsing VDF at", LibraryVDFLocation)
		return nil, err
	}
	e.InfoLog.Println("Loading libraries saved in VDF...")
	for x := range libraries {
		if strings.HasSuffix(libraries[x].Path, "SteamLibrary") {
			e.InfoLog.Println("Found manually added library at", libraries[x].Path)
			libraries[x].Path = strings.ReplaceAll(libraries[x].Path, "/SteamLibrary", "")
		} else {
			e.InfoLog.Println("Found library at", libraries[x].Path)
		}
	}
	return libraries, nil
}

// ParseVDF parses a Steam libraryfolders.vdf file and returns Library slices.
func ParseVDF(file string) ([]Library, error) {
	var libraries []Library

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := vdf.NewParser(f)
	m, err := p.Parse()
	if err != nil {
		return nil, err
	}

	for _, library := range m["libraryfolders"].(map[string]interface{}) {
		var installedGames []int
		for game := range library.(map[string]interface{})["apps"].(map[string]interface{}) {
			intGame, _ := strconv.Atoi(game)
			installedGames = append(installedGames, intGame)
		}
		newLib := Library{
			Path:           library.(map[string]interface{})["path"].(string),
			InstalledGames: installedGames,
		}
		libraries = append(libraries, newLib)
	}
	return libraries, nil
}
