package main

import (
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/collate"
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

func TestCmdEscape(t *testing.T) {
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
		if got := cmdEscape(test.s); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", test.s, test.exp, got)
		}
	}
}

func TestCmdUnescape(t *testing.T) {
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
		if got := cmdUnescape(test.s); !reflect.DeepEqual(got, test.exp) {
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
		{`foo\`, []string{`foo\`}},
		{`foo"`, []string{`foo"`}},
		{"foo bar", []string{"foo", "bar"}},
		{`foo\ bar`, []string{`foo\ bar`}},
		{`"foo bar"`, []string{`"foo bar"`}},
		{`"foo" "bar"`, []string{`"foo"`, `"bar"`}},
		{`"foo "bar"`, []string{`"foo "bar"`}},
		{`"foo\" bar"`, []string{`"foo\" bar"`}},
		{`\"foo bar\"`, []string{`\"foo`, `bar\"`}},
		{`:rename foo\ bar`, []string{":rename", `foo\ bar`}},
		{`!dir "C:\Program Files"`, []string{"!dir", `"C:\Program Files"`}},
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

func TestReadArrays(t *testing.T) {
	tests := []struct {
		s        string
		min_cols int
		max_cols int
		exp      [][]string
	}{
		{"foo bar", 2, 2, [][]string{{"foo", "bar"}}},
		{"foo bar ", 2, 2, [][]string{{"foo", "bar"}}},
		{" foo bar", 2, 2, [][]string{{"foo", "bar"}}},
		{" foo bar ", 2, 2, [][]string{{"foo", "bar"}}},
		{"foo bar#baz", 2, 2, [][]string{{"foo", "bar"}}},
		{"foo bar #baz", 2, 2, [][]string{{"foo", "bar"}}},
		{`'foo#baz' bar`, 2, 2, [][]string{{"foo#baz", "bar"}}},
		{`"foo#baz" bar`, 2, 2, [][]string{{"foo#baz", "bar"}}},
		{"foo bar baz", 3, 3, [][]string{{"foo", "bar", "baz"}}},
		{`"foo bar baz"`, 1, 1, [][]string{{"foo bar baz"}}},
	}

	for _, test := range tests {
		if got, _ := readArrays(strings.NewReader(test.s), test.min_cols, test.max_cols); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("at input '%v' expected '%v' but got '%v'", test.s, test.exp, got)
		}
	}
}

func TestHumanize(t *testing.T) {
	tests := []struct {
		size     uint64
		expected string
	}{
		{0, "0B"},
		{1, "1B"},
		{2, "2B"},
		{10, "10B"},
		{100, "100B"},
		{1000, "1000B"},
		{1023, "1023B"},
		{1024, "1.0K"},
		{1025, "1.0K"},                // 1.000976563 KiB
		{10188, "9.9K"},               // 9.94921875 KiB
		{10189, "10.0K"},              // 9.950195313 KiB
		{10240, "10.0K"},              // 10 KiB
		{10291, "10.0K"},              // 10.049804688 KiB
		{10292, "10.1K"},              // 10.05078125 KiB
		{10342, "10.1K"},              // 10.099609375 KiB
		{102348, "99.9K"},             // 99.94921875 KiB
		{102349, "100K"},              // 99.950195313 KiB
		{1023487, "999K"},             // 999.499023438 KiB
		{1023488, "1000K"},            // 999.5 KiB
		{1048063, "1023K"},            // 1023.499023438 KiB
		{1048064, "1.0M"},             // 1023.5 KiB
		{1072693248, "1023M"},         // 1023 MiB
		{1073217535, "1023M"},         // 1023.499999046 MiB
		{1073217536, "1.0G"},          // 1023.5 MiB
		{1073741824, "1.0G"},          // 1 GiB
		{1610612736, "1.5G"},          // 1.5 GiB
		{1319413953332, "1.2T"},       // 1.2 TiB
		{1463669878895412, "1.3P"},    // 1.3 PiB
		{7955158381787244544, "6.9E"}, // 6.9 EiB
		{math.MaxUint64, "16.0E"},     // 16 EiB
	}

	for _, test := range tests {
		if got := humanize(test.size); got != test.expected {
			t.Errorf("at input '%d' expected '%s' but got '%s'", test.size, test.expected, got)
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

type fakeFileInfo struct {
	name  string
	isDir bool
}

func (fileinfo fakeFileInfo) Name() string       { return fileinfo.name }
func (fileinfo fakeFileInfo) Size() int64        { return 0 }
func (fileinfo fakeFileInfo) Mode() os.FileMode  { return os.FileMode(0o000) }
func (fileinfo fakeFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (fileinfo fakeFileInfo) IsDir() bool        { return fileinfo.isDir }
func (fileinfo fakeFileInfo) Sys() any           { return nil }

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name              string
		fileName          string
		isDir             bool
		expectedExtension string
	}{
		{"normal file", "file.txt", false, ".txt"},
		{"file without extension", "file", false, ""},
		{"hidden file", ".gitignore", false, ""},
		{"hidden file with extension", ".file.txt", false, ".txt"},
		{"directory", "dir", true, ""},
		{"hidden directory", ".git", true, ""},
		{"directory with dot", "profile.d", true, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := getFileExtension(fakeFileInfo{test.fileName, test.isDir}); got != test.expectedExtension {
				t.Errorf("at input '%s' expected '%s' but got '%s'", test.fileName, test.expectedExtension, got)
			}
		})
	}
}

func TestLocaleNaturalLess(t *testing.T) {
	tests := []struct {
		s1  string
		s2  string
		exp bool
	}{
		// preserving behavior of `naturalLess`
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

		// locale sort
		{"你好", "他好", true},     // \u4F60\u597D, \u4ED6\u597D
		{"到这", "到那", false},    // \u5230\u8FD9, \u5230\u90A3
		{"你说", "什么", true},     // \u4f60\u8bf4, \u4ec0\u4e48
		{"你好", "World", false}, // \u4F60\u597D, \u57\u6f\u72\u6c\u64
		{"甲1", "甲乙", true},
		{"甲1", "甲10", true},
		{"甲2", "甲10", true},
		{"甲1", "甲10乙", true},
		{"甲2", "甲10乙", true},
		{"甲1乙", "甲10乙", true},
		{"甲2乙", "甲10乙", true},
		{"甲1乙", "甲10", true},
		{"甲2乙", "甲10", true},
	}

	localeStr := "zh-CN"
	collator, err := makeCollator(localeStr, collate.Numeric)
	if err != nil {
		t.Fatalf("failed to create collator for %q: %s", localeStr, err)
	}

	for _, test := range tests {
		if got := collator.CompareString(test.s1, test.s2) < 0; got != test.exp {
			t.Errorf("at input '%s' and '%s' expected '%t' but got '%t'", test.s1, test.s2, test.exp, got)
		}
	}
}

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		s   string
		exp string
	}{
		{"", ""},                      // empty
		{"foo bar", "foo bar"},        // plain text
		{"\033[31mRed\033[0m", "Red"}, // octal syntax
		{"\x1b[31mRed\x1b[0m", "Red"}, // hexadecimal syntax
		{"foo\x1b[31mRed", "fooRed"},  // no reset parameter
		{
			"foo\x1b[1;31;102mBoldRedGreen\x1b[0mbar",
			"fooBoldRedGreenbar",
		}, // multiple attributes
		{
			"misc.go:func \x1b[01;31m\x1b[KstripAnsi\x1b[m\x1b[K(s string) string {",
			"misc.go:func stripAnsi(s string) string {",
		}, // `grep` output containing `erase in line` sequence
	}

	for _, test := range tests {
		if got := stripAnsi(test.s); got != test.exp {
			t.Errorf("at input %q expected %q but got %q", test.s, test.exp, got)
		}
		// we rely on both functions extracting the same runes
		// to avoid misalignment
		if printLength(test.s) != len(stripAnsi(test.s)) {
			t.Errorf("at input %q expected '%d' but got '%d'", test.s, printLength(test.s), len(stripAnsi(test.s)))
		}
	}
}
