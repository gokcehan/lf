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
	win      *win
	sixel    *string
	lastFile string // TODO maybe use hash of sixels instead
}

func (sxs *sixelScreen) clearSixels(screen tcell.Screen) {
	if sxs.sixel == nil {
		return
	}

	win := sxs.win
	for y := 0; y < win.h; y++ {
		win.print(screen, 0, y, tcell.StyleDefault, strings.Repeat(string(gSixelFiller), win.w))
	}

	sxs.sixel = nil
}

func (sxs *sixelScreen) setSixels(win *win, sixel *string) {
	sxs.win = win
	sxs.sixel = sixel
}

func (sxs *sixelScreen) showSixels(screen tcell.Screen, path string) {
	if sxs.sixel == nil {
		return
	}

	if path != sxs.lastFile {
		tmp := sxs.sixel
		sxs.clearSixels(screen)
		sxs.sixel = tmp
	}

	fmt.Fprint(os.Stderr, "\0337")                                  // Save cursor position
	fmt.Fprintf(os.Stderr, "\033[%d;%dH", sxs.win.y+1, sxs.win.x+1) // Move cursor to position
	fmt.Fprint(os.Stderr, *sxs.sixel)                               //
	fmt.Fprint(os.Stderr, "\0338")                                  // Restore cursor position

	sxs.lastFile = path
}
