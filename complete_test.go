package main

import (
	"reflect"
	"testing"
)

func TestMatchLongest(t *testing.T) {
	tests := []struct {
		s1  string
		s2  string
		exp string
	}{
		{"", "", ""},
		{"", "foo", ""},
		{"foo", "", ""},
		{"foo", "bar", ""},
		{"foo", "foobar", "foo"},
		{"foo", "barfoo", ""},
		{"foobar", "foobaz", "fooba"},
		{"год", "гол", "го"},
	}

	for _, test := range tests {
		if got := string(matchLongest([]rune(test.s1), []rune(test.s2))); got != test.exp {
			t.Errorf("at input '%s' and '%s' expected '%s' but got '%s'", test.s1, test.s2, test.exp, got)
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
		{"", nil, nil, ""},
		{"", []string{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}, ""},
		{"fo", []string{"foo", "bar", "baz"}, []string{"foo"}, "foo "},
		{"ba", []string{"foo", "bar", "baz"}, []string{"bar", "baz"}, "ba"},
		{"fo", []string{"bar", "baz"}, nil, "fo"},
	}

	for _, test := range tests {
		m, l := matchWord(test.s, test.words)

		if !reflect.DeepEqual(m, test.matches) {
			t.Errorf("at input '%s' with '%s' expected '%s' but got '%s'", test.s, test.words, test.matches, m)
		}

		if ls := string(l); ls != test.longest {
			t.Errorf("at input '%s' with '%s' expected '%s' but got '%s'", test.s, test.words, test.longest, ls)
		}
	}
}
