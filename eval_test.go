package main

import (
	"reflect"
	"testing"
)

// These inputs are used in scan and parse tests.
var gEvalTests = []struct {
	inp   string
	toks  []string
	exprs []expr
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
		"echo hello",
		[]string{"echo", "hello", "\n"},
		[]expr{&callExpr{"echo", []string{"hello"}}},
	},

	{
		"echo hello world",
		[]string{"echo", "hello", "world", "\n"},
		[]expr{&callExpr{"echo", []string{"hello", "world"}}},
	},

	{
		"echo 'hello world'",
		[]string{"echo", "hello world", "\n"},
		[]expr{&callExpr{"echo", []string{"hello world"}}},
	},

	{
		`echo hello\ world`,
		[]string{"echo", "hello world", "\n"},
		[]expr{&callExpr{"echo", []string{"hello world"}}},
	},

	{
		`echo hello\	world`,
		[]string{"echo", "hello\tworld", "\n"},
		[]expr{&callExpr{"echo", []string{"hello\tworld"}}},
	},

	{
		"echo hello\\\nworld",
		[]string{"echo", "hello\nworld", "\n"},
		[]expr{&callExpr{"echo", []string{"hello\nworld"}}},
	},

	{
		`echo hello\aworld`,
		[]string{"echo", "helloworld", "\n"},
		[]expr{&callExpr{"echo", []string{"helloworld"}}},
	},

	{
		"set hidden # trailing comments are allowed",
		[]string{"set", "hidden", "\n"},
		[]expr{&setExpr{"hidden", ""}},
	},

	{
		"set hidden; set preview",
		[]string{"set", "hidden", ";", "set", "preview", "\n"},
		[]expr{&setExpr{"hidden", ""}, &setExpr{"preview", ""}},
	},

	{
		"set hidden\nset preview",
		[]string{"set", "hidden", "\n", "set", "preview", "\n"},
		[]expr{&setExpr{"hidden", ""}, &setExpr{"preview", ""}},
	},

	{
		"set ratios 1:2:3",
		[]string{"set", "ratios", "1:2:3", "\n"},
		[]expr{&setExpr{"ratios", "1:2:3"}},
	},

	{
		"set ratios 1:2:3;",
		[]string{"set", "ratios", "1:2:3", ";"},
		[]expr{&setExpr{"ratios", "1:2:3"}},
	},

	{
		":set ratios 1:2:3",
		[]string{":", "set", "ratios", "1:2:3", "\n", "\n"},
		[]expr{&listExpr{[]expr{&setExpr{"ratios", "1:2:3"}}}},
	},

	{
		":set ratios 1:2:3\nset hidden",
		[]string{":", "set", "ratios", "1:2:3", "\n", "\n", "set", "hidden", "\n"},
		[]expr{&listExpr{[]expr{&setExpr{"ratios", "1:2:3"}}}, &setExpr{"hidden", ""}},
	},

	{
		":set ratios 1:2:3;",
		[]string{":", "set", "ratios", "1:2:3", ";", "\n"},
		[]expr{&listExpr{[]expr{&setExpr{"ratios", "1:2:3"}}}},
	},

	{
		":set ratios 1:2:3;\nset hidden",
		[]string{":", "set", "ratios", "1:2:3", ";", "\n", "set", "hidden", "\n"},
		[]expr{&listExpr{[]expr{&setExpr{"ratios", "1:2:3"}}}, &setExpr{"hidden", ""}},
	},

	{
		"map gh cd ~",
		[]string{"map", "gh", "cd", "~", "\n"},
		[]expr{&mapExpr{"gh", &callExpr{"cd", []string{"~"}}}},
	},

	{
		"map gh cd ~;",
		[]string{"map", "gh", "cd", "~", ";"},
		[]expr{&mapExpr{"gh", &callExpr{"cd", []string{"~"}}}},
	},

	{
		"map gh :cd ~",
		[]string{"map", "gh", ":", "cd", "~", "\n", "\n"},
		[]expr{&mapExpr{"gh", &listExpr{[]expr{&callExpr{"cd", []string{"~"}}}}}},
	},

	{
		"map gh :cd ~;",
		[]string{"map", "gh", ":", "cd", "~", ";", "\n"},
		[]expr{&mapExpr{"gh", &listExpr{[]expr{&callExpr{"cd", []string{"~"}}}}}},
	},

	{
		"cmd usage $du -h . | less",
		[]string{"cmd", "usage", "$", "du -h . | less", "\n"},
		[]expr{&cmdExpr{"usage", &execExpr{"$", "du -h . | less"}}},
	},

	{
		"cmd 世界 $echo 世界",
		[]string{"cmd", "世界", "$", "echo 世界", "\n"},
		[]expr{&cmdExpr{"世界", &execExpr{"$", "echo 世界"}}},
	},

	{
		"map u usage",
		[]string{"map", "u", "usage", "\n"},
		[]expr{&mapExpr{"u", &callExpr{"usage", nil}}},
	},

	{
		"map u usage;",
		[]string{"map", "u", "usage", ";"},
		[]expr{&mapExpr{"u", &callExpr{"usage", nil}}},
	},

	{
		"map u :usage",
		[]string{"map", "u", ":", "usage", "\n", "\n"},
		[]expr{&mapExpr{"u", &listExpr{[]expr{&callExpr{"usage", nil}}}}},
	},

	{
		"map u :usage;",
		[]string{"map", "u", ":", "usage", ";", "\n"},
		[]expr{&mapExpr{"u", &listExpr{[]expr{&callExpr{"usage", nil}}}}},
	},

	{
		"map r push :rename<space>",
		[]string{"map", "r", "push", ":rename<space>", "\n"},
		[]expr{&mapExpr{"r", &callExpr{"push", []string{":rename<space>"}}}},
	},

	{
		"map r push :rename<space>;",
		[]string{"map", "r", "push", ":rename<space>;", "\n"},
		[]expr{&mapExpr{"r", &callExpr{"push", []string{":rename<space>;"}}}},
	},

	{
		"map r push :rename<space> # trailing comments are allowed after a space",
		[]string{"map", "r", "push", ":rename<space>", "\n"},
		[]expr{&mapExpr{"r", &callExpr{"push", []string{":rename<space>"}}}},
	},

	{
		"map r :push :rename<space>",
		[]string{"map", "r", ":", "push", ":rename<space>", "\n", "\n"},
		[]expr{&mapExpr{"r", &listExpr{[]expr{&callExpr{"push", []string{":rename<space>"}}}}}},
	},

	{
		"map r :push :rename<space> ; set hidden",
		[]string{"map", "r", ":", "push", ":rename<space>", ";", "set", "hidden", "\n", "\n"},
		[]expr{&mapExpr{"r", &listExpr{[]expr{&callExpr{"push", []string{":rename<space>"}}, &setExpr{"hidden", ""}}}}},
	},

	{
		"map u $du -h . | less",
		[]string{"map", "u", "$", "du -h . | less", "\n"},
		[]expr{&mapExpr{"u", &execExpr{"$", "du -h . | less"}}},
	},

	{
		"cmd usage $du -h \"$1\" | less",
		[]string{"cmd", "usage", "$", `du -h "$1" | less`, "\n"},
		[]expr{&cmdExpr{"usage", &execExpr{"$", `du -h "$1" | less`}}},
	},

	{
		"map u usage /",
		[]string{"map", "u", "usage", "/", "\n"},
		[]expr{&mapExpr{"u", &callExpr{"usage", []string{"/"}}}},
	},

	{
		"map ss :set sortby size; set info size",
		[]string{"map", "ss", ":", "set", "sortby", "size", ";", "set", "info", "size", "\n", "\n"},
		[]expr{&mapExpr{"ss", &listExpr{[]expr{&setExpr{"sortby", "size"}, &setExpr{"info", "size"}}}}},
	},

	{
		"map ss :set sortby size; set info size;",
		[]string{"map", "ss", ":", "set", "sortby", "size", ";", "set", "info", "size", ";", "\n"},
		[]expr{&mapExpr{"ss", &listExpr{[]expr{&setExpr{"sortby", "size"}, &setExpr{"info", "size"}}}}},
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
		[]expr{&cmdExpr{"gohome", &listExpr{[]expr{
			&callExpr{"cd", []string{"~"}},
			&setExpr{"hidden", ""}}},
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
		[]expr{&mapExpr{"gh", &listExpr{[]expr{
			&callExpr{"cd", []string{"~"}},
			&setExpr{"hidden", ""}}},
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
		[]expr{&mapExpr{"c", &execExpr{"$", `
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
		[]expr{&cmdExpr{"compress", &execExpr{"$", `
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
