package main

import (
	"fmt"
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
	xprev, yprev int
	sixel        *sixel
	altFill      bool
	lastFile     string // TODO maybe use hash of sixels instead to flip altFill
}

func (sxs *sixelScreen) clear() {
	sxs.sixel = nil
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

func newSixelScreen() (sxs sixelScreen) {
	return sxs
}

func (sxs *sixelScreen) showSixels() {
	var buf strings.Builder
	buf.WriteString("\0337")
	if sxs.sixel != nil {
		buf.WriteString(fmt.Sprintf("\033[%d;%dH", sxs.yprev, sxs.xprev))
		buf.WriteString(sxs.sixel.data)
	}
	buf.WriteString("\0338")
	fmt.Print(buf.String())
}

// fillers are used to control when tcell redraws the region where a sixel image is drawn.
// alternating between bold and regular is to clear the image before drawing a new one.
func (sxs *sixelScreen) printFiller(win *win, screen tcell.Screen, reg *reg) {
	if reg.sixel == nil {
		return
	}
	fillStyle := sxs.fillerStyle(reg.path)

	for y := 0; y < win.h; y++ {
		win.print(screen, 0, y, fillStyle, strings.Repeat(string(gSixelFiller), win.w))
	}

	// TODO: move logic into showSixel
	sxs.xprev, sxs.yprev = win.x+1, win.y+1
	sxs.sixel = reg.sixel
}
