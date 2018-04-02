package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (e *setExpr) eval(app *app, args []string) {
	switch e.opt {
	case "dircounts":
		gOpts.dircounts = true
	case "nodircounts":
		gOpts.dircounts = false
	case "dircounts!":
		gOpts.dircounts = !gOpts.dircounts
	case "dirfirst":
		gOpts.dirfirst = true
		app.nav.sort()
	case "nodirfirst":
		gOpts.dirfirst = false
		app.nav.sort()
	case "dirfirst!":
		gOpts.dirfirst = !gOpts.dirfirst
		app.nav.sort()
	case "globsearch":
		gOpts.globsearch = true
	case "noglobsearch":
		gOpts.globsearch = false
	case "globsearch!":
		gOpts.globsearch = !gOpts.globsearch
	case "hidden":
		gOpts.hidden = true
		app.nav.sort()
	case "nohidden":
		gOpts.hidden = false
		app.nav.sort()
	case "hidden!":
		gOpts.hidden = !gOpts.hidden
		app.nav.sort()
	case "ignorecase":
		gOpts.ignorecase = true
	case "lscolors":
		gOpts.lscolors = true
	case "nolscolors":
		gOpts.lscolors = false
	case "noignorecase":
		gOpts.ignorecase = false
	case "ignorecase!":
		gOpts.ignorecase = !gOpts.ignorecase
	case "preview":
		gOpts.preview = true
	case "nopreview":
		gOpts.preview = false
	case "preview!":
		gOpts.preview = !gOpts.preview
	case "reverse":
		gOpts.reverse = true
		app.nav.sort()
	case "noreverse":
		gOpts.reverse = false
		app.nav.sort()
	case "reverse!":
		gOpts.reverse = !gOpts.reverse
		app.nav.sort()
	case "smartcase":
		gOpts.smartcase = true
	case "nosmartcase":
		gOpts.smartcase = false
	case "smartcase!":
		gOpts.smartcase = !gOpts.smartcase
	case "wrapscan":
		gOpts.wrapscan = true
	case "nowrapscan":
		gOpts.wrapscan = false
	case "wrapscan!":
		gOpts.wrapscan = !gOpts.wrapscan
	case "scrolloff":
		n, err := strconv.Atoi(e.val)
		if err != nil {
			app.ui.printf("scrolloff: %s", err)
			return
		}
		if n < 0 {
			app.ui.print("scrolloff: value should be a non-negative number")
			return
		}
		max := app.ui.wins[0].h / 2
		if n > max {
			n = max
		}
		gOpts.scrolloff = n
	case "tabstop":
		n, err := strconv.Atoi(e.val)
		if err != nil {
			app.ui.printf("tabstop: %s", err)
			return
		}
		if n <= 0 {
			app.ui.print("tabstop: value should be a positive number")
			return
		}
		gOpts.tabstop = n
	case "filesep":
		gOpts.filesep = e.val
	case "ifs":
		gOpts.ifs = e.val
	case "previewer":
		gOpts.previewer = strings.Replace(e.val, "~", gUser.HomeDir, -1)
	case "promptfmt":
		gOpts.promptfmt = e.val
	case "shell":
		gOpts.shell = e.val
	case "sortby":
		if e.val != "natural" && e.val != "name" && e.val != "size" && e.val != "time" {
			app.ui.print("sortby: value should either be 'natural', 'name', 'size' or 'time'")
			return
		}
		gOpts.sortby = e.val
		app.nav.sort()
	case "timefmt":
		gOpts.timefmt = e.val
	case "ratios":
		toks := strings.Split(e.val, ":")
		var rats []int
		for _, s := range toks {
			n, err := strconv.Atoi(s)
			if err != nil {
				app.ui.printf("ratios: %s", err)
				return
			}
			if n <= 0 {
				app.ui.print("ratios: value should be a positive number")
				return
			}
			rats = append(rats, n)
		}
		if gOpts.preview && len(rats) < 2 {
			app.ui.print("ratios: should consist of at least two numbers when 'preview' is enabled")
			return
		}
		gOpts.ratios = rats
		app.ui.wins = getWins()
		app.ui.loadFile(app.nav)
	case "info":
		toks := strings.Split(e.val, ":")
		for _, s := range toks {
			if s != "" && s != "size" && s != "time" {
				app.ui.print("info: should consist of 'size' or 'time' separated with colon")
				return
			}
		}
		gOpts.info = toks
	default:
		app.ui.printf("unknown option: %s", e.opt)
	}
}

func (e *mapExpr) eval(app *app, args []string) {
	if e.expr == nil {
		delete(gOpts.keys, e.keys)
		return
	}
	gOpts.keys[e.keys] = e.expr
}

func (e *cmapExpr) eval(app *app, args []string) {
	if e.cmd == "" {
		delete(gOpts.cmdkeys, e.key)
		return
	}
	gOpts.cmdkeys[e.key] = &callExpr{e.cmd, nil}
}

func (e *cmdExpr) eval(app *app, args []string) {
	if e.expr == nil {
		delete(gOpts.cmds, e.name)
		return
	}
	gOpts.cmds[e.name] = e.expr
}

func splitKeys(s string) (keys []string) {
	for i := 0; i < len(s); {
		c, w := utf8.DecodeRuneInString(s[i:])
		if c != '<' {
			keys = append(keys, s[i:i+w])
			i += w
		} else {
			j := i + w
			for c != '>' && j < len(s) {
				c, w = utf8.DecodeRuneInString(s[j:])
				j += w
			}
			keys = append(keys, s[i:j])
			i = j
		}
	}
	return
}

func (e *callExpr) eval(app *app, args []string) {
	switch e.name {
	case "up":
		app.nav.up(1)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "half-up":
		app.nav.up(app.nav.height / 2)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "page-up":
		app.nav.up(app.nav.height)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "down":
		app.nav.down(1)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "half-down":
		app.nav.down(app.nav.height / 2)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "page-down":
		app.nav.down(app.nav.height)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "updir":
		if err := app.nav.updir(); err != nil {
			app.ui.printf("%s", err)
			return
		}
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "open":
		curr, err := app.nav.currFile()
		if err != nil {
			app.ui.printf("opening: %s", err)
			return
		}

		if curr.IsDir() {
			err := app.nav.open()
			if err != nil {
				app.ui.printf("opening directory: %s", err)
				return
			}
			app.ui.loadFile(app.nav)
			app.ui.loadFileInfo(app.nav)
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
			if len(app.nav.marks) != 0 {
				marks := app.nav.currMarks()
				path = strings.Join(marks, "\n")
			} else if curr, err := app.nav.currFile(); err == nil {
				path = curr.path
			} else {
				return
			}

			_, err = out.WriteString(path)
			if err != nil {
				log.Printf("writing selection file: %s", err)
			}

			app.quitChan <- true

			return
		}

		if cmd, ok := gOpts.cmds["open-file"]; ok {
			cmd.eval(app, e.args)
		}
	case "quit":
		app.quitChan <- true
	case "top":
		app.nav.top()
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "bot":
		app.nav.bot()
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "toggle":
		app.nav.toggle()
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "invert":
		app.nav.invert()
	case "unmark":
		app.nav.unmark()
	case "yank":
		if err := app.nav.save(true); err != nil {
			app.ui.printf("yank: %s", err)
			return
		}
		app.nav.unmark()
		if err := sendRemote("send sync"); err != nil {
			app.ui.printf("yank: %s", err)
		}
	case "delete":
		if err := app.nav.save(false); err != nil {
			app.ui.printf("delete: %s", err)
			return
		}
		app.nav.unmark()
		if err := sendRemote("send sync"); err != nil {
			app.ui.printf("delete: %s", err)
		}
	case "put":
		if cmd, ok := gOpts.cmds["put"]; ok {
			cmd.eval(app, e.args)
		} else if err := app.nav.put(); err != nil {
			app.ui.printf("put: %s", err)
			return
		}
		app.nav.renew(app.nav.height)
		if err := sendRemote("send sync"); err != nil {
			app.ui.printf("put: %s", err)
		}
	case "clear":
		if err := saveFiles(nil, false); err != nil {
			app.ui.printf("clear: %s", err)
			return
		}
		if err := sendRemote("send sync"); err != nil {
			app.ui.printf("clear: %s", err)
		}
	case "redraw":
		app.ui.sync()
		app.ui.renew()
		app.ui.loadFile(app.nav)
	case "reload":
		app.ui.sync()
		app.ui.renew()
		app.nav.reload()
	case "read":
		app.ui.cmdPrefix = ":"
	case "read-shell":
		app.ui.cmdPrefix = "$"
	case "read-shell-wait":
		app.ui.cmdPrefix = "!"
	case "read-shell-async":
		app.ui.cmdPrefix = "&"
	case "search":
		app.ui.cmdPrefix = "/"
	case "search-back":
		app.ui.cmdPrefix = "?"
	case "search-next":
		if err := app.nav.searchNext(); err != nil {
			app.ui.printf("search: %s: %s", err, app.nav.search)
			return
		}
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "search-prev":
		if err := app.nav.searchPrev(); err != nil {
			app.ui.printf("search: %s: %s", err, app.nav.search)
			return
		}
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "sync":
		if err := app.nav.sync(); err != nil {
			app.ui.printf("sync: %s", err)
		}
	case "echo":
		app.ui.msg = strings.Join(e.args, " ")
	case "cd":
		path := "~"
		if len(e.args) > 0 {
			path = e.args[0]
		}
		if err := app.nav.cd(path); err != nil {
			app.ui.printf("%s", err)
			return
		}
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "push":
		if len(e.args) > 0 {
			log.Println("pushing keys", e.args[0])
			for _, key := range splitKeys(e.args[0]) {
				app.ui.keyChan <- key
			}
		}
	case "cmd-insert":
		if len(e.args) > 0 {
			app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(e.args[0])...)
		}
	case "cmd-escape":
		app.ui.menuBuf = nil
		app.ui.cmdAccLeft = nil
		app.ui.cmdAccRight = nil
		app.ui.cmdPrefix = ""
	case "cmd-comp":
		var matches []string
		if app.ui.cmdPrefix == ":" {
			matches, app.ui.cmdAccLeft = compCmd(app.ui.cmdAccLeft)
		} else {
			matches, app.ui.cmdAccLeft = compShell(app.ui.cmdAccLeft)
		}
		app.ui.draw(app.nav)
		if len(matches) > 1 {
			app.ui.menuBuf = listMatches(matches)
		} else {
			app.ui.menuBuf = nil
		}
	case "cmd-enter":
		s := string(append(app.ui.cmdAccLeft, app.ui.cmdAccRight...))
		if len(s) == 0 {
			return
		}
		app.ui.menuBuf = nil
		app.ui.cmdAccLeft = nil
		app.ui.cmdAccRight = nil
		switch app.ui.cmdPrefix {
		case ":":
			log.Printf("command: %s", s)
			p := newParser(strings.NewReader(s))
			for p.parse() {
				p.expr.eval(app, nil)
			}
			if p.err != nil {
				app.ui.printf("%s", p.err)
			}
		case "$":
			log.Printf("shell: %s", s)
			app.runShell(s, nil, false, false)
		case "!":
			log.Printf("shell-wait: %s", s)
			app.runShell(s, nil, true, false)
		case "&":
			log.Printf("shell-async: %s", s)
			app.runShell(s, nil, false, true)
		case "/":
			log.Printf("search: %s", s)
			app.nav.search = s
			if err := app.nav.searchNext(); err != nil {
				app.ui.printf("search: %s: %s", err, app.nav.search)
			} else {
				app.ui.loadFile(app.nav)
				app.ui.loadFileInfo(app.nav)
			}
		case "?":
			log.Printf("search-back: %s", s)
			app.nav.search = s
			if err := app.nav.searchPrev(); err != nil {
				app.ui.printf("search: %s: %s", err, app.nav.search)
			} else {
				app.ui.loadFile(app.nav)
				app.ui.loadFileInfo(app.nav)
			}
		default:
			log.Printf("entering unknown execution prefix: %q", app.ui.cmdPrefix)
		}
		app.cmdHist = append(app.cmdHist, cmdItem{app.ui.cmdPrefix, s})
		app.ui.cmdPrefix = ""
	case "cmd-hist-next":
		if app.cmdHistInd > 0 {
			app.cmdHistInd--
		}
		if app.cmdHistInd == 0 {
			app.ui.menuBuf = nil
			app.ui.cmdAccLeft = nil
			app.ui.cmdAccRight = nil
			app.ui.cmdPrefix = ""
			return
		}
		cmd := app.cmdHist[len(app.cmdHist)-app.cmdHistInd]
		app.ui.cmdPrefix = cmd.prefix
		app.ui.cmdAccLeft = []rune(cmd.value)
		app.ui.cmdAccRight = nil
		app.ui.menuBuf = nil
	case "cmd-hist-prev":
		if app.cmdHistInd == len(app.cmdHist) {
			return
		}
		app.cmdHistInd++
		cmd := app.cmdHist[len(app.cmdHist)-app.cmdHistInd]
		app.ui.cmdPrefix = cmd.prefix
		app.ui.cmdAccLeft = []rune(cmd.value)
		app.ui.cmdAccRight = nil
		app.ui.menuBuf = nil
	case "cmd-delete":
		if len(app.ui.cmdAccRight) > 0 {
			app.ui.cmdAccRight = app.ui.cmdAccRight[1:]
		}
	case "cmd-delete-back":
		if len(app.ui.cmdAccLeft) > 0 {
			app.ui.cmdAccLeft = app.ui.cmdAccLeft[:len(app.ui.cmdAccLeft)-1]
		}
	case "cmd-left":
		if len(app.ui.cmdAccLeft) > 0 {
			app.ui.cmdAccRight = append([]rune{app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-1]}, app.ui.cmdAccRight...)
			app.ui.cmdAccLeft = app.ui.cmdAccLeft[:len(app.ui.cmdAccLeft)-1]
		}
	case "cmd-right":
		if len(app.ui.cmdAccRight) > 0 {
			app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, app.ui.cmdAccRight[0])
			app.ui.cmdAccRight = app.ui.cmdAccRight[1:]
		}
	case "cmd-beg":
		app.ui.cmdAccRight = append(app.ui.cmdAccLeft, app.ui.cmdAccRight...)
		app.ui.cmdAccLeft = nil
	case "cmd-end":
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, app.ui.cmdAccRight...)
		app.ui.cmdAccRight = nil
	case "cmd-delete-beg":
		if len(app.ui.cmdAccLeft) > 0 {
			app.ui.cmdBuf = app.ui.cmdAccLeft
			app.ui.cmdAccLeft = nil
		}
	case "cmd-delete-end":
		if len(app.ui.cmdAccRight) > 0 {
			app.ui.cmdBuf = app.ui.cmdAccRight
			app.ui.cmdAccRight = nil
		}
	case "cmd-delete-word":
		ind := strings.LastIndex(strings.TrimRight(string(app.ui.cmdAccLeft), " "), " ") + 1
		app.ui.cmdBuf = app.ui.cmdAccLeft[ind:]
		app.ui.cmdAccLeft = app.ui.cmdAccLeft[:ind]
	case "cmd-put":
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, app.ui.cmdBuf...)
	case "cmd-transpose":
		if len(app.ui.cmdAccLeft) > 1 {
			app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-1], app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-2] = app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-2], app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-1]
		}
	default:
		cmd, ok := gOpts.cmds[e.name]
		if !ok {
			app.ui.printf("command not found: %s", e.name)
			return
		}
		cmd.eval(app, e.args)
	}
}

func (e *execExpr) eval(app *app, args []string) {
	switch e.prefix {
	case "$":
		log.Printf("shell: %s -- %s", e, args)
		app.runShell(e.value, args, false, false)
	case "!":
		log.Printf("shell-wait: %s -- %s", e, args)
		app.runShell(e.value, args, true, false)
	case "&":
		log.Printf("shell-async: %s -- %s", e, args)
		app.runShell(e.value, args, false, true)
	case "/":
		log.Printf("search: %s -- %s", e, args)
		// TODO: implement
	case "?":
		log.Printf("search-back: %s -- %s", e, args)
		// TODO: implement
	default:
		log.Printf("evaluating unknown execution prefix: %q", e.prefix)
	}
}

func (e *listExpr) eval(app *app, args []string) {
	for _, expr := range e.exprs {
		expr.eval(app, nil)
	}
}
