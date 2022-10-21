package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

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

		{"31", none, none.Foreground(tcell.ColorMaroon)},
		{"31", none.Foreground(tcell.ColorGreen), none.Foreground(tcell.ColorMaroon)},
		{"31", none.Foreground(tcell.ColorGreen).Bold(true), none.Foreground(tcell.ColorMaroon).Bold(true)},

		{"41", none, none.Background(tcell.ColorMaroon)},
		{"41", none.Background(tcell.ColorGreen), none.Background(tcell.ColorMaroon)},

		{"1;31", none, none.Foreground(tcell.ColorMaroon).Bold(true)},
		{"1;31", none.Foreground(tcell.ColorGreen), none.Foreground(tcell.ColorMaroon).Bold(true)},

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
