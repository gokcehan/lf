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
	data           string
}

type sixelScreen struct {
	wpx, hpx     int
	fontw, fonth int
	sx           []sixel
	altFill      bool
	lastFile     string // TODO maybe use hash of sixels instead to flip altFill
}

func (sxs *sixelScreen) clear() {
	sxs.sx = nil
}

func (sxs *sixelScreen) fillerStyle(filePath string) tcell.Style {
	if sxs.lastFile != filePath {
		sxs.altFill = !sxs.altFill
	}

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
		buf.WriteString(sixel.data)
	}
	buf.WriteString("\0338")
	fmt.Print(buf.String())
}

// fillers are used to control when tcell redraws the region where a sixel image is drawn.
// alternating between bold and regular is to clear the image before drawing a new one.
func (sxs *sixelScreen) printFiller(win *win, screen tcell.Screen, reg *reg) {
	fillStyle := sxs.fillerStyle(reg.path)
	for _, sx := range reg.sixels {
		wc, hc := sxs.pxToCells(sx.wPx, sx.hPx)

		for y := sx.y; y < sx.y+hc; y++ {
			win.print(screen, sx.x, y, fillStyle, strings.Repeat(string(gSixelFiller), min(wc, win.w-sx.x+2)))
		}

		s := sx
		s.x += win.x
		s.y += win.y
		sxs.sx = append(sxs.sx, s)
	}
}

var reNumber = regexp.MustCompile(`^[0-9]+`)

func renderPreviewLine(text string, linenr int, win *win, sxScreen *sixelScreen) (lines []string, sx *sixel) {
	if strings.HasPrefix(text, gSixelBegin) {
		if b := strings.Index(text, gSixelTerminate); b >= 0 {
			data := text[:b+len(gSixelTerminate)]
			wpx, hpx, err := sixelDimPx(data)

			if err == nil {
				xc := 2
				yc := linenr
				maxh := (win.h - yc) * sxScreen.fonth

				// any syntax error should already be caught by sixelDimPx, error is safe to discard
				data, hpx, _ = trimSixelHeight(data, maxh)
				_, hc := sxScreen.pxToCells(wpx, hpx)

				sx = &sixel{xc, yc, wpx, hpx, data}
				for j := 1; j < hc; j++ {
					lines = append(lines, "")
				}
				return lines, sx
			}
		}
	}

	return []string{text}, nil
}

// needs some testing
func sixelDimPx(data string) (w int, h int, err error) {
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
	i := strings.Index(data, "q") + 1
	if i == 0 {
		// syntax error
		return 0, 0, errInvalidSixel
	}

	// Start of (optional) Raster Attributes
	//    "	Pan	;	Pad;	Ph;	Pv
	// pixel aspect ratio = Pan/Pad
	// We are only interested in Ph and Pv (horizontal and vertical size in px)
	if data[i] == '"' {
		i++
		b := strings.Index(data[i:], ";")
		// pan := strconv.Atoi(s[a:b])
		i += b + 1
		b = strings.Index(data[i:], ";")
		// pad := strconv.Atoi(s[a:b])

		i += b + 1
		b = strings.Index(data[i:], ";")
		ph, err1 := strconv.Atoi(data[i : i+b])

		i += b + 1
		b = strings.Index(data[i:], "#")
		pv, err2 := strconv.Atoi(data[i : i+b])
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
	for ; i < len(data)-2; i++ {
		c := data[i]
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
			m := reNumber.FindString(data[i+1:])
			if m == "" {
				// syntax error
				return 0, 0, errInvalidSixel
			}
			if data[i+1+len(m)] < '?' || data[i+1+len(m)] > '~' {
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
	if data[len(data)-3] != '-' {
		w = max(w, w_line)
		h++ // add newline on last row
	}
	return w, h * 6, nil
}

// maybe merge with sixelDimPx()
func trimSixelHeight(data string, maxh int) (res string, trimmedHeight int, err error) {
	var h int
	maxh = maxh - (maxh % 6)

	i := strings.Index(data, "q") + 1
	if i == 0 {
		// syntax error
		return "", -1, errInvalidSixel
	}

	if data[i] == '"' {
		i++
		for j := 0; j < 3; j++ {
			b := strings.Index(data[i:], ";")
			i += b + 1
		}
		b := strings.Index(data[i:], "#")
		pv, err := strconv.Atoi(data[i : i+b])

		if err == nil && pv > maxh {
			mh := strconv.Itoa(maxh)
			data = data[:i] + mh + data[i+b:]
			i += len(mh)
		} else {
			i += b
		}
	}

	for h < maxh {
		k := strings.IndexRune(data[i+1:], '-')
		if k == -1 {
			if data[len(data)-3] != '-' {
				h += 6
				i = len(data) - 3
			}
			break
		}
		i += k + 1
		h += 6
	}

	if i == 0 {
		return data, 6, nil
	}

	if len(data) > i+3 {
		return data[:i+1] + "\x1b\\", h, nil
	}

	return data, h, nil
}
