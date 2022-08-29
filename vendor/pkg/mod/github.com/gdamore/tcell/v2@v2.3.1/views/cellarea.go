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
	"sync"

	"github.com/gdamore/tcell/v2"
)

// CellModel models the content of a CellView.  The dimensions used within
// a CellModel are always logical, and have origin 0, 0.
type CellModel interface {
	GetCell(x, y int) (rune, tcell.Style, []rune, int)
	GetBounds() (int, int)
	SetCursor(int, int)
	GetCursor() (int, int, bool, bool)
	MoveCursor(offx, offy int)
}

// CellView is a flexible view of a CellModel, offering both cursor
// management and a panning.
type CellView struct {
	port     *ViewPort
	view     View
	content  Widget
	contentV *ViewPort
	style    tcell.Style
	lines    []string
	model    CellModel
	once     sync.Once

	WidgetWatchers
}

// Draw draws the content.
func (a *CellView) Draw() {

	port := a.port
	model := a.model
	port.Fill(' ', a.style)

	if a.view == nil {
		return
	}
	if model == nil {
		return
	}
	vw, vh := a.view.Size()
	for y := 0; y < vh; y++ {
		for x := 0; x < vw; x++ {
			a.view.SetContent(x, y, ' ', nil, a.style)
		}
	}

	ex, ey := model.GetBounds()
	vx, vy := port.Size()
	if ex < vx {
		ex = vx
	}
	if ey < vy {
		ey = vy
	}

	cx, cy, en, sh := a.model.GetCursor()
	for y := 0; y < ey; y++ {
		for x := 0; x < ex; x++ {
			ch, style, comb, wid := model.GetCell(x, y)
			if ch == 0 {
				ch = ' '
				style = a.style
			}
			if en && x == cx && y == cy && sh {
				style = style.Reverse(true)
			}
			port.SetContent(x, y, ch, comb, style)
			x += wid - 1
		}
	}
}

func (a *CellView) keyUp() {
	if _, _, en, _ := a.model.GetCursor(); !en {
		a.port.ScrollUp(1)
		return
	}
	a.model.MoveCursor(0, -1)
	a.MakeCursorVisible()
}

func (a *CellView) keyDown() {
	if _, _, en, _ := a.model.GetCursor(); !en {
		a.port.ScrollDown(1)
		return
	}
	a.model.MoveCursor(0, 1)
	a.MakeCursorVisible()
}

func (a *CellView) keyLeft() {
	if _, _, en, _ := a.model.GetCursor(); !en {
		a.port.ScrollLeft(1)
		return
	}
	a.model.MoveCursor(-1, 0)
	a.MakeCursorVisible()
}

func (a *CellView) keyRight() {
	if _, _, en, _ := a.model.GetCursor(); !en {
		a.port.ScrollRight(1)
		return
	}
	a.model.MoveCursor(+1, 0)
	a.MakeCursorVisible()
}

func (a *CellView) keyPgUp() {
	_, vy := a.port.Size()
	if _, _, en, _ := a.model.GetCursor(); !en {
		a.port.ScrollUp(vy)
		return
	}
	a.model.MoveCursor(0, -vy)
	a.MakeCursorVisible()
}

func (a *CellView) keyPgDn() {
	_, vy := a.port.Size()
	if _, _, en, _ := a.model.GetCursor(); !en {
		a.port.ScrollDown(vy)
		return
	}
	a.model.MoveCursor(0, +vy)
	a.MakeCursorVisible()
}

func (a *CellView) keyHome() {
	vx, vy := a.model.GetBounds()
	if _, _, en, _ := a.model.GetCursor(); !en {
		a.port.ScrollUp(vy)
		a.port.ScrollLeft(vx)
		return
	}
	a.model.SetCursor(0, 0)
	a.MakeCursorVisible()
}

func (a *CellView) keyEnd() {
	vx, vy := a.model.GetBounds()
	if _, _, en, _ := a.model.GetCursor(); !en {
		a.port.ScrollDown(vy)
		a.port.ScrollRight(vx)
		return
	}
	a.model.SetCursor(vx, vy)
	a.MakeCursorVisible()
}

// MakeCursorVisible ensures that the cursor is visible, panning the ViewPort
// as necessary, if the cursor is enabled.
func (a *CellView) MakeCursorVisible() {
	if a.model == nil {
		return
	}
	x, y, enabled, _ := a.model.GetCursor()
	if enabled {
		a.MakeVisible(x, y)
	}
}

// HandleEvent handles events.  In particular, it handles certain key events
// to move the cursor or pan the view.
func (a *CellView) HandleEvent(e tcell.Event) bool {
	if a.model == nil {
		return false
	}
	switch e := e.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyUp, tcell.KeyCtrlP:
			a.keyUp()
			return true
		case tcell.KeyDown, tcell.KeyCtrlN:
			a.keyDown()
			return true
		case tcell.KeyRight, tcell.KeyCtrlF:
			a.keyRight()
			return true
		case tcell.KeyLeft, tcell.KeyCtrlB:
			a.keyLeft()
			return true
		case tcell.KeyPgDn:
			a.keyPgDn()
			return true
		case tcell.KeyPgUp:
			a.keyPgUp()
			return true
		case tcell.KeyEnd:
			a.keyEnd()
			return true
		case tcell.KeyHome:
			a.keyHome()
			return true
		}
	}
	return false
}

// Size returns the content size, based on the model.
func (a *CellView) Size() (int, int) {
	// We always return a minimum of two rows, and two columns.
	w, h := a.model.GetBounds()
	// Clip to a 2x2 minimum square; we can scroll within that.
	if w > 2 {
		w = 2
	}
	if h > 2 {
		h = 2
	}
	return w, h
}

// GetModel gets the model for this CellView
func (a *CellView) GetModel() CellModel {
	return a.model
}

// SetModel sets the model for this CellView.
func (a *CellView) SetModel(model CellModel) {
	w, h := model.GetBounds()
	a.model = model
	a.port.SetContentSize(w, h, true)
	a.port.ValidateView()
	a.PostEventWidgetContent(a)
}

// SetView sets the View context.
func (a *CellView) SetView(view View) {
	port := a.port
	port.SetView(view)
	a.view = view
	if view == nil {
		return
	}
	width, height := view.Size()
	a.port.Resize(0, 0, width, height)
	if a.model != nil {
		w, h := a.model.GetBounds()
		a.port.SetContentSize(w, h, true)
	}
	a.Resize()
}

// Resize is called when the View is resized.  It will ensure that the
// cursor is visible, if present.
func (a *CellView) Resize() {
	// We might want to reflow text
	width, height := a.view.Size()
	a.port.Resize(0, 0, width, height)
	a.port.ValidateView()
	a.MakeCursorVisible()
}

// SetCursor sets the the cursor position.
func (a *CellView) SetCursor(x, y int) {
	a.model.SetCursor(x, y)
}

// SetCursorX sets the the cursor column.
func (a *CellView) SetCursorX(x int) {
	_, y, _, _ := a.model.GetCursor()
	a.SetCursor(x, y)
}

// SetCursorY sets the the cursor row.
func (a *CellView) SetCursorY(y int) {
	x, _, _, _ := a.model.GetCursor()
	a.SetCursor(x, y)
}

// MakeVisible makes the given coordinates visible, if they are not already.
// It does this by moving the ViewPort for the CellView.
func (a *CellView) MakeVisible(x, y int) {
	a.port.MakeVisible(x, y)
}

// SetStyle sets the the default fill style.
func (a *CellView) SetStyle(s tcell.Style) {
	a.style = s
}

// Init initializes a new CellView for use.
func (a *CellView) Init() {
	a.once.Do(func() {
		a.port = NewViewPort(nil, 0, 0, 0, 0)
		a.style = tcell.StyleDefault
	})
}

// NewCellView creates a CellView.
func NewCellView() *CellView {
	cv := &CellView{}
	cv.Init()
	return cv
}
