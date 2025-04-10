package main

import (
	"log"
	"os"
	"strconv"

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
	if sxs.lastFile != "" && (filePath != sxs.lastFile || win.w != sxs.lastWinW || win.h != sxs.lastWinH) {
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
	ti, err := tcell.LookupTerminfo(os.Getenv("TERM"))
	if err != nil {
		log.Printf("terminal lookup failed during sixel render %s", err)
		return
	}
	ws, err := tty.WindowSize()
	if err != nil {
		log.Printf("window size lookup failed during sixel render %s", err)
		return
	}
	cw, ch := ws.CellDimensions()
	matches := reSixelSize.FindStringSubmatch(*reg.sixel)
	if matches == nil {
		log.Printf("sixel dimensions cannot be looked up")
		return
	}
	iw, _ := strconv.Atoi(matches[1])
	ih, _ := strconv.Atoi(matches[2])

	// width and height are -1 to avoid showing half filled sixels
	screen.LockRegion(win.x, win.y, iw/cw-1, ih/ch-1, true)
	ti.TPuts(tty, ti.TGoto(win.x, win.y))
	ti.TPuts(tty, *reg.sixel)
	sxs.lastFile = reg.path
	sxs.lastWinW = win.w
	sxs.lastWinH = win.h
}
