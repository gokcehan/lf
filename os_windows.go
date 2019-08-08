package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	envOpener = os.Getenv("OPENER")
	envEditor = os.Getenv("EDITOR")
	envPager  = os.Getenv("PAGER")
	envShell  = os.Getenv("SHELL")
)

var envPathExt = os.Getenv("PATHEXT")

var (
	gDefaultShell      = "cmd"
	gDefaultSocketProt = "tcp"
	gDefaultSocketPath = ":12345"
)

var (
	gUser        *user.User
	gConfigPaths []string
	gMarksPath   string
	gHistoryPath string
)

func init() {
	if envOpener == "" {
		envOpener = `start ""`
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

	data := os.Getenv("LOCALAPPDATA")

	gConfigPaths = []string{
		filepath.Join(os.Getenv("ProgramData"), "lf", "lfrc"),
		filepath.Join(data, "lf", "lfrc"),
	}

	gMarksPath = filepath.Join(data, "lf", "marks")
	gHistoryPath = filepath.Join(data, "lf", "history")
}

func detachedCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 8}
	return cmd
}

func pauseCommand() *exec.Cmd {
	return exec.Command("cmd", "/c", "pause")
}

func shellCommand(s string, args []string) *exec.Cmd {
	args = append([]string{"/c", s}, args...)

	args = append(gOpts.shellopts, args...)

	return exec.Command(gOpts.shell, args...)
}

func setDefaults() {
	gOpts.cmds["open"] = &execExpr{"&", "%OPENER% %f%"}
	gOpts.keys["e"] = &execExpr{"$", "%EDITOR% %f%"}
	gOpts.keys["i"] = &execExpr{"!", "%PAGER% %f%"}
	gOpts.keys["w"] = &execExpr{"$", "%SHELL%"}

	gOpts.cmds["doc"] = &execExpr{"!", "lf -doc | %PAGER%"}
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

func isHidden(f os.FileInfo) bool {
	// TODO: implement
	return false
}

func exportFiles(f string, fs []string) {
	envFile := fmt.Sprintf(`"%s"`, f)

	var quotedFiles []string
	for _, f := range fs {
		quotedFiles = append(quotedFiles, fmt.Sprintf(`"%s"`, f))
	}
	envFiles := strings.Join(quotedFiles, gOpts.filesep)

	os.Setenv("f", envFile)
	os.Setenv("fs", envFiles)

	if len(fs) == 0 {
		os.Setenv("fx", envFile)
	} else {
		os.Setenv("fx", envFiles)
	}
}
