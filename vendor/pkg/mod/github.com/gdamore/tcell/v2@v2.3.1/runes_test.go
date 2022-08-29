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

func TestCanDisplayUTF8(t *testing.T) {
	s := mkTestScreen(t, "UTF-8")
	defer s.Fini()

	if s.CharacterSet() != "UTF-8" {
		t.Errorf("Bad charset: %v", s.CharacterSet())
	}
	if !s.CanDisplay('a', true) {
		t.Errorf("Should be able to display 'a'")
	}
	if !s.CanDisplay(RuneHLine, true) {
		t.Errorf("Should be able to display hline (with fallback)")
	}
	if !s.CanDisplay(RuneHLine, false) {
		t.Errorf("Should be able to display hline (no fallback)")
	}
	if !s.CanDisplay('⌀', false) {
		t.Errorf("Should be able to display null")
	}
}
func TestCanDisplayASCII(t *testing.T) {
	s := mkTestScreen(t, "US-ASCII")
	defer s.Fini()

	if s.CharacterSet() != "US-ASCII" {
		t.Errorf("Wrong character set: %v", s.CharacterSet())
	}
	if !s.CanDisplay('a', true) {
		t.Errorf("Should be able to display 'a'")
	}
	if !s.CanDisplay(RuneHLine, true) {
		t.Errorf("Should be able to display hline (with fallback)")
	}
	if s.CanDisplay(RunePi, false) {
		t.Errorf("Should not be able to display Pi (no fallback)")
	}
	if s.CanDisplay('⌀', false) {
		t.Errorf("Should not be able to display null")
	}
}

func TestRuneFallbacks(t *testing.T) {
	s := mkTestScreen(t, "US-ASCII")
	defer s.Fini()
	if s.CharacterSet() != "US-ASCII" {
		t.Errorf("Wrong character set: %v", s.CharacterSet())
	}

	// Test registering a fallback
	s.RegisterRuneFallback('⌀', "o")
	if s.CanDisplay('⌀', false) {
		t.Errorf("Should not be able to display null (no fallback)")
	}
	if !s.CanDisplay('⌀', true) {
		t.Errorf("Should be able to display null (with fallback)")
	}

	// Test unregistering custom fallback
	s.UnregisterRuneFallback('⌀')
	if s.CanDisplay('⌀', false) {
		t.Errorf("Should not be able to display null (no fallback)")
	}
	if s.CanDisplay('⌀', true) {
		t.Errorf("Should not be able to display null (with fallback)")
	}

	// Test unregistering builtin fallback
	if !s.CanDisplay(RuneHLine, true) {
		t.Errorf("Should be able to display hline")
	}
	s.UnregisterRuneFallback(RuneHLine)
	if s.CanDisplay(RuneHLine, true) {
		t.Errorf("Should not be able to display hline")
	}
}
