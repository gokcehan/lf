package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (e *SetExpr) eval(app *App, args []string) {
	switch e.opt {
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
		if e.val != "name" && e.val != "size" && e.val != "time" {
			msg := "sortby should either be 'name', 'size' or 'time'"
			app.ui.message = msg
			log.Print(msg)
			return
		}
		gOpts.sortby = e.val
		app.nav.renew(app.nav.height)
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
		app.ui = newUI()
	default:
		msg := fmt.Sprintf("unknown option: %s", e.opt)
		app.ui.message = msg
		log.Print(msg)
	}
}

func (e *MapExpr) eval(app *App, args []string) {
	if e.expr == nil {
		delete(gOpts.keys, e.keys)
		return
	}
	gOpts.keys[e.keys] = e.expr
}

func (e *CmdExpr) eval(app *App, args []string) {
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

func (e *CallExpr) eval(app *App, args []string) {
	// TODO: check for extra toks in each case
	switch e.name {
	case "up":
		app.nav.up(1)
		app.ui.loadFile(app.nav)
	case "half-up":
		app.nav.up(app.nav.height / 2)
		app.ui.loadFile(app.nav)
	case "page-up":
		app.nav.up(app.nav.height)
		app.ui.loadFile(app.nav)
	case "down":
		app.nav.down(1)
		app.ui.loadFile(app.nav)
	case "half-down":
		app.nav.down(app.nav.height / 2)
		app.ui.loadFile(app.nav)
	case "page-down":
		app.nav.down(app.nav.height)
		app.ui.loadFile(app.nav)
	case "updir":
		if err := app.nav.updir(); err != nil {
			app.ui.message = err.Error()
			log.Print(err)
			return
		}
		app.ui.loadFile(app.nav)
	case "open":
		dir := app.nav.currDir()

		if len(dir.fi) == 0 {
			return
		}

		path := app.nav.currPath()

		f, err := os.Stat(path)
		if err != nil {
			msg := fmt.Sprintf("open: %s", err)
			app.ui.message = msg
			log.Print(msg)
			return
		}

		if f.IsDir() {
			if err := app.nav.open(); err != nil {
				app.ui.message = err.Error()
				log.Print(err)
				return
			}
			app.ui.loadFile(app.nav)
			return
		}

		if gSelectionPath != "" {
			out, err := os.Create(gSelectionPath)
			if err != nil {
				log.Printf("opening selection file: %s", err)
			}
			defer out.Close()

			if len(app.nav.marks) != 0 {
				marks := app.nav.currMarks()
				path = strings.Join(marks, "\n")
			}

			_, err = out.WriteString(path)
			if err != nil {
				log.Printf("writing selection file: %s", err)
			}

			gExitFlag = true
			return
		}

		if cmd, ok := gOpts.cmds["open-file"]; ok {
			cmd.eval(app, e.args)
		}
	case "quit":
		gExitFlag = true
	case "bot":
		app.nav.bot()
		app.ui.loadFile(app.nav)
	case "top":
		app.nav.top()
		app.ui.loadFile(app.nav)
	case "read":
		s := app.ui.prompt(app.nav, ":")
		if len(s) == 0 {
			return
		}
		log.Printf("command: %s", s)
		p := newParser(strings.NewReader(s))
		for p.parse() {
			p.expr.eval(app, nil)
		}
		if p.err != nil {
			app.ui.message = p.err.Error()
			log.Print(p.err)
		}
	case "read-shell":
		s := app.ui.prompt(app.nav, "$")
		if len(s) == 0 {
			return
		}
		log.Printf("shell: %s", s)
		app.runShell(s, nil, false, false)
	case "read-shell-wait":
		s := app.ui.prompt(app.nav, "!")
		if len(s) == 0 {
			return
		}
		log.Printf("shell-wait: %s", s)
		app.runShell(s, nil, true, false)
	case "read-shell-async":
		s := app.ui.prompt(app.nav, "&")
		if len(s) == 0 {
			return
		}
		log.Printf("shell-async: %s", s)
		app.runShell(s, nil, false, true)
	case "search":
		s := app.ui.prompt(app.nav, "/")
		if len(s) == 0 {
			return
		}
		log.Printf("search: %s", s)
		app.ui.message = "sorry, search is not implemented yet!"
		// TODO: implement
	case "search-back":
		s := app.ui.prompt(app.nav, "?")
		if len(s) == 0 {
			return
		}
		log.Printf("search-back: %s", s)
		app.ui.message = "sorry, search-back is not implemented yet!"
		// TODO: implement
	case "toggle":
		app.nav.toggle()
	case "yank":
		if err := app.nav.save(true); err != nil {
			msg := fmt.Sprintf("yank: %s", err)
			app.ui.message = msg
			log.Printf(msg)
			return
		}
		if len(app.nav.marks) == 0 {
			app.ui.message = "file saved for copy"
		} else {
			app.ui.message = fmt.Sprintf("%d files saved for copy", len(app.nav.marks))
		}
		app.nav.marks = make(map[string]bool)
	case "delete":
		if err := app.nav.save(false); err != nil {
			msg := fmt.Sprintf("delete: %s", err)
			app.ui.message = msg
			log.Printf(msg)
			return
		}
		if len(app.nav.marks) == 0 {
			app.ui.message = "file saved for move"
		} else {
			app.ui.message = fmt.Sprintf("%d files saved for move", len(app.nav.marks))
		}
		app.nav.marks = make(map[string]bool)
	case "paste":
		if err := app.nav.paste(); err != nil {
			msg := fmt.Sprintf("paste: %s", err)
			app.ui.message = msg
			log.Printf(msg)
			return
		}
		app.nav.renew(app.nav.height)
		app.nav.save(false)
		saveFiles(nil, false)
	case "renew":
		app.ui.sync()
		app.ui.renew()
		app.nav.renew(app.ui.wins[0].h)
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
	case "push":
		if len(e.args) > 0 {
			app.ui.keysbuf = append(app.ui.keysbuf, splitKeys(strings.Join(e.args, ""))...)
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

func (e *ExecExpr) eval(app *App, args []string) {
	switch e.pref {
	case "$":
		log.Printf("shell: %s -- %s", e, args)
		app.ui.clearMsg()
		app.runShell(e.expr, args, false, false)
		app.ui.loadFile(app.nav)
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
		log.Printf("unknown execution prefix: %q", e.pref)
	}
}

func (e *ListExpr) eval(app *App, args []string) {
	for _, expr := range e.exprs {
		expr.eval(app, nil)
	}
}
