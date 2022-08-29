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
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// SimpleStyledText is a form of Text that offers highlighting of the text
// using simple in-line markup.  Its intention is to make it easier to mark
// up hot // keys for menubars, etc.
type SimpleStyledText struct {
	styles map[rune]tcell.Style
	markup []rune
	Text
}

// SetMarkup sets the text used for the string.  It applies markup as follows
// (modeled on tcsh style prompt markup):
//
// * %% - emit a single % in current style
// * %N - normal style
// * %A - alternate style
// * %S - start standout (reverse) style
// * %B - start bold style
// * %U - start underline style
//
// Other styles can be set using %<rune>, if styles are registered.
// Upper case characters and punctuation are reserved for use by the system.
// Lower case are available for use by the user.  (Users may define mappings
// for upper case letters to override system defined styles.)
//
// Note that for simplicity, combining styles is not supported.  By default
// the alternate style is the same as standout (reverse) mode.
//
// Arguably we could have used Markdown syntax instead, but properly doing all
// of Markdown is not trivial, and these escape sequences make it clearer that
// we are not even attempting to do that.
func (t *SimpleStyledText) SetMarkup(s string) {

	markup := []rune(s)
	styl := make([]tcell.Style, 0, len(markup))
	text := make([]rune, 0, len(markup))

	style := t.styles['N']

	esc := false
	for _, r := range markup {
		if esc {
			esc = false
			switch r {
			case '%':
				text = append(text, '%')
				styl = append(styl, style)
			default:
				style = t.styles[r]
			}
			continue
		}
		switch r {
		case '%':
			esc = true
			continue
		default:
			text = append(text, r)
			styl = append(styl, style)
		}
	}

	t.Text.SetText(string(text))
	for i, s := range styl {
		t.SetStyleAt(i, s)
	}
	t.markup = markup
}

// Registers a style for the given rune.  This style will be used for
// text marked with %<r>. See SetMarkup() for more detail.  Note that
// this must be done before using any of the styles with SetMarkup().
// Only letters may be used when registering styles, and be advised that
// the system may have predefined uses for upper case letters.
func (t *SimpleStyledText) RegisterStyle(r rune, style tcell.Style) {
	if r == 'N' {
		t.Text.SetStyle(style)
	}
	if unicode.IsLetter(r) {
		t.styles[r] = style
	}
}

// LookupStyle returns the style registered for the given rune.
// Returns tcell.StyleDefault if no style was previously registered
// for the rune.
func (t *SimpleStyledText) LookupStyle(r rune) tcell.Style {
	return t.styles[r]
}

// Markup returns the text that was set, including markup.
func (t *SimpleStyledText) Markup() string {
	return string(t.markup)
}

// NewSimpleStyledText creates an empty Text.
func NewSimpleStyledText() *SimpleStyledText {
	ss := &SimpleStyledText{}
	// Create map and establish default styles.
	ss.styles = make(map[rune]tcell.Style)
	ss.styles['N'] = tcell.StyleDefault
	ss.styles['S'] = tcell.StyleDefault.Reverse(true)
	ss.styles['U'] = tcell.StyleDefault.Underline(true)
	ss.styles['B'] = tcell.StyleDefault.Bold(true)
	return ss
}
