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
	sm := make(styleMap)

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

func parseEscapeSequence(s string) tcell.Style {
	s = strings.TrimPrefix(s, "\033[")
	if i := strings.IndexByte(s, 'm'); i >= 0 {
		s = s[:i]
	}
	return applyAnsiCodes(s, tcell.StyleDefault)
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
		case n == 3:
			st = st.Italic(true)
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
		case n >= 90 && n <= 97:
			st = st.Foreground(tcell.PaletteColor(n - 82))
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
		case n >= 100 && n <= 107:
			st = st.Background(tcell.PaletteColor(n - 92))
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
		key, val := pair[0], pair[1]

		key = replaceTilde(key)

		if filepath.IsAbs(key) {
			key = filepath.Clean(key)
		}

		sm[key] = applyAnsiCodes(val, tcell.StyleDefault)
	}
}

// This function parses $LS_COLORS environment variable.
func (sm styleMap) parseGNU(env string) {
	for _, entry := range strings.Split(env, ":") {
		if entry == "" {
			continue
		}

		pair := strings.Split(entry, "=")

		if len(pair) != 2 {
			log.Printf("invalid $LS_COLORS entry: %s", entry)
			return
		}

		key, val := pair[0], pair[1]

		key = replaceTilde(key)

		if filepath.IsAbs(key) {
			key = filepath.Clean(key)
		}

		sm[key] = applyAnsiCodes(val, tcell.StyleDefault)
	}
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
		sm[key] = getStyle(env[i*2], env[i*2+1])
	}
}

func (sm styleMap) get(f *file) tcell.Style {
	if val, ok := sm[f.path]; ok {
		return val
	}

	if f.IsDir() {
		if val, ok := sm[f.Name()+"/"]; ok {
			return val
		}
	}

	var key string

	switch {
	case f.linkState == working:
		key = "ln"
	case f.linkState == broken:
		key = "or"
	case f.IsDir() && f.Mode()&os.ModeSticky != 0 && f.Mode()&0002 != 0:
		key = "tw"
	case f.IsDir() && f.Mode()&0002 != 0:
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
	case f.Mode()&0111 != 0:
		key = "ex"
	}

	if val, ok := sm[key]; ok {
		return val
	}

	if val, ok := sm[f.Name()+"*"]; ok {
		return val
	}

	if val, ok := sm["*"+f.Name()]; ok {
		return val
	}

	if val, ok := sm[filepath.Base(f.Name())+".*"]; ok {
		return val
	}

	if val, ok := sm["*"+strings.ToLower(f.ext)]; ok {
		return val
	}

	if val, ok := sm["fi"]; ok {
		return val
	}

	return tcell.StyleDefault
}
