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
	"math"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// The UI-1 plan (§4) exists because the old page had 15 spacing values and 11
// font sizes. These tests fail if a token drifts back to a one-off number.

func TestDesignTokensOverrideFyneDefaults(t *testing.T) {
	th := &customTheme{inner: theme.DarkTheme()}
	want := map[fyne.ThemeSizeName]float32{
		theme.SizeNameText:               FontSizeBody,
		theme.SizeNameCaptionText:        FontSizeMicro,
		theme.SizeNameSubHeadingText:     FontSizeTitle,
		theme.SizeNameHeadingText:        FontSizePage,
		theme.SizeNamePadding:            SpaceLabelValue,
		theme.SizeNameInnerPadding:       SpaceControlPad,
		theme.SizeNameSeparatorThickness: SeparatorWidth,
	}
	for name, expect := range want {
		if got := th.Size(name); got != expect {
			t.Errorf("Size(%s) = %v, want %v (the §4 token)", name, got, expect)
		}
	}

	// Text must actually be bigger than Fyne's default, otherwise the token
	// table is decorative. Fyne v2.7.4 ships text = 14.
	if got := theme.DarkTheme().Size(theme.SizeNameText); got >= FontSizeBody {
		t.Errorf("Fyne default text is %v, the token (%v) is not an upgrade",
			got, FontSizeBody)
	}
}

func TestUnmappedSizesFallThrough(t *testing.T) {
	inner := theme.DarkTheme()
	th := &customTheme{inner: inner}
	for _, name := range []fyne.ThemeSizeName{
		theme.SizeNameScrollBar, theme.SizeNameInlineIcon,
		theme.SizeNameInputRadius, theme.SizeNameInputBorder,
	} {
		if got, base := th.Size(name), inner.Size(name); got != base {
			t.Errorf("Size(%s) = %v, want the inner theme's %v", name, got, base)
		}
	}
}

func TestSpacingLadder(t *testing.T) {
	if len(spacingLadder) != 7 {
		t.Fatalf("spacing ladder has %d values, §4.1 defines 7", len(spacingLadder))
	}
	for i, v := range spacingLadder {
		if math.Mod(float64(v), 4) != 0 {
			t.Errorf("spacing ladder[%d] = %v is not on the 8 px grid (§4.1)", i, v)
		}
		if i > 0 && v <= spacingLadder[i-1] {
			t.Errorf("spacing ladder is not increasing at index %d: %v after %v", i, v, spacingLadder[i-1])
		}
	}
}

func TestTypeLadderFollowsOnePointTwoFive(t *testing.T) {
	if len(typeLadder) != 6 {
		t.Fatalf("type ladder has %d values, §4.2 defines 6", len(typeLadder))
	}
	// The ladder is 1.25 from --fs-body upward: 16 → 20 → 25 → 31.
	for _, pair := range [][2]float32{
		{FontSizeBody, FontSizeValue},
		{FontSizeValue, FontSizeTitle},
		{FontSizeTitle, FontSizePage},
	} {
		got := pair[1] / pair[0]
		if math.Abs(float64(got-1.25)) > 0.02 {
			t.Errorf("%v → %v is a ratio of %.3f, want 1.25", pair[0], pair[1], got)
		}
	}
	if FontSizeSmall >= FontSizeBody || FontSizeMicro >= FontSizeSmall {
		t.Errorf("the two smaller steps must stay below --fs-body: %v %v %v",
			FontSizeMicro, FontSizeSmall, FontSizeBody)
	}
}

func TestTokenSizeNeverInventsAValue(t *testing.T) {
	allowed := append(append([]float32{}, spacingLadder...), typeLadder...)
	allowed = append(allowed, SeparatorWidth)
	mapped := []fyne.ThemeSizeName{
		theme.SizeNameText, theme.SizeNameCaptionText, theme.SizeNameSubHeadingText,
		theme.SizeNameHeadingText, theme.SizeNamePadding, theme.SizeNameInnerPadding,
		theme.SizeNameSeparatorThickness,
	}
	for _, name := range mapped {
		got := tokenSize(name, -1)
		found := false
		for _, v := range allowed {
			if got == v {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Size(%s) returned %v, which is not in either ladder (§4.1, §4.2)", name, got)
		}
	}
}

// The theme wrapper must keep the CJK font substitution that stops Chinese text
// from rendering as boxes.
func TestCJKFontSurvivesTheThemeWrapper(t *testing.T) {
	th := &customTheme{inner: theme.DarkTheme()}
	if got := th.Font(fyne.TextStyle{}); got == nil || !strings.Contains(got.Name(), "NotoSansTC") {
		t.Errorf("regular font = %v, want the embedded Noto Sans TC", got)
	}
	if got := th.Font(fyne.TextStyle{Bold: true}); got == nil || !strings.Contains(got.Name(), "Bold") {
		t.Errorf("bold font = %v, want the embedded bold face", got)
	}
}
