package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

const EscapeCode = 27

type Win struct {
	w int
	h int
	x int
	y int
}

func newWin(w, h, x, y int) *Win {
	return &Win{w, h, x, y}
}

func (win *Win) renew(w, h, x, y int) {
	win.w = w
	win.h = h
	win.x = x
	win.y = y
}

func (win *Win) print(x, y int, fg, bg termbox.Attribute, s string) {
	off := x
	for i := 0; i < len(s); i++ {
		r, w := utf8.DecodeRuneInString(s[i:])

		if r == EscapeCode {
			i++
			if s[i] == '[' {
				j := strings.IndexByte(s[i:], 'm')

				toks := strings.Split(s[i+1:i+j], ";")

				var nums []int
				for _, t := range toks {
					if t == "" {
						fg = termbox.ColorDefault
						bg = termbox.ColorDefault
						break
					}
					i, err := strconv.Atoi(t)
					if err != nil {
						log.Printf("converting escape code: %s", err)
						continue
					}
					nums = append(nums, i)
				}

				for _, n := range nums {
					if 30 <= n && n <= 37 {
						fg = termbox.ColorDefault
					}
					if 40 <= n && n <= 47 {
						bg = termbox.ColorDefault
					}
				}

				for _, n := range nums {
					switch n {
					case 1:
						fg = fg | termbox.AttrBold
					case 4:
						fg = fg | termbox.AttrUnderline
					case 7:
						fg = fg | termbox.AttrReverse
					case 30:
						fg = fg | termbox.ColorBlack
					case 31:
						fg = fg | termbox.ColorRed
					case 32:
						fg = fg | termbox.ColorGreen
					case 33:
						fg = fg | termbox.ColorYellow
					case 34:
						fg = fg | termbox.ColorBlue
					case 35:
						fg = fg | termbox.ColorMagenta
					case 36:
						fg = fg | termbox.ColorCyan
					case 37:
						fg = fg | termbox.ColorWhite
					case 40:
						bg = bg | termbox.ColorBlack
					case 41:
						bg = bg | termbox.ColorRed
					case 42:
						bg = bg | termbox.ColorGreen
					case 43:
						bg = bg | termbox.ColorYellow
					case 44:
						bg = bg | termbox.ColorBlue
					case 45:
						bg = bg | termbox.ColorMagenta
					case 46:
						bg = bg | termbox.ColorCyan
					case 47:
						bg = bg | termbox.ColorWhite
					}
				}

				i = i + j
				continue
			}
		}

		if x >= win.w {
			break
		}

		termbox.SetCell(win.x+x, win.y+y, r, fg, bg)

		i += w - 1

		if r == '\t' {
			x += gOpts.tabstop - (x-off)%gOpts.tabstop
		} else {
			x++
		}
	}
}

func (win *Win) printf(x, y int, fg, bg termbox.Attribute, format string, a ...interface{}) {
	win.print(x, y, fg, bg, fmt.Sprintf(format, a...))
}

func (win *Win) printl(x, y int, fg, bg termbox.Attribute, s string) {
	win.printf(x, y, fg, bg, "%s%*s", s, win.w-len(s), "")
}

func (win *Win) printd(dir *Dir, marks map[string]bool) {
	if win.w < 3 {
		return
	}

	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	if len(dir.fi) == 0 {
		fg = termbox.AttrBold
		win.print(0, 0, fg, bg, "empty")
		return
	}

	maxind := len(dir.fi) - 1

	beg := max(dir.ind-dir.pos, 0)
	end := min(beg+win.h, maxind+1)

	for i, f := range dir.fi[beg:end] {
		switch {
		case f.Mode().IsRegular():
			if f.Mode()&0111 != 0 {
				fg = termbox.AttrBold | termbox.ColorGreen
			} else {
				fg = termbox.ColorDefault
			}
		case f.Mode().IsDir():
			fg = termbox.AttrBold | termbox.ColorBlue
		case f.Mode()&os.ModeSymlink != 0:
			fg = termbox.ColorCyan
		case f.Mode()&os.ModeNamedPipe != 0:
			fg = termbox.ColorRed
		case f.Mode()&os.ModeSocket != 0:
			fg = termbox.ColorYellow
		case f.Mode()&os.ModeDevice != 0:
			fg = termbox.ColorWhite
		}

		path := filepath.Join(dir.path, f.Name())

		if marks[path] {
			win.print(0, i, fg, termbox.ColorMagenta, " ")
		}

		if i == dir.pos {
			fg = fg | termbox.AttrReverse
		}

		var s []byte

		s = append(s, ' ')

		s = append(s, f.Name()...)

		if len(s) > win.w-2 {
			s = s[:win.w-2]
		} else {
			s = append(s, make([]byte, win.w-2-len(s))...)
		}

		switch gOpts.showinfo {
		case "none":
			break
		case "size":
			if win.w > 8 {
				h := humanize(f.Size())
				s = append(s[:win.w-3-len(h)])
				s = append(s, ' ')
				s = append(s, h...)
			}
		case "time":
			if win.w > 24 {
				t := f.ModTime().Format("Jan _2 15:04")
				s = append(s[:win.w-3-len(t)])
				s = append(s, ' ')
				s = append(s, t...)
			}
		default:
			log.Printf("unknown showinfo type: %s", gOpts.showinfo)
		}

		// TODO: add a trailing '~' to the name if cut

		win.print(1, i, fg, bg, string(s))
	}
}

func (win *Win) printr(reg []string) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	for i, l := range reg {
		win.print(2, i, fg, bg, l)
	}

	return
}

type UI struct {
	wins    []*Win
	pwdwin  *Win
	msgwin  *Win
	menuwin *Win
	message string
	regprev []string
	dirprev *Dir
}

func getWidths(wtot int) []int {
	rsum := 0
	for _, rat := range gOpts.ratios {
		rsum += rat
	}

	wlen := len(gOpts.ratios)
	widths := make([]int, wlen)

	wsum := 0
	for i := 0; i < wlen-1; i++ {
		widths[i] = gOpts.ratios[i] * (wtot / rsum)
		wsum += widths[i]
	}
	widths[wlen-1] = wtot - wsum

	return widths
}

func newUI() *UI {
	wtot, htot := termbox.Size()

	var wins []*Win

	widths := getWidths(wtot)

	wacc := 0
	wlen := len(widths)
	for i := 0; i < wlen; i++ {
		wins = append(wins, newWin(widths[i], htot-2, wacc, 1))
		wacc += widths[i]
	}

	return &UI{
		wins:    wins,
		pwdwin:  newWin(wtot, 1, 0, 0),
		msgwin:  newWin(wtot, 1, 0, htot-1),
		menuwin: newWin(wtot, 1, 0, htot-2),
	}
}

func (ui *UI) renew() {
	termbox.Flush()

	wtot, htot := termbox.Size()

	widths := getWidths(wtot)

	wacc := 0
	wlen := len(widths)
	for i := 0; i < wlen; i++ {
		ui.wins[i].renew(widths[i], htot-2, wacc, 1)
		wacc += widths[i]
	}

	ui.msgwin.renew(wtot, 1, 0, htot-1)
}

func (ui *UI) loadFile(nav *Nav) {
	dir := nav.currDir()

	if len(dir.fi) == 0 {
		return
	}

	curr := nav.currFile()

	ui.message = fmt.Sprintf("%v %v %v", curr.Mode(), humanize(curr.Size()), curr.ModTime().Format(time.ANSIC))

	if !gOpts.preview {
		return
	}

	path := nav.currPath()

	if curr.IsDir() {
		dir := newDir(path)
		dir.load(nav.inds[path], nav.poss[path], nav.height, nav.names[path])
		ui.dirprev = dir
	} else if curr.Mode().IsRegular() {
		var reader io.Reader

		if len(gOpts.previewer) != 0 {
			cmd := exec.Command(gOpts.previewer, path, strconv.Itoa(nav.height))

			out, err := cmd.StdoutPipe()
			if err != nil {
				msg := fmt.Sprintf("previewing file: %s", err)
				ui.message = msg
				log.Print(msg)
			}

			if err := cmd.Start(); err != nil {
				msg := fmt.Sprintf("previewing file: %s", err)
				ui.message = msg
				log.Print(msg)
			}

			defer out.Close()
			go func() { defer cmd.Wait() }()
			reader = out
		} else {
			f, err := os.Open(path)
			if err != nil {
				msg := fmt.Sprintf("opening file: %s", err)
				ui.message = msg
				log.Print(msg)
			}

			defer f.Close()
			reader = f
		}

		ui.regprev = nil

		buf := bufio.NewScanner(reader)

		for i := 0; i < nav.height && buf.Scan(); i++ {
			for _, r := range buf.Text() {
				if unicode.IsSpace(r) {
					continue
				}
				if !unicode.IsPrint(r) && r != EscapeCode {
					ui.regprev = []string{"binary"}
					return
				}
			}
			ui.regprev = append(ui.regprev, buf.Text())
		}

		if buf.Err() != nil {
			msg := fmt.Sprintf("loading file: %s", buf.Err())
			ui.message = msg
			log.Print(msg)
		}
	}
}

func (ui *UI) clearMsg() {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	win := ui.msgwin
	win.printl(0, 0, fg, bg, "")
	termbox.SetCursor(win.x, win.y)
	termbox.Flush()
}

func (ui *UI) draw(nav *Nav) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	termbox.Clear(fg, bg)
	defer termbox.Flush()

	dir := nav.currDir()

	path := strings.Replace(dir.path, envHome, "~", -1)

	ui.pwdwin.printf(0, 0, termbox.AttrBold|termbox.ColorGreen, bg, "%s@%s", envUser, envHost)
	ui.pwdwin.printf(len(envUser)+len(envHost)+1, 0, fg, bg, ":")
	ui.pwdwin.printf(len(envUser)+len(envHost)+2, 0, termbox.AttrBold|termbox.ColorBlue, bg, "%s", path)

	length := min(len(ui.wins), len(nav.dirs))
	woff := len(ui.wins) - length

	if gOpts.preview {
		length = min(len(ui.wins)-1, len(nav.dirs))
		woff = len(ui.wins) - 1 - length
	}

	doff := len(nav.dirs) - length
	for i := 0; i < length; i++ {
		ui.wins[woff+i].printd(nav.dirs[doff+i], nav.marks)
	}

	defer ui.msgwin.print(0, 0, fg, bg, ui.message)

	if gOpts.preview {
		if len(dir.fi) == 0 {
			return
		}

		preview := ui.wins[len(ui.wins)-1]
		path := nav.currPath()

		f, err := os.Stat(path)
		if err != nil {
			msg := fmt.Sprintf("getting file information: %s", err)
			ui.message = msg
			log.Print(msg)
			return
		}

		if f.IsDir() {
			preview.printd(ui.dirprev, nav.marks)
		} else if f.Mode().IsRegular() {
			preview.printr(ui.regprev)
		}
	}
}

func findBinds(keys map[string]Expr, prefix string) (binds map[string]Expr, ok bool) {
	binds = make(map[string]Expr)
	for key, expr := range keys {
		if strings.HasPrefix(key, prefix) {
			binds[key] = expr
			if key == prefix {
				ok = true
			}
		}
	}
	return
}

func (ui *UI) getExpr(nav *Nav) (expr Expr, count int) {
	expr = &CallExpr{"renew", nil}
	count = 1

	var acc []rune
	var cnt []rune

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Ch != 0 {
				switch {
				case ev.Ch == '<':
					acc = append(acc, '<', 'l', 't', '>')
				case ev.Ch == '>':
					acc = append(acc, '<', 'g', 't', '>')
				// Interpret digits as command count but only do this for
				// digits preceding any non-digit characters
				// (e.g. "42y2k" as 42 times "y2k").
				case unicode.IsDigit(ev.Ch) && len(acc) == 0:
					cnt = append(cnt, ev.Ch)
				default:
					acc = append(acc, ev.Ch)
				}
			} else {
				switch ev.Key {
				case termbox.KeyF1:
					acc = append(acc, '<', 'f', '-', '1', '>')
				case termbox.KeyF2:
					acc = append(acc, '<', 'f', '-', '2', '>')
				case termbox.KeyF3:
					acc = append(acc, '<', 'f', '-', '3', '>')
				case termbox.KeyF4:
					acc = append(acc, '<', 'f', '-', '4', '>')
				case termbox.KeyF5:
					acc = append(acc, '<', 'f', '-', '5', '>')
				case termbox.KeyF6:
					acc = append(acc, '<', 'f', '-', '6', '>')
				case termbox.KeyF7:
					acc = append(acc, '<', 'f', '-', '7', '>')
				case termbox.KeyF8:
					acc = append(acc, '<', 'f', '-', '8', '>')
				case termbox.KeyF9:
					acc = append(acc, '<', 'f', '-', '9', '>')
				case termbox.KeyF10:
					acc = append(acc, '<', 'f', '-', '1', '0', '>')
				case termbox.KeyF11:
					acc = append(acc, '<', 'f', '-', '1', '1', '>')
				case termbox.KeyF12:
					acc = append(acc, '<', 'f', '-', '1', '2', '>')
				case termbox.KeyInsert:
					acc = append(acc, '<', 'i', 'n', 's', 'e', 'r', 't', '>')
				case termbox.KeyDelete:
					acc = append(acc, '<', 'd', 'e', 'l', 'e', 't', 'e', '>')
				case termbox.KeyHome:
					acc = append(acc, '<', 'h', 'o', 'm', 'e', '>')
				case termbox.KeyEnd:
					acc = append(acc, '<', 'e', 'n', 'd', '>')
				case termbox.KeyPgup:
					acc = append(acc, '<', 'p', 'g', 'u', 'p', '>')
				case termbox.KeyPgdn:
					acc = append(acc, '<', 'p', 'g', 'd', 'n', '>')
				case termbox.KeyArrowUp:
					acc = append(acc, '<', 'u', 'p', '>')
				case termbox.KeyArrowDown:
					acc = append(acc, '<', 'd', 'o', 'w', 'n', '>')
				case termbox.KeyArrowLeft:
					acc = append(acc, '<', 'l', 'e', 'f', 't', '>')
				case termbox.KeyArrowRight:
					acc = append(acc, '<', 'r', 'i', 'g', 'h', 't', '>')
				case termbox.KeyCtrlSpace: // also KeyCtrlTilde and KeyCtrl2
					acc = append(acc, '<', 'c', '-', 's', 'p', 'a', 'c', 'e', '>')
				case termbox.KeyCtrlA:
					acc = append(acc, '<', 'c', '-', 'a', '>')
				case termbox.KeyCtrlB:
					acc = append(acc, '<', 'c', '-', 'b', '>')
				case termbox.KeyCtrlC:
					acc = append(acc, '<', 'c', '-', 'c', '>')
				case termbox.KeyCtrlD:
					acc = append(acc, '<', 'c', '-', 'd', '>')
				case termbox.KeyCtrlE:
					acc = append(acc, '<', 'c', '-', 'e', '>')
				case termbox.KeyCtrlF:
					acc = append(acc, '<', 'c', '-', 'f', '>')
				case termbox.KeyCtrlG:
					acc = append(acc, '<', 'c', '-', 'g', '>')
				case termbox.KeyBackspace: // also KeyCtrlH
					acc = append(acc, '<', 'b', 's', '>')
				case termbox.KeyTab: // also KeyCtrlI
					acc = append(acc, '<', 't', 'a', 'b', '>')
				case termbox.KeyCtrlJ:
					acc = append(acc, '<', 'c', '-', 'j', '>')
				case termbox.KeyCtrlK:
					acc = append(acc, '<', 'c', '-', 'k', '>')
				case termbox.KeyCtrlL:
					acc = append(acc, '<', 'c', '-', 'l', '>')
				case termbox.KeyEnter: // also KeyCtrlM
					acc = append(acc, '<', 'e', 'n', 't', 'e', 'r', '>')
				case termbox.KeyCtrlN:
					acc = append(acc, '<', 'c', '-', 'n', '>')
				case termbox.KeyCtrlO:
					acc = append(acc, '<', 'c', '-', 'o', '>')
				case termbox.KeyCtrlP:
					acc = append(acc, '<', 'c', '-', 'p', '>')
				case termbox.KeyCtrlQ:
					acc = append(acc, '<', 'c', '-', 'q', '>')
				case termbox.KeyCtrlR:
					acc = append(acc, '<', 'c', '-', 'r', '>')
				case termbox.KeyCtrlS:
					acc = append(acc, '<', 'c', '-', 's', '>')
				case termbox.KeyCtrlT:
					acc = append(acc, '<', 'c', '-', 't', '>')
				case termbox.KeyCtrlU:
					acc = append(acc, '<', 'c', '-', 'u', '>')
				case termbox.KeyCtrlV:
					acc = append(acc, '<', 'c', '-', 'v', '>')
				case termbox.KeyCtrlW:
					acc = append(acc, '<', 'c', '-', 'w', '>')
				case termbox.KeyCtrlX:
					acc = append(acc, '<', 'c', '-', 'x', '>')
				case termbox.KeyCtrlY:
					acc = append(acc, '<', 'c', '-', 'y', '>')
				case termbox.KeyCtrlZ:
					acc = append(acc, '<', 'c', '-', 'z', '>')
				case termbox.KeyEsc: // also KeyCtrlLsqBracket and KeyCtrl3
					acc = nil
					return
				case termbox.KeyCtrlBackslash: // also KeyCtrl4
					acc = append(acc, '<', 'c', '-', '\\', '>')
				case termbox.KeyCtrlRsqBracket: // also KeyCtrl5
					acc = append(acc, '<', 'c', '-', ']', '>')
				case termbox.KeyCtrl6:
					acc = append(acc, '<', 'c', '-', '6', '>')
				case termbox.KeyCtrlSlash: // also KeyCtrlUnderscore and KeyCtrl7
					acc = append(acc, '<', 'c', '-', '/', '>')
				case termbox.KeySpace:
					acc = append(acc, '<', 's', 'p', 'a', 'c', 'e', '>')
				case termbox.KeyBackspace2: // also KeyCtrl8
					acc = append(acc, '<', 'b', 's', '2', '>')
				}
			}

			binds, ok := findBinds(gOpts.keys, string(acc))

			switch len(binds) {
			case 0:
				ui.message = fmt.Sprintf("unknown mapping: %s", string(acc))
				acc = nil
				return
			case 1:
				if ok {
					if len(cnt) > 0 {
						c, err := strconv.Atoi(string(cnt))
						if err != nil {
							log.Printf("converting command count: %s", err)
						}
						count = c
					}
					return gOpts.keys[string(acc)], count
				}
				ui.draw(nav)
				if len(acc) > 0 {
					ui.listBinds(binds)
				}
			default:
				if ok {
					// TODO: use a delay
					if len(cnt) > 0 {
						c, err := strconv.Atoi(string(cnt))
						if err != nil {
							log.Printf("converting command count: %s", err)
						}
						count = c
					}
					return gOpts.keys[string(acc)], count
				}
				ui.draw(nav)
				if len(acc) > 0 {
					ui.listBinds(binds)
				}
			}
		case termbox.EventResize:
			return
		default:
			// TODO: handle other events
		}
	}
}

func (ui *UI) prompt(nav *Nav, pref string) string {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	win := ui.msgwin

	win.printl(0, 0, fg, bg, pref)
	termbox.SetCursor(win.x+len(pref), win.y)
	defer termbox.HideCursor()
	termbox.Flush()

	var lacc []rune
	var racc []rune

	var buf []rune

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Ch != 0 {
				lacc = append(lacc, ev.Ch)
			} else {
				// TODO: rest of the keys
				switch ev.Key {
				case termbox.KeyEsc:
					return ""
				case termbox.KeySpace:
					lacc = append(lacc, ' ')
				case termbox.KeyTab:
					var matches []string
					if pref == ":" {
						matches, lacc = compCmd(lacc)
					} else {
						matches, lacc = compShell(lacc)
					}
					ui.draw(nav)
					if len(matches) > 1 {
						ui.listMatches(matches)
					}
				case termbox.KeyEnter, termbox.KeyCtrlJ:
					win.printl(0, 0, fg, bg, "")
					termbox.SetCursor(win.x, win.y)
					termbox.Flush()
					return string(append(lacc, racc...))
				case termbox.KeyBackspace, termbox.KeyBackspace2:
					if len(lacc) > 0 {
						lacc = lacc[:len(lacc)-1]
					}
				case termbox.KeyDelete, termbox.KeyCtrlD:
					if len(racc) > 0 {
						racc = racc[1:]
					}
				case termbox.KeyArrowLeft, termbox.KeyCtrlB:
					if len(lacc) > 0 {
						racc = append([]rune{lacc[len(lacc)-1]}, racc...)
						lacc = lacc[:len(lacc)-1]
					}
				case termbox.KeyArrowRight, termbox.KeyCtrlF:
					if len(racc) > 0 {
						lacc = append(lacc, racc[0])
						racc = racc[1:]
					}
				case termbox.KeyHome, termbox.KeyCtrlA:
					racc = append(lacc, racc...)
					lacc = nil
				case termbox.KeyEnd, termbox.KeyCtrlE:
					lacc = append(lacc, racc...)
					racc = nil
				case termbox.KeyCtrlK:
					if len(racc) > 0 {
						buf = racc
						racc = nil
					}
				case termbox.KeyCtrlU:
					if len(lacc) > 0 {
						buf = lacc
						lacc = nil
					}
				case termbox.KeyCtrlW:
					ind := strings.LastIndex(strings.TrimRight(string(lacc), " "), " ") + 1
					buf = lacc[ind:]
					lacc = lacc[:ind]
				case termbox.KeyCtrlY:
					lacc = append(lacc, buf...)
				case termbox.KeyCtrlT:
					if len(lacc) > 1 {
						lacc[len(lacc)-1], lacc[len(lacc)-2] = lacc[len(lacc)-2], lacc[len(lacc)-1]
					}
				}
			}

			win.printl(0, 0, fg, bg, pref)
			win.print(len(pref), 0, fg, bg, string(lacc))
			win.print(len(pref)+len(lacc), 0, fg, bg, string(racc))
			termbox.SetCursor(win.x+len(pref)+len(lacc), win.y)
			termbox.Flush()
		default:
			// TODO: handle other events
		}
	}
}

func (ui *UI) pause() {
	termbox.Close()
}

func (ui *UI) resume() {
	if err := termbox.Init(); err != nil {
		log.Fatalf("initializing termbox: %s", err)
	}
}

func (ui *UI) sync() {
	if err := termbox.Sync(); err != nil {
		log.Printf("syncing termbox: %s", err)
	}
}

func (ui *UI) showMenu(b *bytes.Buffer) {
	lines := strings.Split(b.String(), "\n")

	lines = lines[:len(lines)-1]

	ui.menuwin.h = len(lines) - 1
	ui.menuwin.y = ui.wins[0].h - ui.menuwin.h

	ui.menuwin.printl(0, 0, termbox.AttrBold, termbox.AttrBold, lines[0])
	for i, line := range lines[1:] {
		ui.menuwin.printl(0, i+1, termbox.ColorDefault, termbox.ColorDefault, "")
		ui.menuwin.print(0, i+1, termbox.ColorDefault, termbox.ColorDefault, line)
	}

	termbox.Flush()
}

func (ui *UI) listBinds(binds map[string]Expr) {
	t := new(tabwriter.Writer)
	b := new(bytes.Buffer)

	var keys []string
	for k := range binds {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	t.Init(b, 0, gOpts.tabstop, 2, '\t', 0)
	fmt.Fprintln(t, "keys\tcommand")
	for _, k := range keys {
		fmt.Fprintf(t, "%s\t%v\n", k, binds[k])
	}
	t.Flush()

	ui.showMenu(b)
}

func (ui *UI) listMatches(matches []string) {
	b := new(bytes.Buffer)

	wtot, _ := termbox.Size()

	wcol := 0
	for _, m := range matches {
		wcol = max(wcol, len(m))
	}
	wcol += gOpts.tabstop - wcol%gOpts.tabstop

	ncol := wtot / wcol

	b.WriteString("possible matches\n")
	for i := 0; i < len(matches); i++ {
		for j := 0; j < ncol && i < len(matches); i, j = i+1, j+1 {
			b.WriteString(fmt.Sprintf("%s%*s", matches[i], wcol-len(matches[i]), ""))
		}
		b.WriteByte('\n')
	}

	ui.showMenu(b)
}
