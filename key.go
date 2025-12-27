package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v3"
)

var gKeyVal = map[tcell.Key]string{
	tcell.KeyEnter:     "<enter>",
	tcell.KeyBackspace: "<backspace>",
	tcell.KeyTab:       "<tab>",
	tcell.KeyBacktab:   "<backtab>",
	tcell.KeyEsc:       "<esc>",
	tcell.KeyDelete:    "<delete>",
	tcell.KeyInsert:    "<insert>",
	tcell.KeyUp:        "<up>",
	tcell.KeyDown:      "<down>",
	tcell.KeyLeft:      "<left>",
	tcell.KeyRight:     "<right>",
	tcell.KeyHome:      "<home>",
	tcell.KeyEnd:       "<end>",
	tcell.KeyUpLeft:    "<upleft>",
	tcell.KeyUpRight:   "<upright>",
	tcell.KeyDownLeft:  "<downleft>",
	tcell.KeyDownRight: "<downright>",
	tcell.KeyCenter:    "<center>",
	tcell.KeyPgDn:      "<pgdn>",
	tcell.KeyPgUp:      "<pgup>",
	tcell.KeyClear:     "<clear>",
	tcell.KeyExit:      "<exit>",
	tcell.KeyCancel:    "<cancel>",
	tcell.KeyPause:     "<pause>",
	tcell.KeyPrint:     "<print>",
	tcell.KeyF1:        "<f-1>",
	tcell.KeyF2:        "<f-2>",
	tcell.KeyF3:        "<f-3>",
	tcell.KeyF4:        "<f-4>",
	tcell.KeyF5:        "<f-5>",
	tcell.KeyF6:        "<f-6>",
	tcell.KeyF7:        "<f-7>",
	tcell.KeyF8:        "<f-8>",
	tcell.KeyF9:        "<f-9>",
	tcell.KeyF10:       "<f-10>",
	tcell.KeyF11:       "<f-11>",
	tcell.KeyF12:       "<f-12>",
	tcell.KeyF13:       "<f-13>",
	tcell.KeyF14:       "<f-14>",
	tcell.KeyF15:       "<f-15>",
	tcell.KeyF16:       "<f-16>",
	tcell.KeyF17:       "<f-17>",
	tcell.KeyF18:       "<f-18>",
	tcell.KeyF19:       "<f-19>",
	tcell.KeyF20:       "<f-20>",
	tcell.KeyF21:       "<f-21>",
	tcell.KeyF22:       "<f-22>",
	tcell.KeyF23:       "<f-23>",
	tcell.KeyF24:       "<f-24>",
	tcell.KeyF25:       "<f-25>",
	tcell.KeyF26:       "<f-26>",
	tcell.KeyF27:       "<f-27>",
	tcell.KeyF28:       "<f-28>",
	tcell.KeyF29:       "<f-29>",
	tcell.KeyF30:       "<f-30>",
	tcell.KeyF31:       "<f-31>",
	tcell.KeyF32:       "<f-32>",
	tcell.KeyF33:       "<f-33>",
	tcell.KeyF34:       "<f-34>",
	tcell.KeyF35:       "<f-35>",
	tcell.KeyF36:       "<f-36>",
	tcell.KeyF37:       "<f-37>",
	tcell.KeyF38:       "<f-38>",
	tcell.KeyF39:       "<f-39>",
	tcell.KeyF40:       "<f-40>",
	tcell.KeyF41:       "<f-41>",
	tcell.KeyF42:       "<f-42>",
	tcell.KeyF43:       "<f-43>",
	tcell.KeyF44:       "<f-44>",
	tcell.KeyF45:       "<f-45>",
	tcell.KeyF46:       "<f-46>",
	tcell.KeyF47:       "<f-47>",
	tcell.KeyF48:       "<f-48>",
	tcell.KeyF49:       "<f-49>",
	tcell.KeyF50:       "<f-50>",
	tcell.KeyF51:       "<f-51>",
	tcell.KeyF52:       "<f-52>",
	tcell.KeyF53:       "<f-53>",
	tcell.KeyF54:       "<f-54>",
	tcell.KeyF55:       "<f-55>",
	tcell.KeyF56:       "<f-56>",
	tcell.KeyF57:       "<f-57>",
	tcell.KeyF58:       "<f-58>",
	tcell.KeyF59:       "<f-59>",
	tcell.KeyF60:       "<f-60>",
	tcell.KeyF61:       "<f-61>",
	tcell.KeyF62:       "<f-62>",
	tcell.KeyF63:       "<f-63>",
	tcell.KeyF64:       "<f-64>",
	tcell.KeyCtrlA:     "<c-a>",
	tcell.KeyCtrlB:     "<c-b>",
	tcell.KeyCtrlC:     "<c-c>",
	tcell.KeyCtrlD:     "<c-d>",
	tcell.KeyCtrlE:     "<c-e>",
	tcell.KeyCtrlF:     "<c-f>",
	tcell.KeyCtrlG:     "<c-g>",
	tcell.KeyCtrlJ:     "<c-j>",
	tcell.KeyCtrlK:     "<c-k>",
	tcell.KeyCtrlL:     "<c-l>",
	tcell.KeyCtrlN:     "<c-n>",
	tcell.KeyCtrlO:     "<c-o>",
	tcell.KeyCtrlP:     "<c-p>",
	tcell.KeyCtrlQ:     "<c-q>",
	tcell.KeyCtrlR:     "<c-r>",
	tcell.KeyCtrlS:     "<c-s>",
	tcell.KeyCtrlT:     "<c-t>",
	tcell.KeyCtrlU:     "<c-u>",
	tcell.KeyCtrlV:     "<c-v>",
	tcell.KeyCtrlW:     "<c-w>",
	tcell.KeyCtrlX:     "<c-x>",
	tcell.KeyCtrlY:     "<c-y>",
	tcell.KeyCtrlZ:     "<c-z>",
}

var gValKey map[string]tcell.Key

func init() {
	gValKey = make(map[string]tcell.Key, len(gKeyVal))
	for k, v := range gKeyVal {
		gValKey[v] = k
	}
}

// for simplicity, assume there will only be one modifier (ctrl, shift or alt)
var reModKey = regexp.MustCompile(`<(c|s|a)-(.+)>`)

func wrapModifier(s string, mod string) string {
	s = strings.TrimPrefix(s, "<")
	s = strings.TrimSuffix(s, ">")
	return fmt.Sprintf("<%s-%s>", mod, s)
}

func addKeyModifier(s string, mod tcell.ModMask) string {
	if reModKey.MatchString(s) {
		return s
	}

	switch {
	case mod&tcell.ModCtrl != 0:
		return wrapModifier(s, "c")
	case mod&tcell.ModShift != 0:
		return wrapModifier(s, "s")
	case mod&tcell.ModAlt != 0:
		return wrapModifier(s, "a")
	default:
		return s
	}
}

func readKey(ev *tcell.EventKey) string {
	var s string
	if ev.Key() == tcell.KeyRune {
		switch ev.Str() {
		case "<":
			s = "<lt>"
		case ">":
			s = "<gt>"
		case " ":
			s = "<space>"
		default:
			s = ev.Str()
		}
	} else {
		s = gKeyVal[ev.Key()]
	}

	return addKeyModifier(s, ev.Modifiers())
}

func parseKeyModifier(s string) (tcell.ModMask, string) {
	matches := reModKey.FindStringSubmatch(s)
	if matches == nil {
		return tcell.ModNone, s
	}

	mod := tcell.ModNone
	switch matches[1] {
	case "c":
		mod = tcell.ModCtrl
	case "s":
		mod = tcell.ModShift
	case "a":
		mod = tcell.ModAlt
	}

	s = matches[2]
	if len(s) > 1 {
		s = "<" + s + ">"
	}

	return mod, s
}

func parseKey(s string) *tcell.EventKey {
	if key, ok := gValKey[s]; ok {
		return tcell.NewEventKey(key, "", tcell.ModNone)
	}

	mod, s := parseKeyModifier(s)

	k := tcell.KeyRune
	if key, ok := gValKey[s]; ok {
		k = key
		s = ""
	} else {
		switch s {
		case "<lt>":
			s = "<"
		case "<gt>":
			s = ">"
		case "<space>":
			s = " "
		}
	}

	return tcell.NewEventKey(k, s, mod)
}
