package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type styleMap struct {
	styles        map[string]tcell.Style
	useLinkTarget bool
}

func parseStyles() styleMap {
	sm := styleMap{
		styles:        make(map[string]tcell.Style),
		useLinkTarget: false,
	}

	// Default values from dircolors
	//
	// no*  NORMAL                 00
	// fi   FILE                   00
	// rs*  RESET                  0
	// di   DIR                    01;34
	// ln   LINK                   01;36
	// mh*  MULTIHARDLINK          00
	// pi   FIFO                   40;33
	// so   SOCK                   01;35
	// do*  DOOR                   01;35
	// bd   BLK                    40;33;01
	// cd   CHR                    40;33;01
	// or   ORPHAN                 40;31;01
	// mi*  MISSING                00
	// su   SETUID                 37;41
	// sg   SETGID                 30;43
	// ca*  CAPABILITY             30;41
	// tw   STICKY_OTHER_WRITABLE  30;42
	// ow   OTHER_WRITABLE         34;42
	// st   STICKY                 37;44
	// ex   EXEC                   01;32
	//
	// (Entries marked with * are not implemented in lf)

	// default values from dircolors with background colors removed
	defaultColors := []string{
		"fi=00",
		"di=01;34",
		"ln=01;36",
		"pi=33",
		"so=01;35",
		"bd=33;01",
		"cd=33;01",
		"or=31;01",
		"su=01;32",
		"sg=01;32",
		"tw=01;34",
		"ow=01;34",
		"st=01;34",
		"ex=01;32",
	}

	sm.parseGNU(strings.Join(defaultColors, ":"))

	if env := os.Getenv("LSCOLORS"); env != "" {
		sm.parseBSD(env)
	}

	if env := os.Getenv("LS_COLORS"); env != "" {
		sm.parseGNU(env)
	}

	if env := os.Getenv("LF_COLORS"); env != "" {
		sm.parseGNU(env)
	}

	for _, path := range gColorsPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			sm.parseFile(path)
		}
	}

	return sm
}

func parseColor(toks []string) (tcell.Color, int, error) {
	if len(toks) == 0 {
		return tcell.ColorDefault, 0, fmt.Errorf("invalid args: %v", toks)
	}

	if toks[0] == "5" && len(toks) >= 2 {
		n, err := strconv.Atoi(toks[1])
		if err != nil {
			return tcell.ColorDefault, 0, fmt.Errorf("invalid args: %v", toks)
		}

		return tcell.PaletteColor(n), 2, nil
	}

	if toks[0] == "2" && len(toks) >= 4 {
		r, err := strconv.Atoi(toks[1])
		if err != nil {
			return tcell.ColorDefault, 0, fmt.Errorf("invalid args: %v", toks)
		}

		g, err := strconv.Atoi(toks[2])
		if err != nil {
			return tcell.ColorDefault, 0, fmt.Errorf("invalid args: %v", toks)
		}

		b, err := strconv.Atoi(toks[3])
		if err != nil {
			return tcell.ColorDefault, 0, fmt.Errorf("invalid args: %v", toks)
		}

		return tcell.NewRGBColor(int32(r), int32(g), int32(b)), 4, nil
	}

	return tcell.ColorDefault, 0, fmt.Errorf("invalid args: %v", toks)
}

func applyAnsiCodes(s string, st tcell.Style) tcell.Style {
	toks := strings.Split(s, ";")

	// ECMA-48 details the standard
	tokslen := len(toks)

loop:
	for i := 0; i < tokslen; i++ {
		switch strings.TrimLeft(toks[i], "0") {
		case "":
			st = tcell.StyleDefault
		case "1":
			st = st.Bold(true)
		case "2":
			st = st.Dim(true)
		case "3":
			st = st.Italic(true)
		case "4:0":
			st = st.Underline(false)
		case "4", "4:1":
			st = st.Underline(true)
		case "4:2":
			st = st.Underline(tcell.UnderlineStyleDouble)
		case "4:3":
			st = st.Underline(tcell.UnderlineStyleCurly)
		case "4:4":
			st = st.Underline(tcell.UnderlineStyleDotted)
		case "4:5":
			st = st.Underline(tcell.UnderlineStyleDashed)
		case "5", "6":
			st = st.Blink(true)
		case "7":
			st = st.Reverse(true)
		case "8":
			// TODO: tcell PR for proper conceal
			_, bg, _ := st.Decompose()
			st = st.Foreground(bg)
		case "9":
			st = st.StrikeThrough(true)
		case "22":
			st = st.Bold(false).Dim(false)
		case "23":
			st = st.Italic(false)
		case "24":
			st = st.Underline(false)
		case "25":
			st = st.Blink(false)
		case "27":
			st = st.Reverse(false)
		case "29":
			st = st.StrikeThrough(false)
		case "30", "31", "32", "33", "34", "35", "36", "37":
			n, _ := strconv.Atoi(toks[i])
			st = st.Foreground(tcell.PaletteColor(n - 30))
		case "90", "91", "92", "93", "94", "95", "96", "97":
			n, _ := strconv.Atoi(toks[i])
			st = st.Foreground(tcell.PaletteColor(n - 82))
		case "38":
			color, offset, err := parseColor(toks[i+1:])
			if err != nil {
				log.Printf("error processing ansi code 38: %s", err)
				break loop
			}
			st = st.Foreground(color)
			i += offset
		case "40", "41", "42", "43", "44", "45", "46", "47":
			n, _ := strconv.Atoi(toks[i])
			st = st.Background(tcell.PaletteColor(n - 40))
		case "100", "101", "102", "103", "104", "105", "106", "107":
			n, _ := strconv.Atoi(toks[i])
			st = st.Background(tcell.PaletteColor(n - 92))
		case "48":
			color, offset, err := parseColor(toks[i+1:])
			if err != nil {
				log.Printf("error processing ansi code 48: %s", err)
				break loop
			}
			st = st.Background(color)
			i += offset
		case "58":
			color, offset, err := parseColor(toks[i+1:])
			if err != nil {
				log.Printf("error processing ansi code 58: %s", err)
				break loop
			}
			st = st.Underline(color)
			i += offset
		default:
			log.Printf("unknown ansi code: %s", toks[i])
		}
	}

	return st
}

func (sm styleMap) parseFile(path string) {
	log.Printf("reading file: %s", path)

	f, err := os.Open(path)
	if err != nil {
		log.Printf("opening colors file: %s", err)
		return
	}
	defer f.Close()

	pairs, err := readPairs(f)
	if err != nil {
		log.Printf("reading colors file: %s", err)
		return
	}

	for _, pair := range pairs {
		sm.parsePair(pair)
	}
}

// This function parses $LS_COLORS environment variable.
func (sm *styleMap) parseGNU(env string) {
	for _, entry := range strings.Split(env, ":") {
		if entry == "" {
			continue
		}

		pair := strings.Split(entry, "=")

		if len(pair) != 2 {
			log.Printf("invalid $LS_COLORS entry: %s", entry)
			return
		}

		sm.parsePair(pair)
	}
}

func (sm *styleMap) parsePair(pair []string) {
	key, val := pair[0], pair[1]

	key = replaceTilde(key)

	if filepath.IsAbs(key) {
		key = filepath.Clean(key)
	}

	if key == "ln" && val == "target" {
		sm.useLinkTarget = true
	}

	sm.styles[key] = applyAnsiCodes(val, tcell.StyleDefault)
}

// This function parses $LSCOLORS environment variable.
func (sm styleMap) parseBSD(env string) {
	if len(env) != 22 {
		log.Printf("invalid $LSCOLORS variable: %s", env)
		return
	}

	colorNames := []string{"di", "ln", "so", "pi", "ex", "bd", "cd", "su", "sg", "tw", "ow"}

	getStyle := func(r1, r2 byte) tcell.Style {
		st := tcell.StyleDefault

		switch {
		case r1 == 'x':
			st = st.Foreground(tcell.ColorDefault)
		case 'A' <= r1 && r1 <= 'H':
			st = st.Foreground(tcell.PaletteColor(int(r1 - 'A'))).Bold(true)
		case 'a' <= r1 && r1 <= 'h':
			st = st.Foreground(tcell.PaletteColor(int(r1 - 'a')))
		default:
			log.Printf("invalid $LSCOLORS entry: %c", r1)
			return tcell.StyleDefault
		}

		switch {
		case r2 == 'x':
			st = st.Background(tcell.ColorDefault)
		case 'a' <= r2 && r2 <= 'h':
			st = st.Background(tcell.PaletteColor(int(r2 - 'a')))
		default:
			log.Printf("invalid $LSCOLORS entry: %c", r2)
			return tcell.StyleDefault
		}

		return st
	}

	for i, key := range colorNames {
		sm.styles[key] = getStyle(env[i*2], env[i*2+1])
	}
}

func (sm styleMap) get(f *file) tcell.Style {
	if val, ok := sm.styles[f.path]; ok {
		return val
	}

	if f.IsDir() {
		if val, ok := sm.styles[f.Name()+"/"]; ok {
			return val
		}
	}

	var key string

	switch {
	case f.linkState == working && !sm.useLinkTarget:
		key = "ln"
	case f.linkState == broken:
		key = "or"
	case f.IsDir() && f.Mode()&os.ModeSticky != 0 && f.Mode()&0o002 != 0:
		key = "tw"
	case f.IsDir() && f.Mode()&0o002 != 0:
		key = "ow"
	case f.IsDir() && f.Mode()&os.ModeSticky != 0:
		key = "st"
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
	case f.Mode()&0o111 != 0:
		key = "ex"
	}

	if val, ok := sm.styles[key]; ok {
		return val
	}

	if val, ok := sm.styles[f.Name()+"*"]; ok {
		return val
	}

	if val, ok := sm.styles["*"+f.Name()]; ok {
		return val
	}

	if val, ok := sm.styles[filepath.Base(f.Name())+".*"]; ok {
		return val
	}

	if val, ok := sm.styles["*"+strings.ToLower(f.ext)]; ok {
		return val
	}

	if val, ok := sm.styles["fi"]; ok {
		return val
	}

	return tcell.StyleDefault
}
