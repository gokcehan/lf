package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

const (
	gSixelBegin     = "\033P"
	gSixelTerminate = "\033\\"
)

var (
	errInvalidSixel = errors.New("invalid sixel sequence")
	gSixelFiller    = '\u2800'
)

type sixel struct {
	x, y, wPx, hPx int
	str            string
}

type sixelScreen struct {
	wpx, hpx     int
	fontw, fonth int
	sx           []sixel
	altFill      bool
}

func (sxs *sixelScreen) clear() {
	sxs.sx = nil
}

// fillers are used to control when tcell redraws the region where a sixel image is drawn.
// alternating between bold and regular is to clear the image before drawing a new one.
func (sxs *sixelScreen) fillerStyle() tcell.Style {
	if sxs.altFill {
		return tcell.StyleDefault.Bold(true)
	}
	return tcell.StyleDefault
}

func newSixelScreen(wc, hc int) (sxs sixelScreen) {
	sxs.updateSizes(wc, hc)
	return sxs
}

func (sxs *sixelScreen) updateSizes(wc, hc int) {
	var err error
	sxs.wpx, sxs.hpx, err = getTermPixels()
	if err != nil {
		sxs.wpx, sxs.hpx = -1, -1
		log.Printf("getting terminal pixel size: %s", err)
	}

	sxs.fontw = sxs.wpx / wc
	sxs.fonth = sxs.hpx / hc
}

func (sxs *sixelScreen) pxToCells(x, y int) (int, int) {
	return x/sxs.fontw + 1, y/sxs.fonth + 1
}

func (sxs *sixelScreen) showSixels() {
	var buf strings.Builder
	buf.WriteString("\0337")
	for _, sixel := range sxs.sx {
		buf.WriteString(fmt.Sprintf("\033[%d;%dH", sixel.y+1, sixel.x+1))
		buf.WriteString(sixel.str)
	}
	buf.WriteString("\0338")
	fmt.Print(buf.String())
	sxs.altFill = !sxs.altFill

}

var reNumber = regexp.MustCompile(`^[0-9]+`)

func renderPreviewLine(text string, linenr int, fpath string, win *win, sxScreen *sixelScreen) (lines []string, sixels []sixel) {
	if gOpts.sixel && sxScreen.wpx > 0 && sxScreen.hpx > 0 {
		if a := strings.Index(text, gSixelBegin); a >= 0 {
			if b := strings.Index(text[a:], gSixelTerminate); b >= 0 {
				textbefore := text[:a]
				s := text[a : a+b+len(gSixelTerminate)]
				textafter := text[a+b+len(gSixelTerminate):]
				wpx, hpx, err := sixelDimPx(s)

				if err == nil {
					xc := runeSliceWidth([]rune(textbefore)) + 2
					yc := linenr
					maxh := (win.h - yc) * sxScreen.fonth

					// any syntax error should already be caught by sixelDimPx
					s, hpx, _ = trimSixelHeight(s, maxh)
					_, hc := sxScreen.pxToCells(wpx, hpx)

					lines = append(lines, textbefore)

					sixels = append(sixels, sixel{xc, yc, wpx, hpx, s})
					for j := 1; j < hc; j++ {
						lines = append(lines, "")
					}

					linesAfter, sixelsAfter := renderPreviewLine(textafter, linenr, fpath, win, sxScreen)
					lines = append(lines, linesAfter...)
					sixels = append(sixels, sixelsAfter...)
					return lines, sixels
				}
			}
		}
	}
	return []string{text}, sixels
}

// needs some testing
func sixelDimPx(s string) (w int, h int, err error) {
	// TODO maybe take into account pixel aspect ratio

	// General sixel sequence:
	//    DCS <P1>;<P2>;<P3>;	q  [" <raster_attributes>]   <main_body> ST
	// DCS is "ESC P"
	// We are not interested in P1~P3
	// the optional raster attributes may contain the 'reported' image size in pixels
	// (The actual image can be larger, but is at least this big)
	// ST is the terminating string "ESC \"
	//
	// https://vt100.net/docs/vt3xx-gp/chapter14.html
	i := strings.Index(s, "q") + 1
	if i == 0 {
		// syntax error
		return 0, 0, errInvalidSixel
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
		b = strings.Index(s[i:], "#")
		pv, err2 := strconv.Atoi(s[i : i+b])
		i += b

		if err1 != nil || err2 != nil {
			goto main_body // keep trying
		}

		// TODO
		// ph and pv are more like suggestions, it's still possible to go over the
		// reported size, so we might need to parse the entire main body anyway
		return ph, pv, nil
	}

main_body:
	var w_line int
	for ; i < len(s)-2; i++ {
		c := s[i]
		switch {
		case '?' <= c && c <= '~': // data char
			w_line++
		case c == '-': // next line
			w = max(w, w_line)
			w_line = 0
			h++
		case c == '$': // Graphics Carriage Return: go back to start of same line
			w = max(w, w_line)
			w_line = 0
		case c == '!': // Repeat Introducer
			m := reNumber.FindString(s[i+1:])
			if m == "" {
				// syntax error
				return 0, 0, errInvalidSixel
			}
			if s[i+1+len(m)] < '?' || s[i+1+len(m)] > '~' {
				// syntax error
				return 0, 0, errInvalidSixel
			}
			n, _ := strconv.Atoi(m)
			w_line += n - 1
		default:
			// other cases:
			//   c == '#' (change color)
		}
	}
	if s[len(s)-3] != '-' {
		w = max(w, w_line)
		h++ // add newline on last row
	}
	return w, h * 6, nil
}

// maybe merge with sixelDimPx()
func trimSixelHeight(s string, maxh int) (seq string, trimmedHeight int, err error) {
	var h int
	maxh = maxh - (maxh % 6)

	i := strings.Index(s, "q") + 1
	if i == 0 {
		// syntax error
		return "", -1, errInvalidSixel
	}

	if s[i] == '"' {
		i++
		for j := 0; j < 3; j++ {
			b := strings.Index(s[i:], ";")
			i += b + 1
		}
		b := strings.Index(s[i:], "#")
		pv, err := strconv.Atoi(s[i : i+b])

		if err == nil && pv > maxh {
			mh := strconv.Itoa(maxh)
			s = s[:i] + mh + s[i+b:]
			i += len(mh)
		} else {
			i += b
		}
	}

	for h < maxh {
		k := strings.IndexRune(s[i+1:], '-')
		if k == -1 {
			if s[len(s)-3] != '-' {
				h += 6
				i = len(s) - 3
			}
			break
		}
		i += k + 1
		h += 6
	}

	if i == 0 {
		return s, 6, nil
	}

	if len(s) > i+3 {
		return s[:i+1] + "\x1b\\", h, nil
	}

	return s, h, nil
}
