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
