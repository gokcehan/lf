package main

import "time"

var gOpts struct {
	dirfirst  bool
	hidden    bool
	preview   bool
	scrolloff int
	tabstop   int
	filesep   string
	ifs       string
	previewer string
	shell     string
	showinfo  string
	sortby    string
	timefmt   string
	ratios    []int
	keys      map[string]expr
	cmds      map[string]expr
}

func init() {
	gOpts.dirfirst = true
	gOpts.hidden = false
	gOpts.preview = true
	gOpts.scrolloff = 0
	gOpts.tabstop = 8
	gOpts.filesep = ":"
	gOpts.shell = envShell
	gOpts.showinfo = "none"
	gOpts.sortby = "natural"
	gOpts.timefmt = time.ANSIC
	gOpts.ratios = []int{1, 2, 3}

	gOpts.keys = make(map[string]expr)

	gOpts.keys["k"] = &callExpr{"up", nil}
	gOpts.keys["<up>"] = &callExpr{"up", nil}
	gOpts.keys["<c-u>"] = &callExpr{"half-up", nil}
	gOpts.keys["<c-b>"] = &callExpr{"page-up", nil}
	gOpts.keys["j"] = &callExpr{"down", nil}
	gOpts.keys["<down>"] = &callExpr{"down", nil}
	gOpts.keys["<c-d>"] = &callExpr{"half-down", nil}
	gOpts.keys["<c-f>"] = &callExpr{"page-down", nil}
	gOpts.keys["h"] = &callExpr{"updir", nil}
	gOpts.keys["<left>"] = &callExpr{"updir", nil}
	gOpts.keys["l"] = &callExpr{"open", nil}
	gOpts.keys["<right>"] = &callExpr{"open", nil}
	gOpts.keys["q"] = &callExpr{"quit", nil}
	gOpts.keys["G"] = &callExpr{"bot", nil}
	gOpts.keys["gg"] = &callExpr{"top", nil}
	gOpts.keys[":"] = &callExpr{"read", nil}
	gOpts.keys["$"] = &callExpr{"read-shell", nil}
	gOpts.keys["!"] = &callExpr{"read-shell-wait", nil}
	gOpts.keys["&"] = &callExpr{"read-shell-async", nil}
	gOpts.keys["/"] = &callExpr{"search", nil}
	gOpts.keys["?"] = &callExpr{"search-back", nil}
	gOpts.keys["n"] = &callExpr{"search-next", nil}
	gOpts.keys["N"] = &callExpr{"search-prev", nil}
	gOpts.keys["<space>"] = &callExpr{"toggle", nil}
	gOpts.keys["v"] = &callExpr{"invert", nil}
	gOpts.keys["y"] = &callExpr{"yank", nil}
	gOpts.keys["d"] = &callExpr{"delete", nil}
	gOpts.keys["c"] = &callExpr{"clear", nil}
	gOpts.keys["p"] = &callExpr{"put", nil}
	gOpts.keys["<c-l>"] = &callExpr{"renew", nil}

	gOpts.cmds = make(map[string]expr)
}
