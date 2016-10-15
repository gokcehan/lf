package main

import (
	"reflect"
	"testing"
)

// These inputs are used in scan and parse tests.
var gTests = []struct {
	inp   string
	toks  []string
	exprs []Expr
}{
	{
		"",
		[]string{},
		nil,
	},

	{
		"# comments start with '#'",
		[]string{},
		nil,
	},

	{
		"set hidden # trailing comments are allowed",
		[]string{"set", "hidden", "\n"},
		[]Expr{&SetExpr{"hidden", ""}},
	},

	{
		"set hidden; set preview",
		[]string{"set", "hidden", ";", "set", "preview", "\n"},
		[]Expr{&SetExpr{"hidden", ""}, &SetExpr{"preview", ""}},
	},

	{
		"set ratios 1:2:3",
		[]string{"set", "ratios", "1:2:3", "\n"},
		[]Expr{&SetExpr{"ratios", "1:2:3"}},
	},

	{
		"set ratios 1:2:3;",
		[]string{"set", "ratios", "1:2:3", ";"},
		[]Expr{&SetExpr{"ratios", "1:2:3"}},
	},

	{
		":set ratios 1:2:3",
		[]string{":", "set", "ratios", "1:2:3", "\n", "\n"},
		[]Expr{&ListExpr{[]Expr{&SetExpr{"ratios", "1:2:3"}}}},
	},

	{
		":set ratios 1:2:3;",
		[]string{":", "set", "ratios", "1:2:3", ";", "\n"},
		[]Expr{&ListExpr{[]Expr{&SetExpr{"ratios", "1:2:3"}}}},
	},

	{
		"map gh cd ~",
		[]string{"map", "gh", "cd", "~", "\n"},
		[]Expr{&MapExpr{"gh", &CallExpr{"cd", []string{"~"}}}},
	},

	{
		"map gh cd ~;",
		[]string{"map", "gh", "cd", "~", ";"},
		[]Expr{&MapExpr{"gh", &CallExpr{"cd", []string{"~"}}}},
	},

	{
		"map gh :cd ~",
		[]string{"map", "gh", ":", "cd", "~", "\n", "\n"},
		[]Expr{&MapExpr{"gh", &ListExpr{[]Expr{&CallExpr{"cd", []string{"~"}}}}}},
	},

	{
		"map gh :cd ~;",
		[]string{"map", "gh", ":", "cd", "~", ";", "\n"},
		[]Expr{&MapExpr{"gh", &ListExpr{[]Expr{&CallExpr{"cd", []string{"~"}}}}}},
	},

	{
		"cmd usage $du -h . | less",
		[]string{"cmd", "usage", "$", "du -h . | less", "\n"},
		[]Expr{&CmdExpr{"usage", &ExecExpr{"$", "du -h . | less"}}},
	},

	{
		"map u usage",
		[]string{"map", "u", "usage", "\n"},
		[]Expr{&MapExpr{"u", &CallExpr{"usage", nil}}},
	},

	{
		"map u usage;",
		[]string{"map", "u", "usage", ";"},
		[]Expr{&MapExpr{"u", &CallExpr{"usage", nil}}},
	},

	{
		"map u :usage",
		[]string{"map", "u", ":", "usage", "\n", "\n"},
		[]Expr{&MapExpr{"u", &ListExpr{[]Expr{&CallExpr{"usage", nil}}}}},
	},

	{
		"map u :usage;",
		[]string{"map", "u", ":", "usage", ";", "\n"},
		[]Expr{&MapExpr{"u", &ListExpr{[]Expr{&CallExpr{"usage", nil}}}}},
	},

	{
		"map r push :rename<space>",
		[]string{"map", "r", "push", ":rename<space>", "\n"},
		[]Expr{&MapExpr{"r", &CallExpr{"push", []string{":rename<space>"}}}},
	},

	{
		"map r push :rename<space>;",
		[]string{"map", "r", "push", ":rename<space>;", "\n"},
		[]Expr{&MapExpr{"r", &CallExpr{"push", []string{":rename<space>;"}}}},
	},

	{
		"map r push :rename<space> # trailing comments area allowed after a space",
		[]string{"map", "r", "push", ":rename<space>", "\n"},
		[]Expr{&MapExpr{"r", &CallExpr{"push", []string{":rename<space>"}}}},
	},

	{
		"map r :push :rename<space>",
		[]string{"map", "r", ":", "push", ":rename<space>", "\n", "\n"},
		[]Expr{&MapExpr{"r", &ListExpr{[]Expr{&CallExpr{"push", []string{":rename<space>"}}}}}},
	},

	{
		"map r :push :rename<space> ; set hidden",
		[]string{"map", "r", ":", "push", ":rename<space>", ";", "set", "hidden", "\n", "\n"},
		[]Expr{&MapExpr{"r", &ListExpr{[]Expr{&CallExpr{"push", []string{":rename<space>"}}, &SetExpr{"hidden", ""}}}}},
	},

	{
		"map u $du -h . | less",
		[]string{"map", "u", "$", "du -h . | less", "\n"},
		[]Expr{&MapExpr{"u", &ExecExpr{"$", "du -h . | less"}}},
	},

	{
		"cmd usage $du -h \"$1\" | less",
		[]string{"cmd", "usage", "$", `du -h "$1" | less`, "\n"},
		[]Expr{&CmdExpr{"usage", &ExecExpr{"$", `du -h "$1" | less`}}},
	},

	{
		"map u usage /",
		[]string{"map", "u", "usage", "/", "\n"},
		[]Expr{&MapExpr{"u", &CallExpr{"usage", []string{"/"}}}},
	},

	{
		"map ss :set sortby size; set showinfo size",
		[]string{"map", "ss", ":", "set", "sortby", "size", ";", "set", "showinfo", "size", "\n", "\n"},
		[]Expr{&MapExpr{"ss", &ListExpr{[]Expr{&SetExpr{"sortby", "size"}, &SetExpr{"showinfo", "size"}}}}},
	},

	{
		"map ss :set sortby size; set showinfo size;",
		[]string{"map", "ss", ":", "set", "sortby", "size", ";", "set", "showinfo", "size", ";", "\n"},
		[]Expr{&MapExpr{"ss", &ListExpr{[]Expr{&SetExpr{"sortby", "size"}, &SetExpr{"showinfo", "size"}}}}},
	},

	{
		`cmd gohome :{{
			cd ~
			set hidden
		}}`,
		[]string{"cmd", "gohome", ":", "{{",
			"cd", "~", "\n",
			"set", "hidden", "\n",
			"}}", "\n"},
		[]Expr{&CmdExpr{"gohome", &ListExpr{[]Expr{
			&CallExpr{"cd", []string{"~"}},
			&SetExpr{"hidden", ""}}},
		}},
	},

	{
		`map gh :{{
			cd ~
			set hidden
		}}`,
		[]string{"map", "gh", ":", "{{",
			"cd", "~", "\n",
			"set", "hidden", "\n",
			"}}", "\n"},
		[]Expr{&MapExpr{"gh", &ListExpr{[]Expr{
			&CallExpr{"cd", []string{"~"}},
			&SetExpr{"hidden", ""}}},
		}},
	},

	{
		`map c ${{
			mkdir foo
			IFS=':'; cp ${fs} foo
			tar -czvf "foo.tar.gz" foo
			rm -rf foo
		}}`,
		[]string{"map", "c", "$", "{{", `
			mkdir foo
			IFS=':'; cp ${fs} foo
			tar -czvf "foo.tar.gz" foo
			rm -rf foo
		`, "}}", "\n"},
		[]Expr{&MapExpr{"c", &ExecExpr{"$", `
			mkdir foo
			IFS=':'; cp ${fs} foo
			tar -czvf "foo.tar.gz" foo
			rm -rf foo
		`}}},
	},

	{
		`cmd compress ${{
			mkdir "$1"
			IFS=':'; cp ${fs} "$1"
			tar -czvf "$1.tar.gz" "$1"
			rm -rf "$1"
		}}`,
		[]string{"cmd", "compress", "$", "{{", `
			mkdir "$1"
			IFS=':'; cp ${fs} "$1"
			tar -czvf "$1.tar.gz" "$1"
			rm -rf "$1"
		`, "}}", "\n"},
		[]Expr{&CmdExpr{"compress", &ExecExpr{"$", `
			mkdir "$1"
			IFS=':'; cp ${fs} "$1"
			tar -czvf "$1.tar.gz" "$1"
			rm -rf "$1"
		`}}},
	},
}

func TestSplitKeys(t *testing.T) {
	inps := []struct {
		s    string
		keys []string
	}{
		{"", nil},
		{"j", []string{"j"}},
		{"jk", []string{"j", "k"}},
		{"1j", []string{"1", "j"}},
		{"42j", []string{"4", "2", "j"}},
		{"<space>", []string{"<space>"}},
		{"j<space>", []string{"j", "<space>"}},
		{"j<space>k", []string{"j", "<space>", "k"}},
		{"1j<space>k", []string{"1", "j", "<space>", "k"}},
		{"1j<space>1k", []string{"1", "j", "<space>", "1", "k"}},
		{"<>", []string{"<>"}},
		{"<space", []string{"<space"}},
		{"<space<", []string{"<space<"}},
		{"<space<>", []string{"<space<>"}},
		{"><space>", []string{">", "<space>"}},
		{"><space>>", []string{">", "<space>", ">"}},
	}

	for _, inp := range inps {
		if keys := splitKeys(inp.s); !reflect.DeepEqual(keys, inp.keys) {
			t.Errorf("at input '%s' expected '%v' but got '%v'", inp.s, inp.keys, keys)
		}
	}
}
