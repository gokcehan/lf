// +build !windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

var (
	gDefaultShell      = "/bin/sh"
	gDefaultSocketProt = "unix"
	gDefaultSocketPath string
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

	config := os.Getenv("XDG_CONFIG_HOME")
	if config == "" {
		config = filepath.Join(gUser.HomeDir, ".config")
	}

	gConfigPath = filepath.Join(config, "lf", "lfrc")

	gDefaultSocketPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.sock", gUser.Username))
}

func pauseCommand() *exec.Cmd {
	c := `echo
	      echo -n 'Press any key to continue'
	      old=$(stty -g)
	      stty raw -echo
	      eval "ignore=\$(dd bs=1 count=1 2> /dev/null)"
	      stty $old
	      echo`

	return exec.Command(gOpts.shell, "-c", c)
}

func shellCommand(s string, args []string) *exec.Cmd {
	if len(gOpts.ifs) != 0 {
		s = fmt.Sprintf("IFS='%s'; %s", gOpts.ifs, s)
	}

	args = append([]string{"-c", s, "--"}, args...)

	return exec.Command(gOpts.shell, args...)
}

func putCommand(list []string, dir *dir, copy bool) *exec.Cmd {
	var sh string
	var args []string

	if copy {
		sh = "cp"
		args = append(args, "-R")
	} else {
		sh = "mv"
	}

	// XXX: POSIX standard states that -i flag shall do nothing when the
	// response is not affirmative. Since this command is run with a nil stdin,
	// it should not give an affirmative answer and in return this command
	// should not overwrite existing files. Our intention here is to use the
	// standard -i flag in place of non-standard -n flag to avoid overwrites.
	args = append(args, "-i")

	args = append(args, list...)
	args = append(args, dir.path)

	return exec.Command(sh, args...)
}

func moveCursor(y, x int) {
	fmt.Printf("\033[%d;%dH", y, x)
}
