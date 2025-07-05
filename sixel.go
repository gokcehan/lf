package main

import (
	"errors"
	"fmt"
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

	cw, ch, err := cellSize(screen)
	if err != nil {
		log.Printf("sixel: %s", err)
		return
	}

	matches := reSixelSize.FindStringSubmatch(*reg.sixel)
	if matches == nil {
		log.Printf("sixel: failed to get image size")
		return
	}
	iw, _ := strconv.Atoi(matches[1])
	ih, _ := strconv.Atoi(matches[2])

	if os.Getenv("TMUX") != "" {
		// tmux rounds the image height up to a multiple of 6, so we
		// need to do the same to avoid overwriting the image, as tmux
		// would remove the image if we touched it.
		ih = (ih + 5) / 6 * 6
	}

	screen.LockRegion(win.x, win.y, (iw+cw-1)/cw, (ih+ch-1)/ch, true)
	fmt.Fprint(os.Stderr, "\0337")                          // Save cursor position
	fmt.Fprintf(os.Stderr, "\033[%d;%dH", win.y+1, win.x+1) // Move cursor to position
	fmt.Fprint(os.Stderr, *reg.sixel)                       // Print sixel
	fmt.Fprint(os.Stderr, "\0338")                          // Restore cursor position

	sxs.lastFile = reg.path
	sxs.lastWin = *win
	sxs.forceClear = false
}

func cellSize(screen tcell.Screen) (int, int, error) {
	tty, ok := screen.Tty()
	if !ok {
		// fallback for Windows Terminal
		return 10, 20, nil
	}

	ws, err := tty.WindowSize()
	if err != nil {
		return -1, -1, fmt.Errorf("failed to get window size: %s", err)
	}

	cw, ch := ws.CellDimensions()
	if cw <= 0 || ch <= 0 {
		return -1, -1, errors.New("cell dimensions should be greater than 0")
	}

	return cw, ch, nil
}
