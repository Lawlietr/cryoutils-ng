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

package fyneui

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
)

// contentMaxWidth caps the scroll content width on large displays (4K).
// Without it, selects and buttons stretch across the full ~3750px viewport.
const contentMaxWidth = 1100.0

// cappedCenterLayout wraps the page content: the wrapped object is laid out
// at min(available width, contentMaxWidth) and centered horizontally.
type cappedCenterLayout struct{}

func (l cappedCenterLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	min := objects[0].MinSize()
	if min.Width > contentMaxWidth {
		min.Width = contentMaxWidth
	}
	return min
}

func (l cappedCenterLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	o := objects[0]
	w := size.Width
	if w > contentMaxWidth {
		w = contentMaxWidth
	}
	o.Move(fyne.NewPos((size.Width-w)/2, 0))
	o.Resize(fyne.NewSize(w, o.MinSize().Height))
}

// windowSizeForScreen picks the initial window size (in Fyne logical px, i.e.
// before the display scale is applied):
//   - unknown resolution (headless) or Deck-native 1280×800 → 1280×800 (existing behavior)
//   - larger displays → 80% of the screen in physical px, divided by the scale,
//     so 4K keeps a usable window instead of full 3840px
func windowSizeForScreen() fyne.Size {
	sw, sh := screenDimensions()
	if sw <= 0 {
		return fyne.NewSize(1280, 800)
	}
	if sw <= 1280 && sh <= 800 {
		return fyne.NewSize(1280, 800)
	}
	scale := effectiveScale()
	return fyne.NewSize(float32(sw)*4/5/scale, float32(sh)*4/5/scale)
}

// effectiveScale returns the Fyne scale the app will run at: CRYOUTILS_FYNE_SCALE
// (set by the scale re-exec), else FYNE_SCALE (manual override), else 1.0.
func effectiveScale() float32 {
	for _, key := range []string{"CRYOUTILS_FYNE_SCALE", "FYNE_SCALE"} {
		if v := os.Getenv(key); v != "" {
			if s, err := strconv.ParseFloat(v, 32); err == nil && s > 0 {
				return float32(s)
			}
		}
	}
	return 1.0
}

// xrandrCurrentRE matches the "current W x H" clause on the xrandr
// "Screen 0:" summary line (e.g. "current 3840 x 2160").
var xrandrCurrentRE = regexp.MustCompile(`current\s+(\d+)\s*x\s*(\d+)`)

// screenDimensions reads the active monitor resolution from xrandr.
// Returns (0, 0) when xrandr is unavailable or unparseable.
func screenDimensions() (int, int) {
	if os.Getenv("DISPLAY") == "" {
		return 0, 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "xrandr").Output()
	if err != nil {
		return 0, 0
	}
	if m := xrandrCurrentRE.FindStringSubmatch(string(out)); m != nil {
		w, _ := strconv.Atoi(m[1])
		h, _ := strconv.Atoi(m[2])
		if w > 0 && h > 0 {
			return w, h
		}
	}
	return 0, 0
}
