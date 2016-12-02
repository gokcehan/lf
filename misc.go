package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

func isRoot(name string) bool { return filepath.Dir(name) == name }

func runeWidth(r rune) int {
	w := runewidth.RuneWidth(r)
	if w == 0 || (w == 2 && runewidth.IsAmbiguousWidth(r)) {
		w = 1
	}
	return w
}

func runeSliceWidth(rs []rune) int {
	w := 0
	for _, r := range rs {
		w += runeWidth(r)
	}
	return w
}

func runeSliceWidthRange(rs []rune, beg int, end int) []rune {
	curr := 0
	b := 0
	for i, r := range rs {
		w := runeWidth(r)
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

// This function is used to escape whitespaces with backlashes in a given
// string. If a whitespace is already escaped then it is not escaped again.
func escape(s string) string {
	esc := false
	buf := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\\' {
			esc = true
			continue
		}
		if esc {
			esc = false
			buf = append(buf, '\\', r)
			continue
		}
		if unicode.IsSpace(r) {
			buf = append(buf, '\\')
		}
		buf = append(buf, r)
	}
	return string(buf)
}

// This function is used to remove backlashes that are used to escape
// whitespaces in a given string.
func unescape(s string) string {
	esc := false
	buf := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\\' {
			esc = true
			continue
		}
		if esc {
			esc = false
			if unicode.IsSpace(r) {
				buf = append(buf, r)
			} else {
				buf = append(buf, '\\', r)
			}
			continue
		}
		buf = append(buf, r)
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

// This function converts a size in bytes to a human readable form. For this
// purpose metric suffixes are used (e.g. 1K = 1000). For values less than 10
// the first significant digit is shown, otherwise it is hidden. Numbers are
// always rounded down. For these reasons this function always show somewhat
// smaller values but it should be fine for most human beings.
func humanize(size int64) string {
	if size < 1000 {
		return fmt.Sprintf("%d", size)
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

// This function extracts numbers from a string and returns with the rest.
// It is used for numeric sorting of files when the file name consists of
// both digits and letters.
//
// For instance if your input is 'foo123bar456' you get a slice of number
// consisting of elements '123' and '456' and rest of the string as a slice
// consisting of elements 'foo' and 'bar'. The last return argument denotes
// whether or not the first partition is a number.
func extractNums(s string) (nums []int, rest []string, numFirst bool) {
	var buf []rune

	r, _ := utf8.DecodeRuneInString(s)
	digit := unicode.IsDigit(r)
	numFirst = digit

	for _, c := range s {
		if unicode.IsDigit(c) == digit {
			buf = append(buf, c)
			continue
		}

		if digit {
			i, err := strconv.Atoi(string(buf))
			if err != nil {
				log.Printf("extracting numbers: %s", err)
			}
			nums = append(nums, i)
		} else {
			rest = append(rest, string(buf))
		}

		buf = nil
		buf = append(buf, c)
		digit = !digit
	}

	if digit {
		i, err := strconv.Atoi(string(buf))
		if err != nil {
			log.Printf("extracting numbers: %s", err)
		}
		nums = append(nums, i)
	} else {
		rest = append(rest, string(buf))
	}

	return
}

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
// We don't need no thought control
// No dark templates in compiler
// Haskell leave them kids alone
// Hey Bjarne leave them kids alone
// All in all it's just another brick in the code
// All in all you're just another brick in the code
//
// -- Pink Trolled --
