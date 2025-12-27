package main

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

var gKeyTests = []struct {
	ev *tcell.EventKey
	s  string
}{
	{tcell.NewEventKey(tcell.KeyRune, "<", tcell.ModNone), "<lt>"},
	{tcell.NewEventKey(tcell.KeyRune, ">", tcell.ModNone), "<gt>"},
	{tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone), "<space>"},
	{tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModNone), "a"},
	{tcell.NewEventKey(tcell.KeyCtrlA, "", tcell.ModNone), "<c-a>"},
	{tcell.NewEventKey(tcell.KeyRune, "A", tcell.ModNone), "A"},
	{tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModAlt), "<a-a>"},
	{tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone), "<left>"},
	{tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModCtrl), "<c-left>"},
	{tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModShift), "<s-left>"},
	{tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModAlt), "<a-left>"},
	{tcell.NewEventKey(tcell.KeyEsc, "", tcell.ModNone), "<esc>"},
	{tcell.NewEventKey(tcell.KeyF1, "", tcell.ModNone), "<f-1>"},
}

func TestReadKey(t *testing.T) {
	for _, test := range gKeyTests {
		if got := readKey(test.ev); got != test.s {
			t.Errorf("at input '%#v' expected '%s' but got '%s'", test.ev, test.s, got)
		}
	}
}

func TestParseKey(t *testing.T) {
	keyEqual := func(ev1, ev2 *tcell.EventKey) bool {
		return ev1.Key() == ev2.Key() && ev1.Modifiers() == ev2.Modifiers() && ev1.Str() == ev2.Str()
	}

	for _, test := range gKeyTests {
		if got := parseKey(test.s); !keyEqual(got, test.ev) {
			t.Errorf("at input '%s' expected '%#v' but got '%#v'", test.s, test.ev, got)
		}
	}
}
