package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

const (
	gSixelBegin = "\033P"
)

type sixelScreen struct {
	lastFile string
	lastWinW int
	lastWinH int
}

func (sxs *sixelScreen) clearSixel(win *win, screen tcell.Screen, filePath string) {
	if filePath != sxs.lastFile || win.w != sxs.lastWinW || win.h != sxs.lastWinH {
		screen.LockRegion(win.x, win.y, win.w, win.h, false)
	}

}
func (sxs *sixelScreen) printSixel(win *win, screen tcell.Screen, reg *reg) {

	if reg.path == sxs.lastFile && win.w == sxs.lastWinW && win.h == sxs.lastWinH {
		return
	}
	if reg.sixel == nil {
		sxs.lastFile = ""
		return
	}

	tty, ok := screen.Tty()
	if !ok {
		log.Printf("returning underlying tty failed during sixel render")
		return
	}

	// Get the terminfo for our current terminal
	ti, err := tcell.LookupTerminfo(os.Getenv("TERM"))
	if err != nil {
		log.Printf("terminal lookup failed during sixel render %s", err)
	}
	screen.LockRegion(win.x, win.y, win.w, win.h, true)

	// Move the cursor to our draw position
	ti.TPuts(tty, ti.TGoto(win.x, win.y))
	ti.TPuts(tty, *reg.sixel)
	sxs.lastFile = reg.path
	sxs.lastWinW = win.w
	sxs.lastWinH = win.h
}
