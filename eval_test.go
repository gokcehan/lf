package main

import (
	"reflect"
	"strings"
	"testing"
)

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
		[]expr{&callExpr{"echo", []string{"hello"}, 1}},
	},

	{
		"echo hello world",
		[]string{"echo", "hello", "world", "\n"},
		[]expr{&callExpr{"echo", []string{"hello", "world"}, 1}},
	},

	{
		"echo 'hello world'",
		[]string{"echo", "hello world", "\n"},
		[]expr{&callExpr{"echo", []string{"hello world"}, 1}},
	},

	{
		`echo "hello world"`,
		[]string{"echo", "hello world", "\n"},
		[]expr{&callExpr{"echo", []string{"hello world"}, 1}},
	},

	{
		`echo "hello\"world"`,
		[]string{"echo", `hello"world`, "\n"},
		[]expr{&callExpr{"echo", []string{`hello"world`}, 1}},
	},

	{
		`echo "hello\tworld"`,
		[]string{"echo", "hello\tworld", "\n"},
		[]expr{&callExpr{"echo", []string{"hello\tworld"}, 1}},
	},

	{
		`echo "hello\nworld"`,
		[]string{"echo", "hello\nworld", "\n"},
		[]expr{&callExpr{"echo", []string{"hello\nworld"}, 1}},
	},

	{
		`echo "hello\zworld"`,
		[]string{"echo", `hello\zworld`, "\n"},
		[]expr{&callExpr{"echo", []string{`hello\zworld`}, 1}},
	},

	{
		`echo "hello\0world"`,
		[]string{"echo", "hello\000world", "\n"},
		[]expr{&callExpr{"echo", []string{"hello\000world"}, 1}},
	},

	{
		`echo "hello\101world"`,
		[]string{"echo", "hello\101world", "\n"},
		[]expr{&callExpr{"echo", []string{"hello\101world"}, 1}},
	},

	{
		`echo hello\ world`,
		[]string{"echo", "hello world", "\n"},
		[]expr{&callExpr{"echo", []string{"hello world"}, 1}},
	},

	{
		"echo hello\\\tworld",
		[]string{"echo", "hello\tworld", "\n"},
		[]expr{&callExpr{"echo", []string{"hello\tworld"}, 1}},
	},

	{
		"echo hello\\\nworld",
		[]string{"echo", "hello\nworld", "\n"},
		[]expr{&callExpr{"echo", []string{"hello\nworld"}, 1}},
	},

	{
		`echo hello\\world`,
		[]string{"echo", `hello\world`, "\n"},
		[]expr{&callExpr{"echo", []string{`hello\world`}, 1}},
	},

	{
		`echo hello\zworld`,
		[]string{"echo", "hellozworld", "\n"},
		[]expr{&callExpr{"echo", []string{"hellozworld"}, 1}},
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
		`set ifs ""`,
		[]string{"set", "ifs", "", "\n"},
		[]expr{&setExpr{"ifs", ""}},
	},

	{
		`set ifs "\n"`,
		[]string{"set", "ifs", "\n", "\n"},
		[]expr{&setExpr{"ifs", "\n"}},
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
		[]expr{&listExpr{[]expr{&setExpr{"ratios", "1:2:3"}}, 1}},
	},

	{
		":set ratios 1:2:3\nset hidden",
		[]string{":", "set", "ratios", "1:2:3", "\n", "\n", "set", "hidden", "\n"},
		[]expr{&listExpr{[]expr{&setExpr{"ratios", "1:2:3"}}, 1}, &setExpr{"hidden", ""}},
	},

	{
		":set ratios 1:2:3;",
		[]string{":", "set", "ratios", "1:2:3", ";", "\n"},
		[]expr{&listExpr{[]expr{&setExpr{"ratios", "1:2:3"}}, 1}},
	},

	{
		":set ratios 1:2:3;\nset hidden",
		[]string{":", "set", "ratios", "1:2:3", ";", "\n", "set", "hidden", "\n"},
		[]expr{&listExpr{[]expr{&setExpr{"ratios", "1:2:3"}}, 1}, &setExpr{"hidden", ""}},
	},

	{
		"set ratios 1:2:3\n set hidden",
		[]string{"set", "ratios", "1:2:3", "\n", "set", "hidden", "\n"},
		[]expr{&setExpr{"ratios", "1:2:3"}, &setExpr{"hidden", ""}},
	},

	{
		"set ratios 1:2:3 \nset hidden",
		[]string{"set", "ratios", "1:2:3", "\n", "set", "hidden", "\n"},
		[]expr{&setExpr{"ratios", "1:2:3"}, &setExpr{"hidden", ""}},
	},

	{
		"setlocal /foo/bar hidden # trailing comments are allowed",
		[]string{"setlocal", "/foo/bar", "hidden", "\n"},
		[]expr{&setLocalExpr{"/foo/bar", "hidden", ""}},
	},

	{
		"setlocal /foo/bar hidden; setlocal /foo/bar reverse",
		[]string{"setlocal", "/foo/bar", "hidden", ";", "setlocal", "/foo/bar", "reverse", "\n"},
		[]expr{&setLocalExpr{"/foo/bar", "hidden", ""}, &setLocalExpr{"/foo/bar", "reverse", ""}},
	},

	{
		"setlocal /foo/bar hidden\nsetlocal /foo/bar reverse",
		[]string{"setlocal", "/foo/bar", "hidden", "\n", "setlocal", "/foo/bar", "reverse", "\n"},
		[]expr{&setLocalExpr{"/foo/bar", "hidden", ""}, &setLocalExpr{"/foo/bar", "reverse", ""}},
	},

	{
		`setlocal /foo/bar info ""`,
		[]string{"setlocal", "/foo/bar", "info", "", "\n"},
		[]expr{&setLocalExpr{"/foo/bar", "info", ""}},
	},

	{
		`setlocal /foo/bar info "size"`,
		[]string{"setlocal", "/foo/bar", "info", "size", "\n"},
		[]expr{&setLocalExpr{"/foo/bar", "info", "size"}},
	},

	{
		"setlocal /foo/bar info size:time",
		[]string{"setlocal", "/foo/bar", "info", "size:time", "\n"},
		[]expr{&setLocalExpr{"/foo/bar", "info", "size:time"}},
	},

	{
		"setlocal /foo/bar info size:time;",
		[]string{"setlocal", "/foo/bar", "info", "size:time", ";"},
		[]expr{&setLocalExpr{"/foo/bar", "info", "size:time"}},
	},

	{
		":setlocal /foo/bar info size:time",
		[]string{":", "setlocal", "/foo/bar", "info", "size:time", "\n", "\n"},
		[]expr{&listExpr{[]expr{&setLocalExpr{"/foo/bar", "info", "size:time"}}, 1}},
	},

	{
		":setlocal /foo/bar info size:time\nsetlocal /foo/bar hidden",
		[]string{":", "setlocal", "/foo/bar", "info", "size:time", "\n", "\n", "setlocal", "/foo/bar", "hidden", "\n"},
		[]expr{&listExpr{[]expr{&setLocalExpr{"/foo/bar", "info", "size:time"}}, 1}, &setLocalExpr{"/foo/bar", "hidden", ""}},
	},

	{
		":setlocal /foo/bar info size:time;",
		[]string{":", "setlocal", "/foo/bar", "info", "size:time", ";", "\n"},
		[]expr{&listExpr{[]expr{&setLocalExpr{"/foo/bar", "info", "size:time"}}, 1}},
	},

	{
		":setlocal /foo/bar info size:time;\nsetlocal /foo/bar hidden",
		[]string{":", "setlocal", "/foo/bar", "info", "size:time", ";", "\n", "setlocal", "/foo/bar", "hidden", "\n"},
		[]expr{&listExpr{[]expr{&setLocalExpr{"/foo/bar", "info", "size:time"}}, 1}, &setLocalExpr{"/foo/bar", "hidden", ""}},
	},

	{
		"setlocal /foo/bar info size:time\n setlocal /foo/bar hidden",
		[]string{"setlocal", "/foo/bar", "info", "size:time", "\n", "setlocal", "/foo/bar", "hidden", "\n"},
		[]expr{&setLocalExpr{"/foo/bar", "info", "size:time"}, &setLocalExpr{"/foo/bar", "hidden", ""}},
	},

	{
		"setlocal /foo/bar info size:time \nsetlocal /foo/bar hidden",
		[]string{"setlocal", "/foo/bar", "info", "size:time", "\n", "setlocal", "/foo/bar", "hidden", "\n"},
		[]expr{&setLocalExpr{"/foo/bar", "info", "size:time"}, &setLocalExpr{"/foo/bar", "hidden", ""}},
	},

	{
		"map gh cd ~",
		[]string{"map", "gh", "cd", "~", "\n"},
		[]expr{&mapExpr{"gh", &callExpr{"cd", []string{"~"}, 1}}},
	},

	{
		"map gh cd ~;",
		[]string{"map", "gh", "cd", "~", ";"},
		[]expr{&mapExpr{"gh", &callExpr{"cd", []string{"~"}, 1}}},
	},

	{
		"map gh :cd ~",
		[]string{"map", "gh", ":", "cd", "~", "\n", "\n"},
		[]expr{&mapExpr{"gh", &listExpr{[]expr{&callExpr{"cd", []string{"~"}, 1}}, 1}}},
	},

	{
		"map gh :cd ~;",
		[]string{"map", "gh", ":", "cd", "~", ";", "\n"},
		[]expr{&mapExpr{"gh", &listExpr{[]expr{&callExpr{"cd", []string{"~"}, 1}}, 1}}},
	},

	{
		"cmap <c-g> cmd-escape",
		[]string{"cmap", "<c-g>", "cmd-escape", "\n"},
		[]expr{&cmapExpr{"<c-g>", &callExpr{"cmd-escape", nil, 1}}},
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
		[]expr{&mapExpr{"u", &callExpr{"usage", nil, 1}}},
	},

	{
		"map u usage;",
		[]string{"map", "u", "usage", ";"},
		[]expr{&mapExpr{"u", &callExpr{"usage", nil, 1}}},
	},

	{
		"map u :usage",
		[]string{"map", "u", ":", "usage", "\n", "\n"},
		[]expr{&mapExpr{"u", &listExpr{[]expr{&callExpr{"usage", nil, 1}}, 1}}},
	},

	{
		"map u :usage;",
		[]string{"map", "u", ":", "usage", ";", "\n"},
		[]expr{&mapExpr{"u", &listExpr{[]expr{&callExpr{"usage", nil, 1}}, 1}}},
	},

	{
		"map r push :rename<space>",
		[]string{"map", "r", "push", ":rename<space>", "\n"},
		[]expr{&mapExpr{"r", &callExpr{"push", []string{":rename<space>"}, 1}}},
	},

	{
		"map r push :rename<space>;",
		[]string{"map", "r", "push", ":rename<space>;", "\n"},
		[]expr{&mapExpr{"r", &callExpr{"push", []string{":rename<space>;"}, 1}}},
	},

	{
		"map r push :rename<space> # trailing comments are allowed after a space",
		[]string{"map", "r", "push", ":rename<space>", "\n"},
		[]expr{&mapExpr{"r", &callExpr{"push", []string{":rename<space>"}, 1}}},
	},

	{
		"map r :push :rename<space>",
		[]string{"map", "r", ":", "push", ":rename<space>", "\n", "\n"},
		[]expr{&mapExpr{"r", &listExpr{[]expr{&callExpr{"push", []string{":rename<space>"}, 1}}, 1}}},
	},

	{
		"map r :push :rename<space> ; set hidden",
		[]string{"map", "r", ":", "push", ":rename<space>", ";", "set", "hidden", "\n", "\n"},
		[]expr{&mapExpr{"r", &listExpr{[]expr{&callExpr{"push", []string{":rename<space>"}, 1}, &setExpr{"hidden", ""}}, 1}}},
	},

	{
		"map u $du -h . | less",
		[]string{"map", "u", "$", "du -h . | less", "\n"},
		[]expr{&mapExpr{"u", &execExpr{"$", "du -h . | less"}}},
	},

	{
		"cmd usage $du -h $1 | less",
		[]string{"cmd", "usage", "$", "du -h $1 | less", "\n"},
		[]expr{&cmdExpr{"usage", &execExpr{"$", "du -h $1 | less"}}},
	},

	{
		"map u usage /",
		[]string{"map", "u", "usage", "/", "\n"},
		[]expr{&mapExpr{"u", &callExpr{"usage", []string{"/"}, 1}}},
	},

	{
		"map ss :set sortby size; set info size",
		[]string{"map", "ss", ":", "set", "sortby", "size", ";", "set", "info", "size", "\n", "\n"},
		[]expr{&mapExpr{"ss", &listExpr{[]expr{&setExpr{"sortby", "size"}, &setExpr{"info", "size"}}, 1}}},
	},

	{
		"map ss :set sortby size; set info size;",
		[]string{"map", "ss", ":", "set", "sortby", "size", ";", "set", "info", "size", ";", "\n"},
		[]expr{&mapExpr{"ss", &listExpr{[]expr{&setExpr{"sortby", "size"}, &setExpr{"info", "size"}}, 1}}},
	},

	{
		`cmd gohome :{{
			cd ~
			set hidden
		}}`,
		[]string{
			"cmd", "gohome", ":", "{{",
			"cd", "~", "\n",
			"set", "hidden", "\n",
			"}}", "\n",
		},
		[]expr{&cmdExpr{
			"gohome", &listExpr{[]expr{
				&callExpr{"cd", []string{"~"}, 1},
				&setExpr{"hidden", ""},
			}, 1},
		}},
	},

	{
		`map gh :{{
			cd ~
			set hidden
		}}`,
		[]string{
			"map", "gh", ":", "{{",
			"cd", "~", "\n",
			"set", "hidden", "\n",
			"}}", "\n",
		},
		[]expr{&mapExpr{
			"gh", &listExpr{[]expr{
				&callExpr{"cd", []string{"~"}, 1},
				&setExpr{"hidden", ""},
			}, 1},
		}},
	},

	{
		`map c ${{
			mkdir foo
			cp $fs foo
			tar -czvf foo.tar.gz foo
			rm -rf foo
		}}`,
		[]string{"map", "c", "$", "{{", `
			mkdir foo
			cp $fs foo
			tar -czvf foo.tar.gz foo
			rm -rf foo
		`, "}}", "\n"},
		[]expr{&mapExpr{"c", &execExpr{"$", `
			mkdir foo
			cp $fs foo
			tar -czvf foo.tar.gz foo
			rm -rf foo
		`}}},
	},

	{
		`cmd compress ${{
			mkdir $1
			cp $fs $1
			tar -czvf $1.tar.gz $1
			rm -rf $1
		}}`,
		[]string{"cmd", "compress", "$", "{{", `
			mkdir $1
			cp $fs $1
			tar -czvf $1.tar.gz $1
			rm -rf $1
		`, "}}", "\n"},
		[]expr{&cmdExpr{"compress", &execExpr{"$", `
			mkdir $1
			cp $fs $1
			tar -czvf $1.tar.gz $1
			rm -rf $1
		`}}},
	},
}

func TestScan(t *testing.T) {
	for _, test := range gEvalTests {
		s := newScanner(strings.NewReader(test.inp))

		for _, tok := range test.toks {
			if s.scan(); s.tok != tok {
				t.Errorf("at input '%s' expected '%s' but scanned '%s'", test.inp, tok, s.tok)
			}
		}

		if s.scan() {
			t.Errorf("at input '%s' unexpected '%s'", test.inp, s.tok)
		}
	}
}

func TestParse(t *testing.T) {
	for _, test := range gEvalTests {
		p := newParser(strings.NewReader(test.inp))

		for _, expr := range test.exprs {
			if p.parse(); !reflect.DeepEqual(p.expr, expr) {
				t.Errorf("at input '%s' expected '%s' but parsed '%s'", test.inp, expr, p.expr)
			}
		}

		if p.parse(); p.expr != nil {
			t.Errorf("at input '%s' unexpected '%s'", test.inp, p.expr)
		}
	}
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
