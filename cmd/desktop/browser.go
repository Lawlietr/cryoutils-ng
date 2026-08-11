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

package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// appBrowserCandidates lists Chromium-family browser names to try via LookPath,
// in priority order. These open URLs as chromeless app windows via --app=.
var appBrowserCandidates = []string{
	"google-chrome",
	"chromium",
	"chromium-browser",
	"microsoft-edge",
	"brave-browser",
	"brave",
	"firefox", // fallback: --new-window (no --app support)
}

// flatpakBrowserApps lists Flatpak app IDs that provide a Chromium-compatible browser.
var flatpakBrowserApps = []string{
	"com.brave.Browser",
	"org.chromium.Chromium",
	"com.google.Chrome",
	"com.microsoft.Edge",
}

// findAppBrowsers returns the absolute paths of browser executables found in PATH,
// limited to the candidate list. Order is preserved from candidateCandidates.
func findAppBrowsers() []string {
	var found []string
	for _, name := range appBrowserCandidates {
		if path, err := exec.LookPath(name); err == nil {
			found = append(found, path)
		}
	}
	return found
}

// detectFlatpakBrowsers returns Flatpak browser app IDs available on the system.
// It runs `flatpak list --app` and matches against known browser IDs.
func detectFlatpakBrowsers() []string {
	cmd := exec.Command("flatpak", "list", "--app")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var found []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// flatpak list --app output: <app-id> <version> <...>
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		appID := parts[0]
		for _, known := range flatpakBrowserApps {
			if appID == known {
				found = append(found, known)
				break
			}
		}
	}
	return found
}

// openInAppWindow opens the given URL in a Chromium-family browser as a
// chromeless app window (--new-window --app=<url>).
// browserPath must be an absolute path to a supported browser binary.
func openInAppWindow(url, browserPath string) error {
	args := []string{"--new-window", "--app=" + url}
	cmd := exec.Command(browserPath, args...)
	return cmd.Start()
}

// openInBrowser opens the given URL using xdg-open as a last-resort fallback.
func openInBrowser(url string) error {
	cmd := exec.Command("xdg-open", url)
	return cmd.Start()
}

// openURL orchestrates the browser launch fallback chain:
//   1. If noBrowser is true, prints the URL and returns "none".
//   2. If overrideBrowser is set, uses it directly (--app=).
//   3. Tries native app-browser candidates (LookPath).
//   4. Tries Flatpak browser apps.
//   5. Falls back to steam://openurl (via steam command).
//   6. Falls back to xdg-open.
//
// Returns the method that was used (e.g. "chrome-app", "flatpak", "steam", "xdg-open", "none").
func openURL(url string, noBrowser bool, overrideBrowser string, infoLog *log.Logger) string {
	if noBrowser {
		fmt.Println(url)
		return "none"
	}

	if overrideBrowser != "" {
		if err := openInAppWindow(url, overrideBrowser); err != nil {
			infoLog.Printf("Failed to open override browser %q: %v", overrideBrowser, err)
		} else {
			infoLog.Printf("Opened URL in override browser: %s", overrideBrowser)
			return "override"
		}
	}

	// Try native app-browser candidates
	candidates := findAppBrowsers()
	for _, path := range candidates {
		if err := openInAppWindow(url, path); err != nil {
			infoLog.Printf("Failed to open %q: %v", path, err)
			continue
		}
		infoLog.Printf("Opened URL in app window: %s", path)
		return "app-window"
	}

	// Try Flatpak browsers
	flatpakApps := detectFlatpakBrowsers()
	for _, appID := range flatpakApps {
		cmd := exec.Command("flatpak", "run", "--branch=stable", "--arch=x86_64", appID, "--new-window", "--app=" + url)
		if err := cmd.Start(); err != nil {
			infoLog.Printf("Failed to open Flatpak browser %q: %v", appID, err)
			continue
		}
		infoLog.Printf("Opened URL in Flatpak browser: %s", appID)
		return "flatpak"
	}

	// Try steam://openurl (Steam is guaranteed to exist on Deck)
	steamURL := "steam://openurl/" + url
	cmd := exec.Command("steam", steamURL)
	if err := cmd.Start(); err != nil {
		infoLog.Printf("steam://openurl not available: %v", err)
	} else {
		infoLog.Printf("Opened URL via steam://openurl: %s", steamURL)
		return "steam"
	}

	// Final fallback: xdg-open
	if err := openInBrowser(url); err != nil {
		infoLog.Printf("All browser methods failed, URL: %s", url)
		return "failed"
	}
	infoLog.Printf("Opened URL via xdg-open: %s", url)
	return "xdg-open"
}
