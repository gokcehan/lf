package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

var (
	envOpener = os.Getenv("OPENER")
	envEditor = os.Getenv("VISUAL")
	envPager  = os.Getenv("PAGER")
	envShell  = os.Getenv("SHELL")
)

var envPathExt = os.Getenv("PATHEXT")

var (
	gDefaultShell       = "cmd"
	gDefaultShellFlag   = "/c"
	gDefaultSocketProt  = "unix"
	gDefaultSocketPath  string
	gDefaultHiddenFiles []string
)

var (
	gUser        *user.User
	gConfigPaths []string
	gColorsPaths []string
	gIconsPaths  []string
	gFilesPath   string
	gTagsPath    string
	gMarksPath   string
	gHistoryPath string
)

func init() {
	if envOpener == "" {
		envOpener = `start ""`
	}

	if envEditor == "" {
		envEditor = os.Getenv("EDITOR")
		if envEditor == "" {
			envEditor = "notepad"
		}
	}

	if envPager == "" {
		envPager = "more"
	}

	if envShell == "" {
		envShell = "cmd"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	username := os.Getenv("USERNAME")
	if username == "" {
		panic("$USERNAME variable is empty or not set")
	}

	gUser = &user.User{
		HomeDir:  homeDir,
		Username: username,
	}

	data := os.Getenv("LF_CONFIG_HOME")
	if data == "" {
		data = os.Getenv("LOCALAPPDATA")
	}

	gConfigPaths = []string{
		filepath.Join(os.Getenv("ProgramData"), "lf", "lfrc"),
		filepath.Join(data, "lf", "lfrc"),
	}

	gColorsPaths = []string{
		filepath.Join(os.Getenv("ProgramData"), "lf", "colors"),
		filepath.Join(data, "lf", "colors"),
	}

	gIconsPaths = []string{
		filepath.Join(os.Getenv("ProgramData"), "lf", "icons"),
		filepath.Join(data, "lf", "icons"),
	}

	gFilesPath = filepath.Join(data, "lf", "files")
	gMarksPath = filepath.Join(data, "lf", "marks")
	gTagsPath = filepath.Join(data, "lf", "tags")
	gHistoryPath = filepath.Join(data, "lf", "history")

	socket, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		gDefaultSocketProt = "tcp"
		gDefaultSocketPath = "127.0.0.1:12345"
	} else {
		runtime := os.TempDir()
		gDefaultSocketPath = filepath.Join(runtime, fmt.Sprintf("lf.%s.sock", gUser.Username))
		syscall.Close(socket)
	}
}

func detachedCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &windows.SysProcAttr{CreationFlags: 8}
	return cmd
}

// https://github.com/ActiveState/cli/pull/3389/files
// makeCmdLine builds a command line out of args by escaping "special"
// characters and joining the arguments with spaces.
// Based on syscall\exec_windows.go

// https://learn.microsoft.com/en-us/archive/blogs/twistylittlepassagesallalike/everyone-quotes-command-line-arguments-the-wrong-way

// way around os.EscapeArg is = SysProcAttr.CmdLine
// Set to the misexec application, but don't pass command line arguments
// cmd := exec.Command("msiexec")
//
// // Manually set the command line arguments so they are not escaped
// cmd.SysProcAttr = &syscall.SysProcAttr{
//     HideWindow:    false,
//     CmdLine:       fmt.Sprintf(` /a "%v" TARGETDIR="%v"`, msiFile, targetDir), // Leave a space at the beginning
//     CreationFlags: 0,
// }

// https://github.com/golang/go/issues/15566 syscall: exec_windows.go: arguments should not be escaped to work with msiexec
// argv windows https://daviddeley.com/autohotkey/parameters/parameters.htm#WIN  the best source we have on this topic.

// extensive testing here https://github.com/sergeymakinen/go-quote/blob/main/windows/argv.go

// https://github.com/golang/go/issues/17149 os/exec: Cannot execute command with space in the name on Windows, when there are parameters. Big discussion around cmd

// Go encodes child process parameters in a way that is understood by most programs. Go uses rules similar to what CommandLineToArgvW implements.
//
// Unfortunately, your child process is cmd.exe (cmd.exe is called to execute the batch file you've requested). And cmd.exe parses its input parameters differently.

// manually set cmd:
// command to execute, may contain quotes, backslashes and spaces
// var commandLine = `"C:\Program Files\echo.bat" "hello friend"`
//
// var comSpec = os.Getenv("COMSPEC")
// if comSpec == "" {
// 	comSpec = os.Getenv("SystemRoot") + "\\System32\\cmd.exe"
// }
// childProcess = exec.Command(comSpec)
// childProcess.SysProcAttr = &syscall.SysProcAttr{CmdLine: comSpec + " /C \"" + commandLine + "\""}
//
// // Then execute and read the output
// out, _ := childProcess.CombinedOutput()
// fmt.Printf("Output: %s", out)

// func makeCmd(name string, args ...string) (*exec.Cmd, error) {
// 	if len(args) == 0 {
// 		return exec.Command(name), nil
// 	}
//
// 	name, err := exec.LookPath(name)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if !isBatchFile(name) {
// 		return exec.Command(name, args...), nil
// 	}
//
// 	argsEscaped := make([]string, len(args)+1)
// 	argsEscaped[0] = syscall.EscapeArg(name)
// 	for i, a := range args {
// 		argsEscaped[i+1] = syscall.EscapeArg(a)
// 	}
//
// 	shell := os.Getenv("COMSPEC")
// 	if shell == "" {
// 		shell = "cmd.exe" // Will be expanded by exec.LookPath in exec.Command
// 	}
//
// 	cmd := exec.Command(shell)
// 	cmd.Args = nil
// 	cmd.SysProcAttr = &syscall.SysProcAttr{
// 		CmdLine: fmt.Sprintf(`%s /c "%s"`, syscall.EscapeArg(cmd.Path), strings.Join(argsEscaped, " ")),
// 	}
//
// 	return cmd, nil
// }
//
// func isBatchFile(path string) bool {
// 	ext := filepath.Ext(path)
// 	return strings.EqualFold(ext, ".bat") || strings.EqualFold(ext, ".cmd")
// }

//
// Empirically, I can also confirm that if I force golang to pass nil for lpApplicationName here instead of argv0p, executing exec.Command(`C:\Program Files\echo.bat`, "hello world") works without resorting to using the SysProcAttr escape hatch.
//
// All that said:
//
//     Specifying lpApplicationName for CreateProcess is cited as a security best-practice as a fail-safe should the caller fail to add quotes around the path of the executable in lpCommandLine. FWIW, dotnet ensures that the executable is quoted for this very reason, and it turns out golang does, too, via makeCmdLine's use of appendEscapeArg.
//     I can't find any official documentation around this subtle behavior difference with CreateProcess as to why it works when NOT specifying lpApplicationName, so at this point it's all circumstantial. All I was able to find is someone else pointing out this suble behavioral difference way back in 2001.
//     It's entirely possible I'm glossing over some other important detail. I'm stabbing in the dark at this point. 😅

// It still exists in golang 1.20, e.g:
// cmd.exe /c copy/b "input 1.ts"+"input 2.ts" ouput.ts
// I guess that golang can automatically add double quotes to paths that contain Spaces, but many commands have their own coding rules, such as copy/b. The "+" in copy/b means concatenating the input files, but golang cannot parse it and cannot add double quotes to paths of input files that contain Spaces correctly.
//

// TODO: check this https://github.com/otm/gluash

// https://github.com/golang/go/issues/68313 syscall/exec_windows.go: appendEscapeArg does not escape all necessary characters
// some related commits to go https://go-review.googlesource.com/c/go/+/30947

// https://github.com/golang/go/issues/27199 os/exec: execution of batch-files (.cmd/.bat) is vulnerable in go-lang for windows / insufficient escape

// some workarounds = https://github.com/ActiveState/cli/pull/3389 adn functions for proper escapings

// golang lib https://github.com/golang/go/blob/master/src/syscall/exec_windows.go#L84

// Based on https://github.com/sebres/PoC/blob/master/SB-0D-001-win-exec/SOLUTION.md#definition

// TODO:  syscall\exec_windows.go
// and https://github.com/ActiveState/cli/pull/3389/files
// and https://learn.microsoft.com/en-us/archive/blogs/twistylittlepassagesallalike/everyone-quotes-command-line-arguments-the-wrong-way
func escapeArg(s string) string {
	const argvUnsafeChars = "\t \"<>&|^!()%"
	if len(s) == 0 {
		return `""`
	}

	needsBackslash := strings.ContainsAny(s, `"\`)
	// Based on https://github.com/sebres/PoC/blob/master/SB-0D-001-win-exec/SOLUTION.md#definition
	//"\t \""
	needsQuotes := strings.ContainsAny(s, "\t \"<>&|^!()%")

	if !needsBackslash && !needsQuotes {
		// No special handling required; normal case.
		return s
	}

	if !needsBackslash {
		// just needsQuotes
		return `"` + s + `"`
	}

	var (
		buf     strings.Builder
		slashes int
	)
	if needsQuotes {
		buf.WriteByte('"')
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			for slashes++; slashes > 0; slashes-- {
				buf.WriteByte('\\')
			}
			buf.WriteByte(s[i])
		case '\\':
			slashes++
			buf.WriteByte(s[i])
		default:
			slashes = 0
			buf.WriteByte(s[i])
		}
	}

	if needsQuotes {
		for ; slashes > 0; slashes-- {
			buf.WriteByte('\\')
		}
		buf.WriteByte('"')
	}

	return buf.String()
}

// https://ss64.com/nt/syntax-esc.html
var cmdQuoteReplacer = strings.NewReplacer(
	"\t", "^\t",
	// TODO: do we need to escape Spaces???
	" ", "^ ",
	"!", "^!",
	"&", "^&",
	"'", "^'",
	`"`, `""`,
	"+", "^+",
	",", "^,",
	";", "^;",
	"<", "^<",
	"|", "^|",
	"=", "^=",
	">", "^>",
	"[", "^[",
	"]", "^]",
	"^", "^^",
	"`", "^`",
	"{", "^{",
	"}", "^}",
	"~", "^~",
)

func escapeCmd(s string) string {
	const cmdUnsafeChars = "!\"&'+,;<=>[]^`{}~"
	if strings.ContainsAny(s, cmdUnsafeChars) {
		s = cmdQuoteReplacer.Replace(s)
	}
	return s
}

// TODO: I we ran a command from UI it passes as entire string in `s`
// without args. Should we also escape `s`??

func shellCommand2(s string, args []string) *exec.Cmd {
	for i := 0; i < len(args); i++ {
		args[i] = escapeArg(args[i])
	}

	// s = escapeArg(s)

	// Windows CMD requires special handling to deal with quoted arguments
	exeName := filepath.Base(strings.ToLower(gOpts.shellcmd[0]))
	isCmd := exeName == "cmd" || strings.HasSuffix(exeName, ".bat") || strings.HasSuffix(exeName, ".cmd")
	if isCmd {
		// Go currently does not escape arguments properly on Windows, it account for spaces and tab characters, but not
		// other characters that need escaping such as `<` and `>`.
		// This can be dropped once we update to a Go version that fixes this bug: https://github.com/golang/go/issues/68313
		for i := 0; i < len(args); i++ {
			args[i] = escapeCmd(args[i])
		}
		
		// in case the command has some malicious characters such as &
		// TODO: or it should be handled by the user when they type the command from the UI??
		s = `"`+escapeCmd(s)+`"`
	}

	var words []string
	for _, word := range gOpts.shellcmd {
		switch word {
		case "%c":
			words = append(words, s)
		case "%a":
			words = append(words, args...)
		default:
			words = append(words, word)
		}
	}

	cmd := exec.Command(words[0], words[1:]...)

	if true {
		// If we have to deal with a different from Argv command-line quoting
		// when starting processes on Windows, we need to to manually create a command-line
		// via the CmdLine SysProcAttr

		cmd.SysProcAttr = &windows.SysProcAttr{
			CmdLine: strings.Join(words[1:], " "),
		}

		log.Printf("- %v\n", cmd.SysProcAttr)
		return cmd
	}
	log.Printf("- %v\n", cmd.Args)
	return cmd
}

func shellCommand(s string, args []string) *exec.Cmd {
	if len(gOpts.shellcmd) > 0 {
		return shellCommand2(s, args)
	}

	// original legacy configuration which uses shell, shellopts and shellflag

	// Windows CMD requires special handling to deal with quoted arguments
	if strings.ToLower(gOpts.shell) == "cmd" {
		var builder strings.Builder
		builder.WriteString(s)
		for _, arg := range args {
			fmt.Fprintf(&builder, ` "%s"`, arg)
		}
		shellOpts := strings.Join(gOpts.shellopts, " ")
		cmdline := fmt.Sprintf(`%s %s %s "%s"`, gOpts.shell, shellOpts, gOpts.shellflag, builder.String())

		cmd := exec.Command(gOpts.shell)
		cmd.SysProcAttr = &windows.SysProcAttr{CmdLine: cmdline}

		return cmd
	}

	args = append([]string{gOpts.shellflag, s}, args...)
	args = append(gOpts.shellopts, args...)
	cmd := exec.Command(gOpts.shell, args...)

	return cmd
}

func shellSetPG(cmd *exec.Cmd) {
}

func shellKill(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}

func setDefaults() {
	gOpts.cmds["open"] = &execExpr{"&", "%OPENER% %f%"}
	gOpts.keys["e"] = &execExpr{"$", "%EDITOR% %f%"}
	gOpts.keys["i"] = &execExpr{"!", "%PAGER% %f%"}
	gOpts.keys["w"] = &execExpr{"$", "%SHELL%"}

	gOpts.cmds["doc"] = &execExpr{"!", "%lf% -doc | %PAGER%"}
	gOpts.keys["<f-1>"] = &callExpr{"doc", nil, 1}

	gOpts.cmds["maps"] = &execExpr{"!", `%lf% -remote "query %id% maps" | %PAGER%`}
	gOpts.cmds["cmaps"] = &execExpr{"!", `%lf% -remote "query %id% cmaps" | %PAGER%`}
	gOpts.cmds["cmds"] = &execExpr{"!", `%lf% -remote "query %id% cmds" | %PAGER%`}
}

func setUserUmask() {}

func isExecutable(f os.FileInfo) bool {
	exts := strings.Split(envPathExt, string(filepath.ListSeparator))
	for _, e := range exts {
		if strings.HasSuffix(strings.ToLower(f.Name()), strings.ToLower(e)) {
			log.Print(f.Name(), e)
			return true
		}
	}
	return false
}

func isHidden(f os.FileInfo, path string, hiddenfiles []string) bool {
	ptr, err := windows.UTF16PtrFromString(filepath.Join(path, f.Name()))
	if err != nil {
		return false
	}
	attrs, err := windows.GetFileAttributes(ptr)
	if err != nil {
		return false
	}

	if attrs&windows.FILE_ATTRIBUTE_HIDDEN != 0 {
		return true
	}

	hidden := false
	for _, pattern := range hiddenfiles {
		matched := matchPattern(strings.TrimPrefix(pattern, "!"), f.Name(), path)
		if strings.HasPrefix(pattern, "!") && matched {
			hidden = false
		} else if matched {
			hidden = true
		}
	}

	return hidden
}

func userName(f os.FileInfo) string {
	return ""
}

func groupName(f os.FileInfo) string {
	return ""
}

func linkCount(f os.FileInfo) string {
	return ""
}

func errCrossDevice(err error) bool {
	return err.(*os.LinkError).Err.(windows.Errno) == 17
}

func quoteString(s string) string {
	// Windows CMD requires special handling to deal with quoted arguments
	if strings.ToLower(gOpts.shell) == "cmd" {
		return fmt.Sprintf(`"%s"`, s)
	}
	return s
}
