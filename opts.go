package main

import "time"

var gOpts struct {
	dircounts  bool
	dirfirst   bool
	globsearch bool
	hidden     bool
	ignorecase bool
	lscolors   bool
	preview    bool
	reverse    bool
	smartcase  bool
	wrapscan   bool
	scrolloff  int
	tabstop    int
	filesep    string
	ifs        string
	previewer  string
	promptfmt  string
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
	gOpts.ignorecase = true
	// gOpts.lscolors = true   this option is initialized in lscolors.go init() depending on whether the corresponding environment variable is set
	gOpts.preview = true
	gOpts.reverse = false
	gOpts.smartcase = true
	gOpts.wrapscan = true
	gOpts.scrolloff = 0
	gOpts.tabstop = 8
	gOpts.filesep = "\n"
	gOpts.promptfmt = "\033[32;1m%u@%h\033[0m:\033[34;1m%w/\033[0m\033[1m%f\033[0m"
	gOpts.shell = gDefaultShell
	gOpts.sortby = "natural"
	gOpts.timefmt = time.ANSIC
	gOpts.ratios = []int{1, 2, 3}
	gOpts.info = nil

	gOpts.keys = make(map[string]expr)

	gOpts.keys["k"] = &callExpr{"up", nil, 1}
	gOpts.keys["<up>"] = &callExpr{"up", nil, 1}
	gOpts.keys["<c-u>"] = &callExpr{"half-up", nil, 1}
	gOpts.keys["<c-b>"] = &callExpr{"page-up", nil, 1}
	gOpts.keys["j"] = &callExpr{"down", nil, 1}
	gOpts.keys["<down>"] = &callExpr{"down", nil, 1}
	gOpts.keys["<c-d>"] = &callExpr{"half-down", nil, 1}
	gOpts.keys["<c-f>"] = &callExpr{"page-down", nil, 1}
	gOpts.keys["h"] = &callExpr{"updir", nil, 1}
	gOpts.keys["<left>"] = &callExpr{"updir", nil, 1}
	gOpts.keys["l"] = &callExpr{"open", nil, 1}
	gOpts.keys["<right>"] = &callExpr{"open", nil, 1}
	gOpts.keys["q"] = &callExpr{"quit", nil, 1}
	gOpts.keys["gg"] = &callExpr{"top", nil, 1}
	gOpts.keys["G"] = &callExpr{"bot", nil, 1}
	gOpts.keys["<space>"] = &callExpr{"toggle", nil, 1}
	gOpts.keys["v"] = &callExpr{"invert", nil, 1}
	gOpts.keys["u"] = &callExpr{"unmark", nil, 1}
	gOpts.keys["y"] = &callExpr{"yank", nil, 1}
	gOpts.keys["d"] = &callExpr{"delete", nil, 1}
	gOpts.keys["c"] = &callExpr{"clear", nil, 1}
	gOpts.keys["p"] = &callExpr{"put", nil, 1}
	gOpts.keys["<c-l>"] = &callExpr{"redraw", nil, 1}
	gOpts.keys["<c-r>"] = &callExpr{"reload", nil, 1}
	gOpts.keys[":"] = &callExpr{"read", nil, 1}
	gOpts.keys["$"] = &callExpr{"shell", nil, 1}
	gOpts.keys["%"] = &callExpr{"shell-pipe", nil, 1}
	gOpts.keys["!"] = &callExpr{"shell-wait", nil, 1}
	gOpts.keys["&"] = &callExpr{"shell-async", nil, 1}
	gOpts.keys["/"] = &callExpr{"search", nil, 1}
	gOpts.keys["?"] = &callExpr{"search-back", nil, 1}
	gOpts.keys["n"] = &callExpr{"search-next", nil, 1}
	gOpts.keys["N"] = &callExpr{"search-prev", nil, 1}
	gOpts.keys["<c-n>"] = &callExpr{"cmd-hist-next", nil, 1}
	gOpts.keys["<c-p>"] = &callExpr{"cmd-hist-prev", nil, 1}

	gOpts.keys["zh"] = &setExpr{"hidden!", ""}
	gOpts.keys["zr"] = &setExpr{"reverse!", ""}
	gOpts.keys["zn"] = &setExpr{"info", ""}
	gOpts.keys["zs"] = &setExpr{"info", "size"}
	gOpts.keys["zt"] = &setExpr{"info", "time"}
	gOpts.keys["za"] = &setExpr{"info", "size:time"}
	gOpts.keys["sn"] = &listExpr{[]expr{&setExpr{"sortby", "natural"}, &setExpr{"info", ""}}}
	gOpts.keys["ss"] = &listExpr{[]expr{&setExpr{"sortby", "size"}, &setExpr{"info", "size"}}}
	gOpts.keys["st"] = &listExpr{[]expr{&setExpr{"sortby", "time"}, &setExpr{"info", "time"}}}
	gOpts.keys["gh"] = &callExpr{"cd", []string{"~"}, 1}

	gOpts.cmdkeys = make(map[string]expr)

	// TODO: rest of the keys
	gOpts.cmdkeys["<space>"] = &callExpr{"cmd-insert", []string{" "}, 1}
	gOpts.cmdkeys["<esc>"] = &callExpr{"cmd-escape", nil, 1}
	gOpts.cmdkeys["<tab>"] = &callExpr{"cmd-comp", nil, 1}
	gOpts.cmdkeys["<enter>"] = &callExpr{"cmd-enter", nil, 1}
	gOpts.cmdkeys["<c-j>"] = &callExpr{"cmd-enter", nil, 1}
	gOpts.cmdkeys["<c-n>"] = &callExpr{"cmd-hist-next", nil, 1}
	gOpts.cmdkeys["<c-p>"] = &callExpr{"cmd-hist-prev", nil, 1}
	gOpts.cmdkeys["<delete>"] = &callExpr{"cmd-delete", nil, 1}
	gOpts.cmdkeys["<c-d>"] = &callExpr{"cmd-delete", nil, 1}
	gOpts.cmdkeys["<bs>"] = &callExpr{"cmd-delete-back", nil, 1}
	gOpts.cmdkeys["<bs2>"] = &callExpr{"cmd-delete-back", nil, 1}
	gOpts.cmdkeys["<left>"] = &callExpr{"cmd-left", nil, 1}
	gOpts.cmdkeys["<c-b>"] = &callExpr{"cmd-left", nil, 1}
	gOpts.cmdkeys["<right>"] = &callExpr{"cmd-right", nil, 1}
	gOpts.cmdkeys["<c-f>"] = &callExpr{"cmd-right", nil, 1}
	gOpts.cmdkeys["<home>"] = &callExpr{"cmd-beg", nil, 1}
	gOpts.cmdkeys["<c-a>"] = &callExpr{"cmd-beg", nil, 1}
	gOpts.cmdkeys["<end>"] = &callExpr{"cmd-end", nil, 1}
	gOpts.cmdkeys["<c-e>"] = &callExpr{"cmd-end", nil, 1}
	gOpts.cmdkeys["<c-u>"] = &callExpr{"cmd-delete-beg", nil, 1}
	gOpts.cmdkeys["<c-k>"] = &callExpr{"cmd-delete-end", nil, 1}
	gOpts.cmdkeys["<c-w>"] = &callExpr{"cmd-delete-word", nil, 1}
	gOpts.cmdkeys["<c-y>"] = &callExpr{"cmd-put", nil, 1}
	gOpts.cmdkeys["<c-t>"] = &callExpr{"cmd-transpose", nil, 1}
	gOpts.cmdkeys["<c-c>"] = &callExpr{"cmd-interrupt", nil, 1}

	gOpts.cmds = make(map[string]expr)

	setDefaults()
}
