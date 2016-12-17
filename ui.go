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
	"unicode"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

const gEscapeCode = 27

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

type win struct {
	w int
	h int
	x int
	y int
}

func newWin(w, h, x, y int) *win {
	return &win{w, h, x, y}
}

func (win *win) renew(w, h, x, y int) {
	win.w = w
	win.h = h
	win.x = x
	win.y = y
}

func (win *win) print(x, y int, fg, bg termbox.Attribute, s string) {
	off := x
	for i := 0; i < len(s); i++ {
		r, w := utf8.DecodeRuneInString(s[i:])

		if r == gEscapeCode && i+1 < len(s) && s[i+1] == '[' {
			j := strings.IndexByte(s[i:min(len(s), i+32)], 'm')

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

func (win *win) printf(x, y int, fg, bg termbox.Attribute, format string, a ...interface{}) {
	win.print(x, y, fg, bg, fmt.Sprintf(format, a...))
}

func (win *win) printl(x, y int, fg, bg termbox.Attribute, s string) {
	win.printf(x, y, fg, bg, "%s%*s", s, win.w-len(s), "")
}

func (win *win) printd(dir *dir, marks, saves map[string]bool) {
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
		case f.LinkState != notLink:
			if f.LinkState == working {
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

func (win *win) printr(reg []string) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	for i, l := range reg {
		win.print(2, i, fg, bg, l)
	}

	return
}

type ui struct {
	wins    []*win
	pwdwin  *win
	msgwin  *win
	menuwin *win
	message string
	regprev []string
	dirprev *dir
	keychan chan string
	evschan chan termbox.Event
	cmdpref string
	cmdlacc []rune
	cmdracc []rune
	cmdbuf  []rune
	menubuf *bytes.Buffer
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

func newUI() *ui {
	wtot, htot := termbox.Size()

	var wins []*win

	widths := getWidths(wtot)

	wacc := 0
	wlen := len(widths)
	for i := 0; i < wlen; i++ {
		wins = append(wins, newWin(widths[i], htot-2, wacc, 1))
		wacc += widths[i]
	}

	key := make(chan string, 1000)
	evs := make(chan termbox.Event)

	go func() {
		for {
			evs <- termbox.PollEvent()
		}
	}()

	return &ui{
		wins:    wins,
		pwdwin:  newWin(wtot, 1, 0, 0),
		msgwin:  newWin(wtot, 1, 0, htot-1),
		menuwin: newWin(wtot, 1, 0, htot-2),
		keychan: key,
		evschan: evs,
	}
}

func (ui *ui) renew() {
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

func (ui *ui) loadFileInfo(nav *nav) {
	curr, err := nav.currFile()
	if err != nil {
		return
	}

	ui.message = fmt.Sprintf("%v %v %v", curr.Mode(), humanize(curr.Size()), curr.ModTime().Format(gOpts.timefmt))
}

func (ui *ui) loadFile(nav *nav) {
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

func (ui *ui) draw(nav *nav) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	termbox.Clear(fg, bg)

	dir := nav.currDir()

	path := strings.Replace(dir.path, envHome, "~", -1)
	path = filepath.Clean(path)

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
		ui.wins[woff+i].printd(nav.dirs[doff+i], nav.marks, nav.saves)
	}

	if ui.cmdpref != "" {
		ui.msgwin.printl(0, 0, fg, bg, ui.cmdpref)
		ui.msgwin.print(len(ui.cmdpref), 0, fg, bg, string(ui.cmdlacc))
		ui.msgwin.print(len(ui.cmdpref)+runeSliceWidth(ui.cmdlacc), 0, fg, bg, string(ui.cmdracc))
		termbox.SetCursor(ui.msgwin.x+len(ui.cmdpref)+runeSliceWidth(ui.cmdlacc), ui.msgwin.y)
	} else {
		ui.msgwin.print(0, 0, fg, bg, ui.message)
		termbox.HideCursor()
	}

	if gOpts.preview {
		f, err := nav.currFile()
		if err == nil {
			preview := ui.wins[len(ui.wins)-1]

			if f.IsDir() {
				preview.printd(ui.dirprev, nav.marks, nav.saves)
			} else if f.Mode().IsRegular() {
				preview.printr(ui.regprev)
			}
		}
	}

	if ui.menubuf != nil {
		lines := strings.Split(ui.menubuf.String(), "\n")

		lines = lines[:len(lines)-1]

		ui.menuwin.h = len(lines) - 1
		ui.menuwin.y = ui.wins[0].h - ui.menuwin.h

		ui.menuwin.printl(0, 0, termbox.AttrBold, termbox.AttrBold, lines[0])
		for i, line := range lines[1:] {
			ui.menuwin.printl(0, i+1, termbox.ColorDefault, termbox.ColorDefault, "")
			ui.menuwin.print(0, i+1, termbox.ColorDefault, termbox.ColorDefault, line)
		}
	}

	termbox.Flush()

	if ui.cmdpref == "" {
		// leave the cursor at the beginning of the current file for screen readers
		fmt.Printf("[%d;%dH", ui.wins[woff+length-1].y+nav.dirs[doff+length-1].pos+1, ui.wins[woff+length-1].x+1)
	}
}

func findBinds(keys map[string]expr, prefix string) (binds map[string]expr, ok bool) {
	binds = make(map[string]expr)
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

func listBinds(binds map[string]expr) *bytes.Buffer {
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

	return b
}

func (ui *ui) pollEvent() termbox.Event {
	select {
	case key := <-ui.keychan:
		ev := termbox.Event{Type: termbox.EventKey}

		if len(key) == 1 {
			ev.Ch, _ = utf8.DecodeRuneInString(key)
		} else {
			switch key {
			case "<lt>":
				ev.Ch = '<'
			case "<gt>":
				ev.Ch = '>'
			default:
				if val, ok := gValKey[key]; ok {
					ev.Key = val
				} else {
					ev.Key = termbox.KeyEsc
					msg := fmt.Sprintf("unknown key: %s", key)
					ui.message = msg
					log.Print(msg)
				}
			}
		}

		return ev
	case ev := <-ui.evschan:
		return ev
	}
}

type multiExpr struct {
	expr  expr
	count int
}

// This function is used to read expressions on the client side. Digits are
// interpreted as command counts but this is only done for digits preceding any
// non-digit characters (e.g. "42y2k" as 42 times "y2k").
func (ui *ui) readExpr() chan multiExpr {
	ch := make(chan multiExpr)

	renew := &callExpr{"renew", nil}
	count := 1

	var acc []rune
	var cnt []rune

	go func() {
		for {
			ev := ui.pollEvent()

			if ui.cmdpref != "" {
				switch ev.Type {
				case termbox.EventKey:
					if ev.Ch != 0 {
						ch <- multiExpr{&callExpr{"cmd-insert", []string{string(ev.Ch)}}, 1}
					} else {
						// TODO: rest of the keys
						switch ev.Key {
						case termbox.KeyEsc:
							ch <- multiExpr{&callExpr{"cmd-escape", nil}, 1}
						case termbox.KeySpace:
							ch <- multiExpr{&callExpr{"cmd-insert", []string{" "}}, 1}
						case termbox.KeyTab:
							ch <- multiExpr{&callExpr{"cmd-comp", nil}, 1}
						case termbox.KeyEnter, termbox.KeyCtrlJ:
							ch <- multiExpr{&callExpr{"cmd-enter", nil}, 1}
						case termbox.KeyBackspace, termbox.KeyBackspace2:
							ch <- multiExpr{&callExpr{"cmd-delete-back", nil}, 1}
						case termbox.KeyDelete, termbox.KeyCtrlD:
							ch <- multiExpr{&callExpr{"cmd-delete", nil}, 1}
						case termbox.KeyArrowLeft, termbox.KeyCtrlB:
							ch <- multiExpr{&callExpr{"cmd-left", nil}, 1}
						case termbox.KeyArrowRight, termbox.KeyCtrlF:
							ch <- multiExpr{&callExpr{"cmd-right", nil}, 1}
						case termbox.KeyHome, termbox.KeyCtrlA:
							ch <- multiExpr{&callExpr{"cmd-beg", nil}, 1}
						case termbox.KeyEnd, termbox.KeyCtrlE:
							ch <- multiExpr{&callExpr{"cmd-end", nil}, 1}
						case termbox.KeyCtrlK:
							ch <- multiExpr{&callExpr{"cmd-delete-end", nil}, 1}
						case termbox.KeyCtrlU:
							ch <- multiExpr{&callExpr{"cmd-delete-beg", nil}, 1}
						case termbox.KeyCtrlW:
							ch <- multiExpr{&callExpr{"cmd-delete-word", nil}, 1}
						case termbox.KeyCtrlY:
							ch <- multiExpr{&callExpr{"cmd-put", nil}, 1}
						case termbox.KeyCtrlT:
							ch <- multiExpr{&callExpr{"cmd-transpose", nil}, 1}
						}
					}
					continue
				}
			}

			switch ev.Type {
			case termbox.EventKey:
				if ev.Ch != 0 {
					switch {
					case ev.Ch == '<':
						acc = append(acc, '<', 'l', 't', '>')
					case ev.Ch == '>':
						acc = append(acc, '<', 'g', 't', '>')
					case unicode.IsDigit(ev.Ch) && len(acc) == 0:
						cnt = append(cnt, ev.Ch)
					default:
						acc = append(acc, ev.Ch)
					}
				} else {
					val := gKeyVal[ev.Key]
					if string(val) == "<esc>" {
						ch <- multiExpr{renew, 1}
						acc = nil
						cnt = nil
					}
					acc = append(acc, val...)
				}

				binds, ok := findBinds(gOpts.keys, string(acc))

				switch len(binds) {
				case 0:
					ui.message = fmt.Sprintf("unknown mapping: %s", string(acc))
					ch <- multiExpr{renew, 1}
					acc = nil
					cnt = nil
				case 1:
					if ok {
						if len(cnt) > 0 {
							c, err := strconv.Atoi(string(cnt))
							if err != nil {
								log.Printf("converting command count: %s", err)
							}
							count = c
						} else {
							count = 1
						}
						expr := gOpts.keys[string(acc)]
						ch <- multiExpr{expr, count}
						acc = nil
						cnt = nil
					}
					if len(acc) > 0 {
						ui.menubuf = listBinds(binds)
						ch <- multiExpr{renew, 1}
					} else if ui.menubuf != nil {
						ui.menubuf = nil
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
						} else {
							count = 1
						}
						expr := gOpts.keys[string(acc)]
						ch <- multiExpr{expr, count}
						acc = nil
						cnt = nil
					}
					if len(acc) > 0 {
						ui.menubuf = listBinds(binds)
						ch <- multiExpr{renew, 1}
					} else {
						ui.menubuf = nil
					}
				}
			case termbox.EventResize:
				ch <- multiExpr{renew, 1}
			default:
				// TODO: handle other events
			}
		}
	}()

	return ch
}

func (ui *ui) pause() {
	termbox.Close()
}

func (ui *ui) resume() {
	if err := termbox.Init(); err != nil {
		log.Fatalf("initializing termbox: %s", err)
	}
}

func (ui *ui) sync() {
	if err := termbox.Sync(); err != nil {
		log.Printf("syncing termbox: %s", err)
	}
}

func listMatches(matches []string) *bytes.Buffer {
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

	return b
}
