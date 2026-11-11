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
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// sectionCard is the bordered box around every section of the page.
//
// It replaces widget.Card for two reasons. Fyne's Card renderer is
// NewShadowingRenderer(objects, CardLevel), so a Card has a shadow and no border:
// on the dark window background the six sections ran into one long column
// (docs/ui-redesign-plan.md §2.7, fix option (b) in §5.6). Card also ignores the
// §4 padding ladder, while the box below takes its padding from the tokens.
//
// The rectangle is the first object in the container so the layout can stretch it
// to the box size; the stroke therefore always matches the content instead of
// being a second, separately measured frame.
func sectionCard(title string, body fyne.CanvasObject) *fyne.Container {
	head := widget.NewLabel(title)
	head.TextStyle = fyne.TextStyle{Bold: true}

	surface := canvas.NewRectangle(ColorSectionSurface)
	surface.CornerRadius = RadiusCard
	// v2.7.4 has no SetStroke on Rectangle; the stroke is a pair of fields.
	surface.StrokeColor = ColorSectionBorder
	surface.StrokeWidth = SeparatorWidth

	return container.New(&sectionLayout{}, surface, head, body)
}

// sectionGap is a fixed-height spacer between section boxes.
//
// container.NewVBox spaces its children by theme.Padding(), which is 4 px - far
// below the §4 card gap, and two boxes with a 4 px seam read as one block with a
// doubled stroke. The spacer makes the §4 value actual instead of intended.
type sectionGap struct {
	canvas.Rectangle
	height float32
}

func newSectionGap(height float32) *sectionGap {
	g := &sectionGap{height: height}
	g.Rectangle.Resize(fyne.NewSize(0, height))
	return g
}

func (g *sectionGap) MinSize() fyne.Size { return fyne.NewSize(0, g.height) }

// stackSections lays sections out in one column with the §4 card gap between them.
func stackSections(items ...fyne.CanvasObject) *fyne.Container {
	objs := make([]fyne.CanvasObject, 0, len(items)*2)
	for i, item := range items {
		if i > 0 {
			objs = append(objs, newSectionGap(SpaceCardGap))
		}
		objs = append(objs, item)
	}
	return container.NewVBox(objs...)
}

// sectionLayout puts the title above the body inside the card padding, and gives
// the background rectangle the whole box.
type sectionLayout struct{}

func (l *sectionLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 3 {
		return fyne.NewSize(0, 0)
	}
	headMin, bodyMin := objects[1].MinSize(), objects[2].MinSize()
	wid := headMin.Width
	if bodyMin.Width > wid {
		wid = bodyMin.Width
	}
	return fyne.NewSize(
		2*SpaceCardPad+wid,
		2*SpaceCardPad+headMin.Height+SpaceCardRowGap+bodyMin.Height)
}

func (l *sectionLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 3 {
		return
	}
	surface, head, body := objects[0], objects[1], objects[2]
	surface.Resize(size)

	inner := size.Width - 2*SpaceCardPad
	head.Move(fyne.NewPos(SpaceCardPad, SpaceCardPad))
	head.Resize(fyne.NewSize(inner, head.MinSize().Height))

	bodyY := SpaceCardPad + head.MinSize().Height + SpaceCardRowGap
	body.Move(fyne.NewPos(SpaceCardPad, bodyY))
	body.Resize(fyne.NewSize(inner, size.Height-bodyY-SpaceCardPad))
}
