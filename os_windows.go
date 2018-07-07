package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

var (
	envOpener = os.Getenv("OPENER")
	envEditor = os.Getenv("EDITOR")
	envPager  = os.Getenv("PAGER")
	envShell  = os.Getenv("SHELL")
)

var (
	envPathExt = os.Getenv("PATHEXT")
)

var (
	gDefaultShell      = "cmd"
	gDefaultSocketProt = "tcp"
	gDefaultSocketPath = ":12345"
)

var (
	gUser        *user.User
	gConfigPaths []string
)

func init() {
	if envOpener == "" {
		envOpener = "start"
	}

	if envEditor == "" {
		envEditor = "notepad"
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

	gConfigPaths = []string{
		filepath.Join(os.Getenv("ProgramData"), "lf", "lfrc"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "lf", "lfrc"),
	}
}

func pauseCommand() *exec.Cmd {
	return exec.Command("cmd", "/c", "pause")
}

func shellCommand(s string, args []string) *exec.Cmd {
	args = append([]string{"/c", s}, args...)

	args = append(gOpts.shellopts, args...)

	return exec.Command(gOpts.shell, args...)
}

func pasteCommand(list []string, dir *dir, cp bool) *exec.Cmd {
	var args []string

	sh := "robocopy"
	if !cp {
		args = []string{"/move"}
	}
	for _, f := range list {
		stat, err := os.Stat(f)
		if err != nil {
			log.Printf("getting file information: %s", err)
			continue
		}
		base := filepath.Base(f)
		dest := filepath.Dir(f)
		if stat.IsDir() {
			exec.Command(sh, append(args, f, filepath.Join(dir.path, base))...).Run()
		} else {
			exec.Command(sh, append(args, dest, dir.path, base)...).Run()
		}
	}

	// TODO: return 0 on success

	return exec.Command(sh, args...)
}

func setDefaults() {
	gOpts.cmds["open"] = &execExpr{"&", "%OPENER% %f%"}
	gOpts.keys["e"] = &execExpr{"$", "%EDITOR% %f%"}
	gOpts.keys["i"] = &execExpr{"!", "%PAGER% %f%"}
	gOpts.keys["w"] = &execExpr{"$", "%SHELL%"}

	gOpts.cmds["doc"] = &execExpr{"!", "lf -doc | $PAGER"}
	gOpts.keys["<f-1>"] = &callExpr{"doc", nil, 1}
}

func moveCursor(y, x int) {
	// TODO: implement
	return
}

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
