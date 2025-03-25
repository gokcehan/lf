//go:build windows
// +build windows

package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

type EscapeTest struct {
	args    []string
	expArgv string
	expCmd  string
}

var escapeTests = []EscapeTest{
	{[]string{``}, "\"\"", `^"^"`},
	{
		[]string{`argument1`, `she said, "you had me at hello"`, `"\some\path with\spaces`},
		`argument1 "she said, \"you had me at hello\"" "\"\some\path with\spaces"`,
		`argument1 ^"she^ said^,^ \^"you^ had^ me^ at^ hello\^"^" ^"\^"\some\path^ with\spaces^"`,
	},
	{
		[]string{`argument1`, `argument"2`, `argument3`, `argument4`},
		`argument1 "argument\"2" argument3 argument4`,
		`argument1 ^"argument\^"2^" argument3 argument4`,
	},
	{
		[]string{`\some\directory with\spaces\`, `argument2`},
		`"\some\directory with\spaces\\" argument2`,
		`^"\some\directory^ with\spaces\\^" argument2`,
	},
	{
		[]string{`malicious argument" &whoami`},
		`"malicious argument\" &whoami"`,
		`^"malicious^ argument\^"^ ^&whoami^"`,
	},
}

func TestEscapeArg(t *testing.T) {
	for _, test := range escapeTests {
		t.Run("escapeArg", func(t *testing.T) {
			t.Run("Argv", func(t *testing.T) {
				var args []string
				for _, a := range test.args {
					args = append(args, escapeArg(a))
				}
				arg := strings.Join(args, " ")
				if arg != test.expArgv {
					t.Errorf("expected `%#v` but got `%#v`", test.expArgv, arg)
				}
			})
			t.Run("Cmd", func(t *testing.T) {
				var args []string
				for _, a := range test.args {
					args = append(args, escapeCmd(escapeArg(a)))
				}
				arg := strings.Join(args, " ")
				if arg != test.expCmd {
					t.Errorf("expected %#v but got %#v", test.expCmd, arg)
				}
			})
		})
	}
}

func runCmd(c string, t *testing.T) {
	t.Helper()
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &windows.SysProcAttr{
		CmdLine: fmt.Sprintf(`/D /c cecho %s`, c),
	}
	cmd.Dir = `W:\`
	rtn, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v %v", rtn, err)
	}
	fmt.Println(string(rtn))
}

func TestMyCmd2(t *testing.T) {
	for _, test := range escapeTests {
		args := []string{}
		for _, arg := range test.args {
			args = append(args, escapeCmd(escapeArg(arg)))
		}
		c := strings.Join(args, " ")
		runCmd(c, t)
	}
}
