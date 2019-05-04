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

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

const gEscapeCode = 27

var gKeyVal = map[termbox.Key]string{
	termbox.KeyF1:             "<f-1>",
	termbox.KeyF2:             "<f-2>",
	termbox.KeyF3:             "<f-3>",
	termbox.KeyF4:             "<f-4>",
	termbox.KeyF5:             "<f-5>",
	termbox.KeyF6:             "<f-6>",
	termbox.KeyF7:             "<f-7>",
	termbox.KeyF8:             "<f-8>",
	termbox.KeyF9:             "<f-9>",
	termbox.KeyF10:            "<f-10>",
	termbox.KeyF11:            "<f-11>",
	termbox.KeyF12:            "<f-12>",
	termbox.KeyInsert:         "<insert>",
	termbox.KeyDelete:         "<delete>",
	termbox.KeyHome:           "<home>",
	termbox.KeyEnd:            "<end>",
	termbox.KeyPgup:           "<pgup>",
	termbox.KeyPgdn:           "<pgdn>",
	termbox.KeyArrowUp:        "<up>",
	termbox.KeyArrowDown:      "<down>",
	termbox.KeyArrowLeft:      "<left>",
	termbox.KeyArrowRight:     "<right>",
	termbox.KeyCtrlSpace:      "<c-space>",
	termbox.KeyCtrlA:          "<c-a>",
	termbox.KeyCtrlB:          "<c-b>",
	termbox.KeyCtrlC:          "<c-c>",
	termbox.KeyCtrlD:          "<c-d>",
	termbox.KeyCtrlE:          "<c-e>",
	termbox.KeyCtrlF:          "<c-f>",
	termbox.KeyCtrlG:          "<c-g>",
	termbox.KeyBackspace:      "<bs>",
	termbox.KeyTab:            "<tab>",
	termbox.KeyCtrlJ:          "<c-j>",
	termbox.KeyCtrlK:          "<c-k>",
	termbox.KeyCtrlL:          "<c-l>",
	termbox.KeyEnter:          "<enter>",
	termbox.KeyCtrlN:          "<c-n>",
	termbox.KeyCtrlO:          "<c-o>",
	termbox.KeyCtrlP:          "<c-p>",
	termbox.KeyCtrlQ:          "<c-q>",
	termbox.KeyCtrlR:          "<c-r>",
	termbox.KeyCtrlS:          "<c-s>",
	termbox.KeyCtrlT:          "<c-t>",
	termbox.KeyCtrlU:          "<c-u>",
	termbox.KeyCtrlV:          "<c-v>",
	termbox.KeyCtrlW:          "<c-w>",
	termbox.KeyCtrlX:          "<c-x>",
	termbox.KeyCtrlY:          "<c-y>",
	termbox.KeyCtrlZ:          "<c-z>",
	termbox.KeyEsc:            "<esc>",
	termbox.KeyCtrlBackslash:  "<c-\\>",
	termbox.KeyCtrlRsqBracket: "<c-]>",
	termbox.KeyCtrl6:          "<c-6>",
	termbox.KeyCtrlSlash:      "<c-/>",
	termbox.KeySpace:          "<space>",
	termbox.KeyBackspace2:     "<bs2>",
}

var gDirExactIcons = map[string]rune{

	  ".git"                             : '',
    "Desktop"                          : '',
    "Documents"                        : '',
    "Downloads"                        : '',
    "Dotfiles"                         : '',
    "Dropbox"                          : '',
    "Music"                            : '',
    "Pictures"                         : '',
    "Public"                           : '',
    "Templates"                        : '',
    "Videos"                           : '',
// Spanish
    "Escritorio"                       : '',
    "Documentos"                       : '',
    "Descargas"                        : '',
    "Música"                           : '',
    "Imágenes"                         : '',
    "Plantillas"                       : '',
// French
    "Bureau"                           : '',
    "Images"                           : '',
    "Musique"                          : '',
    "Publique"                         : '',
    "Téléchargements"                  : '',
    "Vidéos"                           : '',
// Portuguese
    "Imagens"                          : '',
    "Modelos"                          : '',
    "Público"                          : '',
    "Área de trabalho"                 : '',
// Italian
    "Documenti"                        : '',
    "Immagini"                         : '',
    "Pubblici"                         : '',
    "Scaricati"                        : '',
    "Scrivania"                        : '',
    "Video"                            : '',
// German
    "Bilder"                           : '',
    "Dokumente"                        : '',
    "Musik"                            : '',
    "Schreibtisch"                     : '',
    "Vorlagen"                         : '',
    "Öffentlich"                       : '',
// Hungarian
    "Dokumentumok"                     : '',
    "Képek"                            : '',
    "Modelli"                          : '',
    "Zene"                             : '',
    "Letöltések"                       : '',
    "Számítógép"                        : '',
		"Videók" 														: '',
}

var gExtensionIcons = map[string]rune{
		"7z"       : '',
    "a"        : '',
    "ai"       : '',
    "apk"      : '',
    "asm"      : '',
    "asp"      : '',
    "aup"      : '',
    "avi"      : '',
    "bat"      : '',
    "bmp"      : '',
    "bz2"      : '',
    "c"        : '',
    "c++"      : '',
    "cab"      : '',
    "cbr"      : '',
    "cbz"      : '',
    "cc"       : '',
    "class"    : '',
    "clj"      : '',
    "cljc"     : '',
    "cljs"     : '',
    "cmake"    : '',
    "coffee"   : '',
    "conf"     : '',
    "cp"       : '',
    "cpio"     : '',
    "cpp"      : '',
    "css"      : '',
    "cue"      : '',
    "cvs"      : '',
    "cxx"      : '',
    "d"        : '',
    "dart"     : '',
    "db"       : '',
    "deb"      : '',
    "diff"     : '',
    "dll"      : '',
    "doc"      : '',
    "docx"     : '',
    "dump"     : '',
    "edn"      : '',
    "efi"      : '',
    "ejs"      : '',
    "elf"      : '',
    "epub"     : '',
    "erl"      : '',
    "exe"      : '',
    "f#"       : '',
    "fifo"     : '|',
    "fish"     : '',
    "flac"     : '',
    "flv"      : '',
    "fs"       : '',
    "fsi"      : '',
    "fsscript" : '',
    "fsx"      : '',
    "gem"      : '',
    "gif"      : '',
    "go"       : '',
    "gz"       : '',
    "gzip"     : '',
    "h"        : '',
    "hbs"      : '',
    "hrl"      : '',
    "hs"       : '',
    "htaccess" : '',
    "htpasswd" : '',
    "htm"      : '',
    "html"     : '',
    "ico"      : '',
    "img"      : '',
    "ini"      : '',
    "iso"      : '',
    "jar"      : '',
    "java"     : '',
    "jl"       : '',
    "jpeg"     : '',
    "jpg"      : '',
    "js"       : '',
    "json"     : '',
    "jsx"      : '',
    "key"      : '',
    "less"     : '',
    "lha"      : '',
    "lhs"      : '',
    "log"      : '',
    "lua"      : '',
    "lzh"      : '',
    "lzma"     : '',
    "m4a"      : '',
    "m4v"      : '',
    "markdown" : '',
    "md"       : '',
    "mkv"      : '',
    "ml"       : 'λ',
    "mli"      : 'λ',
    "mov"      : '',
    "mp3"      : '',
    "mp4"      : '',
    "mpeg"     : '',
    "mpg"      : '',
    "msi"      : '',
    "mustache" : '',
    "o"        : '',
    "ogg"      : '',
    "pdf"      : '',
    "php"      : '',
    "pl"       : '',
    "pm"       : '',
    "png"      : '',
    "pub"      : '',
    "ppt"      : '',
    "pptx"     : '',
    "psb"      : '',
    "psd"      : '',
    "py"       : '',
    "pyc"      : '',
    "pyd"      : '',
    "pyo"      : '',
    "rar"      : '',
    "rb"       : '',
    "rc"       : '',
    "rlib"     : '',
    "rom"      : '',
    "rpm"      : '',
    "rs"       : '',
    "rss"      : '',
    "rtf"      : '',
    "s"        : '',
    "so"       : '',
    "scala"    : '',
    "scss"     : '',
    "sh"       : '',
    "slim"     : '',
    "sln"      : '',
    "sql"      : '',
    "styl"     : '',
    "suo"      : '',
    "t"        : '',
    "tar"      : '',
    "tgz"      : '',
    "ts"       : '',
    "twig"     : '',
    "vim"      : '',
    "vimrc"    : '',
    "wav"      : '',
    "webm"     : '',
    "xbps"     : '',
    "xhtml"    : '',
    "xls"      : '',
    "xlsx"     : '',
    "xml"      : '',
    "xul"      : '',
    "xz"       : '',
    "yaml"     : '',
    "yml"      : '',
		"zip" 		 : '',
}

var gValKey map[string]termbox.Key

func init() {
	gValKey = make(map[string]termbox.Key)
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

	if reg.loading {
		fg = termbox.AttrReverse
		win.print(2, 0, fg, bg, "loading...")
		return
	}

	for i, l := range reg.lines {
		fg, bg = win.print(2, i, fg, bg, l)
	}

	return
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
		default:
			log.Printf("unknown info type: %s", s)
		}
	}

	return info
}

func getEmoji(f *file) rune{

	if f.IsDir() {
		if val, ok := gDirExactIcons[f.Name()]; ok{
			return val
		}else{
			return ''
		}
		
	}else{
		li := strings.LastIndex(f.Name(),".")
		if li == -1{
			return ''
		}
		var ext = f.Name()[li+1:]
		if icon, ok := gExtensionIcons[ext]; ok{
			return icon
		}else{
			return ''
		}

	}

}

func (win *win) printDir(dir *dir, selections map[string]int, saves map[string]bool, colors colorMap) {
	if win.w < 5 || dir == nil {
		return
	}

	if dir.loading {
		win.print(2, 0, termbox.AttrReverse, termbox.ColorDefault, "loading...")
		return
	}

	if len(dir.files) == 0 {
		win.print(2, 0, termbox.AttrReverse, termbox.ColorDefault, "empty")
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
		for j := 10; j < len(dir.files); j *= 10 {
			lnwidth++
		}
		lnformat = fmt.Sprintf("%%%d.d ", lnwidth)
	}

	for i, f := range dir.files[beg:end] {
		fg, bg := colors.get(f)

		if lnwidth > 0 {
			var ln string

			if gOpts.number && (!gOpts.relativenumber || i == dir.pos) {
				ln = fmt.Sprintf(lnformat, i+1+beg)
			} else if gOpts.relativenumber {
				if i < dir.pos {
					ln = fmt.Sprintf(lnformat, dir.pos-i)
				} else {
					ln = fmt.Sprintf(lnformat, i-dir.pos)
				}
			}

			win.print(0, i, termbox.ColorYellow, bg, ln)
		}

		path := filepath.Join(dir.path, f.Name())

		if _, ok := selections[path]; ok {
			win.print(lnwidth, i, fg, termbox.ColorMagenta, " ")
		} else if cp, ok := saves[path]; ok {
			if cp {
				win.print(lnwidth, i, fg, termbox.ColorYellow, " ")
			} else {
				win.print(lnwidth, i, fg, termbox.ColorRed, " ")
			}
		}

		if i == dir.pos {
			fg |= termbox.AttrReverse
		}

		var s []rune

		s = append(s, ' ')

		if gOpts.icons{
			var emoji = getEmoji(f)
			s = append(s, emoji);
			s = append(s, ' ')
		}

		for _, r := range f.Name() {
			s = append(s, r)
		}

		w := runeSliceWidth(s)

		if w > win.w-3 {
			s = runeSliceWidthRange(s, 0, win.w-4)
			s = append(s, '~')
		} else {
			for i := 0; i < win.w-3-w; i++ {
				s = append(s, ' ')
			}
		}

		info := fileInfo(f, dir)

		if len(info) > 0 && win.w-2 > 2*len(info) {
			if win.w-2 > w+len(info) {
				s = runeSliceWidthRange(s, 0, win.w-3-len(info))
			} else {
				s = runeSliceWidthRange(s, 0, win.w-4-len(info))
				s = append(s, '~')
			}
			for _, r := range info {
				s = append(s, r)
			}
		}

		s = append(s, ' ')

		win.print(lnwidth+1, i, fg, bg, string(s))
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
	exprChan    chan expr
	keyChan     chan string
	evChan      chan termbox.Event
	menuBuf     *bytes.Buffer
	cmdPrefix   string
	cmdAccLeft  []rune
	cmdAccRight []rune
	cmdYankBuf  []rune
	keyAcc      []rune
	keyCount    []rune
	colors      colorMap
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

func getWins() []*win {
	wtot, htot := termbox.Size()

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

func newUI() *ui {
	wtot, htot := termbox.Size()

	evQueue := make(chan termbox.Event)
	go func() {
		for {
			evQueue <- termbox.PollEvent()
		}
	}()

	evChan := make(chan termbox.Event)
	go func() {
		for {
			ev := <-evQueue
			if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
				select {
				case ev2 := <-evQueue:
					ev2.Mod = termbox.ModAlt
					evChan <- ev2
					continue
				case <-time.After(100 * time.Millisecond):
				}
			}
			evChan <- ev
		}
	}()

	return &ui{
		wins:      getWins(),
		promptWin: newWin(wtot, 1, 0, 0),
		msgWin:    newWin(wtot, 1, 0, htot-1),
		menuWin:   newWin(wtot, 1, 0, htot-2),
		keyChan:   make(chan string, 1000),
		evChan:    evChan,
		colors:    parseColors(),
	}
}

func (ui *ui) renew() {
	wtot, htot := termbox.Size()

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
		ui.regPrev = nav.loadReg(ui, curr.path)
	}
}

func (ui *ui) loadFileInfo(nav *nav) {
	curr, err := nav.currFile()
	if err != nil {
		return
	}

	ui.echof("%v %4s %v", curr.Mode(), humanize(curr.Size()), curr.ModTime().Format(gOpts.timefmt))
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

	tot := len(currDir.files)
	ind := min(currDir.ind+1, tot)
	acc := string(ui.keyCount) + string(ui.keyAcc)

	var progress string

	if nav.copyTotal > 0 {
		percentage := int((100 * float64(nav.copyBytes)) / float64(nav.copyTotal))
		progress += fmt.Sprintf("  [%d%%]", percentage)
	}

	if nav.moveTotal > 0 {
		progress += fmt.Sprintf("  [%d/%d]", nav.moveCount, nav.moveTotal)
	}

	ruler := fmt.Sprintf("%s%s  %d/%d", acc, progress, ind, tot)

	ui.msgWin.printRight(0, fg, bg, ruler)
}

func (ui *ui) drawBox(nav *nav) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault

	w, h := termbox.Size()

	for i := 1; i < w-1; i++ {
		termbox.SetCell(i, 1, '─', fg, bg)
		termbox.SetCell(i, h-2, '─', fg, bg)
	}

	for i := 2; i < h-2; i++ {
		termbox.SetCell(0, i, '│', fg, bg)
		termbox.SetCell(w-1, i, '│', fg, bg)
	}

	termbox.SetCell(0, 1, '┌', fg, bg)
	termbox.SetCell(w-1, 1, '┐', fg, bg)
	termbox.SetCell(0, h-2, '└', fg, bg)
	termbox.SetCell(w-1, h-2, '┘', fg, bg)

	wacc := 0
	for wind := 0; wind < len(ui.wins)-1; wind++ {
		wacc += ui.wins[wind].w
		termbox.SetCell(wacc, 1, '┬', fg, bg)
		for i := 2; i < h-2; i++ {
			termbox.SetCell(wacc, i, '│', fg, bg)
		}
		termbox.SetCell(wacc, h-2, '┴', fg, bg)
	}
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
		ui.wins[woff+i].printDir(nav.dirs[doff+i], nav.selections, nav.saves, ui.colors)
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
				preview.printDir(ui.dirPrev, nav.selections, nav.saves, ui.colors)
			} else if f.Mode().IsRegular() {
				preview.printReg(ui.regPrev)
			}
		}
	}

	if gOpts.drawbox {
		ui.drawBox(nav)
	}

	if ui.menuBuf != nil {
		lines := strings.Split(ui.menuBuf.String(), "\n")

		lines = lines[:len(lines)-1]

		ui.menuWin.h = len(lines) - 1
		ui.menuWin.y = ui.wins[0].h - ui.menuWin.h

		if gOpts.drawbox {
			ui.menuWin.y += 2
		}

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

func (ui *ui) pollEvent() termbox.Event {
	select {
	case key := <-ui.keyChan:
		ev := termbox.Event{Type: termbox.EventKey}

		if len(key) == 1 {
			ev.Ch, _ = utf8.DecodeRuneInString(key)
		} else {
			switch {
			case key == "<lt>":
				ev.Ch = '<'
			case key == "<gt>":
				ev.Ch = '>'
			case reAltKey.MatchString(key):
				match := reAltKey.FindStringSubmatch(key)[1]
				ev.Ch, _ = utf8.DecodeRuneInString(match)
				ev.Mod = termbox.ModAlt
			default:
				if val, ok := gValKey[key]; ok {
					ev.Key = val
				} else {
					ev.Key = termbox.KeyEsc
					ui.echoerrf("unknown key: %s", key)
				}
			}
		}

		return ev
	case ev := <-ui.evChan:
		return ev
	}
}

// This function is used to read a normal event on the client side. For keys,
// digits are interpreted as command counts but this is only done for digits
// preceding any non-digit characters (e.g. "42y2k" as 42 times "y2k").
func (ui *ui) readEvent(ch chan<- expr, ev termbox.Event) {
	draw := &callExpr{"draw", nil, 1}
	count := 1

	switch ev.Type {
	case termbox.EventKey:
		if ev.Ch != 0 {
			switch {
			case ev.Ch == '<':
				ui.keyAcc = append(ui.keyAcc, []rune("<lt>")...)
			case ev.Ch == '>':
				ui.keyAcc = append(ui.keyAcc, []rune("<gt>")...)
			case ev.Mod == termbox.ModAlt:
				ui.keyAcc = append(ui.keyAcc, '<', 'a', '-', ev.Ch, '>')
			case unicode.IsDigit(ev.Ch) && len(ui.keyAcc) == 0:
				ui.keyCount = append(ui.keyCount, ev.Ch)
			default:
				ui.keyAcc = append(ui.keyAcc, ev.Ch)
			}
		} else {
			val := gKeyVal[ev.Key]
			if val == "<esc>" {
				ch <- draw
				ui.keyAcc = nil
				ui.keyCount = nil
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
		case 1:
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
				}
				ch <- expr
				ui.keyAcc = nil
				ui.keyCount = nil
				ui.menuBuf = nil
			} else {
				ui.menuBuf = listBinds(binds)
				ch <- draw
			}
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
	case termbox.EventResize:
		ch <- &callExpr{"redraw", nil, 1}
	}
}

func readCmdEvent(ch chan<- expr, ev termbox.Event) {
	if ev.Ch != 0 {
		if ev.Mod == termbox.ModAlt {
			val := string([]rune{'<', 'a', '-', ev.Ch, '>'})
			if expr, ok := gOpts.cmdkeys[val]; ok {
				ch <- expr
			}
		} else {
			ch <- &callExpr{"cmd-insert", []string{string(ev.Ch)}, 1}
		}
	} else {
		val := gKeyVal[ev.Key]
		if expr, ok := gOpts.cmdkeys[val]; ok {
			ch <- expr
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

			if ui.cmdPrefix != "" && ev.Type == termbox.EventKey {
				readCmdEvent(ch, ev)
				continue
			}

			ui.readEvent(ch, ev)
		}
	}()

	return ch
}

func setColorMode() {
	if gOpts.color256 {
		termbox.SetOutputMode(termbox.Output256)
	} else {
		termbox.SetOutputMode(termbox.OutputNormal)
	}
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
