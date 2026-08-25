// CryoUtils NG
// Copyright (C) 2025 CryoUtils NG contributors
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

package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAvailable(t *testing.T) {
	codes := Available()
	if len(codes) == 0 {
		t.Fatal("Expected at least one locale")
	}
	if codes[0] != "en" {
		t.Errorf("Expected 'en' first, got %q", codes[0])
	}
	found := false
	for _, c := range codes {
		if c == "zh-TW" {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'zh-TW' in available locales")
	}
}

func TestLoadEn(t *testing.T) {
	l := Load("en")
	if l == nil {
		t.Fatal("Load(en) returned nil")
	}
	if l.Label() != "English" {
		t.Errorf("Expected label 'English', got %q", l.Label())
	}
	if l.T("status.title") != "System Status" {
		t.Errorf("Expected 'System Status', got %q", l.T("status.title"))
	}
	if l.T("swap.resizeSwap") != "Resize Swap" {
		t.Errorf("Expected 'Resize Swap', got %q", l.T("swap.resizeSwap"))
	}
}

func TestLoadZhTW(t *testing.T) {
	l := Load("zh-TW")
	if l == nil {
		t.Fatal("Load(zh-TW) returned nil")
	}
	if l.Label() != "繁體中文" {
		t.Errorf("Expected label '繁體中文', got %q", l.Label())
	}
	if l.T("status.title") != "系統狀態" {
		t.Errorf("Expected '系統狀態', got %q", l.T("status.title"))
	}
	if l.T("swap.resizeSwap") != "調整 Swap" {
		t.Errorf("Expected '調整 Swap', got %q", l.T("swap.resizeSwap"))
	}
}

func TestTFallbackChain(t *testing.T) {
	// Non-existent key should fall back to en, then to key itself
	l := Load("zh-TW")
	// "common.loading" exists in both → should return zh-TW value
	if l.T("common.loading") != "載入中…" {
		t.Errorf("Expected zh-TW value")
	}
	// Non-existent key → should return key itself
	if l.T("nonexistent.key") != "nonexistent.key" {
		t.Errorf("Expected key fallback, got %q", l.T("nonexistent.key"))
	}
}

func TestRuntimeOverride(t *testing.T) {
	// Create a temp runtime dir
	tmpDir := t.TempDir()
	// We can't change core.InstallDirectory at runtime, so test via a direct path
	// Instead, test the loadFromDir function indirectly by creating a file
	// in a temp dir and loading it manually
	// Actually, let's test the override mechanism by creating a temp dir
	// and using loadFromDir directly
	overrideData := []byte(`{"_label": "Override", "native": {"test": "override value"}}`)
	overridePath := filepath.Join(tmpDir, "en.json")
	if err := os.WriteFile(overridePath, overrideData, 0644); err != nil {
		t.Fatal(err)
	}
	l := loadFromDir(tmpDir, "en")
	if l == nil {
		t.Fatal("loadFromDir returned nil")
	}
	if l.T("native.test") != "override value" {
		t.Errorf("Expected override value, got %q", l.T("native.test"))
	}
	if l.Label() != "Override" {
		t.Errorf("Expected label 'Override', got %q", l.Label())
	}
}

func TestBadJSON(t *testing.T) {
	tmpDir := t.TempDir()
	badPath := filepath.Join(tmpDir, "bad.json")
	if err := os.WriteFile(badPath, []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	// Should not crash, should return nil
	l := loadFromDir(tmpDir, "bad")
	if l != nil {
		t.Error("Expected nil for bad JSON")
	}
}

func TestLabelFallback(t *testing.T) {
	// Test that Label() falls back to code when _label is missing
	// We can't easily test this without modifying embedded data,
	// but we can test the logic by checking the en locale has a label
	en := Load("en")
	if en == nil {
		t.Fatal("Load(en) returned nil")
	}
	if en.Label() == "" {
		t.Error("Expected non-empty label")
	}
}
