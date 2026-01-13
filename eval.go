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

func (e *setExpr) eval(app *app, _ []string) {
	var err error
	switch e.opt {
	case "anchorfind", "noanchorfind", "anchorfind!":
		err = applyBoolOpt(&gOpts.anchorfind, e)
	case "autoquit", "noautoquit", "autoquit!":
		err = applyBoolOpt(&gOpts.autoquit, e)
	case "dircounts", "nodircounts", "dircounts!":
		err = applyBoolOpt(&gOpts.dircounts, e)
		if err == nil {
			app.nav.renew()
			app.ui.loadFile(app, false)
		}
	case "dirfirst", "nodirfirst", "dirfirst!":
		err = applyBoolOpt(&gOpts.dirfirst, e)
		if err == nil {
			app.nav.sort()
		}
	case "dironly", "nodironly", "dironly!":
		err = applyBoolOpt(&gOpts.dironly, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.loadFile(app, true)
		}
	case "dirpreviews", "nodirpreviews", "dirpreviews!":
		err = applyBoolOpt(&gOpts.dirpreviews, e)
	case "drawbox", "nodrawbox", "drawbox!":
		err = applyBoolOpt(&gOpts.drawbox, e)
		if err == nil {
			app.ui.renew()
			app.nav.resize(app.ui)
			app.ui.loadFile(app, true)
		}
	case "hidden", "nohidden", "hidden!":
		err = applyBoolOpt(&gOpts.hidden, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
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
			app.ui.loadFile(app, true)
		}
	case "ignoredia", "noignoredia", "ignoredia!":
		err = applyBoolOpt(&gOpts.ignoredia, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.loadFile(app, true)
		}
	case "incfilter", "noincfilter", "incfilter!":
		err = applyBoolOpt(&gOpts.incfilter, e)
	case "incsearch", "noincsearch", "incsearch!":
		err = applyBoolOpt(&gOpts.incsearch, e)
	case "mergeindicators", "nomergeindicators", "mergeindicators!":
		err = applyBoolOpt(&gOpts.mergeindicators, e)
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
	case "preload", "nopreload", "preload!":
		err = applyBoolOpt(&gOpts.preload, e)
	case "preview", "nopreview", "preview!":
		preview := gOpts.preview
		err = applyBoolOpt(&preview, e)
		if preview && len(gOpts.ratios) < 2 {
			err = errors.New("preview: 'ratios' should consist of at least two numbers before enabling 'preview'")
		}
		if err == nil {
			gOpts.preview = preview
			app.ui.sxScreen.forceClear = true
			app.ui.loadFile(app, true)
		}
	case "relativenumber", "norelativenumber", "relativenumber!":
		err = applyBoolOpt(&gOpts.relativenumber, e)
	case "reverse", "noreverse", "reverse!":
		err = applyBoolOpt(&gOpts.reverse, e)
		if err == nil {
			app.nav.sort()
		}
	case "roundbox", "noroundbox", "roundbox!":
		err = applyBoolOpt(&gOpts.roundbox, e)
	case "showbinds", "noshowbinds", "showbinds!":
		err = applyBoolOpt(&gOpts.showbinds, e)
	case "smartcase", "nosmartcase", "smartcase!":
		err = applyBoolOpt(&gOpts.smartcase, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.loadFile(app, true)
		}
	case "smartdia", "nosmartdia", "smartdia!":
		err = applyBoolOpt(&gOpts.smartdia, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
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
	case "filtermethod":
		switch e.val {
		case "text", "glob", "regex":
			gOpts.filtermethod = searchMethod(e.val)
		default:
			app.ui.echoerr("filtermethod: value should either be 'text', 'glob' or 'regex")
			return
		}
		app.nav.sort()
		app.nav.position()
		app.ui.loadFile(app, true)
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
	case "infotimefmtnew":
		gOpts.infotimefmtnew = e.val
	case "infotimefmtold":
		gOpts.infotimefmtold = e.val
	case "menufmt":
		gOpts.menufmt = e.val
	case "menuheaderfmt":
		gOpts.menuheaderfmt = e.val
	case "menuselectfmt":
		gOpts.menuselectfmt = e.val
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
				app.ui.echoerr("preserve: should consist of 'mode' or 'timestamps' separated with colon")
				return
			}
		}
		gOpts.preserve = toks
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
		app.nav.resize(app.ui)
		app.ui.loadFile(app, true)
	case "rulerfile", "norulerfile", "rulerfile!":
		gOpts.rulerfile = replaceTilde(e.val)
		app.ui.ruler, app.ui.rulerErr = parseRuler(gOpts.rulerfile)
	case "rulerfmt":
		gOpts.rulerfmt = e.val
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
	case "searchmethod":
		switch e.val {
		case "text", "glob", "regex":
			gOpts.searchmethod = searchMethod(e.val)
		default:
			app.ui.echoerr("searchmethod: value should either be 'text', 'glob' or 'regex'")
			return
		}
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
	case "sizeunits":
		switch e.val {
		case "binary", "decimal":
			gOpts.sizeunits = e.val
		default:
			app.ui.echoerr("sizeunits: value should either be 'binary' or 'decimal'")
			return
		}
	case "sortby":
		method := sortMethod(e.val)
		if !isValidSortMethod(method) {
			app.ui.echoerr(invalidSortErrorMessage)
			return
		}
		gOpts.sortby = method
		app.nav.sort()
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
	}
}

func (e *setLocalExpr) eval(app *app, _ []string) {
	recursive := strings.HasSuffix(e.path, string(os.PathSeparator)) && e.path != string(os.PathSeparator)
	if recursive {
		e.path = strings.TrimSuffix(e.path, string(os.PathSeparator))
	}

	var err error
	e.path, err = filepath.Abs(replaceTilde(e.path))
	if err != nil {
		app.ui.echoerrf("setlocal: %s", err)
		return
	}

	if recursive && e.path != string(os.PathSeparator) {
		e.path += string(os.PathSeparator)
	}

	switch e.opt {
	case "dircounts", "nodircounts", "dircounts!":
		err = applyLocalBoolOpt(gLocalOpts.dircounts, gOpts.dircounts, e)
	case "dirfirst", "nodirfirst", "dirfirst!":
		err = applyLocalBoolOpt(gLocalOpts.dirfirst, gOpts.dirfirst, e)
		if err == nil {
			app.nav.sort()
		}
	case "dironly", "nodironly", "dironly!":
		err = applyLocalBoolOpt(gLocalOpts.dironly, gOpts.dironly, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.loadFile(app, true)
		}
	case "hidden", "nohidden", "hidden!":
		err = applyLocalBoolOpt(gLocalOpts.hidden, gOpts.hidden, e)
		if err == nil {
			app.nav.sort()
			app.nav.position()
			app.ui.loadFile(app, true)
		}
	case "reverse", "noreverse", "reverse!":
		err = applyLocalBoolOpt(gLocalOpts.reverse, gOpts.reverse, e)
		if err == nil {
			app.nav.sort()
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
	default:
		err = fmt.Errorf("unknown option: %s", e.opt)
	}

	if err != nil {
		app.ui.echoerr(err.Error())
	}
}

func (e *mapExpr) eval(app *app, _ []string) {
	if e.expr == nil {
		delete(gOpts.nkeys, e.keys)
		delete(gOpts.vkeys, e.keys)
	} else {
		gOpts.nkeys[e.keys] = e.expr
		gOpts.vkeys[e.keys] = e.expr
	}
}

func (e *nmapExpr) eval(app *app, _ []string) {
	if e.expr == nil {
		delete(gOpts.nkeys, e.keys)
	} else {
		gOpts.nkeys[e.keys] = e.expr
	}
}

func (e *vmapExpr) eval(app *app, _ []string) {
	if e.expr == nil {
		delete(gOpts.vkeys, e.keys)
	} else {
		gOpts.vkeys[e.keys] = e.expr
	}
}

func (e *cmapExpr) eval(app *app, _ []string) {
	if e.expr == nil {
		delete(gOpts.cmdkeys, e.key)
	} else {
		gOpts.cmdkeys[e.key] = e.expr
	}
}

func (e *cmdExpr) eval(app *app, _ []string) {
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
	app.nav.preload()
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

func update(app *app) {
	exitCompMenu(app)
	app.cmdHistoryInput = nil

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
		}
	case gOpts.incfilter && app.ui.cmdPrefix == "filter: ":
		filter := string(app.ui.cmdAccLeft) + string(app.ui.cmdAccRight)
		dir := app.nav.currDir()
		old := dir.ind

		if err := app.nav.setFilter(strings.Split(filter, " ")); err != nil {
			app.ui.echoerrf("filter: %s", err)
		} else if old != dir.ind {
			app.ui.loadFile(app, true)
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
		}
	} else if gOpts.incfilter && app.ui.cmdPrefix == "filter: " {
		dir := app.nav.currDir()
		old := dir.ind
		if err := app.nav.setFilter(app.nav.prevFilter); err != nil {
			log.Printf("reset filter: %s", err)
		} else if old != dir.ind {
			app.ui.loadFile(app, true)
		}
	}
}

func normal(app *app) {
	resetIncCmd(app)
	exitCompMenu(app)

	app.cmdHistoryInd = 0
	app.cmdHistoryInput = nil

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
				if _, err := remote("send load"); err != nil {
					app.ui.echoerrf("rename: %s", err)
					return
				}
			}
			app.ui.loadFile(app, true)
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
				if _, err := remote("send load"); err != nil {
					app.ui.echoerrf("rename: %s", err)
					return
				}
			}
			app.ui.loadFile(app, true)
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
			if _, err := remote("send sync"); err != nil {
				app.ui.echoerrf("mark-save: %s", err)
				return
			}
		}
	case app.ui.cmdPrefix == "mark-load: ":
		normal(app)

		path, ok := app.nav.marks[arg]
		if !ok {
			app.ui.echoerr("mark-load: no such mark")
			return
		}

		if err := cd(app, path); err != nil {
			app.ui.echoerrf("mark-load: %s", err)
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
			if _, err := remote("send sync"); err != nil {
				app.ui.echoerrf("mark-remove: %s", err)
				return
			}
		}
	case app.ui.cmdPrefix == ":" && len(app.ui.cmdAccLeft) == 0:
		switch arg {
		case "!", "$", "%", "&":
			app.ui.cmdPrefix = arg
			app.cmdHistoryInd = 0
			app.cmdHistoryInput = nil
			return
		}
		fallthrough
	default:
		exitCompMenu(app)
		app.cmdHistoryInput = nil
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(arg)...)
	}
}

func cd(app *app, path string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	path, err = filepath.Abs(replaceTilde(path))
	if err != nil {
		return fmt.Errorf("getting absolute path: %w", err)
	}

	if path == wd {
		return nil
	}

	resetIncCmd(app)
	preChdir(app)

	if err := app.nav.cd(path); err != nil {
		return fmt.Errorf("changing directory: %w", err)
	}

	app.ui.loadFile(app, true)

	app.nav.marks["'"] = wd
	restartIncCmd(app)
	onChdir(app)

	return nil
}

func exitCompMenu(app *app) {
	app.ui.menu = ""
	app.ui.menuSelect = nil
	app.menuCompActive = false
}

func (e *callExpr) eval(app *app, _ []string) {
	os.Setenv("lf_count", strconv.Itoa(e.count))

	// commands that shouldn't clear the message line
	silentCmds := []string{
		"addcustominfo",
		"clearmaps",
		"draw",
		"load",
		"push",
		"redraw",
		"source",
		"sync",
		"tty-write",
		"on-focus-gained",
		"on-focus-lost",
		"on-init",
	}
	if !slices.Contains(silentCmds, e.name) && app.ui.cmdPrefix != ">" {
		app.ui.echo("")
	}

	switch e.name {
	case "quit":
		app.quitChan <- struct{}{}
	case "up":
		if app.nav.up(e.count) {
			app.ui.loadFile(app, true)
		}
	case "half-up":
		if app.nav.up(e.count * app.nav.height / 2) {
			app.ui.loadFile(app, true)
		}
	case "page-up":
		if app.nav.up(e.count * app.nav.height) {
			app.ui.loadFile(app, true)
		}
	case "scroll-up":
		if app.nav.scrollUp(e.count) {
			app.ui.loadFile(app, true)
		}
	case "down":
		if app.nav.down(e.count) {
			app.ui.loadFile(app, true)
		}
	case "half-down":
		if app.nav.down(e.count * app.nav.height / 2) {
			app.ui.loadFile(app, true)
		}
	case "page-down":
		if app.nav.down(e.count * app.nav.height) {
			app.ui.loadFile(app, true)
		}
	case "scroll-down":
		if app.nav.scrollDown(e.count) {
			app.ui.loadFile(app, true)
		}
	case "updir":
		resetIncCmd(app)
		preChdir(app)
		for range e.count {
			if err := app.nav.updir(); err != nil {
				app.ui.echoerrf("%s", err)
				return
			}
		}
		app.ui.loadFile(app, true)
		restartIncCmd(app)
		onChdir(app)
	case "open":
		curr := app.nav.currFile()
		if curr == nil {
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
			restartIncCmd(app)
			onChdir(app)
		} else {
			if gSelectionPath != "" || gPrintSelection {
				app.selectionOut, _ = app.nav.currFileOrSelections()
				app.quitChan <- struct{}{}
				return
			}

			if cmd, ok := gOpts.cmds["open"]; ok {
				cmd.eval(app, e.args)
			}
		}
	case "jump-next":
		resetIncCmd(app)
		preChdir(app)
		for range e.count {
			app.nav.cdJumpListNext()
		}
		app.ui.loadFile(app, true)
		restartIncCmd(app)
		onChdir(app)
	case "jump-prev":
		resetIncCmd(app)
		preChdir(app)
		for range e.count {
			app.nav.cdJumpListPrev()
		}
		app.ui.loadFile(app, true)
		restartIncCmd(app)
		onChdir(app)
	case "top":
		var moved bool
		if e.count == 1 {
			moved = app.nav.top()
		} else {
			moved = app.nav.move(e.count - 1)
		}
		if moved {
			app.ui.loadFile(app, true)
		}
	case "bottom":
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
		}
	case "high":
		if app.nav.high() {
			app.ui.loadFile(app, true)
		}
	case "middle":
		if app.nav.middle() {
			app.ui.loadFile(app, true)
		}
	case "low":
		if app.nav.low() {
			app.ui.loadFile(app, true)
		}
	case "toggle":
		if len(e.args) == 0 {
			app.nav.toggle()
		} else {
			for _, path := range e.args {
				path, err := filepath.Abs(replaceTilde(path))
				if err != nil {
					app.ui.echoerrf("toggle: %s", err)
					continue
				}

				if _, err := os.Lstat(path); os.IsNotExist(err) {
					app.ui.echoerrf("toggle: %s", err)
					continue
				}

				app.nav.toggleSelection(path)
			}
		}
	case "invert":
		app.nav.invert()
	case "unselect":
		app.nav.unselect()
	case "glob-select":
		if len(e.args) != 1 {
			app.ui.echoerr("glob-select: requires a pattern to match")
			return
		}
		if err := app.nav.globSel(e.args[0], false); err != nil {
			app.ui.echoerrf("%s", err)
			return
		}
	case "glob-unselect":
		if len(e.args) != 1 {
			app.ui.echoerr("glob-unselect: requires a pattern to match")
			return
		}
		if err := app.nav.globSel(e.args[0], true); err != nil {
			app.ui.echoerrf("%s", err)
			return
		}
	case "copy":
		if err := app.nav.save(clipboardCopy); err != nil {
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
			if _, err := remote("send sync"); err != nil {
				app.ui.echoerrf("copy: %s", err)
				return
			}
		}
	case "cut":
		if err := app.nav.save(clipboardCut); err != nil {
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
			if _, err := remote("send sync"); err != nil {
				app.ui.echoerrf("cut: %s", err)
				return
			}
		}
	case "paste":
		if cmd, ok := gOpts.cmds["paste"]; ok {
			cmd.eval(app, e.args)
		} else if err := app.nav.paste(app); err != nil {
			app.ui.echoerrf("paste: %s", err)
			return
		}
		app.ui.loadFile(app, true)
	case "clear":
		if err := saveFiles(clipboard{nil, clipboardCut}); err != nil {
			app.ui.echoerrf("clear: %s", err)
			return
		}
		if gSingleMode {
			if err := app.nav.sync(); err != nil {
				app.ui.echoerrf("clear: %s", err)
				return
			}
		} else {
			if _, err := remote("send sync"); err != nil {
				app.ui.echoerrf("clear: %s", err)
				return
			}
		}
	case "sync":
		if err := app.nav.sync(); err != nil {
			app.ui.echoerrf("sync: %s", err)
		}
	case "draw":
	case "redraw":
		app.ui.screen.Sync()
		app.ui.renew()
		app.nav.resize(app.ui)
		app.ui.sxScreen.forceClear = true
		app.ui.loadFile(app, true)
		onRedraw(app)
	case "load":
		if gOpts.watch {
			return
		}
		app.nav.renew()
		app.ui.loadFile(app, false)
	case "reload":
		app.nav.reload()
		app.ui.loadFile(app, true)
	case "delete":
		if cmd, ok := gOpts.cmds["delete"]; ok {
			cmd.eval(app, e.args)
			app.nav.unselect()
			if gSingleMode {
				app.nav.renew()
				app.ui.loadFile(app, true)
			} else {
				if _, err := remote("send load"); err != nil {
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
				app.ui.cmdPrefix = "delete '" + list[0] + "'? [y/N] "
			} else {
				app.ui.cmdPrefix = "delete " + strconv.Itoa(len(list)) + " items? [y/N] "
			}
		}
	case "rename":
		if cmd, ok := gOpts.cmds["rename"]; ok {
			cmd.eval(app, e.args)
			if gSingleMode {
				app.nav.renew()
				app.ui.loadFile(app, true)
			} else {
				if _, err := remote("send load"); err != nil {
					app.ui.echoerrf("rename: %s", err)
					return
				}
			}
		} else {
			curr := app.nav.currFile()
			if curr == nil {
				app.ui.echoerr("rename: empty directory")
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
	case "read":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = ":"
	case "shell":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "$"
	case "shell-pipe":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "%"
	case "shell-wait":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "!"
	case "shell-async":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "&"
	case "find":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "find: "
		app.nav.findBack = false
	case "find-back":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "find-back: "
		app.nav.findBack = true
	case "find-next":
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
		}
	case "find-prev":
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
		}
	case "search":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "/"
		dir := app.nav.currDir()
		app.nav.searchInd = dir.ind
		app.nav.searchPos = dir.pos
		app.nav.searchBack = false
	case "search-back":
		if app.ui.cmdPrefix == ">" {
			return
		}
		normal(app)
		app.ui.cmdPrefix = "?"
		dir := app.nav.currDir()
		app.nav.searchInd = dir.ind
		app.nav.searchPos = dir.pos
		app.nav.searchBack = true
	case "search-next":
		for range e.count {
			if app.nav.searchBack {
				if moved, err := app.nav.searchPrev(); err != nil {
					app.ui.echoerrf("search-back: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app, true)
				}
			} else {
				if moved, err := app.nav.searchNext(); err != nil {
					app.ui.echoerrf("search: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app, true)
				}
			}
		}
	case "search-prev":
		for range e.count {
			if app.nav.searchBack {
				if moved, err := app.nav.searchNext(); err != nil {
					app.ui.echoerrf("search-back: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app, true)
				}
			} else {
				if moved, err := app.nav.searchPrev(); err != nil {
					app.ui.echoerrf("search: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app, true)
				}
			}
		}
	case "filter":
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
	case "setfilter":
		log.Printf("filter: %s", e.args)
		if err := app.nav.setFilter(e.args); err != nil {
			app.ui.echoerrf("filter: %s", err)
		}
		app.ui.loadFile(app, true)
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
	case "tag":
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
			if _, err := remote("send sync"); err != nil {
				app.ui.echoerrf("tag: %s", err)
			}
		}
	case "tag-toggle":
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
			if _, err := remote("send sync"); err != nil {
				app.ui.echoerrf("tag-toggle: %s", err)
			}
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

		if err := cd(app, path); err != nil {
			app.ui.echoerrf("cd: %s", err)
		}
	case "select":
		if len(e.args) != 1 {
			app.ui.echoerr("select: requires an argument")
			return
		}

		path, err := filepath.Abs(replaceTilde(e.args[0]))
		if err != nil {
			app.ui.echoerrf("select: %s", err)
			return
		}

		lstat, err := os.Lstat(path)
		if err != nil {
			app.ui.echoerrf("select: %s", err)
			return
		}

		if err := cd(app, filepath.Dir(path)); err != nil {
			app.ui.echoerrf("select: %s", err)
			return
		}

		dir := app.nav.currDir()
		if dir.loading {
			dir.files = append(dir.files, &file{FileInfo: lstat})
		} else {
			app.nav.currDir().sel(filepath.Base(path), app.nav.height)
			app.ui.loadFile(app, true)
		}
	case "source":
		if len(e.args) != 1 {
			app.ui.echoerr("source: requires an argument")
			return
		}
		app.readFile(replaceTilde(e.args[0]))
	case "push":
		if len(e.args) != 1 {
			app.ui.echoerr("push: requires an argument")
			return
		}
		log.Println("pushing keys", e.args[0])
		for _, val := range splitKeys(e.args[0]) {
			app.ui.keyChan <- val
		}
	case "addcustominfo":
		var k, v string
		switch len(e.args) {
		case 1:
			k, v = e.args[0], ""
		case 2:
			k, v = e.args[0], e.args[1] // don't trim to allow for custom alignment
		default:
			app.ui.echoerr("addcustominfo: requires either 1 or 2 arguments")
			return
		}

		path, err := filepath.Abs(replaceTilde(k))
		if err != nil {
			app.ui.echoerrf("addcustominfo: %s", err)
			return
		}

		d := app.nav.getDir(filepath.Dir(path))

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

		if f.customInfo != v {
			f.customInfo = v
			// only sort when order changes
			if getSortBy(d.path) == customSort {
				d.sort()
			}
		}
	case "calcdirsize":
		err := app.nav.calcDirSize()
		if err != nil {
			app.ui.echoerrf("calcdirsize: %s", err)
			return
		}
		app.nav.sort()
	case "clearmaps":
		// leave `:` and cmaps bound so the user can still exit using `:quit`
		clear(gOpts.nkeys)
		clear(gOpts.vkeys)
		gOpts.nkeys[":"] = &callExpr{"read", nil, 1}
		gOpts.vkeys[":"] = &callExpr{"read", nil, 1}
	case "tty-write":
		if len(e.args) != 1 {
			app.ui.echoerr("tty-write: requires an argument")
			return
		}

		tty, ok := app.ui.screen.Tty()
		if !ok {
			log.Print("tty-write: failed to get tty")
			return
		}

		tty.Write([]byte(e.args[0]))
	case "visual":
		dir := app.nav.currDir()
		dir.visualAnchor = dir.ind
		dir.visualWrap = 0
	case "visual-accept":
		dir := app.nav.currDir()
		for _, path := range dir.visualSelections() {
			if _, ok := app.nav.selections[path]; !ok {
				app.nav.selections[path] = app.nav.selectionInd
				app.nav.selectionInd++
			}
		}
		// resetting Visual mode here instead of inside `normal()`
		// allows us to use Visual mode inside search, find etc.
		dir.visualAnchor = -1
		normal(app)
	case "visual-unselect":
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
		dir.visualWrap = -dir.visualWrap
		dir.boundPos(app.nav.height)
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
		app.doComplete()
	case "cmd-menu-complete":
		app.menuComplete(1)
	case "cmd-menu-complete-back":
		app.menuComplete(-1)
	case "cmd-menu-accept":
		exitCompMenu(app)
	case "cmd-menu-discard":
		if app.menuCompActive {
			app.ui.cmdAccLeft = []rune(strings.Join(app.menuCompTmp, " "))
		}
		exitCompMenu(app)
	case "cmd-enter":
		s := string(append(app.ui.cmdAccLeft, app.ui.cmdAccRight...))
		if len(s) == 0 && app.ui.cmdPrefix != "filter: " && app.ui.cmdPrefix != ">" {
			return
		}

		exitCompMenu(app)

		app.ui.cmdAccLeft = nil
		app.ui.cmdAccRight = nil

		switch app.ui.cmdPrefix {
		case ":":
			log.Printf("command: %s", s)
			app.cmdHistory = append(app.cmdHistory, app.ui.cmdPrefix+s)
			app.ui.cmdPrefix = ""
			p := newParser(strings.NewReader(s))
			for p.parse() {
				p.expr.eval(app, nil)
			}
			if p.err != nil {
				app.ui.echoerrf("%s", p.err)
			}
		case "$":
			log.Printf("shell: %s", s)
			app.cmdHistory = append(app.cmdHistory, app.ui.cmdPrefix+s)
			app.ui.cmdPrefix = ""
			app.runShell(s, nil, "$")
		case "%":
			log.Printf("shell-pipe: %s", s)
			app.cmdHistory = append(app.cmdHistory, app.ui.cmdPrefix+s)
			app.runShell(s, nil, "%")
		case ">":
			io.WriteString(app.cmdIn, s+"\n")
			app.cmdOutBuf = nil
		case "!":
			log.Printf("shell-wait: %s", s)
			app.cmdHistory = append(app.cmdHistory, app.ui.cmdPrefix+s)
			app.ui.cmdPrefix = ""
			app.runShell(s, nil, "!")
		case "&":
			log.Printf("shell-async: %s", s)
			app.cmdHistory = append(app.cmdHistory, app.ui.cmdPrefix+s)
			app.ui.cmdPrefix = ""
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
			}
		case "filter: ":
			log.Printf("filter: %s", s)
			app.ui.cmdPrefix = ""
			if err := app.nav.setFilter(strings.Split(s, " ")); err != nil {
				app.ui.echoerrf("filter: %s", err)
			}
			app.ui.loadFile(app, true)
		case "find: ":
			app.ui.cmdPrefix = ""
			if moved, found := app.nav.findNext(); !found {
				app.ui.echoerrf("find: pattern not found: %s", app.nav.find)
			} else if moved {
				app.ui.loadFile(app, true)
			}
		case "find-back: ":
			app.ui.cmdPrefix = ""
			if moved, found := app.nav.findPrev(); !found {
				app.ui.echoerrf("find-back: pattern not found: %s", app.nav.find)
			} else if moved {
				app.ui.loadFile(app, true)
			}
		case "rename: ":
			app.ui.cmdPrefix = ""

			curr := app.nav.currFile()
			if curr == nil {
				app.ui.echoerr("rename: empty directory")
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
				app.ui.cmdPrefix = "create '" + newDir + "'? [y/N] "
				return
			}

			oldStat, err := os.Lstat(oldPath)
			if err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}
			if newStat, err := os.Lstat(newPath); !os.IsNotExist(err) && !os.SameFile(oldStat, newStat) {
				app.ui.cmdPrefix = "replace '" + newPath + "'? [y/N] "
				return
			}

			if err := app.nav.rename(); err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}

			if gSingleMode {
				app.nav.renew()
			} else {
				if _, err := remote("send load"); err != nil {
					app.ui.echoerrf("rename: %s", err)
					return
				}
			}
			app.ui.loadFile(app, true)
		default:
			log.Printf("entering unknown execution prefix: %q", app.ui.cmdPrefix)
		}
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
	case "cmd-history-next":
		if !slices.Contains([]string{":", "$", "!", "%", "&"}, app.ui.cmdPrefix) {
			return
		}
		input := app.ui.cmdPrefix + string(app.ui.cmdAccLeft)
		if app.cmdHistoryInput == nil {
			app.cmdHistoryInput = &input
		}
		for i := app.cmdHistoryInd - 1; i >= 0; i-- {
			if i == 0 {
				if *app.cmdHistoryInput == "" {
					normal(app)
				} else {
					exitCompMenu(app)
					app.ui.cmdAccLeft = nil
					app.cmdHistoryInd = 0
				}
				break
			}
			cmd := app.cmdHistory[len(app.cmdHistory)-i]
			if strings.HasPrefix(cmd, *app.cmdHistoryInput) && cmd != input {
				exitCompMenu(app)
				app.ui.cmdPrefix = cmd[:1]
				app.ui.cmdAccLeft = []rune(cmd[1:])
				app.cmdHistoryInd = i
				break
			}
		}
	case "cmd-history-prev":
		if !slices.Contains([]string{":", "$", "!", "%", "&", ""}, app.ui.cmdPrefix) {
			return
		}
		input := app.ui.cmdPrefix + string(app.ui.cmdAccLeft)
		if app.cmdHistoryInput == nil {
			app.cmdHistoryInput = &input
		}
		for i := app.cmdHistoryInd + 1; i <= len(app.cmdHistory); i++ {
			cmd := app.cmdHistory[len(app.cmdHistory)-i]
			if strings.HasPrefix(cmd, *app.cmdHistoryInput) && cmd != input {
				exitCompMenu(app)
				app.ui.cmdPrefix = cmd[:1]
				app.ui.cmdAccLeft = []rune(cmd[1:])
				app.cmdHistoryInd = i
				break
			}
		}
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
				app.cmdHistoryInd = 0
				app.cmdHistoryInput = nil
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
	case "cmd-capitalize-word":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		ind := 0
		for ind < len(app.ui.cmdAccRight) && unicode.IsSpace(app.ui.cmdAccRight[ind]) {
			ind++
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
	case "on-focus-gained":
		onFocusGained(app)
	case "on-focus-lost":
		onFocusLost(app)
	case "on-init":
		onInit(app)
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

func (e *listExpr) eval(app *app, _ []string) {
	for range e.count {
		for _, expr := range e.exprs {
			expr.eval(app, nil)
		}
	}
}
