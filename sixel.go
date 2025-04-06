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
	lastSxW  int
	lastSxH  int
	lastWinW int
	lastWinH int
}

func (sxs *sixelScreen) unlockSixel(win *win, screen tcell.Screen, filePath string) {
	if filePath != "" && (filePath != sxs.lastFile || win.w != sxs.lastWinW || win.h != sxs.lastWinH) {
		screen.LockRegion(win.x, win.y, sxs.lastSxW, sxs.lastSxH, false)

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
	v, _ := tty.WindowSize()
	w, h := v.CellDimensions()
	matches := reSixelSize.FindStringSubmatch(*reg.sixel)
	if matches == nil {
		log.Printf("sixel dimensions cannot be looked up")
		return
	}
	iw, _ := strconv.Atoi(matches[1])
	ih, _ := strconv.Atoi(matches[2])

	// width and height are -1 to avoid showing half filled sixels
	screen.LockRegion(win.x, win.y, iw/w-1, ih/h-1, true)
	ti.TPuts(tty, ti.TGoto(win.x, win.y))
	ti.TPuts(tty, *reg.sixel)
	sxs.lastFile = reg.path
	sxs.lastSxW = iw
	sxs.lastSxH = ih
	sxs.lastWinW = win.w
	sxs.lastWinH = win.h
}
