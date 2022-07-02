package main

import (
	"bufio"
	"fmt"
	"io"
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

// This function reads whitespace separated string pairs at each line. Single
// or double quotes can be used to escape whitespaces. Hash characters can be
// used to add a comment until the end of line. Leading and trailing space is
// trimmed. Empty lines are skipped.
func readPairs(r io.Reader) ([][]string, error) {
	var pairs [][]string
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()

		squote, dquote := false, false
		for i := 0; i < len(line); i++ {
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
		pair := strings.FieldsFunc(line, func(r rune) bool {
			if r == '\'' && !dquote {
				squote = !squote
			} else if r == '"' && !squote {
				dquote = !dquote
			}
			return !squote && !dquote && unicode.IsSpace(r)
		})

		if len(pair) != 2 {
			return nil, fmt.Errorf("expected pair but found: %s", s.Text())
		}

		for i := 0; i < len(pair); i++ {
			squote, dquote = false, false
			buf := make([]rune, 0, len(pair[i]))
			for _, r := range pair[i] {
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
			pair[i] = string(buf)
		}

		pairs = append(pairs, pair)
	}

	return pairs, nil
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

// This function compares two strings for natural sorting which takes into
// account values of numbers in strings. For example, '2' is less than '10',
// and similarly 'foo2bar' is less than 'foo10bar', but 'bar2bar' is greater
// than 'foo10bar'.
func naturalLess(s1, s2 string) bool {
	lo1, lo2, hi1, hi2 := 0, 0, 0, 0
	for {
		if hi1 >= len(s1) {
			return hi2 != len(s2)
		}

		if hi2 >= len(s2) {
			return false
		}

		isDigit1 := isDigit(s1[hi1])
		isDigit2 := isDigit(s2[hi2])

		for lo1 = hi1; hi1 < len(s1) && isDigit(s1[hi1]) == isDigit1; hi1++ {
		}

		for lo2 = hi2; hi2 < len(s2) && isDigit(s2[hi2]) == isDigit2; hi2++ {
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

func mod(a, b int) int {
	return (a%b + b) % b
}

var reNumber = regexp.MustCompile(`^[0-9]+`)

// needs some testing
func sixelDimPx(s string) (w int, h int) {
	// TODO maybe take into account pixel aspect ratio
	// TODO handle macro parameters P1 to P3
	// `DCS P1;P2;P3;q...`

	// General sixel sequence:
	//    DCS <P1>;<P2>;<P3>;	q  [" <raster_attributes>];   <main_body> ST
	// DCS is "ESC P"
	// We are not interested in P1~P3
	// the optional raster attributes may contain the image size in pixels
	// ST is the terminating string "ESC \"
	i := strings.Index(s, "q") + 1
	if i == 0 {
		// syntax error
		return -1, -1
	}

	// Start of (optional) Raster Attributes
	//    "	Pan	;	Pad;	Ph;	Pv
	// pixel aspect ratio = Pan/Pad
	// We are only interested in Ph and Pv (horizontal and vertical size in px)
	if s[i] == '"' {
		i++
		b := strings.Index(s[i:], ";")
		// pan := strconv.Atoi(s[a:b])
		i += b + 1
		b = strings.Index(s[i:], ";")
		// pad := strconv.Atoi(s[a:b])

		i += b + 1
		b = strings.Index(s[i:], ";")
		ph, err1 := strconv.Atoi(s[i : i+b])

		i += b + 1
		b = strings.Index(s[i:], ";")
		pv, err2 := strconv.Atoi(s[i : i+b])

		if err1 != nil || err2 != nil {
			goto main_body // keep trying
		}

		// TODO
		// ph and pv are more like suggestions, it's still possible to go over the
		// reported size, so we might need to parse the entire main body anyway
		return ph, pv
	}

main_body:
	var wi int
	for ; i < len(s)-2; i++ {
		c := s[i]
		switch {
		case '?' <= c && c <= '~':
			wi++
		case c == '-':
			w = max(w, wi)
			wi = 0
			h++
		case c == '$':
			w = max(w, wi)
			wi = 0
		case c == '!':
			m := reNumber.FindString(s[i+1:])
			if m == "" {
				// syntax error
				return -1, -1
			}
			if s[i+1+len(m)] < '?' || s[i+1+len(m)] > '~' {
				// syntax error
				return -1, -1
			}
			n, _ := strconv.Atoi(m)
			wi += n - 1
		default:
		}
	}
	if s[len(s)-3] != '-' {
		w = max(w, wi)
		h++ // add newline on last row
	}
	return w, h * 6
}

func pxToCells(x, y, wc, hc, wpx, hpx int) (int, int) {
	var fw int = wpx / wc
	var fh int = hpx / hc
	return x/fw + 1, y/fh + 1 // TODO should probably use ceiling instead of +1
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
