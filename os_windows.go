package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

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
	gDefaultShell      = "cmd"
	gDefaultShellFlag  = "/c"
	gDefaultSocketProt = "tcp"
	gDefaultSocketPath = "127.0.0.1:12345"
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

	u, err := user.Current()
	if err != nil {
		log.Printf("user: %s", err)
	}
	gUser = u

	// remove domain prefix
	gUser.Username = strings.Split(gUser.Username, `\`)[1]

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
	return attrs&windows.FILE_ATTRIBUTE_HIDDEN != 0
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
