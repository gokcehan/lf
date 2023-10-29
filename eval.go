package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

func (e *setExpr) eval(app *app, args []string) {
	switch e.opt {
	case "anchorfind":
		if e.val == "" || e.val == "true" {
			gOpts.anchorfind = true
		} else if e.val == "false" {
			gOpts.anchorfind = false
		} else {
			app.ui.echoerr("anchorfind: value should be empty, 'true', or 'false'")
			return
		}
	case "noanchorfind":
		if e.val != "" {
			app.ui.echoerrf("noanchorfind: unexpected value: %s", e.val)
			return
		}
		gOpts.anchorfind = false
	case "anchorfind!":
		if e.val != "" {
			app.ui.echoerrf("anchorfind!: unexpected value: %s", e.val)
			return
		}
		gOpts.anchorfind = !gOpts.anchorfind
	case "autoquit":
		if e.val == "" || e.val == "true" {
			gOpts.autoquit = true
		} else if e.val == "false" {
			gOpts.autoquit = false
		} else {
			app.ui.echoerr("autoquit: value should be empty, 'true', or 'false'")
			return
		}
	case "noautoquit":
		if e.val != "" {
			app.ui.echoerrf("noautoquit: unexpected value: %s", e.val)
			return
		}
		gOpts.autoquit = false
	case "autoquit!":
		if e.val != "" {
			app.ui.echoerrf("autoquit!: unexpected value: %s", e.val)
			return
		}
		gOpts.autoquit = !gOpts.autoquit
	case "borderfmt":
		gOpts.borderfmt = e.val
	case "cleaner":
		gOpts.cleaner = replaceTilde(e.val)
	case "cursoractivefmt":
		gOpts.cursoractivefmt = e.val
	case "cursorparentfmt":
		gOpts.cursorparentfmt = e.val
	case "cursorpreviewfmt":
		gOpts.cursorpreviewfmt = e.val
	case "dircache":
		if e.val == "" || e.val == "true" {
			gOpts.dircache = true
		} else if e.val == "false" {
			gOpts.dircache = false
		} else {
			app.ui.echoerr("dircache: value should be empty, 'true', or 'false'")
			return
		}
	case "nodircache":
		if e.val != "" {
			app.ui.echoerrf("nodircache: unexpected value: %s", e.val)
			return
		}
		gOpts.dircache = false
	case "dircache!":
		if e.val != "" {
			app.ui.echoerrf("dircache!: unexpected value: %s", e.val)
			return
		}
		gOpts.dircache = !gOpts.dircache
	case "dircounts":
		if e.val == "" || e.val == "true" {
			gOpts.dircounts = true
		} else if e.val == "false" {
			gOpts.dircounts = false
		} else {
			app.ui.echoerr("dircounts: value should be empty, 'true', or 'false'")
			return
		}
	case "nodircounts":
		if e.val != "" {
			app.ui.echoerrf("nodircounts: unexpected value: %s", e.val)
			return
		}
		gOpts.dircounts = false
	case "dircounts!":
		if e.val != "" {
			app.ui.echoerrf("dircounts!: unexpected value: %s", e.val)
			return
		}
		gOpts.dircounts = !gOpts.dircounts
	case "dirfirst":
		if e.val == "" || e.val == "true" {
			gOpts.sortType.option |= dirfirstSort
		} else if e.val == "false" {
			gOpts.sortType.option &= ^dirfirstSort
		} else {
			app.ui.echoerr("dirfirst: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.ui.sort()
	case "nodirfirst":
		if e.val != "" {
			app.ui.echoerrf("nodirfirst: unexpected value: %s", e.val)
			return
		}
		gOpts.sortType.option &= ^dirfirstSort
		app.nav.sort()
		app.ui.sort()
	case "dirfirst!":
		if e.val != "" {
			app.ui.echoerrf("dirfirst!: unexpected value: %s", e.val)
			return
		}
		gOpts.sortType.option ^= dirfirstSort
		app.nav.sort()
		app.ui.sort()
	case "dironly":
		if e.val == "" || e.val == "true" {
			gOpts.dironly = true
		} else if e.val == "false" {
			gOpts.dironly = false
		} else {
			app.ui.echoerr("dironly: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "nodironly":
		if e.val != "" {
			app.ui.echoerrf("nodironly: unexpected value: %s", e.val)
			return
		}
		gOpts.dironly = false
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "dironly!":
		if e.val != "" {
			app.ui.echoerrf("dironly!: unexpected value: %s", e.val)
			return
		}
		gOpts.dironly = !gOpts.dironly
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "dirpreviews":
		if e.val == "" || e.val == "true" {
			gOpts.dirpreviews = true
		} else if e.val == "false" {
			gOpts.dirpreviews = false
		} else {
			app.ui.echoerr("dirpreviews: value should be empty, 'true', or 'false'")
			return
		}
	case "nodirpreviews":
		if e.val != "" {
			app.ui.echoerrf("nodirpreviews: unexpected value: %s", e.val)
			return
		}
		gOpts.dirpreviews = false
	case "dirpreviews!":
		if e.val != "" {
			app.ui.echoerrf("dirpreviews!: unexpected value: %s", e.val)
			return
		}
		gOpts.dirpreviews = !gOpts.dirpreviews
	case "drawbox":
		if e.val == "" || e.val == "true" {
			gOpts.drawbox = true
		} else if e.val == "false" {
			gOpts.drawbox = false
		} else {
			app.ui.echoerr("drawbox: value should be empty, 'true', or 'false'")
			return
		}
		app.ui.renew()
		if app.nav.height != app.ui.wins[0].h {
			app.nav.height = app.ui.wins[0].h
			app.nav.regCache = make(map[string]*reg)
		}
		app.ui.loadFile(app, true)
	case "nodrawbox":
		if e.val != "" {
			app.ui.echoerrf("nodrawbox: unexpected value: %s", e.val)
			return
		}
		gOpts.drawbox = false
		app.ui.renew()
		if app.nav.height != app.ui.wins[0].h {
			app.nav.height = app.ui.wins[0].h
			app.nav.regCache = make(map[string]*reg)
		}
		app.ui.loadFile(app, true)
	case "drawbox!":
		if e.val != "" {
			app.ui.echoerrf("drawbox!: unexpected value: %s", e.val)
			return
		}
		gOpts.drawbox = !gOpts.drawbox
		app.ui.renew()
		if app.nav.height != app.ui.wins[0].h {
			app.nav.height = app.ui.wins[0].h
			app.nav.regCache = make(map[string]*reg)
		}
		app.ui.loadFile(app, true)
	case "dupfilefmt":
		gOpts.dupfilefmt = e.val
	case "errorfmt":
		gOpts.errorfmt = e.val
	case "filesep":
		gOpts.filesep = e.val
	case "findlen":
		n, err := strconv.Atoi(e.val)
		if err != nil {
			app.ui.echoerrf("findlen: %s", err)
			return
		}
		if n < 0 {
			app.ui.echoerr("findlen: value should be a non-negative number")
			return
		}
		gOpts.findlen = n
	case "globsearch":
		if e.val == "" || e.val == "true" {
			gOpts.globsearch = true
		} else if e.val == "false" {
			gOpts.globsearch = false
		} else {
			app.ui.echoerr("globsearch: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "noglobsearch":
		if e.val != "" {
			app.ui.echoerrf("noglobsearch: unexpected value: %s", e.val)
			return
		}
		gOpts.globsearch = false
		app.nav.sort()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "globsearch!":
		if e.val != "" {
			app.ui.echoerrf("globsearch!: unexpected value: %s", e.val)
			return
		}
		gOpts.globsearch = !gOpts.globsearch
		app.nav.sort()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "hidden":
		if e.val == "" || e.val == "true" {
			gOpts.sortType.option |= hiddenSort
		} else if e.val == "false" {
			gOpts.sortType.option &= ^hiddenSort
		} else {
			app.ui.echoerr("hidden: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "nohidden":
		if e.val != "" {
			app.ui.echoerrf("nohidden: unexpected value: %s", e.val)
			return
		}
		gOpts.sortType.option &= ^hiddenSort
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "hidden!":
		if e.val != "" {
			app.ui.echoerrf("hidden!: unexpected value: %s", e.val)
			return
		}
		gOpts.sortType.option ^= hiddenSort
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "hiddenfiles":
		toks := strings.Split(e.val, ":")
		for _, s := range toks {
			if s == "" {
				app.ui.echoerr("hiddenfiles: glob should be non-empty")
				return
			}
			_, err := filepath.Match(s, "a")
			if err != nil {
				app.ui.echoerrf("hiddenfiles: %s", err)
				return
			}
		}
		gOpts.hiddenfiles = toks
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "history":
		if e.val == "" || e.val == "true" {
			gOpts.history = true
		} else if e.val == "false" {
			gOpts.history = false
		} else {
			app.ui.echoerr("history: value should be empty, 'true', or 'false'")
			return
		}
	case "nohistory":
		if e.val != "" {
			app.ui.echoerrf("nohistory: unexpected value: %s", e.val)
			return
		}
		gOpts.history = false
	case "history!":
		if e.val != "" {
			app.ui.echoerrf("history!: unexpected value: %s", e.val)
			return
		}
		gOpts.history = !gOpts.history
	case "icons":
		if e.val == "" || e.val == "true" {
			gOpts.icons = true
		} else if e.val == "false" {
			gOpts.icons = false
		} else {
			app.ui.echoerr("icons: value should be empty, 'true', or 'false'")
			return
		}
	case "noicons":
		if e.val != "" {
			app.ui.echoerrf("noicons: unexpected value: %s", e.val)
			return
		}
		gOpts.icons = false
	case "icons!":
		if e.val != "" {
			app.ui.echoerrf("icons!: unexpected value: %s", e.val)
			return
		}
		gOpts.icons = !gOpts.icons
	case "ifs":
		gOpts.ifs = e.val
	case "ignorecase":
		if e.val == "" || e.val == "true" {
			gOpts.ignorecase = true
		} else if e.val == "false" {
			gOpts.ignorecase = false
		} else {
			app.ui.echoerr("ignorecase: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "noignorecase":
		if e.val != "" {
			app.ui.echoerrf("noignorecase: unexpected value: %s", e.val)
			return
		}
		gOpts.ignorecase = false
		app.nav.sort()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "ignorecase!":
		if e.val != "" {
			app.ui.echoerrf("ignorecase!: unexpected value: %s", e.val)
			return
		}
		gOpts.ignorecase = !gOpts.ignorecase
		app.nav.sort()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "ignoredia":
		if e.val == "" || e.val == "true" {
			gOpts.ignoredia = true
		} else if e.val == "false" {
			gOpts.ignoredia = false
		} else {
			app.ui.echoerr("ignoredia: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.ui.sort()
	case "noignoredia":
		if e.val != "" {
			app.ui.echoerrf("noignoredia: unexpected value: %s", e.val)
			return
		}
		gOpts.ignoredia = false
		app.nav.sort()
		app.ui.sort()
	case "ignoredia!":
		if e.val != "" {
			app.ui.echoerrf("ignoredia!: unexpected value: %s", e.val)
			return
		}
		gOpts.ignoredia = !gOpts.ignoredia
		app.nav.sort()
		app.ui.sort()
	case "incfilter":
		if e.val == "" || e.val == "true" {
			gOpts.incfilter = true
		} else if e.val == "false" {
			gOpts.incfilter = false
		} else {
			app.ui.echoerr("incfilter: value should be empty, 'true', or 'false'")
			return
		}
	case "noincfilter":
		if e.val != "" {
			app.ui.echoerrf("noincfilter: unexpected value: %s", e.val)
			return
		}
		gOpts.incfilter = false
	case "incfilter!":
		if e.val != "" {
			app.ui.echoerrf("incfilter!: unexpected value: %s", e.val)
			return
		}
		gOpts.incfilter = !gOpts.incfilter
	case "incsearch":
		if e.val == "" || e.val == "true" {
			gOpts.incsearch = true
		} else if e.val == "false" {
			gOpts.incsearch = false
		} else {
			app.ui.echoerr("incsearch: value should be empty, 'true', or 'false'")
			return
		}
	case "noincsearch":
		if e.val != "" {
			app.ui.echoerrf("noincsearch: unexpected value: %s", e.val)
			return
		}
		gOpts.incsearch = false
	case "incsearch!":
		if e.val != "" {
			app.ui.echoerrf("incsearch!: unexpected value: %s", e.val)
			return
		}
		gOpts.incsearch = !gOpts.incsearch
	case "info":
		if e.val == "" {
			gOpts.info = nil
			return
		}
		toks := strings.Split(e.val, ":")
		for _, s := range toks {
			switch s {
			case "size", "time", "atime", "ctime":
			default:
				app.ui.echoerr("info: should consist of 'size', 'time', 'atime' or 'ctime' separated with colon")
				return
			}
		}
		gOpts.info = toks
	case "ruler":
		if e.val == "" {
			gOpts.ruler = nil
			return
		}
		toks := strings.Split(e.val, ":")
		for _, s := range toks {
			switch s {
			case "df", "acc", "progress", "selection", "filter", "ind":
			default:
				if !strings.HasPrefix(s, "lf_") {
					app.ui.echoerr("ruler: should consist of 'df', 'acc', 'progress', 'selection', 'filter', 'ind' or 'lf_<option_name>' separated with colon")
					return
				}
			}
		}
		gOpts.ruler = toks
		app.ui.echoerr("option 'ruler' is deprecated, use 'rulerfmt' instead")
		return
	case "rulerfmt":
		gOpts.rulerfmt = e.val
	case "preserve":
		if e.val == "" {
			gOpts.preserve = nil
			return
		}
		toks := strings.Split(e.val, ":")
		for _, s := range toks {
			switch s {
			case "mode", "timestamps":
			default:
				app.ui.echoerr("preserve: should consist of 'mode' or 'timestamps separated with colon")
				return
			}
		}
		gOpts.preserve = toks
	case "infotimefmtnew":
		gOpts.infotimefmtnew = e.val
	case "infotimefmtold":
		gOpts.infotimefmtold = e.val
	case "mouse":
		if e.val == "" || e.val == "true" {
			if !gOpts.mouse {
				gOpts.mouse = true
				app.ui.screen.EnableMouse(tcell.MouseButtonEvents)
			}
		} else if e.val == "false" {
			if gOpts.mouse {
				gOpts.mouse = false
				app.ui.screen.DisableMouse()
			}
		} else {
			app.ui.echoerr("mouse: value should be empty, 'true', or 'false'")
			return
		}
	case "nomouse":
		if e.val != "" {
			app.ui.echoerrf("nomouse: unexpected value: %s", e.val)
			return
		}
		if gOpts.mouse {
			gOpts.mouse = false
			app.ui.screen.DisableMouse()
		}
	case "mouse!":
		if e.val != "" {
			app.ui.echoerrf("mouse!: unexpected value: %s", e.val)
			return
		}
		if gOpts.mouse {
			gOpts.mouse = false
			app.ui.screen.DisableMouse()
		} else {
			gOpts.mouse = true
			app.ui.screen.EnableMouse(tcell.MouseButtonEvents)
		}
	case "number":
		if e.val == "" || e.val == "true" {
			gOpts.number = true
		} else if e.val == "false" {
			gOpts.number = false
		} else {
			app.ui.echoerr("number: value should be empty, 'true', or 'false'")
			return
		}
	case "nonumber":
		if e.val != "" {
			app.ui.echoerrf("nonumber: unexpected value: %s", e.val)
			return
		}
		gOpts.number = false
	case "number!":
		if e.val != "" {
			app.ui.echoerrf("number!: unexpected value: %s", e.val)
			return
		}
		gOpts.number = !gOpts.number
	case "numberfmt":
		gOpts.numberfmt = e.val
	case "period":
		n, err := strconv.Atoi(e.val)
		if err != nil {
			app.ui.echoerrf("period: %s", err)
			return
		}
		if n < 0 {
			app.ui.echoerr("period: value should be a non-negative number")
			return
		}
		gOpts.period = n
		if n == 0 {
			app.ticker.Stop()
		} else {
			app.ticker.Stop()
			app.ticker = time.NewTicker(time.Duration(gOpts.period) * time.Second)
		}
	case "preview":
		if e.val == "" || e.val == "true" {
			if len(gOpts.ratios) < 2 {
				app.ui.echoerr("preview: 'ratios' should consist of at least two numbers before enabling 'preview'")
				return
			}
			gOpts.preview = true
		} else if e.val == "false" {
			gOpts.preview = false
		} else {
			app.ui.echoerr("preview: value should be empty, 'true', or 'false'")
			return
		}
		app.ui.loadFile(app, true)
	case "nopreview":
		if e.val != "" {
			app.ui.echoerrf("nopreview: unexpected value: %s", e.val)
			return
		}
		gOpts.preview = false
		app.ui.loadFile(app, true)
	case "preview!":
		if e.val != "" {
			app.ui.echoerrf("preview!: unexpected value: %s", e.val)
			return
		}
		if len(gOpts.ratios) < 2 {
			app.ui.echoerr("preview: 'ratios' should consist of at least two numbers before enabling 'preview'")
			return
		}
		gOpts.preview = !gOpts.preview
		app.ui.loadFile(app, true)
	case "previewer":
		gOpts.previewer = replaceTilde(e.val)
	case "promptfmt":
		gOpts.promptfmt = e.val
	case "ratios":
		toks := strings.Split(e.val, ":")
		var rats []int
		for _, s := range toks {
			n, err := strconv.Atoi(s)
			if err != nil {
				app.ui.echoerrf("ratios: %s", err)
				return
			}
			if n <= 0 {
				app.ui.echoerr("ratios: value should be a positive number")
				return
			}
			rats = append(rats, n)
		}
		if gOpts.preview && len(rats) < 2 {
			app.ui.echoerr("ratios: should consist of at least two numbers when 'preview' is enabled")
			return
		}
		gOpts.ratios = rats
		app.ui.wins = getWins(app.ui.screen)
		app.ui.loadFile(app, true)
	case "relativenumber":
		if e.val == "" || e.val == "true" {
			gOpts.relativenumber = true
		} else if e.val == "false" {
			gOpts.relativenumber = false
		} else {
			app.ui.echoerr("relativenumber: value should be empty, 'true', or 'false'")
			return
		}
	case "norelativenumber":
		if e.val != "" {
			app.ui.echoerrf("norelativenumber: unexpected value: %s", e.val)
			return
		}
		gOpts.relativenumber = false
	case "relativenumber!":
		if e.val != "" {
			app.ui.echoerrf("relativenumber!: unexpected value: %s", e.val)
			return
		}
		gOpts.relativenumber = !gOpts.relativenumber
	case "reverse":
		if e.val == "" || e.val == "true" {
			gOpts.sortType.option |= reverseSort
		} else if e.val == "false" {
			gOpts.sortType.option &= ^reverseSort
		} else {
			app.ui.echoerr("reverse: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.ui.sort()
	case "noreverse":
		if e.val != "" {
			app.ui.echoerrf("noreverse: unexpected value: %s", e.val)
			return
		}
		gOpts.sortType.option &= ^reverseSort
		app.nav.sort()
		app.ui.sort()
	case "reverse!":
		if e.val != "" {
			app.ui.echoerrf("reverse!: unexpected value: %s", e.val)
			return
		}
		gOpts.sortType.option ^= reverseSort
		app.nav.sort()
		app.ui.sort()
	case "scrolloff":
		n, err := strconv.Atoi(e.val)
		if err != nil {
			app.ui.echoerrf("scrolloff: %s", err)
			return
		}
		if n < 0 {
			app.ui.echoerr("scrolloff: value should be a non-negative number")
			return
		}
		gOpts.scrolloff = n
	case "selmode":
		switch e.val {
		case "all", "dir":
			gOpts.selmode = e.val
		default:
			app.ui.echoerr("selmode: value should either be 'all' or 'dir'")
			return
		}
	case "shell":
		gOpts.shell = e.val
	case "shellflag":
		gOpts.shellflag = e.val
	case "shellopts":
		if e.val == "" {
			gOpts.shellopts = nil
			return
		}
		gOpts.shellopts = strings.Split(e.val, ":")
	case "smartcase":
		if e.val == "" || e.val == "true" {
			gOpts.smartcase = true
		} else if e.val == "false" {
			gOpts.smartcase = false
		} else {
			app.ui.echoerr("smartcase: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "nosmartcase":
		if e.val != "" {
			app.ui.echoerrf("nosmartcase: unexpected value: %s", e.val)
			return
		}
		gOpts.smartcase = false
		app.nav.sort()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "smartcase!":
		if e.val != "" {
			app.ui.echoerrf("smartcase!: unexpected value: %s", e.val)
			return
		}
		gOpts.smartcase = !gOpts.smartcase
		app.nav.sort()
		app.ui.sort()
		app.ui.loadFile(app, true)
	case "smartdia":
		if e.val == "" || e.val == "true" {
			gOpts.smartdia = true
		} else if e.val == "false" {
			gOpts.smartdia = false
		} else {
			app.ui.echoerr("smartdia: value should be empty, 'true', or 'false'")
			return
		}
	case "nosmartdia":
		if e.val != "" {
			app.ui.echoerrf("nosmartdia: unexpected value: %s", e.val)
			return
		}
		gOpts.smartdia = false
	case "smartdia!":
		if e.val != "" {
			app.ui.echoerrf("smartdia!: unexpected value: %s", e.val)
			return
		}
		gOpts.smartdia = !gOpts.smartdia
	case "sortby":
		switch e.val {
		case "natural":
			gOpts.sortType.method = naturalSort
		case "name":
			gOpts.sortType.method = nameSort
		case "size":
			gOpts.sortType.method = sizeSort
		case "time":
			gOpts.sortType.method = timeSort
		case "ctime":
			gOpts.sortType.method = ctimeSort
		case "atime":
			gOpts.sortType.method = atimeSort
		case "ext":
			gOpts.sortType.method = extSort
		default:
			app.ui.echoerr("sortby: value should either be 'natural', 'name', 'size', 'time', 'atime', 'ctime' or 'ext'")
			return
		}
		app.nav.sort()
		app.ui.sort()
	case "statfmt":
		gOpts.statfmt = e.val
	case "tabstop":
		n, err := strconv.Atoi(e.val)
		if err != nil {
			app.ui.echoerrf("tabstop: %s", err)
			return
		}
		if n <= 0 {
			app.ui.echoerr("tabstop: value should be a positive number")
			return
		}
		gOpts.tabstop = n
	case "tagfmt":
		gOpts.tagfmt = e.val
	case "tempmarks":
		if e.val != "" {
			gOpts.tempmarks = "'" + e.val
		} else {
			gOpts.tempmarks = "'"
		}
	case "timefmt":
		gOpts.timefmt = e.val
	case "truncatechar":
		if runeSliceWidth([]rune(e.val)) != 1 {
			app.ui.echoerr("truncatechar: value should be a single character")
			return
		}

		gOpts.truncatechar = e.val
	case "truncatepct":
		n, err := strconv.Atoi(e.val)
		if err != nil {
			app.ui.echoerrf("truncatepct: %s", err)
			return
		}
		if n < 0 || n > 100 {
			app.ui.echoerrf("truncatepct: must be between 0 and 100 (both inclusive), got %d", n)
			return
		}
		gOpts.truncatepct = n
	case "waitmsg":
		gOpts.waitmsg = e.val
	case "wrapscan":
		if e.val == "" || e.val == "true" {
			gOpts.wrapscan = true
		} else if e.val == "false" {
			gOpts.wrapscan = false
		} else {
			app.ui.echoerr("wrapscan: value should be empty, 'true', or 'false'")
			return
		}
	case "nowrapscan":
		if e.val != "" {
			app.ui.echoerrf("nowrapscan: unexpected value: %s", e.val)
			return
		}
		gOpts.wrapscan = false
	case "wrapscan!":
		if e.val != "" {
			app.ui.echoerrf("wrapscan!: unexpected value: %s", e.val)
			return
		}
		gOpts.wrapscan = !gOpts.wrapscan
	case "wrapscroll":
		if e.val == "" || e.val == "true" {
			gOpts.wrapscroll = true
		} else if e.val == "false" {
			gOpts.wrapscroll = false
		} else {
			app.ui.echoerr("wrapscroll: value should be empty, 'true', or 'false'")
			return
		}
	case "nowrapscroll":
		if e.val != "" {
			app.ui.echoerrf("nowrapscroll: unexpected value: %s", e.val)
			return
		}
		gOpts.wrapscroll = false
	case "wrapscroll!":
		if e.val != "" {
			app.ui.echoerrf("wrapscroll!: unexpected value: %s", e.val)
			return
		}
		gOpts.wrapscroll = !gOpts.wrapscroll
	case "sixel":
		if e.val == "" || e.val == "true" {
			gOpts.sixel = true
		} else if e.val == "false" {
			gOpts.sixel = false
		} else {
			app.ui.echoerr("sixel: value should be empty, 'true', or 'false'")
			return
		}
	case "nosixel":
		if e.val != "" {
			app.ui.echoerrf("nosixel: unexpected value: %s", e.val)
			return
		}
		gOpts.sixel = false
	case "sixel!":
		if e.val != "" {
			app.ui.echoerrf("sixel!: unexpected value: %s", e.val)
			return
		}
		gOpts.sixel = !gOpts.sixel
	default:
		// any key with the prefix user_ is accepted as a user defined option
		if strings.HasPrefix(e.opt, "user_") {
			gOpts.user[e.opt[5:]] = e.val
			// Export user defined options immediately, so that the current values
			// are available for some external previewer, which is started in a
			// different thread and thus cannot export (as `setenv` is not thread-safe).
			os.Setenv("lf_"+e.opt, e.val)
		} else {
			app.ui.echoerrf("unknown option: %s", e.opt)
		}
		return
	}
	app.ui.loadFileInfo(app.nav)
}

func (e *setLocalExpr) eval(app *app, args []string) {
	path := replaceTilde(e.path)
	if !filepath.IsAbs(path) {
		app.ui.echoerr("setlocal: path should be absolute")
		return
	}

	switch e.opt {
	case "dirfirst":
		if e.val == "" || e.val == "true" {
			gLocalOpts.dirfirsts[path] = true
		} else if e.val == "false" {
			gLocalOpts.dirfirsts[path] = false
		} else {
			app.ui.echoerr("dirfirst: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.ui.sort()
	case "nodirfirst":
		if e.val != "" {
			app.ui.echoerrf("nodirfirst: unexpected value: %s", e.val)
			return
		}
		gLocalOpts.dirfirsts[path] = false
		app.nav.sort()
		app.ui.sort()
	case "dirfirst!":
		if e.val != "" {
			app.ui.echoerrf("dirfirst!: unexpected value: %s", e.val)
			return
		}
		if val, ok := gLocalOpts.dirfirsts[path]; ok {
			gLocalOpts.dirfirsts[path] = !val
		} else {
			val = gOpts.sortType.option&dirfirstSort != 0
			gLocalOpts.dirfirsts[path] = !val
		}
		app.nav.sort()
		app.ui.sort()
	case "dironly":
		if e.val == "" || e.val == "true" {
			gLocalOpts.dironlys[path] = true
		} else if e.val == "false" {
			gLocalOpts.dironlys[path] = false
		} else {
			app.ui.echoerr("dironly: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.ui.sort()
	case "nodironly":
		if e.val != "" {
			app.ui.echoerrf("nodironly: unexpected value: %s", e.val)
			return
		}
		gLocalOpts.dironlys[path] = false
		app.nav.sort()
		app.ui.sort()
	case "dironly!":
		if e.val != "" {
			app.ui.echoerrf("dironly!: unexpected value: %s", e.val)
			return
		}
		if val, ok := gLocalOpts.dironlys[path]; ok {
			gLocalOpts.dironlys[path] = !val
		} else {
			gLocalOpts.dironlys[path] = !gOpts.dironly
		}
		app.nav.sort()
		app.ui.sort()
	case "hidden":
		if e.val == "" || e.val == "true" {
			gLocalOpts.hiddens[path] = true
		} else if e.val == "false" {
			gLocalOpts.hiddens[path] = false
		} else {
			app.ui.echoerr("hidden: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.ui.sort()
	case "nohidden":
		if e.val != "" {
			app.ui.echoerrf("nohidden: unexpected value: %s", e.val)
			return
		}
		gLocalOpts.hiddens[path] = false
		app.nav.sort()
		app.ui.sort()
	case "hidden!":
		if e.val != "" {
			app.ui.echoerrf("hidden!: unexpected value: %s", e.val)
			return
		}
		if val, ok := gLocalOpts.hiddens[path]; ok {
			gLocalOpts.hiddens[path] = !val
		} else {
			val = gOpts.sortType.option&hiddenSort != 0
			gLocalOpts.hiddens[path] = !val
		}
		app.nav.sort()
		app.ui.sort()
	case "info":
		if e.val == "" {
			gLocalOpts.infos[path] = nil
			return
		}
		toks := strings.Split(e.val, ":")
		for _, s := range toks {
			switch s {
			case "size", "time", "atime", "ctime":
			default:
				app.ui.echoerr("info: should consist of 'size', 'time', 'atime' or 'ctime' separated with colon")
				return
			}
		}
		gLocalOpts.infos[path] = toks
	case "reverse":
		if e.val == "" || e.val == "true" {
			gLocalOpts.reverses[path] = true
		} else if e.val == "false" {
			gLocalOpts.reverses[path] = false
		} else {
			app.ui.echoerr("reverse: value should be empty, 'true', or 'false'")
			return
		}
		app.nav.sort()
		app.ui.sort()
	case "noreverse":
		if e.val != "" {
			app.ui.echoerrf("noreverse: unexpected value: %s", e.val)
			return
		}
		gLocalOpts.reverses[path] = false
		app.nav.sort()
		app.ui.sort()
	case "reverse!":
		if e.val != "" {
			app.ui.echoerrf("reverse!: unexpected value: %s", e.val)
			return
		}
		if val, ok := gLocalOpts.reverses[path]; ok {
			gLocalOpts.reverses[path] = !val
		} else {
			val = gOpts.sortType.option&reverseSort != 0
			gLocalOpts.reverses[path] = !val
		}
		app.nav.sort()
		app.ui.sort()
	case "sortby":
		switch e.val {
		case "natural":
			gLocalOpts.sortMethods[path] = naturalSort
		case "name":
			gLocalOpts.sortMethods[path] = nameSort
		case "size":
			gLocalOpts.sortMethods[path] = sizeSort
		case "time":
			gLocalOpts.sortMethods[path] = timeSort
		case "ctime":
			gLocalOpts.sortMethods[path] = ctimeSort
		case "atime":
			gLocalOpts.sortMethods[path] = atimeSort
		case "ext":
			gLocalOpts.sortMethods[path] = extSort
		default:
			app.ui.echoerr("sortby: value should either be 'natural', 'name', 'size', 'time', 'atime', 'ctime' or 'ext'")
			return
		}
		app.nav.sort()
		app.ui.sort()
	default:
		app.ui.echoerrf("unknown option: %s", e.opt)
		return
	}
	app.ui.loadFileInfo(app.nav)
}

func (e *mapExpr) eval(app *app, args []string) {
	if e.expr == nil {
		delete(gOpts.keys, e.keys)
	} else {
		gOpts.keys[e.keys] = e.expr
	}
	app.ui.loadFileInfo(app.nav)
}

func (e *cmapExpr) eval(app *app, args []string) {
	if e.expr == nil {
		delete(gOpts.cmdkeys, e.key)
	} else {
		gOpts.cmdkeys[e.key] = e.expr
	}
	app.ui.loadFileInfo(app.nav)
}

func (e *cmdExpr) eval(app *app, args []string) {
	if e.expr == nil {
		delete(gOpts.cmds, e.name)
	} else {
		gOpts.cmds[e.name] = e.expr
	}
	app.ui.loadFileInfo(app.nav)
}

func preChdir(app *app) {
	if cmd, ok := gOpts.cmds["pre-cd"]; ok {
		cmd.eval(app, nil)
	}
}

func onChdir(app *app) {
	app.nav.addJumpList()
	if cmd, ok := gOpts.cmds["on-cd"]; ok {
		cmd.eval(app, nil)
	}
}

func onRedraw(app *app) {
	if cmd, ok := gOpts.cmds["on-redraw"]; ok {
		cmd.eval(app, nil)
	}
}

func onSelect(app *app) {
	if cmd, ok := gOpts.cmds["on-select"]; ok {
		cmd.eval(app, nil)
	}
}

func splitKeys(s string) (keys []string) {
	for i := 0; i < len(s); {
		r, w := utf8.DecodeRuneInString(s[i:])
		if r != '<' {
			keys = append(keys, s[i:i+w])
			i += w
		} else {
			j := i + w
			for r != '>' && j < len(s) {
				r, w = utf8.DecodeRuneInString(s[j:])
				j += w
			}
			keys = append(keys, s[i:j])
			i = j
		}
	}
	return
}

func doComplete(app *app) (matches []string) {
	switch app.ui.cmdPrefix {
	case ":":
		matches, app.ui.cmdAccLeft = completeCmd(app.ui.cmdAccLeft)
	case "/", "?":
		matches, app.ui.cmdAccLeft = completeFile(app.ui.cmdAccLeft)
	case "$", "%", "!", "&":
		matches, app.ui.cmdAccLeft = completeShell(app.ui.cmdAccLeft)
	}
	return
}

func menuComplete(app *app, dir int) {
	if !app.menuCompActive {
		toks := tokenize(string(app.ui.cmdAccLeft))
		for i, tok := range toks {
			toks[i] = replaceTilde(tok)
		}
		app.ui.cmdAccLeft = []rune(strings.Join(toks, " "))
		app.ui.cmdTmp = app.ui.cmdAccLeft
		app.menuComps = doComplete(app)
		if len(app.menuComps) > 1 {
			app.menuCompInd = -1
			app.menuCompActive = true
		}
	} else {
		app.menuCompInd += dir
		if app.menuCompInd == len(app.menuComps) {
			app.menuCompInd = 0
		} else if app.menuCompInd < 0 {
			app.menuCompInd = len(app.menuComps) - 1
		}

		comp := app.menuComps[app.menuCompInd]
		toks := tokenize(string(app.ui.cmdTmp))
		last := toks[len(toks)-1]

		if app.ui.cmdPrefix != "/" && app.ui.cmdPrefix != "?" {
			comp = escape(comp)
			_, last = filepath.Split(last)
		}

		ind := len(app.ui.cmdTmp) - len([]rune(last))
		app.ui.cmdAccLeft = append(app.ui.cmdTmp[:ind], []rune(comp)...)
	}
	app.ui.menuBuf = listMatches(app.ui.screen, app.menuComps, app.menuCompInd)
}

func update(app *app) {
	app.ui.menuBuf = nil
	app.menuCompActive = false

	switch {
	case gOpts.incsearch && app.ui.cmdPrefix == "/":
		app.nav.search = string(app.ui.cmdAccLeft) + string(app.ui.cmdAccRight)
		if app.nav.search == "" {
			return
		}

		dir := app.nav.currDir()
		old := dir.ind
		dir.ind = app.nav.searchInd
		dir.pos = app.nav.searchPos

		if _, err := app.nav.searchNext(); err != nil {
			app.ui.echoerrf("search: %s: %s", err, app.nav.search)
		} else if old != dir.ind {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case gOpts.incsearch && app.ui.cmdPrefix == "?":
		app.nav.search = string(app.ui.cmdAccLeft) + string(app.ui.cmdAccRight)
		if app.nav.search == "" {
			return
		}

		dir := app.nav.currDir()
		old := dir.ind
		dir.ind = app.nav.searchInd
		dir.pos = app.nav.searchPos

		if _, err := app.nav.searchPrev(); err != nil {
			app.ui.echoerrf("search: %s: %s", err, app.nav.search)
		} else if old != dir.ind {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case gOpts.incfilter && app.ui.cmdPrefix == "filter: ":
		filter := string(app.ui.cmdAccLeft) + string(app.ui.cmdAccRight)
		dir := app.nav.currDir()
		old := dir.ind

		if err := app.nav.setFilter(strings.Split(filter, " ")); err != nil {
			app.ui.echoerrf("filter: %s", err)
		} else if old != dir.ind {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	}
}

func restartIncCmd(app *app) {
	if gOpts.incsearch && (app.ui.cmdPrefix == "/" || app.ui.cmdPrefix == "?") {
		dir := app.nav.currDir()
		app.nav.searchInd = dir.ind
		app.nav.searchPos = dir.pos
		update(app)
	} else if gOpts.incfilter && app.ui.cmdPrefix == "filter: " {
		dir := app.nav.currDir()
		app.nav.prevFilter = dir.filter
		update(app)
	}
}

func resetIncCmd(app *app) {
	if gOpts.incsearch && (app.ui.cmdPrefix == "/" || app.ui.cmdPrefix == "?") {
		dir := app.nav.currDir()
		dir.pos = app.nav.searchPos
		if dir.ind != app.nav.searchInd {
			dir.ind = app.nav.searchInd
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	} else if gOpts.incfilter && app.ui.cmdPrefix == "filter: " {
		dir := app.nav.currDir()
		old := dir.ind
		app.nav.setFilter(app.nav.prevFilter)
		if old != dir.ind {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	}
}

func normal(app *app) {
	resetIncCmd(app)

	app.cmdHistoryInd = 0
	app.menuCompActive = false

	app.ui.menuBuf = nil
	app.ui.cmdAccLeft = nil
	app.ui.cmdAccRight = nil
	app.ui.cmdPrefix = ""
}

func insert(app *app, arg string) {
	switch {
	case gOpts.incsearch && (app.ui.cmdPrefix == "/" || app.ui.cmdPrefix == "?"):
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(arg)...)
		update(app)
	case gOpts.incfilter && app.ui.cmdPrefix == "filter: ":
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(arg)...)
		update(app)
	case app.ui.cmdPrefix == "find: ":
		app.nav.find = string(app.ui.cmdAccLeft) + arg + string(app.ui.cmdAccRight)

		if gOpts.findlen == 0 {
			switch app.nav.findSingle() {
			case 0:
				app.ui.echoerrf("find: pattern not found: %s", app.nav.find)
			case 1:
				app.ui.loadFile(app, true)
				app.ui.loadFileInfo(app.nav)
			default:
				app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(arg)...)
				return
			}
		} else {
			if len(app.nav.find) < gOpts.findlen {
				app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(arg)...)
				return
			}

			if moved, found := app.nav.findNext(); !found {
				app.ui.echoerrf("find: pattern not found: %s", app.nav.find)
			} else if moved {
				app.ui.loadFile(app, true)
				app.ui.loadFileInfo(app.nav)
			}
		}

		normal(app)
	case app.ui.cmdPrefix == "find-back: ":
		app.nav.find = string(app.ui.cmdAccLeft) + arg + string(app.ui.cmdAccRight)

		if gOpts.findlen == 0 {
			switch app.nav.findSingle() {
			case 0:
				app.ui.echoerrf("find-back: pattern not found: %s", app.nav.find)
			case 1:
				app.ui.loadFile(app, true)
				app.ui.loadFileInfo(app.nav)
			default:
				app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(arg)...)
				return
			}
		} else {
			if len(app.nav.find) < gOpts.findlen {
				app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(arg)...)
				return
			}

			if moved, found := app.nav.findPrev(); !found {
				app.ui.echoerrf("find-back: pattern not found: %s", app.nav.find)
			} else if moved {
				app.ui.loadFile(app, true)
				app.ui.loadFileInfo(app.nav)
			}
		}

		normal(app)
	case strings.HasPrefix(app.ui.cmdPrefix, "delete"):
		normal(app)

		if arg == "y" {
			if err := app.nav.del(app); err != nil {
				app.ui.echoerrf("delete: %s", err)
				return
			}
			app.nav.unselect()
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case strings.HasPrefix(app.ui.cmdPrefix, "replace"):
		normal(app)

		if arg == "y" {
			if err := app.nav.rename(); err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}
			if gSingleMode {
				app.nav.renew()
				app.ui.loadFile(app, true)
			} else {
				if err := remote("send load"); err != nil {
					app.ui.echoerrf("rename: %s", err)
					return
				}
			}
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case strings.HasPrefix(app.ui.cmdPrefix, "create"):
		normal(app)

		if arg == "y" {
			if err := os.MkdirAll(filepath.Dir(app.nav.renameNewPath), os.ModePerm); err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}
			if err := app.nav.rename(); err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}
			if gSingleMode {
				app.nav.renew()
				app.ui.loadFile(app, true)
			} else {
				if err := remote("send load"); err != nil {
					app.ui.echoerrf("rename: %s", err)
					return
				}
			}
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case app.ui.cmdPrefix == "mark-save: ":
		normal(app)

		app.nav.marks[arg] = app.nav.currDir().path
		if err := app.nav.writeMarks(); err != nil {
			app.ui.echoerrf("mark-save: %s", err)
		}
		if gSingleMode {
			if err := app.nav.sync(); err != nil {
				app.ui.echoerrf("mark-save: %s", err)
			}
		} else {
			if err := remote("send sync"); err != nil {
				app.ui.echoerrf("mark-save: %s", err)
			}
		}
	case app.ui.cmdPrefix == "mark-load: ":
		normal(app)

		wd, err := os.Getwd()
		if err != nil {
			log.Printf("getting current directory: %s", err)
		}

		path, ok := app.nav.marks[arg]
		if !ok {
			app.ui.echoerr("mark-load: no such mark")
			return
		}

		if wd != path {
			resetIncCmd(app)
			preChdir(app)
		}

		if err := app.nav.cd(path); err != nil {
			app.ui.echoerrf("%s", err)
			return
		}
		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)

		if wd != path {
			app.nav.marks["'"] = wd
			restartIncCmd(app)
			onChdir(app)
		}
	case app.ui.cmdPrefix == "mark-remove: ":
		normal(app)
		if err := app.nav.removeMark(arg); err != nil {
			app.ui.echoerrf("mark-remove: %s", err)
			return
		}
		if err := app.nav.writeMarks(); err != nil {
			app.ui.echoerrf("mark-remove: %s", err)
			return
		}
		if gSingleMode {
			if err := app.nav.sync(); err != nil {
				app.ui.echoerrf("mark-remove: %s", err)
			}
		} else {
			if err := remote("send sync"); err != nil {
				app.ui.echoerrf("mark-remove: %s", err)
			}
		}
	case app.ui.cmdPrefix == ":" && len(app.ui.cmdAccLeft) == 0:
		switch arg {
		case "!", "$", "%", "&":
			app.ui.cmdPrefix = arg
			return
		}
		fallthrough
	default:
		app.ui.menuBuf = nil
		app.menuCompActive = false
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(arg)...)
	}
}

func (e *callExpr) eval(app *app, args []string) {
	os.Setenv("lf_count", strconv.Itoa(e.count))

	switch e.name {
	case "up":
		if !app.nav.init {
			return
		}
		if app.nav.up(e.count) {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "half-up":
		if !app.nav.init {
			return
		}
		if app.nav.up(e.count * app.nav.height / 2) {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "page-up":
		if !app.nav.init {
			return
		}
		if app.nav.up(e.count * app.nav.height) {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "scroll-up":
		if !app.nav.init {
			return
		}
		if app.nav.scrollUp(e.count) {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "down":
		if !app.nav.init {
			return
		}
		if app.nav.down(e.count) {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "half-down":
		if !app.nav.init {
			return
		}
		if app.nav.down(e.count * app.nav.height / 2) {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "page-down":
		if !app.nav.init {
			return
		}
		if app.nav.down(e.count * app.nav.height) {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "scroll-down":
		if !app.nav.init {
			return
		}
		if app.nav.scrollDown(e.count) {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "updir":
		if !app.nav.init {
			return
		}
		resetIncCmd(app)
		preChdir(app)
		for i := 0; i < e.count; i++ {
			if err := app.nav.updir(); err != nil {
				app.ui.echoerrf("%s", err)
				return
			}
		}
		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)
		restartIncCmd(app)
		onChdir(app)
	case "open":
		if !app.nav.init {
			return
		}
		curr, err := app.nav.currFile()
		if err != nil {
			app.ui.echoerrf("opening: %s", err)
			return
		}

		if curr.IsDir() {
			resetIncCmd(app)
			preChdir(app)
			err := app.nav.open()
			if err != nil {
				app.ui.echoerrf("opening directory: %s", err)
				return
			}
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
			restartIncCmd(app)
			onChdir(app)
			return
		}

		if gSelectionPath != "" || gPrintSelection {
			app.selectionOut, _ = app.nav.currFileOrSelections()
			app.quitChan <- struct{}{}
			return
		}

		if cmd, ok := gOpts.cmds["open"]; ok {
			cmd.eval(app, e.args)
		}
	case "jump-prev":
		resetIncCmd(app)
		preChdir(app)
		for i := 0; i < e.count; i++ {
			app.nav.cdJumpListPrev()
		}
		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)
		restartIncCmd(app)
		onChdir(app)
	case "jump-next":
		resetIncCmd(app)
		preChdir(app)
		for i := 0; i < e.count; i++ {
			app.nav.cdJumpListNext()
		}
		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)
		restartIncCmd(app)
		onChdir(app)
	case "quit":
		app.quitChan <- struct{}{}
	case "top":
		if !app.nav.init {
			return
		}
		var moved bool
		if e.count == 1 {
			moved = app.nav.top()
		} else {
			moved = app.nav.move(e.count - 1)
		}
		if moved {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "bottom":
		if !app.nav.init {
			return
		}
		var moved bool
		if e.count == 1 {
			// Different from Vim, which would treat a count of 1 as meaning to
			// move to the first line (i.e. the top)
			moved = app.nav.bottom()
		} else {
			moved = app.nav.move(e.count - 1)
		}
		if moved {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "high":
		if !app.nav.init {
			return
		}
		if app.nav.high() {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "middle":
		if !app.nav.init {
			return
		}
		if app.nav.middle() {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "low":
		if !app.nav.init {
			return
		}
		if app.nav.low() {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "toggle":
		if !app.nav.init {
			return
		}
		if len(e.args) == 0 {
			app.nav.toggle()
		} else {
			dir := app.nav.currDir()
			for _, path := range e.args {
				path = replaceTilde(path)
				if !filepath.IsAbs(path) {
					path = filepath.Join(dir.path, path)
				}
				if _, err := os.Lstat(path); !os.IsNotExist(err) {
					app.nav.toggleSelection(path)
				} else {
					app.ui.echoerrf("toggle: %s", err)
				}
			}
		}
	case "tag-toggle":
		if !app.nav.init {
			return
		}

		tag := "*"
		if len(e.args) != 0 {
			tag = e.args[0]
		}

		if err := app.nav.tagToggle(tag); err != nil {
			app.ui.echoerrf("tag-toggle: %s", err)
		} else if err := app.nav.writeTags(); err != nil {
			app.ui.echoerrf("tag-toggle: %s", err)
		}

		if gSingleMode {
			if err := app.nav.sync(); err != nil {
				app.ui.echoerrf("tag-toggle: %s", err)
			}
		} else {
			if err := remote("send sync"); err != nil {
				app.ui.echoerrf("tag-toggle: %s", err)
			}
		}
	case "tag":
		if !app.nav.init {
			return
		}

		tag := "*"
		if len(e.args) != 0 {
			tag = e.args[0]
		}

		if err := app.nav.tag(tag); err != nil {
			app.ui.echoerrf("tag: %s", err)
		} else if err := app.nav.writeTags(); err != nil {
			app.ui.echoerrf("tag: %s", err)
		}

		if gSingleMode {
			if err := app.nav.sync(); err != nil {
				app.ui.echoerrf("tag: %s", err)
			}
		} else {
			if err := remote("send sync"); err != nil {
				app.ui.echoerrf("tag: %s", err)
			}
		}
	case "invert":
		if !app.nav.init {
			return
		}
		app.nav.invert()
	case "invert-below":
		if !app.nav.init {
			return
		}
		app.nav.invertBelow()
	case "unselect":
		app.nav.unselect()
	case "calcdirsize":
		if !app.nav.init {
			return
		}
		err := app.nav.calcDirSize()
		if err != nil {
			app.ui.echoerrf("calcdirsize: %s", err)
			return
		}
		app.ui.loadFileInfo(app.nav)
		app.nav.sort()
		app.ui.sort()
	case "clearmaps":
		// leave `:` and cmaps bound so the user can still exit using `:quit`
		gOpts.keys = make(map[string]expr)
		gOpts.keys[":"] = &callExpr{"read", nil, 1}
	case "copy":
		if !app.nav.init {
			return
		}

		if err := app.nav.save(true); err != nil {
			app.ui.echoerrf("copy: %s", err)
			return
		}
		app.nav.unselect()
		if gSingleMode {
			if err := app.nav.sync(); err != nil {
				app.ui.echoerrf("copy: %s", err)
				return
			}
		} else {
			if err := remote("send sync"); err != nil {
				app.ui.echoerrf("copy: %s", err)
				return
			}
		}
		app.ui.loadFileInfo(app.nav)
	case "cut":
		if !app.nav.init {
			return
		}

		if err := app.nav.save(false); err != nil {
			app.ui.echoerrf("cut: %s", err)
			return
		}
		app.nav.unselect()
		if gSingleMode {
			if err := app.nav.sync(); err != nil {
				app.ui.echoerrf("cut: %s", err)
				return
			}
		} else {
			if err := remote("send sync"); err != nil {
				app.ui.echoerrf("cut: %s", err)
				return
			}
		}
		app.ui.loadFileInfo(app.nav)
	case "paste":
		if !app.nav.init {
			return
		}

		if cmd, ok := gOpts.cmds["paste"]; ok {
			cmd.eval(app, e.args)
		} else if err := app.nav.paste(app); err != nil {
			app.ui.echoerrf("paste: %s", err)
			return
		}
		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)
	case "delete":
		if !app.nav.init {
			return
		}

		if cmd, ok := gOpts.cmds["delete"]; ok {
			cmd.eval(app, e.args)
			app.nav.unselect()
			if gSingleMode {
				app.nav.renew()
				app.ui.loadFile(app, true)
			} else {
				if err := remote("send load"); err != nil {
					app.ui.echoerrf("delete: %s", err)
					return
				}
			}
		} else {
			list, err := app.nav.currFileOrSelections()
			if err != nil {
				app.ui.echoerrf("delete: %s", err)
				return
			}

			if app.ui.cmdPrefix == ">" {
				return
			}
			normal(app)
			if len(list) == 1 {
				app.ui.cmdPrefix = "delete '" + list[0] + "' ? [y/N] "
			} else {
				app.ui.cmdPrefix = "delete " + strconv.Itoa(len(list)) + " items? [y/N] "
			}
		}
		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)
	case "clear":
		if !app.nav.init {
			return
		}
		if err := saveFiles(nil, false); err != nil {
			app.ui.echoerrf("clear: %s", err)
			return
		}
		if gSingleMode {
			if err := app.nav.sync(); err != nil {
				app.ui.echoerrf("clear: %s", err)
				return
			}
		} else {
			if err := remote("send sync"); err != nil {
				app.ui.echoerrf("clear: %s", err)
				return
			}
		}
		app.ui.loadFileInfo(app.nav)
	case "draw":
	case "redraw":
		if !app.nav.init {
			return
		}
		app.ui.renew()
		app.ui.screen.Sync()
		if app.nav.height != app.ui.wins[0].h {
			app.nav.height = app.ui.wins[0].h
			app.nav.regCache = make(map[string]*reg)
		}
		for _, dir := range app.nav.dirs {
			dir.boundPos(app.nav.height)
		}
		app.ui.loadFile(app, true)
		onRedraw(app)
	case "load":
		if !app.nav.init {
			return
		}
		app.nav.renew()
		app.ui.loadFile(app, false)
	case "reload":
		if !app.nav.init {
			return
		}
		if err := app.nav.reload(); err != nil {
			app.ui.echoerrf("reload: %s", err)
		}
		app.ui.loadFile(app, true)
		// clear file information, will be loaded asynchronously
		app.ui.msg = ""
	case "read":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = ":"
		app.ui.loadFileInfo(app.nav)
	case "shell":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "$"
		app.ui.loadFileInfo(app.nav)
	case "shell-pipe":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "%"
		app.ui.loadFileInfo(app.nav)
	case "shell-wait":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "!"
		app.ui.loadFileInfo(app.nav)
	case "shell-async":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "&"
		app.ui.loadFileInfo(app.nav)
	case "find":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "find: "
		app.nav.findBack = false
		app.ui.loadFileInfo(app.nav)
	case "find-back":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "find-back: "
		app.nav.findBack = true
		app.ui.loadFileInfo(app.nav)
	case "find-next":
		if !app.nav.init {
			return
		}
		dir := app.nav.currDir()
		old := dir.ind
		for i := 0; i < e.count; i++ {
			if app.nav.findBack {
				app.nav.findPrev()
			} else {
				app.nav.findNext()
			}
		}
		if old != dir.ind {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "find-prev":
		if !app.nav.init {
			return
		}
		dir := app.nav.currDir()
		old := dir.ind
		for i := 0; i < e.count; i++ {
			if app.nav.findBack {
				app.nav.findNext()
			} else {
				app.nav.findPrev()
			}
		}
		if old != dir.ind {
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "search":
		if !app.nav.init {
			return
		}
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "/"
		dir := app.nav.currDir()
		app.nav.searchInd = dir.ind
		app.nav.searchPos = dir.pos
		app.nav.searchBack = false
		app.ui.loadFileInfo(app.nav)
	case "search-back":
		if !app.nav.init {
			return
		}
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "?"
		dir := app.nav.currDir()
		app.nav.searchInd = dir.ind
		app.nav.searchPos = dir.pos
		app.nav.searchBack = true
		app.ui.loadFileInfo(app.nav)
	case "search-next":
		if !app.nav.init {
			return
		}
		for i := 0; i < e.count; i++ {
			if app.nav.searchBack {
				if moved, err := app.nav.searchPrev(); err != nil {
					app.ui.echoerrf("search-back: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app, true)
					app.ui.loadFileInfo(app.nav)
				}
			} else {
				if moved, err := app.nav.searchNext(); err != nil {
					app.ui.echoerrf("search: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app, true)
					app.ui.loadFileInfo(app.nav)
				}
			}
		}
	case "search-prev":
		if !app.nav.init {
			return
		}
		for i := 0; i < e.count; i++ {
			if app.nav.searchBack {
				if moved, err := app.nav.searchNext(); err != nil {
					app.ui.echoerrf("search-back: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app, true)
					app.ui.loadFileInfo(app.nav)
				}
			} else {
				if moved, err := app.nav.searchPrev(); err != nil {
					app.ui.echoerrf("search: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app, true)
					app.ui.loadFileInfo(app.nav)
				}
			}
		}
	case "filter":
		if !app.nav.init {
			return
		}
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "filter: "
		dir := app.nav.currDir()
		app.nav.prevFilter = dir.filter
		if len(e.args) == 0 {
			app.ui.cmdAccLeft = []rune(strings.Join(dir.filter, " "))
		} else {
			app.ui.cmdAccLeft = []rune(strings.Join(e.args, " "))
		}
		app.ui.loadFileInfo(app.nav)
	case "setfilter":
		if !app.nav.init {
			return
		}
		log.Printf("filter: %s", e.args)
		if err := app.nav.setFilter(e.args); err != nil {
			app.ui.echoerrf("filter: %s", err)
		}
		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)
	case "mark-save":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "mark-save: "
	case "mark-load":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.menuBuf = listMarks(app.nav.marks)
		app.ui.cmdPrefix = "mark-load: "
	case "mark-remove":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.menuBuf = listMarks(app.nav.marks)
		app.ui.cmdPrefix = "mark-remove: "
	case "rename":
		if !app.nav.init {
			return
		}
		if cmd, ok := gOpts.cmds["rename"]; ok {
			cmd.eval(app, e.args)
			if gSingleMode {
				app.nav.renew()
				app.ui.loadFile(app, true)
			} else {
				if err := remote("send load"); err != nil {
					app.ui.echoerrf("rename: %s", err)
					return
				}
			}
		} else {
			curr, err := app.nav.currFile()
			if err != nil {
				app.ui.echoerrf("rename: %s:", err)
				return
			}
			if app.ui.cmdPrefix == ">" {
				return
			}
			normal(app)
			app.ui.cmdPrefix = "rename: "
			extension := filepath.Ext(curr.Name())
			if len(extension) == 0 || extension == curr.Name() {
				// no extension or .hidden
				app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(curr.Name())...)
			} else {
				app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(curr.Name()[:len(curr.Name())-len(extension)])...)
				app.ui.cmdAccRight = append(app.ui.cmdAccRight, []rune(extension)...)
			}
		}
		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)
	case "sync":
		if err := app.nav.sync(); err != nil {
			app.ui.echoerrf("sync: %s", err)
		}
	case "echo":
		app.ui.echo(strings.Join(e.args, " "))
	case "echomsg":
		app.ui.echomsg(strings.Join(e.args, " "))
	case "echoerr":
		app.ui.echoerr(strings.Join(e.args, " "))
	case "cd":
		path := "~"
		if len(e.args) > 0 {
			path = e.args[0]
		}

		wd, err := os.Getwd()
		if err != nil {
			log.Printf("getting current directory: %s", err)
		}

		path = replaceTilde(path)
		if !filepath.IsAbs(path) {
			path = filepath.Join(wd, path)
		} else {
			path = filepath.Clean(path)
		}

		if wd != path {
			resetIncCmd(app)
			preChdir(app)
		}

		if err := app.nav.cd(path); err != nil {
			app.ui.echoerrf("%s", err)
			return
		}

		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)

		if wd != path {
			app.nav.marks["'"] = wd
			restartIncCmd(app)
			onChdir(app)
		}
	case "select":
		if !app.nav.init {
			return
		}

		if len(e.args) != 1 {
			app.ui.echoerr("select: requires an argument")
			return
		}

		wd, err := os.Getwd()
		if err != nil {
			log.Printf("getting current directory: %s", err)
		}

		path := filepath.Dir(replaceTilde(e.args[0]))
		if !filepath.IsAbs(path) {
			path = filepath.Join(wd, path)
		} else {
			path = filepath.Clean(path)
		}

		if wd != path {
			resetIncCmd(app)
			preChdir(app)
		}

		if err := app.nav.sel(e.args[0]); err != nil {
			app.ui.echoerrf("%s", err)
			return
		}

		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)

		if wd != path {
			app.nav.marks["'"] = wd
			restartIncCmd(app)
			onChdir(app)
		}
	case "glob-select":
		if !app.nav.init {
			return
		}
		if len(e.args) != 1 {
			app.ui.echoerr("glob-select: requires a pattern to match")
			return
		}
		if err := app.nav.globSel(e.args[0], false); err != nil {
			app.ui.echoerrf("%s", err)
			return
		}
	case "glob-unselect":
		if !app.nav.init {
			return
		}
		if len(e.args) != 1 {
			app.ui.echoerr("glob-unselect: requires a pattern to match")
			return
		}
		if err := app.nav.globSel(e.args[0], true); err != nil {
			app.ui.echoerrf("%s", err)
			return
		}
	case "source":
		if len(e.args) != 1 {
			app.ui.echoerr("source: requires an argument")
			return
		}
		app.readFile(replaceTilde(e.args[0]))
		app.ui.loadFileInfo(app.nav)
	case "push":
		if len(e.args) != 1 {
			app.ui.echoerr("push: requires an argument")
			return
		}
		log.Println("pushing keys", e.args[0])
		for _, val := range splitKeys(e.args[0]) {
			app.ui.keyChan <- val
		}
	case "cmd-insert":
		if len(e.args) == 0 {
			return
		}
		insert(app, e.args[0])
	case "cmd-escape":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
	case "cmd-complete":
		matches := doComplete(app)
		app.ui.menuBuf = listMatches(app.ui.screen, matches, -1)
	case "cmd-menu-complete":
		menuComplete(app, 1)
	case "cmd-menu-complete-back":
		menuComplete(app, -1)
	case "cmd-menu-accept":
		app.ui.menuBuf = nil
		app.menuCompActive = false
	case "cmd-enter":
		s := string(append(app.ui.cmdAccLeft, app.ui.cmdAccRight...))
		if len(s) == 0 && app.ui.cmdPrefix != "filter: " && app.ui.cmdPrefix != ">" {
			return
		}

		app.ui.menuBuf = nil
		app.menuCompActive = false

		app.ui.cmdAccLeft = nil
		app.ui.cmdAccRight = nil

		switch app.ui.cmdPrefix {
		case ":":
			log.Printf("command: %s", s)
			app.ui.cmdPrefix = ""
			app.cmdHistory = append(app.cmdHistory, cmdItem{":", s})
			p := newParser(strings.NewReader(s))
			for p.parse() {
				p.expr.eval(app, nil)
			}
			if p.err != nil {
				app.ui.echoerrf("%s", p.err)
			}
		case "$":
			log.Printf("shell: %s", s)
			app.ui.cmdPrefix = ""
			app.cmdHistory = append(app.cmdHistory, cmdItem{"$", s})
			app.runShell(s, nil, "$")
		case "%":
			log.Printf("shell-pipe: %s", s)
			app.cmdHistory = append(app.cmdHistory, cmdItem{"%", s})
			app.runShell(s, nil, "%")
		case ">":
			io.WriteString(app.cmdIn, s+"\n")
			app.cmdOutBuf = nil
		case "!":
			log.Printf("shell-wait: %s", s)
			app.ui.cmdPrefix = ""
			app.cmdHistory = append(app.cmdHistory, cmdItem{"!", s})
			app.runShell(s, nil, "!")
		case "&":
			log.Printf("shell-async: %s", s)
			app.ui.cmdPrefix = ""
			app.cmdHistory = append(app.cmdHistory, cmdItem{"&", s})
			app.runShell(s, nil, "&")
		case "/":
			dir := app.nav.currDir()
			old := dir.ind
			if gOpts.incsearch {
				dir.ind = app.nav.searchInd
				dir.pos = app.nav.searchPos
			}
			log.Printf("search: %s", s)
			app.ui.cmdPrefix = ""
			app.nav.search = s
			if _, err := app.nav.searchNext(); err != nil {
				app.ui.echoerrf("search: %s: %s", err, app.nav.search)
			} else if old != dir.ind {
				app.ui.loadFile(app, true)
				app.ui.loadFileInfo(app.nav)
			}
		case "?":
			dir := app.nav.currDir()
			old := dir.ind
			if gOpts.incsearch {
				dir.ind = app.nav.searchInd
				dir.pos = app.nav.searchPos
			}
			log.Printf("search-back: %s", s)
			app.ui.cmdPrefix = ""
			app.nav.search = s
			if _, err := app.nav.searchPrev(); err != nil {
				app.ui.echoerrf("search-back: %s: %s", err, app.nav.search)
			} else if old != dir.ind {
				app.ui.loadFile(app, true)
				app.ui.loadFileInfo(app.nav)
			}
		case "filter: ":
			log.Printf("filter: %s", s)
			app.ui.cmdPrefix = ""
			if err := app.nav.setFilter(strings.Split(s, " ")); err != nil {
				app.ui.echoerrf("filter: %s", err)
			}
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		case "find: ":
			app.ui.cmdPrefix = ""
			if moved, found := app.nav.findNext(); !found {
				app.ui.echoerrf("find: pattern not found: %s", app.nav.find)
			} else if moved {
				app.ui.loadFile(app, true)
				app.ui.loadFileInfo(app.nav)
			}
		case "find-back: ":
			app.ui.cmdPrefix = ""
			if moved, found := app.nav.findPrev(); !found {
				app.ui.echoerrf("find-back: pattern not found: %s", app.nav.find)
			} else if moved {
				app.ui.loadFile(app, true)
				app.ui.loadFileInfo(app.nav)
			}
		case "rename: ":
			app.ui.cmdPrefix = ""

			curr, err := app.nav.currFile()
			if err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}
			wd, err := os.Getwd()
			if err != nil {
				log.Printf("getting current directory: %s", err)
				return
			}

			oldPath := filepath.Join(wd, curr.Name())
			newPath := filepath.Clean(replaceTilde(s))
			if !filepath.IsAbs(newPath) {
				newPath = filepath.Join(wd, newPath)
			}
			if oldPath == newPath {
				return
			}
			app.nav.renameOldPath = oldPath
			app.nav.renameNewPath = newPath

			newDir := filepath.Dir(newPath)
			if _, err := os.Stat(newDir); os.IsNotExist(err) {
				app.ui.cmdPrefix = "create '" + newDir + "' ? [y/N] "
				return
			}

			oldStat, err := os.Lstat(oldPath)
			if err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}
			if newStat, err := os.Lstat(newPath); !os.IsNotExist(err) && !os.SameFile(oldStat, newStat) {
				app.ui.cmdPrefix = "replace '" + newPath + "' ? [y/N] "
				return
			}

			if err := app.nav.rename(); err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}

			if gSingleMode {
				app.nav.renew()
			} else {
				if err := remote("send load"); err != nil {
					app.ui.echoerrf("rename: %s", err)
					return
				}
			}
			app.ui.loadFile(app, true)
			app.ui.loadFileInfo(app.nav)
		default:
			log.Printf("entering unknown execution prefix: %q", app.ui.cmdPrefix)
		}
	case "cmd-history-next":
		if app.ui.cmdPrefix == "" || app.ui.cmdPrefix == ">" {
			return
		}
		if app.cmdHistoryInd > 0 {
			app.cmdHistoryInd--
		}
		if app.cmdHistoryInd == 0 {
			normal(app)
			app.ui.cmdPrefix = ":"
			return
		}
		historyInd := app.cmdHistoryInd
		cmd := app.cmdHistory[len(app.cmdHistory)-historyInd]
		normal(app)
		app.cmdHistoryInd = historyInd
		app.ui.cmdPrefix = cmd.prefix
		app.ui.cmdAccLeft = []rune(cmd.value)
	case "cmd-history-prev":
		if app.ui.cmdPrefix == ">" {
			return
		}
		if app.cmdHistoryInd == len(app.cmdHistory) {
			return
		}
		historyInd := app.cmdHistoryInd + 1
		cmd := app.cmdHistory[len(app.cmdHistory)-historyInd]
		normal(app)
		app.cmdHistoryInd = historyInd
		app.ui.cmdPrefix = cmd.prefix
		app.ui.cmdAccLeft = []rune(cmd.value)
	case "cmd-delete":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		app.ui.cmdAccRight = app.ui.cmdAccRight[1:]
		update(app)
	case "cmd-delete-back":
		if len(app.ui.cmdAccLeft) == 0 {
			switch app.ui.cmdPrefix {
			case "!", "$", "%", "&":
				app.ui.cmdPrefix = ":"
			case ">", "rename: ", "filter: ":
				// Don't mess with programs waiting for input.
				// Exiting on backspace is also inconvenient for 'rename' and 'filter',
				// since the text field can start out nonempty.
			default:
				normal(app)
			}
			return
		}
		app.ui.cmdAccLeft = app.ui.cmdAccLeft[:len(app.ui.cmdAccLeft)-1]
		update(app)
	case "cmd-left":
		if len(app.ui.cmdAccLeft) == 0 {
			return
		}
		app.ui.cmdAccRight = append([]rune{app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-1]}, app.ui.cmdAccRight...)
		app.ui.cmdAccLeft = app.ui.cmdAccLeft[:len(app.ui.cmdAccLeft)-1]
	case "cmd-right":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, app.ui.cmdAccRight[0])
		app.ui.cmdAccRight = app.ui.cmdAccRight[1:]
	case "cmd-home":
		app.ui.cmdAccRight = append(app.ui.cmdAccLeft, app.ui.cmdAccRight...)
		app.ui.cmdAccLeft = nil
	case "cmd-end":
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, app.ui.cmdAccRight...)
		app.ui.cmdAccRight = nil
	case "cmd-delete-home":
		if len(app.ui.cmdAccLeft) == 0 {
			return
		}
		app.ui.cmdYankBuf = app.ui.cmdAccLeft
		app.ui.cmdAccLeft = nil
		update(app)
	case "cmd-delete-end":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		app.ui.cmdYankBuf = app.ui.cmdAccRight
		app.ui.cmdAccRight = nil
		update(app)
	case "cmd-delete-unix-word":
		ind := strings.LastIndex(strings.TrimRight(string(app.ui.cmdAccLeft), " "), " ") + 1
		app.ui.cmdYankBuf = []rune(string(app.ui.cmdAccLeft)[ind:])
		app.ui.cmdAccLeft = []rune(string(app.ui.cmdAccLeft)[:ind])
		update(app)
	case "cmd-yank":
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, app.ui.cmdYankBuf...)
		update(app)
	case "cmd-transpose":
		if len(app.ui.cmdAccLeft) < 2 {
			return
		}
		app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-1], app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-2] = app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-2], app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-1]
		update(app)
	case "cmd-interrupt":
		if app.cmd != nil {
			err := shellKill(app.cmd)
			if err != nil {
				app.ui.echoerrf("kill: %s", err)
			} else {
				app.ui.echoerr("process interrupt")
			}
		}
		normal(app)
	case "cmd-word":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		loc := reWordEnd.FindStringSubmatchIndex(string(app.ui.cmdAccRight))
		if loc == nil {
			return
		}
		ind := loc[3]
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(string(app.ui.cmdAccRight)[:ind])...)
		app.ui.cmdAccRight = []rune(string(app.ui.cmdAccRight)[ind:])
	case "cmd-word-back":
		if len(app.ui.cmdAccLeft) == 0 {
			return
		}
		locs := reWordBeg.FindAllStringSubmatchIndex(string(app.ui.cmdAccLeft), -1)
		if locs == nil {
			return
		}
		ind := locs[len(locs)-1][3]
		old := app.ui.cmdAccRight
		app.ui.cmdAccRight = append([]rune{}, []rune(string(app.ui.cmdAccLeft)[ind:])...)
		app.ui.cmdAccRight = append(app.ui.cmdAccRight, old...)
		app.ui.cmdAccLeft = []rune(string(app.ui.cmdAccLeft)[:ind])
	case "cmd-capitalize-word":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		ind := 0
		for ; ind < len(app.ui.cmdAccRight) && unicode.IsSpace(app.ui.cmdAccRight[ind]); ind++ {
		}
		if ind >= len(app.ui.cmdAccRight) {
			return
		}
		app.ui.cmdAccRight[ind] = unicode.ToUpper(app.ui.cmdAccRight[ind])
		loc := reWordEnd.FindStringSubmatchIndex(string(app.ui.cmdAccRight))
		if loc == nil {
			return
		}
		ind = loc[3]
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(string(app.ui.cmdAccRight)[:ind])...)
		app.ui.cmdAccRight = []rune(string(app.ui.cmdAccRight)[ind:])
		update(app)
	case "cmd-delete-word":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		loc := reWordEnd.FindStringSubmatchIndex(string(app.ui.cmdAccRight))
		if loc == nil {
			return
		}
		ind := loc[3]
		app.ui.cmdYankBuf = []rune(string(app.ui.cmdAccRight)[:ind])
		app.ui.cmdAccRight = []rune(string(app.ui.cmdAccRight)[ind:])
		update(app)
	case "cmd-delete-word-back":
		if len(app.ui.cmdAccLeft) == 0 {
			return
		}
		locs := reWordBeg.FindAllStringSubmatchIndex(string(app.ui.cmdAccLeft), -1)
		if locs == nil {
			return
		}
		ind := locs[len(locs)-1][3]
		app.ui.cmdYankBuf = []rune(string(app.ui.cmdAccLeft)[ind:])
		app.ui.cmdAccLeft = app.ui.cmdAccLeft[:ind]
		update(app)
	case "cmd-uppercase-word":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		loc := reWordEnd.FindStringSubmatchIndex(string(app.ui.cmdAccRight))
		if loc == nil {
			return
		}
		ind := loc[3]
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(strings.ToUpper(string(app.ui.cmdAccRight)[:ind]))...)
		app.ui.cmdAccRight = []rune(string(app.ui.cmdAccRight)[ind:])
		update(app)
	case "cmd-lowercase-word":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		loc := reWordEnd.FindStringSubmatchIndex(string(app.ui.cmdAccRight))
		if loc == nil {
			return
		}
		ind := loc[3]
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(strings.ToLower(string(app.ui.cmdAccRight)[:ind]))...)
		app.ui.cmdAccRight = []rune(string(app.ui.cmdAccRight)[ind:])
		update(app)
	case "cmd-transpose-word":
		if len(app.ui.cmdAccLeft) == 0 {
			return
		}

		locs := reWord.FindAllStringIndex(string(app.ui.cmdAccLeft), -1)
		if len(locs) < 2 {
			return
		}

		if len(app.ui.cmdAccRight) > 0 {
			loc := reWordEnd.FindStringSubmatchIndex(string(app.ui.cmdAccRight))
			if loc != nil {
				ind := loc[3]
				app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(string(app.ui.cmdAccRight)[:ind])...)
				app.ui.cmdAccRight = []rune(string(app.ui.cmdAccRight)[ind:])
			}
		}

		locs = reWord.FindAllStringIndex(string(app.ui.cmdAccLeft), -1)

		beg1, end1 := locs[len(locs)-2][0], locs[len(locs)-2][1]
		beg2, end2 := locs[len(locs)-1][0], locs[len(locs)-1][1]

		var acc []rune

		acc = append(acc, []rune(string(app.ui.cmdAccLeft)[:beg1])...)
		acc = append(acc, []rune(string(app.ui.cmdAccLeft)[beg2:end2])...)
		acc = append(acc, []rune(string(app.ui.cmdAccLeft)[end1:beg2])...)
		acc = append(acc, []rune(string(app.ui.cmdAccLeft)[beg1:end1])...)
		acc = append(acc, []rune(string(app.ui.cmdAccLeft)[end2:])...)

		app.ui.cmdAccLeft = acc
		update(app)
	default:
		cmd, ok := gOpts.cmds[e.name]
		if !ok {
			app.ui.echoerrf("command not found: %s", e.name)
			return
		}
		cmd.eval(app, e.args)
	}
}

func (e *execExpr) eval(app *app, args []string) {
	switch e.prefix {
	case "$":
		log.Printf("shell: %s -- %s", e, args)
		app.runShell(e.value, args, e.prefix)
	case "%":
		log.Printf("shell-pipe: %s -- %s", e, args)
		app.runShell(e.value, args, e.prefix)
	case "!":
		log.Printf("shell-wait: %s -- %s", e, args)
		app.runShell(e.value, args, e.prefix)
	case "&":
		log.Printf("shell-async: %s -- %s", e, args)
		app.runShell(e.value, args, e.prefix)
	default:
		log.Printf("evaluating unknown execution prefix: %q", e.prefix)
	}
}

func (e *listExpr) eval(app *app, args []string) {
	for i := 0; i < e.count; i++ {
		for _, expr := range e.exprs {
			expr.eval(app, nil)
		}
	}
}
