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
	"embed"
	"image/color"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed fonts/NotoSansTC-Regular.ttf fonts/NotoSansTC-Bold.ttf
var fontFS embed.FS

var notoRegular *fyne.StaticResource
var notoBold *fyne.StaticResource

func init() {
	r, err := fontFS.Open("fonts/NotoSansTC-Regular.ttf")
	if err == nil {
		data, _ := io.ReadAll(r)
		r.Close()
		notoRegular = fyne.NewStaticResource("fonts/NotoSansTC-Regular.ttf", data)
	}
	b, err := fontFS.Open("fonts/NotoSansTC-Bold.ttf")
	if err == nil {
		data, _ := io.ReadAll(b)
		b.Close()
		notoBold = fyne.NewStaticResource("fonts/NotoSansTC-Bold.ttf", data)
	}
}

// customTheme wraps the default theme and substitutes CJK fonts.
type customTheme struct {
	inner fyne.Theme
}

func (c *customTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (c *customTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Bold {
		if notoBold != nil {
			return notoBold
		}
	}
	if notoRegular != nil {
		return notoRegular
	}
	return theme.DefaultTheme().Font(style)
}

func (c *customTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (c *customTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n)
}
