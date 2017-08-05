package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var gDefaultShell = "cmd"
var gDefaultSocketProt = "tcp"
var gDefaultSocketPath = ":12345"

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

	// TODO: add backup options
	// TODO: return 0 on success

	return exec.Command(sh, args...)
}

func moveCursor(y, x int) {
	// TODO: implement
	return
}
