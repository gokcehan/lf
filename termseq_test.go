package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestStripTermSequence(t *testing.T) {
	tests := []struct {
		s   string
		exp string
	}{
		{"", ""},                      // empty
		{"foo bar", "foo bar"},        // plain text
		{"\033[31mRed\033[0m", "Red"}, // octal syntax
		{"\x1b[31mRed\x1b[0m", "Red"}, // hexadecimal syntax
		{"foo\x1b[31mRed", "fooRed"},  // no reset parameter
		{
			"foo\x1b[1;31;102mBoldRedGreen\x1b[0mbar",
			"fooBoldRedGreenbar",
		}, // multiple attributes
		{
			"misc.go:func \x1b[01;31m\x1b[KstripTermSequence\x1b[m\x1b[K(s string) string {",
			"misc.go:func stripTermSequence(s string) string {",
		}, // `grep` output containing `erase in line` sequence

		// OSC 8 hyperlinks
		{
			"\x1b]8;;https://example.com\x1b\\example.com\x1b]8;;\x1b\\",
			"example.com",
		}, // open/close with ST (ESC\)
		{
			"\x1b]8;;https://example.com\x07example.com\x1b]8;;\x07",
			"example.com",
		}, // open/close with BEL
		{
			"\x1b]8;id=42;https://example.com\x1b\\label\x1b]8;;\x1b\\",
			"label",
		}, // params present
		{
			"\x1b]8;;https://example.com\x1b\\example.com",
			"example.com",
		}, // open without close
	}

	for _, test := range tests {
		if got := stripTermSequence(test.s); got != test.exp {
			t.Errorf("at input %q expected %q but got %q", test.s, test.exp, got)
		}
		// we rely on both functions extracting the same runes
		// to avoid misalignment
		if printLength(test.s) != len(stripTermSequence(test.s)) {
			t.Errorf("at input %q expected '%d' but got '%d'", test.s, printLength(test.s), len(stripTermSequence(test.s)))
		}
	}
}

func TestReadTermSequence(t *testing.T) {
	tests := []struct {
		s, exp string
	}{
		{"", ""},        // empty
		{"foo bar", ""}, // plain text
		{"\x1b", ""},    // lone ESC
		{"\x1bX", ""},   // unknown ESC sequence

		{"\x1b[31m", "\x1b[31m"},     // CSI SGR
		{"\x1b[K", "\x1b[K"},         // CSI EL
		{"\x1b[1;31m", "\x1b[1;31m"}, // CSI SGR (multiple params)
		{"\x1b[31", ""},              // CSI incomplete (no terminator)
		{"foo\x1b[31m", ""},          // doesn't start with ESC

		{"\x1b]8;;https://example.com\x1b\\", "\x1b]8;;https://example.com\x1b\\"}, // OSC 8 (ST terminator)
		{"\x1b]8;;https://example.com\x07", "\x1b]8;;https://example.com\x07"},     // OSC 8 (BEL terminator)
		{"\x1b]0;title\x07", ""}, // non-OSC8 OSC (ignored)
	}

	for _, tc := range tests {
		if got := readTermSequence(tc.s); got != tc.exp {
			t.Errorf("input %q: got %q, want %q", tc.s, got, tc.exp)
		}
	}
}

func TestOptionToFmtstr(t *testing.T) {
	tests := []struct {
		s   string
		exp string
	}{
		{"\033[1m", "\033[1m%s\033[0m"},
		{"\033[1;7;31;42m", "\033[1;7;31;42m%s\033[0m"},
	}

	for _, test := range tests {
		if got := optionToFmtstr(test.s); got != test.exp {
			t.Errorf("at input %q expected %q but got %q", test.s, test.exp, got)
		}
	}
}

func TestParseEscapeSequence(t *testing.T) {
	tests := []struct {
		s   string
		exp tcell.Style
	}{
		{"\033[1m", tcell.StyleDefault.Bold(true)},
		{"\033[1;7;31;42m", tcell.StyleDefault.Bold(true).Reverse(true).Foreground(tcell.ColorMaroon).Background(tcell.ColorGreen)},
	}

	for _, test := range tests {
		if got := parseEscapeSequence(test.s); got != test.exp {
			t.Errorf("at input %q expected '%v' but got '%v'", test.s, test.exp, got)
		}
	}
}

func TestApplyTermSequence(t *testing.T) {
	tests := []struct {
		s   string
		exp tcell.Style
	}{
		{"", tcell.StyleDefault},
		{"\x1b[1m", tcell.StyleDefault.Bold(true)},
		{"\x1b[1;7;31;42m", tcell.StyleDefault.Bold(true).Reverse(true).Foreground(tcell.ColorMaroon).Background(tcell.ColorGreen)},
		{
			"\x1b]8;;https://example.com\x1b\\",
			tcell.StyleDefault.UrlId("lf_hyperlink_100680ad546ce6a5").Url("https://example.com"),
		}, // OSC 8 terminated with ST (ESC\), no `id` provided
		{
			"\x1b]8;;https://example.com\x07",
			tcell.StyleDefault.UrlId("lf_hyperlink_100680ad546ce6a5").Url("https://example.com"),
		}, // OSC 8 terminated with BEL, no `id` provided
		{
			"\x1b]8;id=42;https://example.com\x1b\\",
			tcell.StyleDefault.UrlId("42").Url("https://example.com"),
		}, // OSC 8, `id` provided
	}

	for _, test := range tests {
		if got := applyTermSequence(test.s, tcell.StyleDefault); got != test.exp {
			t.Errorf("at input %q expected '%v' but got '%v'", test.s, test.exp, got)
		}
	}
}

func TestApplySGR(t *testing.T) {
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
		if stGot := applySGR(test.s, test.st); stGot != test.stExp {
			t.Errorf("at input '%s' with '%v' expected '%v' but got '%v'",
				test.s, test.st, test.stExp, stGot)
		}
	}
}
