package main

import (
	"testing"
)

func TestPrintLength(t *testing.T) {
	gOpts.tabstop = 8

	tests := []struct {
		s   string
		exp int
	}{
		{"", 0},
		{"hello", 5},
		{"日本語", 6},
		{"\x1b[31mred\x1b[0m", 3},
		{"\x1b]8;;http://x\x1b\\link\x1b]8;;\x1b\\", 4},
		{"\x00", 1},
		{"\t", 8},
		{"ab\t", 8},
	}
	for _, test := range tests {
		if got := printLength(test.s); got != test.exp {
			t.Errorf("at input %q expected %d but got %d", test.s, test.exp, got)
		}
	}
}
