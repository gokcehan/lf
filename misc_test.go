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

func TestExtractNums(t *testing.T) {
	names := []struct {
		s        string
		nums     []int
		rest     []string
		numFirst bool
	}{
		{"foo123bar456", []int{123, 456}, []string{"foo", "bar"}, false},
		{"123foo456bar", []int{123, 456}, []string{"foo", "bar"}, true},
		{"a-1-1.txt", []int{1, 1}, []string{"a-", "-", ".txt"}, false},
		{"a-1-10.txt", []int{1, 10}, []string{"a-", "-", ".txt"}, false},
		{"a-10-1.txt", []int{10, 1}, []string{"a-", "-", ".txt"}, false},
	}

	for _, name := range names {
		nums, rest, numFirst := extractNums(name.s)
		if !reflect.DeepEqual(nums, name.nums) {
			t.Errorf("at input '%s' expected '%v' but got '%v'", name.s, name.nums, nums)
		}
		if !reflect.DeepEqual(rest, name.rest) {
			t.Errorf("at input '%s' expected '%v' but got '%v'", name.s, name.rest, rest)
		}
		if !numFirst == name.numFirst {
			t.Errorf("at input '%s' expected '%t' but got '%t'", name.s, name.numFirst, numFirst)
		}
	}
}
