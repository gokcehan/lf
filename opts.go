package main

import (
	"path/filepath"
	"time"
)

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
	anchorfind       bool
	autoquit         bool
	borderfmt        string
	cursoractivefmt  string
	cursorparentfmt  string
	cursorpreviewfmt string
	dircache         bool
	dircounts        bool
	dironly          bool
	dirpreviews      bool
	drawbox          bool
	dupfilefmt       string
	globsearch       bool
	icons            bool
	ignorecase       bool
	ignoredia        bool
	incfilter        bool
	incsearch        bool
	mouse            bool
	number           bool
	preview          bool
	sixel            bool
	relativenumber   bool
	smartcase        bool
	smartdia         bool
	waitmsg          string
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
	ruler            []string
	rulerfmt         string
	preserve         []string
	shellopts        []string
	keys             map[string]expr
	cmdkeys          map[string]expr
	cmds             map[string]expr
	user             map[string]string
	sortType         sortType
	tempmarks        string
	numberfmt        string
	tagfmt           string
}

var gLocalOpts struct {
	sortMethods map[string]sortMethod
	dirfirsts   map[string]bool
	dironlys    map[string]bool
	hiddens     map[string]bool
	reverses    map[string]bool
	infos       map[string][]string
}

func localOptPaths(path string) []string {
	var list []string
	list = append(list, path)
	for curr := path; !isRoot(curr); curr = filepath.Dir(curr) {
		list = append(list, curr+string(filepath.Separator))
	}
	return list
}

func getSortMethod(path string) sortMethod {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.sortMethods[key]; ok {
			return val
		}
	}
	return gOpts.sortType.method
}

func getSortType(path string) sortType {
	method := getSortMethod(path)
	option := gOpts.sortType.option
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.dirfirsts[key]; ok {
			if val {
				option |= dirfirstSort
			} else {
				option &= ^dirfirstSort
			}
			break
		}
	}
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.hiddens[key]; ok {
			if val {
				option |= hiddenSort
			} else {
				option &= ^hiddenSort
			}
			break
		}
	}
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.reverses[key]; ok {
			if val {
				option |= reverseSort
			} else {
				option &= ^reverseSort
			}
			break
		}
	}
	val := sortType{
		method: method,
		option: option,
	}
	return val
}

func getDirOnly(path string) bool {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.dironlys[key]; ok {
			return val
		}
	}
	return gOpts.dironly
}

func getInfo(path string) []string {
	for _, key := range localOptPaths(path) {
		if val, ok := gLocalOpts.infos[key]; ok {
			return val
		}
	}
	return gOpts.info
}

func init() {
	gOpts.anchorfind = true
	gOpts.autoquit = false
	gOpts.dircache = true
	gOpts.dircounts = false
	gOpts.dironly = false
	gOpts.dirpreviews = false
	gOpts.drawbox = false
	gOpts.dupfilefmt = "%f.~%n~"
	gOpts.borderfmt = "\033[0m"
	gOpts.cursoractivefmt = "\033[7m"
	gOpts.cursorparentfmt = "\033[7m"
	gOpts.cursorpreviewfmt = "\033[4m"
	gOpts.globsearch = false
	gOpts.icons = false
	gOpts.ignorecase = true
	gOpts.ignoredia = true
	gOpts.incfilter = false
	gOpts.incsearch = false
	gOpts.mouse = false
	gOpts.number = false
	gOpts.preview = true
	gOpts.sixel = false
	gOpts.relativenumber = false
	gOpts.smartcase = true
	gOpts.smartdia = false
	gOpts.waitmsg = "Press any key to continue"
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
	gOpts.hiddenfiles = []string{".*"}
	gOpts.history = true
	gOpts.info = nil
	gOpts.ruler = nil
	gOpts.rulerfmt = "  %a|  %p|  \033[7;31m %m \033[0m|  \033[7;33m %c \033[0m|  \033[7;35m %s \033[0m|  \033[7;34m %f \033[0m|  %i/%t"
	gOpts.preserve = []string{"mode"}
	gOpts.shellopts = nil
	gOpts.sortType = sortType{naturalSort, dirfirstSort}
	gOpts.tempmarks = "'"
	gOpts.numberfmt = "\033[33m"
	gOpts.tagfmt = "\033[31m"

	gOpts.keys = make(map[string]expr)

	gOpts.keys["k"] = &callExpr{"up", nil, 1}
	gOpts.keys["<up>"] = &callExpr{"up", nil, 1}
	gOpts.keys["<m-up>"] = &callExpr{"up", nil, 1}
	gOpts.keys["<c-u>"] = &callExpr{"half-up", nil, 1}
	gOpts.keys["<c-b>"] = &callExpr{"page-up", nil, 1}
	gOpts.keys["<pgup>"] = &callExpr{"page-up", nil, 1}
	gOpts.keys["<c-y>"] = &callExpr{"scroll-up", nil, 1}
	gOpts.keys["<c-m-up>"] = &callExpr{"scroll-up", nil, 1}
	gOpts.keys["j"] = &callExpr{"down", nil, 1}
	gOpts.keys["<down>"] = &callExpr{"down", nil, 1}
	gOpts.keys["<m-down>"] = &callExpr{"down", nil, 1}
	gOpts.keys["<c-d>"] = &callExpr{"half-down", nil, 1}
	gOpts.keys["<c-f>"] = &callExpr{"page-down", nil, 1}
	gOpts.keys["<pgdn>"] = &callExpr{"page-down", nil, 1}
	gOpts.keys["<c-e>"] = &callExpr{"scroll-down", nil, 1}
	gOpts.keys["<c-m-down>"] = &callExpr{"scroll-down", nil, 1}
	gOpts.keys["h"] = &callExpr{"updir", nil, 1}
	gOpts.keys["<left>"] = &callExpr{"updir", nil, 1}
	gOpts.keys["l"] = &callExpr{"open", nil, 1}
	gOpts.keys["<right>"] = &callExpr{"open", nil, 1}
	gOpts.keys["q"] = &callExpr{"quit", nil, 1}
	gOpts.keys["gg"] = &callExpr{"top", nil, 1}
	gOpts.keys["<home>"] = &callExpr{"top", nil, 1}
	gOpts.keys["G"] = &callExpr{"bottom", nil, 1}
	gOpts.keys["<end>"] = &callExpr{"bottom", nil, 1}
	gOpts.keys["H"] = &callExpr{"high", nil, 1}
	gOpts.keys["M"] = &callExpr{"middle", nil, 1}
	gOpts.keys["L"] = &callExpr{"low", nil, 1}
	gOpts.keys["["] = &callExpr{"jump-prev", nil, 1}
	gOpts.keys["]"] = &callExpr{"jump-next", nil, 1}
	gOpts.keys["<space>"] = &listExpr{[]expr{&callExpr{"toggle", nil, 1}, &callExpr{"down", nil, 1}}, 1}
	gOpts.keys["t"] = &callExpr{"tag-toggle", nil, 1}
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
	gOpts.keys["sn"] = &listExpr{[]expr{&setExpr{"sortby", "natural"}, &setExpr{"info", ""}}, 1}
	gOpts.keys["ss"] = &listExpr{[]expr{&setExpr{"sortby", "size"}, &setExpr{"info", "size"}}, 1}
	gOpts.keys["st"] = &listExpr{[]expr{&setExpr{"sortby", "time"}, &setExpr{"info", "time"}}, 1}
	gOpts.keys["sa"] = &listExpr{[]expr{&setExpr{"sortby", "atime"}, &setExpr{"info", "atime"}}, 1}
	gOpts.keys["sc"] = &listExpr{[]expr{&setExpr{"sortby", "ctime"}, &setExpr{"info", "ctime"}}, 1}
	gOpts.keys["se"] = &listExpr{[]expr{&setExpr{"sortby", "ext"}, &setExpr{"info", ""}}, 1}
	gOpts.keys["gh"] = &callExpr{"cd", []string{"~"}, 1}

	gOpts.cmdkeys = make(map[string]expr)

	gOpts.cmdkeys["<space>"] = &callExpr{"cmd-insert", []string{" "}, 1}
	gOpts.cmdkeys["<esc>"] = &callExpr{"cmd-escape", nil, 1}
	gOpts.cmdkeys["<tab>"] = &callExpr{"cmd-complete", nil, 1}
	gOpts.cmdkeys["<enter>"] = &callExpr{"cmd-enter", nil, 1}
	gOpts.cmdkeys["<c-j>"] = &callExpr{"cmd-enter", nil, 1}
	gOpts.cmdkeys["<down>"] = &callExpr{"cmd-history-next", nil, 1}
	gOpts.cmdkeys["<c-n>"] = &callExpr{"cmd-history-next", nil, 1}
	gOpts.cmdkeys["<up>"] = &callExpr{"cmd-history-prev", nil, 1}
	gOpts.cmdkeys["<c-p>"] = &callExpr{"cmd-history-prev", nil, 1}
	gOpts.cmdkeys["<delete>"] = &callExpr{"cmd-delete", nil, 1}
	gOpts.cmdkeys["<c-d>"] = &callExpr{"cmd-delete", nil, 1}
	gOpts.cmdkeys["<backspace>"] = &callExpr{"cmd-delete-back", nil, 1}
	gOpts.cmdkeys["<backspace2>"] = &callExpr{"cmd-delete-back", nil, 1}
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
	gOpts.cmdkeys["<a-backspace>"] = &callExpr{"cmd-delete-word-back", nil, 1}
	gOpts.cmdkeys["<a-backspace2>"] = &callExpr{"cmd-delete-word-back", nil, 1}
	gOpts.cmdkeys["<a-u>"] = &callExpr{"cmd-uppercase-word", nil, 1}
	gOpts.cmdkeys["<a-l>"] = &callExpr{"cmd-lowercase-word", nil, 1}
	gOpts.cmdkeys["<a-t>"] = &callExpr{"cmd-transpose-word", nil, 1}

	gOpts.cmds = make(map[string]expr)
	gOpts.user = make(map[string]string)

	gLocalOpts.sortMethods = make(map[string]sortMethod)
	gLocalOpts.dirfirsts = make(map[string]bool)
	gLocalOpts.dironlys = make(map[string]bool)
	gLocalOpts.hiddens = make(map[string]bool)
	gLocalOpts.reverses = make(map[string]bool)
	gLocalOpts.infos = make(map[string][]string)

	setDefaults()
}
