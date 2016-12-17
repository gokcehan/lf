package main

import (
	"reflect"
	"testing"
)

func TestIsRoot(t *testing.T) {
	if !isRoot("/") {
		t.Errorf(`"/" is root`)
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

func TestRuneWidth(t *testing.T) {
	chars := []struct {
		r rune
		w int
	}{
		{' ', 1},
		{'a', 1},
		{'ı', 1},
		{'ş', 1},
		{'世', 2},
		{'界', 2},
	}

	for _, char := range chars {
		if w := runeWidth(char.r); w != char.w {
			t.Errorf("at input '%c' expected '%d' but got '%d'", char.r, char.w, w)
		}
	}
}

func TestRuneSliceWidth(t *testing.T) {
	slices := []struct {
		s []rune
		w int
	}{
		{[]rune{'a', 'b'}, 2},
		{[]rune{'ı', 'ş'}, 2},
		{[]rune{'世', '界'}, 4},
		{[]rune{'世', 'a', '界', 'ı'}, 6},
	}

	for _, slice := range slices {
		if w := runeSliceWidth(slice.s); w != slice.w {
			t.Errorf("at input '%v' expected '%d' but got '%d'", slice.s, slice.w, w)
		}
	}
}

func TestRuneSliceWidthRange(t *testing.T) {
	slices := []struct {
		s []rune
		i int
		j int
		r []rune
	}{
		{[]rune{'a', 'b', 'c', 'd'}, 1, 3, []rune{'b', 'c'}},
		{[]rune{'a', 'ı', 'b', 'ş'}, 1, 3, []rune{'ı', 'b'}},
		{[]rune{'世', '界', '世', '界'}, 2, 6, []rune{'界', '世'}},
		{[]rune{'世', '界', '世', '界'}, 3, 6, []rune{'世'}},
		{[]rune{'世', '界', '世', '界'}, 2, 5, []rune{'界'}},
		{[]rune{'世', '界', '世', '界'}, 3, 5, []rune{}},
		{[]rune{'世', 'a', '界', 'ı'}, 2, 5, []rune{'a', '界'}},
		{[]rune{'世', 'a', '界', 'ı'}, 2, 4, []rune{'a'}},
		{[]rune{'世', 'a', '界', 'ı'}, 3, 5, []rune{'界'}},
		{[]rune{'世', 'a', '界', 'ı'}, 3, 4, []rune{}},
	}

	for _, slice := range slices {
		if r := runeSliceWidthRange(slice.s, slice.i, slice.j); !reflect.DeepEqual(r, slice.r) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", slice.s, slice.r, r)
		}
	}
}

func TestEscape(t *testing.T) {
	strs := []struct {
		s string
		e string
	}{
		{"", ""},
		{"foo", "foo"},
		{"foo bar", `foo\ bar`},
		{"foo  bar", `foo\ \ bar`},
		{`foo\ bar`, `foo\ bar`},
		{`foo\bar`, `foo\bar`},
	}

	for _, str := range strs {
		if e := escape(str.s); !reflect.DeepEqual(e, str.e) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", str.s, str.e, e)
		}
	}
}

func TestUnescape(t *testing.T) {
	strs := []struct {
		s string
		u string
	}{
		{"", ""},
		{"foo", "foo"},
		{`foo\bar`, `foo\bar`},
		{`foo\ bar`, "foo bar"},
		{`foo\ \ bar`, "foo  bar"},
	}

	for _, str := range strs {
		if u := unescape(str.s); !reflect.DeepEqual(u, str.u) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", str.s, str.u, u)
		}
	}
}

func TestTokenize(t *testing.T) {
	strs := []struct {
		s    string
		toks []string
	}{
		{"", []string{""}},
		{"foo", []string{"foo"}},
		{"foo bar", []string{"foo", "bar"}},
		{`:rename foo\ bar`, []string{":rename", `foo\ bar`}},
	}

	for _, str := range strs {
		if toks := tokenize(str.s); !reflect.DeepEqual(toks, str.toks) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", str.s, str.toks, toks)
		}
	}
}

func TestSplitWord(t *testing.T) {
	strs := []struct {
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

	for _, str := range strs {
		if w, r := splitWord(str.s); w != str.word || r != str.rest {
			t.Errorf("at input '%s' expected '%s' and '%s' but got '%s' and '%s'", str.s, str.word, str.rest, w, r)
		}
	}
}

func TestHumanize(t *testing.T) {
	nums := []struct {
		i int64
		s string
	}{
		{0, "0"},
		{9, "9"},
		{99, "99"},
		{999, "999"},
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

	for _, num := range nums {
		if h := humanize(num.i); h != num.s {
			t.Errorf("at input '%d' expected '%s' but got '%s'", num.i, num.s, h)
		}
	}
}

func TestNaturalLess(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected bool
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
		if b := naturalLess(test.s1, test.s2); b != test.expected {
			t.Errorf("at input '%s' and '%s' expected '%t' but got '%t'", test.s1, test.s2, test.expected, b)
		}
	}
}
