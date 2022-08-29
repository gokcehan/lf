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

// View represents a logical view on an area.  It will have some underlying
// physical area as well, generally.  Views are operated on by Widgets.
type View interface {
	// SetContent is used to update the content of the View at the given
	// location.  This will generally be called by the Draw() method of
	// a Widget.
	SetContent(x int, y int, ch rune, comb []rune, style tcell.Style)

	// Size represents the visible size.  The actual content may be
	// larger or smaller.
	Size() (int, int)

	// Resize tells the View that its visible dimensions have changed.
	// It also tells it that it has a new offset relative to any parent
	// view.
	Resize(x, y, width, height int)

	// Fill fills the displayed content with the given rune and style.
	Fill(rune, tcell.Style)

	// Clear clears the content.  Often just Fill(' ', tcell.StyleDefault)
	Clear()
}

// ViewPort is an implementation of a View, that provides a smaller logical
// view of larger content area.  For example, a scrollable window of text,
// the visible window would be the ViewPort, on the underlying content.
// ViewPorts have a two dimensional size, and a two dimensional offset.
//
// In some ways, as the underlying content is not kept persistently by the
// view port, it can be thought perhaps a little more precisely as a clipping
// region.
type ViewPort struct {
	physx  int  // Anchor to the real world, usually 0
	physy  int  // Again, anchor to the real world, usually 3
	viewx  int  // Logical offset of the view
	viewy  int  // Logical offset of the view
	limx   int  // Content limits -- can't right scroll past this
	limy   int  // Content limits -- can't down scroll past this
	width  int  // View width
	height int  // View height
	locked bool // if true, don't autogrow
	v      View
}

// Clear clears the displayed content, filling it with spaces of default
// text attributes.
func (v *ViewPort) Clear() {
	v.Fill(' ', tcell.StyleDefault)
}

// Fill fills the displayed view port with the given character and style.
func (v *ViewPort) Fill(ch rune, style tcell.Style) {
	if v.v != nil {
		for y := 0; y < v.height; y++ {
			for x := 0; x < v.width; x++ {
				v.v.SetContent(x+v.physx, y+v.physy, ch, nil, style)
			}
		}
	}
}

// Size returns the visible size of the ViewPort in character cells.
func (v *ViewPort) Size() (int, int) {
	return v.width, v.height
}

// Reset resets the record of content, and also resets the offset back
// to the origin.  It doesn't alter the dimensions of the view port, nor
// the physical location relative to its parent.
func (v *ViewPort) Reset() {
	v.limx = 0
	v.limy = 0
	v.viewx = 0
	v.viewy = 0
}

// SetContent is used to place data at the given cell location.  Note that
// since the ViewPort doesn't retain this data, if the location is outside
// of the visible area, it is simply discarded.
//
// Generally, this is called during the Draw() phase by the object that
// represents the content.
func (v *ViewPort) SetContent(x, y int, ch rune, comb []rune, s tcell.Style) {
	if v.v == nil {
		return
	}
	if x > v.limx && !v.locked {
		v.limx = x
	}
	if y > v.limy && !v.locked {
		v.limy = y
	}
	if x < v.viewx || y < v.viewy {
		return
	}
	if x >= (v.viewx + v.width) {
		return
	}
	if y >= (v.viewy + v.height) {
		return
	}
	v.v.SetContent(x-v.viewx+v.physx, y-v.viewy+v.physy, ch, comb, s)
}

// MakeVisible moves the ViewPort the minimum necessary to make the given
// point visible.  This should be called before any content is changed with
// SetContent, since otherwise it may be possible to move the location onto
// a region whose contents have been discarded.
func (v *ViewPort) MakeVisible(x, y int) {
	if x < v.limx && x >= v.viewx+v.width {
		v.viewx = x - (v.width - 1)
	}
	if x >= 0 && x < v.viewx {
		v.viewx = x
	}
	if y < v.limy && y >= v.viewy+v.height {
		v.viewy = y - (v.height - 1)
	}
	if y >= 0 && y < v.viewy {
		v.viewy = y
	}
	v.ValidateView()
}

// ValidateViewY ensures that the Y offset of the view port is limited so that
// it cannot scroll away from the content.
func (v *ViewPort) ValidateViewY() {
	if v.viewy >= v.limy-v.height {
		v.viewy = (v.limy - v.height)
	}
	if v.viewy < 0 {
		v.viewy = 0
	}
}

// ValidateViewX ensures that the X offset of the view port is limited so that
// it cannot scroll away from the content.
func (v *ViewPort) ValidateViewX() {
	if v.viewx >= v.limx-v.width {
		v.viewx = (v.limx - v.width)
	}
	if v.viewx < 0 {
		v.viewx = 0
	}
}

// ValidateView does both ValidateViewX and ValidateViewY, ensuring both
// offsets are valid.
func (v *ViewPort) ValidateView() {
	v.ValidateViewX()
	v.ValidateViewY()
}

// Center centers the point, if possible, in the View.
func (v *ViewPort) Center(x, y int) {
	if x < 0 || y < 0 || x >= v.limx || y >= v.limy || v.v == nil {
		return
	}
	v.viewx = x - (v.width / 2)
	v.viewy = y - (v.height / 2)
	v.ValidateView()
}

// ScrollUp moves the view up, showing lower numbered rows of content.
func (v *ViewPort) ScrollUp(rows int) {
	v.viewy -= rows
	v.ValidateViewY()
}

// ScrollDown moves the view down, showingh higher numbered rows of content.
func (v *ViewPort) ScrollDown(rows int) {
	v.viewy += rows
	v.ValidateViewY()
}

// ScrollLeft moves the view to the left.
func (v *ViewPort) ScrollLeft(cols int) {
	v.viewx -= cols
	v.ValidateViewX()
}

// ScrollRight moves the view to the left.
func (v *ViewPort) ScrollRight(cols int) {
	v.viewx += cols
	v.ValidateViewX()
}

// SetSize is used to set the visible size of the view.  Enclosing views or
// layout managers can use this to inform the View of its correct visible size.
func (v *ViewPort) SetSize(width, height int) {
	v.height = height
	v.width = width
	v.ValidateView()
}

// GetVisible returns the upper left and lower right coordinates of the visible
// content.  That is, it will return x1, y1, x2, y2 where the upper left cell
// is position x1, y1, and the lower right is x2, y2.  These coordinates are
// in the space of the content, that is the content area uses coordinate 0,0
// as its first cell position.
func (v *ViewPort) GetVisible() (int, int, int, int) {
	return v.viewx, v.viewy, v.viewx + v.width - 1, v.viewy + v.height - 1
}

// GetPhysical returns the upper left and lower right coordinates of the visible
// content in the coordinate space of the parent.  This is may be the physical
// coordinates of the screen, if the screen is the view's parent.
func (v *ViewPort) GetPhysical() (int, int, int, int) {
	return v.physx, v.physy, v.physx + v.width - 1, v.physy + v.height - 1
}

// SetContentSize sets the size of the content area; this is used to limit
// scrolling and view moment.  If locked is true, then the content size will
// not automatically grow even if content is placed outside of this area
// with the SetContent() method.  If false, and content is drawn outside
// of the existing size, then the size will automatically grow to include
// the new content.
func (v *ViewPort) SetContentSize(width, height int, locked bool) {
	v.limx = width
	v.limy = height
	v.locked = locked
	v.ValidateView()
}

// GetContentSize returns the size of content as width, height in character
// cells.
func (v *ViewPort) GetContentSize() (int, int) {
	return v.limx, v.limy
}

// Resize is called by the enclosing view to change the size of the ViewPort,
// usually in response to a window resize event.  The x, y refer are the
// ViewPort's location relative to the parent View.  A negative value for either
// width or height will cause the ViewPort to expand to fill to the end of parent
// View in the relevant dimension.
func (v *ViewPort) Resize(x, y, width, height int) {
	if v.v == nil {
		return
	}
	px, py := v.v.Size()
	if x >= 0 && x < px {
		v.physx = x
	}
	if y >= 0 && y < py {
		v.physy = y
	}
	if width < 0 {
		width = px - x
	}
	if height < 0 {
		height = py - y
	}
	if width <= x+px {
		v.width = width
	}
	if height <= y+py {
		v.height = height
	}
}

// SetView is called during setup, to provide the parent View.
func (v *ViewPort) SetView(view View) {
	v.v = view
}

// NewViewPort returns a new ViewPort (and hence also a View).
// The x and y coordinates are an offset relative to the parent.
// The origin 0,0 represents the upper left.  The width and height
// indicate a width and height. If the value -1 is supplied, then the
// dimension is calculated from the parent.
func NewViewPort(view View, x, y, width, height int) *ViewPort {
	v := &ViewPort{v: view}
	// initial (and possibly poor) assumptions -- all visible
	// cells are addressible, but none beyond that.
	v.limx = width
	v.limy = height
	v.Resize(x, y, width, height)
	return v
}
