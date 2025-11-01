package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestParseColor(t *testing.T) {
	tests := []struct {
		toks    []string
		color   tcell.Color
		offset  int
		success bool
	}{
		{[]string{}, tcell.ColorDefault, 0, false},
		{[]string{"foo"}, tcell.ColorDefault, 0, false},
		{[]string{"5"}, tcell.ColorDefault, 0, false},
		{[]string{"5", "foo"}, tcell.ColorDefault, 0, false},
		{[]string{"5", "42"}, tcell.PaletteColor(42), 2, true},
		{[]string{"2"}, tcell.ColorDefault, 0, false},
		{[]string{"2", "foo"}, tcell.ColorDefault, 0, false},
		{[]string{"2", "42", "foo"}, tcell.ColorDefault, 0, false},
		{[]string{"2", "42", "43", "foo"}, tcell.ColorDefault, 0, false},
		{[]string{"2", "42", "43", "44"}, tcell.NewRGBColor(42, 43, 44), 4, true},
	}

	for _, test := range tests {
		color, offset, err := parseColor(test.toks)
		success := err == nil
		if color != test.color || offset != test.offset || success != test.success {
			t.Errorf("at input %v expected (%v, %v, %v) but got (%v, %v, %v)",
				test.toks, test.color, test.offset, test.success, color, offset, success)
		}
	}
}
