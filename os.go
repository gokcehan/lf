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
	envOpener = os.Getenv("OPENER")
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
	gUser        *user.User
	gConfigPaths []string
	gMarksPath   string
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
		log.Printf("user: %s", err)
		if os.Getenv("HOME") == "" {
			log.Print("$HOME variable is empty or not set")
		}
		if os.Getenv("USER") == "" {
			log.Print("$USER variable is empty or not set")
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

	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		data = filepath.Join(gUser.HomeDir, ".local", "share")
	}

	gMarksPath = filepath.Join(data, "lf", "marks")
	gHistoryPath = filepath.Join(data, "lf", "history")

	gDefaultSocketPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.sock", gUser.Username))
}

func detachedCommand(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
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

	args = append(gOpts.shellopts, args...)

	return exec.Command(gOpts.shell, args...)
}

func pasteCommand(list []string, dir *dir, cp bool) *exec.Cmd {
	var sh string
	var args []string

	if cp {
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

func deleteCommand(list []string) *exec.Cmd {
	var sh string
	var args []string

	sh = "rm"

	args = append(args, "-r")
	args = append(args, list...)

	return exec.Command(sh, args...)
}

func setDefaults() {
	gOpts.cmds["open"] = &execExpr{"&", `$OPENER "$f"`}
	gOpts.keys["e"] = &execExpr{"$", `$EDITOR "$f"`}
	gOpts.keys["i"] = &execExpr{"$", `$PAGER "$f"`}
	gOpts.keys["w"] = &execExpr{"$", "$SHELL"}

	gOpts.cmds["doc"] = &execExpr{"$", "lf -doc | $PAGER"}
	gOpts.keys["<f-1>"] = &callExpr{"doc", nil, 1}
}

func moveCursor(y, x int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

func isExecutable(f os.FileInfo) bool {
	return f.Mode()&0111 != 0
}
