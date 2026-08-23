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

// Package i18n provides locale loading for CryoUtils NG.
// Canonical locale files are embedded via //go:embed i18n/locales/*.json.
// At runtime, files in ~/.cryoutils_ng/locales/*.json override built-in ones
// (same language code → runtime file wins).
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cryoutils-ng/core"
)

//go:embed locales/*.json
var localeFS embed.FS

// Lang holds a single locale's data.
type Lang struct {
	code    string
	data    map[string]any
	label   string
	builtIn bool // true if loaded from embed, false if from runtime override
}

// Available returns all available language codes (built-in + runtime overrides, deduped, sorted, en first).
func Available() []string {
	seen := map[string]bool{}
	var result []string

	// built-in
	for _, f := range localeFiles() {
		code := strings.TrimSuffix(filepath.Base(f), ".json")
		if !seen[code] {
			seen[code] = true
			result = append(result, code)
		}
	}

	// runtime overrides
	overDir := runtimeLocalesDir()
	entries, err := os.ReadDir(overDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			code := strings.TrimSuffix(e.Name(), ".json")
			if !seen[code] {
				seen[code] = true
				result = append(result, code)
			}
		}
	}

	// sort: en first, then alphabetical
	sort.Slice(result, func(i, j int) bool {
		if result[i] == "en" {
			return true
		}
		if result[j] == "en" {
			return false
		}
		return result[i] < result[j]
	})
	return result
}

// Load loads a locale by code. Returns nil if not found.
// Priority: runtime override > built-in.
func Load(code string) *Lang {
	// Try runtime override first
	if l := loadFromDir(runtimeLocalesDir(), code); l != nil {
		return l
	}
	// Fall back to built-in
	if l := loadFromEmbed(code); l != nil {
		return l
	}
	return nil
}

func loadFromDir(dir, code string) *Lang {
	path := filepath.Join(dir, code+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return parseLang(code, data, false)
}

func loadFromEmbed(code string) *Lang {
	name := "locales/" + code + ".json"
	data, err := localeFS.ReadFile(name)
	if err != nil {
		return nil
	}
	return parseLang(code, data, true)
}

func parseLang(code string, data []byte, builtIn bool) *Lang {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Printf("i18n: failed to parse %s.json: %v", code, err)
		return nil
	}
	label := ""
	if raw["_label"] != nil {
		label, _ = raw["_label"].(string)
	}
	return &Lang{code: code, data: raw, label: label, builtIn: builtIn}
}

// T returns the translated string for a dot-notation key.
// Fallback chain: selected language → en built-in → key itself.
func (l *Lang) T(key string) string {
	val := lookup(l.data, key)
	if val != nil {
		s, ok := val.(string)
		if ok {
			return s
		}
	}
	// Fallback to en
	if l.code != "en" {
		en := Load("en")
		if en != nil {
			val = lookup(en.data, key)
			if val != nil {
				s, ok := val.(string)
				if ok {
					return s
				}
			}
		}
	}
	return key
}

// Label returns the display name for this language (from "_label" field; falls back to code).
func (l *Lang) Label() string {
	if l.label != "" {
		return l.label
	}
	return l.code
}

func lookup(m map[string]any, key string) any {
	parts := strings.Split(key, ".")
	cur := any(m)
	for _, p := range parts {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = mm[p]
		if cur == nil {
			return nil
		}
	}
	return cur
}

func localeFiles() []string {
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".json") {
			files = append(files, "locales/"+e.Name())
		}
	}
	return files
}

func runtimeLocalesDir() string {
	return filepath.Join(core.InstallDirectory, "locales")
}

// EnsureRuntimeDir creates the runtime locales directory if it doesn't exist.
func EnsureRuntimeDir() error {
	return os.MkdirAll(runtimeLocalesDir(), 0755)
}

// DumpToRuntime copies a built-in locale to the runtime directory (for user customization).
func DumpToRuntime(code string) error {
	if err := EnsureRuntimeDir(); err != nil {
		return err
	}
	name := "locales/" + code + ".json"
	data, err := localeFS.ReadFile(name)
	if err != nil {
		return fmt.Errorf("locale %s not found in built-in: %w", code, err)
	}
	return os.WriteFile(filepath.Join(runtimeLocalesDir(), code+".json"), data, 0644)
}
