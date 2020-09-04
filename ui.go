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
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

const gEscapeCode = 27

var gKeyVal = map[tcell.Key]string{
	tcell.KeyF1:            "<f-1>",
	tcell.KeyF2:            "<f-2>",
	tcell.KeyF3:            "<f-3>",
	tcell.KeyF4:            "<f-4>",
	tcell.KeyF5:            "<f-5>",
	tcell.KeyF6:            "<f-6>",
	tcell.KeyF7:            "<f-7>",
	tcell.KeyF8:            "<f-8>",
	tcell.KeyF9:            "<f-9>",
	tcell.KeyF10:           "<f-10>",
	tcell.KeyF11:           "<f-11>",
	tcell.KeyF12:           "<f-12>",
	tcell.KeyInsert:        "<insert>",
	tcell.KeyDelete:        "<delete>",
	tcell.KeyHome:          "<home>",
	tcell.KeyEnd:           "<end>",
	tcell.KeyPgUp:          "<pgup>",
	tcell.KeyPgDn:          "<pgdn>",
	tcell.KeyUp:            "<up>",
	tcell.KeyDown:          "<down>",
	tcell.KeyLeft:          "<left>",
	tcell.KeyRight:         "<right>",
	tcell.KeyCtrlSpace:     "<c-space>",
	tcell.KeyCtrlA:         "<c-a>",
	tcell.KeyCtrlB:         "<c-b>",
	tcell.KeyCtrlC:         "<c-c>",
	tcell.KeyCtrlD:         "<c-d>",
	tcell.KeyCtrlE:         "<c-e>",
	tcell.KeyCtrlF:         "<c-f>",
	tcell.KeyCtrlG:         "<c-g>",
	tcell.KeyBackspace:     "<bs>",
	tcell.KeyTab:           "<tab>",
	tcell.KeyCtrlJ:         "<c-j>",
	tcell.KeyCtrlK:         "<c-k>",
	tcell.KeyCtrlL:         "<c-l>",
	tcell.KeyEnter:         "<enter>",
	tcell.KeyCtrlN:         "<c-n>",
	tcell.KeyCtrlO:         "<c-o>",
	tcell.KeyCtrlP:         "<c-p>",
	tcell.KeyCtrlQ:         "<c-q>",
	tcell.KeyCtrlR:         "<c-r>",
	tcell.KeyCtrlS:         "<c-s>",
	tcell.KeyCtrlT:         "<c-t>",
	tcell.KeyCtrlU:         "<c-u>",
	tcell.KeyCtrlV:         "<c-v>",
	tcell.KeyCtrlW:         "<c-w>",
	tcell.KeyCtrlX:         "<c-x>",
	tcell.KeyCtrlY:         "<c-y>",
	tcell.KeyCtrlZ:         "<c-z>",
	tcell.KeyEsc:           "<esc>",
	tcell.KeyCtrlBackslash: "<c-\\>",
	tcell.KeyCtrlRightSq:   "<c-]>",
	// TODO: compatability only, they are discouraged
	tcell.KeyCtrlCarat:      "<c-6>",
	tcell.KeyCtrlUnderscore: "<c-/>",
	tcell.KeyBackspace2:     "<bs2>",
}

var gValKey map[string]tcell.Key

func init() {
	gValKey = make(map[string]tcell.Key)
	for k, v := range gKeyVal {
		gValKey[v] = k
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

func (win *win) print(screen tcell.Screen, x, y int, st tcell.Style, s string) tcell.Style {
	off := x
	for i := 0; i < len(s); i++ {
		r, w := utf8.DecodeRuneInString(s[i:])

		if r == gEscapeCode && i+1 < len(s) && s[i+1] == '[' {
			j := strings.IndexByte(s[i:min(len(s), i+32)], 'm')
			if j == -1 {
				continue
			}

			st = applyAnsiCodes(s[i+2:i+j], st)

			i += j
			continue
		}

		if x < win.w {
			screen.SetContent(win.x+x, win.y+y, r, nil, st)
		}

		i += w - 1

		if r == '\t' {
			s := gOpts.tabstop - (x-off)%gOpts.tabstop
			for i := 0; i < s && x+i < win.w; i++ {
				screen.SetContent(win.x+x+i, win.y+y, ' ', nil, st)
			}
			x += s
		} else {
			x += runewidth.RuneWidth(r)
		}
	}

	return st
}

func (win *win) printf(screen tcell.Screen, x, y int, st tcell.Style, format string, a ...interface{}) {
	win.print(screen, x, y, st, fmt.Sprintf(format, a...))
}

func (win *win) printLine(screen tcell.Screen, x, y int, st tcell.Style, s string) {
	win.printf(screen, x, y, st, "%s%*s", s, win.w-len(s), "")
}

func (win *win) printRight(screen tcell.Screen, y int, st tcell.Style, s string) {
	win.print(screen, win.w-len(s), y, st, s)
}

func (win *win) printReg(screen tcell.Screen, reg *reg) {
	if reg == nil {
		return
	}

	st := tcell.StyleDefault

	if reg.loading {
		st = st.Reverse(true)
		win.print(screen, 2, 0, st, "loading...")
		return
	}

	for i, l := range reg.lines {
		if i > win.h-1 {
			break
		}

		st = win.print(screen, 2, i, st, l)
	}
}

func fileInfo(f *file, d *dir) string {
	var info string

	path := filepath.Join(d.path, f.Name())

	for _, s := range gOpts.info {
		switch s {
		case "size":
			if !(gOpts.dircounts && f.IsDir()) {
				info = fmt.Sprintf("%s %4s", info, humanize(f.Size()))
				continue
			}

			if f.dirCount == -1 {
				d, err := os.Open(path)
				if err != nil {
					f.dirCount = -2
				}

				names, err := d.Readdirnames(1000)
				d.Close()

				if names == nil && err != io.EOF {
					f.dirCount = -2
				} else {
					f.dirCount = len(names)
				}
			}

			switch {
			case f.dirCount < 0:
				info = fmt.Sprintf("%s    ?", info)
			case f.dirCount < 1000:
				info = fmt.Sprintf("%s %4d", info, f.dirCount)
			default:
				info = fmt.Sprintf("%s 999+", info)
			}
		case "time":
			info = fmt.Sprintf("%s %12s", info, f.ModTime().Format("Jan _2 15:04"))
		case "atime":
			info = fmt.Sprintf("%s %12s", info, f.accessTime.Format("Jan _2 15:04"))
		case "ctime":
			info = fmt.Sprintf("%s %12s", info, f.changeTime.Format("Jan _2 15:04"))
		default:
			log.Printf("unknown info type: %s", s)
		}
	}

	return info
}

func (win *win) printDir(screen tcell.Screen, dir *dir, selections map[string]int, saves map[string]bool, colors styleMap, icons iconMap) {
	if win.w < 5 || dir == nil {
		return
	}

	if dir.loading {
		win.print(screen, 2, 0, tcell.StyleDefault.Reverse(true), "loading...")
		return
	}

	if dir.noPerm {
		win.print(screen, 2, 0, tcell.StyleDefault.Reverse(true), "permission denied")
		return
	}

	if len(dir.files) == 0 {
		win.print(screen, 2, 0, tcell.StyleDefault.Reverse(true), "empty")
		return
	}

	beg := max(dir.ind-dir.pos, 0)
	end := min(beg+win.h, len(dir.files))

	if beg > end {
		return
	}

	var lnwidth int
	var lnformat string

	if gOpts.number || gOpts.relativenumber {
		lnwidth = 1
		if gOpts.number && gOpts.relativenumber {
			lnwidth++
		}
		for j := 10; j < len(dir.files); j *= 10 {
			lnwidth++
		}
		lnformat = fmt.Sprintf("%%%d.d ", lnwidth)
	}

	for i, f := range dir.files[beg:end] {
		st := colors.get(f)

		if lnwidth > 0 {
			var ln string

			if gOpts.number && (!gOpts.relativenumber) {
				ln = fmt.Sprintf(lnformat, i+1+beg)
			} else if gOpts.relativenumber {
				switch {
				case i < dir.pos:
					ln = fmt.Sprintf(lnformat, dir.pos-i)
				case i > dir.pos:
					ln = fmt.Sprintf(lnformat, i-dir.pos)
				case gOpts.number:
					ln = fmt.Sprintf(fmt.Sprintf("%%%d.d ", lnwidth-1), i+1+beg)
				default:
					ln = ""
				}
			}

			win.print(screen, 0, i, st.Foreground(tcell.ColorOlive), ln)
		}

		path := filepath.Join(dir.path, f.Name())

		if _, ok := selections[path]; ok {
			win.print(screen, lnwidth, i, st.Background(tcell.ColorPurple), " ")
		} else if cp, ok := saves[path]; ok {
			if cp {
				win.print(screen, lnwidth, i, st.Background(tcell.ColorOlive), " ")
			} else {
				win.print(screen, lnwidth, i, st.Background(tcell.ColorMaroon), " ")
			}
		}

		if i == dir.pos {
			st = st.Reverse(true)
		}

		var s []rune

		s = append(s, ' ')

		if gOpts.icons {
			s = append(s, []rune(icons.get(f))...)
			s = append(s, ' ')
		}

		for _, r := range f.Name() {
			s = append(s, r)
		}

		w := runeSliceWidth(s)

		if w > win.w-3 {
			s = runeSliceWidthRange(s, 0, win.w-4)
			s = append(s, []rune(gOpts.truncatechar)...)
		} else {
			for i := 0; i < win.w-3-w; i++ {
				s = append(s, ' ')
			}
		}

		info := fileInfo(f, dir)

		if len(info) > 0 && win.w-2 > 2*len(info) {
			if win.w-2 > w+len(info) {
				s = runeSliceWidthRange(s, 0, win.w-3-len(info)-lnwidth)
			} else {
				s = runeSliceWidthRange(s, 0, win.w-4-len(info)-lnwidth)
				s = append(s, []rune(gOpts.truncatechar)...)
			}
			for _, r := range info {
				s = append(s, r)
			}
		}

		s = append(s, ' ')

		win.print(screen, lnwidth+1, i, st, string(s))
	}
}

type ui struct {
	screen      tcell.Screen
	wins        []*win
	promptWin   *win
	msgWin      *win
	menuWin     *win
	msg         string
	regPrev     *reg
	dirPrev     *dir
	exprChan    chan expr
	keyChan     chan string
	evChan      chan tcell.Event
	menuBuf     *bytes.Buffer
	cmdPrefix   string
	cmdAccLeft  []rune
	cmdAccRight []rune
	cmdYankBuf  []rune
	keyAcc      []rune
	keyCount    []rune
	styles      styleMap
	icons       iconMap
}

func getWidths(wtot int) []int {
	rsum := 0
	for _, r := range gOpts.ratios {
		rsum += r
	}

	wlen := len(gOpts.ratios)
	widths := make([]int, wlen)

	wsum := 0
	for i := 0; i < wlen-1; i++ {
		widths[i] = gOpts.ratios[i] * (wtot / rsum)
		wsum += widths[i]
	}
	widths[wlen-1] = wtot - wsum

	if gOpts.drawbox {
		widths[wlen-1]--
	}

	return widths
}

func getWins(screen tcell.Screen) []*win {
	wtot, htot := screen.Size()

	var wins []*win

	widths := getWidths(wtot)

	wacc := 0
	wlen := len(widths)
	for i := 0; i < wlen; i++ {
		if gOpts.drawbox {
			wins = append(wins, newWin(widths[i], htot-4, wacc+1, 2))
		} else {
			wins = append(wins, newWin(widths[i], htot-2, wacc, 1))
		}
		wacc += widths[i]
	}

	return wins
}

func newUI(screen tcell.Screen) *ui {
	wtot, htot := screen.Size()

	ui := &ui{
		screen:    screen,
		wins:      getWins(screen),
		promptWin: newWin(wtot, 1, 0, 0),
		msgWin:    newWin(wtot, 1, 0, htot-1),
		menuWin:   newWin(wtot, 1, 0, htot-2),
		keyChan:   make(chan string, 1000),
		evChan:    make(chan tcell.Event),
		styles:    parseStyles(),
		icons:     parseIcons(),
	}

	go func() {
		var ev tcell.Event
		for {
			ev = ui.screen.PollEvent()
			ui.evChan <- ev
		}
	}()

	return ui
}

func (ui *ui) renew() {
	wtot, htot := ui.screen.Size()

	widths := getWidths(wtot)

	wacc := 0
	wlen := len(widths)
	for i := 0; i < wlen; i++ {
		if gOpts.drawbox {
			ui.wins[i].renew(widths[i], htot-4, wacc+1, 2)
		} else {
			ui.wins[i].renew(widths[i], htot-2, wacc, 1)
		}
		wacc += widths[i]
	}

	ui.promptWin.renew(wtot, 1, 0, 0)
	ui.msgWin.renew(wtot, 1, 0, htot-1)
	ui.menuWin.renew(wtot, 1, 0, htot-2)
}

func (ui *ui) sort() {
	if ui.dirPrev == nil {
		return
	}
	name := ui.dirPrev.name()
	ui.dirPrev.sort()
	ui.dirPrev.sel(name, ui.wins[0].h)
}

func (ui *ui) echo(msg string) {
	ui.msg = msg
}

func (ui *ui) echof(format string, a ...interface{}) {
	ui.echo(fmt.Sprintf(format, a...))
}

func (ui *ui) echomsg(msg string) {
	ui.msg = msg
	log.Print(msg)
}

func (ui *ui) echomsgf(format string, a ...interface{}) {
	ui.echomsg(fmt.Sprintf(format, a...))
}

func (ui *ui) echoerr(msg string) {
	ui.msg = fmt.Sprintf(gOpts.errorfmt, msg)
	log.Printf("error: %s", msg)
}

func (ui *ui) echoerrf(format string, a ...interface{}) {
	ui.echoerr(fmt.Sprintf(format, a...))
}

type reg struct {
	loading  bool
	loadTime time.Time
	path     string
	lines    []string
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
		ui.regPrev = nav.loadReg(curr.path)
	}
}

func (ui *ui) loadFileInfo(nav *nav) {
	curr, err := nav.currFile()
	if err != nil {
		return
	}

	var linkTarget string
	if curr.linkTarget != "" {
		linkTarget = " -> " + curr.linkTarget
	}
	ui.echof("%v %4s %v%s", curr.Mode(), humanize(curr.Size()), curr.ModTime().Format(gOpts.timefmt), linkTarget)
}

func (ui *ui) drawPromptLine(nav *nav) {
	st := tcell.StyleDefault

	pwd := nav.currDir().path

	if strings.HasPrefix(pwd, gUser.HomeDir) {
		pwd = filepath.Join("~", strings.TrimPrefix(pwd, gUser.HomeDir))
	}

	sep := string(filepath.Separator)

	if !strings.HasSuffix(pwd, sep) {
		pwd += sep
	}

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
		names := strings.Split(pwd, sep)
		for i := range names {
			if names[i] == "" {
				continue
			}
			r, _ := utf8.DecodeRuneInString(names[i])
			names[i] = string(r)
			if printLength(strings.Replace(prompt, "%w", strings.Join(names, sep), -1)) <= ui.promptWin.w {
				break
			}
		}
		pwd = strings.Join(names, sep)
	}

	prompt = strings.Replace(prompt, "%w", pwd, -1)

	ui.promptWin.print(ui.screen, 0, 0, st, prompt)
}

func (ui *ui) drawStatLine(nav *nav) {
	st := tcell.StyleDefault

	dir := nav.currDir()

	ui.msgWin.print(ui.screen, 0, 0, st, ui.msg)

	tot := len(dir.files)
	ind := min(dir.ind+1, tot)
	acc := string(ui.keyCount) + string(ui.keyAcc)

	var progress string

	if nav.copyTotal > 0 {
		percentage := int((100 * float64(nav.copyBytes)) / float64(nav.copyTotal))
		progress += fmt.Sprintf("  [%d%%]", percentage)
	}

	if nav.moveTotal > 0 {
		progress += fmt.Sprintf("  [%d/%d]", nav.moveCount, nav.moveTotal)
	}

	if nav.deleteTotal > 0 {
		progress += fmt.Sprintf("  [%d/%d]", nav.deleteCount, nav.deleteTotal)
	}

	ruler := fmt.Sprintf("%s%s  %d/%d", acc, progress, ind, tot)

	ui.msgWin.printRight(ui.screen, 0, st, ruler)
}

func (ui *ui) drawBox() {
	st := tcell.StyleDefault

	w, h := ui.screen.Size()

	for i := 1; i < w-1; i++ {
		ui.screen.SetContent(i, 1, '─', nil, st)
		ui.screen.SetContent(i, h-2, '─', nil, st)
	}

	for i := 2; i < h-2; i++ {
		ui.screen.SetContent(0, i, '│', nil, st)
		ui.screen.SetContent(w-1, i, '│', nil, st)
	}

	ui.screen.SetContent(0, 1, '┌', nil, st)
	ui.screen.SetContent(w-1, 1, '┐', nil, st)
	ui.screen.SetContent(0, h-2, '└', nil, st)
	ui.screen.SetContent(w-1, h-2, '┘', nil, st)

	wacc := 0
	for wind := 0; wind < len(ui.wins)-1; wind++ {
		wacc += ui.wins[wind].w
		ui.screen.SetContent(wacc, 1, '┬', nil, st)
		for i := 2; i < h-2; i++ {
			ui.screen.SetContent(wacc, i, '│', nil, st)
		}
		ui.screen.SetContent(wacc, h-2, '┴', nil, st)
	}
}

func (ui *ui) draw(nav *nav) {
	st := tcell.StyleDefault

	wtot, htot := ui.screen.Size()
	for i := 0; i < wtot; i++ {
		for j := 0; j < htot; j++ {
			ui.screen.SetContent(i, j, ' ', nil, st)
		}
	}

	ui.drawPromptLine(nav)

	length := min(len(ui.wins), len(nav.dirs))
	woff := len(ui.wins) - length

	if gOpts.preview {
		length = min(len(ui.wins)-1, len(nav.dirs))
		woff = len(ui.wins) - 1 - length
	}

	doff := len(nav.dirs) - length
	for i := 0; i < length; i++ {
		ui.wins[woff+i].printDir(ui.screen, nav.dirs[doff+i], nav.selections, nav.saves, ui.styles, ui.icons)
	}

	switch ui.cmdPrefix {
	case "":
		ui.drawStatLine(nav)
		ui.screen.HideCursor()
	case ">":
		ui.msgWin.printLine(ui.screen, 0, 0, st, ui.cmdPrefix)
		ui.msgWin.print(ui.screen, len(ui.cmdPrefix), 0, st, ui.msg)
		ui.msgWin.print(ui.screen, len(ui.cmdPrefix)+len(ui.msg), 0, st, string(ui.cmdAccLeft))
		ui.msgWin.print(ui.screen, len(ui.cmdPrefix)+len(ui.msg)+runeSliceWidth(ui.cmdAccLeft), 0, st, string(ui.cmdAccRight))
		ui.screen.ShowCursor(ui.msgWin.x+len(ui.cmdPrefix)+len(ui.msg)+runeSliceWidth(ui.cmdAccLeft), ui.msgWin.y)
	default:
		ui.msgWin.printLine(ui.screen, 0, 0, st, ui.cmdPrefix)
		ui.msgWin.print(ui.screen, len(ui.cmdPrefix), 0, st, string(ui.cmdAccLeft))
		ui.msgWin.print(ui.screen, len(ui.cmdPrefix)+runeSliceWidth(ui.cmdAccLeft), 0, st, string(ui.cmdAccRight))
		ui.screen.ShowCursor(ui.msgWin.x+len(ui.cmdPrefix)+runeSliceWidth(ui.cmdAccLeft), ui.msgWin.y)
	}

	if gOpts.preview {
		curr, err := nav.currFile()
		if err == nil {
			preview := ui.wins[len(ui.wins)-1]

			if curr.IsDir() {
				preview.printDir(ui.screen, ui.dirPrev, nav.selections, nav.saves, ui.styles, ui.icons)
			} else if curr.Mode().IsRegular() {
				preview.printReg(ui.screen, ui.regPrev)
			}
		}
	}

	if gOpts.drawbox {
		ui.drawBox()
	}

	if ui.menuBuf != nil {
		lines := strings.Split(ui.menuBuf.String(), "\n")

		lines = lines[:len(lines)-1]

		ui.menuWin.h = len(lines) - 1
		ui.menuWin.y = ui.wins[0].h - ui.menuWin.h

		if gOpts.drawbox {
			ui.menuWin.y += 2
		}

		ui.menuWin.printLine(ui.screen, 0, 0, st.Bold(true), lines[0])
		for i, line := range lines[1:] {
			ui.menuWin.printLine(ui.screen, 0, i+1, st, "")
			ui.menuWin.print(ui.screen, 0, i+1, st, line)
		}
	}

	ui.screen.Show()
}

func findBinds(keys map[string]expr, prefix string) (binds map[string]expr, ok bool) {
	binds = make(map[string]expr)
	for key, expr := range keys {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		binds[key] = expr
		if key == prefix {
			ok = true
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

func listMarks(marks map[string]string) *bytes.Buffer {
	t := new(tabwriter.Writer)
	b := new(bytes.Buffer)

	var keys []string
	for k := range marks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	t.Init(b, 0, gOpts.tabstop, 2, '\t', 0)
	fmt.Fprintln(t, "mark\tpath")
	for _, k := range keys {
		fmt.Fprintf(t, "%s\t%s\n", k, marks[k])
	}
	t.Flush()

	return b
}

func (ui *ui) pollEvent() tcell.Event {
	select {
	case key := <-ui.keyChan:
		var ch rune
		var mod tcell.ModMask
		var Key tcell.Key

		if utf8.RuneCountInString(key) == 1 {
			ch, _ = utf8.DecodeRuneInString(key)
		} else {
			switch {
			case key == "<lt>":
				ch = '<'
			case key == "<gt>":
				ch = '>'
			case reAltKey.MatchString(key):
				match := reAltKey.FindStringSubmatch(key)[1]
				ch, _ = utf8.DecodeRuneInString(match)
				mod = tcell.ModMask(tcell.ModAlt)
			default:
				if val, ok := gValKey[key]; ok {
					Key = val
				} else {
					Key = tcell.KeyESC
					ui.echoerrf("unknown key: %s", key)
				}
			}
		}

		return tcell.NewEventKey(Key, ch, mod)
	case ev := <-ui.evChan:
		return ev
	}
}

// This function is used to read a normal event on the client side. For keys,
// digits are interpreted as command counts but this is only done for digits
// preceding any non-digit characters (e.g. "42y2k" as 42 times "y2k").
func (ui *ui) readEvent(ch chan<- expr, ev tcell.Event) {
	if ev == nil {
		return
	}

	draw := &callExpr{"draw", nil, 1}
	count := 1

	switch tev := ev.(type) {
	case *tcell.EventKey:
		// KeyRune is a regular character
		if tev.Key() == tcell.KeyRune {
			switch {
			case tev.Rune() == '<':
				ui.keyAcc = append(ui.keyAcc, []rune("<lt>")...)
			case tev.Rune() == '>':
				ui.keyAcc = append(ui.keyAcc, []rune("<gt>")...)
			case tev.Rune() == ' ':
				ui.keyAcc = append(ui.keyAcc, []rune("<space>")...)
			case tev.Modifiers() == tcell.ModAlt:
				ui.keyAcc = append(ui.keyAcc, '<', 'a', '-', tev.Rune(), '>')
			case unicode.IsDigit(tev.Rune()) && len(ui.keyAcc) == 0:
				ui.keyCount = append(ui.keyCount, tev.Rune())
			default:
				ui.keyAcc = append(ui.keyAcc, tev.Rune())
			}
		} else {
			val := gKeyVal[tev.Key()]
			if val == "<esc>" {
				ch <- draw
				ui.keyAcc = nil
				ui.keyCount = nil
				ui.menuBuf = nil
				return
			}
			ui.keyAcc = append(ui.keyAcc, []rune(val)...)
		}

		if len(ui.keyAcc) == 0 {
			ch <- draw
			break
		}

		binds, ok := findBinds(gOpts.keys, string(ui.keyAcc))

		switch len(binds) {
		case 0:
			ui.echoerrf("unknown mapping: %s", string(ui.keyAcc))
			ch <- draw
			ui.keyAcc = nil
			ui.keyCount = nil
			ui.menuBuf = nil
		default:
			if ok {
				if len(ui.keyCount) > 0 {
					c, err := strconv.Atoi(string(ui.keyCount))
					if err != nil {
						log.Printf("converting command count: %s", err)
					}
					count = c
				}
				expr := gOpts.keys[string(ui.keyAcc)]
				if e, ok := expr.(*callExpr); ok {
					e.count = count
				} else if e, ok := expr.(*listExpr); ok {
					e.count = count
				}
				ch <- expr
				ui.keyAcc = nil
				ui.keyCount = nil
				ui.menuBuf = nil
			} else {
				ui.menuBuf = listBinds(binds)
				ch <- draw
			}
		}
	case *tcell.EventResize:
		ch <- &callExpr{"redraw", nil, 1}
	case *tcell.EventError:
		log.Printf("Got EventError: '%s' at %s", tev.Error(), tev.When())
		return
	case *tcell.EventInterrupt:
		log.Printf("Got EventInterrupt: at %s", tev.When())
		return
	}
}

func readCmdEvent(ch chan<- expr, ev tcell.Event) {
	switch tev := ev.(type) {
	case *tcell.EventKey:
		if tev.Key() == tcell.KeyRune {
			if tev.Modifiers() == tcell.ModMask(tcell.ModAlt) {
				val := string([]rune{'<', 'a', '-', tev.Rune(), '>'})
				if expr, ok := gOpts.cmdkeys[val]; ok {
					ch <- expr
				}
			} else {
				ch <- &callExpr{"cmd-insert", []string{string(tev.Rune())}, 1}
			}
		} else {
			val := gKeyVal[tev.Key()]
			if expr, ok := gOpts.cmdkeys[val]; ok {
				ch <- expr
			}
		}
	}
}

func (ui *ui) readExpr() <-chan expr {
	ch := make(chan expr)

	ui.exprChan = ch

	go func() {
		ch <- &callExpr{"draw", nil, 1}

		for {
			ev := ui.pollEvent()

			switch ev.(type) {
			case *tcell.EventKey:
				if ui.cmdPrefix != "" {
					readCmdEvent(ch, ev)
					continue
				}
			}

			ui.readEvent(ch, ev)
		}
	}()

	return ch
}

func (ui *ui) pause() {
	ui.screen.Fini()
}

func (ui *ui) resume() {
	var screen tcell.Screen
	var err error
	if screen, err = tcell.NewScreen(); err != nil {
		log.Fatalf("creating screen: %s", err)
	} else if err = screen.Init(); err != nil {
		log.Fatalf("initializing screen: %s", err)
	}

	ui.screen = screen

	ui.renew()
}

func listMatches(screen tcell.Screen, matches []string) *bytes.Buffer {
	b := new(bytes.Buffer)

	wtot, _ := screen.Size()

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
