package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

const (
	gSixelBegin     = "\033P"
	gSixelTerminate = "\033\\"
)

var (
	errInvalidSixel  = errors.New("invalid sixel sequence")
	errSixelTooLarge = errors.New("sixel too large")
	gSixelFiller     = '\u2800'
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

type CountingReader struct {
	counter int
	reader  io.Reader
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	r.counter += 1
	return r.reader.Read(p)
}

func (sxs *sixelScreen) showSixels() {
	var buf strings.Builder
	buf.WriteString("\0337")
	for _, sixel := range sxs.sx {
		buf.WriteString(fmt.Sprintf("\033[%d;%dH", sixel.y+1, sixel.x+1))
		log.Println(sixel.data)
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

func renderSixel(data io.Reader, win *win, sxScreen *sixelScreen) (lines []string, sx *sixel) {
	maxh := (win.h) * sxScreen.fonth

	sixelBuilder := strings.Builder{}
	data = io.TeeReader(data, &sixelBuilder)
	wpx, hpx, err := sixelSize(data, maxh)
	if err == errSixelTooLarge {
		return []string{"\033[7msixel image too large\033[0m"}, nil
	}
	if err != nil {
		log.Printf("measuring sixel size: %s", err)
		return []string{"\033[7merror parsing sixel\033[0m"}, nil
	}
	_, hc := sxScreen.pxToCells(wpx, hpx)

	sx = &sixel{2, 0, wpx, hpx, sixelBuilder.String()}
	for j := 1; j < hc; j++ {
		lines = append(lines, "")
	}
	return lines, sx
}

func sixelSize(data io.Reader, maxh int) (width int, height int, err error) {
	var w, h int
	counter := &CountingReader{reader: data}
	reader := bufio.NewReader(counter)
	maxh = maxh - (maxh % 6)
	defer func() {
		if errors.Is(err, errInvalidSixel) {
			err = fmt.Errorf("At position %d: %w", counter.counter, err)
		}
	}()

	// General sixel sequence:
	//    DCS <P1>;<P2>;<P3>   q   [" <raster_attributes>]   <main_body> ST
	// - DCS is "ESC P"
	// - We are not interested in P1~P3
	// - the optional raster attributes may contain the 'reported' image size in pixels
	//   (The actual image can be larger, but is at least this big)
	// - ST is the terminating string "ESC \"
	//
	// https://vt100.net/docs/vt3xx-gp/chapter14.html

	// read DCS (ESC P)
	dcs := make([]byte, 2)
	reader.Read(dcs)
	if string(dcs) != gSixelBegin {
		return 0, 0, errInvalidSixel
	}

	// read <P1>;<P2>;<P3>
	for i := 0; i < 2; i++ {
		s, err := reader.ReadString(';')
		if err != nil {
			return 0, 0, errInvalidSixel
		}
		if !isNumber(s[:len(s)-1]) {
			return 0, 0, errInvalidSixel
		}
	}
	_, err = readNumber(reader)
	if err != nil {
		return 0, 0, errInvalidSixel
	}

	// skip "q"
	if c, _, err := reader.ReadRune(); err != nil || c != 'q' {
		return 0, 0, errInvalidSixel
	}

	peek, err := reader.Peek(1)
	if err != nil {
		return 0, 0, errInvalidSixel
	}

	// optional raster attributes
	if rune(peek[0]) == '"' {
		width, height, err = parseRasterAttributes(reader)
		if err != nil {
			return 0, 0, errInvalidSixel
		}
		if height > maxh {
			return 0, 0, errSixelTooLarge
		}
	}

	// main body
	var w_line int
	newline := false
loop:
	for true {
		newline_last := newline
		newline = false
		c, _, err := reader.ReadRune()
		if err != nil {
			return 0, 0, err
		}

		switch {
		case '?' <= c && c <= '~': // data char
			w_line++
		case c == '-': // next line
			w = max(w, w_line)
			w_line = 0
			newline = true
			h++
			if h*6 > maxh {
				return w, h, errSixelTooLarge
			}
		case c == '$': // Graphics Carriage Return: go back to start of same line
			w = max(w, w_line)
			w_line = 0
		case c == '!': // Repeat Introducer
			// parse "!<number><code>"
			rep, err := readNumber(reader)
			if err != nil {
				return 0, 0, err
			}

			// the number must be followed by a data char (anything between '?' and '~')
			if next, _, err := reader.ReadRune(); err != nil {
				return 0, 0, err
			} else if next < '?' || next > '~' {
				return 0, 0, errInvalidSixel
			}

			w_line += rep
		case c == '#': // Color Controller
			err := readColorController(reader)
			if err != nil {
				return 0, 0, err
			}

		case c == gEscapeCode: // string terminator, "ESC \"
			next, _, err := reader.ReadRune()
			if err != nil {
				return 0, 0, err
			}
			if next != '\\' {
				return 0, 0, errInvalidSixel
			}
			if !newline_last {
				w = max(w, w_line)
				w_line = 0
				h++
			}
			break loop

		default:
			return 0, 0, errInvalidSixel
		}
	}

	return max(w, width), max(h*6, height), nil
}

// Start of (optional) Raster Attributes
//
//	"	Pan	;	Pad;	Ph;	Pv
//
// pixel aspect ratio = Pan/Pad
// We are only interested in Ph and Pv (horizontal and vertical size in px)
func parseRasterAttributes(reader *bufio.Reader) (height, width int, err error) {
	// skip '"'
	_, _, err = reader.ReadRune()
	if err != nil {
		return 0, 0, err
	}

	// reads and discards Pan;Pad;
	for i := 0; i < 2; i++ {
		s, err := reader.ReadString(';')
		if err != nil {
			return 0, 0, errInvalidSixel
		}
		if !isNumber(s[:len(s)-1]) {
			return 0, 0, errInvalidSixel
		}
	}

	s, err := reader.ReadString(';')
	if err != nil {
		return 0, 0, errInvalidSixel
	}
	height, err = strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return 0, 0, errInvalidSixel
	}

	width, err = readNumber(reader)
	if err != nil {
		return 0, 0, errInvalidSixel
	}

	return height, width, nil
}

// parses either:
//  1. Color Picker:
//     # id
//  2. Color Introducer:
//     # id; colorSystem; x; y; z
func readColorController(reader *bufio.Reader) error {
	_, err := readNumber(reader)
	if err != nil {
		return err
	}

	if peek, err := reader.Peek(1); err != nil {
		return err
	} else if rune(peek[0]) == ';' {
		for i := 0; i < 4; i++ {
			r, _, err := reader.ReadRune()
			if err != nil {
				return err
			}
			if r != ';' {
				return errInvalidSixel
			}
			_, err = readNumber(reader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !unicode.IsNumber(c) {
			return false
		}
	}
	return true
}

func readNumber(data *bufio.Reader) (int, error) {
	builder := strings.Builder{}

	peek, err := data.Peek(1)
	for ; true; peek, err = data.Peek(1) {
		if err != nil {
			return 0, err
		}
		if !unicode.IsNumber(rune(peek[0])) {
			break
		}
		r, _, _ := data.ReadRune()
		builder.WriteRune(r)
	}

	if builder.Len() == 0 {
		return 0, err
	}

	n, err := strconv.Atoi(builder.String())
	if err != nil {
		return 0, err
	}

	return n, nil
}
