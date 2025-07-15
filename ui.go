package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

const gEscapeCode = 27

var gKeyVal = map[tcell.Key]string{
	tcell.KeyEnter:          "<enter>",
	tcell.KeyBackspace:      "<backspace>",
	tcell.KeyTab:            "<tab>",
	tcell.KeyBacktab:        "<backtab>",
	tcell.KeyEsc:            "<esc>",
	tcell.KeyBackspace2:     "<backspace2>",
	tcell.KeyDelete:         "<delete>",
	tcell.KeyInsert:         "<insert>",
	tcell.KeyUp:             "<up>",
	tcell.KeyDown:           "<down>",
	tcell.KeyLeft:           "<left>",
	tcell.KeyRight:          "<right>",
	tcell.KeyHome:           "<home>",
	tcell.KeyEnd:            "<end>",
	tcell.KeyUpLeft:         "<upleft>",
	tcell.KeyUpRight:        "<upright>",
	tcell.KeyDownLeft:       "<downleft>",
	tcell.KeyDownRight:      "<downright>",
	tcell.KeyCenter:         "<center>",
	tcell.KeyPgDn:           "<pgdn>",
	tcell.KeyPgUp:           "<pgup>",
	tcell.KeyClear:          "<clear>",
	tcell.KeyExit:           "<exit>",
	tcell.KeyCancel:         "<cancel>",
	tcell.KeyPause:          "<pause>",
	tcell.KeyPrint:          "<print>",
	tcell.KeyF1:             "<f-1>",
	tcell.KeyF2:             "<f-2>",
	tcell.KeyF3:             "<f-3>",
	tcell.KeyF4:             "<f-4>",
	tcell.KeyF5:             "<f-5>",
	tcell.KeyF6:             "<f-6>",
	tcell.KeyF7:             "<f-7>",
	tcell.KeyF8:             "<f-8>",
	tcell.KeyF9:             "<f-9>",
	tcell.KeyF10:            "<f-10>",
	tcell.KeyF11:            "<f-11>",
	tcell.KeyF12:            "<f-12>",
	tcell.KeyF13:            "<f-13>",
	tcell.KeyF14:            "<f-14>",
	tcell.KeyF15:            "<f-15>",
	tcell.KeyF16:            "<f-16>",
	tcell.KeyF17:            "<f-17>",
	tcell.KeyF18:            "<f-18>",
	tcell.KeyF19:            "<f-19>",
	tcell.KeyF20:            "<f-20>",
	tcell.KeyF21:            "<f-21>",
	tcell.KeyF22:            "<f-22>",
	tcell.KeyF23:            "<f-23>",
	tcell.KeyF24:            "<f-24>",
	tcell.KeyF25:            "<f-25>",
	tcell.KeyF26:            "<f-26>",
	tcell.KeyF27:            "<f-27>",
	tcell.KeyF28:            "<f-28>",
	tcell.KeyF29:            "<f-29>",
	tcell.KeyF30:            "<f-30>",
	tcell.KeyF31:            "<f-31>",
	tcell.KeyF32:            "<f-32>",
	tcell.KeyF33:            "<f-33>",
	tcell.KeyF34:            "<f-34>",
	tcell.KeyF35:            "<f-35>",
	tcell.KeyF36:            "<f-36>",
	tcell.KeyF37:            "<f-37>",
	tcell.KeyF38:            "<f-38>",
	tcell.KeyF39:            "<f-39>",
	tcell.KeyF40:            "<f-40>",
	tcell.KeyF41:            "<f-41>",
	tcell.KeyF42:            "<f-42>",
	tcell.KeyF43:            "<f-43>",
	tcell.KeyF44:            "<f-44>",
	tcell.KeyF45:            "<f-45>",
	tcell.KeyF46:            "<f-46>",
	tcell.KeyF47:            "<f-47>",
	tcell.KeyF48:            "<f-48>",
	tcell.KeyF49:            "<f-49>",
	tcell.KeyF50:            "<f-50>",
	tcell.KeyF51:            "<f-51>",
	tcell.KeyF52:            "<f-52>",
	tcell.KeyF53:            "<f-53>",
	tcell.KeyF54:            "<f-54>",
	tcell.KeyF55:            "<f-55>",
	tcell.KeyF56:            "<f-56>",
	tcell.KeyF57:            "<f-57>",
	tcell.KeyF58:            "<f-58>",
	tcell.KeyF59:            "<f-59>",
	tcell.KeyF60:            "<f-60>",
	tcell.KeyF61:            "<f-61>",
	tcell.KeyF62:            "<f-62>",
	tcell.KeyF63:            "<f-63>",
	tcell.KeyF64:            "<f-64>",
	tcell.KeyCtrlA:          "<c-a>",
	tcell.KeyCtrlB:          "<c-b>",
	tcell.KeyCtrlC:          "<c-c>",
	tcell.KeyCtrlD:          "<c-d>",
	tcell.KeyCtrlE:          "<c-e>",
	tcell.KeyCtrlF:          "<c-f>",
	tcell.KeyCtrlG:          "<c-g>",
	tcell.KeyCtrlJ:          "<c-j>",
	tcell.KeyCtrlK:          "<c-k>",
	tcell.KeyCtrlL:          "<c-l>",
	tcell.KeyCtrlN:          "<c-n>",
	tcell.KeyCtrlO:          "<c-o>",
	tcell.KeyCtrlP:          "<c-p>",
	tcell.KeyCtrlQ:          "<c-q>",
	tcell.KeyCtrlR:          "<c-r>",
	tcell.KeyCtrlS:          "<c-s>",
	tcell.KeyCtrlT:          "<c-t>",
	tcell.KeyCtrlU:          "<c-u>",
	tcell.KeyCtrlV:          "<c-v>",
	tcell.KeyCtrlW:          "<c-w>",
	tcell.KeyCtrlX:          "<c-x>",
	tcell.KeyCtrlY:          "<c-y>",
	tcell.KeyCtrlZ:          "<c-z>",
	tcell.KeyCtrlSpace:      "<c-space>",
	tcell.KeyCtrlUnderscore: "<c-_>",
	tcell.KeyCtrlRightSq:    "<c-]>",
	tcell.KeyCtrlBackslash:  "<c-\\>",
	tcell.KeyCtrlCarat:      "<c-^>",
}

var gValKey map[string]tcell.Key

func init() {
	gValKey = make(map[string]tcell.Key, len(gKeyVal))
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
	slen := len(s)
	for i := 0; i < slen; i++ {
		r, w := utf8.DecodeRuneInString(s[i:])

		if r == gEscapeCode && i+1 < slen && s[i+1] == '[' {
			j := strings.IndexAny(s[i:min(slen, i+64)], "mK")
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
	var comb []rune
	slen := len(s)
	for i := 0; i < slen; i++ {
		r, w := utf8.DecodeRuneInString(s[i:])

		if r == gEscapeCode && i+1 < slen && s[i+1] == '[' {
			j := strings.IndexAny(s[i:min(slen, i+64)], "mK")
			if j == -1 {
				continue
			}
			if s[i+j] == 'm' {
				st = applyAnsiCodes(s[i+2:i+j], st)
			}

			i += j
			continue
		}

		for {
			rc, wc := utf8.DecodeRuneInString(s[i+w:])
			if !unicode.Is(unicode.Mn, rc) {
				break
			}
			comb = append(comb, rc)
			i += wc
		}

		if x < win.w {
			screen.SetContent(win.x+x, win.y+y, r, comb, st)
			comb = nil
		}

		i += w - 1

		if r == '\t' {
			ind := gOpts.tabstop - (x-off)%gOpts.tabstop
			for i := 0; i < ind && x+i < win.w; i++ {
				screen.SetContent(win.x+x+i, win.y+y, ' ', nil, st)
			}
			x += ind
		} else {
			x += runewidth.RuneWidth(r)
		}
	}

	return st
}

func (win *win) printf(screen tcell.Screen, x, y int, st tcell.Style, format string, a ...any) {
	win.print(screen, x, y, st, fmt.Sprintf(format, a...))
}

func (win *win) printLine(screen tcell.Screen, x, y int, st tcell.Style, s string) {
	win.printf(screen, x, y, st, "%s%*s", s, win.w-printLength(s), "")
}

func (win *win) printRight(screen tcell.Screen, y int, st tcell.Style, s string) {
	win.print(screen, win.w-printLength(s), y, st, s)
}

func (win *win) printReg(screen tcell.Screen, reg *reg, previewLoading bool, sxs *sixelScreen) {
	if reg == nil {
		return
	}

	st := tcell.StyleDefault

	if reg.loading {
		if previewLoading {
			st = st.Reverse(true)
			win.print(screen, 2, 0, st, "loading...")
		}
		return
	}

	for i, l := range reg.lines {
		if i > win.h-1 {
			break
		}

		st = win.print(screen, 2, i, st, l)
	}

	sxs.printSixel(win, screen, reg)
}

var gThisYear = time.Now().Year()

func infotimefmt(t time.Time) string {
	if t.Year() == gThisYear {
		return t.Format(gOpts.infotimefmtnew)
	}
	return t.Format(gOpts.infotimefmtold)
}

func fileInfo(f *file, d *dir, userWidth int, groupWidth int, customWidth int) (string, string, int) {
	var info strings.Builder
	var custom string
	var off int

	for _, s := range getInfo(d.path) {
		switch s {
		case "size":
			if f.IsDir() && getDirCounts(d.path) {
				switch {
				case f.dirCount < -1:
					info.WriteString("    !")
				case f.dirCount < 0:
					info.WriteString("    ?")
				case f.dirCount < 1000:
					fmt.Fprintf(&info, " %4d", f.dirCount)
				default:
					info.WriteString(" 999+")
				}
				continue
			}

			var sz string
			if f.IsDir() && f.dirSize < 0 {
				sz = "-"
			} else {
				sz = humanize(f.TotalSize())
			}
			fmt.Fprintf(&info, " %4s", sz)
		case "time":
			fmt.Fprintf(&info, " %*s", max(len(gOpts.infotimefmtnew), len(gOpts.infotimefmtold)), infotimefmt(f.ModTime()))
		case "atime":
			fmt.Fprintf(&info, " %*s", max(len(gOpts.infotimefmtnew), len(gOpts.infotimefmtold)), infotimefmt(f.accessTime))
		case "btime":
			fmt.Fprintf(&info, " %*s", max(len(gOpts.infotimefmtnew), len(gOpts.infotimefmtold)), infotimefmt(f.birthTime))
		case "ctime":
			fmt.Fprintf(&info, " %*s", max(len(gOpts.infotimefmtnew), len(gOpts.infotimefmtold)), infotimefmt(f.changeTime))
		case "perm":
			info.WriteString(" " + f.FileInfo.Mode().String())
		case "user":
			fmt.Fprintf(&info, " %-*s", userWidth, userName(f.FileInfo))
		case "group":
			fmt.Fprintf(&info, " %-*s", groupWidth, groupName(f.FileInfo))
		case "custom":
			// To allow for the usage of escape sequences, store `custom`
			// separately and print it later using the offset.
			off = info.Len()
			fmt.Fprintf(&info, " %*s", customWidth, "")
			custom = fmt.Sprintf(" %s%*s", f.customInfo, customWidth-printLength(f.customInfo), "")
		default:
			log.Printf("unknown info type: %s", s)
		}
	}

	return info.String(), custom, off
}

type dirContext struct {
	selections map[string]int
	saves      map[string]bool
	tags       map[string]string
}

type dirRole byte

const (
	Active dirRole = iota
	Parent
	Preview
)

type dirStyle struct {
	colors styleMap
	icons  iconMap
	role   dirRole
}

func (win *win) printDir(ui *ui, dir *dir, context *dirContext, dirStyle *dirStyle) {
	if win.w < 5 || dir == nil {
		return
	}

	messageStyle := tcell.StyleDefault.Reverse(true)

	if dir.noPerm {
		win.print(ui.screen, 2, 0, messageStyle, "permission denied")
		return
	}
	fileslen := len(dir.files)
	if dir.loading && fileslen == 0 {
		win.print(ui.screen, 2, 0, messageStyle, "loading...")
		return
	}

	if fileslen == 0 {
		win.print(ui.screen, 2, 0, messageStyle, "empty")
		return
	}

	beg := max(dir.ind-dir.pos, 0)
	end := min(beg+win.h, fileslen)

	if beg > end {
		return
	}

	var lnwidth int

	if dirStyle.role == Active && (gOpts.number || gOpts.relativenumber) {
		lnwidth = 1
		if gOpts.number && gOpts.relativenumber {
			lnwidth++
		}
		for j := 10; j <= fileslen; j *= 10 {
			lnwidth++
		}
	}

	var userWidth int
	var groupWidth int
	var customWidth int

	// Only fetch user/group/custom widths if configured to display them
	for _, s := range getInfo(dir.path) {
		switch s {
		case "user":
			userWidth = getUserWidth(dir, beg, end)
		case "group":
			groupWidth = getGroupWidth(dir, beg, end)
		case "custom":
			customWidth = getCustomWidth(dir, beg, end)
		}

		if userWidth > 0 && groupWidth > 0 && customWidth > 0 {
			break
		}
	}

	visualSelections := dir.visualSelections()
	for i, f := range dir.files[beg:end] {
		st := dirStyle.colors.get(f)

		if lnwidth > 0 {
			var ln string

			if gOpts.number && (!gOpts.relativenumber) {
				ln = fmt.Sprintf("%*d", lnwidth, i+1+beg)
			} else if gOpts.relativenumber {
				switch {
				case i < dir.pos:
					ln = fmt.Sprintf("%*d", lnwidth, dir.pos-i)
				case i > dir.pos:
					ln = fmt.Sprintf("%*d", lnwidth, i-dir.pos)
				case gOpts.number:
					ln = fmt.Sprintf("%*d ", lnwidth-1, i+1+beg)
				default:
					ln = fmt.Sprintf("%*d", lnwidth, 0)
				}
			}

			win.print(ui.screen, 0, i, tcell.StyleDefault, fmt.Sprintf(optionToFmtstr(gOpts.numberfmt), ln))
		}

		path := filepath.Join(dir.path, f.Name())

		if slices.Contains(visualSelections, path) {
			win.print(ui.screen, lnwidth, i, parseEscapeSequence(gOpts.visualfmt), " ")
		} else if _, ok := context.selections[path]; ok {
			win.print(ui.screen, lnwidth, i, parseEscapeSequence(gOpts.selectfmt), " ")
		} else if cp, ok := context.saves[path]; ok {
			if cp {
				win.print(ui.screen, lnwidth, i, parseEscapeSequence(gOpts.copyfmt), " ")
			} else {
				win.print(ui.screen, lnwidth, i, parseEscapeSequence(gOpts.cutfmt), " ")
			}
		}

		// make space for select marker, and leave another space at the end
		maxWidth := win.w - lnwidth - 2
		// make extra space to separate windows if drawbox is not enabled
		if !gOpts.drawbox {
			maxWidth -= 1
		}

		tag := " "
		if val, ok := context.tags[evalSymlinks(path)]; ok && len(val) > 0 {
			tag = val
		}

		var icon []rune
		var iconDef iconDef
		if gOpts.icons {
			iconDef = dirStyle.icons.get(f)
			icon = append(icon, []rune(iconDef.icon)...)
			icon = append(icon, ' ')
		}

		// subtract space for tag and icon
		maxFilenameWidth := maxWidth - 1 - runeSliceWidth(icon)

		info, custom, off := fileInfo(f, dir, userWidth, groupWidth, customWidth)
		infolen := len(info)
		showInfo := infolen > 0 && 2*infolen < maxWidth
		if showInfo {
			maxFilenameWidth -= infolen
		}

		filename := []rune(f.Name())
		if runeSliceWidth(filename) > maxFilenameWidth {
			truncatePos := (maxFilenameWidth - 1) * gOpts.truncatepct / 100
			lastPart := runeSliceWidthLastRange(filename, maxFilenameWidth-truncatePos-1)
			filename = runeSliceWidthRange(filename, 0, truncatePos)
			filename = append(filename, []rune(gOpts.truncatechar)...)
			filename = append(filename, lastPart...)
		}
		for j := runeSliceWidth(filename); j < maxFilenameWidth; j++ {
			filename = append(filename, ' ')
		}

		if showInfo {
			filename = append(filename, []rune(info)...)
			off += lnwidth + 2 + runeSliceWidth(icon) + maxFilenameWidth
		}

		if i == dir.pos {
			var cursorFmt string
			switch dirStyle.role {
			case Active:
				cursorFmt = optionToFmtstr(gOpts.cursoractivefmt)
			case Parent:
				cursorFmt = optionToFmtstr(gOpts.cursorparentfmt)
			case Preview:
				cursorFmt = optionToFmtstr(gOpts.cursorpreviewfmt)
			}

			// print tag separately as it can contain color escape sequences
			win.print(ui.screen, lnwidth+1, i, st, fmt.Sprintf(cursorFmt, tag))

			line := append(icon, filename...)
			line = append(line, ' ')
			win.print(ui.screen, lnwidth+2, i, st, fmt.Sprintf(cursorFmt, string(line)))

			// print over the empty space we reserved for the custom info
			if showInfo && custom != "" {
				win.print(ui.screen, off, i, st, fmt.Sprintf(cursorFmt, stripAnsi(custom)))
			}
		} else {
			if tag == " " {
				win.print(ui.screen, lnwidth+1, i, st, " ")
			} else {
				tagStr := fmt.Sprintf(optionToFmtstr(gOpts.tagfmt), tag)
				win.print(ui.screen, lnwidth+1, i, tcell.StyleDefault, tagStr)
			}

			if len(icon) > 0 {
				iconStyle := st
				if iconDef.hasStyle {
					iconStyle = iconDef.style
				}
				win.print(ui.screen, lnwidth+2, i, iconStyle, string(icon))
			}

			line := append(filename, ' ')
			win.print(ui.screen, lnwidth+2+runeSliceWidth(icon), i, st, string(line))

			// print over the empty space we reserved for the custom info
			if showInfo && custom != "" {
				win.print(ui.screen, off, i, st, custom)
			}
		}
	}
}

func getUserWidth(dir *dir, beg int, end int) int {
	maxw := 0

	for _, f := range dir.files[beg:end] {
		maxw = max(len(userName(f.FileInfo)), maxw)
	}

	return maxw
}

func getGroupWidth(dir *dir, beg int, end int) int {
	maxw := 0

	for _, f := range dir.files[beg:end] {
		maxw = max(len(groupName(f.FileInfo)), maxw)
	}

	return maxw
}

func getCustomWidth(dir *dir, beg int, end int) int {
	maxw := 0

	for _, f := range dir.files[beg:end] {
		maxw = max(printLength(f.customInfo), maxw)
	}

	return maxw
}

func getWidths(wtot int) []int {
	rsum := 0
	for _, r := range gOpts.ratios {
		rsum += r
	}

	wlen := len(gOpts.ratios)
	widths := make([]int, wlen)

	if gOpts.drawbox {
		wtot -= (wlen + 1)
	}

	wsum := 0
	for i := range wlen - 1 {
		widths[i] = gOpts.ratios[i] * wtot / rsum
		wsum += widths[i]
	}
	widths[wlen-1] = wtot - wsum

	return widths
}

func getWins(screen tcell.Screen) []*win {
	wtot, htot := screen.Size()

	widths := getWidths(wtot)

	wacc := 0
	wlen := len(widths)
	wins := make([]*win, 0, wlen)
	for i := range wlen {
		if gOpts.drawbox {
			wacc++
			wins = append(wins, newWin(widths[i], htot-4, wacc, 2))
		} else {
			wins = append(wins, newWin(widths[i], htot-2, wacc, 1))
		}
		wacc += widths[i]
	}

	return wins
}

type ui struct {
	screen      tcell.Screen
	sxScreen    sixelScreen
	polling     bool
	wins        []*win
	promptWin   *win
	msgWin      *win
	menuWin     *win
	msg         string
	msgIsStat   bool
	regPrev     *reg
	dirPrev     *dir
	exprChan    chan expr
	keyChan     chan string
	tevChan     chan tcell.Event
	evChan      chan tcell.Event
	menu        string
	cmdPrefix   string
	cmdAccLeft  []rune
	cmdAccRight []rune
	cmdYankBuf  []rune
	cmdTmp      []rune
	keyAcc      []rune
	keyCount    []rune
	styles      styleMap
	icons       iconMap
	currentFile string
}

func newUI(screen tcell.Screen) *ui {
	wtot, htot := screen.Size()

	ui := &ui{
		screen:      screen,
		polling:     true,
		wins:        getWins(screen),
		promptWin:   newWin(wtot, 1, 0, 0),
		msgWin:      newWin(wtot, 1, 0, htot-1),
		menuWin:     newWin(wtot, 1, 0, htot-2),
		msgIsStat:   true,
		exprChan:    make(chan expr, 1000),
		keyChan:     make(chan string, 1000),
		tevChan:     make(chan tcell.Event, 1000),
		evChan:      make(chan tcell.Event, 1000),
		styles:      parseStyles(),
		icons:       parseIcons(),
		currentFile: "",
		sxScreen:    sixelScreen{},
	}

	go ui.pollEvents()

	return ui
}

func (ui *ui) winAt(x, y int) (int, *win) {
	for i := len(ui.wins) - 1; i >= 0; i-- {
		w := ui.wins[i]
		if x >= w.x && y >= w.y && y < w.y+w.h {
			return i, w
		}
	}
	return -1, nil
}

func (ui *ui) pollEvents() {
	var ev tcell.Event
	for {
		ev = ui.screen.PollEvent()
		if ev == nil {
			ui.polling = false
			return
		}
		ui.tevChan <- ev
	}
}

func (ui *ui) renew() {
	ui.wins = getWins(ui.screen)

	wtot, htot := ui.screen.Size()
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
	ui.msgIsStat = false
}

func (ui *ui) echomsg(msg string) {
	ui.echo(msg)
	log.Print(msg)
}

func optionToFmtstr(optstr string) string {
	if !strings.Contains(optstr, "%s") {
		return optstr + "%s\033[0m"
	} else {
		return optstr
	}
}

func (ui *ui) echoerr(msg string) {
	ui.echo(fmt.Sprintf(optionToFmtstr(gOpts.errorfmt), msg))
	log.Printf("error: %s", msg)
}

func (ui *ui) echoerrf(format string, a ...any) {
	ui.echoerr(fmt.Sprintf(format, a...))
}

// This represents the preview for a regular file.
// This can also be used to represent the preview of a directory if
// `dirpreviews` is enabled.
type reg struct {
	loading  bool
	volatile bool
	loadTime time.Time
	path     string
	lines    []string
	sixel    *string
}

func (ui *ui) loadFile(app *app, volatile bool) {
	if !app.nav.init {
		return
	}

	curr, err := app.nav.currFile()
	if err != nil {
		return
	}

	if curr.path != ui.currentFile {
		ui.currentFile = curr.path
		onSelect(app)
	}

	if volatile {
		app.nav.previewChan <- ""
	}

	if !gOpts.preview {
		return
	}

	if curr.Mode().IsRegular() || (curr.IsDir() && gOpts.dirpreviews) {
		ui.regPrev = app.nav.loadReg(curr.path, volatile)
	} else if curr.IsDir() {
		ui.dirPrev = app.nav.loadDir(curr.path)
	}
}

func (ui *ui) loadFileInfo(nav *nav) {
	if !nav.init {
		return
	}

	ui.msg = ""
	ui.msgIsStat = true

	curr, err := nav.currFile()
	if err != nil {
		return
	}

	if curr.err != nil {
		ui.echoerrf("stat: %s", curr.err)
		return
	}

	statfmt := strings.ReplaceAll(gOpts.statfmt, "|", "\x1f")
	replace := func(s string, val string) {
		if val == "" {
			val = "\x00"
		}
		statfmt = strings.ReplaceAll(statfmt, s, val)
	}
	if nav.isVisualMode() {
		replace("%m", "VISUAL")
		replace("%M", "VISUAL")
	} else {
		replace("%m", "")
		replace("%M", "NORMAL")
	}
	replace("%p", curr.Mode().String())
	replace("%c", linkCount(curr))
	replace("%u", userName(curr))
	replace("%g", groupName(curr))
	replace("%s", humanize(curr.Size()))
	replace("%S", fmt.Sprintf("%4s", humanize(curr.Size())))
	replace("%t", curr.ModTime().Format(gOpts.timefmt))
	replace("%l", curr.linkTarget)

	var fileInfo strings.Builder
	for _, section := range strings.Split(statfmt, "\x1f") {
		if !strings.Contains(section, "\x00") {
			fileInfo.WriteString(section)
		}
	}

	ui.msg = fileInfo.String()
}

func (ui *ui) drawPromptLine(nav *nav) {
	st := tcell.StyleDefault

	dir := nav.currDir()
	pwd := dir.path

	if strings.HasPrefix(pwd, gUser.HomeDir) {
		pwd = filepath.Join("~", strings.TrimPrefix(pwd, gUser.HomeDir))
	}

	sep := string(filepath.Separator)

	var fname string
	curr, err := nav.currFile()
	if err == nil {
		fname = filepath.Base(curr.path)
	}

	var prompt string

	prompt = strings.ReplaceAll(gOpts.promptfmt, "%u", gUser.Username)
	prompt = strings.ReplaceAll(prompt, "%h", gHostname)
	prompt = strings.ReplaceAll(prompt, "%f", fname)

	if printLength(strings.ReplaceAll(strings.ReplaceAll(prompt, "%w", pwd), "%d", pwd)) > ui.promptWin.w {
		names := strings.Split(pwd, sep)
		for i := range names {
			if names[i] == "" {
				continue
			}
			r, _ := utf8.DecodeRuneInString(names[i])
			names[i] = string(r)
			if printLength(strings.ReplaceAll(strings.ReplaceAll(prompt, "%w", strings.Join(names, sep)), "%d", strings.Join(names, sep))) <= ui.promptWin.w {
				break
			}
		}
		pwd = strings.Join(names, sep)
	}

	prompt = strings.ReplaceAll(prompt, "%w", pwd)
	if !strings.HasSuffix(pwd, sep) {
		pwd += sep
	}
	prompt = strings.ReplaceAll(prompt, "%d", pwd)

	if len(dir.filter) != 0 {
		prompt = strings.ReplaceAll(prompt, "%F", fmt.Sprint(dir.filter))
	} else {
		prompt = strings.ReplaceAll(prompt, "%F", "")
	}

	// spacer
	avail := ui.promptWin.w - printLength(prompt) + 2
	if avail > 0 {
		prompt = strings.Replace(prompt, "%S", strings.Repeat(" ", avail), 1)
	}
	prompt = strings.ReplaceAll(prompt, "%S", "")

	ui.promptWin.print(ui.screen, 0, 0, st, prompt)
}

func formatRulerOpt(name string, val string) string {
	// handle escape character so it doesn't mess up the ruler
	val = strings.ReplaceAll(val, "\033", "\033[7m\\033\033[0m")

	// display name of builtin options for clarity
	if !strings.HasPrefix(name, "lf_user_") {
		return fmt.Sprintf("%s=%s", strings.TrimPrefix(name, "lf_"), val)
	}

	return val
}

func (ui *ui) drawRuler(nav *nav) {
	st := tcell.StyleDefault

	dir := nav.currDir()

	ui.msgWin.print(ui.screen, 0, 0, st, ui.msg)

	tot := len(dir.files)
	ind := min(dir.ind+1, tot)
	hid := len(dir.allFiles) - tot
	acc := string(ui.keyCount) + string(ui.keyAcc)

	var percentage string
	beg := max(dir.ind-dir.pos, 0)
	switch {
	case tot <= nav.height:
		percentage = "All"
	case beg == 0:
		percentage = "Top"
	case beg == tot-nav.height:
		percentage = "Bot"
	default:
		percentage = fmt.Sprintf("%2d%%", beg*100/(tot-nav.height))
	}

	copy := 0
	move := 0
	for _, cp := range nav.saves {
		if cp {
			copy++
		} else {
			move++
		}
	}

	currSelections := nav.currSelections()
	currVSelections := nav.currDir().visualSelections()

	progress := []string{}

	if nav.copyTotal > 0 {
		progress = append(progress, fmt.Sprintf("[%d%%]", nav.copyBytes*100/nav.copyTotal))
	}

	if nav.moveTotal > 0 {
		progress = append(progress, fmt.Sprintf("[%d/%d]", nav.moveCount, nav.moveTotal))
	}

	if nav.deleteTotal > 0 {
		progress = append(progress, fmt.Sprintf("[%d/%d]", nav.deleteCount, nav.deleteTotal))
	}

	opts := getOptsMap()

	rulerfmt := strings.ReplaceAll(gOpts.rulerfmt, "|", "\x1f")
	rulerfmt = reRulerSub.ReplaceAllStringFunc(rulerfmt, func(s string) string {
		var result string
		switch s {
		case "%a":
			result = acc
		case "%p":
			result = strings.Join(progress, " ")
		case "%m":
			result = fmt.Sprintf("%.d", move)
		case "%c":
			result = fmt.Sprintf("%.d", copy)
		case "%s":
			result = fmt.Sprintf("%.d", len(currSelections))
		case "%v":
			result = fmt.Sprintf("%.d", len(currVSelections))
		case "%f":
			result = strings.Join(dir.filter, " ")
		case "%i":
			result = strconv.Itoa(ind)
		case "%t":
			result = strconv.Itoa(tot)
		case "%h":
			result = strconv.Itoa(hid)
		case "%P":
			result = percentage
		case "%d":
			result = diskFree(dir.path)
		default:
			s = strings.TrimSuffix(strings.TrimPrefix(s, "%{"), "}")
			if val, ok := opts[s]; ok {
				result = formatRulerOpt(s, val)
			}
		}
		if result == "" {
			return "\x00"
		}
		return result
	})
	var ruler strings.Builder
	for _, section := range strings.Split(rulerfmt, "\x1f") {
		if !strings.Contains(section, "\x00") {
			ruler.WriteString(section)
		}
	}
	ui.msgWin.printRight(ui.screen, 0, st, ruler.String())
}

func (ui *ui) drawBox() {
	st := parseEscapeSequence(gOpts.borderfmt)

	w, h := ui.screen.Size()

	for i := 1; i < w-1; i++ {
		ui.screen.SetContent(i, 1, tcell.RuneHLine, nil, st)
		ui.screen.SetContent(i, h-2, tcell.RuneHLine, nil, st)
	}

	for i := 2; i < h-2; i++ {
		ui.screen.SetContent(0, i, tcell.RuneVLine, nil, st)
		ui.screen.SetContent(w-1, i, tcell.RuneVLine, nil, st)
	}

	if gOpts.roundbox {
		ui.screen.SetContent(0, 1, '╭', nil, st)
		ui.screen.SetContent(w-1, 1, '╮', nil, st)
		ui.screen.SetContent(0, h-2, '╰', nil, st)
		ui.screen.SetContent(w-1, h-2, '╯', nil, st)
	} else {
		ui.screen.SetContent(0, 1, tcell.RuneULCorner, nil, st)
		ui.screen.SetContent(w-1, 1, tcell.RuneURCorner, nil, st)
		ui.screen.SetContent(0, h-2, tcell.RuneLLCorner, nil, st)
		ui.screen.SetContent(w-1, h-2, tcell.RuneLRCorner, nil, st)
	}

	wacc := 0
	for wind := range len(ui.wins) - 1 {
		wacc += ui.wins[wind].w + 1
		ui.screen.SetContent(wacc, 1, tcell.RuneTTee, nil, st)
		for i := 2; i < h-2; i++ {
			ui.screen.SetContent(wacc, i, tcell.RuneVLine, nil, st)
		}
		ui.screen.SetContent(wacc, h-2, tcell.RuneBTee, nil, st)
	}
}

func (ui *ui) dirOfWin(nav *nav, wind int) *dir {
	wins := len(ui.wins)
	if gOpts.preview {
		wins--
	}
	ind := len(nav.dirs) - wins + wind
	if ind < 0 {
		return nil
	}
	return nav.dirs[ind]
}

func (ui *ui) draw(nav *nav) {
	st := tcell.StyleDefault
	context := dirContext{selections: nav.selections, saves: nav.saves, tags: nav.tags}

	ui.screen.Clear()

	ui.drawPromptLine(nav)

	wins := len(ui.wins)
	if gOpts.preview {
		wins--
	}
	for i := range wins {
		role := Parent
		if i == wins-1 {
			role = Active
		}
		if dir := ui.dirOfWin(nav, i); dir != nil {
			ui.wins[i].printDir(ui, dir, &context,
				&dirStyle{colors: ui.styles, icons: ui.icons, role: role})
		}
	}

	switch ui.cmdPrefix {
	case "":
		ui.drawRuler(nav)
		ui.screen.HideCursor()
	case ">":
		maxWidth := ui.msgWin.w - 1 // leave space for cursor at the end
		prefix := runeSliceWidthRange([]rune(ui.cmdPrefix), 0, maxWidth)
		left := runeSliceWidthLastRange(ui.cmdAccLeft, maxWidth-runeSliceWidth(prefix)-printLength(ui.msg))
		ui.msgWin.printLine(ui.screen, 0, 0, st, string(prefix)+ui.msg)
		ui.msgWin.print(ui.screen, runeSliceWidth(prefix)+printLength(ui.msg), 0, st, string(left)+string(ui.cmdAccRight))
		ui.screen.ShowCursor(ui.msgWin.x+runeSliceWidth(prefix)+printLength(ui.msg)+runeSliceWidth(left), ui.msgWin.y)
	default:
		maxWidth := ui.msgWin.w - 1 // leave space for cursor at the end
		prefix := runeSliceWidthRange([]rune(ui.cmdPrefix), 0, maxWidth)
		left := runeSliceWidthLastRange(ui.cmdAccLeft, maxWidth-runeSliceWidth(prefix))
		ui.msgWin.printLine(ui.screen, 0, 0, st, string(prefix)+string(left)+string(ui.cmdAccRight))
		ui.screen.ShowCursor(ui.msgWin.x+runeSliceWidth(prefix)+runeSliceWidth(left), ui.msgWin.y)
	}

	curr, err := nav.currFile()
	if err == nil {
		preview := ui.wins[len(ui.wins)-1]
		ui.sxScreen.clearSixel(preview, ui.screen, curr.path)
		if gOpts.preview {
			if curr.Mode().IsRegular() || (curr.IsDir() && gOpts.dirpreviews) {
				preview.printReg(ui.screen, ui.regPrev, nav.previewLoading, &ui.sxScreen)
			} else if curr.IsDir() {
				ui.sxScreen.lastFile = ""
				preview.printDir(ui, ui.dirPrev, &context,
					&dirStyle{colors: ui.styles, icons: ui.icons, role: Preview})
			}
		}
	}

	if gOpts.drawbox {
		ui.drawBox()
	}

	if ui.menu != "" {
		lines := strings.Split(ui.menu, "\n")

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

func listBinds(binds map[string]map[string]expr) string {
	t := new(tabwriter.Writer)
	b := new(bytes.Buffer)

	m := make(map[string]map[string]string)
	for mode, keys := range binds {
		for key, expr := range keys {
			if _, ok := m[key]; !ok {
				m[key] = make(map[string]string)
			}
			m[key][expr.String()] += mode
		}
	}

	type entry struct {
		mode, key, cmd string
	}

	var entries []entry
	for key, cmds := range m {
		for cmd, modes := range cmds {
			tmp := []rune(modes)
			slices.Sort(tmp)
			entries = append(entries, entry{string(tmp), key, cmd})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].key != entries[j].key {
			return entries[i].key < entries[j].key
		}
		return entries[i].mode < entries[j].mode
	})

	t.Init(b, 0, gOpts.tabstop, 2, '\t', 0)
	fmt.Fprintln(t, "mode\tkeys\tcommand")
	for _, e := range entries {
		fmt.Fprintf(t, "%s\t%s\t%s\n", e.mode, e.key, e.cmd)
	}
	t.Flush()

	return b.String()
}

func listCmds(cmds map[string]expr) string {
	t := new(tabwriter.Writer)
	b := new(bytes.Buffer)

	keys := make([]string, 0, len(cmds))
	for k := range cmds {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	t.Init(b, 0, gOpts.tabstop, 2, '\t', 0)
	fmt.Fprintln(t, "name\tcommand")
	for _, k := range keys {
		fmt.Fprintf(t, "%s\t%v\n", k, cmds[k])
	}
	t.Flush()

	return b.String()
}

func listJumps(jumps []string, ind int) string {
	t := new(tabwriter.Writer)
	b := new(bytes.Buffer)

	maxlength := len(strconv.Itoa(max(ind, len(jumps)-1-ind)))

	t.Init(b, 0, gOpts.tabstop, 2, '\t', 0)
	fmt.Fprintln(t, "  jump\tpath")
	// print jumps in order of most recent, Vim uses the opposite order
	for i := len(jumps) - 1; i >= 0; i-- {
		switch {
		case i < ind:
			fmt.Fprintf(t, "  %*d\t%s\n", maxlength, ind-i, jumps[i])
		case i > ind:
			fmt.Fprintf(t, "  %*d\t%s\n", maxlength, i-ind, jumps[i])
		default:
			fmt.Fprintf(t, "> %*d\t%s\n", maxlength, 0, jumps[i])
		}
	}
	t.Flush()

	return b.String()
}

func listHistory(history []cmdItem) string {
	t := new(tabwriter.Writer)
	b := new(bytes.Buffer)

	maxlength := len(strconv.Itoa(len(history)))

	t.Init(b, 0, gOpts.tabstop, 2, '\t', 0)
	fmt.Fprintln(t, "number\tcommand")
	for i, cmd := range history {
		fmt.Fprintf(t, "%*d\t%s%s\n", maxlength, i+1, cmd.prefix, cmd.value)
	}
	t.Flush()

	return b.String()
}

func listMarks(marks map[string]string) string {
	t := new(tabwriter.Writer)
	b := new(bytes.Buffer)

	keys := make([]string, 0, len(marks))
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

	return b.String()
}

func listFilesInCurrDir(nav *nav) string {
	if !nav.init {
		return ""
	}
	dir := nav.currDir()
	if dir.loading {
		log.Printf("listFilesInCurrDir(): %s is still loading, `files` isn't ready for remote query", dir.path)
		return ""
	}

	b := new(strings.Builder)
	for _, file := range dir.files {
		fmt.Fprintln(b, file.path)
	}

	return b.String()
}

func (ui *ui) pollEvent() tcell.Event {
	select {
	case val := <-ui.keyChan:
		var ch rune
		var mod tcell.ModMask
		k := tcell.KeyRune

		if key, ok := gValKey[val]; ok {
			return tcell.NewEventKey(key, ch, mod)
		}

		switch {
		case utf8.RuneCountInString(val) == 1:
			ch, _ = utf8.DecodeRuneInString(val)
		case val == "<lt>":
			ch = '<'
		case val == "<gt>":
			ch = '>'
		case val == "<space>":
			ch = ' '
		case reModKey.MatchString(val):
			matches := reModKey.FindStringSubmatch(val)
			switch matches[1] {
			case "c":
				mod = tcell.ModCtrl
			case "s":
				mod = tcell.ModShift
			case "a":
				mod = tcell.ModAlt
			}
			val = matches[2]
			if utf8.RuneCountInString(val) == 1 {
				ch, _ = utf8.DecodeRuneInString(val)
				break
			} else if key, ok := gValKey["<"+val+">"]; ok {
				k = key
				break
			}
			fallthrough
		default:
			k = tcell.KeyESC
			ui.echoerrf("unknown key: %s", val)
		}

		return tcell.NewEventKey(k, ch, mod)
	case ev := <-ui.tevChan:
		return ev
	}
}

func addSpecialKeyModifier(val string, mod tcell.ModMask) string {
	switch {
	case !strings.HasPrefix(val, "<"):
		return val
	case mod == tcell.ModCtrl && !strings.HasPrefix(val, "<c-"):
		return "<c-" + val[1:]
	case mod == tcell.ModShift:
		return "<s-" + val[1:]
	case mod == tcell.ModAlt:
		return "<a-" + val[1:]
	default:
		return val
	}
}

// This function is used to read a normal event on the client side. For keys,
// digits are interpreted as command counts but this is only done for digits
// preceding any non-digit characters (e.g. "42y2k" as 42 times "y2k").
func (ui *ui) readNormalEvent(ev tcell.Event, nav *nav) expr {
	draw := &callExpr{"draw", nil, 1}
	count := 0

	keys := gOpts.nkeys
	mode := "n"
	if nav.isVisualMode() {
		keys = gOpts.vkeys
		mode = "v"
	}

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
			val = addSpecialKeyModifier(val, tev.Modifiers())
			if val == "<esc>" && len(ui.keyAcc) != 0 {
				ui.keyAcc = nil
				ui.keyCount = nil
				ui.menu = ""
				return draw
			}
			ui.keyAcc = append(ui.keyAcc, []rune(val)...)
		}

		if len(ui.keyAcc) == 0 {
			return draw
		}

		binds, ok := findBinds(keys, string(ui.keyAcc))

		switch len(binds) {
		case 0:
			ui.echoerrf("unknown mapping: %s", string(ui.keyAcc))
			ui.keyAcc = nil
			ui.keyCount = nil
			ui.menu = ""
			return draw
		default:
			if ok {
				if len(ui.keyCount) > 0 {
					c, err := strconv.Atoi(string(ui.keyCount))
					if err != nil {
						log.Printf("converting command count: %s", err)
					}
					count = c
				}
				expr := keys[string(ui.keyAcc)]

				if count != 0 {
					switch e := expr.(type) {
					case *callExpr:
						expr = &callExpr{name: e.name, args: e.args, count: count}
					case *listExpr:
						expr = &listExpr{exprs: e.exprs, count: count}
					}
				}

				ui.keyAcc = nil
				ui.keyCount = nil
				ui.menu = ""
				return expr
			}
			if gOpts.showbinds {
				ui.menu = listBinds(map[string]map[string]expr{
					mode: binds,
				})
			}
			return draw
		}
	case *tcell.EventMouse:
		if ui.cmdPrefix != "" {
			return nil
		}

		var button string

		switch tev.Buttons() {
		case tcell.Button1:
			button = "<m-1>"
		case tcell.Button2:
			button = "<m-2>"
		case tcell.Button3:
			button = "<m-3>"
		case tcell.Button4:
			button = "<m-4>"
		case tcell.Button5:
			button = "<m-5>"
		case tcell.Button6:
			button = "<m-6>"
		case tcell.Button7:
			button = "<m-7>"
		case tcell.Button8:
			button = "<m-8>"
		case tcell.WheelUp:
			button = "<m-up>"
		case tcell.WheelDown:
			button = "<m-down>"
		case tcell.WheelLeft:
			button = "<m-left>"
		case tcell.WheelRight:
			button = "<m-right>"
		case tcell.ButtonNone:
			return nil
		}
		if tev.Modifiers() == tcell.ModCtrl {
			button = "<c-" + button[1:]
		}
		if expr, ok := keys[button]; ok {
			return expr
		}
		if button != "<m-1>" && button != "<m-2>" {
			ui.echoerrf("unknown mapping: %s", button)
			ui.keyAcc = nil
			ui.keyCount = nil
			ui.menu = ""
			return draw
		}

		x, y := tev.Position()
		wind, w := ui.winAt(x, y)
		if wind == -1 {
			return nil
		}

		var dir *dir
		if gOpts.preview && wind == len(ui.wins)-1 {
			curr, err := nav.currFile()
			if err != nil {
				return nil
			} else if !curr.IsDir() || gOpts.dirpreviews {
				if tev.Buttons() != tcell.Button2 {
					return nil
				}
				return &callExpr{"open", nil, 1}
			}
			dir = ui.dirPrev
		} else {
			dir = ui.dirOfWin(nav, wind)
			if dir == nil {
				return nil
			}
		}

		var file *file
		ind := dir.ind - dir.pos + y - w.y
		if ind < len(dir.files) {
			file = dir.files[ind]
		}

		if file != nil {
			sel := &callExpr{"select", []string{file.path}, 1}

			if tev.Buttons() == tcell.Button1 {
				return sel
			}
			if file.IsDir() {
				return &callExpr{"cd", []string{file.path}, 1}
			}
			return &listExpr{[]expr{sel, &callExpr{"open", nil, 1}}, 1}
		}
		if tev.Buttons() == tcell.Button1 {
			return &callExpr{"cd", []string{dir.path}, 1}
		}
	case *tcell.EventResize:
		return &callExpr{"redraw", nil, 1}
	case *tcell.EventError:
		log.Printf("Got EventError: '%s' at %s", tev.Error(), tev.When())
	case *tcell.EventInterrupt:
		log.Printf("Got EventInterrupt: at %s", tev.When())
	case *tcell.EventFocus:
		if tev.Focused {
			return &callExpr{"on-focus-gained", nil, 1}
		} else {
			return &callExpr{"on-focus-lost", nil, 1}
		}
	}
	return nil
}

func readCmdEvent(ev tcell.Event) expr {
	switch tev := ev.(type) {
	case *tcell.EventKey:
		if tev.Key() == tcell.KeyRune {
			if tev.Modifiers() == tcell.ModMask(tcell.ModAlt) {
				val := string([]rune{'<', 'a', '-', tev.Rune(), '>'})
				if expr, ok := gOpts.cmdkeys[val]; ok {
					return expr
				}
			} else {
				return &callExpr{"cmd-insert", []string{string(tev.Rune())}, 1}
			}
		} else {
			val := gKeyVal[tev.Key()]
			val = addSpecialKeyModifier(val, tev.Modifiers())
			if expr, ok := gOpts.cmdkeys[val]; ok {
				return expr
			}
		}
	}
	return nil
}

func (ui *ui) readEvent(ev tcell.Event, nav *nav) expr {
	if ev == nil {
		return nil
	}

	if _, ok := ev.(*tcell.EventKey); ok && ui.cmdPrefix != "" {
		return readCmdEvent(ev)
	}

	return ui.readNormalEvent(ev, nav)
}

func (ui *ui) readExpr() {
	go func() {
		for {
			ui.evChan <- ui.pollEvent()
		}
	}()
}

func (ui *ui) suspend() error {
	ui.sxScreen.forceClear = true
	return ui.screen.Suspend()
}

func (ui *ui) resume() error {
	err := ui.screen.Resume()
	if !ui.polling {
		go ui.pollEvents()
		ui.polling = true
	}
	return err
}

func (ui *ui) exportSizes() {
	w, h := ui.screen.Size()
	os.Setenv("lf_width", strconv.Itoa(w))
	os.Setenv("lf_height", strconv.Itoa(h))
}

func anyKey() {
	fmt.Fprint(os.Stderr, gOpts.waitmsg)
	defer fmt.Fprint(os.Stderr, "\n")
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	b := make([]byte, 8)
	os.Stdin.Read(b)
}

func listMatches(screen tcell.Screen, matches []string, selectedInd int) string {
	mlen := len(matches)
	if mlen < 2 {
		return ""
	}

	var b strings.Builder

	wtot, _ := screen.Size()
	wcol := 0
	for _, m := range matches {
		wcol = max(wcol, len(m))
	}
	wcol += gOpts.tabstop - wcol%gOpts.tabstop
	ncol := max(wtot/wcol, 1)

	b.WriteString("possible matches\n")

	for i := 0; i < mlen; {
		for j := 0; j < ncol && i < mlen; i, j = i+1, j+1 {
			target := matches[i]

			if selectedInd == i {
				fmt.Fprintf(&b, "\033[7m%s\033[0m%*s", target, wcol-len(target), "")
			} else {
				fmt.Fprintf(&b, "%s%*s", target, wcol-len(target), "")
			}
		}
		b.WriteByte('\n')
	}

	return b.String()
}
