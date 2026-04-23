package main

import (
	"fmt"
	"maps"
	"time"
)

// String values match the sortby string sent by the user at startup
type sortMethod string

const (
	naturalSort sortMethod = "natural"
	nameSort    sortMethod = "name"
	sizeSort    sortMethod = "size"
	timeSort    sortMethod = "time"
	atimeSort   sortMethod = "atime"
	btimeSort   sortMethod = "btime"
	ctimeSort   sortMethod = "ctime"
	extSort     sortMethod = "ext"
	customSort  sortMethod = "custom"
)

func isValidSortMethod(method sortMethod) bool {
	switch method {
	case naturalSort, nameSort, sizeSort, timeSort, atimeSort, btimeSort, ctimeSort, extSort, customSort:
		return true
	}
	return false
}

const invalidSortErrorMessage = `sortby: value should either be 'natural', 'name', 'size', 'time', 'atime', 'btime', 'ctime', 'ext' or 'custom'`

type searchMethod string

const (
	textSearch  searchMethod = "text"
	globSearch  searchMethod = "glob"
	regexSearch searchMethod = "regex"
)

type cursorStyle string

const (
	defaultCursor        cursorStyle = "default"
	blockCursor          cursorStyle = "block"
	underlineCursor      cursorStyle = "underline"
	barCursor            cursorStyle = "bar"
	blinkBlockCursor     cursorStyle = "blinkblock"
	blinkUnderlineCursor cursorStyle = "blinkunderline"
	blinkBarCursor       cursorStyle = "blinkbar"
)

type borderStyle byte

const (
	borderOutline borderStyle = 1 << iota
	borderSeparators
	borderRound

	borderBox          = borderOutline | borderSeparators
	borderRoundOutline = borderOutline | borderRound
	borderRoundBox     = borderBox | borderRound
)

func (s borderStyle) String() string {
	switch s {
	case borderBox:
		return "box"
	case borderRoundBox:
		return "roundbox"
	case borderOutline:
		return "outline"
	case borderRoundOutline:
		return "roundoutline"
	case borderSeparators:
		return "separators"
	default:
		return fmt.Sprintf("borderStyle(%d)", s)
	}
}

var gOpts struct {
	anchorfind       bool
	autoquit         bool
	borderfmt        string
	borderstyle      borderStyle
	cleaner          string
	copyfmt          string
	cursoractivefmt  string
	cursorparentfmt  string
	cursorpreviewfmt string
	cutfmt           string
	dircounts        bool
	dirfirst         bool
	dironly          bool
	dirpreviews      bool
	drawbox          bool
	dupfilefmt       string
	errorfmt         string
	filesep          string
	filtermethod     searchMethod
	findlen          int
	hidden           bool
	hiddenfiles      []string
	history          bool
	icons            bool
	ifs              string
	ignorecase       bool
	ignoredia        bool
	incfilter        bool
	incsearch        bool
	info             []string
	infotimefmtnew   string
	infotimefmtold   string
	menufmt          string
	menuheaderfmt    string
	menuselectfmt    string
	mergeindicators  bool
	mouse            bool
	number           bool
	numbercursorfmt  string
	numberfmt        string
	period           int
	preload          bool
	preserve         []string
	preview          bool
	previewer        string
	promptfmt        string
	ratios           []int
	relativenumber   bool
	reverse          bool
	rulerfile        string
	rulerfmt         string
	scrolloff        int
	searchmethod     searchMethod
	selectfmt        string
	selmode          string
	shell            string
	shellflag        string
	shellopts        []string
	showbinds        bool
	sizeunits        string
	smartcase        bool
	smartdia         bool
	sortby           sortMethod
	sortignorecase   bool
	sortignoredia    bool
	statfmt          string
	tabstop          int
	tagfmt           string
	tempmarks        string
	terminalcursor   cursorStyle
	timefmt          string
	truncatechar     string
	truncatepct      int
	visualfmt        string
	waitmsg          string
	watch            bool
	wrapscan         bool
	wrapscroll       bool
	nkeys            map[string]expr
	vkeys            map[string]expr
	cmdkeys          map[string]expr
	cmds             map[string]expr
	user             map[string]string
}

var gLocalOpts struct {
	dircounts      map[string]bool
	dirfirst       map[string]bool
	dironly        map[string]bool
	hidden         map[string]bool
	info           map[string][]string
	reverse        map[string]bool
	sortby         map[string]sortMethod
	sortignorecase map[string]bool
	sortignoredia  map[string]bool
}

func getDirCounts(path string) bool {
	if val, ok := gLocalOpts.dircounts[path]; ok {
		return val
	}
	return gOpts.dircounts
}

func getDirFirst(path string) bool {
	if val, ok := gLocalOpts.dirfirst[path]; ok {
		return val
	}
	return gOpts.dirfirst
}

func getDirOnly(path string) bool {
	if val, ok := gLocalOpts.dironly[path]; ok {
		return val
	}
	return gOpts.dironly
}

func getHidden(path string) bool {
	if val, ok := gLocalOpts.hidden[path]; ok {
		return val
	}
	return gOpts.hidden
}

func getInfo(path string) []string {
	if val, ok := gLocalOpts.info[path]; ok {
		return val
	}
	return gOpts.info
}

func getReverse(path string) bool {
	if val, ok := gLocalOpts.reverse[path]; ok {
		return val
	}
	return gOpts.reverse
}

func getSortBy(path string) sortMethod {
	if val, ok := gLocalOpts.sortby[path]; ok {
		return val
	}
	return gOpts.sortby
}

func getSortIgnoreCase(path string) bool {
	if val, ok := gLocalOpts.sortignorecase[path]; ok {
		return val
	}
	return gOpts.sortignorecase
}

func getSortIgnoreDia(path string) bool {
	if val, ok := gLocalOpts.sortignoredia[path]; ok {
		return val
	}
	return gOpts.sortignoredia
}

func init() {
	gOpts.anchorfind = true
	gOpts.autoquit = true
	gOpts.borderfmt = "\033[0m"
	gOpts.borderstyle = borderBox
	gOpts.cleaner = ""
	gOpts.copyfmt = "\033[7;33m"
	gOpts.cursoractivefmt = "\033[7m"
	gOpts.cursorparentfmt = "\033[7m"
	gOpts.cursorpreviewfmt = "\033[4m"
	gOpts.cutfmt = "\033[7;31m"
	gOpts.dircounts = false
	gOpts.dirfirst = true
	gOpts.dironly = false
	gOpts.dirpreviews = false
	gOpts.drawbox = false
	gOpts.dupfilefmt = "%f.~%n~"
	gOpts.errorfmt = "\033[7;31;47m"
	gOpts.filesep = "\n"
	gOpts.filtermethod = textSearch
	gOpts.findlen = 1
	gOpts.hidden = false
	gOpts.hiddenfiles = gDefaultHiddenFiles
	gOpts.history = true
	gOpts.icons = false
	gOpts.ifs = ""
	gOpts.ignorecase = true
	gOpts.ignoredia = true
	gOpts.incfilter = false
	gOpts.incsearch = false
	gOpts.info = nil
	gOpts.infotimefmtnew = "Jan _2 15:04"
	gOpts.infotimefmtold = "Jan _2  2006"
	gOpts.menufmt = "\033[0m"
	gOpts.menuheaderfmt = "\033[1m"
	gOpts.menuselectfmt = "\033[7m"
	gOpts.mergeindicators = false
	gOpts.mouse = false
	gOpts.number = false
	gOpts.numbercursorfmt = ""
	gOpts.numberfmt = "\033[33m"
	gOpts.period = 0
	gOpts.preload = false
	gOpts.preserve = []string{"mode"}
	gOpts.preview = true
	gOpts.previewer = ""
	gOpts.promptfmt = "\033[32;1m%u@%h\033[0m:\033[34;1m%d\033[0m\033[1m%f\033[0m"
	gOpts.ratios = []int{1, 2, 3}
	gOpts.relativenumber = false
	gOpts.reverse = false
	gOpts.rulerfile = ""
	gOpts.rulerfmt = ""
	gOpts.scrolloff = 0
	gOpts.searchmethod = textSearch
	gOpts.selectfmt = "\033[7;35m"
	gOpts.selmode = "all"
	gOpts.shell = gDefaultShell
	gOpts.shellflag = gDefaultShellFlag
	gOpts.shellopts = nil
	gOpts.showbinds = true
	gOpts.sizeunits = "binary"
	gOpts.smartcase = true
	gOpts.smartdia = false
	gOpts.sortby = naturalSort
	gOpts.sortignorecase = true
	gOpts.sortignoredia = true
	gOpts.statfmt = "\033[36m%p\033[0m| %c| %u| %g| %S| %t| -> %l"
	gOpts.tabstop = 8
	gOpts.tagfmt = "\033[31m"
	gOpts.tempmarks = "'"
	gOpts.terminalcursor = defaultCursor
	gOpts.timefmt = time.ANSIC
	gOpts.truncatechar = "~"
	gOpts.truncatepct = 100
	gOpts.visualfmt = "\033[7;36m"
	gOpts.waitmsg = "Press any key to continue"
	gOpts.watch = false
	gOpts.wrapscan = true
	gOpts.wrapscroll = false

	// Normal and Visual mode
	keys := map[string]expr{
		"k":          &callExpr{"up", nil, 1},
		"<up>":       &callExpr{"up", nil, 1},
		"<m-up>":     &callExpr{"up", nil, 1},
		"<c-u>":      &callExpr{"half-up", nil, 1},
		"<c-b>":      &callExpr{"page-up", nil, 1},
		"<pgup>":     &callExpr{"page-up", nil, 1},
		"<c-y>":      &callExpr{"scroll-up", nil, 1},
		"<c-m-up>":   &callExpr{"scroll-up", nil, 1},
		"j":          &callExpr{"down", nil, 1},
		"<down>":     &callExpr{"down", nil, 1},
		"<m-down>":   &callExpr{"down", nil, 1},
		"<c-d>":      &callExpr{"half-down", nil, 1},
		"<c-f>":      &callExpr{"page-down", nil, 1},
		"<pgdn>":     &callExpr{"page-down", nil, 1},
		"<c-e>":      &callExpr{"scroll-down", nil, 1},
		"<c-m-down>": &callExpr{"scroll-down", nil, 1},
		"h":          &callExpr{"updir", nil, 1},
		"<left>":     &callExpr{"updir", nil, 1},
		"l":          &callExpr{"open", nil, 1},
		"<right>":    &callExpr{"open", nil, 1},
		"q":          &callExpr{"quit", nil, 1},
		"gg":         &callExpr{"top", nil, 1},
		"<home>":     &callExpr{"top", nil, 1},
		"G":          &callExpr{"bottom", nil, 1},
		"<end>":      &callExpr{"bottom", nil, 1},
		"H":          &callExpr{"high", nil, 1},
		"M":          &callExpr{"middle", nil, 1},
		"L":          &callExpr{"low", nil, 1},
		"[":          &callExpr{"jump-prev", nil, 1},
		"]":          &callExpr{"jump-next", nil, 1},
		"t":          &callExpr{"tag-toggle", nil, 1},
		"u":          &callExpr{"unselect", nil, 1},
		"y":          &callExpr{"copy", nil, 1},
		"d":          &callExpr{"cut", nil, 1},
		"c":          &callExpr{"clear", nil, 1},
		"p":          &callExpr{"paste", nil, 1},
		"<c-l>":      &callExpr{"redraw", nil, 1},
		"<c-r>":      &callExpr{"reload", nil, 1},
		":":          &callExpr{"read", nil, 1},
		"$":          &callExpr{"shell", nil, 1},
		"%":          &callExpr{"shell-pipe", nil, 1},
		"!":          &callExpr{"shell-wait", nil, 1},
		"&":          &callExpr{"shell-async", nil, 1},
		"f":          &callExpr{"find", nil, 1},
		"F":          &callExpr{"find-back", nil, 1},
		";":          &callExpr{"find-next", nil, 1},
		",":          &callExpr{"find-prev", nil, 1},
		"/":          &callExpr{"search", nil, 1},
		"?":          &callExpr{"search-back", nil, 1},
		"n":          &callExpr{"search-next", nil, 1},
		"N":          &callExpr{"search-prev", nil, 1},
		"m":          &callExpr{"mark-save", nil, 1},
		"'":          &callExpr{"mark-load", nil, 1},
		`"`:          &callExpr{"mark-remove", nil, 1},
		`r`:          &callExpr{"rename", nil, 1},
		"<c-n>":      &callExpr{"cmd-history-next", nil, 1},
		"<c-p>":      &callExpr{"cmd-history-prev", nil, 1},

		"zh": &setExpr{"hidden!", ""},
		"zr": &setExpr{"reverse!", ""},
		"zn": &setExpr{"info", ""},
		"zs": &setExpr{"info", "size"},
		"zt": &setExpr{"info", "time"},
		"za": &setExpr{"info", "size:time"},
		"sn": &listExpr{[]expr{&setExpr{"sortby", "natural"}, &setExpr{"info", ""}}, 1},
		"ss": &listExpr{[]expr{&setExpr{"sortby", "size"}, &setExpr{"info", "size"}}, 1},
		"st": &listExpr{[]expr{&setExpr{"sortby", "time"}, &setExpr{"info", "time"}}, 1},
		"sa": &listExpr{[]expr{&setExpr{"sortby", "atime"}, &setExpr{"info", "atime"}}, 1},
		"sb": &listExpr{[]expr{&setExpr{"sortby", "btime"}, &setExpr{"info", "btime"}}, 1},
		"sc": &listExpr{[]expr{&setExpr{"sortby", "ctime"}, &setExpr{"info", "ctime"}}, 1},
		"se": &listExpr{[]expr{&setExpr{"sortby", "ext"}, &setExpr{"info", ""}}, 1},
		"gh": &callExpr{"cd", []string{"~"}, 1},
	}

	// insert bindings that apply to both Normal & Visual mode first
	gOpts.nkeys = maps.Clone(keys)
	// now add Normal mode specific ones
	gOpts.nkeys["<space>"] = &listExpr{[]expr{&callExpr{"toggle", nil, 1}, &callExpr{"down", nil, 1}}, 1}
	gOpts.nkeys["V"] = &callExpr{"visual", nil, 1}
	gOpts.nkeys["v"] = &callExpr{"invert", nil, 1}

	// now do the same for Visual mode
	gOpts.vkeys = maps.Clone(keys)
	gOpts.vkeys["<esc>"] = &callExpr{"visual-discard", nil, 1}
	gOpts.vkeys["V"] = &callExpr{"visual-accept", nil, 1}
	gOpts.vkeys["o"] = &callExpr{"visual-change", nil, 1}

	// Command-line mode bindings can be assigned directly
	gOpts.cmdkeys = map[string]expr{
		"<space>":       &callExpr{"cmd-insert", []string{" "}, 1},
		"<esc>":         &callExpr{"cmd-escape", nil, 1},
		"<tab>":         &callExpr{"cmd-complete", nil, 1},
		"<enter>":       &callExpr{"cmd-enter", nil, 1},
		"<c-j>":         &callExpr{"cmd-enter", nil, 1},
		"<down>":        &callExpr{"cmd-history-next", nil, 1},
		"<c-n>":         &callExpr{"cmd-history-next", nil, 1},
		"<up>":          &callExpr{"cmd-history-prev", nil, 1},
		"<c-p>":         &callExpr{"cmd-history-prev", nil, 1},
		"<delete>":      &callExpr{"cmd-delete", nil, 1},
		"<c-d>":         &callExpr{"cmd-delete", nil, 1},
		"<backspace>":   &callExpr{"cmd-delete-back", nil, 1},
		"<left>":        &callExpr{"cmd-left", nil, 1},
		"<c-b>":         &callExpr{"cmd-left", nil, 1},
		"<right>":       &callExpr{"cmd-right", nil, 1},
		"<c-f>":         &callExpr{"cmd-right", nil, 1},
		"<home>":        &callExpr{"cmd-home", nil, 1},
		"<c-a>":         &callExpr{"cmd-home", nil, 1},
		"<end>":         &callExpr{"cmd-end", nil, 1},
		"<c-e>":         &callExpr{"cmd-end", nil, 1},
		"<c-u>":         &callExpr{"cmd-delete-home", nil, 1},
		"<c-k>":         &callExpr{"cmd-delete-end", nil, 1},
		"<c-w>":         &callExpr{"cmd-delete-unix-word", nil, 1},
		"<c-y>":         &callExpr{"cmd-yank", nil, 1},
		"<c-t>":         &callExpr{"cmd-transpose", nil, 1},
		"<c-c>":         &callExpr{"cmd-interrupt", nil, 1},
		"<a-f>":         &callExpr{"cmd-word", nil, 1},
		"<a-b>":         &callExpr{"cmd-word-back", nil, 1},
		"<a-c>":         &callExpr{"cmd-capitalize-word", nil, 1},
		"<a-d>":         &callExpr{"cmd-delete-word", nil, 1},
		"<a-backspace>": &callExpr{"cmd-delete-word-back", nil, 1},
		"<a-u>":         &callExpr{"cmd-uppercase-word", nil, 1},
		"<a-l>":         &callExpr{"cmd-lowercase-word", nil, 1},
		"<a-t>":         &callExpr{"cmd-transpose-word", nil, 1},
	}

	gOpts.cmds = make(map[string]expr)
	gOpts.user = make(map[string]string)

	gLocalOpts.dircounts = make(map[string]bool)
	gLocalOpts.dirfirst = make(map[string]bool)
	gLocalOpts.dironly = make(map[string]bool)
	gLocalOpts.hidden = make(map[string]bool)
	gLocalOpts.info = make(map[string][]string)
	gLocalOpts.reverse = make(map[string]bool)
	gLocalOpts.sortby = make(map[string]sortMethod)
	gLocalOpts.sortignorecase = make(map[string]bool)
	gLocalOpts.sortignoredia = make(map[string]bool)

	setDefaults()
}
