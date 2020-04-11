package main

import (
	"testing"

	"github.com/doronbehar/termbox-go"
)

func TestApplyAnsiCodes(t *testing.T) {
	none := termbox.ColorDefault

	tests := []struct {
		s     string
		fg    termbox.Attribute
		bg    termbox.Attribute
		fgExp termbox.Attribute
		bgExp termbox.Attribute
	}{
		{"", none, none, none, none},
		{"", termbox.ColorRed, termbox.ColorRed, none, none},
		{"", termbox.AttrBold, none, none, none},
		{"", termbox.ColorRed | termbox.AttrBold, none, none, none},

		{"0", none, none, none, none},
		{"0", termbox.ColorRed, termbox.ColorRed, none, none},
		{"0", termbox.AttrBold, none, none, none},
		{"0", termbox.ColorRed | termbox.AttrBold, none, none, none},

		{"1", none, none, termbox.AttrBold, none},
		{"4", none, none, termbox.AttrUnderline, none},
		{"7", none, none, termbox.AttrReverse, none},

		{"1", termbox.ColorRed, none, termbox.ColorRed | termbox.AttrBold, none},
		{"4", termbox.ColorRed, none, termbox.ColorRed | termbox.AttrUnderline, none},
		{"7", termbox.ColorRed, none, termbox.ColorRed | termbox.AttrReverse, none},

		{"4", termbox.AttrBold, none, termbox.AttrBold | termbox.AttrUnderline, none},
		{"4", termbox.ColorRed | termbox.AttrBold, none, termbox.ColorRed | termbox.AttrBold | termbox.AttrUnderline, none},

		{"31", none, none, termbox.ColorRed, none},
		{"31", termbox.ColorGreen, none, termbox.ColorRed, none},
		{"31", termbox.ColorGreen | termbox.AttrBold, none, termbox.ColorRed | termbox.AttrBold, none},

		{"41", none, none, none, termbox.ColorRed},
		{"41", none, termbox.ColorGreen, none, termbox.ColorRed},

		{"1;31", none, none, termbox.ColorRed | termbox.AttrBold, none},
		{"1;31", termbox.ColorGreen, none, termbox.ColorRed | termbox.AttrBold, none},

		{"38;5;0", none, none, termbox.ColorBlack, none},
		{"38;5;1", none, none, termbox.ColorRed, none},
		{"38;5;8", none, none, termbox.Attribute(9), none},
		{"38;5;16", none, none, termbox.Attribute(17), none},
		{"38;5;232", none, none, termbox.Attribute(233), none},

		{"38;5;1", termbox.ColorGreen, none, termbox.ColorRed, none},
		{"38;5;1", termbox.ColorGreen | termbox.AttrBold, none, termbox.ColorRed | termbox.AttrBold, none},

		{"48;5;0", none, none, none, termbox.ColorBlack},
		{"48;5;1", none, none, none, termbox.ColorRed},
		{"48;5;8", none, none, none, termbox.Attribute(9)},
		{"48;5;16", none, none, none, termbox.Attribute(17)},
		{"48;5;232", none, none, none, termbox.Attribute(233)},

		{"48;5;1", none, termbox.ColorGreen, none, termbox.ColorRed},

		{"1;38;5;1", none, none, termbox.ColorRed | termbox.AttrBold, none},
		{"1;38;5;1", termbox.ColorGreen, none, termbox.ColorRed | termbox.AttrBold, none},
	}

	for _, test := range tests {
		if fgGot, bgGot := applyAnsiCodes(test.s, test.fg, test.bg); fgGot != test.fgExp || bgGot != test.bgExp {
			t.Errorf("at input '%s' with '%d' and '%d' expected '%d' and '%d' but got '%d' and '%d'",
				test.s, test.fg, test.bg, test.fgExp, test.bgExp, fgGot, bgGot)
		}
	}
}
