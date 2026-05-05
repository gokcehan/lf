package main

import (
	"bufio"
	"bytes"
	"cmp"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/clipperhouse/displaywidth"
)

var (
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

// firstGraphemeCluster returns the string containing the first grapheme cluster
// of the input.
func firstGraphemeCluster(s string) string {
	gr := displaywidth.StringGraphemes(s)
	gr.Next()
	return gr.Value()
}

// lastGraphemeCluster returns the string containing the last grapheme cluster
// of the input.
func lastGraphemeCluster(s string) string {
	gr := displaywidth.StringGraphemes(s)
	var last string
	for gr.Next() {
		last = gr.Value()
	}
	return last
}

// truncateRight truncates a string from the right based on Unicode widths,
// taking into account grapheme clusters.
func truncateRight(s string, maxWidth int) string {
	buf := make([]byte, 0, len(s))
	width := 0

	gr := displaywidth.StringGraphemes(s)
	for gr.Next() {
		width += gr.Width()
		if width > maxWidth {
			break
		}

		buf = append(buf, gr.Value()...)
	}

	return string(buf)
}

// truncateLeft truncates a string from the left based on Unicode widths,
// taking into account grapheme clusters.
func truncateLeft(s string, maxWidth int) string {
	type cluster struct {
		bytes []byte
		width int
	}

	var clusters []cluster
	totalWidth := 0
	gr := displaywidth.StringGraphemes(s)
	for gr.Next() {
		clusters = append(clusters, cluster{[]byte(gr.Value()), gr.Width()})
		totalWidth += gr.Width()
	}

	buf := make([]byte, 0, len(s))
	width := 0
	for _, cluster := range clusters {
		if totalWidth-width <= maxWidth {
			buf = append(buf, cluster.bytes...)
		}

		width += cluster.width
	}

	return string(buf)
}

// cmdEscape is used to escape whitespace and special characters with
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

// cmdUnescape is used to remove backslashes that are used to escape
// whitespace and special characters in a given string.
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

// tokenize splits the given string by whitespace. It is aware of escaped
// and quoted whitespace so that they are not split unintentionally.
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

// splitWord splits the first word of a string delimited by whitespace from
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

// readArrays reads whitespace-separated string arrays on each line. Single
// or double quotes can be used to escape whitespace. Hash characters can be
// used to add a comment until the end of line. Leading and trailing space is
// trimmed. Empty lines are skipped.
func readArrays(r io.Reader, minCols, maxCols int) ([][]string, error) {
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

		if arrlen < minCols || arrlen > maxCols {
			if minCols == maxCols {
				return nil, fmt.Errorf("expected %d columns but found: %s", minCols, s.Text())
			}
			return nil, fmt.Errorf("expected %d~%d columns but found: %s", minCols, maxCols, s.Text())
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

	return arrays, s.Err()
}

func readPairs(r io.Reader) ([][]string, error) {
	return readArrays(r, 2, 2)
}

// humanize converts a size in bytes to a human-readable form using
// prefixes for either binary (1 KiB = 1024 B) or decimal (1 KB = 1000 B)
// multiples. The output should be no more than 5 characters long.
func humanize(size int64) string {
	var base int64 = 1024
	if gOpts.sizeunits == "decimal" {
		base = 1000
	}

	if size < base {
		return fmt.Sprintf("%dB", size)
	}

	// Note: due to [fs.FileInfo.Size] being `int64`, the maximum
	// possible representable value would be 8 EiB or 9.2 EB.
	prefixes := []string{
		"K", // kibi (2^10) or kilo (10^3)
		"M", // mebi (2^20) or mega (10^6)
		"G", // gibi (2^30) or giga (10^9)
		"T", // tebi (2^40) or tera (10^12)
		"P", // pebi (2^50) or peta (10^15)
		"E", // exbi (2^60) or exa (10^18)
		"Z", // zebi (2^70) or zetta (10^21)
		"Y", // yobi (2^80) or yotta (10^24)
		"R", // robi (2^90) or ronna (10^27)
		"Q", // quebi (2^100) or quetta (10^30)
	}

	curr := big.NewRat(size, base)

	for _, prefix := range prefixes {
		// if curr < 99.95 then round to 1 decimal place
		if curr.Cmp(big.NewRat(9995, 100)) < 0 {
			return fmt.Sprintf("%s%s", curr.FloatString(1), prefix)
		}

		// if curr < base-0.5 then round to the nearest integer
		if curr.Cmp(new(big.Rat).Sub(big.NewRat(base, 1), big.NewRat(1, 2))) < 0 {
			return fmt.Sprintf("%s%s", curr.FloatString(0), prefix)
		}

		curr.Quo(curr, big.NewRat(base, 1))
	}

	return fmt.Sprintf("+999%s", prefixes[len(prefixes)-1])
}

// permString returns an ls(1)-style string representation of the given file
// mode, to avoid using [fs.FileMode.String], which differs slightly.
func permString(m os.FileMode) string {
	// re-use Perm()'s "-rwxrwxrwx" output and write type into b[0]
	b := []byte(m.Perm().String())
	switch {
	case m&os.ModeSymlink != 0:
		b[0] = 'l'
	case m&os.ModeDir != 0:
		b[0] = 'd'
	case m&os.ModeNamedPipe != 0:
		b[0] = 'p'
	case m&os.ModeSocket != 0:
		b[0] = 's'
	case m&os.ModeCharDevice != 0:
		b[0] = 'c'
	case m&os.ModeDevice != 0:
		b[0] = 'b'
	default:
		b[0] = '-'
	}
	// patch exec slots with suid/sgid/sticky flags
	if m&os.ModeSetuid != 0 {
		if b[3] == 'x' {
			b[3] = 's'
		} else {
			b[3] = 'S'
		}
	}
	if m&os.ModeSetgid != 0 {
		if b[6] == 'x' {
			b[6] = 's'
		} else {
			b[6] = 'S'
		}
	}
	if m&os.ModeSticky != 0 {
		if b[9] == 'x' {
			b[9] = 't'
		} else {
			b[9] = 'T'
		}
	}

	return string(b)
}

// naturalCmp compares two strings for natural sorting which takes into
// account the values of numbers in strings. For example, '2' is ordered before
// '10', and similarly 'foo2bar' ordered before 'foo10bar'. When comparing
// numbers, if they have the same value then the length of the string is also
// compared, so '0' is ordered before '00'.
func naturalCmp(s1, s2 string) int {
	s1len := len(s1)
	s2len := len(s2)

	var lo1, lo2, hi1, hi2 int
	for {
		switch {
		case hi1 >= s1len && hi2 >= s2len:
			return 0
		case hi1 >= s1len && hi2 < s2len:
			return -1
		case hi1 < s1len && hi2 >= s2len:
			return 1
		}

		lo1 = hi1
		isDigit1 := isDigit(s1[hi1])
		for hi1 < s1len && isDigit(s1[hi1]) == isDigit1 {
			hi1++
		}
		tok1 := s1[lo1:hi1]

		lo2 = hi2
		isDigit2 := isDigit(s2[hi2])
		for hi2 < s2len && isDigit(s2[hi2]) == isDigit2 {
			hi2++
		}
		tok2 := s2[lo2:hi2]

		if isDigit1 && isDigit2 {
			num1, err1 := strconv.Atoi(tok1)
			num2, err2 := strconv.Atoi(tok2)
			if err1 == nil && err2 == nil {
				if num1 != num2 {
					return cmp.Compare(num1, num2)
				} else if len(tok1) != len(tok2) {
					return cmp.Compare(len(tok1), len(tok2))
				}
			}
		}

		if tok1 != tok2 {
			return cmp.Compare(tok1, tok2)
		}
	}
}

// getFileExtension returns the extension of a file with a leading dot.
// It returns an empty string if extension could not be determined
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

// truncateFilename truncates a filename at a given position.
// The position is specified as percentage indicating where the truncation
// character will appear (0 means left, 50 means middle, 100 means right).
// The file extension is not affected by truncation, however it will be clipped
// if it exceeds the allowed width.
func truncateFilename(file fs.FileInfo, maxWidth, truncatePct int, truncateChar string) string {
	filename := sanitizeName(file.Name())
	if displaywidth.String(filename) <= maxWidth {
		return filename
	}

	ext := sanitizeName(getFileExtension(file))
	avail := maxWidth - displaywidth.String(truncateChar) - displaywidth.String(ext)
	if avail < 0 {
		return truncateRight(truncateChar+ext, maxWidth)
	}

	basename := strings.TrimSuffix(filename, ext)
	left := truncateRight(basename, avail*truncatePct/100)
	right := truncateLeft(basename, avail-displaywidth.String(left))
	return left + truncateChar + right + ext
}

// deletePathRecursive deletes entries from a map if the key is either the given
// path or a subpath of it.
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

// readLines reads lines from a file to be displayed as a preview.
// The number of lines to read is capped since files can be very large.
// Lines are split on `\n` characters, and `\r` characters are discarded.
// Individual lines are truncated to avoid unbounded memory usage on files
// with very long or no newlines.
// Sixel images are also detected and stored as separate lines.
// The presence of a null byte outside a sixel image indicates a binary file.
func readLines(reader io.ByteReader, maxLines int) (lines []string, binary bool, sixel bool) {
	const maxLineBytes = 1 << 16 // 64 KiB per line

	type state int
	const (
		stateNormal state = iota
		stateEsc
		stateSixel
		stateSixelEsc
	)
	currState := stateNormal

	var buf bytes.Buffer
	maxLinesReached := false
	flush := func(force bool) {
		if buf.Len() > 0 || force {
			lines = append(lines, buf.String())
		}
		buf.Reset()
		if len(lines) >= maxLines {
			maxLinesReached = true
		}
	}

	for !maxLinesReached {
		b, err := reader.ReadByte()
		if err != nil {
			flush(false)
			return
		}

		switch currState {
		case stateNormal:
			switch b {
			case 0:
				return nil, true, false
			case '\033':
				currState = stateEsc
			case '\r':
				// filter out carriage return
			case '\n':
				flush(true)
			default:
				if buf.Len() >= maxLineBytes {
					flush(true)
				}
				buf.WriteByte(b)
			}
		case stateEsc:
			if b == 'P' {
				flush(false)
				buf.WriteString("\033P")
				currState = stateSixel
			} else {
				buf.WriteByte('\033')
				buf.WriteByte(b)
				currState = stateNormal
			}
		case stateSixel:
			// Accept printable bytes (0x20-0x7E) and ESC inside the
			// DCS frame. Reject everything else (C0, DEL, 0x80+).
			switch {
			case b == '\033':
				buf.WriteByte(b)
				currState = stateSixelEsc
			case b >= 0x20 && b <= 0x7E:
				buf.WriteByte(b)
			default:
				buf.Reset()
				currState = stateNormal
			}
		case stateSixelEsc:
			buf.WriteByte(b)
			if b == '\\' {
				flush(true)
				sixel = true
				currState = stateNormal
			} else {
				currState = stateSixel
			}
		}
	}

	return
}

// getWidths calculates the widths of windows as the result of applying the
// `ratios` option to the screen width. One column is allocated for each divider
// between windows. When `drawbox` is enabled and `borderstyle` includes an outline,
// getWidths reserves two additional columns for the left and right borders.
func getWidths(wtot int, ratios []int, drawbox bool, borderstyle borderStyle) []int {
	rlen := len(ratios)
	wtot -= rlen - 1
	if drawbox && borderstyle&borderOutline != 0 {
		wtot -= 2
	}
	wtot = max(wtot, 0)

	rtot := 0
	for _, r := range ratios {
		rtot += r
	}

	divround := func(x, y int) int {
		return (x + y/2) / y
	}

	widths := make([]int, rlen)
	rsum := 0
	wsum := 0
	for i, r := range ratios {
		rsum += r
		widths[i] = divround(wtot*rsum, rtot) - wsum
		wsum += widths[i]
	}

	return widths
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
