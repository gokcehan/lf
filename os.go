//go:build !windows

package main

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

var (
	envOpener = os.Getenv("OPENER")
	envEditor = os.Getenv("VISUAL")
	envPager  = os.Getenv("PAGER")
	envShell  = os.Getenv("SHELL")
)

var (
	gDefaultShell       = "sh"
	gDefaultShellFlag   = "-c"
	gDefaultSocketProt  = "unix"
	gDefaultSocketPath  string
	gDefaultHiddenFiles = []string{".*"}
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
		envEditor = cmp.Or(os.Getenv("EDITOR"), "vi")
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

	config := cmp.Or(
		os.Getenv("LF_CONFIG_HOME"),
		os.Getenv("XDG_CONFIG_HOME"),
		filepath.Join(gUser.HomeDir, ".config"),
	)

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

	data := cmp.Or(
		os.Getenv("LF_DATA_HOME"),
		os.Getenv("XDG_DATA_HOME"),
		filepath.Join(gUser.HomeDir, ".local", "share"),
	)

	gFilesPath = filepath.Join(data, "lf", "files")
	gMarksPath = filepath.Join(data, "lf", "marks")
	gTagsPath = filepath.Join(data, "lf", "tags")
	gHistoryPath = filepath.Join(data, "lf", "history")

	runtime := cmp.Or(os.Getenv("XDG_RUNTIME_DIR"), os.TempDir())

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
		err = unix.Kill(-pgid, syscall.SIGTERM)
		if err == nil {
			return nil
		}
	}
	return cmd.Process.Kill()
}

func setDefaults() {
	gOpts.cmds["open"] = &execExpr{"&", `$OPENER "$f"`}
	gOpts.nkeys["e"] = &execExpr{"$", `$EDITOR "$f"`}
	gOpts.vkeys["e"] = &execExpr{"$", `$EDITOR "$f"`}
	gOpts.nkeys["i"] = &execExpr{"$", `$PAGER "$f"`}
	gOpts.vkeys["i"] = &execExpr{"$", `$PAGER "$f"`}
	gOpts.nkeys["w"] = &execExpr{"$", "$SHELL"}
	gOpts.vkeys["w"] = &execExpr{"$", "$SHELL"}

	gOpts.cmds["doc"] = &execExpr{"$", `"$lf" -doc | $PAGER`}
	gOpts.nkeys["<f-1>"] = &callExpr{"doc", nil, 1}
	gOpts.vkeys["<f-1>"] = &callExpr{"doc", nil, 1}

	gOpts.cmds["maps"] = &execExpr{"$", `"$lf" -remote "query $id maps" | $PAGER`}
	gOpts.cmds["nmaps"] = &execExpr{"$", `"$lf" -remote "query $id nmaps" | $PAGER`}
	gOpts.cmds["vmaps"] = &execExpr{"$", `"$lf" -remote "query $id vmaps" | $PAGER`}
	gOpts.cmds["cmaps"] = &execExpr{"$", `"$lf" -remote "query $id cmaps" | $PAGER`}
	gOpts.cmds["cmds"] = &execExpr{"$", `"$lf" -remote "query $id cmds" | $PAGER`}
}

func setUserUmask() {
	unix.Umask(0o077)
}

func isExecutable(f os.FileInfo) bool {
	return f.Mode()&0o111 != 0
}

func isHidden(f os.FileInfo, path string, hiddenfiles []string) bool {
	hidden := false
	for _, pattern := range hiddenfiles {
		if matchPattern(strings.TrimPrefix(pattern, "!"), f.Name(), path) {
			hidden = !strings.HasPrefix(pattern, "!")
		}
	}
	return hidden
}

func userName(f os.FileInfo) string {
	if stat, ok := f.Sys().(*syscall.Stat_t); ok {
		uid := strconv.FormatUint(uint64(stat.Uid), 10)
		if u, err := user.LookupId(uid); err == nil {
			return u.Username
		}
		return uid
	}
	return ""
}

func groupName(f os.FileInfo) string {
	if stat, ok := f.Sys().(*syscall.Stat_t); ok {
		gid := strconv.FormatUint(uint64(stat.Gid), 10)
		if g, err := user.LookupGroupId(gid); err == nil {
			return g.Name
		}
		return gid
	}
	return ""
}

func linkCount(f os.FileInfo) string {
	if stat, ok := f.Sys().(*syscall.Stat_t); ok {
		return strconv.FormatUint(uint64(stat.Nlink), 10)
	}
	return ""
}

func errCrossDevice(err error) bool {
	return err.(*os.LinkError).Err.(unix.Errno) == unix.EXDEV
}

func quoteString(s string) string {
	return s
}

func shellEscape(s string) string {
	buf := make([]rune, 0, len(s))
	for _, r := range s {
		if strings.ContainsRune(" !\"$&'()*,:;<=>?@[\\]^`{|}", r) {
			buf = append(buf, '\\')
		}
		buf = append(buf, r)
	}
	return string(buf)
}

func shellUnescape(s string) string {
	esc := false
	buf := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\\' && !esc {
			esc = true
			continue
		}
		esc = false
		buf = append(buf, r)
	}
	if esc {
		buf = append(buf, '\\')
	}
	return string(buf)
}
