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

func TestApplyAnsiCodes(t *testing.T) {
	none := tcell.StyleDefault

	tests := []struct {
		s     string
		st    tcell.Style
		stExp tcell.Style
	}{
		{"", none, none},
		{"", none.Foreground(tcell.ColorMaroon).Background(tcell.ColorMaroon), none},
		{"", none.Bold(true), none},
		{"", none.Foreground(tcell.ColorMaroon).Bold(true), none},

		{"0", none, none},
		{"0", none.Foreground(tcell.ColorMaroon).Background(tcell.ColorMaroon), none},
		{"0", none.Bold(true), none},
		{"0", none.Foreground(tcell.ColorMaroon).Bold(true), none},

		{"1", none, none.Bold(true)},
		{"4", none, none.Underline(true)},
		{"7", none, none.Reverse(true)},

		{"1", none.Foreground(tcell.ColorMaroon), none.Foreground(tcell.ColorMaroon).Bold(true)},
		{"4", none.Foreground(tcell.ColorMaroon), none.Foreground(tcell.ColorMaroon).Underline(true)},
		{"7", none.Foreground(tcell.ColorMaroon), none.Foreground(tcell.ColorMaroon).Reverse(true)},

		{"4", none.Bold(true), none.Bold(true).Underline(true)},
		{"4", none.Foreground(tcell.ColorMaroon).Bold(true), none.Foreground(tcell.ColorMaroon).Bold(true).Underline(true)},

		{"4:0", none, none},
		{"4:0", none.Underline(true), none},
		{"4:1", none, none.Underline(true)},
		{"4:2", none, none.Underline(tcell.UnderlineStyleDouble)},
		{"4:3", none, none.Underline(tcell.UnderlineStyleCurly)},
		{"4:4", none, none.Underline(tcell.UnderlineStyleDotted)},
		{"4:5", none, none.Underline(tcell.UnderlineStyleDashed)},

		{"22", none.Italic(true).Bold(true).Dim(true), none.Italic(true)},
		{"23", none.Bold(true).Italic(true), none.Bold(true)},
		{"24", none.Bold(true).Underline(true), none.Bold(true)},
		{"25", none.Bold(true).Blink(true), none.Bold(true)},
		{"27", none.Bold(true).Reverse(true), none.Bold(true)},
		{"29", none.Bold(true).StrikeThrough(true), none.Bold(true)},

		{"31", none, none.Foreground(tcell.ColorMaroon)},
		{"31", none.Foreground(tcell.ColorGreen), none.Foreground(tcell.ColorMaroon)},
		{"31", none.Foreground(tcell.ColorGreen).Bold(true), none.Foreground(tcell.ColorMaroon).Bold(true)},

		{"41", none, none.Background(tcell.ColorMaroon)},
		{"41", none.Background(tcell.ColorGreen), none.Background(tcell.ColorMaroon)},

		{"1;31", none, none.Foreground(tcell.ColorMaroon).Bold(true)},
		{"1;31", none.Foreground(tcell.ColorGreen), none.Foreground(tcell.ColorMaroon).Bold(true)},

		{"01;31", none, none.Foreground(tcell.ColorMaroon).Bold(true)},
		{"01;31", none.Foreground(tcell.ColorGreen), none.Foreground(tcell.ColorMaroon).Bold(true)},

		{"38;5;0", none, none.Foreground(tcell.ColorBlack)},
		{"38;5;1", none, none.Foreground(tcell.ColorMaroon)},
		{"38;5;8", none, none.Foreground(tcell.ColorGray)},
		{"38;5;16", none, none.Foreground(tcell.Color16)},
		{"38;5;232", none, none.Foreground(tcell.Color232)},

		{"38;5;1", none.Foreground(tcell.ColorGreen), none.Foreground(tcell.ColorMaroon)},
		{"38;5;1", none.Foreground(tcell.ColorGreen).Bold(true), none.Foreground(tcell.ColorMaroon).Bold(true)},

		{"48;5;0", none, none.Background(tcell.ColorBlack)},
		{"48;5;1", none, none.Background(tcell.ColorMaroon)},
		{"48;5;8", none, none.Background(tcell.ColorGray)},
		{"48;5;16", none, none.Background(tcell.Color16)},
		{"48;5;232", none, none.Background(tcell.Color232)},

		{"48;5;1", none.Background(tcell.ColorGreen), none.Background(tcell.ColorMaroon)},

		{"1;38;5;1", none, none.Foreground(tcell.ColorMaroon).Bold(true)},
		{"1;38;5;1", none.Foreground(tcell.ColorGreen), none.Foreground(tcell.ColorMaroon).Bold(true)},

		{"38;2;5;102;8", none, none.Foreground(tcell.NewRGBColor(5, 102, 8))},
		{"48;2;0;48;143", none, none.Background(tcell.NewRGBColor(0, 48, 143))},

		// Fixes color construction issue: https://github.com/gokcehan/lf/pull/439#issuecomment-674409446
		{"38;5;34;1", none, none.Foreground(tcell.Color34).Bold(true)},
	}

	for _, test := range tests {
		if stGot := applyAnsiCodes(test.s, test.st); stGot != test.stExp {
			t.Errorf("at input '%s' with '%v' expected '%v' but got '%v'",
				test.s, test.st, test.stExp, stGot)
		}
	}
}
