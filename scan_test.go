package main

import (
	"strings"
	"testing"
)

func TestScan(t *testing.T) {
	for _, test := range gEvalTests {
		s := newScanner(strings.NewReader(test.inp))

		for _, tok := range test.toks {
			if s.scan(); s.tok != tok {
				t.Errorf("at input '%s' expected '%s' but scanned '%s'", test.inp, tok, s.tok)
			}
		}

		if s.scan() {
			t.Errorf("at input '%s' unexpected '%s'", test.inp, s.tok)
		}
	}
}
