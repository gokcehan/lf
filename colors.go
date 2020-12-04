package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type styleMap map[string]tcell.Style

func parseStyles() styleMap {
	if env := os.Getenv("LS_COLORS"); env != "" {
		return parseStylesGNU(env)
	}

	if env := os.Getenv("LSCOLORS"); env != "" {
		return parseStylesBSD(env)
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

	return parseStylesGNU(strings.Join(defaultColors, ":"))
}

func applyAnsiCodes(s string, st tcell.Style) tcell.Style {
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

	// ECMA-48 details the standard
	// TODO: should we support turning off attributes?
	//    Probably because this is used for previewers too
	for i := 0; i < len(nums); i++ {
		n := nums[i]
		switch {
		case n == 0:
			st = tcell.StyleDefault
		case n == 1:
			st = st.Bold(true)
		case n == 2:
			st = st.Dim(true)
		case n == 4:
			st = st.Underline(true)
		case n == 5 || n == 6:
			st = st.Blink(true)
		case n == 7:
			st = st.Reverse(true)
		case n == 8:
			// TODO: tcell PR for proper conceal
			_, bg, _ := st.Decompose()
			st = st.Foreground(bg)
		case n == 9:
			st = st.StrikeThrough(true)
		case n >= 30 && n <= 37:
			st = st.Foreground(tcell.PaletteColor(n - 30))
		case n == 38:
			if i+3 <= len(nums) && nums[i+1] == 5 {
				st = st.Foreground(tcell.PaletteColor(nums[i+2]))
				i += 2
			} else if i+5 <= len(nums) && nums[i+1] == 2 {
				st = st.Foreground(tcell.NewRGBColor(
					int32(nums[i+2]),
					int32(nums[i+3]),
					int32(nums[i+4])))
				i += 4
			} else {
				log.Printf("unknown ansi code or incorrect form: %d", n)
			}
		case n >= 40 && n <= 47:
			st = st.Background(tcell.PaletteColor(n - 40))
		case n == 48:
			if i+3 <= len(nums) && nums[i+1] == 5 {
				st = st.Background(tcell.PaletteColor(nums[i+2]))
				i += 2
			} else if i+5 <= len(nums) && nums[i+1] == 2 {
				st = st.Background(tcell.NewRGBColor(
					int32(nums[i+2]),
					int32(nums[i+3]),
					int32(nums[i+4])))
				i += 4
			} else {
				log.Printf("unknown ansi code or incorrect form: %d", n)
			}
		default:
			log.Printf("unknown ansi code: %d", n)
		}
	}

	return st
}

// This function parses $LS_COLORS environment variable.
func parseStylesGNU(env string) styleMap {
	styles := make(styleMap)

	entries := strings.Split(env, ":")
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		pair := strings.Split(entry, "=")
		if len(pair) != 2 {
			log.Printf("invalid $LS_COLORS entry: %s", entry)
			return styles
		}
		key, val := pair[0], pair[1]
		styles[key] = applyAnsiCodes(val, tcell.StyleDefault)
	}

	return styles
}

// This function parses $LSCOLORS environment variable.
func parseStylesBSD(env string) styleMap {
	styles := make(styleMap)

	if len(env) != 22 {
		log.Printf("invalid $LSCOLORS variable: %s", env)
		return styles
	}

	colorNames := []string{"di", "ln", "so", "pi", "ex", "bd", "cd", "su", "sg", "tw", "ow"}

	getStyle := func(r1, r2 byte) tcell.Style {
		st := tcell.StyleDefault

		switch {
		case r1 == 'x':
			st = st.Foreground(tcell.ColorDefault)
		case 'A' <= r1 && r1 <= 'H':
			st = st.Foreground(tcell.Color(r1 - 'A')).Bold(true)
		case 'a' <= r1 && r1 <= 'h':
			st = st.Foreground(tcell.Color(r1 - 'a'))
		default:
			log.Printf("invalid $LSCOLORS entry: %c", r1)
			return tcell.StyleDefault
		}

		switch {
		case r2 == 'x':
			st = st.Background(tcell.ColorDefault)
		case 'a' <= r2 && r2 <= 'h':
			st = st.Background(tcell.Color(r2 - 'a'))
		default:
			log.Printf("invalid $LSCOLORS entry: %c", r2)
			return tcell.StyleDefault
		}

		return st
	}

	for i, key := range colorNames {
		styles[key] = getStyle(env[i*2], env[i*2+1])
	}

	return styles
}

func (cm styleMap) get(f *file) tcell.Style {
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
		return val
	}

	if val, ok := cm["fi"]; ok {
		return val
	}

	return tcell.StyleDefault
}
