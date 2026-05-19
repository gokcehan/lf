package main

// Sanitation helpers for untrusted text (filenames, previews, messages).
// Pick one of these when handling untrusted input:
//
//   sanitizePreview - replace control chars, keep tabs (preview content).
//   sanitizeName    - replace control chars and tabs (names in width/column slots).
//   sanitizeMessage - replace control chars, keep lf's own SGR/OSC8 (messages).
//
// isControlChar and isPrintable are internal predicates used by the above
// and the renderer; they are not sanitation entry points.

import (
	"strings"
	"unicode/utf8"
)

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
