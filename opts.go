package main

import (
	"maps"
	"path/filepath"
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
	return method == naturalSort ||
		method == nameSort ||
		method == sizeSort ||
		method == timeSort ||
		method == atimeSort ||
		method == btimeSort ||
		method == ctimeSort ||
		method == extSort ||
		method == customSort
}

const invalidSortErrorMessage = `sortby: value should either be 'natural', 'name', 'size', 'time', 'atime', 'btime', 'ctime', 'ext' or 'custom'`

var gOpts struct {
	anchorfind       bool
	autoquit         bool
	borderfmt        string
	copyfmt          string
	cursoractivefmt  string
	cursorparentfmt  string
	cursorpreviewfmt string
	cutfmt           string
	dircache         bool
	dircounts        bool
	dirfirst         bool
	dironly          bool
	dirpreviews      bool
	drawbox          bool
	dupfilefmt       string
	globfilter       bool
	globsearch       bool
	hidden           bool
	icons            bool
	ignorecase       bool
	ignoredia        bool
	incfilter        bool
	incsearch        bool
	locale           string
	mouse            bool
	number           bool
	preview          bool
	relativenumber   bool
	reverse          bool
	roundbox         bool
	selectfmt        string
	visualfmt        string
	showbinds        bool
	sixel            bool
	sortby           sortMethod
	smartcase        bool
	smartdia         bool
	waitmsg          string
	watch            bool
	wrapscan         bool
	wrapscroll       bool
	findlen          int
	period           int
	scrolloff        int
	tabstop          int
	errorfmt         string
	filesep          string
	ifs              string
	previewer        string
	cleaner          string
	promptfmt        string
	selmode          string
	shell            string
	shellflag        string
	statfmt          string
	timefmt          string
	infotimefmtnew   string
	infotimefmtold   string
	truncatechar     string
	truncatepct      int
	ratios           []int
	hiddenfiles      []string
	history          bool
	info             []string
	rulerfmt         string
	preserve         []string
	shellopts        []string
	nkeys            map[string]expr
	vkeys            map[string]expr
	cmdkeys          map[string]expr
	cmds             map[string]expr
	user             map[string]string
	tempmarks        string
	numberfmt        string
	tagfmt           string
}

var gLocalOpts struct {
	sortby    map[string]sortMethod
	dircounts map[string]bool
	dirfirst  map[string]bool
	dironly   map[string]bool
	hidden    map[string]bool
	reverse   map[string]bool
	info      map[string][]string
	locale    map[string]string
}

func localOptPaths(path string) []string {
	list := []string{path}
	for curr := path; !isRoot(curr); curr = filepath.Dir(curr) {
		list = append(list, curr+string(filepath.Separator))
	}
	return list
}

func getDirCounts(path string) bool {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.dircounts[key]; ok {
			return val
		}
	}
	return gOpts.dircounts
}

func getDirFirst(path string) bool {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.dirfirst[key]; ok {
			return val
		}
	}
	return gOpts.dirfirst
}

func getDirOnly(path string) bool {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.dironly[key]; ok {
			return val
		}
	}
	return gOpts.dironly
}

func getHidden(path string) bool {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.hidden[key]; ok {
			return val
		}
	}
	return gOpts.hidden
}

func getInfo(path string) []string {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.info[key]; ok {
			return val
		}
	}
	return gOpts.info
}

func getReverse(path string) bool {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.reverse[key]; ok {
			return val
		}
	}
	return gOpts.reverse
}

func getSortBy(path string) sortMethod {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.sortby[key]; ok {
			return val
		}
	}
	return gOpts.sortby
}

func getLocale(path string) string {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.locale[key]; ok {
			return val
		}
	}
	return gOpts.locale
}

func init() {
	gOpts.anchorfind = true
	gOpts.autoquit = true
	gOpts.dircache = true
	gOpts.dircounts = false
	gOpts.dirfirst = true
	gOpts.dironly = false
	gOpts.dirpreviews = false
	gOpts.drawbox = false
	gOpts.dupfilefmt = "%f.~%n~"
	gOpts.borderfmt = "\033[0m"
	gOpts.copyfmt = "\033[7;33m"
	gOpts.cursoractivefmt = "\033[7m"
	gOpts.cursorparentfmt = "\033[7m"
	gOpts.cursorpreviewfmt = "\033[4m"
	gOpts.cutfmt = "\033[7;31m"
	gOpts.globfilter = false
	gOpts.globsearch = false
	gOpts.hidden = false
	gOpts.icons = false
	gOpts.ignorecase = true
	gOpts.ignoredia = true
	gOpts.incfilter = false
	gOpts.incsearch = false
	gOpts.locale = localeStrDisable
	gOpts.mouse = false
	gOpts.number = false
	gOpts.preview = true
	gOpts.relativenumber = false
	gOpts.reverse = false
	gOpts.roundbox = false
	gOpts.selectfmt = "\033[7;35m"
	gOpts.visualfmt = "\033[7;36m"
	gOpts.showbinds = true
	gOpts.sixel = false
	gOpts.sortby = naturalSort
	gOpts.smartcase = true
	gOpts.smartdia = false
	gOpts.waitmsg = "Press any key to continue"
	gOpts.watch = false
	gOpts.wrapscan = true
	gOpts.wrapscroll = false
	gOpts.findlen = 1
	gOpts.period = 0
	gOpts.scrolloff = 0
	gOpts.tabstop = 8
	gOpts.errorfmt = "\033[7;31;47m"
	gOpts.filesep = "\n"
	gOpts.ifs = ""
	gOpts.previewer = ""
	gOpts.cleaner = ""
	gOpts.promptfmt = "\033[32;1m%u@%h\033[0m:\033[34;1m%d\033[0m\033[1m%f\033[0m"
	gOpts.selmode = "all"
	gOpts.shell = gDefaultShell
	gOpts.shellflag = gDefaultShellFlag
	gOpts.statfmt = "\033[36m%p\033[0m| %c| %u| %g| %S| %t| -> %l"
	gOpts.timefmt = time.ANSIC
	gOpts.infotimefmtnew = "Jan _2 15:04"
	gOpts.infotimefmtold = "Jan _2  2006"
	gOpts.truncatechar = "~"
	gOpts.truncatepct = 100
	gOpts.ratios = []int{1, 2, 3}
	gOpts.hiddenfiles = gDefaultHiddenFiles
	gOpts.history = true
	gOpts.info = nil
	gOpts.rulerfmt = "  %a|  %p|  \033[7;31m %m \033[0m|  \033[7;33m %c \033[0m|  \033[7;35m %s \033[0m|  \033[7;36m %v \033[0m|  \033[7;34m %f \033[0m|  %i/%t"
	gOpts.preserve = []string{"mode"}
	gOpts.shellopts = nil
	gOpts.tempmarks = "'"
	gOpts.numberfmt = "\033[33m"
	gOpts.tagfmt = "\033[31m"

	// normal and visual mode
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

	// insert bindings that apply to both normal & visual mode first
	gOpts.nkeys = maps.Clone(keys)
	// now add normal mode specific ones
	gOpts.nkeys["<space>"] = &listExpr{[]expr{&callExpr{"toggle", nil, 1}, &callExpr{"down", nil, 1}}, 1}
	gOpts.nkeys["V"] = &callExpr{"visual", nil, 1}
	gOpts.nkeys["v"] = &callExpr{"invert", nil, 1}

	// now do the same for visual mode
	gOpts.vkeys = maps.Clone(keys)
	gOpts.vkeys["<esc>"] = &callExpr{"visual-discard", nil, 1}
	gOpts.vkeys["V"] = &callExpr{"visual-accept", nil, 1}
	gOpts.vkeys["o"] = &callExpr{"visual-change", nil, 1}

	// command line mode bindings can be assigned directly
	gOpts.cmdkeys = map[string]expr{
		"<space>":        &callExpr{"cmd-insert", []string{" "}, 1},
		"<esc>":          &callExpr{"cmd-escape", nil, 1},
		"<tab>":          &callExpr{"cmd-complete", nil, 1},
		"<enter>":        &callExpr{"cmd-enter", nil, 1},
		"<c-j>":          &callExpr{"cmd-enter", nil, 1},
		"<down>":         &callExpr{"cmd-history-next", nil, 1},
		"<c-n>":          &callExpr{"cmd-history-next", nil, 1},
		"<up>":           &callExpr{"cmd-history-prev", nil, 1},
		"<c-p>":          &callExpr{"cmd-history-prev", nil, 1},
		"<delete>":       &callExpr{"cmd-delete", nil, 1},
		"<c-d>":          &callExpr{"cmd-delete", nil, 1},
		"<backspace>":    &callExpr{"cmd-delete-back", nil, 1},
		"<backspace2>":   &callExpr{"cmd-delete-back", nil, 1},
		"<left>":         &callExpr{"cmd-left", nil, 1},
		"<c-b>":          &callExpr{"cmd-left", nil, 1},
		"<right>":        &callExpr{"cmd-right", nil, 1},
		"<c-f>":          &callExpr{"cmd-right", nil, 1},
		"<home>":         &callExpr{"cmd-home", nil, 1},
		"<c-a>":          &callExpr{"cmd-home", nil, 1},
		"<end>":          &callExpr{"cmd-end", nil, 1},
		"<c-e>":          &callExpr{"cmd-end", nil, 1},
		"<c-u>":          &callExpr{"cmd-delete-home", nil, 1},
		"<c-k>":          &callExpr{"cmd-delete-end", nil, 1},
		"<c-w>":          &callExpr{"cmd-delete-unix-word", nil, 1},
		"<c-y>":          &callExpr{"cmd-yank", nil, 1},
		"<c-t>":          &callExpr{"cmd-transpose", nil, 1},
		"<c-c>":          &callExpr{"cmd-interrupt", nil, 1},
		"<a-f>":          &callExpr{"cmd-word", nil, 1},
		"<a-b>":          &callExpr{"cmd-word-back", nil, 1},
		"<a-c>":          &callExpr{"cmd-capitalize-word", nil, 1},
		"<a-d>":          &callExpr{"cmd-delete-word", nil, 1},
		"<a-backspace>":  &callExpr{"cmd-delete-word-back", nil, 1},
		"<a-backspace2>": &callExpr{"cmd-delete-word-back", nil, 1},
		"<a-u>":          &callExpr{"cmd-uppercase-word", nil, 1},
		"<a-l>":          &callExpr{"cmd-lowercase-word", nil, 1},
		"<a-t>":          &callExpr{"cmd-transpose-word", nil, 1},
	}

	gOpts.cmds = make(map[string]expr)
	gOpts.user = make(map[string]string)

	gLocalOpts.sortby = make(map[string]sortMethod)
	gLocalOpts.dircounts = make(map[string]bool)
	gLocalOpts.dirfirst = make(map[string]bool)
	gLocalOpts.dironly = make(map[string]bool)
	gLocalOpts.hidden = make(map[string]bool)
	gLocalOpts.reverse = make(map[string]bool)
	gLocalOpts.info = make(map[string][]string)
	gLocalOpts.locale = make(map[string]string)

	setDefaults()
}
