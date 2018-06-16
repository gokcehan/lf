package main

import (
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
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
		app.ui.loadFile(app.nav)
	case "nohidden":
		gOpts.sortType.option &= ^hiddenSort
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app.nav)
	case "hidden!":
		gOpts.sortType.option ^= hiddenSort
		app.nav.sort()
		app.nav.position()
		app.ui.sort()
		app.ui.loadFile(app.nav)
	case "ignorecase":
		gOpts.ignorecase = true
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
	case "wrapscan":
		gOpts.wrapscan = true
	case "nowrapscan":
		gOpts.wrapscan = false
	case "wrapscan!":
		gOpts.wrapscan = !gOpts.wrapscan
	case "period":
		n, err := strconv.Atoi(e.val)
		if err != nil {
			app.ui.printf("period: %s", err)
			return
		}
		if n < 0 {
			app.ui.print("period: value should be a non-negative number")
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
			app.ui.printf("scrolloff: %s", err)
			return
		}
		if n < 0 {
			app.ui.print("scrolloff: value should be a non-negative number")
			return
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
		switch e.val {
		case "natural":
			gOpts.sortType.method = naturalSort
		case "name":
			gOpts.sortType.method = nameSort
		case "size":
			gOpts.sortType.method = sizeSort
		case "time":
			gOpts.sortType.method = timeSort
		}
		app.nav.sort()
		app.ui.sort()
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
	gOpts.cmdkeys[e.key] = &callExpr{e.cmd, nil, 1}
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

func (e *callExpr) eval(app *app, args []string) {
	switch e.name {
	case "up":
		app.nav.up(e.count)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "half-up":
		app.nav.up(e.count * app.nav.height / 2)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "page-up":
		app.nav.up(e.count * app.nav.height)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "down":
		app.nav.down(e.count)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "half-down":
		app.nav.down(e.count * app.nav.height / 2)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "page-down":
		app.nav.down(e.count * app.nav.height)
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "updir":
		for i := 0; i < e.count; i++ {
			if err := app.nav.updir(); err != nil {
				app.ui.printf("%s", err)
				return
			}
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
	case "bottom":
		app.nav.bottom()
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "toggle":
		for i := 0; i < e.count; i++ {
			app.nav.toggle()
		}
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
		if err := sendRemote("send load"); err != nil {
			app.ui.printf("put: %s", err)
		}
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
	case "draw":
	case "redraw":
		app.ui.sync()
		app.ui.renew()
		app.nav.height = app.ui.wins[0].h
	case "load":
		app.nav.renew()
	case "reload":
		app.nav.reload()
	case "read":
		app.ui.cmdPrefix = ":"
	case "shell":
		app.ui.cmdPrefix = "$"
	case "shell-pipe":
		app.ui.cmdPrefix = "%"
	case "shell-wait":
		app.ui.cmdPrefix = "!"
	case "shell-async":
		app.ui.cmdPrefix = "&"
	case "search":
		app.ui.cmdPrefix = "/"
	case "search-back":
		app.ui.cmdPrefix = "?"
	case "search-next":
		for i := 0; i < e.count; i++ {
			if err := app.nav.searchNext(); err != nil {
				app.ui.printf("search: %s: %s", err, app.nav.search)
				return
			}
		}
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "search-prev":
		for i := 0; i < e.count; i++ {
			if err := app.nav.searchPrev(); err != nil {
				app.ui.printf("search: %s: %s", err, app.nav.search)
				return
			}
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
	case "select":
		if len(e.args) != 1 {
			app.ui.print("select: requires an argument")
			return
		}
		if err := app.nav.find(e.args[0]); err != nil {
			app.ui.printf("%s", err)
			return
		}
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "push":
		if len(e.args) != 1 {
			app.ui.print("push: requires an argument")
			return
		}
		log.Println("pushing keys", e.args[0])
		for _, key := range splitKeys(e.args[0]) {
			app.ui.keyChan <- key
		}
	case "cmd-insert":
		if len(e.args) == 0 {
			return
		}
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(e.args[0])...)
	case "cmd-escape":
		if app.ui.cmdPrefix == ">" {
			return
		}
		app.ui.menuBuf = nil
		app.ui.cmdAccLeft = nil
		app.ui.cmdAccRight = nil
		app.ui.cmdPrefix = ""
	case "cmd-complete":
		var matches []string
		if app.ui.cmdPrefix == ":" {
			matches, app.ui.cmdAccLeft = completeCmd(app.ui.cmdAccLeft)
		} else {
			matches, app.ui.cmdAccLeft = completeShell(app.ui.cmdAccLeft)
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
			app.cmdHistory = append(app.cmdHistory, cmdItem{app.ui.cmdPrefix, s})
			app.ui.cmdPrefix = ""
		case "$":
			log.Printf("shell: %s", s)
			app.runShell(s, nil, app.ui.cmdPrefix)
			app.cmdHistory = append(app.cmdHistory, cmdItem{app.ui.cmdPrefix, s})
			app.ui.cmdPrefix = ""
		case "%":
			log.Printf("shell-pipe: %s", s)
			app.runShell(s, nil, app.ui.cmdPrefix)
			app.cmdHistory = append(app.cmdHistory, cmdItem{app.ui.cmdPrefix, s})
		case ">":
			io.WriteString(app.cmdIn, s+"\n")
			app.cmdOutBuf = nil
		case "!":
			log.Printf("shell-wait: %s", s)
			app.runShell(s, nil, app.ui.cmdPrefix)
			app.cmdHistory = append(app.cmdHistory, cmdItem{app.ui.cmdPrefix, s})
			app.ui.cmdPrefix = ""
		case "&":
			log.Printf("shell-async: %s", s)
			app.runShell(s, nil, app.ui.cmdPrefix)
			app.cmdHistory = append(app.cmdHistory, cmdItem{app.ui.cmdPrefix, s})
			app.ui.cmdPrefix = ""
		case "/":
			log.Printf("search: %s", s)
			app.nav.search = s
			if err := app.nav.searchNext(); err != nil {
				app.ui.printf("search: %s: %s", err, app.nav.search)
			} else {
				app.ui.loadFile(app.nav)
				app.ui.loadFileInfo(app.nav)
			}
			app.ui.cmdPrefix = ""
		case "?":
			log.Printf("search-back: %s", s)
			app.nav.search = s
			if err := app.nav.searchPrev(); err != nil {
				app.ui.printf("search: %s: %s", err, app.nav.search)
			} else {
				app.ui.loadFile(app.nav)
				app.ui.loadFileInfo(app.nav)
			}
			app.ui.cmdPrefix = ""
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
			app.ui.cmdAccLeft = nil
			app.ui.cmdAccRight = nil
			app.ui.cmdPrefix = ""
			return
		}
		cmd := app.cmdHistory[len(app.cmdHistory)-app.cmdHistoryInd]
		app.ui.cmdPrefix = cmd.prefix
		app.ui.cmdAccLeft = []rune(cmd.value)
		app.ui.cmdAccRight = nil
		app.ui.menuBuf = nil
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
		app.ui.menuBuf = nil
	case "cmd-delete":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		app.ui.cmdAccRight = app.ui.cmdAccRight[1:]
	case "cmd-delete-back":
		if len(app.ui.cmdAccLeft) == 0 {
			return
		}
		app.ui.cmdAccLeft = app.ui.cmdAccLeft[:len(app.ui.cmdAccLeft)-1]
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
	case "cmd-delete-end":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		app.ui.cmdYankBuf = app.ui.cmdAccRight
		app.ui.cmdAccRight = nil
	case "cmd-delete-unix-word":
		ind := strings.LastIndex(strings.TrimRight(string(app.ui.cmdAccLeft), " "), " ") + 1
		app.ui.cmdYankBuf = app.ui.cmdAccLeft[ind:]
		app.ui.cmdAccLeft = app.ui.cmdAccLeft[:ind]
	case "cmd-yank":
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, app.ui.cmdYankBuf...)
	case "cmd-transpose":
		if len(app.ui.cmdAccLeft) < 2 {
			return
		}
		app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-1], app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-2] = app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-2], app.ui.cmdAccLeft[len(app.ui.cmdAccLeft)-1]
	case "cmd-interrupt":
		if app.cmd != nil {
			app.cmd.Process.Kill()
		}
		app.ui.menuBuf = nil
		app.ui.cmdAccLeft = nil
		app.ui.cmdAccRight = nil
		app.ui.cmdPrefix = ""
	case "cmd-word":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		loc := reWordEnd.FindStringIndex(string(app.ui.cmdAccRight))
		if loc == nil {
			return
		}
		ind := loc[0] + 1
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, app.ui.cmdAccRight[:ind]...)
		app.ui.cmdAccRight = app.ui.cmdAccRight[ind:]
	case "cmd-word-back":
		if len(app.ui.cmdAccLeft) == 0 {
			return
		}
		locs := reWordBeg.FindAllStringIndex(string(app.ui.cmdAccLeft), -1)
		if locs == nil {
			return
		}
		ind := locs[len(locs)-1][1] - 1
		old := app.ui.cmdAccRight
		app.ui.cmdAccRight = append([]rune{}, app.ui.cmdAccLeft[ind:]...)
		app.ui.cmdAccRight = append(app.ui.cmdAccRight, old...)
		app.ui.cmdAccLeft = app.ui.cmdAccLeft[:ind]
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
		loc := reWordEnd.FindStringIndex(string(app.ui.cmdAccRight))
		if loc == nil {
			return
		}
		ind = loc[0] + 1
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, app.ui.cmdAccRight[:ind]...)
		app.ui.cmdAccRight = app.ui.cmdAccRight[ind:]
	case "cmd-delete-word":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		loc := reWordEnd.FindStringIndex(string(app.ui.cmdAccRight))
		if loc == nil {
			return
		}
		ind := loc[0] + 1
		app.ui.cmdAccRight = app.ui.cmdAccRight[ind:]
	case "cmd-uppercase-word":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		loc := reWordEnd.FindStringIndex(string(app.ui.cmdAccRight))
		if loc == nil {
			return
		}
		ind := loc[0] + 1
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(strings.ToUpper(string(app.ui.cmdAccRight[:ind])))...)
		app.ui.cmdAccRight = app.ui.cmdAccRight[ind:]
	case "cmd-lowercase-word":
		if len(app.ui.cmdAccRight) == 0 {
			return
		}
		loc := reWordEnd.FindStringIndex(string(app.ui.cmdAccRight))
		if loc == nil {
			return
		}
		ind := loc[0] + 1
		app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, []rune(strings.ToLower(string(app.ui.cmdAccRight[:ind])))...)
		app.ui.cmdAccRight = app.ui.cmdAccRight[ind:]
	case "cmd-transpose-word":
		if len(app.ui.cmdAccLeft) == 0 {
			return
		}

		locs := reWord.FindAllStringIndex(string(app.ui.cmdAccLeft), -1)
		if len(locs) < 2 {
			return
		}

		if len(app.ui.cmdAccRight) > 0 {
			loc := reWordEnd.FindStringIndex(string(app.ui.cmdAccRight))
			if loc != nil {
				ind := loc[0] + 1
				app.ui.cmdAccLeft = append(app.ui.cmdAccLeft, app.ui.cmdAccRight[:ind]...)
				app.ui.cmdAccRight = app.ui.cmdAccRight[ind:]
			}
		}

		locs = reWord.FindAllStringIndex(string(app.ui.cmdAccLeft), -1)

		beg1, end1 := locs[len(locs)-2][0], locs[len(locs)-2][1]
		beg2, end2 := locs[len(locs)-1][0], locs[len(locs)-1][1]

		var acc []rune

		acc = append(acc, app.ui.cmdAccLeft[:beg1]...)
		acc = append(acc, app.ui.cmdAccLeft[beg2:end2]...)
		acc = append(acc, app.ui.cmdAccLeft[end1:beg2]...)
		acc = append(acc, app.ui.cmdAccLeft[beg1:end1]...)
		acc = append(acc, app.ui.cmdAccLeft[end2:]...)

		app.ui.cmdAccLeft = acc
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
	for _, expr := range e.exprs {
		expr.eval(app, nil)
	}
}
