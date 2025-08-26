package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

var (
	reModKey    = regexp.MustCompile(`<(c|s|a)-(.+)>`)
	reRulerSub  = regexp.MustCompile(`%[apmcsvfithPd]|%\{[^}]+\}`)
	reSixelSize = regexp.MustCompile(`"1;1;(\d+);(\d+)`)
)

var (
	reWord    = regexp.MustCompile(`(\pL|\pN)+`)
	reWordBeg = regexp.MustCompile(`([^\pL\pN]|^)(\pL|\pN)`)
	reWordEnd = regexp.MustCompile(`(\pL|\pN)([^\pL\pN]|$)`)
)

func isRoot(name string) bool { return filepath.Dir(name) == name }

func replaceTilde(s string) string {
	if strings.HasPrefix(s, "~") {
		return gUser.HomeDir + s[1:]
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
	if beg == end {
		return []rune{}
	}

	curr := 0
	b := 0
	foundb := false
	for i, r := range rs {
		w := runewidth.RuneWidth(r)
		if curr >= beg && !foundb {
			b = i
			foundb = true
		}
		if curr == end || curr+w > end {
			return rs[b:i]
		}
		curr += w
	}

	return rs[b:]
}

// Returns the last runes of `rs` that take up at most `maxWidth` space.
func runeSliceWidthLastRange(rs []rune, maxWidth int) []rune {
	lastWidth := 0
	for i := len(rs) - 1; i >= 0; i-- {
		w := runewidth.RuneWidth(rs[i])
		if lastWidth+w > maxWidth {
			return rs[i+1:]
		}
		lastWidth += w
	}
	return rs
}

// This function is used to escape whitespaces and special characters with
// backslashes in a given string.
func cmdEscape(s string) string {
	buf := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsSpace(r) || r == '\\' || r == ';' || r == '#' {
			buf = append(buf, '\\')
		}
		buf = append(buf, r)
	}
	return string(buf)
}

// This function is used to remove backslashes that are used to escape
// whitespaces and special characters in a given string.
func cmdUnescape(s string) string {
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
// and quoted whitespaces so that they are not split unintentionally.
func tokenize(s string) []string {
	esc := false
	quote := false
	var buf []rune
	var toks []string
	for _, r := range s {
		switch {
		case esc:
			esc = false
			buf = append(buf, r)
		case r == '\\':
			esc = true
			buf = append(buf, r)
		case r == '"':
			quote = !quote
			buf = append(buf, r)
		case unicode.IsSpace(r) && !quote:
			toks = append(toks, string(buf))
			buf = nil
		default:
			buf = append(buf, r)
		}
	}
	return append(toks, string(buf))
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

// This function reads whitespace separated string arrays at each line. Single
// or double quotes can be used to escape whitespaces. Hash characters can be
// used to add a comment until the end of line. Leading and trailing space is
// trimmed. Empty lines are skipped.
func readArrays(r io.Reader, min_cols, max_cols int) ([][]string, error) {
	var arrays [][]string
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()

		squote, dquote := false, false
		for i := range len(line) {
			if line[i] == '\'' && !dquote {
				squote = !squote
			} else if line[i] == '"' && !squote {
				dquote = !dquote
			}
			if !squote && !dquote && line[i] == '#' {
				line = line[:i]
				break
			}
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		squote, dquote = false, false
		arr := strings.FieldsFunc(line, func(r rune) bool {
			if r == '\'' && !dquote {
				squote = !squote
			} else if r == '"' && !squote {
				dquote = !dquote
			}
			return !squote && !dquote && unicode.IsSpace(r)
		})
		arrlen := len(arr)

		if arrlen < min_cols || arrlen > max_cols {
			if min_cols == max_cols {
				return nil, fmt.Errorf("expected %d columns but found: %s", min_cols, s.Text())
			}
			return nil, fmt.Errorf("expected %d~%d columns but found: %s", min_cols, max_cols, s.Text())
		}

		for i := range arrlen {
			squote, dquote = false, false
			buf := make([]rune, 0, len(arr[i]))
			for _, r := range arr[i] {
				if r == '\'' && !dquote {
					squote = !squote
					continue
				}
				if r == '"' && !squote {
					dquote = !dquote
					continue
				}
				buf = append(buf, r)
			}
			arr[i] = string(buf)
		}

		arrays = append(arrays, arr)
	}

	return arrays, nil
}

func readPairs(r io.Reader) ([][]string, error) {
	return readArrays(r, 2, 2)
}

// This function converts a size in bytes to a human readable form using
// prefixes for either binary (1 KiB = 1024 B) or decimal (1 KB = 1000 B)
// multiples. The output should be no more than 5 characters long.
func humanize(size uint64) string {
	var base uint64 = 1024
	if gOpts.sizeunits == "decimal" {
		base = 1000
	}

	if size < base {
		return fmt.Sprintf("%dB", size)
	}

	// *Note*: due to [`FileSize.Size()`](https://pkg.go.dev/io/fs#FileInfo)
	// being `int64`, the maximum possible representable value would be 8 EiB or
	// 9.2 EB.
	prefixes := []string{
		"K", // kibi (2^10) or kilo (10^3)
		"M", // mebi (2^20) or mega (10^6)
		"G", // gibi (2^30) or giga (10^9)
		"T", // tebi (2^40) or tera (10^2)
		"P", // pebi (2^50) or peta (10^15)
		"E", // exbi (2^60) or exa (10^18)
		"Z", // zebi (2^70) or zetta (10^21)
		"Y", // yobi (2^80) or yotta (10^24)
		"R", // robi (2^90) or ronna (10^27)
		"Q", // quebi (2^100) or quetta (10^30)
	}

	curr := big.NewRat(int64(size), int64(base))

	for _, prefix := range prefixes {
		// if curr < 99.95 then round to 1 decimal place
		if curr.Cmp(big.NewRat(9995, 100)) < 0 {
			return fmt.Sprintf("%s%s", curr.FloatString(1), prefix)
		}

		// if curr < base-0.5 then round to the nearest integer
		if curr.Cmp(new(big.Rat).Sub(big.NewRat(int64(base), 1), big.NewRat(1, 2))) < 0 {
			return fmt.Sprintf("%s%s", curr.FloatString(0), prefix)
		}

		curr.Quo(curr, big.NewRat(int64(base), 1))
	}

	return fmt.Sprintf("+999%s", prefixes[len(prefixes)-1])
}

// This function compares two strings for natural sorting which takes into
// account values of numbers in strings. For example, '2' is less than '10',
// and similarly 'foo2bar' is less than 'foo10bar', but 'bar2bar' is greater
// than 'foo10bar'.
func naturalLess(s1, s2 string) bool {
	lo1, lo2, hi1, hi2 := 0, 0, 0, 0
	s1len := len(s1)
	s2len := len(s2)
	for {
		if hi1 >= s1len {
			return hi2 != s2len
		}

		if hi2 >= s2len {
			return false
		}

		isDigit1 := isDigit(s1[hi1])
		isDigit2 := isDigit(s2[hi2])

		for lo1 = hi1; hi1 < s1len && isDigit(s1[hi1]) == isDigit1; hi1++ {
		}

		for lo2 = hi2; hi2 < s2len && isDigit(s2[hi2]) == isDigit2; hi2++ {
		}

		if s1[lo1:hi1] == s2[lo2:hi2] {
			continue
		}

		if isDigit1 && isDigit2 {
			num1, err1 := strconv.Atoi(s1[lo1:hi1])
			num2, err2 := strconv.Atoi(s2[lo2:hi2])

			if err1 == nil && err2 == nil {
				return num1 < num2
			}
		}

		return s1[lo1:hi1] < s2[lo2:hi2]
	}
}

// This function returns the extension of a file with a leading dot
// it returns an empty string if extension could not be determined
// i.e. directories, filenames without extensions
func getFileExtension(file fs.FileInfo) string {
	if file.IsDir() {
		return ""
	}
	if strings.Count(file.Name(), ".") == 1 && file.Name()[0] == '.' {
		// hidden file without extension
		return ""
	}
	return filepath.Ext(file.Name())
}

// This function deletes entries from a map if the key is either the given path
// or a subpath of it.
// This is useful for clearing cached data when a directory is moved or deleted.
func deletePathRecursive[T any](m map[string]T, path string) {
	delete(m, path)
	prefix := path + string(filepath.Separator)
	for k := range m {
		if strings.HasPrefix(k, prefix) {
			delete(m, k)
		}
	}
}

// This function takes an escape sequence option (e.g. `\033[1m`) and outputs a
// complete format string (e.g. `\033[1m%s\033[0m`).
func optionToFmtstr(optstr string) string {
	if !strings.Contains(optstr, "%s") {
		return optstr + "%s\033[0m"
	} else {
		return optstr
	}
}

// This function takes an escape sequence option (e.g. `\033[1m`) and converts
// it to a `tcell.Style` object.
func parseEscapeSequence(s string) tcell.Style {
	s = strings.TrimPrefix(s, "\033[")
	if i := strings.IndexByte(s, 'm'); i >= 0 {
		s = s[:i]
	}
	return applyAnsiCodes(s, tcell.StyleDefault)
}

// This function is used to remove style-related ANSI escape sequences from
// a given string.
//
// *Note*: this function is based entirely on `printLength()` and strips only
// style-related escape sequences and the `erase in line` sequence. Other codes
// (e.g., cursor moves), as well as broken escape sequences, aren't removed.
// This prevents mismatches between the two functions and avoids misalignment
// when rendering the UI.
func stripAnsi(s string) string {
	var b strings.Builder
	slen := len(s)
	for i := 0; i < slen; i++ {
		r, w := utf8.DecodeRuneInString(s[i:])

		if r == gEscapeCode && i+1 < slen && s[i+1] == '[' {
			j := strings.IndexAny(s[i:min(slen, i+64)], "mK")
			if j == -1 {
				continue
			}

			i += j
			continue
		}

		i += w - 1
		b.WriteRune(r)
	}

	return b.String()
}

// This function reads lines from a file to be displayed as a preview.
// The number of lines to read is capped since files can be very large.
// Lines are split on `\n` characters, and `\r` characters are discarded.
// Sixel images are also detected and stored as separate lines.
// The presence of a null byte outside a sixel image indicates a binary file.
func readLines(reader io.ByteReader, maxLines int) (lines []string, binary bool, sixel bool) {
	var buf bytes.Buffer
	var last byte
	inSixel := false

	for {
		b, err := reader.ReadByte()
		if err != nil {
			if buf.Len() > 0 {
				lines = append(lines, buf.String())
			}
			return
		}

		if inSixel {
			buf.WriteByte(b)
			if b == '\\' && last == '\033' {
				lines = append(lines, buf.String())
				buf.Reset()
				if len(lines) >= maxLines {
					return
				}
				inSixel = false
			}
		} else {
			switch {
			case b == 0:
				return nil, true, false
			case b == '\033':
				// withhold as it could be the start of a sixel image
			case b == 'P' && last == '\033':
				if buf.Len() > 0 {
					lines = append(lines, buf.String())
					buf.Reset()
					if len(lines) >= maxLines {
						return
					}
				}
				buf.WriteByte(last)
				buf.WriteByte(b)
				inSixel = true
				sixel = true
			case last == '\033':
				// not a sixel image
				buf.WriteByte(last)
				buf.WriteByte(b)
			case b == '\r':
			case b == '\n':
				lines = append(lines, buf.String())
				buf.Reset()
				if len(lines) >= maxLines {
					return
				}
			default:
				buf.WriteByte(b)
			}
		}

		last = b
	}
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
