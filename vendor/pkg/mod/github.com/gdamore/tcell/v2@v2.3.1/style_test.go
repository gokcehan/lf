// Copyright 2018 The TCell Authors
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

package tcell

import (
	"testing"
)

func TestStyle(t *testing.T) {
	s := mkTestScreen(t, "")
	defer s.Fini()

	style := StyleDefault
	fg, bg, attr := style.Decompose()

	if fg != ColorDefault || bg != ColorDefault || attr != AttrNone {
		t.Errorf("Bad default style (%v, %v, %v)", fg, bg, attr)
	}

	s2 := style.
		Background(ColorRed).
		Foreground(ColorBlue).
		Blink(true)

	fg, bg, attr = s2.Decompose()
	if fg != ColorBlue || bg != ColorRed || attr != AttrBlink {
		t.Errorf("Bad custom style (%v, %v, %v)", fg, bg, attr)
	}
}
