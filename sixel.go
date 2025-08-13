package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/gdamore/tcell/v2"
)

const gSixelBegin = "\033P"

type sixelScreen struct {
	lastFile   string
	lastWin    win
	forceClear bool
}

func loadSixel(reader *bufio.Reader) (string, error) {
	buffer := new(bytes.Buffer)
	inSixel := false
	shift := false
	last := '\000'

	// Sixels can start / end on the same line as regular text,
	// so we can't iterate over lines.
	//
	// Start of text (and every new line) needs to be shifted
	// over to the start of the box, otherwise it gets written
	// to x=0 of the terminal and not of the preview box.
	for {
		rune, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		if last == '\033' && rune == 'P' {
			// sixel start
			inSixel = true
		} else if last == '\033' && rune == '\\' {
			// sixel end, (possible) start of regular line
			inSixel = false
			shift = true
		} else if rune == '\n' && !inSixel {
			// start of line
			shift = true
		}

		buffer.WriteRune(rune)
		if shift {
			// for now use 0 as the shift position,
			// will be updated on every redraw since that can
			// change
			buffer.WriteString("\033[0G")
			shift = false
		}
		last = rune
	}
	return buffer.String(), nil
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

	xShift := fmt.Sprintf("\033[%dG", win.x+1)
	sixel := reAnsiShift.ReplaceAllString(*reg.sixel, xShift)

	screen.LockRegion(win.x, win.y, win.w, win.h, true)
	fmt.Fprint(os.Stderr, "\0337")                          // Save cursor position
	fmt.Fprintf(os.Stderr, "\033[%d;%dH", win.y+1, win.x+1) // Move cursor to position
	fmt.Fprint(os.Stderr, sixel)                            // Print sixel
	fmt.Fprint(os.Stderr, "\0338")                          // Restore cursor position

	sxs.lastFile = reg.path
	sxs.lastWin = *win
	sxs.forceClear = false
}
