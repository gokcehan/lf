package main

import "time"

var gOpts struct {
	dircounts  bool
	dirfirst   bool
	globsearch bool
	hidden     bool
	ignorecase bool
	preview    bool
	reverse    bool
	smartcase  bool
	wrapscan   bool
	scrolloff  int
	tabstop    int
	filesep    string
	ifs        string
	previewer  string
	shell      string
	sortby     string
	timefmt    string
	ratios     []int
	info       []string
	keys       map[string]expr
	cmdkeys    map[string]expr
	cmds       map[string]expr
}

func init() {
	gOpts.dircounts = false
	gOpts.dirfirst = true
	gOpts.globsearch = false
	gOpts.hidden = false
	gOpts.ignorecase = false
	gOpts.preview = true
	gOpts.reverse = false
	gOpts.smartcase = false
	gOpts.wrapscan = true
	gOpts.scrolloff = 0
	gOpts.tabstop = 8
	gOpts.filesep = ":"
	gOpts.shell = gDefaultShell
	gOpts.sortby = "natural"
	gOpts.timefmt = time.ANSIC
	gOpts.ratios = []int{1, 2, 3}
	gOpts.info = nil

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
	gOpts.keys["<c-n>"] = &callExpr{"cmd-hist-next", nil}
	gOpts.keys["<c-p>"] = &callExpr{"cmd-hist-prev", nil}
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

	gOpts.cmdkeys = make(map[string]expr)

	// TODO: rest of the keys
	gOpts.cmdkeys["<esc>"] = &callExpr{"cmd-escape", nil}
	gOpts.cmdkeys["<space>"] = &callExpr{"cmd-insert", []string{" "}}
	gOpts.cmdkeys["<tab>"] = &callExpr{"cmd-comp", nil}
	gOpts.cmdkeys["<enter>"] = &callExpr{"cmd-enter", nil}
	gOpts.cmdkeys["<c-j>"] = &callExpr{"cmd-enter", nil}
	gOpts.cmdkeys["<c-n>"] = &callExpr{"cmd-hist-next", nil}
	gOpts.cmdkeys["<c-p>"] = &callExpr{"cmd-hist-prev", nil}
	gOpts.cmdkeys["<bs>"] = &callExpr{"cmd-delete-back", nil}
	gOpts.cmdkeys["<bs2>"] = &callExpr{"cmd-delete-back", nil}
	gOpts.cmdkeys["<delete>"] = &callExpr{"cmd-delete", nil}
	gOpts.cmdkeys["<c-d>"] = &callExpr{"cmd-delete", nil}
	gOpts.cmdkeys["<left>"] = &callExpr{"cmd-left", nil}
	gOpts.cmdkeys["<c-b>"] = &callExpr{"cmd-left", nil}
	gOpts.cmdkeys["<right>"] = &callExpr{"cmd-right", nil}
	gOpts.cmdkeys["<c-f>"] = &callExpr{"cmd-right", nil}
	gOpts.cmdkeys["<home>"] = &callExpr{"cmd-beg", nil}
	gOpts.cmdkeys["<c-a>"] = &callExpr{"cmd-beg", nil}
	gOpts.cmdkeys["<end>"] = &callExpr{"cmd-end", nil}
	gOpts.cmdkeys["<c-e>"] = &callExpr{"cmd-end", nil}
	gOpts.cmdkeys["<c-k>"] = &callExpr{"cmd-delete-end", nil}
	gOpts.cmdkeys["<c-u>"] = &callExpr{"cmd-delete-beg", nil}
	gOpts.cmdkeys["<c-w>"] = &callExpr{"cmd-delete-word", nil}
	gOpts.cmdkeys["<c-y>"] = &callExpr{"cmd-put", nil}
	gOpts.cmdkeys["<c-t>"] = &callExpr{"cmd-transpose", nil}

	gOpts.cmds = make(map[string]expr)
}
