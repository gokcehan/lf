package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
)

const (
	gSixelBegin = "\033P"
)

var (
	gSixelFiller = '\u2800'
)

type sixel struct {
	data string
}

type sixelScreen struct {
	wpx, hpx     int
	xprev, yprev int
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
}

func (sxs *sixelScreen) showSixels() {
	var buf strings.Builder
	buf.WriteString("\0337")
	for _, sixel := range sxs.sx {
		buf.WriteString(fmt.Sprintf("\033[%d;%dH", sxs.yprev, sxs.xprev))
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
		hc := win.h

		for y := win.y; y < win.y+hc; y++ {
			win.print(screen, 0, y, fillStyle, strings.Repeat(string(gSixelFiller), win.w))
		}

		// TODO: move logic into showSixel
		sxs.xprev, sxs.yprev = win.x, win.y
		sxs.sx = append(sxs.sx, sx)
	}
}
