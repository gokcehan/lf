package main

import "time"

type sortMethod byte

const (
	naturalSort sortMethod = iota
	nameSort
	sizeSort
	timeSort
	atimeSort
	ctimeSort
	extSort
)

type sortOption byte

const (
	dirfirstSort sortOption = 1 << iota
	hiddenSort
	reverseSort
)

type sortType struct {
	method sortMethod
	option sortOption
}

var gOpts struct {
	anchorfind     bool
	color256       bool
	dircounts      bool
	drawbox        bool
	globsearch     bool
	icons          bool
	ignorecase     bool
	ignoredia      bool
	incsearch      bool
	number         bool
	preview        bool
	relativenumber bool
	smartcase      bool
	smartdia       bool
	wrapscan       bool
	wrapscroll     bool
	findlen        int
	period         int
	scrolloff      int
	tabstop        int
	errorfmt       string
	filesep        string
	ifs            string
	previewer      string
	promptfmt      string
	shell          string
	timefmt        string
	ratios         []int
	hiddenfiles    []string
	info           []string
	shellopts      []string
	keys           map[string]expr
	cmdkeys        map[string]expr
	cmds           map[string]expr
	sortType       sortType
}

func init() {
	gOpts.anchorfind = true
	gOpts.color256 = false
	gOpts.dircounts = false
	gOpts.drawbox = false
	gOpts.globsearch = false
	gOpts.icons = false
	gOpts.ignorecase = true
	gOpts.ignoredia = false
	gOpts.incsearch = false
	gOpts.number = false
	gOpts.preview = true
	gOpts.relativenumber = false
	gOpts.smartcase = true
	gOpts.smartdia = false
	gOpts.wrapscan = true
	gOpts.wrapscroll = false
	gOpts.findlen = 1
	gOpts.period = 0
	gOpts.scrolloff = 0
	gOpts.tabstop = 8
	gOpts.errorfmt = "\033[7;31;47m%s\033[0m"
	gOpts.filesep = "\n"
	gOpts.ifs = ""
	gOpts.previewer = ""
	gOpts.promptfmt = "\033[32;1m%u@%h\033[0m:\033[34;1m%w/\033[0m\033[1m%f\033[0m"
	gOpts.shell = gDefaultShell
	gOpts.timefmt = time.ANSIC
	gOpts.ratios = []int{1, 2, 3}
	gOpts.hiddenfiles = []string{".*"}
	gOpts.info = nil
	gOpts.shellopts = nil
	gOpts.sortType = sortType{naturalSort, dirfirstSort}

	gOpts.keys = make(map[string]expr)

	gOpts.keys["k"] = &callExpr{"up", nil, 1}
	gOpts.keys["<up>"] = &callExpr{"up", nil, 1}
	gOpts.keys["<c-u>"] = &callExpr{"half-up", nil, 1}
	gOpts.keys["<c-b>"] = &callExpr{"page-up", nil, 1}
	gOpts.keys["<pgup>"] = &callExpr{"page-up", nil, 1}
	gOpts.keys["j"] = &callExpr{"down", nil, 1}
	gOpts.keys["<down>"] = &callExpr{"down", nil, 1}
	gOpts.keys["<c-d>"] = &callExpr{"half-down", nil, 1}
	gOpts.keys["<c-f>"] = &callExpr{"page-down", nil, 1}
	gOpts.keys["<pgdn>"] = &callExpr{"page-down", nil, 1}
	gOpts.keys["h"] = &callExpr{"updir", nil, 1}
	gOpts.keys["<left>"] = &callExpr{"updir", nil, 1}
	gOpts.keys["l"] = &callExpr{"open", nil, 1}
	gOpts.keys["<right>"] = &callExpr{"open", nil, 1}
	gOpts.keys["q"] = &callExpr{"quit", nil, 1}
	gOpts.keys["gg"] = &callExpr{"top", nil, 1}
	gOpts.keys["<home>"] = &callExpr{"top", nil, 1}
	gOpts.keys["G"] = &callExpr{"bottom", nil, 1}
	gOpts.keys["<end>"] = &callExpr{"bottom", nil, 1}
	gOpts.keys["<space>"] = &listExpr{[]expr{&callExpr{"toggle", nil, 1}, &callExpr{"down", nil, 1}}}
	gOpts.keys["v"] = &callExpr{"invert", nil, 1}
	gOpts.keys["u"] = &callExpr{"unselect", nil, 1}
	gOpts.keys["y"] = &callExpr{"copy", nil, 1}
	gOpts.keys["d"] = &callExpr{"cut", nil, 1}
	gOpts.keys["c"] = &callExpr{"clear", nil, 1}
	gOpts.keys["p"] = &callExpr{"paste", nil, 1}
	gOpts.keys["<c-l>"] = &callExpr{"redraw", nil, 1}
	gOpts.keys["<c-r>"] = &callExpr{"reload", nil, 1}
	gOpts.keys[":"] = &callExpr{"read", nil, 1}
	gOpts.keys["$"] = &callExpr{"shell", nil, 1}
	gOpts.keys["%"] = &callExpr{"shell-pipe", nil, 1}
	gOpts.keys["!"] = &callExpr{"shell-wait", nil, 1}
	gOpts.keys["&"] = &callExpr{"shell-async", nil, 1}
	gOpts.keys["f"] = &callExpr{"find", nil, 1}
	gOpts.keys["F"] = &callExpr{"find-back", nil, 1}
	gOpts.keys[";"] = &callExpr{"find-next", nil, 1}
	gOpts.keys[","] = &callExpr{"find-prev", nil, 1}
	gOpts.keys["/"] = &callExpr{"search", nil, 1}
	gOpts.keys["?"] = &callExpr{"search-back", nil, 1}
	gOpts.keys["n"] = &callExpr{"search-next", nil, 1}
	gOpts.keys["N"] = &callExpr{"search-prev", nil, 1}
	gOpts.keys["m"] = &callExpr{"mark-save", nil, 1}
	gOpts.keys["'"] = &callExpr{"mark-load", nil, 1}
	gOpts.keys[`"`] = &callExpr{"mark-remove", nil, 1}
	gOpts.keys[`r`] = &callExpr{"rename", nil, 1}
	gOpts.keys["<c-n>"] = &callExpr{"cmd-history-next", nil, 1}
	gOpts.keys["<c-p>"] = &callExpr{"cmd-history-prev", nil, 1}

	gOpts.keys["zh"] = &setExpr{"hidden!", ""}
	gOpts.keys["zr"] = &setExpr{"reverse!", ""}
	gOpts.keys["zn"] = &setExpr{"info", ""}
	gOpts.keys["zs"] = &setExpr{"info", "size"}
	gOpts.keys["zt"] = &setExpr{"info", "time"}
	gOpts.keys["za"] = &setExpr{"info", "size:time"}
	gOpts.keys["sn"] = &listExpr{[]expr{&setExpr{"sortby", "natural"}, &setExpr{"info", ""}}}
	gOpts.keys["ss"] = &listExpr{[]expr{&setExpr{"sortby", "size"}, &setExpr{"info", "size"}}}
	gOpts.keys["st"] = &listExpr{[]expr{&setExpr{"sortby", "time"}, &setExpr{"info", "time"}}}
	gOpts.keys["sa"] = &listExpr{[]expr{&setExpr{"sortby", "atime"}, &setExpr{"info", "atime"}}}
	gOpts.keys["sc"] = &listExpr{[]expr{&setExpr{"sortby", "ctime"}, &setExpr{"info", "ctime"}}}
	gOpts.keys["se"] = &listExpr{[]expr{&setExpr{"sortby", "ext"}, &setExpr{"info", ""}}}
	gOpts.keys["gh"] = &callExpr{"cd", []string{"~"}, 1}

	gOpts.cmdkeys = make(map[string]expr)

	gOpts.cmdkeys["<space>"] = &callExpr{"cmd-insert", []string{" "}, 1}
	gOpts.cmdkeys["<esc>"] = &callExpr{"cmd-escape", nil, 1}
	gOpts.cmdkeys["<tab>"] = &callExpr{"cmd-complete", nil, 1}
	gOpts.cmdkeys["<enter>"] = &callExpr{"cmd-enter", nil, 1}
	gOpts.cmdkeys["<c-j>"] = &callExpr{"cmd-enter", nil, 1}
	gOpts.cmdkeys["<c-n>"] = &callExpr{"cmd-history-next", nil, 1}
	gOpts.cmdkeys["<c-p>"] = &callExpr{"cmd-history-prev", nil, 1}
	gOpts.cmdkeys["<delete>"] = &callExpr{"cmd-delete", nil, 1}
	gOpts.cmdkeys["<c-d>"] = &callExpr{"cmd-delete", nil, 1}
	gOpts.cmdkeys["<bs>"] = &callExpr{"cmd-delete-back", nil, 1}
	gOpts.cmdkeys["<bs2>"] = &callExpr{"cmd-delete-back", nil, 1}
	gOpts.cmdkeys["<left>"] = &callExpr{"cmd-left", nil, 1}
	gOpts.cmdkeys["<c-b>"] = &callExpr{"cmd-left", nil, 1}
	gOpts.cmdkeys["<right>"] = &callExpr{"cmd-right", nil, 1}
	gOpts.cmdkeys["<c-f>"] = &callExpr{"cmd-right", nil, 1}
	gOpts.cmdkeys["<home>"] = &callExpr{"cmd-home", nil, 1}
	gOpts.cmdkeys["<c-a>"] = &callExpr{"cmd-home", nil, 1}
	gOpts.cmdkeys["<end>"] = &callExpr{"cmd-end", nil, 1}
	gOpts.cmdkeys["<c-e>"] = &callExpr{"cmd-end", nil, 1}
	gOpts.cmdkeys["<c-u>"] = &callExpr{"cmd-delete-home", nil, 1}
	gOpts.cmdkeys["<c-k>"] = &callExpr{"cmd-delete-end", nil, 1}
	gOpts.cmdkeys["<c-w>"] = &callExpr{"cmd-delete-unix-word", nil, 1}
	gOpts.cmdkeys["<c-y>"] = &callExpr{"cmd-yank", nil, 1}
	gOpts.cmdkeys["<c-t>"] = &callExpr{"cmd-transpose", nil, 1}
	gOpts.cmdkeys["<c-c>"] = &callExpr{"cmd-interrupt", nil, 1}
	gOpts.cmdkeys["<a-f>"] = &callExpr{"cmd-word", nil, 1}
	gOpts.cmdkeys["<a-b>"] = &callExpr{"cmd-word-back", nil, 1}
	gOpts.cmdkeys["<a-c>"] = &callExpr{"cmd-capitalize-word", nil, 1}
	gOpts.cmdkeys["<a-d>"] = &callExpr{"cmd-delete-word", nil, 1}
	gOpts.cmdkeys["<a-u>"] = &callExpr{"cmd-uppercase-word", nil, 1}
	gOpts.cmdkeys["<a-l>"] = &callExpr{"cmd-lowercase-word", nil, 1}
	gOpts.cmdkeys["<a-t>"] = &callExpr{"cmd-transpose-word", nil, 1}

	gOpts.cmds = make(map[string]expr)

	setDefaults()
}
