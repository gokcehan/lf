package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

const (
	gEscapeCode         = 27
	gAnsiColorResetMask = termbox.AttrBold | termbox.AttrUnderline | termbox.AttrReverse
)

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

var gKeyVal = map[termbox.Key][]rune{
	termbox.KeyF1:             {'<', 'f', '-', '1', '>'},
	termbox.KeyF2:             {'<', 'f', '-', '2', '>'},
	termbox.KeyF3:             {'<', 'f', '-', '3', '>'},
	termbox.KeyF4:             {'<', 'f', '-', '4', '>'},
	termbox.KeyF5:             {'<', 'f', '-', '5', '>'},
	termbox.KeyF6:             {'<', 'f', '-', '6', '>'},
	termbox.KeyF7:             {'<', 'f', '-', '7', '>'},
	termbox.KeyF8:             {'<', 'f', '-', '8', '>'},
	termbox.KeyF9:             {'<', 'f', '-', '9', '>'},
	termbox.KeyF10:            {'<', 'f', '-', '1', '0', '>'},
	termbox.KeyF11:            {'<', 'f', '-', '1', '1', '>'},
	termbox.KeyF12:            {'<', 'f', '-', '1', '2', '>'},
	termbox.KeyInsert:         {'<', 'i', 'n', 's', 'e', 'r', 't', '>'},
	termbox.KeyDelete:         {'<', 'd', 'e', 'l', 'e', 't', 'e', '>'},
	termbox.KeyHome:           {'<', 'h', 'o', 'm', 'e', '>'},
	termbox.KeyEnd:            {'<', 'e', 'n', 'd', '>'},
	termbox.KeyPgup:           {'<', 'p', 'g', 'u', 'p', '>'},
	termbox.KeyPgdn:           {'<', 'p', 'g', 'd', 'n', '>'},
	termbox.KeyArrowUp:        {'<', 'u', 'p', '>'},
	termbox.KeyArrowDown:      {'<', 'd', 'o', 'w', 'n', '>'},
	termbox.KeyArrowLeft:      {'<', 'l', 'e', 'f', 't', '>'},
	termbox.KeyArrowRight:     {'<', 'r', 'i', 'g', 'h', 't', '>'},
	termbox.KeyCtrlSpace:      {'<', 'c', '-', 's', 'p', 'a', 'c', 'e', '>'},
	termbox.KeyCtrlA:          {'<', 'c', '-', 'a', '>'},
	termbox.KeyCtrlB:          {'<', 'c', '-', 'b', '>'},
	termbox.KeyCtrlC:          {'<', 'c', '-', 'c', '>'},
	termbox.KeyCtrlD:          {'<', 'c', '-', 'd', '>'},
	termbox.KeyCtrlE:          {'<', 'c', '-', 'e', '>'},
	termbox.KeyCtrlF:          {'<', 'c', '-', 'f', '>'},
	termbox.KeyCtrlG:          {'<', 'c', '-', 'g', '>'},
	termbox.KeyBackspace:      {'<', 'b', 's', '>'},
	termbox.KeyTab:            {'<', 't', 'a', 'b', '>'},
	termbox.KeyCtrlJ:          {'<', 'c', '-', 'j', '>'},
	termbox.KeyCtrlK:          {'<', 'c', '-', 'k', '>'},
	termbox.KeyCtrlL:          {'<', 'c', '-', 'l', '>'},
	termbox.KeyEnter:          {'<', 'e', 'n', 't', 'e', 'r', '>'},
	termbox.KeyCtrlN:          {'<', 'c', '-', 'n', '>'},
	termbox.KeyCtrlO:          {'<', 'c', '-', 'o', '>'},
	termbox.KeyCtrlP:          {'<', 'c', '-', 'p', '>'},
	termbox.KeyCtrlQ:          {'<', 'c', '-', 'q', '>'},
	termbox.KeyCtrlR:          {'<', 'c', '-', 'r', '>'},
	termbox.KeyCtrlS:          {'<', 'c', '-', 's', '>'},
	termbox.KeyCtrlT:          {'<', 'c', '-', 't', '>'},
	termbox.KeyCtrlU:          {'<', 'c', '-', 'u', '>'},
	termbox.KeyCtrlV:          {'<', 'c', '-', 'v', '>'},
	termbox.KeyCtrlW:          {'<', 'c', '-', 'w', '>'},
	termbox.KeyCtrlX:          {'<', 'c', '-', 'x', '>'},
	termbox.KeyCtrlY:          {'<', 'c', '-', 'y', '>'},
	termbox.KeyCtrlZ:          {'<', 'c', '-', 'z', '>'},
	termbox.KeyEsc:            {'<', 'e', 's', 'c', '>'},
	termbox.KeyCtrlBackslash:  {'<', 'c', '-', '\\', '>'},
	termbox.KeyCtrlRsqBracket: {'<', 'c', '-', ']', '>'},
	termbox.KeyCtrl6:          {'<', 'c', '-', '6', '>'},
	termbox.KeyCtrlSlash:      {'<', 'c', '-', '/', '>'},
	termbox.KeySpace:          {'<', 's', 'p', 'a', 'c', 'e', '>'},
	termbox.KeyBackspace2:     {'<', 'b', 's', '2', '>'},
}

var gValKey map[string]termbox.Key

func init() {
	gValKey = make(map[string]termbox.Key)
	for k, v := range gKeyVal {
		gValKey[string(v)] = k
	}
}

type win struct {
	w, h, x, y int
}

func newWin(w, h, x, y int) *win {
	return &win{w, h, x, y}
}

func (win *win) renew(w, h, x, y int) {
	win.w, win.h, win.x, win.y = w, h, x, y
}

func applyAnsiCodes(s string, fg, bg termbox.Attribute) (termbox.Attribute, termbox.Attribute) {
	toks := strings.Split(s, ";")

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

	// Parse 256 color terminal ansi codes
	// termbox-go has a color offset of one, because of attributes
	if len(nums) == 3 {
		if nums[0] == 48 && nums[1] == 5 {
			bg = termbox.Attribute(nums[2])
			bg++
		}
		if nums[0] == 38 && nums[1] == 5 {
			fg = termbox.Attribute(nums[2])
			fg++
		}
		return fg, bg
	}

	for _, n := range nums {
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

func printLength(s string) int {
	ind := 0
	off := 0
	for i := 0; i < len(s); i++ {
		r, w := utf8.DecodeRuneInString(s[i:])

		if r == gEscapeCode && i+1 < len(s) && s[i+1] == '[' {
			j := strings.IndexByte(s[i:min(len(s), i+32)], 'm')
			if j == -1 {
				continue
			}

			i += j
			continue
		}

		i += w - 1

		if r == '\t' {
			ind += gOpts.tabstop - (ind-off)%gOpts.tabstop
		} else {
			ind += runewidth.RuneWidth(r)
		}
	}

	return ind
}

func (win *win) print(x, y int, fg, bg termbox.Attribute, s string) (termbox.Attribute, termbox.Attribute) {
	off := x
	for i := 0; i < len(s); i++ {
		r, w := utf8.DecodeRuneInString(s[i:])

		if r == gEscapeCode && i+1 < len(s) && s[i+1] == '[' {
			j := strings.IndexByte(s[i:min(len(s), i+32)], 'm')
			if j == -1 {
				continue
			}

			fg, bg = applyAnsiCodes(s[i+2:i+j], fg, bg)

			i += j
			continue
		}

		if x < win.w {
			termbox.SetCell(win.x+x, win.y+y, r, fg, bg)
		}

		i += w - 1

		if r == '\t' {
			x += gOpts.tabstop - (x-off)%gOpts.tabstop
		} else {
			x += runewidth.RuneWidth(r)
		}
	}

	return fg, bg
}

func (win *win) printf(x, y int, fg, bg termbox.Attribute, format string, a ...interface{}) {
	win.print(x, y, fg, bg, fmt.Sprintf(format, a...))
}

func (win *win) printLine(x, y int, fg, bg termbox.Attribute, s string) {
	win.printf(x, y, fg, bg, "%s%*s", s, win.w-len(s), "")
}

func (win *win) printRight(y int, fg, bg termbox.Attribute, s string) {
	win.print(win.w-len(s), y, fg, bg, s)
}

func (win *win) printReg(reg *reg) {
	if reg == nil {
		return
	}

	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	for i, l := range reg.lines {
		fg, bg = win.print(2, i, fg, bg, l)
	}

	return
}

func (win *win) printDir(dir *dir, marks map[string]int, saves map[string]bool) {
	if win.w < 3 {
		return
	}

	if dir == nil {
		return
	}

	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	if dir.loading {
		fg = termbox.AttrBold
		win.print(2, 0, fg, bg, "loading...")
		return
	}

	if len(dir.fi) == 0 {
		fg = termbox.AttrBold
		win.print(2, 0, fg, bg, "empty")
		return
	}

	maxind := len(dir.fi) - 1

	beg := max(dir.ind-dir.pos, 0)
	end := min(beg+win.h, maxind+1)

	for i, f := range dir.fi[beg:end] {
		switch {
		case f.linkState == working:
			fg = termbox.ColorCyan
			if f.Mode().IsDir() {
				fg |= termbox.AttrBold
			}
		case f.linkState == broken:
			fg = termbox.ColorMagenta
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

		if _, ok := marks[path]; ok {
			win.print(0, i, fg, termbox.ColorMagenta, " ")
		} else if copy, ok := saves[path]; ok {
			if copy {
				win.print(0, i, fg, termbox.ColorYellow, " ")
			} else {
				win.print(0, i, fg, termbox.ColorRed, " ")
			}
		}

		if i == dir.pos {
			fg |= termbox.AttrReverse
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
			for i := 0; i < win.w-2-w; i++ {
				s = append(s, ' ')
			}
		}

		var info string

		for _, s := range gOpts.info {
			switch s {
			case "size":
				if !(gOpts.dircounts && f.IsDir()) {
					info = fmt.Sprintf("%s %4s", info, humanize(f.Size()))
					continue
				}

				if f.count == -1 {
					d, err := os.Open(path)
					if err != nil {
						f.count = -2
					}

					names, err := d.Readdirnames(1000)
					d.Close()

					if names == nil && err != io.EOF {
						f.count = -2
					} else {
						f.count = len(names)
					}
				}

				switch {
				case f.count < 0:
					info = fmt.Sprintf("%s    ?", info)
				case f.count < 1000:
					info = fmt.Sprintf("%s %4d", info, f.count)
				default:
					info = fmt.Sprintf("%s 999+", info)
				}
			case "time":
				info = fmt.Sprintf("%s %12s", info, f.ModTime().Format("Jan _2 15:04"))
			default:
				log.Printf("unknown info type: %s", s)
			}
		}

		if len(info) > 0 && win.w > 2*len(info) {
			s = runeSliceWidthRange(s, 0, win.w-2-len(info))
			for _, r := range info {
				s = append(s, r)
			}
		}

		// TODO: add a trailing '~' to the name if cut

		win.print(1, i, fg, bg, string(s))
	}
}

type ui struct {
	wins        []*win
	promptWin   *win
	msgWin      *win
	menuWin     *win
	msg         string
	regPrev     *reg
	dirPrev     *dir
	keyChan     chan string
	evChan      chan termbox.Event
	menuBuf     *bytes.Buffer
	cmdPrefix   string
	cmdAccLeft  []rune
	cmdAccRight []rune
	cmdBuf      []rune
	keyAcc      []rune
	keyCount    []rune
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

func getWins() []*win {
	wtot, htot := termbox.Size()

	var wins []*win

	widths := getWidths(wtot)

	wacc := 0
	wlen := len(widths)
	for i := 0; i < wlen; i++ {
		wins = append(wins, newWin(widths[i], htot-2, wacc, 1))
		wacc += widths[i]
	}

	return wins
}

func newUI() *ui {
	wtot, htot := termbox.Size()

	evChan := make(chan termbox.Event)

	go func() {
		for {
			evChan <- termbox.PollEvent()
		}
	}()

	return &ui{
		wins:      getWins(),
		promptWin: newWin(wtot, 1, 0, 0),
		msgWin:    newWin(wtot, 1, 0, htot-1),
		menuWin:   newWin(wtot, 1, 0, htot-2),
		keyChan:   make(chan string, 1000),
		evChan:    evChan,
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

	ui.msgWin.renew(wtot, 1, 0, htot-1)
}

func (ui *ui) print(msg string) {
	ui.msg = msg
	log.Print(msg)
}

func (ui *ui) printf(format string, a ...interface{}) {
	ui.print(fmt.Sprintf(format, a...))
}

type reg struct {
	path  string
	lines []string
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
		ui.dirPrev = nav.loadDir(curr.path)
	} else if curr.Mode().IsRegular() {
		ui.regPrev = nav.loadReg(ui, curr.path)
	}
}

func (ui *ui) loadFileInfo(nav *nav) {
	curr, err := nav.currFile()
	if err != nil {
		return
	}

	ui.msg = fmt.Sprintf("%v %4s %v", curr.Mode(), humanize(curr.Size()), curr.ModTime().Format(gOpts.timefmt))
}

func (ui *ui) drawPromptLine(nav *nav) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	dir := nav.currDir()

	pwd := strings.Replace(dir.path, gUser.HomeDir, "~", -1)
	pwd = filepath.Clean(pwd)

	var fname string
	curr, err := nav.currFile()
	if err == nil {
		fname = filepath.Base(curr.path)
	}

	var prompt string

	prompt = strings.Replace(gOpts.promptfmt, "%u", gUser.Username, -1)
	prompt = strings.Replace(prompt, "%h", gHostname, -1)
	prompt = strings.Replace(prompt, "%f", fname, -1)

	if printLength(strings.Replace(prompt, "%w", pwd, -1)) > ui.promptWin.w {
		sep := string(filepath.Separator)
		names := strings.Split(pwd, sep)
		for i := range names {
			r, _ := utf8.DecodeRuneInString(names[i])
			names[i] = string(r)
			if printLength(strings.Replace(prompt, "%w", strings.Join(names, sep), -1)) <= ui.promptWin.w {
				break
			}
		}
		pwd = strings.Join(names, sep)
	}

	prompt = strings.Replace(prompt, "%w", pwd, -1)

	ui.promptWin.print(0, 0, fg, bg, prompt)
}

func (ui *ui) drawStatLine(nav *nav) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	currDir := nav.currDir()

	ui.msgWin.print(0, 0, fg, bg, ui.msg)

	tot := len(currDir.fi)
	ind := min(currDir.ind+1, tot)

	ruler := fmt.Sprintf("%d/%d", ind, tot)

	ui.msgWin.printRight(0, fg, bg, ruler)
}

func (ui *ui) draw(nav *nav) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	termbox.Clear(fg, bg)

	ui.drawPromptLine(nav)

	length := min(len(ui.wins), len(nav.dirs))
	woff := len(ui.wins) - length

	if gOpts.preview {
		length = min(len(ui.wins)-1, len(nav.dirs))
		woff = len(ui.wins) - 1 - length
	}

	doff := len(nav.dirs) - length
	for i := 0; i < length; i++ {
		ui.wins[woff+i].printDir(nav.dirs[doff+i], nav.marks, nav.saves)
	}

	switch ui.cmdPrefix {
	case "":
		ui.drawStatLine(nav)
		termbox.HideCursor()
	case ">":
		ui.msgWin.printLine(0, 0, fg, bg, ui.cmdPrefix)
		ui.msgWin.print(len(ui.cmdPrefix), 0, fg, bg, ui.msg)
		ui.msgWin.print(len(ui.cmdPrefix)+len(ui.msg), 0, fg, bg, string(ui.cmdAccLeft))
		ui.msgWin.print(len(ui.cmdPrefix)+len(ui.msg)+runeSliceWidth(ui.cmdAccLeft), 0, fg, bg, string(ui.cmdAccRight))
		termbox.SetCursor(ui.msgWin.x+len(ui.cmdPrefix)+len(ui.msg)+runeSliceWidth(ui.cmdAccLeft), ui.msgWin.y)
	default:
		ui.msgWin.printLine(0, 0, fg, bg, ui.cmdPrefix)
		ui.msgWin.print(len(ui.cmdPrefix), 0, fg, bg, string(ui.cmdAccLeft))
		ui.msgWin.print(len(ui.cmdPrefix)+runeSliceWidth(ui.cmdAccLeft), 0, fg, bg, string(ui.cmdAccRight))
		termbox.SetCursor(ui.msgWin.x+len(ui.cmdPrefix)+runeSliceWidth(ui.cmdAccLeft), ui.msgWin.y)
	}

	if gOpts.preview {
		f, err := nav.currFile()
		if err == nil {
			preview := ui.wins[len(ui.wins)-1]

			if f.IsDir() {
				preview.printDir(ui.dirPrev, nav.marks, nav.saves)
			} else if f.Mode().IsRegular() {
				preview.printReg(ui.regPrev)
			}
		}
	}

	if ui.menuBuf != nil {
		lines := strings.Split(ui.menuBuf.String(), "\n")

		lines = lines[:len(lines)-1]

		ui.menuWin.h = len(lines) - 1
		ui.menuWin.y = ui.wins[0].h - ui.menuWin.h

		ui.menuWin.printLine(0, 0, termbox.AttrBold, termbox.AttrBold, lines[0])
		for i, line := range lines[1:] {
			ui.menuWin.printLine(0, i+1, fg, bg, "")
			ui.menuWin.print(0, i+1, fg, bg, line)
		}
	}

	termbox.Flush()

	if ui.cmdPrefix == "" {
		// leave the cursor at the beginning of the current file for screen readers
		moveCursor(ui.wins[woff+length-1].y+nav.dirs[doff+length-1].pos+1, ui.wins[woff+length-1].x+1)
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
	case key := <-ui.keyChan:
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
					ui.printf("unknown key: %s", key)
				}
			}
		}

		return ev
	case ev := <-ui.evChan:
		return ev
	}
}

type multiExpr struct {
	expr  expr
	count int
}

func readCmdEvent(ch chan<- multiExpr, ev termbox.Event) {
	if ev.Ch != 0 {
		ch <- multiExpr{&callExpr{"cmd-insert", []string{string(ev.Ch)}}, 1}
	} else {
		val := gKeyVal[ev.Key]
		if expr, ok := gOpts.cmdkeys[string(val)]; ok {
			ch <- multiExpr{expr, 1}
		}
	}
}

func (ui *ui) readEvent(ch chan<- multiExpr, ev termbox.Event) {
	redraw := &callExpr{"redraw", nil}
	count := 1

	switch ev.Type {
	case termbox.EventKey:
		if ev.Ch != 0 {
			switch {
			case ev.Ch == '<':
				ui.keyAcc = append(ui.keyAcc, '<', 'l', 't', '>')
			case ev.Ch == '>':
				ui.keyAcc = append(ui.keyAcc, '<', 'g', 't', '>')
			case unicode.IsDigit(ev.Ch) && len(ui.keyAcc) == 0:
				ui.keyCount = append(ui.keyCount, ev.Ch)
			default:
				ui.keyAcc = append(ui.keyAcc, ev.Ch)
			}
		} else {
			val := gKeyVal[ev.Key]
			if string(val) == "<esc>" {
				ch <- multiExpr{redraw, 1}
				ui.keyAcc = nil
				ui.keyCount = nil
			}
			ui.keyAcc = append(ui.keyAcc, val...)
		}

		binds, ok := findBinds(gOpts.keys, string(ui.keyAcc))

		switch len(binds) {
		case 0:
			ui.printf("unknown mapping: %s", string(ui.keyAcc))
			ch <- multiExpr{redraw, 1}
			ui.keyAcc = nil
			ui.keyCount = nil
			ui.menuBuf = nil
		case 1:
			if ok {
				if len(ui.keyCount) > 0 {
					c, err := strconv.Atoi(string(ui.keyCount))
					if err != nil {
						log.Printf("converting command count: %s", err)
					}
					count = c
				} else {
					count = 1
				}
				expr := gOpts.keys[string(ui.keyAcc)]
				ch <- multiExpr{expr, count}
				ui.keyAcc = nil
				ui.keyCount = nil
			}
			if len(ui.keyAcc) > 0 {
				ui.menuBuf = listBinds(binds)
				ch <- multiExpr{redraw, 1}
			} else if ui.menuBuf != nil {
				ui.menuBuf = nil
			}
		default:
			if ok {
				// TODO: use a delay
				if len(ui.keyCount) > 0 {
					c, err := strconv.Atoi(string(ui.keyCount))
					if err != nil {
						log.Printf("converting command count: %s", err)
					}
					count = c
				} else {
					count = 1
				}
				expr := gOpts.keys[string(ui.keyAcc)]
				ch <- multiExpr{expr, count}
				ui.keyAcc = nil
				ui.keyCount = nil
			}
			if len(ui.keyAcc) > 0 {
				ui.menuBuf = listBinds(binds)
				ch <- multiExpr{redraw, 1}
			} else {
				ui.menuBuf = nil
			}
		}
	case termbox.EventResize:
		ch <- multiExpr{redraw, 1}
	default:
		// TODO: handle other events
	}
}

// This function is used to read expressions on the client side. Digits are
// interpreted as command counts but this is only done for digits preceding any
// non-digit characters (e.g. "42y2k" as 42 times "y2k").
func (ui *ui) readExpr() <-chan multiExpr {
	ch := make(chan multiExpr)

	go func() {
		ch <- multiExpr{&callExpr{"redraw", nil}, 1}

		for {
			ev := ui.pollEvent()

			if ui.cmdPrefix != "" && ev.Type == termbox.EventKey {
				readCmdEvent(ch, ev)
				continue
			}

			ui.readEvent(ch, ev)
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
