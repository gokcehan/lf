//go:build !windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

var (
	envOpener = os.Getenv("OPENER")
	envEditor = os.Getenv("EDITOR")
	envPager  = os.Getenv("PAGER")
	envShell  = os.Getenv("SHELL")
)

var (
	gDefaultShell      = "sh"
	gDefaultShellFlag  = "-c"
	gDefaultSocketProt = "unix"
	gDefaultSocketPath string
)

var (
	gUser        *user.User
	gConfigPaths []string
	gColorsPaths []string
	gIconsPaths  []string
	gFilesPath   string
	gMarksPath   string
	gTagsPath    string
	gHistoryPath string
)

func init() {
	if envOpener == "" {
		if runtime.GOOS == "darwin" {
			envOpener = "open"
		} else {
			envOpener = "xdg-open"
		}
	}

	if envEditor == "" {
		envEditor = "vi"
	}

	if envPager == "" {
		envPager = "less"
	}

	if envShell == "" {
		envShell = "sh"
	}

	u, err := user.Current()
	if err != nil {
		// When the user is not in /etc/passwd (for e.g. LDAP) and CGO_ENABLED=1 in go env,
		// the cgo implementation of user.Current() fails even when HOME and USER are set.

		log.Printf("user: %s", err)
		if os.Getenv("HOME") == "" {
			panic("$HOME variable is empty or not set")
		}
		if os.Getenv("USER") == "" {
			panic("$USER variable is empty or not set")
		}
		u = &user.User{
			Username: os.Getenv("USER"),
			HomeDir:  os.Getenv("HOME"),
		}
	}
	gUser = u

	config := os.Getenv("XDG_CONFIG_HOME")
	if config == "" {
		config = filepath.Join(gUser.HomeDir, ".config")
	}

	gConfigPaths = []string{
		filepath.Join("/etc", "lf", "lfrc"),
		filepath.Join(config, "lf", "lfrc"),
	}

	gColorsPaths = []string{
		filepath.Join("/etc", "lf", "colors"),
		filepath.Join(config, "lf", "colors"),
	}

	gIconsPaths = []string{
		filepath.Join("/etc", "lf", "icons"),
		filepath.Join(config, "lf", "icons"),
	}

	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		data = filepath.Join(gUser.HomeDir, ".local", "share")
	}

	gFilesPath = filepath.Join(data, "lf", "files")
	gMarksPath = filepath.Join(data, "lf", "marks")
	gTagsPath = filepath.Join(data, "lf", "tags")
	gHistoryPath = filepath.Join(data, "lf", "history")

	runtime := os.Getenv("XDG_RUNTIME_DIR")
	if runtime == "" {
		runtime = os.TempDir()
	}

	gDefaultSocketPath = filepath.Join(runtime, fmt.Sprintf("lf.%s.sock", gUser.Username))
}

func detachedCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &unix.SysProcAttr{Setsid: true}
	return cmd
}

func shellCommand(s string, args []string) *exec.Cmd {
	if len(gOpts.ifs) != 0 {
		s = fmt.Sprintf("IFS='%s'; %s", gOpts.ifs, s)
	}

	args = append([]string{gOpts.shellflag, s, "--"}, args...)

	args = append(gOpts.shellopts, args...)

	return exec.Command(gOpts.shell, args...)
}

func shellSetPG(cmd *exec.Cmd) {
	cmd.SysProcAttr = &unix.SysProcAttr{Setpgid: true}
}

func shellKill(cmd *exec.Cmd) error {
	pgid, err := unix.Getpgid(cmd.Process.Pid)
	if err == nil && cmd.Process.Pid == pgid {
		// kill the process group
		err = unix.Kill(-pgid, 15)
		if err == nil {
			return nil
		}
	}
	return cmd.Process.Kill()
}

func pageFileCommand(filename string) string {
	return fmt.Sprintf(`$PAGER "%s"`, filename)
}

func setDefaults() {
	gOpts.cmds["open"] = &execExpr{"&", `$OPENER "$f"`}
	gOpts.keys["e"] = &execExpr{"$", `$EDITOR "$f"`}
	gOpts.keys["i"] = &execExpr{"$", pageFileCommand(`$f`)}
	gOpts.keys["w"] = &execExpr{"$", "$SHELL"}

	gOpts.cmds["doc"] = &execExpr{"$", "lf -doc | $PAGER"}
	gOpts.keys["<f-1>"] = &callExpr{"doc", nil, 1}
}

func setUserUmask() {
	unix.Umask(0077)
}

func isExecutable(f os.FileInfo) bool {
	return f.Mode()&0111 != 0
}

func isHidden(f os.FileInfo, path string, hiddenfiles []string) bool {
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
	if stat, ok := f.Sys().(*syscall.Stat_t); ok {
		if u, err := user.LookupId(fmt.Sprint(stat.Uid)); err == nil {
			return fmt.Sprintf("%v ", u.Username)
		}
	}
	return ""
}

func groupName(f os.FileInfo) string {
	if stat, ok := f.Sys().(*syscall.Stat_t); ok {
		if g, err := user.LookupGroupId(fmt.Sprint(stat.Gid)); err == nil {
			return fmt.Sprintf("%v ", g.Name)
		}
	}
	return ""
}

func linkCount(f os.FileInfo) string {
	if stat, ok := f.Sys().(*syscall.Stat_t); ok {
		return fmt.Sprintf("%v ", stat.Nlink)
	}
	return ""
}

func errCrossDevice(err error) bool {
	return err.(*os.LinkError).Err.(unix.Errno) == unix.EXDEV
}

func exportFiles(f string, fs []string, pwd string) {
	envFile := f
	envFiles := strings.Join(fs, gOpts.filesep)

	os.Setenv("f", envFile)
	os.Setenv("fs", envFiles)
	os.Setenv("PWD", pwd)

	if len(fs) == 0 {
		os.Setenv("fx", envFile)
	} else {
		os.Setenv("fx", envFiles)
	}
}
