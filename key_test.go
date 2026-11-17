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

func TestMouseWheelKey(t *testing.T) {
	tests := []struct {
		name    string
		buttons tcell.ButtonMask
		want    string
	}{
		{name: "up", buttons: tcell.WheelUp, want: "<m-up>"},
		{name: "down", buttons: tcell.WheelDown, want: "<m-down>"},
		{name: "left", buttons: tcell.WheelLeft, want: "<m-left>"},
		{name: "right", buttons: tcell.WheelRight, want: "<m-right>"},
		{name: "wheel with held button", buttons: tcell.Button2 | tcell.WheelUp, want: "<m-up>"},
		{name: "no wheel", buttons: tcell.Button2, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mouseWheelKey(tt.buttons); got != tt.want {
				t.Errorf("mouseWheelKey(%v) = %q, want %q", tt.buttons, got, tt.want)
			}
		})
	}
}

func TestReadNormalEventWheelWithHeldButton(t *testing.T) {
	previousKeys := gOpts.nkeys
	gOpts.nkeys = map[string]expr{"<m-up>": &callExpr{name: "up", count: 1}}
	t.Cleanup(func() { gOpts.nkeys = previousKeys })

	nav := &nav{
		dirPaths: []string{"."},
		dirCache: map[string]*dir{".": {visualAnchor: -1}},
	}
	want := gOpts.nkeys["<m-up>"]
	event := tcell.NewEventMouse(0, 0, tcell.Button2|tcell.WheelUp, tcell.ModNone)

	if got := (&ui{}).readNormalEvent(event, nav); got != want {
		t.Errorf("readNormalEvent() = %#v, want wheel-up binding %#v", got, want)
	}
}
