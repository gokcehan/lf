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

// SimpleStyledTextBar is a Widget that provides a single line of text, but
// with distinct left, center, and right areas.  Each of the areas can be
// styled differently, and can contain internal style markup.
// They align to the left, center, and right respectively.
// This is basically a convenience type on top SimpleStyledText and BoxLayout.
type SimpleStyledTextBar struct {
	left   *SimpleStyledText
	right  *SimpleStyledText
	center *SimpleStyledText
	once   sync.Once

	BoxLayout
}

// SetRight sets the right text for the textbar.
// It is always right-aligned.
func (s *SimpleStyledTextBar) SetRight(m string) {
	s.initialize()
	s.right.SetMarkup(m)
}

// SetLeft sets the left text for the textbar.
// It is always left-aligned.
func (s *SimpleStyledTextBar) SetLeft(m string) {
	s.initialize()
	s.left.SetMarkup(m)
}

// SetCenter sets the center text for the textbar.
// It is always centered.
func (s *SimpleStyledTextBar) SetCenter(m string) {
	s.initialize()
	s.center.SetMarkup(m)
}

func (s *SimpleStyledTextBar) RegisterRightStyle(r rune, style tcell.Style) {
	s.initialize()
	s.right.RegisterStyle(r, style)
}

func (s *SimpleStyledTextBar) RegisterLeftStyle(r rune, style tcell.Style) {
	s.initialize()
	s.left.RegisterStyle(r, style)
}

func (s *SimpleStyledTextBar) RegisterCenterStyle(r rune, style tcell.Style) {
	s.initialize()
	s.center.RegisterStyle(r, style)
}

func (s *SimpleStyledTextBar) Size() (int, int) {
	s.initialize()
	w, h := s.BoxLayout.Size()
	if h < 1 {
		h = 1
	}
	if w < 1 {
		w = 1
	}
	return w, h
}

func (s *SimpleStyledTextBar) initialize() {
	s.once.Do(func() {
		s.center = NewSimpleStyledText()
		s.left = NewSimpleStyledText()
		s.right = NewSimpleStyledText()
		s.center.SetAlignment(VAlignTop | HAlignCenter)
		s.left.SetAlignment(VAlignTop | HAlignLeft)
		s.right.SetAlignment(VAlignTop | HAlignRight)
		s.BoxLayout.SetOrientation(Horizontal)
		s.BoxLayout.AddWidget(s.left, 0.0)
		s.BoxLayout.AddWidget(NewSpacer(), 1.0)
		s.BoxLayout.AddWidget(s.center, 0.0)
		s.BoxLayout.AddWidget(NewSpacer(), 1.0)
		s.BoxLayout.AddWidget(s.right, 0.0)
	})
}

// Init prepares the widget for use, ensuring resources needed are
// allocated, etc.
func (s *SimpleStyledTextBar) Init() {
	s.initialize()
}

// NewSimpleStyledTextBar creates an empty, initialized TextBar.
func NewSimpleStyledTextBar() *SimpleStyledTextBar {
	s := &SimpleStyledTextBar{}
	s.initialize()
	return s
}
