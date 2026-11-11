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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Design tokens. The single source of truth is docs/ui-redesign-plan.md §4.
// The original complaint ("设置挤在一起、不整齊") traced back to 15 different
// spacing values and 11 font sizes with no ladder behind them, so a value that
// is not in one of the ladders below must not be added. All numbers are logical
// px; Fyne multiplies every theme size by the runtime scale, so scale 2 doubles
// them without a second set of numbers.

// Spacing ladder, §4.1 (8 px grid).
const (
	SpaceLabelValue = 4.0  // --s-1 a label and the value that belongs to it
	SpaceControlPad = 8.0  // --s-2 padding inside a control, vertical
	SpaceRowGap     = 12.0 // --s-3 neighbours inside one row
	SpaceCardRowGap = 16.0 // --s-4 rows inside one card
	SpaceCardPad    = 24.0 // --s-5 padding inside a card
	SpaceCardGap    = 32.0 // --s-6 between cards
	SpaceSectionGap = 48.0 // --s-7 between sections, page top and bottom
)

// Type ladder, §4.2 (ratio 1.25).
const (
	FontSizeMicro = 12.0 // --fs-micro status labels
	FontSizeSmall = 14.0 // --fs-sm hints and the version string
	FontSizeBody  = 16.0 // --fs-body labels, entries, button text
	FontSizeValue = 20.0 // --fs-md status values
	FontSizeTitle = 25.0 // --fs-lg section titles
	FontSizePage  = 31.0 // --fs-xl page title
)

// Control rules, §4.3.
const (
	ControlHeightTouch = 44.0  // Steam Deck touch target
	ControlHeightDesk  = 36.0  // desktop
	ControlMaxWidth    = 320.0 // select and entry width cap, never full width
	ButtonMinWidth     = 96.0
	SeparatorWidth     = 1.0
)

// spacingLadder and typeLadder exist so the tests can reject a one-off value.
var spacingLadder = []float32{
	SpaceLabelValue, SpaceControlPad, SpaceRowGap, SpaceCardRowGap,
	SpaceCardPad, SpaceCardGap, SpaceSectionGap,
}

var typeLadder = []float32{
	FontSizeMicro, FontSizeSmall, FontSizeBody, FontSizeValue, FontSizeTitle, FontSizePage,
}

// tokenSize maps the ladders onto the Fyne size names.
//
// Fyne v2.7.4 exposes only two padding steps (theme.go has `padding` and
// `innerPadding`), so --s-3 … --s-7 cannot be expressed as theme sizes: those
// gaps are applied by the containers themselves with the constants above.
// `fallback` is the value the wrapped theme would have returned.
func tokenSize(name fyne.ThemeSizeName, fallback float32) float32 {
	switch name {
	case theme.SizeNameText:
		return FontSizeBody
	case theme.SizeNameCaptionText:
		return FontSizeMicro
	case theme.SizeNameSubHeadingText:
		return FontSizeTitle
	case theme.SizeNameHeadingText:
		return FontSizePage
	case theme.SizeNamePadding:
		return SpaceLabelValue
	case theme.SizeNameInnerPadding:
		return SpaceControlPad
	case theme.SizeNameSeparatorThickness:
		return SeparatorWidth
	}
	// FontSizeValue (--fs-md) has no Fyne size name: the status values carry an
	// explicit text style instead, because Fyne has no "value text" role.
	return fallback
}

// customTheme wraps a built-in theme and applies the design tokens plus the
// embedded CJK fonts. v2.7.4 has no theme constructor to build on top of
// another theme, so this wrapper is the extension point.
type customTheme struct {
	inner fyne.Theme
}

func (c *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return c.inner.Color(name, variant)
}

func (c *customTheme) Font(style fyne.TextStyle) fyne.Resource {
	return cjkFont(style)
}

func (c *customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return c.inner.Icon(name)
}

func (c *customTheme) Size(name fyne.ThemeSizeName) float32 {
	return tokenSize(name, c.inner.Size(name))
}
