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
	"github.com/gdamore/tcell/v2"
)

// Spacer is a Widget that occupies no visible real-estate.  It is useful to
// add this to layouts when expansion space is required.  It expands as needed
// with blank space.
type Spacer struct {
	WidgetWatchers
}

// Draw is called to update the displayed content.
func (*Spacer) Draw() {}

// Size always returns 0, 0, since no size is ever *requird* to display nothing.
func (*Spacer) Size() (int, int) {
	return 0, 0
}

// SetView sets the View object used for the text bar.
func (*Spacer) SetView(View) {}

// HandleEvent implements a tcell.EventHandler, but does nothing.
func (*Spacer) HandleEvent(tcell.Event) bool {
	return false
}

// Resize is called when our View changes sizes.
func (s *Spacer) Resize() {
	s.PostEventWidgetResize(s)
}

// NewSpacer creates an empty Spacer.  It's probably easier just to declare it
// directly.
func NewSpacer() *Spacer {
	return &Spacer{}
}
