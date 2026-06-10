package main

import (
	"log"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// gEscapeCode is the byte that starts ANSI control sequences.
const gEscapeCode byte = '\x1b'

// stripTermSequence is used to remove style-related ANSI escape sequences from
// a given string.
//
// Note: this function is based on [printLength] and only strips style-related
// sequences, `erase in line`, and `OSC 8` sequences. Other codes (e.g. cursor
// moves), as well as broken escape sequences, are not removed. This prevents
// mismatches between the two functions and avoids misalignment when rendering
// the UI.
func stripTermSequence(s string) string {
	var b strings.Builder
	slen := len(s)
	for i := 0; i < slen; {
		seq := readTermSequence(s[i:])
		if seq != "" {
			i += len(seq) // skip known sequence
			continue
		}

		r, w := utf8.DecodeRuneInString(s[i:])
		i += w
		b.WriteRune(r)
	}

	return b.String()
}

// readTermSequence is used to extract and return a terminal sequence from a
// given string. If no supported sequence could be found, an empty string is
// returned.
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
	if slen < 2 || s[0] != gEscapeCode {
		return ""
	}

	switch s[1] {
	case '[': // CSI
		// Find the final byte (0x40-0x7E per ECMA-48), then check
		// if it indicates a sequence we support (SGR or EL).
		for i := 2; i < min(slen, 64); i++ {
			if s[i] >= 0x40 && s[i] <= 0x7E {
				if s[i] == 'm' || s[i] == 'K' {
					return s[:i+1]
				}
				return ""
			}
		}
		return ""
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
			if b == gEscapeCode && i+1 < slen && s[i+1] == '\\' {
				return s[:i+2]
			}
		}
		// TODO: C1 forms?
		return ""
	default:
		return ""
	}
}

// optionToFmtstr takes an escape sequence option (e.g. `\033[1m`) and outputs
// a complete format string (e.g. `\033[1m%s\033[0m`).
func optionToFmtstr(optstr string) string {
	if !strings.Contains(optstr, "%s") {
		return optstr + "%s\033[0m"
	} else {
		return optstr
	}
}

// parseEscapeSequence takes an escape sequence option (e.g. `\033[1m`) and
// converts it to a [tcell.Style] object.
// Legacy function that only accepts SGR. Kept for convenience.
func parseEscapeSequence(s string) tcell.Style {
	s = strings.TrimPrefix(s, "\033[")
	if i := strings.IndexByte(s, 'm'); i >= 0 {
		s = s[:i]
	}
	return applySGR(s, tcell.StyleDefault)
}

// applyTermSequence takes an escape sequence (e.g. `\033[1m`) and applies it
// to the given [tcell.Style] object.
// Accepts SGR and OSC sequences.
func applyTermSequence(s string, st tcell.Style) tcell.Style {
	slen := len(s)
	if slen < 2 || s[0] != gEscapeCode {
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
		} else if slen >= 2 && s[slen-2] == gEscapeCode && s[slen-1] == '\\' {
			return applyOSC(s[2:slen-2], st)
		}
		return st
	default:
		return st
	}
}

// applySGR takes an SGR sequence and applies it to the given [tcell.Style]
// object.
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
			st = st.Foreground(st.GetBackground())
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

// applyOSC takes an OSC sequence and applies it to the given [tcell.Style]
// object.
// It currently supports OSC 8 hyperlinks only, implemented as specified by
// https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda.
func applyOSC(body string, st tcell.Style) tcell.Style {
	extractID := func(params string) string {
		for seg := range strings.SplitSeq(params, ":") {
			if seg == "" {
				continue
			}
			if k, v, ok := strings.Cut(seg, "="); ok && k == "id" {
				return v
			}
		}
		return ""
	}

	toks := strings.SplitN(body, ";", 3)
	if len(toks) < 2 {
		return st
	}
	switch toks[0] {
	case "8":
		if len(toks) < 3 {
			return st
		}
		url := toks[2]
		if url == "" {
			return st
		}

		st = st.Url(url)
		// Optional property used to identify grouped hyperlinks.
		if id := extractID(toks[1]); id != "" {
			st = st.UrlId(id)
		}
		return st
	default:
		return st
	}
}

// tcellStyleToString converts a Style object to string
func tcellStyleToString(st tcell.Style) string {
	args := []string{}
	fg := st.GetForeground()

	addColor := func(c color.Color) {
		if c&color.IsRGB == 0 {
			args = append(args, "5", strconv.Itoa(int(c&^color.IsValid)))
		} else {
			r, g, b := c.RGB()
			args = append(args, "2", strconv.Itoa(int(r)), strconv.Itoa(int(g)), strconv.Itoa(int(b)))
		}
	}

	if fg != color.Default {
		if fg > color.White {
			args = append(args, "38")
			addColor(fg)
		} else if (fg - color.IsValid) < 8 {
			args = append(args, strconv.Itoa(int(30+(fg-color.IsValid))))
		} else {
			args = append(args, strconv.Itoa(int(82+(fg-color.IsValid))))
		}
	}

	bg := st.GetBackground()
	if bg != color.Default {
		if bg > color.White {
			args = append(args, "48")
			addColor(bg)
		} else if (bg - color.IsValid) < 8 {
			args = append(args, strconv.Itoa(int(40+(bg-color.IsValid))))
		} else {
			args = append(args, strconv.Itoa(int(92+(bg-color.IsValid))))
		}
	}

	if st.HasBold() {
		args = append(args, "1")
	}

	if st.HasDim() {
		args = append(args, "2")
	}

	if st.HasItalic() {
		args = append(args, "3")
	}

	if st.HasUnderline() {
		ulArg := "4"
		switch st.GetUnderlineStyle() {
		case tcell.UnderlineStyleSolid:
			ulArg = "4:1"
		case tcell.UnderlineStyleDouble:
			ulArg = "4:2"
		case tcell.UnderlineStyleCurly:
			ulArg = "4:3"
		case tcell.UnderlineStyleDotted:
			ulArg = "4:4"
		case tcell.UnderlineStyleDashed:
			ulArg = "4:5"
		}
		args = append(args, ulArg)
	}

	if st.HasBlink() {
		args = append(args, "5")
	}

	if st.HasReverse() {
		args = append(args, "7")
	}

	if st.HasStrikeThrough() {
		args = append(args, "9")
	}

	return "\x1b[" + strings.Join(args, ";") + "m"
}

// Sanitation helpers for untrusted text (filenames, previews, messages).
// Pick one of these when handling untrusted input:
//
//   sanitizePreview - replace control chars, keep tabs (preview content).
//   sanitizeName    - replace control chars and tabs (names in width/column slots).
//   sanitizeMessage - replace control chars, keep lf's own SGR/OSC8 (messages).
//
// isControlChar and isPrintable are internal predicates used by the above
// and the renderer; they are not sanitation entry points.

// isControlChar reports whether a rune is a control character or otherwise
// unsafe to display in a terminal.
// Covers C0 (0x00-0x1F), DEL (0x7F), and C1 (0x80-0x9F).
func isControlChar(r rune) bool {
	return r < 0x20 || r == 0x7F || r >= 0x80 && r <= 0x9F
}

// sanitizePreview replaces control characters and invalid bytes with the
// Unicode replacement character (U+FFFD). Tabs are preserved for content
// where tab expansion is handled by the renderer (e.g. preview panes).
func sanitizePreview(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\t' || !isControlChar(r) {
			return r
		}
		return '\uFFFD'
	}, s)
}

// sanitizeName sanitizes a filename, path, or symlink target for display.
// Unlike sanitizePreview it also replaces tabs, because tabs in names
// are expanded by the renderer to tabstop width while displaywidth.String
// counts them as width 1, causing column overflow.
func sanitizeName(s string) string {
	return strings.Map(func(r rune) rune {
		if !isControlChar(r) {
			return r
		}
		return '\uFFFD'
	}, s)
}

// sanitizeMessage sanitizes a message intended for the message line. Like
// sanitizeName it strips control runes, but it preserves terminal sequences
// that lf itself recognizes (SGR, EL, OSC 8) so internal messages that use
// color or hyperlinks still render correctly.
func sanitizeMessage(s string) string {
	s = strings.TrimRight(s, "\n\r")
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if seq := readTermSequence(s[i:]); seq != "" {
			b.WriteString(seq)
			i += len(seq)
			continue
		}
		r, w := utf8.DecodeRuneInString(s[i:])
		if isControlChar(r) {
			b.WriteRune('\uFFFD')
		} else {
			b.WriteRune(r)
		}
		i += w
	}
	return b.String()
}

// isPrintable reports whether a grapheme cluster is safe to display.
// It rejects C0/C1 controls, DEL, and invalid UTF-8.
func isPrintable(gc string) bool {
	r, size := utf8.DecodeRuneInString(gc)
	if r == utf8.RuneError && size <= 1 {
		return false
	}
	return !isControlChar(r)
}
