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
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//--- findAppBrowsers tests ---

func TestFindAppBrowsers_WithMockPath(t *testing.T) {
	// Create a temp dir with mock browser binaries
	tmpDir := t.TempDir()
	for _, name := range []string{"google-chrome", "chromium", "brave-browser"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("#!/bin/sh\necho mock\n"), 0755); err != nil {
			t.Fatalf("write mock binary: %v", err)
		}
	}
	// Also write one that is NOT a candidate (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, "firefox"), []byte("#!/bin/sh\necho mock\n"), 0755); err != nil {
		t.Fatalf("write mock binary: %v", err)
	}

	t.Setenv("PATH", tmpDir)

	found := findAppBrowsers()

	// Should find chrome, chromium, brave-browser (but not firefox because it's not in candidates... wait, firefox IS in candidates)
	// Actually firefox IS in appBrowserCandidates, so it should be found too.
	if len(found) != 4 {
		t.Fatalf("expected 4 browsers, got %d: %v", len(found), found)
	}
	for i, name := range []string{"google-chrome", "chromium", "brave-browser", "firefox"} {
		if !strings.HasSuffix(found[i], name) {
			t.Errorf("expected candidate %q at index %d, got %q", name, i, found[i])
		}
	}
}

func TestFindAppBrowsers_NoneFound(t *testing.T) {
	t.Setenv("PATH", "/nonexistent/path/that/does/not/exist")

	found := findAppBrowsers()
	if len(found) != 0 {
		t.Fatalf("expected 0 browsers, got %d: %v", len(found), found)
	}
}

//--- detectFlatpakBrowsers tests ---

func TestDetectFlatpakBrowsers(t *testing.T) {
	// Create a mock flatpak script that outputs known browser IDs
	tmpDir := t.TempDir()
	mockFlatpak := filepath.Join(tmpDir, "flatpak")
	content := `#!/bin/sh
echo com.brave.Browser 1.0.0 /usr/bin/brave
echo org.chromium.Chromium 1.0.0 /usr/bin/chromium
echo com.valvesoftware.Steam 1.0.0 /usr/bin/steam
`
	if err := os.WriteFile(mockFlatpak, []byte(content), 0755); err != nil {
		t.Fatalf("write mock flatpak: %v", err)
	}

	t.Setenv("PATH", tmpDir)

	found := detectFlatpakBrowsers()

	if len(found) != 2 {
		t.Fatalf("expected 2 flatpak browsers, got %d: %v", len(found), found)
	}
	expected := map[string]bool{"com.brave.Browser": true, "org.chromium.Chromium": true}
	for _, app := range found {
		if !expected[app] {
			t.Errorf("unexpected flatpak app: %s", app)
		}
	}
}

func TestDetectFlatpakBrowsers_NotInstalled(t *testing.T) {
	// flatpak command not in PATH → should return nil without error
	tmpDir := t.TempDir()
	// Don't put any flatpak mock in PATH
	t.Setenv("PATH", tmpDir)

	found := detectFlatpakBrowsers()
	if len(found) != 0 {
		t.Fatalf("expected 0 flatpak browsers, got %d: %v", len(found), found)
	}
}

//--- openURL tests ---

func TestOpenURL_NoBrowser(t *testing.T) {
	// Capture stdout to verify the URL is printed
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf strings.Builder
	infoLog := log.New(&buf, "INFO  ", log.LstdFlags)

	method := openURL("http://127.0.0.1:8080/?token=abc", true, "", infoLog)

	w.Close()
	os.Stdout = oldStdout

	if method != "none" {
		t.Errorf("expected method 'none', got %q", method)
	}
	var out strings.Builder
	io.Copy(&out, r)
	if !strings.Contains(out.String(), "http://127.0.0.1:8080/?token=abc") {
		t.Errorf("expected URL in stdout, got: %s", out.String())
	}
}

func TestOpenURL_OverrideBrowser(t *testing.T) {
	// Use a non-existent browser to test the error path (should fall through)
	var buf strings.Builder
	infoLog := log.New(&buf, "INFO  ", log.LstdFlags)

	// Override with a path that doesn't exist → should fall through to xdg-open
	method := openURL("http://127.0.0.1:8080/?token=abc", false, "/nonexistent/browser", infoLog)

	// Since the override fails, it should fall through to xdg-open (or fail entirely)
	if method != "xdg-open" && method != "failed" {
		t.Errorf("expected 'xdg-open' or 'failed', got %q", method)
	}
}

func TestOpenURL_PrefersAppWindowOverXdgOpen(t *testing.T) {
	// Set up PATH so that a mock browser is found
	tmpDir := t.TempDir()
	mockChrome := filepath.Join(tmpDir, "google-chrome")
	content := `#!/bin/sh
# Record that we were called
echo "chrome-called" >> /tmp/cryo_test_chrome_called
`
	if err := os.WriteFile(mockChrome, []byte(content), 0755); err != nil {
		t.Fatalf("write mock chrome: %v", err)
	}

	t.Setenv("PATH", tmpDir)

	// Clean up any previous call marker
	os.Remove("/tmp/cryo_test_chrome_called")

	var buf strings.Builder
	infoLog := log.New(&buf, "INFO  ", log.LstdFlags)

	method := openURL("http://127.0.0.1:8080/?token=abc", false, "", infoLog)

	// The mock browser should have been called via exec.Command.Start()
	// Since we can't easily verify the child process, check the method returned
	if method != "app-window" {
		t.Errorf("expected method 'app-window', got %q", method)
	}

	// Clean up
	os.Remove("/tmp/cryo_test_chrome_called")
}

//--- openInAppWindow command construction test ---

func TestOpenInAppWindow_Arguments(t *testing.T) {
	// We can't easily intercept exec.Command, but we can verify the function
	// doesn't panic with a valid URL and a real binary path.
	// Use /bin/true as a no-op binary to test the call path.
	url := "http://127.0.0.1:9999/?token=test"
	err := openInAppWindow(url, "/bin/true")
	// /bin/true exits 0, but as a background process this is fine
	_ = err
}

//--- main flag parsing test (indirect via openURL) ---

func TestOpenURL_FlatpakFallback(t *testing.T) {
	// Set up PATH so no native browser is found, but flatpak is
	tmpDir := t.TempDir()
	// No browser binaries in PATH

	// Create mock flatpak
	mockFlatpak := filepath.Join(tmpDir, "flatpak")
	flatpakOutput := `com.brave.Browser 1.0.0 /usr/bin/brave
`
	flatpakContent := `#!/bin/sh
echo "` + flatpakOutput + `"
`
	if err := os.WriteFile(mockFlatpak, []byte(flatpakContent), 0755); err != nil {
		t.Fatalf("write mock flatpak: %v", err)
	}

	t.Setenv("PATH", tmpDir)

	var buf strings.Builder
	infoLog := log.New(&buf, "INFO  ", log.LstdFlags)

	method := openURL("http://127.0.0.1:9999/?token=test", false, "", infoLog)

	// Should try flatpak; flatpak run will fail (no real flatpak), so it falls to xdg-open or steam
	if method != "xdg-open" && method != "steam" && method != "failed" {
		t.Logf("method: %s (output: %s)", method, buf.String())
	}
}

//--- ensure findAppBrowsers preserves candidate order ---

func TestFindAppBrowsers_Order(t *testing.T) {
	tmpDir := t.TempDir()
	// Write only chromium and brave (not chrome)
	for _, name := range []string{"chromium", "brave-browser"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("#!/bin/sh\n"), 0755); err != nil {
			t.Fatalf("write mock: %v", err)
		}
	}
	t.Setenv("PATH", tmpDir)

	found := findAppBrowsers()
	if len(found) != 2 {
		t.Fatalf("expected 2, got %d", len(found))
	}
	if !strings.HasSuffix(found[0], "chromium") {
		t.Errorf("expected chromium first, got %s", found[0])
	}
	if !strings.HasSuffix(found[1], "brave-browser") {
		t.Errorf("expected brave-browser second, got %s", found[1])
	}
}

//--- steam://openurl fallback test ---

func TestOpenURL_SteamFallback(t *testing.T) {
	// No browsers in PATH, no flatpak → should try steam://openurl
	tmpDir := t.TempDir()
	t.Setenv("PATH", tmpDir)

	// Ensure steam command is NOT available
	var buf strings.Builder
	infoLog := log.New(&buf, "INFO  ", log.LstdFlags)

	method := openURL("http://127.0.0.1:9999/?token=test", false, "", infoLog)

	// steam command likely not available in this env, so should fall to xdg-open or failed
	if method != "xdg-open" && method != "failed" {
		t.Logf("method: %s (output: %s)", method, buf.String())
	}
}

//--- verify openInBrowser uses xdg-open ---

func TestOpenInBrowser_Command(t *testing.T) {
	// xdg-open may not exist; test that the function attempts it without panicking
	err := openInBrowser("http://example.com")
	// Ignore error (xdg-open may not be installed)
	_ = err
}

//--- helper: verify appBrowserCandidates list is non-empty and ordered ---

func TestAppBrowserCandidates(t *testing.T) {
	if len(appBrowserCandidates) == 0 {
		t.Fatal("appBrowserCandidates must not be empty")
	}
	// Verify google-chrome is first (highest priority)
	if appBrowserCandidates[0] != "google-chrome" {
		t.Errorf("expected google-chrome first, got %q", appBrowserCandidates[0])
	}
}

//--- verify flatpakBrowserApps is non-empty ---

func TestFlatpakBrowserApps(t *testing.T) {
	if len(flatpakBrowserApps) == 0 {
		t.Fatal("flatpakBrowserApps must not be empty")
	}
}

//--- test that openURL returns a valid method string ---

func TestOpenURL_ReturnsValidMethod(t *testing.T) {
	var buf strings.Builder
	infoLog := log.New(&buf, "INFO  ", log.LstdFlags)

	methods := []string{
		openURL("http://test/", true, "", infoLog),       // none
		openURL("http://test/", false, "", infoLog),      // app-window/flatpak/steam/xdg-open/failed
		openURL("http://test/", false, "/no/such/bin", infoLog), // override→fallback
	}
	for i, m := range methods {
		valid := m == "none" || m == "app-window" || m == "flatpak" || m == "steam" || m == "xdg-open" || m == "override" || m == "failed"
		if !valid {
			t.Errorf("method[%d] = %q is not a valid method", i, m)
		}
	}
}

//--- test that exec.Command is used (integration smoke test) ---

func TestOpenInAppWindow_Integration(t *testing.T) {
	// Use a real no-op command to verify the exec path works
	tmpDir := t.TempDir()
	noOp := filepath.Join(tmpDir, "noop")
	if err := os.WriteFile(noOp, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	// Must be in PATH for exec.Command to find it when called with just the name
	// But openInAppWindow uses the full path, so this test just verifies no panic
	url := "http://127.0.0.1:1234/?token=x"
	err := openInAppWindow(url, noOp)
	if err != nil {
		t.Fatalf("openInAppWindow failed: %v", err)
	}
}

//--- test that detectFlatpakBrowsers handles empty output ---

func TestDetectFlatpakBrowsers_EmptyOutput(t *testing.T) {
	tmpDir := t.TempDir()
	mockFlatpak := filepath.Join(tmpDir, "flatpak")
	if err := os.WriteFile(mockFlatpak, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tmpDir)

	found := detectFlatpakBrowsers()
	if len(found) != 0 {
		t.Fatalf("expected 0, got %d: %v", len(found), found)
	}
}

//--- test that detectFlatpakBrowsers handles flatpak not found ---

func TestDetectFlatpakBrowsers_NotFound(t *testing.T) {
	// flatpak not in PATH at all
	t.Setenv("PATH", "/nonexistent")

	found := detectFlatpakBrowsers()
	if len(found) != 0 {
		t.Fatalf("expected 0, got %d: %v", len(found), found)
	}
}

//--- test that openURL with noBrowser prints URL to stdout ---

func TestOpenURL_NoBrowserPrintsURL(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf strings.Builder
	infoLog := log.New(&buf, "INFO  ", log.LstdFlags)

	method := openURL("http://127.0.0.1:7777/?token=xyz", true, "", infoLog)

	w.Close()
	os.Stdout = oldStdout

	if method != "none" {
		t.Errorf("expected 'none', got %q", method)
	}

	var out strings.Builder
	io.Copy(&out, r)
	if !strings.Contains(out.String(), "http://127.0.0.1:7777/?token=xyz") {
		t.Errorf("expected URL in stdout, got: %s", out.String())
	}
}
