package main

var gOpts struct {
	hidden    bool
	preview   bool
	scrolloff int
	tabstop   int
	ifs       string
	previewer string
	shell     string
	showinfo  string
	sortby    string
	dirfirst  bool
	ratios    []int
	keys      map[string]Expr
	cmds      map[string]Expr
}

func init() {
	gOpts.hidden = false
	gOpts.preview = true
	gOpts.scrolloff = 0
	gOpts.tabstop = 8
	gOpts.shell = envShell
	gOpts.showinfo = "none"
	gOpts.sortby = "name"
	gOpts.dirfirst = true
	gOpts.ratios = []int{1, 2, 3}

	gOpts.keys = make(map[string]Expr)

	gOpts.keys["k"] = &CallExpr{"up", nil}
	gOpts.keys["<up>"] = &CallExpr{"up", nil}
	gOpts.keys["<c-u>"] = &CallExpr{"half-up", nil}
	gOpts.keys["<c-b>"] = &CallExpr{"page-up", nil}
	gOpts.keys["j"] = &CallExpr{"down", nil}
	gOpts.keys["<down>"] = &CallExpr{"down", nil}
	gOpts.keys["<c-d>"] = &CallExpr{"half-down", nil}
	gOpts.keys["<c-f>"] = &CallExpr{"page-down", nil}
	gOpts.keys["h"] = &CallExpr{"updir", nil}
	gOpts.keys["<left>"] = &CallExpr{"updir", nil}
	gOpts.keys["l"] = &CallExpr{"open", nil}
	gOpts.keys["<right>"] = &CallExpr{"open", nil}
	gOpts.keys["q"] = &CallExpr{"quit", nil}
	gOpts.keys["G"] = &CallExpr{"bot", nil}
	gOpts.keys["gg"] = &CallExpr{"top", nil}
	gOpts.keys[":"] = &CallExpr{"read", nil}
	gOpts.keys["$"] = &CallExpr{"read-shell", nil}
	gOpts.keys["!"] = &CallExpr{"read-shell-wait", nil}
	gOpts.keys["&"] = &CallExpr{"read-shell-async", nil}
	gOpts.keys["/"] = &CallExpr{"search", nil}
	gOpts.keys["?"] = &CallExpr{"search-back", nil}
	gOpts.keys["<space>"] = &CallExpr{"toggle", nil}
	gOpts.keys["y"] = &CallExpr{"yank", nil}
	gOpts.keys["d"] = &CallExpr{"delete", nil}
	gOpts.keys["p"] = &CallExpr{"paste", nil}
	gOpts.keys["<c-l>"] = &CallExpr{"renew", nil}

	gOpts.cmds = make(map[string]Expr)
}
