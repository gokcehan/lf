package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

const (
	gSixelBegin = "\033P"

	// The filler character should be:
	// - rarely used: the filler is used to trick tcell into redrawing, if the
	//   filler character appears in the user's preview, that cell might not
	//   be cleaned up properly
	// - ideally renders as empty space: the filler alternates between bold
	//   and regular, using a non-space would look weird to the user.
	gSixelFiller = '\u2000'
)

type sixelScreen struct {
	xprev, yprev int
	sixel        *string
	altFill      bool
	lastFile     string // TODO maybe use hash of sixels instead to flip altFill
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

func (sxs *sixelScreen) showSixels() {
	if sxs.sixel == nil {
		return
	}

	// XXX: workaround for bug where quitting lf might leave the terminal in bold
	fmt.Fprint(os.Stderr, "\033[0m")

	fmt.Fprint(os.Stderr, "\0337")                              // Save cursor position
	fmt.Fprintf(os.Stderr, "\033[%d;%dH", sxs.yprev, sxs.xprev) // Move cursor to position
	fmt.Fprint(os.Stderr, *sxs.sixel)                           //
	fmt.Fprint(os.Stderr, "\0338")                              // Restore cursor position
}

func (sxs *sixelScreen) printSixel(win *win, screen tcell.Screen, reg *reg) {
	if reg.sixel == nil {
		return
	}

	// HACK: fillers are used to control when tcell redraws the region where a sixel image is drawn.
	// alternating between bold and regular is to clear the image before drawing a new one.
	st := sxs.fillerStyle(reg.path)
	for y := 0; y < win.h; y++ {
		st = win.print(screen, 0, y, st, strings.Repeat(string(gSixelFiller), win.w))
	}

	sxs.xprev, sxs.yprev = win.x+1, win.y+1
	sxs.sixel = reg.sixel
}
