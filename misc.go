package main

import (
	"fmt"
	"path/filepath"
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

// naturalLess sorts strings in natural order.
// This means that e.g. "abc2" < "abc12".
//
// Non-digit sequences and numbers are compared separately. The former are
// compared bytewise, while the latter are compared numerically (except that
// the number of leading zeros is used as a tie-breaker, so e.g. "2" < "02")
//
// Based on NaturalLess function by Frits van Bommel
// https://github.com/fvbommel/util/blob/master/sortorder/natsort.go
func naturalLess(s1, s2 string) bool {
	s1, s2 = strings.ToLower(s1), strings.ToLower(s2)

	for len(s1) > 0 && len(s2) > 0 {
		r1, size1 := utf8.DecodeRuneInString(s1)
		r2, size2 := utf8.DecodeRuneInString(s2)

		dig1, dig2 := unicode.IsDigit(r1), unicode.IsDigit(r2)

		switch {
		case dig1 != dig2: // Digits before other characters.
			return dig1 // True if LHS is a digit, false if RHS is one.
		case !dig1: // && !dig2, becuase dig1 == dig2
			// UTF-8 compares bytewise-lexicographically, no need to decode
			// codepoints.
			if r1 != r2 {
				return r1 < r2
			}
			s1 = s1[size1:]
			s2 = s2[size2:]
		default: // Digits
			// Eat zeros.
			var nonZero1, nonZero2 int
			for len(s1) > 0 {
				r1, size := utf8.DecodeRuneInString(s1)
				if r1 != '0' {
					break
				}
				s1 = s1[size:]
				nonZero1 += size
			}
			for len(s2) > 0 {
				r2, size := utf8.DecodeRuneInString(s2)
				if r2 != '0' {
					break
				}
				s2 = s2[size:]
				nonZero2 += size
			}
			// Eat all digits.
			var len1, len2 int
			s1a, s2a := s1, s2
			for len(s1) > 0 {
				r1, size := utf8.DecodeRuneInString(s1)
				if !unicode.IsDigit(r1) {
					break
				}
				s1 = s1[size:]
				len1 += size
			}
			for len(s2) > 0 {
				r2, size := utf8.DecodeRuneInString(s2)
				if !unicode.IsDigit(r2) {
					break
				}
				s2 = s2[size:]
				len2 += size
			}
			// If lengths of numbers with non-zero prefix differ, the shorter
			// one is less.
			if len1 != len2 {
				return len1 < len2
			}
			// If they're not equal, string comparison is correct.
			if nr1, nr2 := s1a[:len1], s2a[:len2]; nr1 != nr2 {
				return nr1 < nr2
			}
			// Otherwise, the one with less zeros is less.
			// Because everything up to the number is equal, comparing the index
			// after the zeros is sufficient.
			if nonZero1 != nonZero2 {
				return nonZero1 < nonZero2
			}
		}
		// They're identical so far, so continue comparing.
	}
	// So far they are identical. At least one is ended. If the other continues,
	// it sorts last.
	return len(s1) < len(s2)
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
