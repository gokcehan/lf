package main

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestIsRoot(t *testing.T) {
	sep := string(os.PathSeparator)
	if !isRoot(sep) {
		t.Errorf(`"%s" is root`, sep)
	}

	paths := []string{
		"",
		"~",
		"foo",
		"foo/bar",
		"foo/bar",
		"/home",
		"/home/user",
	}

	for _, p := range paths {
		if isRoot(p) {
			t.Errorf("'%s' is not root", p)
		}
	}
}

func TestRuneSliceWidth(t *testing.T) {
	tests := []struct {
		rs  []rune
		exp int
	}{
		{[]rune{'a', 'b'}, 2},
		{[]rune{'ı', 'ş'}, 2},
		{[]rune{'世', '界'}, 4},
		{[]rune{'世', 'a', '界', 'ı'}, 6},
	}

	for _, test := range tests {
		if got := runeSliceWidth(test.rs); got != test.exp {
			t.Errorf("at input '%v' expected '%d' but got '%d'", test.rs, test.exp, got)
		}
	}
}

func TestRuneSliceWidthRange(t *testing.T) {
	tests := []struct {
		rs  []rune
		beg int
		end int
		exp []rune
	}{
		{[]rune{}, 0, 0, []rune{}},
		{[]rune{'a', 'b', 'c', 'd'}, 1, 3, []rune{'b', 'c'}},
		{[]rune{'a', 'ı', 'b', 'ş'}, 1, 3, []rune{'ı', 'b'}},
		{[]rune{'世', '界', '世', '界'}, 2, 6, []rune{'界', '世'}},
		{[]rune{'世', '界', '世', '界'}, 3, 6, []rune{'世'}},
		{[]rune{'世', '界', '世', '界'}, 2, 5, []rune{'界'}},
		{[]rune{'世', '界', '世', '界'}, 3, 5, []rune{}},
		{[]rune{'世', '界', '世', '界'}, 4, 4, []rune{}},
		{[]rune{'世', '界', '世', '界'}, 5, 5, []rune{}},
		{[]rune{'世', '界', '世', '界'}, 4, 7, []rune{'世'}},
		{[]rune{'世', '界', '世', '界'}, 4, 8, []rune{'世', '界'}},
		{[]rune{'世', 'a', '界', 'ı'}, 2, 5, []rune{'a', '界'}},
		{[]rune{'世', 'a', '界', 'ı'}, 2, 4, []rune{'a'}},
		{[]rune{'世', 'a', '界', 'ı'}, 3, 5, []rune{'界'}},
		{[]rune{'世', 'a', '界', 'ı'}, 3, 4, []rune{}},
		{[]rune{'世', 'a', '界', 'ı'}, 3, 3, []rune{}},
		{[]rune{'世', 'a', '界', 'ı'}, 4, 4, []rune{}},
		{[]rune{'世', 'a', '界', 'ı'}, 4, 6, []rune{'ı'}},
		{[]rune{'世', 'a', '界', 'ı'}, 5, 6, []rune{'ı'}},
	}

	for _, test := range tests {
		if got := runeSliceWidthRange(test.rs, test.beg, test.end); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", test.rs, test.exp, got)
		}
	}
}

func TestRuneSliceWidthLastRange(t *testing.T) {
	tests := []struct {
		rs       []rune
		maxWidth int
		exp      []rune
	}{
		{[]rune{}, 0, []rune{}},
		{[]rune{}, 1, []rune{}},
		{[]rune{'a', 'ı', 'b', ' '}, 0, []rune{}},
		{[]rune{'a', 'ı', 'b', ' '}, 1, []rune{' '}},
		{[]rune{'a', 'ı', 'b', ' '}, 2, []rune{'b', ' '}},
		{[]rune{'a', 'ı', 'b', ' '}, 3, []rune{'ı', 'b', ' '}},
		{[]rune{'a', 'ı', 'b', ' '}, 4, []rune{'a', 'ı', 'b', ' '}},
		{[]rune{'a', 'ı', 'b', ' '}, 5, []rune{'a', 'ı', 'b', ' '}},
		{[]rune{'世', '界', '世', '界'}, 0, []rune{}},
		{[]rune{'世', '界', '世', '界'}, 1, []rune{}},
		{[]rune{'世', '界', '世', '界'}, 2, []rune{'界'}},
		{[]rune{'世', '界', '世', '界'}, 3, []rune{'界'}},
		{[]rune{'世', '界', '世', '界'}, 4, []rune{'世', '界'}},
		{[]rune{'世', '界', '世', '界'}, 5, []rune{'世', '界'}},
		{[]rune{'世', '界', '世', '界'}, 6, []rune{'界', '世', '界'}},
		{[]rune{'世', '界', '世', '界'}, 7, []rune{'界', '世', '界'}},
		{[]rune{'世', '界', '世', '界'}, 8, []rune{'世', '界', '世', '界'}},
		{[]rune{'世', '界', '世', '界'}, 9, []rune{'世', '界', '世', '界'}},
		{[]rune{'世', 'a', '界', 'ı'}, 0, []rune{}},
		{[]rune{'世', 'a', '界', 'ı'}, 1, []rune{'ı'}},
		{[]rune{'世', 'a', '界', 'ı'}, 2, []rune{'ı'}},
		{[]rune{'世', 'a', '界', 'ı'}, 3, []rune{'界', 'ı'}},
		{[]rune{'世', 'a', '界', 'ı'}, 4, []rune{'a', '界', 'ı'}},
		{[]rune{'世', 'a', '界', 'ı'}, 5, []rune{'a', '界', 'ı'}},
		{[]rune{'世', 'a', '界', 'ı'}, 6, []rune{'世', 'a', '界', 'ı'}},
		{[]rune{'世', 'a', '界', 'ı'}, 7, []rune{'世', 'a', '界', 'ı'}},
	}

	for _, test := range tests {
		if got := runeSliceWidthLastRange(test.rs, test.maxWidth); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", test.rs, test.exp, got)
		}
	}
}

func TestEscape(t *testing.T) {
	tests := []struct {
		s   string
		exp string
	}{
		{"", ""},
		{"foo", "foo"},
		{"foo bar", `foo\ bar`},
		{"foo  bar", `foo\ \ bar`},
		{`foo\bar`, `foo\\bar`},
		{`foo\ bar`, `foo\\\ bar`},
		{`foo;bar`, `foo\;bar`},
		{`foo#bar`, `foo\#bar`},
		{`foo\tbar`, `foo\\tbar`},
		{"foo\tbar", "foo\\\tbar"},
		{`foo\`, `foo\\`},
	}

	for _, test := range tests {
		if got := escape(test.s); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", test.s, test.exp, got)
		}
	}
}

func TestUnescape(t *testing.T) {
	tests := []struct {
		s   string
		exp string
	}{
		{"", ""},
		{"foo", "foo"},
		{`foo\ bar`, "foo bar"},
		{`foo\ \ bar`, "foo  bar"},
		{`foo\\bar`, `foo\bar`},
		{`foo\\\ bar`, `foo\ bar`},
		{`foo\;bar`, `foo;bar`},
		{`foo\#bar`, `foo#bar`},
		{`foo\\tbar`, `foo\tbar`},
		{"foo\\\tbar", "foo\tbar"},
		{`foo\`, `foo\`},
	}

	for _, test := range tests {
		if got := unescape(test.s); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", test.s, test.exp, got)
		}
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		s   string
		exp []string
	}{
		{"", []string{""}},
		{"foo", []string{"foo"}},
		{"foo bar", []string{"foo", "bar"}},
		{`:rename foo\ bar`, []string{":rename", `foo\ bar`}},
	}

	for _, test := range tests {
		if got := tokenize(test.s); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", test.s, test.exp, got)
		}
	}
}

func TestSplitWord(t *testing.T) {
	tests := []struct {
		s    string
		word string
		rest string
	}{
		{"", "", ""},
		{"foo", "foo", ""},
		{"  foo", "foo", ""},
		{"foo  ", "foo", ""},
		{"  foo  ", "foo", ""},
		{"foo bar baz", "foo", "bar baz"},
		{"  foo bar baz", "foo", "bar baz"},
		{"foo   bar baz", "foo", "bar baz"},
		{"  foo   bar baz", "foo", "bar baz"},
	}

	for _, test := range tests {
		if w, r := splitWord(test.s); w != test.word || r != test.rest {
			t.Errorf("at input '%s' expected '%s' and '%s' but got '%s' and '%s'", test.s, test.word, test.rest, w, r)
		}
	}
}

func TestReadPairs(t *testing.T) {
	tests := []struct {
		s   string
		exp [][]string
	}{
		{"foo bar", [][]string{{"foo", "bar"}}},
		{"foo bar ", [][]string{{"foo", "bar"}}},
		{" foo bar", [][]string{{"foo", "bar"}}},
		{" foo bar ", [][]string{{"foo", "bar"}}},
		{"foo bar#baz", [][]string{{"foo", "bar"}}},
		{"foo bar #baz", [][]string{{"foo", "bar"}}},
		{`'foo#baz' bar`, [][]string{{"foo#baz", "bar"}}},
		{`"foo#baz" bar`, [][]string{{"foo#baz", "bar"}}},
	}

	for _, test := range tests {
		if got, _ := readPairs(strings.NewReader(test.s)); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", test.s, test.exp, got)
		}
	}
}

func TestHumanize(t *testing.T) {
	tests := []struct {
		i   int64
		exp string
	}{
		{0, "0B"},
		{9, "9B"},
		{99, "99B"},
		{999, "999B"},
		{1000, "1.0K"},
		{1023, "1.0K"},
		{1025, "1.0K"},
		{1049, "1.0K"},
		{1050, "1.0K"},
		{1099, "1.0K"},
		{9999, "9.9K"},
		{10000, "10K"},
		{10100, "10K"},
		{10500, "10K"},
		{1000000, "1.0M"},
	}

	for _, test := range tests {
		if got := humanize(test.i); got != test.exp {
			t.Errorf("at input '%d' expected '%s' but got '%s'", test.i, test.exp, got)
		}
	}
}

func TestNaturalLess(t *testing.T) {
	tests := []struct {
		s1  string
		s2  string
		exp bool
	}{
		{"foo", "bar", false},
		{"bar", "baz", true},
		{"foo", "123", false},
		{"foo1", "foobar", true},
		{"foo1", "foo10", true},
		{"foo2", "foo10", true},
		{"foo1", "foo10bar", true},
		{"foo2", "foo10bar", true},
		{"foo1bar", "foo10bar", true},
		{"foo2bar", "foo10bar", true},
		{"foo1bar", "foo10", true},
		{"foo2bar", "foo10", true},
	}

	for _, test := range tests {
		if got := naturalLess(test.s1, test.s2); got != test.exp {
			t.Errorf("at input '%s' and '%s' expected '%t' but got '%t'", test.s1, test.s2, test.exp, got)
		}
	}
}
