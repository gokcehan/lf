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
)

func (e *setExpr) eval(app *app, args []string) {
	switch e.opt {
	case "anchorfind":
		gOpts.anchorfind = true
	case "noanchorfind":
		gOpts.anchorfind = false
	case "anchorfind!":
		gOpts.anchorfind = !gOpts.anchorfind
	case "dircounts":
		gOpts.dircounts = true
	case "nodircounts":
		gOpts.dircounts = false
	case "dircounts!":
		gOpts.dircounts = !gOpts.dircounts
	case "dirfirst":
		gOpts.sortType.option |= dirfirstSort
		app.nav.sort()
		app.ui.sort()
	case "nodirfirst":
		gOpts.sortType.option &= ^dirfirstSort
		app.nav.sort()
		app.ui.sort()
	case "dirfirst!":
		gOpts.sortType.option ^= dirfirstSort
		app.nav.sort()
		app.ui.sort()
	case "drawbox":
		gOpts.drawbox = true
		app.ui.renew()
		app.nav.height = app.ui.wins[0].h
	case "nodrawbox":
		gOpts.drawbox = false
		app.ui.renew()
		app.nav.height = app.ui.wins[0].h
	case "drawbox!":
		gOpts.drawbox = !gOpts.drawbox
		app.ui.renew()
		app.nav.height = app.ui.wins[0].h
	case "globsearch":
		gOpts.globsearch = true
	case "noglobsearch":
		gOpts.globsearch = false
	case "globsearch!":
		gOpts.globsearch = !gOpts.globsearch
	case "hidden":
		gOpts.sortType.option |= hiddenSort
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app.nav, true)
	case "nohidden":
		gOpts.sortType.option &= ^hiddenSort
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app.nav, true)
	case "hidden!":
		gOpts.sortType.option ^= hiddenSort
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app.nav, true)
	case "icons":
		gOpts.icons = true
	case "noicons":
		gOpts.icons = false
	case "icons!":
		gOpts.icons = !gOpts.icons
	case "ignorecase":
		gOpts.ignorecase = true
		app.nav.sort()
		app.ui.sort()
	case "noignorecase":
		gOpts.ignorecase = false
		app.nav.sort()
		app.ui.sort()
	case "ignorecase!":
		gOpts.ignorecase = !gOpts.ignorecase
		app.nav.sort()
		app.ui.sort()
	case "ignoredia":
		gOpts.ignoredia = true
		app.nav.sort()
		app.ui.sort()
	case "noignoredia":
		gOpts.ignoredia = false
		app.nav.sort()
		app.ui.sort()
	case "ignoredia!":
		gOpts.ignoredia = !gOpts.ignoredia
		app.nav.sort()
		app.ui.sort()
	case "incsearch":
		gOpts.incsearch = true
	case "noincsearch":
		gOpts.incsearch = false
	case "incsearch!":
		gOpts.incsearch = !gOpts.incsearch
	case "mouse":
		if !gOpts.mouse {
			gOpts.mouse = true
			app.ui.screen.EnableMouse()
		}
	case "nomouse":
		if gOpts.mouse {
			gOpts.mouse = false
			app.ui.screen.DisableMouse()
		}
	case "mouse!":
		if gOpts.mouse {
			gOpts.mouse = false
			app.ui.screen.DisableMouse()
		} else {
			gOpts.mouse = true
			app.ui.screen.EnableMouse()
		}
	case "number":
		gOpts.number = true
	case "nonumber":
		gOpts.number = false
	case "number!":
		gOpts.number = !gOpts.number
	case "preview":
		if len(gOpts.ratios) < 2 {
			app.ui.echoerr("preview: 'ratios' should consist of at least two numbers before enabling 'preview'")
			return
		}
		gOpts.preview = true
	case "nopreview":
		gOpts.preview = false
	case "preview!":
		if len(gOpts.ratios) < 2 {
			app.ui.echoerr("preview: 'ratios' should consist of at least two numbers before enabling 'preview'")
			return
		}
		gOpts.preview = !gOpts.preview
	case "relativenumber":
		gOpts.relativenumber = true
	case "norelativenumber":
		gOpts.relativenumber = false
	case "relativenumber!":
		gOpts.relativenumber = !gOpts.relativenumber
	case "reverse":
		gOpts.sortType.option |= reverseSort
		app.nav.sort()
		app.ui.sort()
	case "noreverse":
		gOpts.sortType.option &= ^reverseSort
		app.nav.sort()
		app.ui.sort()
	case "reverse!":
		gOpts.sortType.option ^= reverseSort
		app.nav.sort()
		app.ui.sort()
	case "smartcase":
		gOpts.smartcase = true
	case "nosmartcase":
		gOpts.smartcase = false
	case "smartcase!":
		gOpts.smartcase = !gOpts.smartcase
	case "smartdia":
		gOpts.smartdia = true
	case "nosmartdia":
		gOpts.smartdia = false
	case "smartdia!":
		gOpts.smartdia = !gOpts.smartdia
	case "wrapscan":
		gOpts.wrapscan = true
	case "nowrapscan":
		gOpts.wrapscan = false
	case "wrapscan!":
		gOpts.wrapscan = !gOpts.wrapscan
	case "wrapscroll":
		gOpts.wrapscroll = true
	case "nowrapscroll":
		gOpts.wrapscroll = false
	case "wrapscroll!":
		gOpts.wrapscroll = !gOpts.wrapscroll
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
	case "errorfmt":
		gOpts.errorfmt = e.val
	case "filesep":
		gOpts.filesep = e.val
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
		app.ui.loadFile(app.nav, true)
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
			case "size", "time", "atime", "ctime":
			default:
				app.ui.echoerr("info: should consist of 'size', 'time', 'atime' or 'ctime' separated with colon")
				return
			}
		}
		gOpts.info = toks
	case "previewer":
		gOpts.previewer = replaceTilde(e.val)
	case "cleaner":
		gOpts.cleaner = replaceTilde(e.val)
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
		app.ui.loadFile(app.nav, true)
	case "shell":
		gOpts.shell = e.val
	case "shellopts":
		gOpts.shellopts = strings.Split(e.val, ":")
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
	case "timefmt":
		gOpts.timefmt = e.val
	case "truncatechar":
		if runeSliceWidth([]rune(e.val)) != 1 {
			app.ui.echoerr("truncatechar: value should be a single character")
			return
		}

		gOpts.truncatechar = e.val
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
	if e.cmd == "" {
		delete(gOpts.cmdkeys, e.key)
	} else {
		gOpts.cmdkeys[e.key] = &callExpr{e.cmd, nil, 1}
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

func onChdir(app *app) {
	if cmd, ok := gOpts.cmds["on-cd"]; ok {
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
	// Remove the current menu buffer
	app.ui.menuBuf = nil
	app.ui.menuSelected = -2

	switch {
	case gOpts.incsearch && app.ui.cmdPrefix == "/":
		app.nav.search = string(app.ui.cmdAccLeft) + string(app.ui.cmdAccRight)

		dir := app.nav.currDir()
		old := dir.ind
		dir.ind = app.nav.searchInd
		dir.pos = app.nav.searchPos

		if _, err := app.nav.searchNext(); err != nil {
			app.ui.echoerrf("search: %s: %s", err, app.nav.search)
		} else if old != dir.ind {
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case gOpts.incsearch && app.ui.cmdPrefix == "?":
		app.nav.search = string(app.ui.cmdAccLeft) + string(app.ui.cmdAccRight)

		dir := app.nav.currDir()
		old := dir.ind
		dir.ind = app.nav.searchInd
		dir.pos = app.nav.searchPos

		if _, err := app.nav.searchPrev(); err != nil {
			app.ui.echoerrf("search: %s: %s", err, app.nav.search)
		} else if old != dir.ind {
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	}
}

func normal(app *app) {
	app.ui.menuBuf = nil
	app.ui.menuSelected = -2

	app.ui.cmdAccLeft = nil
	app.ui.cmdAccRight = nil
	app.ui.cmdTmp = nil
	app.ui.cmdPrefix = ""
}

func insert(app *app, arg string) {
	switch {
	case gOpts.incsearch && (app.ui.cmdPrefix == "/" || app.ui.cmdPrefix == "?"):
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(arg)...)
		update(app)
	case app.ui.cmdPrefix == "find: ":
		app.nav.find = string(app.ui.cmdAccLeft) + arg + string(app.ui.cmdAccRight)

		if gOpts.findlen == 0 {
			switch app.nav.findSingle() {
			case 0:
				app.ui.echoerrf("find: pattern not found: %s", app.nav.find)
			case 1:
				app.ui.loadFile(app.nav, true)
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
				app.ui.loadFile(app.nav, true)
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
				app.ui.loadFile(app.nav, true)
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
				app.ui.loadFile(app.nav, true)
				app.ui.loadFileInfo(app.nav)
			}
		}

		normal(app)
	case strings.HasPrefix(app.ui.cmdPrefix, "delete"):
		normal(app)

		if arg == "y" {
			if err := app.nav.del(app.ui); err != nil {
				app.ui.echoerrf("delete: %s", err)
				return
			}
			app.nav.unselect()
			if err := remote("send load"); err != nil {
				app.ui.echoerrf("delete: %s", err)
				return
			}
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case strings.HasPrefix(app.ui.cmdPrefix, "replace"):
		normal(app)

		if arg == "y" {
			if err := app.nav.rename(); err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}
			if err := remote("send load"); err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}
			app.ui.loadFile(app.nav, true)
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
			if err := remote("send load"); err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case app.ui.cmdPrefix == "mark-save: ":
		normal(app)

		wd, err := os.Getwd()
		if err != nil {
			log.Printf("getting current directory: %s", err)
			return
		}
		app.nav.marks[arg] = wd
		if err := app.nav.writeMarks(); err != nil {
			app.ui.echoerrf("mark-save: %s", err)
		}
		if err := remote("send sync"); err != nil {
			app.ui.echoerrf("mark-save: %s", err)
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
		if err := app.nav.cd(path); err != nil {
			app.ui.echoerrf("%s", err)
			return
		}
		app.ui.loadFile(app.nav, true)
		app.ui.loadFileInfo(app.nav)

		if wd != path {
			app.nav.marks["'"] = wd
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
		if err := remote("send sync"); err != nil {
			app.ui.echoerrf("mark-remove: %s", err)
		}
	default:
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(arg)...)
	}
}

func (e *callExpr) eval(app *app, args []string) {
	switch e.name {
	case "up":
		if app.ui.cmdPrefix != "" && app.ui.cmdPrefix != ">" {
			normal(app)
		}
		if app.nav.up(e.count) {
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "half-up":
		if app.ui.cmdPrefix != "" && app.ui.cmdPrefix != ">" {
			normal(app)
		}
		if app.nav.up(e.count * app.nav.height / 2) {
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "page-up":
		if app.ui.cmdPrefix != "" && app.ui.cmdPrefix != ">" {
			normal(app)
		}
		if app.nav.up(e.count * app.nav.height) {
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "down":
		if app.ui.cmdPrefix != "" && app.ui.cmdPrefix != ">" {
			normal(app)
		}
		if app.nav.down(e.count) {
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "half-down":
		if app.ui.cmdPrefix != "" && app.ui.cmdPrefix != ">" {
			normal(app)
		}
		if app.nav.down(e.count * app.nav.height / 2) {
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "page-down":
		if app.ui.cmdPrefix != "" && app.ui.cmdPrefix != ">" {
			normal(app)
		}
		if app.nav.down(e.count * app.nav.height) {
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "updir":
		if app.ui.cmdPrefix != "" && app.ui.cmdPrefix != ">" {
			normal(app)
		}
		for i := 0; i < e.count; i++ {
			if err := app.nav.updir(); err != nil {
				app.ui.echoerrf("%s", err)
				return
			}
		}
		app.ui.loadFile(app.nav, true)
		app.ui.loadFileInfo(app.nav)
		onChdir(app)
	case "open":
		if app.ui.cmdPrefix != "" && app.ui.cmdPrefix != ">" {
			normal(app)
		}
		curr, err := app.nav.currFile()
		if err != nil {
			app.ui.echoerrf("opening: %s", err)
			return
		}

		if curr.IsDir() {
			err := app.nav.open()
			if err != nil {
				app.ui.echoerrf("opening directory: %s", err)
				return
			}
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
			onChdir(app)
			return
		}

		if gSelectionPath != "" {
			out, err := os.Create(gSelectionPath)
			if err != nil {
				log.Printf("opening selection file: %s", err)
				return
			}
			defer out.Close()

			var path string
			if list, err := app.nav.currFileOrSelections(); err == nil {
				path = strings.Join(list, "\n")
			} else {
				return
			}

			_, err = out.WriteString(path)
			if err != nil {
				log.Printf("writing selection file: %s", err)
			}

			app.quitChan <- struct{}{}

			return
		}

		if cmd, ok := gOpts.cmds["open"]; ok {
			cmd.eval(app, e.args)
		}
	case "quit":
		app.quitChan <- struct{}{}
	case "top":
		if app.ui.cmdPrefix != "" && app.ui.cmdPrefix != ">" {
			normal(app)
		}
		if app.nav.top() {
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "bottom":
		if app.ui.cmdPrefix != "" && app.ui.cmdPrefix != ">" {
			normal(app)
		}
		if app.nav.bottom() {
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "toggle":
		if len(e.args) == 0 {
			app.nav.toggle()
		} else {
			dir := app.nav.currDir()
			for _, path := range e.args {
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
	case "invert":
		app.nav.invert()
	case "unselect":
		app.nav.unselect()
	case "copy":
		if err := app.nav.save(true); err != nil {
			app.ui.echoerrf("copy: %s", err)
			return
		}
		app.nav.unselect()
		if err := remote("send sync"); err != nil {
			app.ui.echoerrf("copy: %s", err)
			return
		}
		app.ui.loadFileInfo(app.nav)
	case "cut":
		if err := app.nav.save(false); err != nil {
			app.ui.echoerrf("cut: %s", err)
			return
		}
		app.nav.unselect()
		if err := remote("send sync"); err != nil {
			app.ui.echoerrf("cut: %s", err)
			return
		}
		app.ui.loadFileInfo(app.nav)
	case "paste":
		if cmd, ok := gOpts.cmds["paste"]; ok {
			cmd.eval(app, e.args)
		} else if err := app.nav.paste(app.ui); err != nil {
			app.ui.echoerrf("paste: %s", err)
			return
		}
		app.ui.loadFile(app.nav, true)
		app.ui.loadFileInfo(app.nav)
	case "delete":
		if cmd, ok := gOpts.cmds["delete"]; ok {
			cmd.eval(app, e.args)
			app.nav.unselect()
			if err := remote("send load"); err != nil {
				app.ui.echoerrf("delete: %s", err)
				return
			}
		} else {
			list, err := app.nav.currFileOrSelections()
			if err != nil {
				app.ui.echoerrf("delete: %s", err)
				return
			}

			if len(list) == 1 {
				app.ui.cmdPrefix = "delete '" + list[0] + "' ? [y/N] "
			} else {
				app.ui.cmdPrefix = "delete " + strconv.Itoa(len(list)) + " items? [y/N] "
			}
		}
		app.ui.loadFile(app.nav, true)
		app.ui.loadFileInfo(app.nav)
	case "clear":
		if err := saveFiles(nil, false); err != nil {
			app.ui.echoerrf("clear: %s", err)
			return
		}
		if err := remote("send sync"); err != nil {
			app.ui.echoerrf("clear: %s", err)
			return
		}
		app.ui.loadFileInfo(app.nav)
	case "draw":
	case "redraw":
		app.ui.renew()
		app.ui.screen.Sync()
		app.nav.height = app.ui.wins[0].h
		app.ui.loadFile(app.nav, true)
	case "load":
		app.nav.renew()
		app.ui.loadFile(app.nav, true)
	case "reload":
		if err := app.nav.reload(); err != nil {
			app.ui.echoerrf("reload: %s", err)
		}
		app.ui.loadFile(app.nav, true)
		app.ui.loadFileInfo(app.nav)
	case "read":
		app.ui.cmdPrefix = ":"
		app.ui.loadFileInfo(app.nav)
	case "shell":
		app.ui.cmdPrefix = "$"
		app.ui.loadFileInfo(app.nav)
	case "shell-pipe":
		app.ui.cmdPrefix = "%"
		app.ui.loadFileInfo(app.nav)
	case "shell-wait":
		app.ui.cmdPrefix = "!"
		app.ui.loadFileInfo(app.nav)
	case "shell-async":
		app.ui.cmdPrefix = "&"
		app.ui.loadFileInfo(app.nav)
	case "find":
		app.ui.cmdPrefix = "find: "
		app.nav.findBack = false
		app.ui.loadFileInfo(app.nav)
	case "find-back":
		app.ui.cmdPrefix = "find-back: "
		app.nav.findBack = true
		app.ui.loadFileInfo(app.nav)
	case "find-next":
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
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "find-prev":
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
			app.ui.loadFile(app.nav, true)
			app.ui.loadFileInfo(app.nav)
		}
	case "search":
		app.ui.cmdPrefix = "/"
		dir := app.nav.currDir()
		app.nav.searchInd = dir.ind
		app.nav.searchPos = dir.pos
		app.nav.searchBack = false
		app.ui.loadFileInfo(app.nav)
	case "search-back":
		app.ui.cmdPrefix = "?"
		dir := app.nav.currDir()
		app.nav.searchInd = dir.ind
		app.nav.searchPos = dir.pos
		app.nav.searchBack = true
		app.ui.loadFileInfo(app.nav)
	case "search-next":
		for i := 0; i < e.count; i++ {
			if app.nav.searchBack {
				if moved, err := app.nav.searchPrev(); err != nil {
					app.ui.echoerrf("search-back: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app.nav, true)
					app.ui.loadFileInfo(app.nav)
				}
			} else {
				if moved, err := app.nav.searchNext(); err != nil {
					app.ui.echoerrf("search: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app.nav, true)
					app.ui.loadFileInfo(app.nav)
				}
			}
		}
	case "search-prev":
		for i := 0; i < e.count; i++ {
			if app.nav.searchBack {
				if moved, err := app.nav.searchNext(); err != nil {
					app.ui.echoerrf("search-back: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app.nav, true)
					app.ui.loadFileInfo(app.nav)
				}
			} else {
				if moved, err := app.nav.searchPrev(); err != nil {
					app.ui.echoerrf("search: %s: %s", err, app.nav.search)
				} else if moved {
					app.ui.loadFile(app.nav, true)
					app.ui.loadFileInfo(app.nav)
				}
			}
		}
	case "mark-save":
		app.ui.cmdPrefix = "mark-save: "
	case "mark-load":
		app.ui.menuBuf = listMarks(app.nav.marks)
		app.ui.cmdPrefix = "mark-load: "
	case "mark-remove":
		app.ui.menuBuf = listMarks(app.nav.marks)
		app.ui.cmdPrefix = "mark-remove: "
	case "rename":
		if cmd, ok := gOpts.cmds["rename"]; ok {
			cmd.eval(app, e.args)
			if err := remote("send load"); err != nil {
				app.ui.echoerrf("rename: %s", err)
				return
			}
		} else {
			curr, err := app.nav.currFile()
			if err != nil {
				app.ui.echoerrf("rename: %s:", err)
				return
			}
			app.ui.cmdPrefix = "rename: "
			app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(curr.Name())...)
		}
		app.ui.loadFile(app.nav, true)
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

		if err := app.nav.cd(path); err != nil {
			app.ui.echoerrf("%s", err)
			return
		}

		app.ui.loadFile(app.nav, true)
		app.ui.loadFileInfo(app.nav)

		if !filepath.IsAbs(path) {
			path = filepath.Join(wd, path)
		} else {
			path = filepath.Clean(path)
		}

		if wd != path {
			app.nav.marks["'"] = wd
			onChdir(app)
		}
	case "select":
		if len(e.args) != 1 {
			app.ui.echoerr("select: requires an argument")
			return
		}

		wd, err := os.Getwd()
		if err != nil {
			log.Printf("getting current directory: %s", err)
		}

		if err := app.nav.sel(e.args[0]); err != nil {
			app.ui.echoerrf("%s", err)
			return
		}

		app.ui.loadFile(app.nav, true)
		app.ui.loadFileInfo(app.nav)

		path := filepath.Dir(e.args[0])
		if !filepath.IsAbs(path) {
			path = filepath.Join(wd, path)
		} else {
			path = filepath.Clean(path)
		}

		if wd != path {
			app.nav.marks["'"] = wd
			onChdir(app)
		}
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

		// Reset the completition menu as in bash/vim
		// and update the pertinent variables
		if app.ui.menuBuf != nil {
			app.ui.menuBuf = nil
			app.ui.menuSelected = -2
		}

		insert(app, e.args[0])
	case "cmd-escape":
		if app.ui.cmdPrefix == ">" {
			return
		}
		if gOpts.incsearch && (app.ui.cmdPrefix == "/" || app.ui.cmdPrefix == "?") {
			dir := app.nav.currDir()
			dir.pos = app.nav.searchPos
			if dir.ind != app.nav.searchInd {
				dir.ind = app.nav.searchInd
				app.ui.loadFile(app.nav, true)
				app.ui.loadFileInfo(app.nav)
			}
		}
		normal(app)
	case "cmd-complete":
		var matches []string
		switch app.ui.cmdPrefix {
		case ":":
			matches, app.ui.cmdAccLeft = completeCmd(app.ui.cmdAccLeft)
		case "/", "?":
			matches, app.ui.cmdAccLeft = completeFile(app.ui.cmdAccLeft)
		case "$", "%", "!", "&":
			matches, app.ui.cmdAccLeft = completeShell(app.ui.cmdAccLeft)
		default:
			return
		}

		app.ui.draw(app.nav)

		// If there are no multiple matches
		// we do not need to show the list of matches
		if len(matches) <= 1 {
			app.ui.menuBuf = nil
			return
		}

		if b, err := listMatches(app.ui.screen, matches); err != nil {
			app.ui.echoerrf("cmd-complete: %s", err)
		} else {
			app.ui.menuBuf = b
		}
	case "cmd-menu-complete":
		var target []rune

		// target will store the current menu query
		// note that, as we will store the current selected value
		// in cmdAccLeft, we can not call the complete function with its
		// value, as it is already a final completition result
		if app.ui.menuBuf == nil {
			target = app.ui.cmdAccLeft
			app.ui.cmdTmp = app.ui.cmdAccLeft
		} else {
			target = app.ui.cmdTmp
		}

		var matches []string
		switch app.ui.cmdPrefix {
		case ":":
			matches, app.ui.cmdAccLeft = completeCmd(target)
		case "/", "?":
			matches, app.ui.cmdAccLeft = completeFile(target)
		case "$", "%", "!", "&":
			matches, app.ui.cmdAccLeft = completeShell(target)
		default:
			return
		}

		app.ui.draw(app.nav)

		if len(matches) > 1 {
			step := 1

			// Check if the tab-selecttion was inited
			if app.ui.menuSelected == -2 {
				app.ui.menuSelected = -1
			} else {
				app.ui.menuSelected = mod(app.ui.menuSelected+step, len(matches))
			}

			if err := listMatchesMenu(app.ui, matches); err != nil {
				app.ui.echoerrf("cmd-menu-complete: %s", err)
			}
		} else {
			app.ui.menuBuf = nil
			app.ui.menuSelected = -2
		}
	case "cmd-menu-complete-back":
		var target []rune

		if app.ui.menuBuf == nil {
			target = app.ui.cmdAccLeft
			app.ui.cmdTmp = app.ui.cmdAccLeft
		} else {
			target = app.ui.cmdTmp
		}

		var matches []string
		switch app.ui.cmdPrefix {
		case ":":
			matches, app.ui.cmdAccLeft = completeCmd(target)
		case "/", "?":
			matches, app.ui.cmdAccLeft = completeFile(target)
		case "$", "%", "!", "&":
			matches, app.ui.cmdAccLeft = completeShell(target)
		default:
			return
		}

		app.ui.draw(app.nav)

		if len(matches) > 1 {
			step := -1

			// Check if the tab-selecttion was inited
			if app.ui.menuSelected == -2 {
				app.ui.menuSelected = -1
			} else {
				app.ui.menuSelected = mod(app.ui.menuSelected+step, len(matches))
			}

			if err := listMatchesMenu(app.ui, matches); err != nil {
				app.ui.echoerrf("cmd-menu-complete-back: %s", err)
			}
		} else {
			app.ui.menuBuf = nil
			app.ui.menuSelected = -2
		}
	case "cmd-enter":
		s := string(append(app.ui.cmdAccLeft, app.ui.cmdAccRight...))
		if len(s) == 0 {
			return
		}

		app.ui.menuBuf = nil
		app.ui.menuSelected = -2

		app.ui.cmdAccLeft = nil
		app.ui.cmdAccRight = nil
		app.ui.cmdTmp = nil

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
				app.ui.loadFile(app.nav, true)
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
				app.ui.loadFile(app.nav, true)
				app.ui.loadFileInfo(app.nav)
			}
		case "rename: ":
			app.ui.cmdPrefix = ""
			if curr, err := app.nav.currFile(); err != nil {
				app.ui.echoerrf("rename: %s", err)
			} else {
				wd, err := os.Getwd()
				if err != nil {
					log.Printf("getting current directory: %s", err)
					return
				}

				oldPath := filepath.Join(wd, curr.Name())

				newPath := filepath.Clean(s)
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

				if err := remote("send load"); err != nil {
					app.ui.echoerrf("rename: %s", err)
					return
				}

				app.ui.loadFile(app.nav, true)
				app.ui.loadFileInfo(app.nav)
			}
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
			app.ui.menuBuf = nil
			app.ui.menuSelected = -2
			app.ui.cmdAccLeft = nil
			app.ui.cmdAccRight = nil
			app.ui.cmdTmp = nil
			app.ui.cmdPrefix = ":"
			return
		}
		cmd := app.cmdHistory[len(app.cmdHistory)-app.cmdHistoryInd]
		app.ui.cmdPrefix = cmd.prefix
		app.ui.cmdAccLeft = []rune(cmd.value)
		app.ui.cmdAccRight = nil
		app.ui.cmdTmp = nil
		app.ui.menuBuf = nil
		app.ui.menuSelected = -2
	case "cmd-history-prev":
		if app.ui.cmdPrefix == ">" {
			return
		}
		if app.ui.cmdPrefix == "" {
			app.cmdHistoryInd = 0
		}
		if app.cmdHistoryInd == len(app.cmdHistory) {
			return
		}
		app.cmdHistoryInd++
		cmd := app.cmdHistory[len(app.cmdHistory)-app.cmdHistoryInd]
		app.ui.cmdPrefix = cmd.prefix
		app.ui.cmdAccLeft = []rune(cmd.value)
		app.ui.cmdAccRight = nil
		app.ui.cmdTmp = nil
		app.ui.menuBuf = nil
		app.ui.menuSelected = -2
	case "cmd-delete":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		app.ui.cmdAccRight = app.ui.cmdAccRight[1:]
		update(app)
	case "cmd-delete-back":
		if len(app.ui.cmdAccLeft) == 0 {
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
			app.cmd.Process.Kill()
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
