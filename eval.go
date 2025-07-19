package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

func applyBoolOpt(opt *bool, e *setExpr) error {
	switch {
	case strings.HasPrefix(e.opt, "no"):
		if e.val != "" {
			return fmt.Errorf("%s: unexpected value: %s", e.opt, e.val)
		}
		*opt = false
	case strings.HasSuffix(e.opt, "!"):
		if e.val != "" {
			return fmt.Errorf("%s: unexpected value: %s", e.opt, e.val)
		}
		*opt = !*opt
	default:
		switch e.val {
		case "", "true":
			*opt = true
		case "false":
			*opt = false
		default:
			return fmt.Errorf("%s: value should be empty, 'true', or 'false'", e.opt)
		}
	}

	return nil
}

func applyLocalBoolOpt(localOpt map[string]bool, globalOpt bool, e *setLocalExpr) error {
	opt, ok := localOpt[e.path]
	if !ok {
		opt = globalOpt
	}

	if err := applyBoolOpt(&opt, &setExpr{e.opt, e.val}); err != nil {
		return err
	}

	localOpt[e.path] = opt
	return nil
}

func (e *setExpr) eval(app *app, args []string) {
	var err error
	switch e.opt {
	case "anchorfind", "noanchorfind", "anchorfind!":
		err = applyBoolOpt(&gOpts.anchorfind, e)
	case "autoquit", "noautoquit", "autoquit!":
		err = applyBoolOpt(&gOpts.autoquit, e)
	case "dircache", "nodircache", "dircache!":
		err = applyBoolOpt(&gOpts.dircache, e)
	case "dircounts", "nodircounts", "dircounts!":
		err = applyBoolOpt(&gOpts.dircounts, e)
	case "dirfirst", "nodirfirst", "dirfirst!":
		err = applyBoolOpt(&gOpts.dirfirst, e)
		if err == nil {
			app.nav.sort()
			app.ui.sort()
		}
	case "dironly", "nodironly", "dironly!":
		err = applyBoolOpt(&gOpts.dironly, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.sort()
			app.ui.loadFile(app, true)
		}
	case "dirpreviews", "nodirpreviews", "dirpreviews!":
		err = applyBoolOpt(&gOpts.dirpreviews, e)
	case "drawbox", "nodrawbox", "drawbox!":
		err = applyBoolOpt(&gOpts.drawbox, e)
		if err == nil {
			app.ui.renew()
			if app.nav.height != app.ui.wins[0].h {
				app.nav.height = app.ui.wins[0].h
				clear(app.nav.regCache)
			}
			app.ui.loadFile(app, true)
		}
	case "globfilter", "noglobfilter", "globfilter!":
		err = applyBoolOpt(&gOpts.globfilter, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.sort()
			app.ui.loadFile(app, true)
		}
	case "globsearch", "noglobsearch", "globsearch!":
		err = applyBoolOpt(&gOpts.globsearch, e)
	case "hidden", "nohidden", "hidden!":
		err = applyBoolOpt(&gOpts.hidden, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.sort()
			app.ui.loadFile(app, true)
		}
	case "history", "nohistory", "history!":
		err = applyBoolOpt(&gOpts.history, e)
	case "icons", "noicons", "icons!":
		err = applyBoolOpt(&gOpts.icons, e)
	case "ignorecase", "noignorecase", "ignorecase!":
		err = applyBoolOpt(&gOpts.ignorecase, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.sort()
			app.ui.loadFile(app, true)
		}
	case "ignoredia", "noignoredia", "ignoredia!":
		err = applyBoolOpt(&gOpts.ignoredia, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.sort()
			app.ui.loadFile(app, true)
		}
	case "incfilter", "noincfilter", "incfilter!":
		err = applyBoolOpt(&gOpts.incfilter, e)
	case "incsearch", "noincsearch", "incsearch!":
		err = applyBoolOpt(&gOpts.incsearch, e)
	case "mouse", "nomouse", "mouse!":
		err = applyBoolOpt(&gOpts.mouse, e)
		if err == nil {
			if gOpts.mouse {
				app.ui.screen.EnableMouse(tcell.MouseButtonEvents)
			} else {
				app.ui.screen.DisableMouse()
			}
		}
	case "number", "nonumber", "number!":
		err = applyBoolOpt(&gOpts.number, e)
	case "preview", "nopreview", "preview!":
		preview := gOpts.preview
		err = applyBoolOpt(&preview, e)
		if preview && len(gOpts.ratios) < 2 {
			err = errors.New("preview: 'ratios' should consist of at least two numbers before enabling 'preview'")
		}
		if err == nil {
			if gOpts.sixel {
				app.ui.sxScreen.forceClear = true
			}
			gOpts.preview = preview
			app.ui.loadFile(app, true)
		}
	case "relativenumber", "norelativenumber", "relativenumber!":
		err = applyBoolOpt(&gOpts.relativenumber, e)
	case "reverse", "noreverse", "reverse!":
		err = applyBoolOpt(&gOpts.reverse, e)
		if err == nil {
			app.nav.sort()
			app.ui.sort()
		}
	case "roundbox", "noroundbox", "roundbox!":
		err = applyBoolOpt(&gOpts.roundbox, e)
	case "showbinds", "noshowbinds", "showbinds!":
		err = applyBoolOpt(&gOpts.showbinds, e)
	case "sixel", "nosixel", "sixel!":
		err = applyBoolOpt(&gOpts.sixel, e)
		clear(app.nav.regCache)
		app.ui.sxScreen.forceClear = true
		app.ui.loadFile(app, true)
	case "smartcase", "nosmartcase", "smartcase!":
		err = applyBoolOpt(&gOpts.smartcase, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.sort()
			app.ui.loadFile(app, true)
		}
	case "smartdia", "nosmartdia", "smartdia!":
		err = applyBoolOpt(&gOpts.smartdia, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.sort()
			app.ui.loadFile(app, true)
		}
	case "watch", "nowatch", "watch!":
		err = applyBoolOpt(&gOpts.watch, e)
		if err == nil {
			if gOpts.watch {
				app.watch.start()
				for _, dir := range app.nav.dirCache {
					app.watchDir(dir)
				}
			} else {
				app.watch.stop()
			}
		}
	case "wrapscan", "nowrapscan", "wrapscan!":
		err = applyBoolOpt(&gOpts.wrapscan, e)
	case "wrapscroll", "nowrapscroll", "wrapscroll!":
		err = applyBoolOpt(&gOpts.wrapscroll, e)
	case "borderfmt":
		gOpts.borderfmt = e.val
	case "cleaner":
		gOpts.cleaner = replaceTilde(e.val)
	case "copyfmt":
		gOpts.copyfmt = e.val
	case "cursoractivefmt":
		gOpts.cursoractivefmt = e.val
	case "cursorparentfmt":
		gOpts.cursorparentfmt = e.val
	case "cursorpreviewfmt":
		gOpts.cursorpreviewfmt = e.val
	case "cutfmt":
		gOpts.cutfmt = e.val
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
	case "ifs":
		gOpts.ifs = e.val
	case "info":
		if e.val == "" {
			gOpts.info = nil
			return
		}
		toks := strings.Split(e.val, ":")
		for _, s := range toks {
			switch s {
			case "size", "time", "atime", "btime", "ctime", "perm", "user", "group", "custom":
			default:
				app.ui.echoerr("info: should consist of 'size', 'time', 'atime', 'btime', 'ctime', 'perm', 'user', 'group' or 'custom' separated with colon")
				return
			}
		}
		gOpts.info = toks
	case "locale":
		localeStr := e.val
		if localeStr != localeStrDisable {
			if _, err = getLocaleTag(localeStr); err != nil {
				app.ui.echoerrf("locale: %s", err.Error())
				return
			}
		}
		gOpts.locale = localeStr
		app.nav.sort()
		app.ui.sort()
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
	case "previewer":
		gOpts.previewer = replaceTilde(e.val)
	case "promptfmt":
		gOpts.promptfmt = e.val
	case "ratios":
		toks := strings.Split(e.val, ":")
		rats := make([]int, 0, len(toks))
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
		if gOpts.sixel {
			clear(app.nav.regCache)
		}
		app.ui.loadFile(app, true)
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
	case "selectfmt":
		gOpts.selectfmt = e.val
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
	case "sortby":
		method := sortMethod(e.val)
		if !isValidSortMethod(method) {
			app.ui.echoerr(invalidSortErrorMessage)
			return
		}
		gOpts.sortby = method
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
		gOpts.tempmarks = "'" + e.val
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
	case "visualfmt":
		gOpts.visualfmt = e.val
	case "waitmsg":
		gOpts.waitmsg = e.val
	default:
		// any key with the prefix user_ is accepted as a user defined option
		if strings.HasPrefix(e.opt, "user_") {
			gOpts.user[e.opt[5:]] = e.val
			// Export user defined options immediately, so that the current values
			// are available for some external previewer, which is started in a
			// different thread and thus cannot export (as `setenv` is not thread-safe).
			os.Setenv("lf_"+e.opt, e.val)
		} else {
			err = fmt.Errorf("unknown option: %s", e.opt)
		}
	}

	if err != nil {
		app.ui.echoerr(err.Error())
		return
	}

	app.ui.loadFileInfo(app.nav)
}

func (e *setLocalExpr) eval(app *app, args []string) {
	e.path = replaceTilde(e.path)
	if !filepath.IsAbs(e.path) {
		app.ui.echoerr("setlocal: path should be absolute")
		return
	}

	var err error
	switch e.opt {
	case "dircounts", "nodircounts", "dircounts!":
		err = applyLocalBoolOpt(gLocalOpts.dircounts, gOpts.dircounts, e)
	case "dirfirst", "nodirfirst", "dirfirst!":
		err = applyLocalBoolOpt(gLocalOpts.dirfirst, gOpts.dirfirst, e)
		if err == nil {
			app.nav.sort()
			app.ui.sort()
		}
	case "dironly", "nodironly", "dironly!":
		err = applyLocalBoolOpt(gLocalOpts.dironly, gOpts.dironly, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.sort()
			app.ui.loadFile(app, true)
		}
	case "hidden", "nohidden", "hidden!":
		err = applyLocalBoolOpt(gLocalOpts.hidden, gOpts.hidden, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.sort()
			app.ui.loadFile(app, true)
		}
	case "reverse", "noreverse", "reverse!":
		err = applyLocalBoolOpt(gLocalOpts.reverse, gOpts.reverse, e)
		if err == nil {
			app.nav.sort()
			app.ui.sort()
		}
	case "info":
		if e.val == "" {
			gLocalOpts.info[e.path] = nil
			return
		}
		toks := strings.Split(e.val, ":")
		for _, s := range toks {
			switch s {
			case "size", "time", "atime", "btime", "ctime", "perm", "user", "group", "custom":
			default:
				app.ui.echoerr("info: should consist of 'size', 'time', 'atime', 'btime', 'ctime', 'perm', 'user', 'group' or 'custom' separated with colon")
				return
			}
		}
		gLocalOpts.info[e.path] = toks
	case "sortby":
		method := sortMethod(e.val)
		if !isValidSortMethod(method) {
			app.ui.echoerr(invalidSortErrorMessage)
			return
		}
		gLocalOpts.sortby[e.path] = method
		app.nav.sort()
		app.ui.sort()
	case "locale":
		localeStr := e.val
		if localeStr != localeStrDisable {
			if _, err = getLocaleTag(localeStr); err != nil {
				app.ui.echoerrf("locale: %s", err.Error())
				return
			}
		}
		gLocalOpts.locale[e.path] = localeStr
		app.nav.sort()
		app.ui.sort()
	default:
		err = fmt.Errorf("unknown option: %s", e.opt)
	}

	if err != nil {
		app.ui.echoerr(err.Error())
		return
	}

	app.ui.loadFileInfo(app.nav)
}

func (e *mapExpr) eval(app *app, args []string) {
	if e.expr == nil {
		delete(gOpts.nkeys, e.keys)
		delete(gOpts.vkeys, e.keys)
	} else {
		gOpts.nkeys[e.keys] = e.expr
		gOpts.vkeys[e.keys] = e.expr
	}
	app.ui.loadFileInfo(app.nav)
}

func (e *nmapExpr) eval(app *app, args []string) {
	if e.expr == nil {
		delete(gOpts.nkeys, e.keys)
	} else {
		gOpts.nkeys[e.keys] = e.expr
	}
	app.ui.loadFileInfo(app.nav)
}

func (e *vmapExpr) eval(app *app, args []string) {
	if e.expr == nil {
		delete(gOpts.vkeys, e.keys)
	} else {
		gOpts.vkeys[e.keys] = e.expr
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

	// only enable focus reporting if required by the user
	if e.name == "on-focus-gained" || e.name == "on-focus-lost" {
		_, onFocusGainedExists := gOpts.cmds["on-focus-gained"]
		_, onFocusLostExists := gOpts.cmds["on-focus-lost"]
		if onFocusGainedExists || onFocusLostExists {
			app.ui.screen.EnableFocus()
		} else {
			app.ui.screen.DisableFocus()
		}
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

func onLoad(app *app, files []string) {
	// prevent infinite loops
	if !gOpts.dircache {
		return
	}

	if cmd, ok := gOpts.cmds["on-load"]; ok {
		cmd.eval(app, files)
	}
}

func onFocusGained(app *app) {
	if cmd, ok := gOpts.cmds["on-focus-gained"]; ok {
		cmd.eval(app, nil)
	}
}

func onFocusLost(app *app) {
	if cmd, ok := gOpts.cmds["on-focus-lost"]; ok {
		cmd.eval(app, nil)
	}
}

func onInit(app *app) {
	if cmd, ok := gOpts.cmds["on-init"]; ok {
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

func onQuit(app *app) {
	if cmd, ok := gOpts.cmds["on-quit"]; ok {
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
	app.ui.menu = listMatches(app.ui.screen, app.menuComps, app.menuCompInd)
}

func update(app *app) {
	app.ui.menu = ""
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

	app.ui.menu = ""
	app.ui.cmdAccLeft = nil
	app.ui.cmdAccRight = nil
	app.ui.cmdPrefix = ""

	// ensure the mode indicator in `statfmt` is updated properly
	app.ui.loadFileInfo(app.nav)
}

func visual(app *app) {
	dir := app.nav.currDir()
	dir.visualAnchor = dir.ind

	app.ui.loadFileInfo(app.nav)
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
			return
		}
		if gSingleMode {
			if err := app.nav.sync(); err != nil {
				app.ui.echoerrf("mark-save: %s", err)
				return
			}
		} else {
			if err := remote("send sync"); err != nil {
				app.ui.echoerrf("mark-save: %s", err)
				return
			}
		}
		app.ui.loadFileInfo(app.nav)
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
				return
			}
		} else {
			if err := remote("send sync"); err != nil {
				app.ui.echoerrf("mark-remove: %s", err)
				return
			}
		}
		app.ui.loadFileInfo(app.nav)
	case app.ui.cmdPrefix == ":" && len(app.ui.cmdAccLeft) == 0:
		switch arg {
		case "!", "$", "%", "&":
			app.ui.cmdPrefix = arg
			return
		}
		fallthrough
	default:
		app.ui.menu = ""
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
	case "tty-write":
		if len(e.args) != 1 {
			app.ui.echoerr("tty-write: requires an argument")
			return
		}

		tty, ok := app.ui.screen.Tty()
		if !ok {
			log.Printf("tty-write: failed to get tty")
			return
		}

		tty.Write([]byte(e.args[0]))
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
		for range e.count {
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

		app.ui.loadFileInfo(app.nav)

		if cmd, ok := gOpts.cmds["open"]; ok {
			cmd.eval(app, e.args)
		}
	case "jump-prev":
		resetIncCmd(app)
		preChdir(app)
		for range e.count {
			app.nav.cdJumpListPrev()
		}
		app.ui.loadFile(app, true)
		app.ui.loadFileInfo(app.nav)
		restartIncCmd(app)
		onChdir(app)
	case "jump-next":
		resetIncCmd(app)
		preChdir(app)
		for range e.count {
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
		clear(gOpts.nkeys)
		clear(gOpts.vkeys)
		gOpts.nkeys[":"] = &callExpr{"read", nil, 1}
		gOpts.vkeys[":"] = &callExpr{"read", nil, 1}
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
			clear(app.nav.regCache)
		}
		if gOpts.sixel {
			clear(app.nav.regCache)
			app.ui.sxScreen.forceClear = true
		}
		for _, dir := range app.nav.dirs {
			dir.boundPos(app.nav.height)
		}
		app.ui.loadFile(app, true)
		onRedraw(app)
	case "load":
		if !app.nav.init || gOpts.watch {
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
		app.ui.loadFileInfo(app.nav)
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
		for range e.count {
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
		for range e.count {
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
		for range e.count {
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
		for range e.count {
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
		app.ui.menu = listMarks(app.nav.marks)
		app.ui.cmdPrefix = "mark-load: "
	case "mark-remove":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.menu = listMarks(app.nav.marks)
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
			extension := getFileExtension(curr)
			if len(extension) == 0 {
				// no extension or .hidden or is directory
				app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(curr.Name())...)
			} else {
				app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(curr.Name()[:len(curr.Name())-len(extension)])...)
				app.ui.cmdAccRight = append(app.ui.cmdAccRight, []rune(extension)...)
			}
		}
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
	case "on-focus-gained":
		onFocusGained(app)
	case "on-focus-lost":
		onFocusLost(app)
	case "on-init":
		onInit(app)
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
		app.ui.menu = listMatches(app.ui.screen, matches, -1)
	case "cmd-menu-complete":
		menuComplete(app, 1)
	case "cmd-menu-complete-back":
		menuComplete(app, -1)
	case "cmd-menu-accept":
		app.ui.menu = ""
		app.menuCompActive = false
	case "cmd-enter":
		s := string(append(app.ui.cmdAccLeft, app.ui.cmdAccRight...))
		if len(s) == 0 && app.ui.cmdPrefix != "filter: " && app.ui.cmdPrefix != ">" {
			return
		}

		app.ui.menu = ""
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
		if !slices.Contains([]string{":", "$", "!", "%", "&"}, app.ui.cmdPrefix) {
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
		if !slices.Contains([]string{":", "$", "!", "%", "&", ""}, app.ui.cmdPrefix) {
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
		app.ui.cmdAccRight = append([]rune(string(app.ui.cmdAccLeft)[ind:]), old...)
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
		app.ui.cmdAccLeft = []rune(string(app.ui.cmdAccLeft)[:ind])
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

		app.ui.cmdAccLeft = slices.Concat(
			[]rune(string(app.ui.cmdAccLeft)[:beg1]),
			[]rune(string(app.ui.cmdAccLeft)[beg2:end2]),
			[]rune(string(app.ui.cmdAccLeft)[end1:beg2]),
			[]rune(string(app.ui.cmdAccLeft)[beg1:end1]),
			[]rune(string(app.ui.cmdAccLeft)[end2:]),
		)
		update(app)
	case "addcustominfo":
		var k, v string
		switch len(e.args) {
		case 1:
			k, v = e.args[0], ""
		case 2:
			k, v = e.args[0], e.args[1]
		default:
			app.ui.echoerr("addcustominfo: requires either 1 or 2 arguments")
			return
		}

		path, err := filepath.Abs(replaceTilde(k))
		if err != nil {
			app.ui.echoerrf("addcustominfo: %s", err)
			return
		}

		dir := filepath.Dir(path)
		d, ok := app.nav.dirCache[dir]
		if !ok {
			app.ui.echoerrf("addcustominfo: dir not loaded: %s", dir)
			return
		}

		var f *file
		for _, file := range d.allFiles {
			if file.path == path {
				f = file
				break
			}
		}
		if f == nil {
			app.ui.echoerrf("addcustominfo: file not found: %s", path)
			return
		}

		if len(strings.Trim(v, " ")) == 0 {
			v = ""
		}
		if f.customInfo != v {
			f.customInfo = v
			// only sort when order changes
			if getSortBy(dir) == customSort {
				d.sort()
			}
		}
	case "visual":
		if !app.nav.init {
			return
		}
		visual(app)
	case "visual-accept":
		if !app.nav.init {
			return
		}
		dir := app.nav.currDir()
		for _, path := range dir.visualSelections() {
			if _, ok := app.nav.selections[path]; !ok {
				app.nav.selections[path] = app.nav.selectionInd
				app.nav.selectionInd++
			}
		}
		// resetting visual mode here instead of inside `normal()`
		// allows us to use visual mode inside search, find etc.
		dir.visualAnchor = -1
		normal(app)
	case "visual-unselect":
		if !app.nav.init {
			return
		}
		dir := app.nav.currDir()
		for _, path := range dir.visualSelections() {
			delete(app.nav.selections, path)
		}
		if len(app.nav.selections) == 0 {
			app.nav.selectionInd = 0
		}
		dir.visualAnchor = -1
		normal(app)
	case "visual-discard":
		if !app.nav.init {
			return
		}
		dir := app.nav.currDir()
		dir.visualAnchor = -1
		normal(app)
	case "visual-change":
		if !app.nav.isVisualMode() {
			return
		}
		dir := app.nav.currDir()
		beg := max(dir.ind-dir.pos, 0)
		dir.ind, dir.visualAnchor = dir.visualAnchor, dir.ind
		dir.pos = dir.ind - beg
		dir.boundPos(app.nav.height)
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
	for range e.count {
		for _, expr := range e.exprs {
			expr.eval(app, nil)
		}
	}
}
