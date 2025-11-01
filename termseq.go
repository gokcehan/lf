package main

import (
	"log"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

// This function is used to remove style-related ANSI escape sequences from
// a given string.
//
// *Note*: this function is based entirely on `printLength()` and strips only
// style-related sequences, `erase in line` and `OSC 8` sequences. Other codes
// (e.g., cursor moves), as well as broken escape sequences, aren't removed.
// This prevents mismatches between the two functions and avoids misalignment
// when rendering the UI.
func stripTermSequence(s string) string {
	var b strings.Builder
	slen := len(s)
	for i := 0; i < slen; i++ {
		seq := readTermSequence(s[i:])
		if seq != "" {
			i += len(seq) - 1 // skip known sequence
			continue
		}

		r, w := utf8.DecodeRuneInString(s[i:])
		i += w - 1
		b.WriteRune(r)
	}

	return b.String()
}

// This function is used to extract and return a terminal sequence from a given
// string. If no supported sequence could be found, an empty string is returned.
//
// CSI (Control Sequence Introducer):
//   - SGR (Select Graphic Rendition) `m`, used for text styling
//   - EL (Erase in Line) `K`, returned only so we can skip it
//
// OSC (Operating System Command):
//   - OSC 8, hyperlinks
func readTermSequence(s string) string {
	slen := len(s)
	// must start with ESC
	if slen < 2 || s[0] != byte(gEscapeCode) {
		return ""
	}

	switch s[1] {
	case '[': // CSI
		i := strings.IndexAny(s[:min(slen, 64)], "mK")
		if i == -1 {
			return ""
		}
		return s[:i+1]
	case ']': // OSC
		if slen < 4 || s[2] != '8' || s[3] != ';' {
			return ""
		}
		// find string terminator
		for i := 4; i < slen; i++ {
			b := s[i]
			// BEL (XTerm)
			if b == 0x07 {
				return s[:i+1]
			}
			// ESC\ (ECMA-48)
			if b == byte(gEscapeCode) && i+1 < slen && s[i+1] == '\\' {
				return s[:i+2]
			}
		}
		// TODO: C1 forms?
		return ""
	default:
		return ""
	}
}

// This function takes an escape sequence (e.g. `\033[1m`) and applies it
// to the given `tcell.Style` object.
func applyTermSequence(s string, st tcell.Style) tcell.Style {
	slen := len(s)
	if slen < 2 || s[0] != byte(gEscapeCode) {
		return st
	}
	switch s[1] {
	case '[':
		if s[slen-1] == 'm' {
			return applySGR(s[2:slen-1], st)
		}
		return st
	case ']':
		// trim terminator (BEL or ESC\), then parse body
		if s[slen-1] == 0x07 {
			return applyOSC(s[2:slen-1], st)
		} else if slen >= 2 && s[slen-2] == byte(gEscapeCode) && s[slen-1] == '\\' {
			return applyOSC(s[2:slen-2], st)
		}
		return st
	default:
		return st
	}
}

func applySGR(s string, st tcell.Style) tcell.Style {
	toks := strings.Split(s, ";")

	// ECMA-48 details the standard
	tokslen := len(toks)

loop:
	for i := 0; i < tokslen; i++ {
		switch strings.TrimLeft(toks[i], "0") {
		case "":
			st = tcell.StyleDefault
		case "1":
			st = st.Bold(true)
		case "2":
			st = st.Dim(true)
		case "3":
			st = st.Italic(true)
		case "4:0":
			st = st.Underline(false)
		case "4", "4:1":
			st = st.Underline(true)
		case "4:2":
			st = st.Underline(tcell.UnderlineStyleDouble)
		case "4:3":
			st = st.Underline(tcell.UnderlineStyleCurly)
		case "4:4":
			st = st.Underline(tcell.UnderlineStyleDotted)
		case "4:5":
			st = st.Underline(tcell.UnderlineStyleDashed)
		case "5", "6":
			st = st.Blink(true)
		case "7":
			st = st.Reverse(true)
		case "8":
			// TODO: tcell PR for proper conceal
			_, bg, _ := st.Decompose()
			st = st.Foreground(bg)
		case "9":
			st = st.StrikeThrough(true)
		case "22":
			st = st.Bold(false).Dim(false)
		case "23":
			st = st.Italic(false)
		case "24":
			st = st.Underline(false)
		case "25":
			st = st.Blink(false)
		case "27":
			st = st.Reverse(false)
		case "29":
			st = st.StrikeThrough(false)
		case "30", "31", "32", "33", "34", "35", "36", "37":
			n, _ := strconv.Atoi(toks[i])
			st = st.Foreground(tcell.PaletteColor(n - 30))
		case "90", "91", "92", "93", "94", "95", "96", "97":
			n, _ := strconv.Atoi(toks[i])
			st = st.Foreground(tcell.PaletteColor(n - 82))
		case "38":
			color, offset, err := parseColor(toks[i+1:])
			if err != nil {
				log.Printf("error processing ansi code 38: %s", err)
				break loop
			}
			st = st.Foreground(color)
			i += offset
		case "40", "41", "42", "43", "44", "45", "46", "47":
			n, _ := strconv.Atoi(toks[i])
			st = st.Background(tcell.PaletteColor(n - 40))
		case "100", "101", "102", "103", "104", "105", "106", "107":
			n, _ := strconv.Atoi(toks[i])
			st = st.Background(tcell.PaletteColor(n - 92))
		case "48":
			color, offset, err := parseColor(toks[i+1:])
			if err != nil {
				log.Printf("error processing ansi code 48: %s", err)
				break loop
			}
			st = st.Background(color)
			i += offset
		case "58":
			color, offset, err := parseColor(toks[i+1:])
			if err != nil {
				log.Printf("error processing ansi code 58: %s", err)
				break loop
			}
			st = st.Underline(color)
			i += offset
		default:
			log.Printf("unknown ansi code: %s", toks[i])
		}
	}

	return st
}

// https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
func applyOSC(body string, st tcell.Style) tcell.Style {
	toks := strings.SplitN(body, ";", 3)
	if len(toks) < 2 {
		return st
	}
	switch toks[0] {
	case "8":
		if len(toks) < 3 {
			return st
		}
		// Property used to identify grouped hyperlinks.
		// Assign URL by default to ensure a "unique" id.
		if toks[2] != "" {
			st = st.UrlId(toks[2])
		}
		// handle optional parameters
		if toks[1] != "" {
			for seg := range strings.SplitSeq(toks[1], ":") {
				if seg == "" {
					continue
				}
				if k, v, ok := strings.Cut(seg, "="); ok && k == "id" && v != "" {
					st = st.UrlId(v) // override fallback
				}
			}
		}
		return st.Url(toks[2])
	default:
		return st
	}
}
