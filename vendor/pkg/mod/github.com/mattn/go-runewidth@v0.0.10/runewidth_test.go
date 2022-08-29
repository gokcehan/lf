// +build !js,!appengine

package runewidth

import (
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"testing"
	"unicode/utf8"
)

var _ sort.Interface = (*table)(nil) // ensure that type "table" does implement sort.Interface

func init() {
	os.Setenv("RUNEWIDTH_EASTASIAN", "")
	handleEnv()
}

func (t table) Len() int {
	return len(t)
}

func (t table) Less(i, j int) bool {
	return t[i].first < t[j].first
}

func (t *table) Swap(i, j int) {
	(*t)[i], (*t)[j] = (*t)[j], (*t)[i]
}

type tableInfo struct {
	tbl     table
	name    string
	wantN   int
	wantSHA string
}

var tables = []tableInfo{
	{private, "private", 137468, "a4a641206dc8c5de80bd9f03515a54a706a5a4904c7684dc6a33d65c967a51b2"},
	{nonprint, "nonprint", 2143, "288904683eb225e7c4c0bd3ee481b53e8dace404ec31d443afdbc4d13729fe95"},
	{combining, "combining", 465, "3cce13deb5e23f9f7327f2b1ef162328285a7dcf277a98302a8f7cdd43971268"},
	{doublewidth, "doublewidth", 182440, "3d16eda8650dc2c92d6318d32f0b4a74fda5a278db2d4544b1dd65863394823c"},
	{ambiguous, "ambiguous", 138739, "d05e339a10f296de6547ff3d6c5aee32f627f6555477afebd4a3b7e3cf74c9e3"},
	{emoji, "emoji", 3535, "9ec17351601d49c535658de8d129c1d0ccda2e620669fc39a2faaee7dedcef6d"},
	{notassigned, "notassigned", 10, "68441e98eca1450efbe857ac051fcc872eed347054dfd0bc662d1c4ee021d69f"},
	{neutral, "neutral", 27333, "5455f5e75c307f70b4e9b2384dc5a8bcd91a4c5e2b24b2b185dfad4d860ee5c2"},
}

func TestTableChecksums(t *testing.T) {
	for _, ti := range tables {
		gotN := 0
		buf := make([]byte, utf8.MaxRune+1)
		for r := rune(0); r <= utf8.MaxRune; r++ {
			if inTable(r, ti.tbl) {
				gotN++
				buf[r] = 1
			}
		}
		gotSHA := fmt.Sprintf("%x", sha256.Sum256(buf))
		if gotN != ti.wantN || gotSHA != ti.wantSHA {
			t.Errorf("table = %s,\n\tn = %d want %d,\n\tsha256 = %s want %s", ti.name, gotN, ti.wantN, gotSHA, ti.wantSHA)
		}
	}
}

func checkInterval(first, last rune) bool {
	return first >= 0 && first <= utf8.MaxRune &&
		last >= 0 && last <= utf8.MaxRune &&
		first <= last
}

func isCompact(t *testing.T, ti *tableInfo) bool {
	tbl := ti.tbl
	for i := range tbl {
		e := tbl[i]
		if !checkInterval(e.first, e.last) { // sanity check
			t.Errorf("table invalid: table = %s index = %d %v", ti.name, i, e)
			return false
		}
		if i+1 < len(tbl) && e.last+1 >= tbl[i+1].first { // can be combined into one entry
			t.Errorf("table not compact: table = %s index = %d %v %v", ti.name, i, e, tbl[i+1])
			return false
		}
	}
	return true
}

// This is a utility function in case that a table has changed.
func printCompactTable(tbl table) {
	counter := 0
	printEntry := func(first, last rune) {
		if counter%3 == 0 {
			fmt.Printf("\t")
		}
		fmt.Printf("{0x%04X, 0x%04X},", first, last)
		if (counter+1)%3 == 0 {
			fmt.Printf("\n")
		} else {
			fmt.Printf(" ")
		}
		counter++
	}

	sort.Sort(&tbl) // just in case
	first := rune(-1)
	for i := range tbl {
		e := tbl[i]
		if !checkInterval(e.first, e.last) { // sanity check
			panic("invalid table")
		}
		if first < 0 {
			first = e.first
		}
		if i+1 < len(tbl) && e.last+1 >= tbl[i+1].first { // can be combined into one entry
			continue
		}
		printEntry(first, e.last)
		first = -1
	}
	fmt.Printf("\n\n")
}

func TestSorted(t *testing.T) {
	for _, ti := range tables {
		if !sort.IsSorted(&ti.tbl) {
			t.Errorf("table not sorted: %s", ti.name)
		}
		if !isCompact(t, &ti) {
			t.Errorf("table not compact: %s", ti.name)
			//printCompactTable(ti.tbl)
		}
	}
}

var runewidthtests = []struct {
	in    rune
	out   int
	eaout int
}{
	{'世', 2, 2},
	{'界', 2, 2},
	{'ｾ', 1, 1},
	{'ｶ', 1, 1},
	{'ｲ', 1, 1},
	{'☆', 1, 2}, // double width in ambiguous
	{'☺', 1, 1},
	{'☻', 1, 1},
	{'♥', 1, 2},
	{'♦', 1, 1},
	{'♣', 1, 2},
	{'♠', 1, 2},
	{'♂', 1, 2},
	{'♀', 1, 2},
	{'♪', 1, 2},
	{'♫', 1, 1},
	{'☼', 1, 1},
	{'↕', 1, 2},
	{'‼', 1, 1},
	{'↔', 1, 2},
	{'\x00', 0, 0},
	{'\x01', 0, 0},
	{'\u0300', 0, 0},
	{'\u2028', 0, 0},
	{'\u2029', 0, 0},
}

func TestRuneWidth(t *testing.T) {
	c := NewCondition()
	c.EastAsianWidth = false
	for _, tt := range runewidthtests {
		if out := c.RuneWidth(tt.in); out != tt.out {
			t.Errorf("RuneWidth(%q) = %d, want %d", tt.in, out, tt.out)
		}
	}
	c.EastAsianWidth = true
	for _, tt := range runewidthtests {
		if out := c.RuneWidth(tt.in); out != tt.eaout {
			t.Errorf("RuneWidth(%q) = %d, want %d", tt.in, out, tt.eaout)
		}
	}
}

var isambiguouswidthtests = []struct {
	in  rune
	out bool
}{
	{'世', false},
	{'■', true},
	{'界', false},
	{'○', true},
	{'㈱', false},
	{'①', true},
	{'②', true},
	{'③', true},
	{'④', true},
	{'⑤', true},
	{'⑥', true},
	{'⑦', true},
	{'⑧', true},
	{'⑨', true},
	{'⑩', true},
	{'⑪', true},
	{'⑫', true},
	{'⑬', true},
	{'⑭', true},
	{'⑮', true},
	{'⑯', true},
	{'⑰', true},
	{'⑱', true},
	{'⑲', true},
	{'⑳', true},
	{'☆', true},
}

func TestIsAmbiguousWidth(t *testing.T) {
	for _, tt := range isambiguouswidthtests {
		if out := IsAmbiguousWidth(tt.in); out != tt.out {
			t.Errorf("IsAmbiguousWidth(%q) = %v, want %v", tt.in, out, tt.out)
		}
	}
}

var stringwidthtests = []struct {
	in    string
	out   int
	eaout int
}{
	{"■㈱の世界①", 10, 12},
	{"スター☆", 7, 8},
	{"つのだ☆HIRO", 11, 12},
}

func TestStringWidth(t *testing.T) {
	c := NewCondition()
	c.EastAsianWidth = false
	for _, tt := range stringwidthtests {
		if out := c.StringWidth(tt.in); out != tt.out {
			t.Errorf("StringWidth(%q) = %d, want %d", tt.in, out, tt.out)
		}
	}
	c.EastAsianWidth = true
	for _, tt := range stringwidthtests {
		if out := c.StringWidth(tt.in); out != tt.eaout {
			t.Errorf("StringWidth(%q) = %d, want %d (EA)", tt.in, out, tt.eaout)
		}
	}
}

func TestStringWidthInvalid(t *testing.T) {
	s := "こんにちわ\x00世界"
	if out := StringWidth(s); out != 14 {
		t.Errorf("StringWidth(%q) = %d, want %d", s, out, 14)
	}
}

func TestTruncateSmaller(t *testing.T) {
	s := "あいうえお"
	expected := "あいうえお"

	if out := Truncate(s, 10, "..."); out != expected {
		t.Errorf("Truncate(%q) = %q, want %q", s, out, expected)
	}
}

func TestTruncate(t *testing.T) {
	s := "あいうえおあいうえおえおおおおおおおおおおおおおおおおおおおおおおおおおおおおおお"
	expected := "あいうえおあいうえおえおおおおおおおおおおおおおおおおおおおおおおおおおおお..."
	out := Truncate(s, 80, "...")
	if out != expected {
		t.Errorf("Truncate(%q) = %q, want %q", s, out, expected)
	}
	width := StringWidth(out)
	if width != 79 {
		t.Errorf("width of Truncate(%q) should be %d, but %d", s, 79, width)
	}
}

func TestTruncateFit(t *testing.T) {
	s := "aあいうえおあいうえおえおおおおおおおおおおおおおおおおおおおおおおおおおおおおおお"
	expected := "aあいうえおあいうえおえおおおおおおおおおおおおおおおおおおおおおおおおおおお..."

	out := Truncate(s, 80, "...")
	if out != expected {
		t.Errorf("Truncate(%q) = %q, want %q", s, out, expected)
	}
	width := StringWidth(out)
	if width != 80 {
		t.Errorf("width of Truncate(%q) should be %d, but %d", s, 80, width)
	}
}

func TestTruncateJustFit(t *testing.T) {
	s := "あいうえおあいうえおえおおおおおおおおおおおおおおおおおおおおおおおおおおおおお"
	expected := "あいうえおあいうえおえおおおおおおおおおおおおおおおおおおおおおおおおおおおおお"

	out := Truncate(s, 80, "...")
	if out != expected {
		t.Errorf("Truncate(%q) = %q, want %q", s, out, expected)
	}
	width := StringWidth(out)
	if width != 80 {
		t.Errorf("width of Truncate(%q) should be %d, but %d", s, 80, width)
	}
}

func TestWrap(t *testing.T) {
	s := `東京特許許可局局長はよく柿喰う客だ/東京特許許可局局長はよく柿喰う客だ
123456789012345678901234567890

END`
	expected := `東京特許許可局局長はよく柿喰う
客だ/東京特許許可局局長はよく
柿喰う客だ
123456789012345678901234567890

END`

	if out := Wrap(s, 30); out != expected {
		t.Errorf("Wrap(%q) = %q, want %q", s, out, expected)
	}
}

func TestTruncateNoNeeded(t *testing.T) {
	s := "あいうえおあい"
	expected := "あいうえおあい"

	if out := Truncate(s, 80, "..."); out != expected {
		t.Errorf("Truncate(%q) = %q, want %q", s, out, expected)
	}
}

var isneutralwidthtests = []struct {
	in  rune
	out bool
}{
	{'→', false},
	{'┊', false},
	{'┈', false},
	{'～', false},
	{'└', false},
	{'⣀', true},
	{'⣀', true},
}

func TestIsNeutralWidth(t *testing.T) {
	for _, tt := range isneutralwidthtests {
		if out := IsNeutralWidth(tt.in); out != tt.out {
			t.Errorf("IsNeutralWidth(%q) = %v, want %v", tt.in, out, tt.out)
		}
	}
}

func TestFillLeft(t *testing.T) {
	s := "あxいうえお"
	expected := "    あxいうえお"

	if out := FillLeft(s, 15); out != expected {
		t.Errorf("FillLeft(%q) = %q, want %q", s, out, expected)
	}
}

func TestFillLeftFit(t *testing.T) {
	s := "あいうえお"
	expected := "あいうえお"

	if out := FillLeft(s, 10); out != expected {
		t.Errorf("FillLeft(%q) = %q, want %q", s, out, expected)
	}
}

func TestFillRight(t *testing.T) {
	s := "あxいうえお"
	expected := "あxいうえお    "

	if out := FillRight(s, 15); out != expected {
		t.Errorf("FillRight(%q) = %q, want %q", s, out, expected)
	}
}

func TestFillRightFit(t *testing.T) {
	s := "あいうえお"
	expected := "あいうえお"

	if out := FillRight(s, 10); out != expected {
		t.Errorf("FillRight(%q) = %q, want %q", s, out, expected)
	}
}

func TestEnv(t *testing.T) {
	old := os.Getenv("RUNEWIDTH_EASTASIAN")
	defer os.Setenv("RUNEWIDTH_EASTASIAN", old)

	os.Setenv("RUNEWIDTH_EASTASIAN", "0")
	handleEnv()

	if w := RuneWidth('│'); w != 1 {
		t.Errorf("RuneWidth('│') = %d, want %d", w, 1)
	}
}

func TestZeroWidthJoiner(t *testing.T) {
	c := NewCondition()

	var tests = []struct {
		in   string
		want int
	}{
		{"👩", 2},
		{"👩‍", 2},
		{"👩‍🍳", 2},
		{"‍🍳", 2},
		{"👨‍👨", 2},
		{"👨‍👨‍👧", 2},
		{"🏳️‍🌈", 1},
		{"あ👩‍🍳い", 6},
		{"あ‍🍳い", 6},
		{"あ‍い", 4},
	}

	for _, tt := range tests {
		if got := c.StringWidth(tt.in); got != tt.want {
			t.Errorf("StringWidth(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
