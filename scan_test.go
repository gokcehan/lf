package main

import (
	"strings"
	"testing"
)

var inps = []struct {
	s    string
	toks []string
}{
	// single line commands
	{"", []string{}},
	{"# comments start with '#'", []string{}},
	{"set hidden # trailing comments are allowed", []string{"set", "hidden", "\n"}},
	{"set hidden; set preview", []string{"set", "hidden", ";", "set", "preview", "\n"}},
	{"set ratios 1:2:3", []string{"set", "ratios", "1:2:3", "\n"}},
	{"set ratios 1:2:3;", []string{"set", "ratios", "1:2:3", ";"}},
	{":set ratios 1:2:3", []string{":", "set", "ratios", "1:2:3", "\n", "\n"}},
	{":set ratios 1:2:3;", []string{":", "set", "ratios", "1:2:3", ";", "\n"}},
	{"map gh cd ~", []string{"map", "gh", "cd", "~", "\n"}},
	{"map gh cd ~;", []string{"map", "gh", "cd", "~", ";"}},
	{"map gh :cd ~", []string{"map", "gh", ":", "cd", "~", "\n", "\n"}},
	{"map gh :cd ~;", []string{"map", "gh", ":", "cd", "~", ";", "\n"}},
	{"cmd usage $du -h . | less", []string{"cmd", "usage", "$", "du -h . | less", "\n"}},
	{"map u usage", []string{"map", "u", "usage", "\n"}},
	{"map u usage;", []string{"map", "u", "usage", ";"}},
	{"map u :usage", []string{"map", "u", ":", "usage", "\n", "\n"}},
	{"map u :usage;", []string{"map", "u", ":", "usage", ";", "\n"}},
	{"map u $du -h . | less", []string{"map", "u", "$", "du -h . | less", "\n"}},
	{"cmd usage $du -h \"$1\" | less", []string{"cmd", "usage", "$", `du -h "$1" | less`, "\n"}},
	{"map u usage /", []string{"map", "u", "usage", "/", "\n"}},

	// multiline commands
	{"cmd gohome :{{\n\tcd ~\n\tset hidden\n}}",
		[]string{"cmd", "gohome", ":", "{{",
			"cd", "~", "\n",
			"set", "hidden", "\n",
			"}}", "\n"}},

	{"map gh :{{\n\tcd ~\n\tset hidden\n}}",
		[]string{"map", "gh", ":", "{{",
			"cd", "~", "\n",
			"set", "hidden", "\n",
			"}}", "\n"}},

	{"map c ${{\n\tmkdir foo\n\tIFS=':'; cp ${fs} foo\n\ttar -czvf \"foo.tar.gz\" foo\n\trm -rf foo\n}}",
		[]string{"map", "c", "$", "{{",
			"\n\tmkdir foo\n\tIFS=':'; cp ${fs} foo\n\ttar -czvf \"foo.tar.gz\" foo\n\trm -rf foo\n",
			"}}", "\n"}},

	{"cmd compress ${{\n\tmkdir \"$1\"\n\tIFS=':'; cp ${fs} \"$1\"\n\ttar -czvf \"$1.tar.gz\" \"$1\"\n\trm -rf \"$1\"\n}}",
		[]string{"cmd", "compress", "$", "{{",
			"\n\tmkdir \"$1\"\n\tIFS=':'; cp ${fs} \"$1\"\n\ttar -czvf \"$1.tar.gz\" \"$1\"\n\trm -rf \"$1\"\n",
			"}}", "\n"}},

	// unfinished command
	{"cmd compress ${{\n\tmkdir \"$1\"",
		[]string{"cmd", "compress", "$", "{{"}},
}

func compare(t *testing.T, inp string, out []string) {
	s := newScanner(strings.NewReader(inp))

	for _, tok := range out {
		if s.scan(); s.tok != tok {
			t.Errorf("at input '%s' expected '%s' but scanned '%s'", inp, tok, s.tok)
		}
	}

	if s.scan() {
		t.Errorf("at input '%s' unexpected '%s'", inp, s.tok)
	}
}

func TestScan(t *testing.T) {
	for _, inp := range inps {
		compare(t, inp.s, inp.toks)
	}
}
