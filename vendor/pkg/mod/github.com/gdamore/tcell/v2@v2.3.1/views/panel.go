// Copyright 2015 The Tops'l Authors
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

// Panel is a modified Layout that includes a primary content pane,
// prefixed with an optional title, and an optional menubar, and then
// suffixed by an optional status.
//
// Only the content pane is resizable.  The panel will be formatted
// like this:
//
//  +----------
//  | title
//  | menu
//  | content....
//  | <padding>
//  | status
//  +----------
//
// Each of these components may be any valid widget; their names are
// only meant to be indicative of conventional use, not prescriptive.
type Panel struct {
	title   Widget
	menu    Widget
	content Widget
	status  Widget
	inited  bool
	BoxLayout
}

// Draw draws the Panel.
func (p *Panel) Draw() {
	p.BoxLayout.SetOrientation(Vertical)
	p.BoxLayout.Draw()
}

// SetTitle sets the Widget to display in the title area.
func (p *Panel) SetTitle(w Widget) {
	if p.title != nil {
		p.RemoveWidget(p.title)
	}
	p.InsertWidget(0, w, 0.0)
	p.title = w
}

// SetMenu sets the Widget to display in the menu area, which is
// just below the title.
func (p *Panel) SetMenu(w Widget) {
	index := 0
	if p.title != nil {
		index++
	}
	if p.menu != nil {
		p.RemoveWidget(p.menu)
	}
	p.InsertWidget(index, w, 0.0)
	p.menu = w
}

// SetContent sets the Widget to display in the content area.
func (p *Panel) SetContent(w Widget) {
	index := 0
	if p.title != nil {
		index++
	}
	if p.menu != nil {
		index++
	}
	if p.content != nil {
		p.RemoveWidget(p.content)
	}
	p.InsertWidget(index, w, 1.0)
	p.content = w
}

// SetStatus sets the Widget to display in the status area, which is at
// the bottom of the panel.
func (p *Panel) SetStatus(w Widget) {
	index := 0
	if p.title != nil {
		index++
	}
	if p.menu != nil {
		index++
	}
	if p.content != nil {
		index++
	}
	if p.status != nil {
		p.RemoveWidget(p.status)
	}
	p.InsertWidget(index, w, 0.0)
	p.status = w
}

// NewPanel creates a new Panel.  A zero valued panel can be created too.
func NewPanel() *Panel {
	return &Panel{}
}
