package main

import (
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
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
		s       string
		minCols int
		maxCols int
		exp     [][]string
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
		if got, _ := readArrays(strings.NewReader(test.s), test.minCols, test.maxCols); !reflect.DeepEqual(got, test.exp) {
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
		{math.MaxInt64, "8.0E"},       // 8 EiB
	}

	gOpts.sizeunits = "binary"
	for _, test := range tests {
		if got := humanize(test.size); got != test.expected {
			t.Errorf("at input ('%d', '%s') expected '%s' but got '%s'", test.size, gOpts.sizeunits, test.expected, got)
		}
	}

	tests = []struct {
		size     uint64
		expected string
	}{
		{0, "0B"},
		{1, "1B"},
		{2, "2B"},
		{10, "10B"},
		{100, "100B"},
		{999, "999B"},
		{1000, "1.0K"},
		{1001, "1.0K"},
		{1049, "1.0K"},
		{1050, "1.1K"},
		{1051, "1.1K"},
		{9949, "9.9K"},
		{9950, "10.0K"},
		{9951, "10.0K"},
		{9999, "10.0K"},
		{10000, "10.0K"},
		{10001, "10.0K"},
		{99949, "99.9K"},
		{99950, "100K"},
		{99951, "100K"},
		{999499, "999K"},
		{999500, "1.0M"},
		{999501, "1.0M"},
		{999999, "1.0M"},
		{1000000, "1.0M"},
		{1000001, "1.0M"},
		{999499999, "999M"},
		{999500000, "1.0G"},
		{999500001, "1.0G"},
		{999999999, "1.0G"},
		{1000000000, "1.0G"},
		{1000000001, "1.0G"},
		{math.MaxInt64, "9.2E"},
	}

	gOpts.sizeunits = "decimal"
	for _, test := range tests {
		if got := humanize(test.size); got != test.expected {
			t.Errorf("at input ('%d', '%s') expected '%s' but got '%s'", test.size, gOpts.sizeunits, test.expected, got)
		}
	}
}

func TestPermString(t *testing.T) {
	tests := []struct {
		name string
		m    os.FileMode
		exp  string
	}{
		{"none", 0, "----------"},
		{"regular file", 0o644, "-rw-r--r--"},
		{"executable", 0o755, "-rwxr-xr-x"},
		{"directory", 0o755 | os.ModeDir, "drwxr-xr-x"},
		{"symbolic link", 0o777 | os.ModeSymlink, "lrwxrwxrwx"},
		{"named pipe", 0o644 | os.ModeNamedPipe, "prw-r--r--"},
		{"socket", 0o777 | os.ModeSocket, "srwxrwxrwx"},
		{"character device", 0o660 | os.ModeCharDevice, "crw-rw----"},
		{"block device", 0o660 | os.ModeDevice, "brw-rw----"},
		{"setuid", 0o644 | os.ModeSetuid, "-rwSr--r--"},
		{"setuid executable", 0o755 | os.ModeSetuid, "-rwsr-xr-x"},
		{"setgid", 0o644 | os.ModeSetgid, "-rw-r-Sr--"},
		{"setgid executable", 0o755 | os.ModeSetgid, "-rwxr-sr-x"},
		{"sticky", 0o644 | os.ModeSticky, "-rw-r--r-T"},
		{"sticky executable", 0o755 | os.ModeSticky, "-rwxr-xr-t"},
		{"sticky directory", 0o777 | os.ModeDir | os.ModeSticky, "drwxrwxrwt"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := permString(test.m); got != test.exp {
				t.Errorf("at input '%#o' expected '%s' but got '%s'", test.m, test.exp, got)
			}
		})
	}
}

func TestNaturalCmp(t *testing.T) {
	tests := []struct {
		s1  string
		s2  string
		exp int
	}{
		{"", "", 0},
		{"a", "a", 0},
		{"", "a", -1},
		{"a", "b", -1},
		{"a", "ab", -1},
		{"0", "0", 0},
		{"0", "00", -1},
		{"1", "1", 0},
		{"1", "01", -1},
		{"2", "10", -1},
		{"123", "foo", -1},
		{"foo", "foo1", -1},
		{"foo1", "foobar", -1},
		{"foo1", "foobar1", -1},
		{"foo2", "foo10", -1},
		{"foo2bar", "foo10bar", -1},
		{"foo0", "foo00", -1},
		{"foo0bar", "foo00bar", -1},
		{"foo1", "foo01", -1},
		{"foo1bar", "foo01bar", -1},
	}

	for _, test := range tests {
		if got := naturalCmp(test.s1, test.s2); got != test.exp {
			t.Errorf("at input '%s' and '%s' expected '%d' but got '%d'", test.s1, test.s2, test.exp, got)
		}

		if got := naturalCmp(test.s2, test.s1); got != -test.exp {
			t.Errorf("at input '%s' and '%s' expected '%d' but got '%d'", test.s2, test.s1, -test.exp, got)
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

func TestTruncateFilename(t *testing.T) {
	tests := []struct {
		file        fakeFileInfo
		maxWidth    int
		truncatePct int
		exp         string
	}{
		{fakeFileInfo{"foo", false}, 0, 0, ""},
		{fakeFileInfo{"foo", false}, 0, 100, ""},
		{fakeFileInfo{"foo", false}, 1, 0, "~"},
		{fakeFileInfo{"foo", false}, 1, 100, "~"},
		{fakeFileInfo{"foo", false}, 2, 0, "~o"},
		{fakeFileInfo{"foo", false}, 2, 100, "f~"},
		{fakeFileInfo{"foo", false}, 3, 0, "foo"},
		{fakeFileInfo{"foo", false}, 3, 100, "foo"},
		{fakeFileInfo{"foo", false}, 4, 0, "foo"},
		{fakeFileInfo{"foo", false}, 4, 100, "foo"},
		{fakeFileInfo{"foo.txt", false}, 0, 0, ""},
		{fakeFileInfo{"foo.txt", false}, 0, 100, ""},
		{fakeFileInfo{"foo.txt", false}, 1, 0, "~"},
		{fakeFileInfo{"foo.txt", false}, 1, 100, "~"},
		{fakeFileInfo{"foo.txt", false}, 2, 0, "~."},
		{fakeFileInfo{"foo.txt", false}, 2, 100, "~."},
		{fakeFileInfo{"foo.txt", false}, 3, 0, "~.t"},
		{fakeFileInfo{"foo.txt", false}, 3, 100, "~.t"},
		{fakeFileInfo{"foo.txt", false}, 4, 0, "~.tx"},
		{fakeFileInfo{"foo.txt", false}, 4, 100, "~.tx"},
		{fakeFileInfo{"foo.txt", false}, 5, 0, "~.txt"},
		{fakeFileInfo{"foo.txt", false}, 5, 100, "~.txt"},
		{fakeFileInfo{"foo.txt", false}, 6, 0, "~o.txt"},
		{fakeFileInfo{"foo.txt", false}, 6, 100, "f~.txt"},
		{fakeFileInfo{"foobarbaz", false}, 7, 0, "~barbaz"},
		{fakeFileInfo{"foobarbaz", false}, 7, 50, "foo~baz"},
		{fakeFileInfo{"foobarbaz", false}, 7, 100, "foobar~"},
		{fakeFileInfo{"foobarbaz.txt", false}, 11, 0, "~barbaz.txt"},
		{fakeFileInfo{"foobarbaz.txt", false}, 11, 50, "foo~baz.txt"},
		{fakeFileInfo{"foobarbaz.txt", false}, 11, 100, "foobar~.txt"},
		{fakeFileInfo{"foobarbaz.d", true}, 9, 0, "~barbaz.d"},
		{fakeFileInfo{"foobarbaz.d", true}, 9, 50, "foob~az.d"},
		{fakeFileInfo{"foobarbaz.d", true}, 9, 100, "foobarba~"},
		{fakeFileInfo{"世界世界.txt", false}, 10, 0, "~世界.txt"},
		{fakeFileInfo{"世界世界.txt", false}, 10, 50, "世~界.txt"},
		{fakeFileInfo{"世界世界.txt", false}, 10, 100, "世界~.txt"},
		{fakeFileInfo{"世界世界.txt", false}, 11, 0, "~界世界.txt"},
		{fakeFileInfo{"世界世界.txt", false}, 11, 50, "世~世界.txt"},
		{fakeFileInfo{"世界世界.txt", false}, 11, 100, "世界世~.txt"},
	}

	for _, test := range tests {
		if got := truncateFilename(test.file, test.maxWidth, test.truncatePct, '~'); got != test.exp {
			t.Errorf("at input (%v, %v, %v) expected '%s' but got '%s'", test.file, test.maxWidth, test.truncatePct, test.exp, got)
		}
	}
}

func TestReadLines(t *testing.T) {
	tests := []struct {
		s        string
		maxLines int
		lines    []string
		binary   bool
		sixel    bool
	}{
		{"", 10, nil, false, false},
		{"\r", 10, nil, false, false},
		{"\r\n", 10, []string{""}, false, false},
		{"\r\r\n", 10, []string{""}, false, false},
		{"\n\n", 10, []string{"", ""}, false, false},
		{"foo", 10, []string{"foo"}, false, false},
		{"foo\n", 10, []string{"foo"}, false, false},
		{"foo\r\n", 10, []string{"foo"}, false, false},
		{"foo\nbar", 10, []string{"foo", "bar"}, false, false},
		{"foo\nbar\n", 10, []string{"foo", "bar"}, false, false},
		{"foo\r\nbar", 10, []string{"foo", "bar"}, false, false},
		{"foo\r\nbar\r\n", 10, []string{"foo", "bar"}, false, false},
		{"\033[31mfoo\033[0m", 10, []string{"\033[31mfoo\033[0m"}, false, false},
		{"\000", 10, nil, true, false},
		{"foo\r\n\000\r\nbar\r\n", 10, nil, true, false},
		{"\033P\033\\", 10, []string{"\033P\033\\"}, false, true},
		{"\033Pq\"1;1;1;1#0@\033\\", 10, []string{"\033Pq\"1;1;1;1#0@\033\\"}, false, true},
		{"\033P\000\033\\", 10, []string{"\033P\000\033\\"}, false, true},
		{"\033P\n\033\\", 10, []string{"\033P\n\033\\"}, false, true},
		{"\033P\r\n\033\\", 10, []string{"\033P\r\n\033\\"}, false, true},
		{"\033P\033\\\033P\033\\", 10, []string{"\033P\033\\", "\033P\033\\"}, false, true},
		{"foo\033P\033\\bar", 10, []string{"foo", "\033P\033\\", "bar"}, false, true},
		{"foo\033P\033\\bar\033P\033\\baz", 10, []string{"foo", "\033P\033\\", "bar", "\033P\033\\", "baz"}, false, true},
		{"foo\nbar\nbaz", 2, []string{"foo", "bar"}, false, false},
		{"foo\nbar\nbaz\n", 2, []string{"foo", "bar"}, false, false},
		{"foo\nbar\033P\033\\", 2, []string{"foo", "bar"}, false, false},
		{"foo\nbar\nbaz", 3, []string{"foo", "bar", "baz"}, false, false},
		{"foo\nbar\nbaz\n", 3, []string{"foo", "bar", "baz"}, false, false},
		{"foo\nbar\033P\033\\", 3, []string{"foo", "bar", "\033P\033\\"}, false, true},
	}

	for _, test := range tests {
		lines, binary, sixel := readLines(strings.NewReader(test.s), test.maxLines)
		if !reflect.DeepEqual(lines, test.lines) || binary != test.binary || sixel != test.sixel {
			t.Errorf(
				"at input (%q, %v) expected (%#v, %v, %v) but got (%#v, %v, %v)",
				test.s, test.maxLines,
				test.lines, test.binary, test.sixel,
				lines, binary, sixel,
			)
		}
	}
}

func TestGetWidths(t *testing.T) {
	tests := []struct {
		wtot    int
		ratios  []int
		drawbox bool
		exp     []int
	}{
		{0, []int{1}, false, []int{0}},
		{0, []int{1}, true, []int{0}},
		{0, []int{1, 3, 2}, false, []int{0, 0, 0}},
		{0, []int{1, 3, 2}, true, []int{0, 0, 0}},
		{14, []int{1, 3, 2}, false, []int{2, 6, 4}},
		{16, []int{1, 3, 2}, true, []int{2, 6, 4}},
		{23, []int{1, 3, 2, 4}, false, []int{2, 6, 4, 8}}, // windows end at 2.0, 8.0, 12.0, 20.0 respectively
		{24, []int{1, 3, 2, 4}, false, []int{2, 6, 5, 8}}, // windows end at 2.1, 8.4, 12.6, 21.0 respectively
		{25, []int{1, 3, 2, 4}, false, []int{2, 7, 4, 9}}, // windows end at 2.2, 8.8, 13.2, 22.0 respectively
		{26, []int{1, 3, 2, 4}, false, []int{2, 7, 5, 9}}, // windows end at 2.3, 9.2, 13.8, 23.0 respectively
	}

	for _, test := range tests {
		widths := getWidths(test.wtot, test.ratios, test.drawbox)
		if !reflect.DeepEqual(widths, test.exp) {
			t.Errorf("at input (%v, %v, %v) expected %v but got %v", test.wtot, test.ratios, test.drawbox, test.exp, widths)
		}
	}
}
