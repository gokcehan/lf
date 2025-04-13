package main

import (
	"log"
	"os"
	"strconv"

	"github.com/gdamore/tcell/v2"
)

const gSixelBegin = "\033P"

type sixelScreen struct {
	lastFile   string
	lastWin    win
	forceClear bool
}

func (sxs *sixelScreen) clearSixel(win *win, screen tcell.Screen, filePath string) {
	if sxs.lastFile != "" && (filePath != sxs.lastFile || *win != sxs.lastWin || sxs.forceClear) {
		screen.LockRegion(sxs.lastWin.x, sxs.lastWin.y, sxs.lastWin.w, sxs.lastWin.h, false)
	}
}

func (sxs *sixelScreen) printSixel(win *win, screen tcell.Screen, reg *reg) {
	if reg.path == sxs.lastFile && *win == sxs.lastWin && !sxs.forceClear {
		return
	}

	if reg.sixel == nil {
		sxs.lastFile = ""
		return
	}

	ti, err := tcell.LookupTerminfo(os.Getenv("TERM"))
	if err != nil {
		log.Printf("sixel: failed to look up term into %s", err)
		return
	}

	tty, ok := screen.Tty()
	if !ok {
		log.Printf("sixel: failed to get tty")
		return
	}

	ws, err := tty.WindowSize()
	if err != nil {
		log.Printf("sixel: failed to get window size %s", err)
		return
	}
	cw, ch := ws.CellDimensions()

	matches := reSixelSize.FindStringSubmatch(*reg.sixel)
	if matches == nil {
		log.Printf("sixel: failed to get image size")
		return
	}
	iw, _ := strconv.Atoi(matches[1])
	ih, _ := strconv.Atoi(matches[2])

	screen.LockRegion(win.x, win.y, iw/cw, ih/ch, true)
	ti.TPuts(tty, ti.TGoto(win.x, win.y))
	ti.TPuts(tty, *reg.sixel)

	sxs.lastFile = reg.path
	sxs.lastWin = *win
	sxs.forceClear = false
}
