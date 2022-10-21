package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

const (
	gSixelBegin     = "\033P"
	gSixelTerminate = "\033\\"
)

var (
	gSixelFiller = '\u2800'
)

type sixel struct {
	x, y, wPx, hPx int
	str            string
}

type sixelScreen struct {
	wpx, hpx     int
	fontw, fonth int
	sx           []sixel
	lastFile     string
	altFill      bool
}

func (sxs *sixelScreen) clear() {
	sxs.sx = nil
}

// fillers are used to control when tcell redraws the region where a sixel image is drawn.
// alternating between bold("ESC [1m") and regular is to clear the image before drawing a new one.
func (sxs *sixelScreen) filler(path string, l int) (fill string) {
	if path != sxs.lastFile {
		sxs.altFill = !sxs.altFill
		sxs.lastFile = path
	}

	if sxs.altFill {
		fill = "\033[1m"
		defer func() {
			fill += "\033[0m"
		}()
	}

	fill += strings.Repeat(string(gSixelFiller), l)
	return
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
}

var reNumber = regexp.MustCompile(`^[0-9]+`)

// needs some testing
func sixelDimPx(s string) (w int, h int) {
	// TODO maybe take into account pixel aspect ratio

	// General sixel sequence:
	//    DCS <P1>;<P2>;<P3>;	q  [" <raster_attributes>]   <main_body> ST
	// DCS is "ESC P"
	// We are not interested in P1~P3
	// the optional raster attributes may contain the 'reported' image size in pixels
	// (The actual image can be larger, but is at least this big)
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
		b = strings.Index(s[i:], "#")
		pv, err2 := strconv.Atoi(s[i : i+b])
		i += b

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

// maybe merge with sixelDimPx()
func trimSixelHeight(s string, maxh int) (string, int) {
	var h int
	maxh = maxh - (maxh % 6)

	i := strings.Index(s, "q") + 1
	if i == 0 {
		// syntax error
		return "", -1
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
		return s, 6
	}

	if len(s) > i+3 {
		return s[:i+1] + "\x1b\\", h
	}

	return s, h
}
