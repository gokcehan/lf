package main

import (
	"strings"
	"testing"
)

// TODO: use a slice and merge with outputs

var inp0 = ""
var inp1 = "# comments start with '#'"
var inp2 = "set hidden # trailing comments are allowed"
var inp3 = "set hidden; set preview"
var inp4 = "set ratios 1:2:3"
var inp5 = "set ratios 1:2:3;"
var inp6 = ":set ratios 1:2:3"
var inp7 = ":set ratios 1:2:3;"
var inp8 = "map gh cd ~"
var inp9 = "map gh cd ~;"
var inp10 = "map gh :cd ~"
var inp11 = "map gh :cd ~;"
var inp12 = "cmd usage $du -h . | less"
var inp13 = "map u usage"
var inp14 = "map u usage;"
var inp15 = "map u :usage"
var inp16 = "map u :usage;"
var inp17 = "map u $du -h . | less"
var inp18 = `cmd usage $du -h "$1" | less`
var inp19 = "map u usage /"

var inp20 = `cmd gohome :{{
	cd ~
	set hidden
}}`

var inp21 = `map gh :{{
	cd ~
	set hidden
}}`

var inp22 = `map c ${{
	mkdir foo
	IFS=':'; cp ${fs} foo
	tar -czvf "foo.tar.gz" foo
	rm -rf foo
}}`

var inp23 = `cmd compress ${{
	mkdir "$1"
	IFS=':'; cp ${fs} "$1"
	tar -czvf "$1.tar.gz" "$1"
	rm -rf "$1"
}}`

// unfinished command
var inp24 = `cmd compress ${{
	mkdir "$1"`

var out0 = []string{}
var out1 = []string{}
var out2 = []string{"set", "hidden", "\n"}
var out3 = []string{"set", "hidden", ";", "set", "preview", "\n"}
var out4 = []string{"set", "ratios", "1:2:3", "\n"}
var out5 = []string{"set", "ratios", "1:2:3", ";"}
var out6 = []string{":", "set", "ratios", "1:2:3", "\n", "\n"}
var out7 = []string{":", "set", "ratios", "1:2:3", ";", "\n"}
var out8 = []string{"map", "gh", "cd", "~", "\n"}
var out9 = []string{"map", "gh", "cd", "~", ";"}
var out10 = []string{"map", "gh", ":", "cd", "~", "\n", "\n"}
var out11 = []string{"map", "gh", ":", "cd", "~", ";", "\n"}
var out12 = []string{"cmd", "usage", "$", "du -h . | less", "\n"}
var out13 = []string{"map", "u", "usage", "\n"}
var out14 = []string{"map", "u", "usage", ";"}
var out15 = []string{"map", "u", ":", "usage", "\n", "\n"}
var out16 = []string{"map", "u", ":", "usage", ";", "\n"}
var out17 = []string{"map", "u", "$", "du -h . | less", "\n"}
var out18 = []string{"cmd", "usage", "$", `du -h "$1" | less`, "\n"}
var out19 = []string{"map", "u", "usage", "/", "\n"}
var out20 = []string{"cmd", "gohome", ":", "{{", "cd", "~", "\n", "set", "hidden", "\n", "}}", "\n"}
var out21 = []string{"map", "gh", ":", "{{", "cd", "~", "\n", "set", "hidden", "\n", "}}", "\n"}
var out22 = []string{"map", "c", "$", "{{", "\n\tmkdir foo\n\tIFS=':'; cp ${fs} foo\n\ttar -czvf \"foo.tar.gz\" foo\n\trm -rf foo\n", "}}", "\n"}
var out23 = []string{"cmd", "compress", "$", "{{", "\n\tmkdir \"$1\"\n\tIFS=':'; cp ${fs} \"$1\"\n\ttar -czvf \"$1.tar.gz\" \"$1\"\n\trm -rf \"$1\"\n", "}}", "\n"}
var out24 = []string{"cmd", "compress", "$", "{{"}

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
	compare(t, inp0, out0)
	compare(t, inp1, out1)
	compare(t, inp2, out2)
	compare(t, inp3, out3)
	compare(t, inp4, out4)
	compare(t, inp5, out5)
	compare(t, inp6, out6)
	compare(t, inp7, out7)
	compare(t, inp8, out8)
	compare(t, inp9, out9)
	compare(t, inp10, out10)
	compare(t, inp11, out11)
	compare(t, inp12, out12)
	compare(t, inp13, out13)
	compare(t, inp14, out14)
	compare(t, inp15, out15)
	compare(t, inp16, out16)
	compare(t, inp17, out17)
	compare(t, inp18, out18)
	compare(t, inp19, out19)
	compare(t, inp20, out20)
	compare(t, inp21, out21)
	compare(t, inp22, out22)
	compare(t, inp23, out23)
	compare(t, inp24, out24)
}
