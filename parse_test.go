package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, test := range gTests {
		p := newParser(strings.NewReader(test.inp))

		for _, expr := range test.exprs {
			if p.parse(); !reflect.DeepEqual(p.expr, expr) {
				t.Errorf("at input '%s' expected '%s' but parsed '%s'", test.inp, expr, p.expr)
			}
		}

		if p.parse(); p.expr != nil {
			t.Errorf("at input '%s' unexpected '%s'", test.inp, p.expr)
		}
	}
}
