package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

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

	if !reg.sixel {
		sxs.lastFile = ""
		return
	}

	cw, ch, err := cellSize(screen)
	if err != nil {
		log.Printf("sixel: %s", err)
		return
	}

	y := win.y
	var b strings.Builder

	for _, line := range reg.lines {
		if !strings.HasPrefix(line, "\033P") {
			if y >= win.y+win.h {
				break
			}

			screen.LockRegion(win.x, y, printLength(line), 1, true)
			fmt.Fprintf(&b, "\033[%d;%dH", y+1, win.x+1)
			b.WriteString(line)
			y += 1
			continue
		}

		matches := reSixelSize.FindStringSubmatch(line)
		if matches == nil {
			log.Printf("sixel: failed to get image size")
			continue
		}

		iw, _ := strconv.Atoi(matches[1])
		ih, _ := strconv.Atoi(matches[2])

		if os.Getenv("TMUX") != "" {
			// tmux rounds the image height up to a multiple of 6, so we
			// need to do the same to avoid overwriting the image, as tmux
			// would remove the image if we touched it.
			ih = (ih + 5) / 6 * 6
		}

		sw := (iw + cw - 1) / cw
		sh := (ih + ch - 1) / ch

		if y+sh-1 >= win.y+win.h {
			break
		}

		screen.LockRegion(win.x, y, sw, sh, true)
		fmt.Fprintf(&b, "\033[%d;%dH", y+1, win.x+1)
		b.WriteString(line)
		y += sh
	}

	fmt.Fprint(os.Stderr, "\033[?2026h") // Begin synchronized update
	fmt.Fprint(os.Stderr, "\0337")       // Save cursor position
	fmt.Fprint(os.Stderr, b.String())    // Write data
	fmt.Fprint(os.Stderr, "\0338")       // Restore cursor position
	fmt.Fprint(os.Stderr, "\033[?2026l") // End synchronized update

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
