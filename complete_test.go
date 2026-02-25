package main

import (
	"reflect"
	"testing"
)

func TestGetOptWords(t *testing.T) {
	tests := []struct {
		opts any
		exp  []string
	}{
		{struct{ feature bool }{}, []string{"feature", "feature!", "nofeature"}},
		{struct{ feature int }{}, []string{"feature"}},
		{struct{ feature string }{}, []string{"feature"}},
		{struct{ feature []string }{}, []string{"feature"}},
	}

	for _, test := range tests {
		result := getOptWords(test.opts)
		if !reflect.DeepEqual(result, test.exp) {
			t.Errorf("at input '%#v' expected '%s' but got '%s'", test.opts, test.exp, result)
		}
	}
}

func TestGetLocalOptWords(t *testing.T) {
	tests := []struct {
		localOpts any
		exp       []string
	}{
		{struct{ feature map[string]bool }{}, []string{"feature", "feature!", "nofeature"}},
		{struct{ feature map[string]int }{}, []string{"feature"}},
		{struct{ feature map[string]string }{}, []string{"feature"}},
		{struct{ feature map[string][]string }{}, []string{"feature"}},
	}

	for _, test := range tests {
		result := getLocalOptWords(test.localOpts)
		if !reflect.DeepEqual(result, test.exp) {
			t.Errorf("at input '%#v' expected '%s' but got '%s'", test.localOpts, test.exp, result)
		}
	}
}

func TestGetLongest(t *testing.T) {
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
		if got := getLongest(test.s1, test.s2); got != test.exp {
			t.Errorf("at input '%s' and '%s' expected '%s' but got '%s'", test.s1, test.s2, test.exp, got)
		}
	}
}

func TestMatchWord(t *testing.T) {
	tests := []struct {
		s       string
		words   []string
		matches []compMatch
		longest string
	}{
		{"", nil, nil, ""},
		{"", []string{"foo", "bar", "baz"}, []compMatch{{"foo", "foo"}, {"bar", "bar"}, {"baz", "baz"}}, ""},
		{"f", []string{"foo", "bar", "baz"}, []compMatch{{"foo", "foo"}}, "foo "},
		{"b", []string{"foo", "bar", "baz"}, []compMatch{{"bar", "bar"}, {"baz", "baz"}}, "ba"},
		{"fo", []string{"foo", "bar", "baz"}, []compMatch{{"foo", "foo"}}, "foo "},
		{"ba", []string{"foo", "bar", "baz"}, []compMatch{{"bar", "bar"}, {"baz", "baz"}}, "ba"},
		{"fo", []string{"bar", "baz"}, nil, "fo"},
	}

	for _, test := range tests {
		matches, longest := matchWord(test.s, test.words)

		if !reflect.DeepEqual(matches, test.matches) {
			t.Errorf("at input '%s' with '%s' expected '%v' but got '%v'", test.s, test.words, test.matches, matches)
		}

		if longest != test.longest {
			t.Errorf("at input '%s' with '%s' expected '%s' but got '%s'", test.s, test.words, test.longest, longest)
		}
	}
}

func TestMatchList(t *testing.T) {
	tests := []struct {
		s       string
		words   []string
		matches []compMatch
		longest string
	}{
		{"", nil, nil, ""},
		{"", []string{"foo", "bar", "baz"}, []compMatch{{"foo", "foo"}, {"bar", "bar"}, {"baz", "baz"}}, ""},
		{"f", []string{"bar", "baz"}, nil, "f"},
		{"f", []string{"foo", "bar", "baz"}, []compMatch{{"foo", "foo"}}, "foo"},
		{"b", []string{"foo", "bar", "baz"}, []compMatch{{"bar", "bar"}, {"baz", "baz"}}, "ba"},
		{"ba", []string{"foo", "bar", "baz"}, []compMatch{{"bar", "bar"}, {"baz", "baz"}}, "ba"},
		{"foo", []string{"foo", "bar", "baz"}, []compMatch{{"foo", "foo"}}, "foo "},
		{"foo:", []string{"foo", "bar", "baz"}, []compMatch{{"bar", "foo:bar"}, {"baz", "foo:baz"}}, "foo:ba"},
		{"foo:f", []string{"foo", "bar", "baz"}, nil, "foo:f"},
		{"foo:b", []string{"foo", "bar", "baz"}, []compMatch{{"bar", "foo:bar"}, {"baz", "foo:baz"}}, "foo:ba"},
		{"foo:ba", []string{"foo", "bar", "baz"}, []compMatch{{"bar", "foo:bar"}, {"baz", "foo:baz"}}, "foo:ba"},
		{"bar:b", []string{"foo", "bar", "baz"}, []compMatch{{"baz", "bar:baz"}}, "bar:baz"},
		{"bar:f", []string{"foo", "bar", "baz"}, []compMatch{{"foo", "bar:foo"}}, "bar:foo"},
		{"bar:foo", []string{"foo", "bar", "baz"}, []compMatch{{"foo", "bar:foo"}}, "bar:foo "},
	}

	for _, test := range tests {
		matches, longest := matchList(test.s, test.words)

		if !reflect.DeepEqual(matches, test.matches) {
			t.Errorf("at input '%s' with '%s' expected '%v' but got '%v'", test.s, test.words, test.matches, matches)
		}

		if longest != test.longest {
			t.Errorf("at input '%s' with '%s' expected '%s' but got '%s'", test.s, test.words, test.longest, longest)
		}
	}
}
