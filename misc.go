package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
)

func isRoot(name string) bool { return filepath.Dir(name) == name }

func replaceTilde(s string) string {
	if strings.HasPrefix(s, "~") {
		s = strings.Replace(s, "~", gUser.HomeDir, 1)
	}
	return s
}

func runeSliceWidth(rs []rune) int {
	w := 0
	for _, r := range rs {
		w += runewidth.RuneWidth(r)
	}
	return w
}

func runeSliceWidthRange(rs []rune, beg, end int) []rune {
	curr := 0
	b := 0
	for i, r := range rs {
		w := runewidth.RuneWidth(r)
		switch {
		case curr == beg:
			b = i
		case curr < beg && curr+w > beg:
			b = i + 1
		case curr == end:
			return rs[b:i]
		case curr > end:
			return rs[b : i-1]
		}
		curr += w
	}
	return nil
}

// This function is used to escape whitespaces and special characters with
// backlashes in a given string.
func escape(s string) string {
	buf := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsSpace(r) || r == '\\' || r == ';' || r == '#' {
			buf = append(buf, '\\')
		}
		buf = append(buf, r)
	}
	return string(buf)
}

// This function is used to remove backlashes that are used to escape
// whitespaces and special characters in a given string.
func unescape(s string) string {
	esc := false
	buf := make([]rune, 0, len(s))
	for _, r := range s {
		if esc {
			if !unicode.IsSpace(r) && r != '\\' && r != ';' && r != '#' {
				buf = append(buf, '\\')
			}
			buf = append(buf, r)
			esc = false
			continue
		}
		if r == '\\' {
			esc = true
			continue
		}
		esc = false
		buf = append(buf, r)
	}
	if esc {
		buf = append(buf, '\\')
	}
	return string(buf)
}

// This function splits the given string by whitespaces. It is aware of escaped
// whitespaces so that they are not splitted unintentionally.
func tokenize(s string) []string {
	esc := false
	var buf []rune
	var toks []string
	for _, r := range s {
		if r == '\\' {
			esc = true
			buf = append(buf, r)
			continue
		}
		if esc {
			esc = false
			buf = append(buf, r)
			continue
		}
		if !unicode.IsSpace(r) {
			buf = append(buf, r)
		} else {
			toks = append(toks, string(buf))
			buf = nil
		}
	}
	toks = append(toks, string(buf))
	return toks
}

// This function splits the first word of a string delimited by whitespace from
// the rest. This is used to tokenize a string one by one without touching the
// rest. Whitespace on the left side of both the word and the rest are trimmed.
func splitWord(s string) (word, rest string) {
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	ind := len(s)
	for i, c := range s {
		if unicode.IsSpace(c) {
			ind = i
			break
		}
	}
	word = s[0:ind]
	rest = strings.TrimLeftFunc(s[ind:], unicode.IsSpace)
	return
}

// This function converts a size in bytes to a human readable form using metric
// suffixes (e.g. 1K = 1000). For values less than 10 the first significant
// digit is shown, otherwise it is hidden. Numbers are always rounded down.
// This should be fine for most human beings.
func humanize(size int64) string {
	if size < 1000 {
		return fmt.Sprintf("%dB", size)
	}

	suffix := []string{
		"K", // kilo
		"M", // mega
		"G", // giga
		"T", // tera
		"P", // peta
		"E", // exa
		"Z", // zeta
		"Y", // yotta
	}

	curr := float64(size) / 1000
	for _, s := range suffix {
		if curr < 10 {
			return fmt.Sprintf("%.1f%s", curr-0.0499, s)
		} else if curr < 1000 {
			return fmt.Sprintf("%d%s", int(curr), s)
		}
		curr /= 1000
	}

	return ""
}

// This regexp is used to partition a given string as numbers and non-numbers.
// For instance, if your input is 'foo123bar456' you get a slice of 'foo',
// '123', 'bar', and '456'. This is useful for natural sorting which takes into
// account values of numbers within strings.
var rePart = regexp.MustCompile(`\d+|\D+`)

func naturalLess(s1, s2 string) bool {
	parts1 := rePart.FindAllString(s1, -1)
	parts2 := rePart.FindAllString(s2, -1)

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		if parts1[i] == parts2[i] {
			continue
		}

		num1, err1 := strconv.Atoi(parts1[i])
		num2, err2 := strconv.Atoi(parts2[i])

		if err1 == nil && err2 == nil {
			return num1 < num2
		}

		return parts1[i] < parts2[i]
	}

	return len(parts1) < len(parts2)
}

var reAltKey = regexp.MustCompile(`<a-(.)>`)

var reWord = regexp.MustCompile(`(\pL|\pN)+`)
var reWordBeg = regexp.MustCompile(`([^\pL\pN]|^)(\pL|\pN)`)
var reWordEnd = regexp.MustCompile(`(\pL|\pN)([^\pL\pN]|$)`)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// We don't need no generic code
// We don't need no type control
// No dark templates in compiler
// Haskell leave them kids alone
// Hey Bjarne leave them kids alone
// All in all it's just another brick in the code
// All in all you're just another brick in the code
//
// -- Pink Trolled --
