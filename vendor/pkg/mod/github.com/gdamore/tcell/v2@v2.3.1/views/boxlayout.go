// Copyright 2016 The Tcell Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package views

import (
	"github.com/gdamore/tcell/v2"
)

// BoxLayout is a container Widget that lays out its child widgets in
// either a horizontal row or a vertical column.
type BoxLayout struct {
	view    View
	orient  Orientation
	style   tcell.Style // backing style
	cells   []*boxLayoutCell
	width   int
	height  int
	changed bool

	WidgetWatchers
}

type boxLayoutCell struct {
	widget Widget
	fill   float64 // fill factor - 0.0 means no expansion
	pad    int     // count of padding spaces (stretch)
	frac   float64 // calculated residual spacing, used internally
	view   *ViewPort
}

func (b *BoxLayout) hLayout() {
	w, h := b.view.Size()

	totf := 0.0
	for _, c := range b.cells {
		x, y := c.widget.Size()
		totf += c.fill
		b.width += x
		if y > b.height {
			b.height = y
		}
		c.pad = 0
		c.frac = 0
	}

	extra := w - b.width
	if extra < 0 {
		extra = 0
	}
	resid := extra
	if totf == 0 {
		resid = 0
	}

	for _, c := range b.cells {
		if c.fill > 0 {
			c.frac = float64(extra) * c.fill / totf
			c.pad = int(c.frac)
			c.frac -= float64(c.pad)
			resid -= c.pad
		}
	}

	// Distribute any left over padding.  We try to give it to the
	// the cells with the highest residual fraction.  It should be
	// the case that no single cell gets more than one more cell.
	for resid > 0 {
		var best *boxLayoutCell
		for _, c := range b.cells {
			if c.fill == 0 {
				continue
			}
			if best == nil || c.frac > best.frac {
				best = c
			}
		}
		best.pad++
		best.frac = 0
		resid--
	}

	x, y, xinc := 0, 0, 0
	for _, c := range b.cells {
		cw, _ := c.widget.Size()

		xinc = cw + c.pad
		cw += c.pad

		c.view.Resize(x, y, cw, h)
		c.widget.Resize()
		x += xinc
	}
}

func (b *BoxLayout) vLayout() {
	w, h := b.view.Size()

	totf := 0.0
	for _, c := range b.cells {
		x, y := c.widget.Size()
		b.height += y
		totf += c.fill
		if x > b.width {
			b.width = x
		}
		c.pad = 0
		c.frac = 0
	}

	extra := h - b.height
	if extra < 0 {
		extra = 0
	}

	resid := extra
	if totf == 0 {
		resid = 0
	}

	for _, c := range b.cells {
		if c.fill > 0 {
			c.frac = float64(extra) * c.fill / totf
			c.pad = int(c.frac)
			c.frac -= float64(c.pad)
			resid -= c.pad
		}
	}

	// Distribute any left over padding.  We try to give it to the
	// the cells with the highest residual fraction.  It should be
	// the case that no single cell gets more than one more cell.
	for resid > 0 {
		var best *boxLayoutCell
		for _, c := range b.cells {
			if c.fill == 0 {
				continue
			}
			if best == nil || c.frac > best.frac {
				best = c
			}
		}
		best.pad++
		best.frac = 0
		resid--
	}

	x, y, yinc := 0, 0, 0
	for _, c := range b.cells {
		_, ch := c.widget.Size()

		yinc = ch + c.pad
		ch += c.pad
		c.view.Resize(x, y, w, ch)
		c.widget.Resize()
		y += yinc
	}
}

func (b *BoxLayout) layout() {
	if b.view == nil {
		return
	}
	b.width, b.height = 0, 0
	switch b.orient {
	case Horizontal:
		b.hLayout()
	case Vertical:
		b.vLayout()
	default:
		panic("Bad orientation")
	}
	b.changed = false
}

// Resize adjusts the layout when the underlying View changes size.
func (b *BoxLayout) Resize() {
	b.layout()

	// Now also let the children know we resized.
	for i := range b.cells {
		b.cells[i].widget.Resize()
	}
	b.PostEventWidgetResize(b)
}

// Draw is called to update the displayed content.
func (b *BoxLayout) Draw() {

	if b.view == nil {
		return
	}
	if b.changed {
		b.layout()
	}
	b.view.Fill(' ', b.style)
	for i := range b.cells {
		b.cells[i].widget.Draw()
	}
}

// Size returns the preferred size in character cells (width, height).
func (b *BoxLayout) Size() (int, int) {
	return b.width, b.height
}

// SetView sets the View object used for the text bar.
func (b *BoxLayout) SetView(view View) {
	b.changed = true
	b.view = view
	for _, c := range b.cells {
		c.view.SetView(view)
	}
}

// HandleEvent implements a tcell.EventHandler.  The only events
// we care about are Widget change events from our children. We
// watch for those so that if the child changes, we can arrange
// to update our layout.
func (b *BoxLayout) HandleEvent(ev tcell.Event) bool {
	switch ev.(type) {
	case *EventWidgetContent:
		// This can only have come from one of our children.
		b.changed = true
		b.PostEventWidgetContent(b)
		return true
	}
	for _, c := range b.cells {
		if c.widget.HandleEvent(ev) {
			return true
		}
	}
	return false
}

// AddWidget adds a widget to the end of the BoxLayout.
func (b *BoxLayout) AddWidget(widget Widget, fill float64) {
	c := &boxLayoutCell{
		widget: widget,
		fill:   fill,
		view:   NewViewPort(b.view, 0, 0, 0, 0),
	}
	widget.SetView(c.view)
	b.cells = append(b.cells, c)
	b.changed = true
	widget.Watch(b)
	b.layout()
	b.PostEventWidgetContent(b)
}

// InsertWidget inserts a widget at the given offset.  Offset 0 is the
// front.  If the index is longer than the number of widgets, then it
// just gets appended to the end.
func (b *BoxLayout) InsertWidget(index int, widget Widget, fill float64) {
	c := &boxLayoutCell{
		widget: widget,
		fill:   fill,
		view:   NewViewPort(b.view, 0, 0, 0, 0),
	}
	c.widget.SetView(c.view)
	if index < 0 {
		index = 0
	}
	if index > len(b.cells) {
		index = len(b.cells)
	}
	b.cells = append(b.cells, c)
	copy(b.cells[index+1:], b.cells[index:])
	b.cells[index] = c
	widget.Watch(b)
	b.layout()
	b.PostEventWidgetContent(b)
}

// RemoveWidget removes a Widget from the layout.
func (b *BoxLayout) RemoveWidget(widget Widget) {
	changed := false
	for i := 0; i < len(b.cells); i++ {
		if b.cells[i].widget == widget {
			b.cells = append(b.cells[:i], b.cells[i+1:]...)
			changed = true
		}
	}
	if !changed {
		return
	}
	b.changed = true
	widget.Unwatch(b)
	b.layout()
	b.PostEventWidgetContent(b)
}

// Widgets returns the list of Widgets for this BoxLayout.
func (b *BoxLayout) Widgets() []Widget {
	w := make([]Widget, 0, len(b.cells))
	for _, c := range b.cells {
		w = append(w, c.widget)
	}
	return w
}

// SetOrientation sets the orientation as either Horizontal or Vertical.
func (b *BoxLayout) SetOrientation(orient Orientation) {
	if b.orient != orient {
		b.orient = orient
		b.changed = true
		b.PostEventWidgetContent(b)
	}
}

// SetStyle sets the style used.
func (b *BoxLayout) SetStyle(style tcell.Style) {
	b.style = style
	b.PostEventWidgetContent(b)
}

// NewBoxLayout creates an empty BoxLayout.
func NewBoxLayout(orient Orientation) *BoxLayout {
	return &BoxLayout{orient: orient}
}
