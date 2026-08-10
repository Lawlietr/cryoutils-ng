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

// ProgressEvent represents a progress update sent via SSE.
type ProgressEvent struct {
	Message string `json:"message"`
}

// ProgressChannel is a broadcast channel for progress updates.
// The desktop server writes to this channel; SSE clients read from it.
var ProgressChannel chan ProgressEvent

func init() {
	ProgressChannel = make(chan ProgressEvent, 32)
}
