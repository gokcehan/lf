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
	gDefaultShell      = "cmd"
	gDefaultSocketProt = "tcp"
	gDefaultSocketPath = ":12345"
)

var (
	gUser       *user.User
	gConfigPath string
)

func init() {
	u, err := user.Current()
	if err != nil {
		log.Printf("user: %s", err)
	}
	gUser = u

	// remove domain prefix
	gUser.Username = strings.Split(gUser.Username, `\`)[1]

	gConfigPath = filepath.Join(gUser.HomeDir, "AppData", "Local", "lf", "lfrc")
}

func pauseCommand() *exec.Cmd {
	return exec.Command("cmd", "/c", "pause")
}

func shellCommand(s string, args []string) *exec.Cmd {
	args = append([]string{"/c", s}, args...)
	return exec.Command(gOpts.shell, args...)
}

func putCommand(list []string, dir *dir, copy bool) *exec.Cmd {
	var args []string

	sh := "robocopy"
	if !copy {
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
	gOpts.keys["e"] = &execExpr{"$", `notepad %f%`}
	gOpts.keys["i"] = &execExpr{"$", `more %f%`}
	gOpts.keys["w"] = &execExpr{"$", "cmd"}
	gOpts.cmds["open-file"] = &execExpr{"&", `start %f%`}
}

func moveCursor(y, x int) {
	// TODO: implement
	return
}
