package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
)

func init() {
	if lscolors := os.Getenv("LSCOLORS"); lscolors != "" {
		lsColors.parseBSD(lscolors)
		gOpts.lscolors = true
	}
	if ls_colors := os.Getenv("LS_COLORS"); ls_colors != "" {
		lsColors.parseLinux(ls_colors)
		gOpts.lscolors = true
	}
}

type lsColorsEntry struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

type lsColorsT map[string]lsColorsEntry

var lsColors = make(lsColorsT)

// This function parses LS_COLORS environment variable
func (lsc lsColorsT) parseLinux(env string) {
	e := strings.Split(env, ":")
	for _, e := range e {
		i := strings.IndexRune(e, '=')
		if i >= 0 {
			key := e[:i]
			values := strings.Split(e[i+1:], ";")
			var fg, bg termbox.Attribute
			for _, a := range values {
				i, _ := strconv.Atoi(a) // strconv.Atoi() will returns 0 on error which is termbox.ColorDefault. No need to check the error.

				switch {
				case i == 1:
					fg = fg | termbox.AttrBold
				case i == 4:
					fg = fg | termbox.AttrUnderline
				case i == 5:
					// Flashing text
				case i == 7:
					fg = fg | termbox.AttrReverse
				case i == 8:
					// Concealed
				case 30 <= i && i <= 37:
					fg = fg | termbox.Attribute(i-29)
				case 40 <= i && i <= 47:
					bg = bg | termbox.Attribute(i-39)
				}
			}
			lsc[key] = lsColorsEntry{fg: fg, bg: bg}
		}
	}
}

// This function parses LSCOLORS variable. See http://www.manpages.info/freebsd/ls.1.html
func (lsc lsColorsT) parseBSD(env string) {
	if len(env) != 22 {
		log.Print("LSCOLORS variable invalid")
		return
	}

	unixLsColors := []string{"di", "so", "ln", "pi", "ex", "bd", "cd", "su", "sg", "tw", "ow"}
	unixColors := map[byte]termbox.Attribute{
		'a': termbox.ColorBlack,
		'b': termbox.ColorRed,
		'c': termbox.ColorGreen,
		'd': termbox.ColorYellow, // should be brown
		'e': termbox.ColorBlue,
		'f': termbox.ColorMagenta,
		'g': termbox.ColorCyan,
		'h': termbox.ColorWhite, // should be light grey
		'A': termbox.AttrBold | termbox.ColorBlack,
		'B': termbox.AttrBold | termbox.ColorRed,
		'C': termbox.AttrBold | termbox.ColorGreen,
		'D': termbox.AttrBold | termbox.ColorWhite, // brown
		'E': termbox.AttrBold | termbox.ColorBlue,
		'F': termbox.AttrBold | termbox.ColorMagenta,
		'G': termbox.AttrBold | termbox.ColorCyan,
		'H': termbox.AttrBold | termbox.ColorWhite, // light grey
	}

	getColor := func(r byte) termbox.Attribute {
		if color, ok := unixColors[r]; ok {
			return color
		} else {
			return termbox.ColorDefault
		}
	}

	for i, key := range unixLsColors {
		lsc[key] = lsColorsEntry{fg: getColor(env[i*2]), bg: getColor(env[i*2+1])}
	}
}

// This function returns foreground and background colors for given file
func (lsc lsColorsT) getColors(f *file) (termbox.Attribute, termbox.Attribute) {
	var key = ""

	switch {
	case f.Mode()&os.ModeSticky != 0:
		key = "st"
	case f.Mode()&os.ModeSetuid != 0:
		key = "su"
	case f.Mode()&os.ModeSetgid != 0:
		key = "sg"
	case f.IsDir():
		key = "di"
	case f.linkState == working:
		key = "ln"
	case f.Mode()&os.ModeNamedPipe != 0:
		key = "pi"
	case f.Mode()&os.ModeSocket != 0:
		key = "so"
	case f.Mode()&os.ModeCharDevice != 0:
		key = "cd"
	case f.Mode()&os.ModeDevice != 0:
		key = "bd"
	case f.linkState == broken:
		key = "or"
	case f.Mode().IsRegular() && f.Mode()&0111 != 0:
		key = "ex"
	default:
		if extI := strings.LastIndexByte(f.Name(), '.'); extI > 0 {
			key = "*" + f.Name()[extI:]
		}
	}

	if val, ok := lsc[key]; ok {
		return val.fg, val.bg
	} else {
		if val, ok = lsc["no"]; ok {
			return val.fg, val.bg
		} else {
			return termbox.ColorDefault, termbox.ColorDefault
		}
	}
}
