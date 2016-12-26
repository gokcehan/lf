package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (e *setExpr) eval(app *app, args []string) {
	switch e.opt {
	case "dirfirst":
		gOpts.dirfirst = true
		app.nav.renew(app.nav.height)
	case "nodirfirst":
		gOpts.dirfirst = false
		app.nav.renew(app.nav.height)
	case "dirfirst!":
		gOpts.dirfirst = !gOpts.dirfirst
		app.nav.renew(app.nav.height)
	case "hidden":
		gOpts.hidden = true
		app.nav.renew(app.nav.height)
	case "nohidden":
		gOpts.hidden = false
		app.nav.renew(app.nav.height)
	case "hidden!":
		gOpts.hidden = !gOpts.hidden
		app.nav.renew(app.nav.height)
	case "preview":
		gOpts.preview = true
	case "nopreview":
		gOpts.preview = false
	case "preview!":
		gOpts.preview = !gOpts.preview
	case "scrolloff":
		n, err := strconv.Atoi(e.val)
		if err != nil {
			msg := fmt.Sprintf("scrolloff: %s", err)
			app.ui.message = msg
			log.Print(msg)
			return
		}
		if n < 0 {
			msg := "scrolloff: value should be a non-negative number"
			app.ui.message = msg
			log.Print(msg)
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
			msg := fmt.Sprintf("tabstop: %s", err)
			app.ui.message = msg
			log.Print(msg)
			return
		}
		if n <= 0 {
			msg := "tabstop: value should be a positive number"
			app.ui.message = msg
			log.Print(msg)
			return
		}
		gOpts.tabstop = n
	case "filesep":
		gOpts.filesep = e.val
	case "ifs":
		gOpts.ifs = e.val
	case "previewer":
		gOpts.previewer = strings.Replace(e.val, "~", envHome, -1)
	case "shell":
		gOpts.shell = e.val
	case "showinfo":
		if e.val != "none" && e.val != "size" && e.val != "time" {
			msg := "showinfo should either be 'none', 'size' or 'time'"
			app.ui.message = msg
			log.Print(msg)
			return
		}
		gOpts.showinfo = e.val
	case "sortby":
		if e.val != "natural" && e.val != "name" && e.val != "size" && e.val != "time" {
			msg := "sortby should either be 'natural', 'name', 'size' or 'time'"
			app.ui.message = msg
			log.Print(msg)
			return
		}
		gOpts.sortby = e.val
		app.nav.renew(app.nav.height)
	case "timefmt":
		gOpts.timefmt = e.val
	case "ratios":
		toks := strings.Split(e.val, ":")
		var rats []int
		for _, s := range toks {
			i, err := strconv.Atoi(s)
			if err != nil {
				msg := fmt.Sprintf("ratios: %s", err)
				app.ui.message = msg
				log.Print(msg)
				return
			}
			rats = append(rats, i)
		}
		gOpts.ratios = rats
		app.ui.wins = getWins()
		app.ui.loadFile(app.nav)
	default:
		msg := fmt.Sprintf("unknown option: %s", e.opt)
		app.ui.message = msg
		log.Print(msg)
	}
}

func (e *mapExpr) eval(app *app, args []string) {
	if e.expr == nil {
		delete(gOpts.keys, e.keys)
		return
	}
	gOpts.keys[e.keys] = e.expr
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
	// TODO: check for extra toks in each case
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
			app.ui.message = err.Error()
			log.Print(err)
			return
		}
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "open":
		curr, err := app.nav.currFile()
		if err != nil {
			msg := fmt.Sprintf("opening: %s", err)
			app.ui.message = msg
			log.Print(msg)
			return
		}

		if curr.IsDir() {
			app.nav.open()
			if err != nil {
				msg := fmt.Sprintf("opening directory: %s", err)
				app.ui.message = msg
				log.Print(msg)
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
				path = curr.Path
			} else {
				return
			}

			_, err = out.WriteString(path)
			if err != nil {
				log.Printf("writing selection file: %s", err)
			}

			app.quit <- true

			return
		}

		if cmd, ok := gOpts.cmds["open-file"]; ok {
			cmd.eval(app, e.args)
		}
	case "quit":
		app.quit <- true
	case "bot":
		app.nav.bot()
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "top":
		app.nav.top()
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "read":
		app.ui.cmdpref = ":"
	case "read-shell":
		app.ui.cmdpref = "$"
	case "read-shell-wait":
		app.ui.cmdpref = "!"
	case "read-shell-async":
		app.ui.cmdpref = "&"
	case "search":
		app.ui.cmdpref = "/"
	case "search-back":
		app.ui.cmdpref = "?"
	case "search-next":
		app.nav.searchNext()
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "search-prev":
		app.nav.searchPrev()
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "toggle":
		app.nav.toggle()
	case "invert":
		app.nav.invert()
	case "yank":
		if err := app.nav.save(true); err != nil {
			msg := fmt.Sprintf("yank: %s", err)
			app.ui.message = msg
			log.Printf(msg)
			return
		}
		app.nav.marks = make(map[string]bool)
		if err := sendRemote("send sync"); err != nil {
			msg := fmt.Sprintf("yank: %s", err)
			app.ui.message = msg
			log.Printf(msg)
		}
	case "delete":
		if err := app.nav.save(false); err != nil {
			msg := fmt.Sprintf("delete: %s", err)
			app.ui.message = msg
			log.Printf(msg)
			return
		}
		app.nav.marks = make(map[string]bool)
		if err := sendRemote("send sync"); err != nil {
			msg := fmt.Sprintf("delete: %s", err)
			app.ui.message = msg
			log.Printf(msg)
		}
	case "put":
		if err := app.nav.put(); err != nil {
			msg := fmt.Sprintf("put: %s", err)
			app.ui.message = msg
			log.Printf(msg)
			return
		}
		app.nav.renew(app.nav.height)
		app.nav.save(false)
		app.nav.saves = make(map[string]bool)
		saveFiles(nil, false)
		if err := sendRemote("send sync"); err != nil {
			msg := fmt.Sprintf("put: %s", err)
			app.ui.message = msg
			log.Printf(msg)
		}
	case "clear":
		if err := saveFiles(nil, false); err != nil {
			msg := fmt.Sprintf("clear: %s", err)
			app.ui.message = msg
			log.Printf(msg)
			return
		}
		if err := sendRemote("send sync"); err != nil {
			msg := fmt.Sprintf("clear: %s", err)
			app.ui.message = msg
			log.Printf(msg)
		}
	case "renew":
		app.ui.sync()
		app.ui.renew()
		app.nav.renew(app.ui.wins[0].h)
	case "sync":
		if err := app.nav.sync(); err != nil {
			msg := fmt.Sprintf("sync: %s", err)
			app.ui.message = msg
			log.Printf(msg)
		}
	case "echo":
		app.ui.message = strings.Join(e.args, " ")
	case "cd":
		path := "~"
		if len(e.args) > 0 {
			path = e.args[0]
		}
		if err := app.nav.cd(path); err != nil {
			app.ui.message = err.Error()
			log.Print(err)
			return
		}
		app.ui.loadFile(app.nav)
		app.ui.loadFileInfo(app.nav)
	case "push":
		if len(e.args) > 0 {
			log.Println("pushing keys", e.args[0])
			for _, key := range splitKeys(e.args[0]) {
				app.ui.keychan <- key
			}
		}
	case "cmd-insert":
		if len(e.args) > 0 {
			app.ui.cmdlacc = append(app.ui.cmdlacc, []rune(e.args[0])...)
		}
	case "cmd-escape":
		app.ui.menubuf = nil
		app.ui.cmdbuf = nil
		app.ui.cmdlacc = nil
		app.ui.cmdracc = nil
		app.ui.cmdpref = ""
	case "cmd-comp":
		var matches []string
		if app.ui.cmdpref == ":" {
			matches, app.ui.cmdlacc = compCmd(app.ui.cmdlacc)
		} else {
			matches, app.ui.cmdlacc = compShell(app.ui.cmdlacc)
		}
		app.ui.draw(app.nav)
		if len(matches) > 1 {
			app.ui.menubuf = listMatches(matches)
		} else {
			app.ui.menubuf = nil
		}
	case "cmd-enter":
		s := string(append(app.ui.cmdlacc, app.ui.cmdracc...))
		if len(s) == 0 {
			return
		}
		app.ui.menubuf = nil
		app.ui.cmdbuf = nil
		app.ui.cmdlacc = nil
		app.ui.cmdracc = nil
		switch app.ui.cmdpref {
		case ":":
			log.Printf("command: %s", s)
			p := newParser(strings.NewReader(s))
			for p.parse() {
				p.expr.eval(app, nil)
			}
			if p.err != nil {
				app.ui.message = p.err.Error()
				log.Print(p.err)
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
			app.nav.searchNext()
			app.ui.loadFile(app.nav)
			app.ui.loadFileInfo(app.nav)
		case "?":
			log.Printf("search-back: %s", s)
			app.nav.search = s
			app.nav.searchPrev()
			app.ui.loadFile(app.nav)
			app.ui.loadFileInfo(app.nav)
		default:
			log.Printf("entering unknown execution prefix: %q", app.ui.cmdpref)
		}
		app.ui.cmdpref = ""
	case "cmd-delete-back":
		if len(app.ui.cmdlacc) > 0 {
			app.ui.cmdlacc = app.ui.cmdlacc[:len(app.ui.cmdlacc)-1]
		}
	case "cmd-delete":
		if len(app.ui.cmdracc) > 0 {
			app.ui.cmdracc = app.ui.cmdracc[1:]
		}
	case "cmd-left":
		if len(app.ui.cmdlacc) > 0 {
			app.ui.cmdracc = append([]rune{app.ui.cmdlacc[len(app.ui.cmdlacc)-1]}, app.ui.cmdracc...)
			app.ui.cmdlacc = app.ui.cmdlacc[:len(app.ui.cmdlacc)-1]
		}
	case "cmd-right":
		if len(app.ui.cmdracc) > 0 {
			app.ui.cmdlacc = append(app.ui.cmdlacc, app.ui.cmdracc[0])
			app.ui.cmdracc = app.ui.cmdracc[1:]
		}
	case "cmd-beg":
		app.ui.cmdracc = append(app.ui.cmdlacc, app.ui.cmdracc...)
		app.ui.cmdlacc = nil
	case "cmd-end":
		app.ui.cmdlacc = append(app.ui.cmdlacc, app.ui.cmdracc...)
		app.ui.cmdracc = nil
	case "cmd-delete-end":
		if len(app.ui.cmdracc) > 0 {
			app.ui.cmdbuf = app.ui.cmdracc
			app.ui.cmdracc = nil
		}
	case "cmd-delete-beg":
		if len(app.ui.cmdlacc) > 0 {
			app.ui.cmdbuf = app.ui.cmdlacc
			app.ui.cmdlacc = nil
		}
	case "cmd-delete-word":
		ind := strings.LastIndex(strings.TrimRight(string(app.ui.cmdlacc), " "), " ") + 1
		app.ui.cmdbuf = app.ui.cmdlacc[ind:]
		app.ui.cmdlacc = app.ui.cmdlacc[:ind]
	case "cmd-put":
		app.ui.cmdlacc = append(app.ui.cmdlacc, app.ui.cmdbuf...)
	case "cmd-transpose":
		if len(app.ui.cmdlacc) > 1 {
			app.ui.cmdlacc[len(app.ui.cmdlacc)-1], app.ui.cmdlacc[len(app.ui.cmdlacc)-2] = app.ui.cmdlacc[len(app.ui.cmdlacc)-2], app.ui.cmdlacc[len(app.ui.cmdlacc)-1]
		}
	default:
		cmd, ok := gOpts.cmds[e.name]
		if !ok {
			msg := fmt.Sprintf("command not found: %s", e.name)
			app.ui.message = msg
			log.Print(msg)
			return
		}
		cmd.eval(app, e.args)
	}
}

func (e *execExpr) eval(app *app, args []string) {
	switch e.pref {
	case "$":
		log.Printf("shell: %s -- %s", e, args)
		app.runShell(e.expr, args, false, false)
	case "!":
		log.Printf("shell-wait: %s -- %s", e, args)
		app.runShell(e.expr, args, true, false)
	case "&":
		log.Printf("shell-async: %s -- %s", e, args)
		app.runShell(e.expr, args, false, true)
	case "/":
		log.Printf("search: %s -- %s", e, args)
		// TODO: implement
	case "?":
		log.Printf("search-back: %s -- %s", e, args)
		// TODO: implement
	default:
		log.Printf("evaluating unknown execution prefix: %q", e.pref)
	}
}

func (e *listExpr) eval(app *app, args []string) {
	for _, expr := range e.exprs {
		expr.eval(app, nil)
	}
}
