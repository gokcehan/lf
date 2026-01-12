package main

import (
	"cmp"
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
		envEditor = cmp.Or(os.Getenv("EDITOR"), "notepad")
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

	config := cmp.Or(
		os.Getenv("LF_CONFIG_HOME"),
		os.Getenv("XDG_CONFIG_HOME"),
		os.Getenv("APPDATA"),
	)

	gConfigPaths = []string{
		filepath.Join(os.Getenv("ProgramData"), "lf", "lfrc"),
		filepath.Join(config, "lf", "lfrc"),
	}

	gColorsPaths = []string{
		filepath.Join(os.Getenv("ProgramData"), "lf", "colors"),
		filepath.Join(config, "lf", "colors"),
	}

	gIconsPaths = []string{
		filepath.Join(os.Getenv("ProgramData"), "lf", "icons"),
		filepath.Join(config, "lf", "icons"),
	}

	data := cmp.Or(
		os.Getenv("LF_DATA_HOME"),
		os.Getenv("XDG_DATA_HOME"),
		os.Getenv("LOCALAPPDATA"),
	)

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

func shellCommand(s string, args []string) *exec.Cmd {
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
	return exec.Command(gOpts.shell, args...)
}

func shellSetPG(_ *exec.Cmd) {
}

func shellKill(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}

func setDefaults() {
	gOpts.cmds["open"] = &execExpr{"&", "%OPENER% %f%"}
	gOpts.nkeys["e"] = &execExpr{"$", "%EDITOR% %f%"}
	gOpts.vkeys["e"] = &execExpr{"$", "%EDITOR% %f%"}
	gOpts.nkeys["i"] = &execExpr{"!", "%PAGER% %f%"}
	gOpts.vkeys["i"] = &execExpr{"!", "%PAGER% %f%"}
	gOpts.nkeys["w"] = &execExpr{"$", "%SHELL%"}
	gOpts.vkeys["w"] = &execExpr{"$", "%SHELL%"}

	gOpts.cmds["help"] = &execExpr{"!", "%lf% -doc | %PAGER%"}
	gOpts.nkeys["<f-1>"] = &callExpr{"help", nil, 1}
	gOpts.vkeys["<f-1>"] = &callExpr{"help", nil, 1}

	gOpts.cmds["maps"] = &execExpr{"!", `%lf% -remote "query %id% maps" | %PAGER%`}
	gOpts.cmds["nmaps"] = &execExpr{"!", `%lf% -remote "query %id% nmaps" | %PAGER%`}
	gOpts.cmds["vmaps"] = &execExpr{"!", `%lf% -remote "query %id% vmaps" | %PAGER%`}
	gOpts.cmds["cmaps"] = &execExpr{"!", `%lf% -remote "query %id% cmaps" | %PAGER%`}
	gOpts.cmds["cmds"] = &execExpr{"!", `%lf% -remote "query %id% cmds" | %PAGER%`}
}

func setUserUmask() {}

func isExecutable(f os.FileInfo) bool {
	for e := range strings.SplitSeq(envPathExt, string(filepath.ListSeparator)) {
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
		if matchPattern(strings.TrimPrefix(pattern, "!"), f.Name(), path) {
			hidden = !strings.HasPrefix(pattern, "!")
		}
	}

	return hidden
}

func userName(_ os.FileInfo) string {
	return ""
}

func groupName(_ os.FileInfo) string {
	return ""
}

func linkCount(_ os.FileInfo) string {
	return ""
}

func errCrossDevice(err error) bool {
	return err.(*os.LinkError).Err.(windows.Errno) == windows.ERROR_NOT_SAME_DEVICE
}

func quoteString(s string) string {
	// Windows CMD requires special handling to deal with quoted arguments
	if strings.ToLower(gOpts.shell) == "cmd" {
		return fmt.Sprintf(`"%s"`, s)
	}
	return s
}

func shellEscape(s string) string {
	for _, r := range s {
		if strings.ContainsRune(" !%&'()+,;=[]^`{}~", r) {
			return fmt.Sprintf(`"%s"`, s)
		}
	}
	return s
}

func shellUnescape(s string) string {
	return strings.ReplaceAll(s, `"`, "")
}
