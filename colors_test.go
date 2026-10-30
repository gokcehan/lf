package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestLinkTargetColorOverrides(t *testing.T) {
	link := &file{FileInfo: &fakeStat{name: "link"}, linkState: working}
	for _, test := range []struct {
		name, initial, override, config string
		want                            string
	}{
		{"file-target", "ln=35:fi=31", "", "ln target\n", "31"},
		{"env-color", "ln=target:fi=31", "ln=35", "", "35"},
		{"file-color", "ln=target:fi=31", "", "ln 35\n", "35"},
		{"last-file-entry", "fi=31", "", "ln target\nln 35\n", "35"},
	} {
		t.Run(test.name, func(t *testing.T) {
			sm := styleMap{styles: make(map[string]tcell.Style)}
			sm.parseGNU(test.initial)
			sm.parseGNU(test.override)
			if test.config != "" {
				path := filepath.Join(t.TempDir(), "colors")
				if err := os.WriteFile(path, []byte(test.config), 0o600); err != nil {
					t.Fatal(err)
				}
				sm.parseFile(path)
			}
			if got, want := sm.get(link), applySGR(test.want, tcell.StyleDefault); got != want {
				t.Errorf("link style = %v, want %v", got, want)
			}
		})
	}
}

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
