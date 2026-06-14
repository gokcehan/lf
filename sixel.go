package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v3"
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

			line = sanitizePreview(line)
			screen.LockRegion(win.x, y, printLength(line), 1, true)
			fmt.Fprintf(&b, "\033[%d;%dH", y+1, win.x+1)
			b.WriteString(line)
			y++
			continue
		}

		matches := reSixelSize.FindStringSubmatch(line)
		if matches == nil {
			log.Print("sixel: failed to get image size")
			continue
		}

		iw, _ := strconv.Atoi(matches[1])
		ih, _ := strconv.Atoi(matches[2])

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

	// Clear the preview pane in tcell's buffer so old text from
	// a previous file doesn't linger around or behind the image.
	st := tcell.StyleDefault
	for row := range win.h {
		for col := range win.w {
			screen.SetContent(win.x+col, win.y+row, ' ', nil, st)
		}
	}

	// Also write clear-to-end-of-line for each row of the preview
	// pane directly to the terminal, so old text is erased even if
	// tcell's Show() doesn't fully clear it.
	var clearBuf bytes.Buffer
	for row := range win.h {
		fmt.Fprintf(&clearBuf, "\033[%d;%dH\033[0K", win.y+row+1, win.x+1)
	}
	clearStr := clearBuf.String()

	fmt.Fprint(os.Stderr, "\033[?2026h") // Begin synchronized update
	fmt.Fprint(os.Stderr, "\0337")       // Save cursor position
	screen.Show()                        // Flush tcell's cleared buffer
	fmt.Fprint(os.Stderr, clearStr)      // Clear terminal rows
	fmt.Fprint(os.Stderr, b.String())    // Write sixel + text data
	fmt.Fprint(os.Stderr, "\0338")       // Restore cursor position
	fmt.Fprint(os.Stderr, "\033[?2026l") // End synchronized update

	sxs.lastFile = reg.path
	sxs.lastWin = *win
	sxs.forceClear = false
}

func cellSize(screen tcell.Screen) (int, int, error) {
	tty, ok := screen.Tty()
	if !ok {
		return -1, -1, fmt.Errorf("failed to get tty")
	}

	ws, err := tty.WindowSize()
	if err != nil {
		return -1, -1, fmt.Errorf("failed to get window size: %w", err)
	}

	cw, ch := ws.CellDimensions()
	if cw <= 0 || ch <= 0 {
		// fallback for Windows Terminal
		return 10, 20, nil
	}

	return cw, ch, nil
}
