// +build !windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

var (
	envEditor = os.Getenv("EDITOR")
	envPager  = os.Getenv("PAGER")
	envShell  = os.Getenv("SHELL")
)

var (
	gDefaultShell      = "sh"
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
	cmd := `echo
	        echo -n 'Press any key to continue'
	        old=$(stty -g)
	        stty raw -echo
	        eval "ignore=\$(dd bs=1 count=1 2> /dev/null)"
	        stty $old
	        echo`

	return exec.Command(gOpts.shell, "-c", cmd)
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
	// response is not affirmative. Since this command is run with a nil
	// stdin, it should not give an affirmative answer and in return this
	// command should not overwrite existing files. Our intention here is
	// to use the standard -i flag in place of non-standard -n flag to
	// avoid overwrites.
	args = append(args, "-i")

	args = append(args, list...)
	args = append(args, dir.path)

	return exec.Command(sh, args...)
}

func setDefaults() {
	if envEditor == "" {
		gOpts.keys["e"] = &execExpr{"$", `vi "$f"`}
	} else {
		gOpts.keys["e"] = &execExpr{"$", envEditor + ` "$f"`}
	}

	if envPager == "" {
		gOpts.keys["i"] = &execExpr{"$", `less "$f"`}
	} else {
		gOpts.keys["i"] = &execExpr{"$", envPager + ` "$f"`}
	}

	if envShell == "" {
		gOpts.keys["w"] = &execExpr{"$", "sh"}
	} else {
		gOpts.keys["w"] = &execExpr{"$", envShell}
	}

	if runtime.GOOS == "darwin" {
		gOpts.cmds["open-file"] = &execExpr{"&", `open "$f"`}
	} else {
		gOpts.cmds["open-file"] = &execExpr{"&", `xdg-open "$f"`}
	}
}

func moveCursor(y, x int) {
	fmt.Printf("\033[%d;%dH", y, x)
}
