package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/sys/windows"
)

func createTestApp() *app {
	screen := tcell.NewSimulationScreen("UTF-8")
	err := screen.Init()
	if err != nil {
		panic(err)
	}
	ui := newUI(screen)
	nav := newNav(ui.wins[0].h)
	app := newApp(ui, nav)
	return app
}

func readConfig(app *app, config_string string) {
	filepath := "lfrc"
	white := func() {
		rc, err := os.Create(filepath)
		if err != nil {
			panic(err)
		}
		defer rc.Close()

		_, err = rc.WriteString(config_string)
		if err != nil {
			panic(err)
		}
		_, err = rc.Seek(0, io.SeekStart)
		if err != nil {
			panic(err)
		}
	}
	white()
	app.readFile(filepath)
}

func evalShellExpr(app *app, exp string) string {
	p := newParser(strings.NewReader(exp))
	for p.parse() {
		old := os.Stdout // keep backup of the real stdout
		r, w, _ := os.Pipe()
		os.Stderr = w
		defer func() { os.Stderr = old }() // restoring the real stdout

		outC := make(chan string)
		// copy the output in a separate goroutine so printing can't block indefinitely
		go func() {
			var buf bytes.Buffer
			_, err := io.Copy(&buf, r)
			if err != nil {
				panic(err)
			}
			outC <- buf.String()
		}()

		p.expr.eval(app, nil)

		// back to normal state
		w.Close()
		out := <-outC
		return out
	}
	if p.err != nil {
		panic(p.err)
	}
	return ""
}

func cdToTempDir(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	wd, err := os.Getwd()
	_ = wd
	if err != nil {
		panic(err)
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		err = os.Chdir(wd)
		if err != nil {
			panic(err)
		}
	})
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

type testCmd struct {
	rc_cmd string
	ui_cmd string
}

type cmdShells struct {
	posix testCmd
	pwsh  testCmd
	cmd   testCmd
	nu    testCmd
}

func sameTestCmd(rc_cmd string, ui_cmd string) cmdShells {
	c := testCmd{rc_cmd, ui_cmd}
	return cmdShells{c, c, c, c}
}

type expected struct {
	args string
	exp  string
}

var gShellCommandtests = []struct {
	name string
	cmd  cmdShells
	exp  []expected
}{
	// {
	// 	"lfdoc",
	// 	sameTestCmd(`lf -doc`, `lf -doc`),
	// 	[]expected{{"", genDocString}},
	// },
	{
		"test_echo",
		cmdShells{
			posix: testCmd{`printf '%s\n' "$@"`, `printf '%s\n'`},
			pwsh:  testCmd{`Write-Output $args`, `Write-Output`},
			cmd:   testCmd{`".\test echo.bat"`, `".\test echo.bat"`},
			nu:    testCmd{`print`, `print`},
		},
		[]expected{
			{`a "b c"`, "a\nb c\n"},
			{`a "'b'"`, "a\n'b'\n"},
			{`"iam|not a pipe"`, "iam|not a pipe\n"},
			{`"I'm Special"`, "I'm Special\n"},
			// {`"%NotAppData%"`, "%NotAppData%\n"},
			// {`"^NotEscaped"`, "^NotEscaped\n"},
			// {`"(NotAGroup)"`, "(NotAGroup)\n"},
			// {`"a<b"`, "a<b\n"},
			// {`"malicious&whoami"`, "a<b\n"},
		},
	},

	// {
	// 	"mkdir",
	// 	cmdShells{
	// 		posix: testCmd{`mkdir  "$1" && rm -r "$1"`, `mkdir`},
	// 		pwsh:  testCmd{`$null = New-Item $args[0] -ItemType Directory && Remove-Item $args[0]`, ` $null = New-Item -ItemType Directory`},
	// 		cmd:   testCmd{``, `mkdir`},
	// 		nu:    testCmd{``, `mkdir`},
	// 	},
	// 	[]expected{
	// 		{`"foo bar"`, ""},
	// 		{`"foo 'bar'"`, ""},
	// 	},
	// },
}

func runTestCases(app *app, t *testing.T, shell_type string) {
	for _, c := range gShellCommandtests {
		t.Run(c.name, func(t *testing.T) {
			var cmd testCmd
			switch shell_type {
			case "posix":
				cmd = c.cmd.posix
			case "pwsh":
				cmd = c.cmd.pwsh
			case "cmd":
				cmd = c.cmd.cmd
			case "nu":
				cmd = c.cmd.nu
			}
			if cmd.rc_cmd != "" {
				readConfig(app, fmt.Sprintf(`cmd %s $ {{ %s }}`, c.name, cmd.rc_cmd))
			}
			for _, e := range c.exp {
				t.Run(c.name, func(t *testing.T) {
					for i, run_cmd := range [][]string{{"rc", c.name}, {"ui", "$" + cmd.ui_cmd}} {
						t.Run(run_cmd[0], func(t *testing.T) {
							if i == 0 && cmd.rc_cmd == "" {
								t.Skip("Command does not supported", c.name)
								return
							}
							rtn := evalShellExpr(app, ":"+run_cmd[1]+" "+e.args)
							// remove windows specific crap
							rtn = strings.ReplaceAll(rtn, "\r\n", "\n")

							fmt.Println(rtn)
							if rtn != e.exp {
								t.Errorf("expected '%#v' but got '%#v'", e.exp, rtn)
							}
						})
					}
				})
			}
		})
	}
}

func TestShellCommand_Cmd(t *testing.T) {
	if runtime.GOOS != "windows" || !commandExists("cmd.exe") {
		t.Skip("cmd.exe is not available")
	}
	cdToTempDir(t)

	app := createTestApp()
	readConfig(app, `
set shellcmd 'cmd /D /c %c %a'
`)

	createBat := func() {
		tmpBat := "test echo.bat"
		f, err := os.Create(tmpBat)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		_, err = f.WriteString(`
@echo off
for %%x in (%*) do echo %%~x
`)
		if err != nil {
			panic(err)
		}
	}
	createBat()
	runTestCases(app, t, "cmd")
}

func TestShellCommand_Bash(t *testing.T) {
	if !commandExists("bash") {
		t.Skip("bash is not available")
	}
	cdToTempDir(t)

	app := createTestApp()
	readConfig(app, `
set shellcmd 'bash --norc --noprofile -c %c -- %a'
`)
	runTestCases(app, t, "posix")
}

func TestShellCommand_Pwsh(t *testing.T) {
	if !commandExists("pwsh") {
		t.Skip("pwsh is not available")
	}
	cdToTempDir(t)

	app := createTestApp()
	readConfig(app, `
set shellcmd 'pwsh -NoLogo -NoProfile -NonInteractive -CommandWithArgs %c %a'
`)
	runTestCases(app, t, "pwsh")
}

func TestShellCommand_Nu(t *testing.T) {
	if !commandExists("nu") {
		t.Skip("nu is not available")
	}
	cdToTempDir(t)

	app := createTestApp()

	readConfig(app, `
set shellcmd 'nu test_runner.nu -c %c %a'
`)

	// NOTE: nushell does not support arguments to -c yet
	// https://github.com/nushell/nushell/pull/12344
	createRunner := func() {
		tmpNu := "test_runner.nu"
		f, err := os.Create(tmpNu)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		_, err = f.WriteString(`

def main [-c:string, ...args] {
	let args = ($args | each { $"r#'($in)'#" } | str join ' ')
	nu -c $"($c) ($args)"
}
`)
		if err != nil {
			panic(err)
		}
	}
	createRunner()
	runTestCases(app, t, "nu")

	// TODO: if we want to support replacement %c inside any string " %c"
	// than we should follow some quoting rules
	// but each shell may have different rules
	// so we cannot support each external case
	// 	readConfig(app, `
	// set shellcmd 'nu -c "do {|...args| %c %a"'
	// cmd test $ echo ...$args | str join ' ' o> testoutput
	// cmd mkdir $ mkdir -v ...$args
	// cmd doc $ lf -doc | ignore
	// cmd toggle2 $ lf -remote "send $env.id toggle"
	// `)
}


// When user invokes cmd from command line he already escapes all arguments himself
// so no need to escape already escaped arg





func TestMyCmd(t *testing.T) {
	cases := []string{
		`argument1 "she said, "you had me at hello""  "\some\path with\spaces`,
		// We need to escape " -> \" inside "..."
		`argument1 "she said, \"you had me at hello\""  "\some\path with\spaces`,

		// also for cmd.exe
		`argument1 ^"she said, \^"you had me at hello\^"^"  ^"\some\path with\spaces`,

		`argument1 "argument"2" argument3 argument4`,
		`argument1 "argument\"2" argument3 argument4`,
		// for cmd
		`argument1 ^"argument\^"2^" argument3 argument4`,

		`"\some\directory with\spaces\" argument2`,
		// We need to escape '\' only if it follows by " \" -> \\"
		`"\some\directory with\spaces\\" argument2`,
		// for cmd
		`^"\some\directory with\spaces\\^" argument2`,

		`"malicious argument\" &whoami"`,

		`^"malicious argument\^" ^&whoami^"`,
	}
	for _, c := range cases {
		_ = c
		// for _, i := range []string{"echo.bat", "cecho"} {
		// _ = i
		cmd := exec.Command("cmd.exe")
		cmd.SysProcAttr = &windows.SysProcAttr{
			CmdLine: fmt.Sprintf(`/D /c printf "%%s\n" %s`, c),
		}
		cmd.Dir = `W:\`
		fmt.Println(cmd.SysProcAttr)
		rtn, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v %v", rtn, err)
		}
		fmt.Println(string(rtn))
		// }
	}
}

