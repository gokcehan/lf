// Copyright 2015 The Tcell Authors
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
	"sync"

	"github.com/gdamore/tcell/v2"
)

// TextBar is a Widget that provides a single line of text, but with
// distinct left, center, and right areas.  Each of the areas can be styled
// differently, and they align to the left, center, and right respectively.
// This is basically a convenience type on top Text and BoxLayout.
type TextBar struct {
	changed bool
	style   tcell.Style
	left    Text
	right   Text
	center  Text
	view    View
	lview   ViewPort
	rview   ViewPort
	cview   ViewPort
	once    sync.Once

	WidgetWatchers
}

// SetCenter sets the center text for the textbar.  The text is
// always center aligned.
func (t *TextBar) SetCenter(s string, style tcell.Style) {
	t.initialize()
	if style == tcell.StyleDefault {
		style = t.style
	}
	t.center.SetText(s)
	t.center.SetStyle(style)
}

// SetLeft sets the left text for the textbar.  It is always left-aligned.
func (t *TextBar) SetLeft(s string, style tcell.Style) {
	t.initialize()
	if style == tcell.StyleDefault {
		style = t.style
	}
	t.left.SetText(s)
	t.left.SetStyle(style)
}

// SetRight sets the right text for the textbar.  It is always right-aligned.
func (t *TextBar) SetRight(s string, style tcell.Style) {
	t.initialize()
	if style == tcell.StyleDefault {
		style = t.style
	}
	t.right.SetText(s)
	t.right.SetStyle(style)
}

// SetStyle is used to set a default style to use for the textbar, including
// areas where no text is present.  Note that this will not change the text
// already displayed, so call this before changing or setting text.
func (t *TextBar) SetStyle(style tcell.Style) {
	t.initialize()
	t.style = style
}

func (t *TextBar) initialize() {
	t.once.Do(func() {
		t.center.SetView(&t.cview)
		t.left.SetView(&t.lview)
		t.right.SetView(&t.rview)
		t.center.SetAlignment(VAlignTop | HAlignCenter)
		t.left.SetAlignment(VAlignTop | HAlignLeft)
		t.right.SetAlignment(VAlignTop | HAlignRight)
		t.center.Watch(t)
		t.left.Watch(t)
		t.right.Watch(t)
	})
}

func (t *TextBar) layout() {
	w, _ := t.view.Size()
	ww, wh := t.left.Size()
	t.lview.Resize(0, 0, ww, wh)

	ww, wh = t.center.Size()
	t.cview.Resize((w-ww)/2, 0, ww, wh)

	ww, wh = t.right.Size()
	t.rview.Resize(w-ww, 0, ww, wh)

	t.changed = false
}

// SetView sets the View drawing context for this TextBar.
func (t *TextBar) SetView(view View) {
	t.initialize()
	t.view = view
	t.lview.SetView(view)
	t.rview.SetView(view)
	t.cview.SetView(view)
	t.changed = true
}

// Draw draws the TextBar into its View context.
func (t *TextBar) Draw() {

	t.initialize()
	if t.changed {
		t.layout()
	}
	w, h := t.view.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			t.view.SetContent(x, y, ' ', nil, t.style)
		}
	}

	// Draw in reverse order -- if we clip, we will clip at the
	// right side.
	t.right.Draw()
	t.center.Draw()
	t.left.Draw()
}

// Resize is called when the TextBar's View changes size, and
// updates the layout.
func (t *TextBar) Resize() {
	t.initialize()
	t.layout()

	t.left.Resize()
	t.center.Resize()
	t.right.Resize()

	t.PostEventWidgetResize(t)
}

// Size implements the Size method for Widget, returning the width
// and height in character cells.
func (t *TextBar) Size() (int, int) {
	w, h := 0, 0

	ww, wh := t.left.Size()
	w += ww
	if wh > h {
		h = wh
	}
	ww, wh = t.center.Size()
	w += ww
	if wh > h {
		h = wh
	}
	ww, wh = t.right.Size()
	w += ww
	if wh > h {
		h = wh
	}
	return w, h
}

// HandleEvent handles incoming events.  The only events handled are
// those for the Text objects; when those change, the TextBar adjusts
// the layout to accommodate.
func (t *TextBar) HandleEvent(ev tcell.Event) bool {
	switch ev.(type) {
	case *EventWidgetContent:
		t.changed = true
		return true
	}
	return false
}

// NewTextBar creates an empty, initialized TextBar.
func NewTextBar() *TextBar {
	t := &TextBar{}
	t.initialize()
	return t
}
