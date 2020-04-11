package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/doronbehar/termbox-go"
)

const gAnsiColorResetMask = termbox.AttrBold | termbox.AttrUnderline | termbox.AttrReverse

var gAnsiCodes = map[int]termbox.Attribute{
	0:  termbox.ColorDefault,
	1:  termbox.AttrBold,
	4:  termbox.AttrUnderline,
	7:  termbox.AttrReverse,
	30: termbox.ColorBlack,
	31: termbox.ColorRed,
	32: termbox.ColorGreen,
	33: termbox.ColorYellow,
	34: termbox.ColorBlue,
	35: termbox.ColorMagenta,
	36: termbox.ColorCyan,
	37: termbox.ColorWhite,
	40: termbox.ColorBlack,
	41: termbox.ColorRed,
	42: termbox.ColorGreen,
	43: termbox.ColorYellow,
	44: termbox.ColorBlue,
	45: termbox.ColorMagenta,
	46: termbox.ColorCyan,
	47: termbox.ColorWhite,
}

type colorEntry struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

type colorMap map[string]colorEntry

func parseColors() colorMap {
	if env := os.Getenv("LS_COLORS"); env != "" {
		return parseColorsGNU(env)
	}

	if env := os.Getenv("LSCOLORS"); env != "" {
		return parseColorsBSD(env)
	}

	// default values from dircolors with removed background colors
	defaultColors := []string{
		// "rs=0",
		"di=01;34",
		"ln=01;36",
		// "mh=00",
		"pi=33", // "pi=40;33",
		"so=01;35",
		"do=01;35",
		"bd=33;01", // "bd=40;33;01",
		"cd=33;01", // "cd=40;33;01",
		"or=31;01", // "or=40;31;01",
		// "mi=00",
		"su=01;32", // "su=37;41",
		"sg=01;32", // "sg=30;43",
		// "ca=30;41",
		"tw=01;34", // "tw=30;42",
		"ow=01;34", // "ow=34;42",
		"st=01;34", // "st=37;44",
		"ex=01;32",
	}

	return parseColorsGNU(strings.Join(defaultColors, ":"))
}

func applyAnsiCodes(s string, fg, bg termbox.Attribute) (termbox.Attribute, termbox.Attribute) {
	toks := strings.Split(s, ";")

	var nums []int
	for _, tok := range toks {
		if tok == "" {
			nums = append(nums, 0)
			continue
		}
		n, err := strconv.Atoi(tok)
		if err != nil {
			log.Printf("converting escape code: %s", err)
			continue
		}
		nums = append(nums, n)
	}

	for i := 0; i < len(nums); i++ {
		n := nums[i]
		switch {
		case n == 38 && i+2 < len(nums) && nums[i+1] == 5:
			fg &= gAnsiColorResetMask
			fg |= termbox.Attribute(nums[i+2] + 1)
			i += 2
			continue
		case n == 48 && i+2 < len(nums) && nums[i+1] == 5:
			bg = termbox.Attribute(nums[i+2] + 1)
			i += 2
			continue
		}
		attr, ok := gAnsiCodes[n]
		if !ok {
			log.Printf("unknown ansi code: %d", n)
			continue
		}
		switch {
		case n == 0:
			fg, bg = attr, attr
		case n == 1 || n == 4 || n == 7:
			fg |= attr
		case 30 <= n && n <= 37:
			fg &= gAnsiColorResetMask
			fg |= attr
		case 40 <= n && n <= 47:
			bg = attr
		}
	}

	return fg, bg
}

// This function parses $LS_COLORS environment variable.
func parseColorsGNU(env string) colorMap {
	colors := make(colorMap)

	entries := strings.Split(env, ":")
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		pair := strings.Split(entry, "=")
		if len(pair) != 2 {
			log.Printf("invalid $LS_COLORS entry: %s", entry)
			return colors
		}
		key, val := pair[0], pair[1]
		fg, bg := applyAnsiCodes(val, termbox.ColorDefault, termbox.ColorDefault)
		colors[key] = colorEntry{fg: fg, bg: bg}
	}

	return colors
}

// This function parses $LSCOLORS environment variable.
func parseColorsBSD(env string) colorMap {
	colors := make(colorMap)

	if len(env) != 22 {
		log.Printf("invalid $LSCOLORS variable: %s", env)
		return colors
	}

	colorNames := []string{"di", "ln", "so", "pi", "ex", "bd", "cd", "su", "sg", "tw", "ow"}
	colorCodes := map[byte]termbox.Attribute{
		'a': termbox.ColorBlack,
		'b': termbox.ColorRed,
		'c': termbox.ColorGreen,
		'd': termbox.ColorYellow, // brown
		'e': termbox.ColorBlue,
		'f': termbox.ColorMagenta,
		'g': termbox.ColorCyan,
		'h': termbox.ColorWhite, // light grey
		'A': termbox.AttrBold | termbox.ColorBlack,
		'B': termbox.AttrBold | termbox.ColorRed,
		'C': termbox.AttrBold | termbox.ColorGreen,
		'D': termbox.AttrBold | termbox.ColorYellow, // brown
		'E': termbox.AttrBold | termbox.ColorBlue,
		'F': termbox.AttrBold | termbox.ColorMagenta,
		'G': termbox.AttrBold | termbox.ColorCyan,
		'H': termbox.AttrBold | termbox.ColorWhite, // light grey
		'x': termbox.ColorDefault,
	}

	getColor := func(r byte) termbox.Attribute {
		if color, ok := colorCodes[r]; ok {
			return color
		}

		log.Printf("invalid $LSCOLORS entry: %c", r)
		return termbox.ColorDefault
	}

	for i, key := range colorNames {
		colors[key] = colorEntry{fg: getColor(env[i*2]), bg: getColor(env[i*2+1])}
	}

	return colors
}

func (cm colorMap) get(f *file) (termbox.Attribute, termbox.Attribute) {
	var key string

	switch {
	case f.linkState == working:
		key = "ln"
	case f.linkState == broken:
		key = "or"
	case f.IsDir() && f.Mode()&os.ModeSticky != 0 && f.Mode()&0002 != 0:
		key = "tw"
	case f.IsDir() && f.Mode()&os.ModeSticky != 0:
		key = "st"
	case f.IsDir() && f.Mode()&0002 != 0:
		key = "ow"
	case f.IsDir():
		key = "di"
	case f.Mode()&os.ModeNamedPipe != 0:
		key = "pi"
	case f.Mode()&os.ModeSocket != 0:
		key = "so"
	case f.Mode()&os.ModeCharDevice != 0:
		key = "cd"
	case f.Mode()&os.ModeDevice != 0:
		key = "bd"
	case f.Mode()&os.ModeSetuid != 0:
		key = "su"
	case f.Mode()&os.ModeSetgid != 0:
		key = "sg"
	case f.Mode().IsRegular() && f.Mode()&0111 != 0:
		key = "ex"
	default:
		key = "*" + filepath.Ext(f.Name())
	}

	if val, ok := cm[key]; ok {
		return val.fg, val.bg
	}

	if val, ok := cm["fi"]; ok {
		return val.fg, val.bg
	}

	return termbox.ColorDefault, termbox.ColorDefault
}
