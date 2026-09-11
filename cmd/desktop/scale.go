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

package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
	"time"
)

// maybeReexecWithScale makes the native UI adapt its scale to the display
// resolution. Fyne v2.7.4 has no runtime SetScale() API — the scale is read
// from FYNE_SCALE / Xft.dpi at app initialization — so when no scale was
// chosen we query xrandr, compute a scale, and re-exec ourselves with both
// FYNE_SCALE (read by Fyne) and CRYOUTILS_FYNE_SCALE (loop guard, also read
// by the UI for logical window sizing) set. The re-exec happens before the
// Fyne app exists.
//
// Precedence:
//  1. FYNE_SCALE manually set → respected, no re-exec.
//  2. CRYOUTILS_FYNE_SCALE already set → we are the re-exec'd process, no re-exec (loop guard).
//  3. No DISPLAY (headless) → no re-exec.
//  4. Otherwise xrandr current mode width → scale = clamp(roundTo025(width/1280), 1.0, 2.0).
//
// Mapping: 1280×800→1.0 · 1920×1080→1.5 · 2560×1440→2.0 · 3840×2160→2.0 (cap).
func maybeReexecWithScale() {
	if os.Getenv("FYNE_SCALE") != "" || os.Getenv("CRYOUTILS_FYNE_SCALE") != "" {
		return
	}
	if os.Getenv("DISPLAY") == "" {
		return
	}
	width := currentScreenWidth()
	if width <= 0 {
		return
	}
	scale := computeScale(width)
	if scale == 1.0 {
		// Native-scale display: Fyne's default (1.0 / Xft.dpi) is fine,
		// skip the pointless process restart.
		log.Printf("scale: display width %d px → scale 1.0, no re-exec", width)
		return
	}
	s := strconv.FormatFloat(scale, 'f', -1, 64)

	exe, err := os.Executable()
	if err != nil {
		log.Printf("scale: cannot re-exec (%v), continuing with Fyne default scale", err)
		return
	}
	log.Printf("scale: display width %d px → re-exec with FYNE_SCALE=%s", width, s)
	// FYNE_SCALE is the variable Fyne itself reads at init; CRYOUTILS_FYNE_SCALE
	// is our loop guard + the value the UI reads for logical window sizing.
	if err := syscall.Exec(exe, append([]string{exe}, os.Args[1:]...),
		append(os.Environ(), "FYNE_SCALE="+s, "CRYOUTILS_FYNE_SCALE="+s)); err != nil {
		log.Printf("scale: re-exec failed (%v), continuing with Fyne default scale", err)
	}
}

// computeScale maps a screen width in px to a Fyne scale:
// width/1280 rounded to the nearest 0.25, clamped to [1.0, 2.0].
func computeScale(width int) float64 {
	s := float64(width) / 1280.0 * 4
	// round half away from zero (widths are positive).
	s = float64(int(s+0.5)) / 4
	if s < 1.0 {
		s = 1.0
	}
	if s > 2.0 {
		s = 2.0
	}
	return s
}

// xrandrCurrentRE matches the "current W x H" clause on the xrandr
// "Screen 0:" summary line (e.g. "current 3840 x 2160").
var xrandrCurrentRE = regexp.MustCompile(`current\s+(\d+)\s*x\s*(\d+)`)

// currentScreenWidth reads the active monitor width (px) from xrandr.
// Returns 0 when xrandr is unavailable or unparseable.
func currentScreenWidth() int {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "xrandr").Output()
	if err != nil {
		return 0
	}
	if m := xrandrCurrentRE.FindStringSubmatch(string(out)); m != nil {
		w, _ := strconv.Atoi(m[1])
		return w
	}
	return 0
}
