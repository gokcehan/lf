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
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

const EscapeCode = 27

var gKeyVal = map[termbox.Key][]rune{
	termbox.KeyF1:             []rune{'<', 'f', '-', '1', '>'},
	termbox.KeyF2:             []rune{'<', 'f', '-', '2', '>'},
	termbox.KeyF3:             []rune{'<', 'f', '-', '3', '>'},
	termbox.KeyF4:             []rune{'<', 'f', '-', '4', '>'},
	termbox.KeyF5:             []rune{'<', 'f', '-', '5', '>'},
	termbox.KeyF6:             []rune{'<', 'f', '-', '6', '>'},
	termbox.KeyF7:             []rune{'<', 'f', '-', '7', '>'},
	termbox.KeyF8:             []rune{'<', 'f', '-', '8', '>'},
	termbox.KeyF9:             []rune{'<', 'f', '-', '9', '>'},
	termbox.KeyF10:            []rune{'<', 'f', '-', '1', '0', '>'},
	termbox.KeyF11:            []rune{'<', 'f', '-', '1', '1', '>'},
	termbox.KeyF12:            []rune{'<', 'f', '-', '1', '2', '>'},
	termbox.KeyInsert:         []rune{'<', 'i', 'n', 's', 'e', 'r', 't', '>'},
	termbox.KeyDelete:         []rune{'<', 'd', 'e', 'l', 'e', 't', 'e', '>'},
	termbox.KeyHome:           []rune{'<', 'h', 'o', 'm', 'e', '>'},
	termbox.KeyEnd:            []rune{'<', 'e', 'n', 'd', '>'},
	termbox.KeyPgup:           []rune{'<', 'p', 'g', 'u', 'p', '>'},
	termbox.KeyPgdn:           []rune{'<', 'p', 'g', 'd', 'n', '>'},
	termbox.KeyArrowUp:        []rune{'<', 'u', 'p', '>'},
	termbox.KeyArrowDown:      []rune{'<', 'd', 'o', 'w', 'n', '>'},
	termbox.KeyArrowLeft:      []rune{'<', 'l', 'e', 'f', 't', '>'},
	termbox.KeyArrowRight:     []rune{'<', 'r', 'i', 'g', 'h', 't', '>'},
	termbox.KeyCtrlSpace:      []rune{'<', 'c', '-', 's', 'p', 'a', 'c', 'e', '>'},
	termbox.KeyCtrlA:          []rune{'<', 'c', '-', 'a', '>'},
	termbox.KeyCtrlB:          []rune{'<', 'c', '-', 'b', '>'},
	termbox.KeyCtrlC:          []rune{'<', 'c', '-', 'c', '>'},
	termbox.KeyCtrlD:          []rune{'<', 'c', '-', 'd', '>'},
	termbox.KeyCtrlE:          []rune{'<', 'c', '-', 'e', '>'},
	termbox.KeyCtrlF:          []rune{'<', 'c', '-', 'f', '>'},
	termbox.KeyCtrlG:          []rune{'<', 'c', '-', 'g', '>'},
	termbox.KeyBackspace:      []rune{'<', 'b', 's', '>'},
	termbox.KeyTab:            []rune{'<', 't', 'a', 'b', '>'},
	termbox.KeyCtrlJ:          []rune{'<', 'c', '-', 'j', '>'},
	termbox.KeyCtrlK:          []rune{'<', 'c', '-', 'k', '>'},
	termbox.KeyCtrlL:          []rune{'<', 'c', '-', 'l', '>'},
	termbox.KeyEnter:          []rune{'<', 'e', 'n', 't', 'e', 'r', '>'},
	termbox.KeyCtrlN:          []rune{'<', 'c', '-', 'n', '>'},
	termbox.KeyCtrlO:          []rune{'<', 'c', '-', 'o', '>'},
	termbox.KeyCtrlP:          []rune{'<', 'c', '-', 'p', '>'},
	termbox.KeyCtrlQ:          []rune{'<', 'c', '-', 'q', '>'},
	termbox.KeyCtrlR:          []rune{'<', 'c', '-', 'r', '>'},
	termbox.KeyCtrlS:          []rune{'<', 'c', '-', 's', '>'},
	termbox.KeyCtrlT:          []rune{'<', 'c', '-', 't', '>'},
	termbox.KeyCtrlU:          []rune{'<', 'c', '-', 'u', '>'},
	termbox.KeyCtrlV:          []rune{'<', 'c', '-', 'v', '>'},
	termbox.KeyCtrlW:          []rune{'<', 'c', '-', 'w', '>'},
	termbox.KeyCtrlX:          []rune{'<', 'c', '-', 'x', '>'},
	termbox.KeyCtrlY:          []rune{'<', 'c', '-', 'y', '>'},
	termbox.KeyCtrlZ:          []rune{'<', 'c', '-', 'z', '>'},
	termbox.KeyEsc:            []rune{'<', 'e', 's', 'c', '>'},
	termbox.KeyCtrlBackslash:  []rune{'<', 'c', '-', '\\', '>'},
	termbox.KeyCtrlRsqBracket: []rune{'<', 'c', '-', ']', '>'},
	termbox.KeyCtrl6:          []rune{'<', 'c', '-', '6', '>'},
	termbox.KeyCtrlSlash:      []rune{'<', 'c', '-', '/', '>'},
	termbox.KeySpace:          []rune{'<', 's', 'p', 'a', 'c', 'e', '>'},
	termbox.KeyBackspace2:     []rune{'<', 'b', 's', '2', '>'},
}

var gValKey map[string]termbox.Key

func init() {
	gValKey = make(map[string]termbox.Key)
	for k, v := range gKeyVal {
		gValKey[string(v)] = k
	}
}

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

		if r == EscapeCode && i+1 < len(s) && s[i+1] == '[' {
			j := strings.IndexByte(s[i:min(len(s), i+8)], 'm')

			if j == -1 {
				continue
			}

			toks := strings.Split(s[i+2:i+j], ";")

			var nums []int
			for _, t := range toks {
				if t == "" {
					fg = termbox.ColorDefault
					bg = termbox.ColorDefault
					break
				}
				n, err := strconv.Atoi(t)
				if err != nil {
					log.Printf("converting escape code: %s", err)
					continue
				}
				nums = append(nums, n)
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

			i += j
			continue
		}

		if x >= win.w {
			break
		}

		termbox.SetCell(win.x+x, win.y+y, r, fg, bg)

		i += w - 1

		if r == '\t' {
			x += gOpts.tabstop - (x-off)%gOpts.tabstop
		} else {
			x += runeWidth(r)
		}
	}
}

func (win *Win) printf(x, y int, fg, bg termbox.Attribute, format string, a ...interface{}) {
	win.print(x, y, fg, bg, fmt.Sprintf(format, a...))
}

func (win *Win) printl(x, y int, fg, bg termbox.Attribute, s string) {
	win.printf(x, y, fg, bg, "%s%*s", s, win.w-len(s), "")
}

func (win *Win) printd(dir *Dir, marks, saves map[string]bool) {
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
		case f.LinkState != NotLink:
			if f.LinkState == Working {
				fg = termbox.ColorCyan
				if f.Mode().IsDir() {
					fg |= termbox.AttrBold
				}
			} else {
				fg = termbox.ColorMagenta
			}
		case f.Mode().IsRegular():
			if f.Mode()&0111 != 0 {
				fg = termbox.AttrBold | termbox.ColorGreen
			} else {
				fg = termbox.ColorDefault
			}
		case f.Mode().IsDir():
			fg = termbox.AttrBold | termbox.ColorBlue
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
		} else if copy, ok := saves[path]; ok {
			if copy {
				win.print(0, i, fg, termbox.ColorYellow, " ")
			} else {
				win.print(0, i, fg, termbox.ColorRed, " ")
			}
		}

		if i == dir.pos {
			fg = fg | termbox.AttrReverse
		}

		var s []rune

		s = append(s, ' ')

		for _, r := range f.Name() {
			s = append(s, r)
		}

		w := runeSliceWidth(s)
		if w > win.w-2 {
			s = runeSliceWidthRange(s, 0, win.w-2)
		} else {
			s = append(s, make([]rune, win.w-2-w)...)
		}

		switch gOpts.showinfo {
		case "none":
			break
		case "size":
			if win.w > 8 {
				h := humanize(f.Size())
				s = runeSliceWidthRange(s, 0, win.w-3-len(h))
				s = append(s, ' ')
				for _, r := range h {
					s = append(s, r)
				}
			}
		case "time":
			if win.w > 24 {
				t := f.ModTime().Format("Jan _2 15:04")
				s = runeSliceWidthRange(s, 0, win.w-3-len(t))
				s = append(s, ' ')
				for _, r := range t {
					s = append(s, r)
				}
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
	keysbuf []string
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

func (ui *UI) loadFileInfo(nav *Nav) {
	curr, err := nav.currFile()
	if err != nil {
		return
	}

	ui.message = fmt.Sprintf("%v %v %v", curr.Mode(), humanize(curr.Size()), curr.ModTime().Format(gOpts.timefmt))
}

func (ui *UI) loadFile(nav *Nav) {
	curr, err := nav.currFile()
	if err != nil {
		return
	}

	if !gOpts.preview {
		return
	}

	if curr.IsDir() {
		dir := newDir(curr.Path)
		dir.load(nav.inds[curr.Path], nav.poss[curr.Path], nav.height, nav.names[curr.Path])
		ui.dirprev = dir
	} else if curr.Mode().IsRegular() {
		var reader io.Reader

		if len(gOpts.previewer) != 0 {
			cmd := exec.Command(gOpts.previewer, curr.Path, strconv.Itoa(nav.height))

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

			defer cmd.Wait()
			defer out.Close()
			reader = out
		} else {
			f, err := os.Open(curr.Path)
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
				if r == 0 {
					ui.regprev = []string{"[1mbinary[0m"}
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
	termbox.Flush()
}

func (ui *UI) draw(nav *Nav) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	termbox.Clear(fg, bg)

	// leave the cursor at the beginning of the current file for screen readers
	var length, woff, doff int
	defer func() {
		fmt.Printf("[%d;%dH", ui.wins[woff+length-1].y+nav.dirs[doff+length-1].pos+1, ui.wins[woff+length-1].x+1)
	}()

	defer termbox.Flush()

	dir := nav.currDir()

	path := strings.Replace(dir.path, envHome, "~", -1)
	path = filepath.Clean(path)

	ui.pwdwin.printf(0, 0, termbox.AttrBold|termbox.ColorGreen, bg, "%s@%s", envUser, envHost)
	ui.pwdwin.printf(len(envUser)+len(envHost)+1, 0, fg, bg, ":")
	ui.pwdwin.printf(len(envUser)+len(envHost)+2, 0, termbox.AttrBold|termbox.ColorBlue, bg, "%s", path)

	length = min(len(ui.wins), len(nav.dirs))
	woff = len(ui.wins) - length

	if gOpts.preview {
		length = min(len(ui.wins)-1, len(nav.dirs))
		woff = len(ui.wins) - 1 - length
	}

	doff = len(nav.dirs) - length
	for i := 0; i < length; i++ {
		ui.wins[woff+i].printd(nav.dirs[doff+i], nav.marks, nav.saves)
	}

	defer ui.msgwin.print(0, 0, fg, bg, ui.message)

	if gOpts.preview {
		f, err := nav.currFile()
		if err != nil {
			return
		}

		preview := ui.wins[len(ui.wins)-1]

		if f.IsDir() {
			preview.printd(ui.dirprev, nav.marks, nav.saves)
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

func (ui *UI) pollEvent() termbox.Event {
	if len(ui.keysbuf) > 0 {
		ev := termbox.Event{Type: termbox.EventKey}
		keys := ui.keysbuf[0]
		if len(keys) == 1 {
			ev.Ch, _ = utf8.DecodeRuneInString(keys)
		} else {
			switch keys {
			case "<lt>":
				ev.Ch = '<'
			case "<gt>":
				ev.Ch = '>'
			default:
				if val, ok := gValKey[keys]; ok {
					ev.Key = val
				} else {
					ev.Key = termbox.KeyEsc
					msg := fmt.Sprintf("unknown key: %s", keys)
					ui.message = msg
					log.Print(msg)
				}
			}
		}
		ui.keysbuf = ui.keysbuf[1:]
		return ev
	}
	return termbox.PollEvent()
}

type MultiExpr struct {
	expr  Expr
	count int
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
		switch ev := ui.pollEvent(); ev.Type {
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
			win.print(len(pref)+runeSliceWidth(lacc), 0, fg, bg, string(racc))
			termbox.SetCursor(win.x+len(pref)+runeSliceWidth(lacc), win.y)
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
