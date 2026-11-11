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
	"image/color"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// channelDiff sums the per-channel distance between two colours, the same way
// tools/ui1shots/measure.py decides whether a pixel differs from its background.
// A pair below the 30 threshold cannot be told apart in a screenshot.
func channelDiff(a, b color.Color) int {
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	d := func(x, y uint32) int {
		if x > y {
			return int(x - y)
		}
		return int(y - x)
	}
	return d(ar>>8, br>>8) + d(ag>>8, bg>>8) + d(ab>>8, bb>>8)
}

const screenshotTolerance = 30

// contrastRatio is the WCAG luminance ratio, which unlike a channel sum can tell
// a light line on a dark page from a dark line on a dark page.
func contrastRatio(a, b color.Color) float64 {
	lum := func(c color.Color) float64 {
		r, g, bl, _ := c.RGBA()
		chan_ := func(v uint32) float64 {
			s := float64(v>>8) / 255.0
			if s <= 0.03928 {
				return s / 12.92
			}
			return math.Pow((s+0.055)/1.055, 2.4)
		}
		return 0.2126*chan_(r) + 0.7152*chan_(g) + 0.0722*chan_(bl)
	}
	x, y := lum(a), lum(b)
	if x < y {
		x, y = y, x
	}
	return (x + 0.05) / (y + 0.05)
}

// The complaint this step fixes is that six sections read as one column. That
// only works if the box can be seen, so the colours are gated numerically rather
// than by eye.
func TestSectionColoursSurviveTheScreenshotTest(t *testing.T) {
	window := theme.DarkTheme().Color(theme.ColorNameBackground, theme.VariantDark)

	if got := channelDiff(ColorSectionSurface, window); got < screenshotTolerance {
		t.Errorf("card surface differs from the window background by only %d, want >= %d: the sections would still merge into one column",
			got, screenshotTolerance)
	}
	if got := channelDiff(ColorSectionBorder, window); got < screenshotTolerance {
		t.Errorf("card border differs from the window background by only %d, want >= %d", got, screenshotTolerance)
	}
	if got := channelDiff(ColorSectionBorder, ColorSectionSurface); got < screenshotTolerance {
		t.Errorf("card border differs from its own surface by only %d, want >= %d", got, screenshotTolerance)
	}
}

// The theme's own separator name is measurable but weak on the dark theme: it
// resolves to opaque black, so it is a line *darker* than the window background.
// The section stroke has to beat it, otherwise the box would be no better than
// the hairline option the plan rejected.
func TestSectionBorderBeatsTheThemeSeparator(t *testing.T) {
	window := theme.DarkTheme().Color(theme.ColorNameBackground, theme.VariantDark)
	separator := theme.DarkTheme().Color(theme.ColorNameSeparator, theme.VariantDark)

	ratioBorder := contrastRatio(ColorSectionBorder, window)
	ratioSeparator := contrastRatio(separator, window)
	if ratioBorder < 2.0 {
		t.Errorf("section border against the window is only %.2f contrast, want >= 2.0", ratioBorder)
	}
	if ratioBorder <= ratioSeparator {
		t.Errorf("section border contrast %.2f is no better than the theme separator %.2f",
			ratioBorder, ratioSeparator)
	}
	t.Logf("border/window contrast %.2f, theme separator/window contrast %.2f", ratioBorder, ratioSeparator)
}

// A select or entry filled with the input colour must not disappear into the
// card face it stands on. Fyne's own fill fails this, which is why the theme
// wrapper overrides it.
func TestControlsStaySeparableFromTheCardFace(t *testing.T) {
	th := &customTheme{inner: theme.DarkTheme()}
	input := th.Color(theme.ColorNameInputBackground, theme.VariantDark)

	if got := contrastRatio(ColorSectionSurface, input); got < 1.2 {
		t.Errorf("card face and the input fill differ by only %.2f contrast, want >= 1.2: a select would vanish into its card", got)
	}
	if plain := theme.DarkTheme().Color(theme.ColorNameInputBackground, theme.VariantDark); contrastRatio(ColorSectionSurface, plain) >= 1.2 {
		t.Errorf("Fyne's stock input fill no longer needs the override; the override can be dropped")
	}
}

func TestSectionBoxPutsPaddingAroundTitleAndBody(t *testing.T) {
	body := widget.NewLabel("body")
	box := sectionCard("Status", body)

	// Measure the objects the box actually built: the title is bold, so a plain
	// label of the same text would be a few points narrower and the arithmetic
	// would look wrong.
	headMin, bodyMin := box.Objects[1].MinSize(), box.Objects[2].MinSize()
	wid := headMin.Width
	if bodyMin.Width > wid {
		wid = bodyMin.Width
	}
	wantWidth := 2*SpaceCardPad + wid
	wantHeight := 2*SpaceCardPad + headMin.Height + SpaceCardRowGap + bodyMin.Height

	min := box.MinSize()
	if min.Width != wantWidth {
		t.Errorf("MinSize().Width = %v, want %v (the §4 card padding on both sides)", min.Width, wantWidth)
	}
	if min.Height != wantHeight {
		t.Errorf("MinSize().Height = %v, want %v (card padding + title + row gap + body)", min.Height, wantHeight)
	}
}

// The background rectangle has to be stretched to the box, otherwise the stroke
// sits on a rectangle of its own size and stops matching the content.
func TestSectionBoxStretchesTheSurfaceToTheBox(t *testing.T) {
	box := sectionCard("Status", widget.NewLabel("body"))
	size := fyne.NewSize(900, 400)
	box.Resize(size)

	surface := box.Objects[0]
	if got := surface.Size(); got.Width != size.Width || got.Height != size.Height {
		t.Errorf("surface size = %v, want the full box %v", got, size)
	}

	head, body := box.Objects[1], box.Objects[2]
	if got := head.Position(); got.X != SpaceCardPad || got.Y != SpaceCardPad {
		t.Errorf("title position = %v, want (%v,%v) = the card padding", got, SpaceCardPad, SpaceCardPad)
	}
	wantBodyY := SpaceCardPad + head.MinSize().Height + SpaceCardRowGap
	if got := body.Position(); got.Y != wantBodyY {
		t.Errorf("body y = %v, want %v (padding + title + §4 row gap)", got.Y, wantBodyY)
	}
}

func TestStackSectionsSeparatesBoxesByTheCardGap(t *testing.T) {
	a := sectionCard("A", widget.NewLabel("a"))
	b := sectionCard("B", widget.NewLabel("b"))
	cc := sectionCard("C", widget.NewLabel("c"))
	page := stackSections(a, b, cc)

	if len(page.Objects) != 5 {
		t.Fatalf("stackSections built %d object(s), want 5 (three boxes and two gaps)", len(page.Objects))
	}
	for _, i := range []int{1, 3} {
		gap, ok := page.Objects[i].(*sectionGap)
		if !ok {
			t.Fatalf("object %d is %T, want a section gap", i, page.Objects[i])
		}
		if gap.MinSize().Height != SpaceCardGap {
			t.Errorf("gap height = %v, want %v (--s-6)", gap.MinSize().Height, SpaceCardGap)
		}
	}
	want := []fyne.CanvasObject{a, b, cc}
	for _, i := range []int{0, 2, 4} {
		if page.Objects[i] != want[i/2] {
			t.Errorf("object %d is not the section box it should be", i)
		}
	}
}

// sectionBody returns the body object of a section box, so the other tests do
// not each carry their own cast. It fails loudly when a section stops using the
// box, which keeps the tests honest about the page structure.
func sectionBody(t *testing.T, section fyne.CanvasObject) fyne.CanvasObject {
	t.Helper()
	box, ok := section.(*fyne.Container)
	if !ok {
		t.Fatalf("section is %T, want the section box container", section)
	}
	if len(box.Objects) != 3 {
		t.Fatalf("section box has %d object(s), want 3 (surface, title, body)", len(box.Objects))
	}
	return box.Objects[2]
}
