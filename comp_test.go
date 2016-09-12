package main

import (
	"reflect"
	"testing"
)

func TestMatchLongest(t *testing.T) {
	tests := []struct {
		fst string
		snd string
		res string
	}{
		{"", "", ""},
		{"", "foo", ""},
		{"foo", "", ""},
		{"foo", "bar", ""},
		{"foo", "foobar", "foo"},
		{"foo", "barfoo", ""},
		{"foobar", "foobaz", "fooba"},
	}

	for _, test := range tests {
		if l := matchLongest(test.fst, test.snd); l != test.res {
			t.Errorf("at input '%s' and '%s' expected '%s' but got '%s'", test.fst, test.snd, test.res, l)
		}
	}
}

func TestMatchWord(t *testing.T) {
	tests := []struct {
		s       string
		words   []string
		matches []string
		longest string
	}{
		{"fo", []string{"foo", "bar", "baz"}, []string{"foo"}, "foo "},
		{"ba", []string{"foo", "bar", "baz"}, []string{"bar", "baz"}, "ba"},
		{"fo", []string{"bar", "baz"}, nil, "fo"},
	}

	for _, test := range tests {
		m, l := matchWord(test.s, test.words)

		if !reflect.DeepEqual(m, test.matches) {
			t.Errorf("at input '%s' with '%s' expected '%s' but got '%s'", test.s, test.words, test.matches, m)
		}

		if l != test.longest {
			t.Errorf("at input '%s' with '%s' expected '%s' but got '%s'", test.s, test.words, test.longest, l)
		}
	}
}
