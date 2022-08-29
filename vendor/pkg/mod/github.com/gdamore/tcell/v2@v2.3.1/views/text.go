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
	"github.com/mattn/go-runewidth"

	"github.com/gdamore/tcell/v2"
)

// Text is a Widget with containing a block of text, which can optionally
// be styled.
type Text struct {
	view    View
	align   Alignment
	style   tcell.Style
	text    []rune
	widths  []int
	styles  []tcell.Style
	lengths []int
	width   int
	height  int

	WidgetWatchers
}

func (t *Text) clear() {
	v := t.view
	w, h := v.Size()
	v.Clear()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v.SetContent(x, y, ' ', nil, t.style)
		}
	}
}

// calcY figures the initial Y offset.  Alignment is top by default.
func (t *Text) calcY(height int) int {
	if t.align&VAlignCenter != 0 {
		return (height - len(t.lengths)) / 2
	}
	if t.align&VAlignBottom != 0 {
		return height - len(t.lengths)
	}
	return 0
}

// calcX figures the initial X offset for the given line.
// Alignment is left by default.
func (t *Text) calcX(width, line int) int {
	if t.align&HAlignCenter != 0 {
		return (width - t.lengths[line]) / 2
	}
	if t.align&HAlignRight != 0 {
		return width - t.lengths[line]
	}
	return 0
}

// Draw draws the Text.
func (t *Text) Draw() {
	v := t.view
	if v == nil {
		return
	}

	width, height := v.Size()
	if width == 0 || height == 0 {
		return
	}

	t.clear()

	// Note that we might wind up with a negative X if the width
	// is larger than the length.  That's OK, and correct even.
	// The view will clip it properly in that case.

	// We align to the left & top by default.
	y := t.calcY(height)
	r := rune(0)
	w := 0
	x := 0
	var styl tcell.Style
	var comb []rune
	line := 0
	newline := true
	for i, l := range t.text {

		if newline {
			x = t.calcX(width, line)
			newline = false
		}
		if l == '\n' {
			if w != 0 {
				v.SetContent(x, y, r, comb, styl)
			}
			newline = true
			w = 0
			comb = nil
			line++
			y++
			continue
		}
		if t.widths[i] == 0 {
			comb = append(comb, l)
			continue
		}
		if w != 0 {
			v.SetContent(x, y, r, comb, styl)
			x += w
		}
		r = l
		w = t.widths[i]
		styl = t.styles[i]
		comb = nil
	}
	if w != 0 {
		v.SetContent(x, y, r, comb, styl)
	}
}

// Size returns the width and height in character cells of the Text.
func (t *Text) Size() (int, int) {
	if len(t.text) != 0 {
		return t.width, t.height
	}
	return 0, 0
}

// SetAlignment sets the alignment.  Negative values
// indicate right justification, positive values are left,
// and zero indicates center aligned.
func (t *Text) SetAlignment(align Alignment) {
	if align != t.align {
		t.align = align
		t.PostEventWidgetContent(t)
	}
}

// Alignment returns the alignment of the Text.
func (t *Text) Alignment() Alignment {
	return t.align
}

// SetView sets the View object used for the text bar.
func (t *Text) SetView(view View) {
	t.view = view
}

// HandleEvent implements a tcell.EventHandler, but does nothing.
func (t *Text) HandleEvent(tcell.Event) bool {
	return false
}

// SetText sets the text used for the string.  Any previously set
// styles on individual rune indices are reset, and the default style
// for the widget is set.
func (t *Text) SetText(s string) {
	t.width = 0
	t.text = []rune(s)
	if len(t.widths) < len(t.text) {
		t.widths = make([]int, len(t.text))
	} else {
		t.widths = t.widths[0:len(t.text)]
	}
	if len(t.styles) < len(t.text) {
		t.styles = make([]tcell.Style, len(t.text))
	} else {
		t.styles = t.styles[0:len(t.text)]
	}
	t.lengths = []int{}
	length := 0
	for i, r := range t.text {
		t.widths[i] = runewidth.RuneWidth(r)
		t.styles[i] = t.style
		if r == '\n' {
			t.lengths = append(t.lengths, length)
			if length > t.width {
				t.width = length
			}
			length = 0
		} else if t.widths[i] == 0 && length == 0 {
			// If first character on line is combining, inject
			// a leading space.  (Shame on the caller!)
			t.widths = append(t.widths, 0)
			copy(t.widths[i+1:], t.widths[i:])
			t.widths[i] = 1

			t.text = append(t.text, ' ')
			copy(t.text[i+1:], t.text[i:])
			t.text[i] = ' '

			t.styles = append(t.styles, t.style)
			copy(t.styles[i+1:], t.styles[i:])
			t.styles[i] = t.style
			length++
		} else {
			length += t.widths[i]
		}
	}
	if length > 0 {
		t.lengths = append(t.lengths, length)
		if length > t.width {
			t.width = length
		}
	}
	t.height = len(t.lengths)
	t.PostEventWidgetContent(t)
}

// Text returns the text that was set.
func (t *Text) Text() string {
	return string(t.text)
}

// SetStyle sets the style used.  This applies to every cell in the
// in the text.
func (t *Text) SetStyle(style tcell.Style) {
	t.style = style
	for i := 0; i < len(t.text); i++ {
		if t.widths[i] != 0 {
			t.styles[i] = t.style
		}
	}
	t.PostEventWidgetContent(t)
}

// Style returns the previously set default style.  Note that
// individual characters may have different styles.
func (t *Text) Style() tcell.Style {
	return t.style
}

// SetStyleAt sets the style at the given rune index.  Note that for
// strings containing combining characters, it is not possible to
// change the style at the position of the combining character, but
// those positions *do* count for calculating the index.  A lot of
// complexity can be avoided by avoiding the use of combining characters.
func (t *Text) SetStyleAt(pos int, style tcell.Style) {
	if pos < 0 || pos >= len(t.text) || t.widths[pos] < 1 {
		return
	}
	t.styles[pos] = style
	t.PostEventWidgetContent(t)
}

// StyleAt gets the style at the given rune index.  If an invalid
// index is given, or the index is a combining character, then
// tcell.StyleDefault is returned.
func (t *Text) StyleAt(pos int) tcell.Style {
	if pos < 0 || pos >= len(t.text) || t.widths[pos] < 1 {
		return tcell.StyleDefault
	}
	return t.styles[pos]
}

// Resize is called when our View changes sizes.
func (t *Text) Resize() {
	t.PostEventWidgetResize(t)
}

// NewText creates an empty Text.
func NewText() *Text {
	return &Text{}
}
