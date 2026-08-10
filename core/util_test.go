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
	"path/filepath"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		slice []string
		str   string
		want  bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{[]string{"a"}, "a", true},
		{[]string{"a"}, "b", false},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			if got := contains(tt.slice, tt.str); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoesFileExist(t *testing.T) {
	tmpDir := t.TempDir()
	goodFile := filepath.Join(tmpDir, "exists.txt")
	_ = os.WriteFile(goodFile, []byte("hello"), 0644)

	if !doesFileExist(goodFile) {
		t.Errorf("doesFileExist(%q) = false, want true", goodFile)
	}
	if doesFileExist(filepath.Join(tmpDir, "nope.txt")) {
		t.Errorf("doesFileExist(nonexistent) = true, want false")
	}
}

func TestIsSubPath(t *testing.T) {
	tests := []struct {
		parent string
		sub    string
		want   bool
	}{
		{"/a/b/c", "/a/b/c/d", true},
		{"/a/b/c", "/a/b/x", false},
		{"/a/b/c", "/x/y/z", false},
		{"/a/b/c", "/a/b/c", true},
	}
	for _, tt := range tests {
		t.Run(tt.parent+"_"+tt.sub, func(t *testing.T) {
			if got := isSubPath(tt.parent, tt.sub); got != tt.want {
				t.Errorf("isSubPath(%q, %q) = %v, want %v", tt.parent, tt.sub, got, tt.want)
			}
		})
	}
}

func TestRemoveElementFromStringSlice(t *testing.T) {
	input := []string{"a", "b", "c", "b"}
	got := removeElementFromStringSlice("b", input)
	want := []string{"a", "c"}
	if len(got) != len(want) {
		t.Errorf("removeElementFromStringSlice() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("removeElementFromStringSlice()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestGetHumanVRAMSize(t *testing.T) {
	tests := []struct {
		input  int
		want   string
	}{
		{512, "512MB"},
		{1024, "1GB"},
		{2048, "2GB"},
		{4096, "4GB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := GetHumanVRAMSize(tt.input); got != tt.want {
				t.Errorf("GetHumanVRAMSize(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
