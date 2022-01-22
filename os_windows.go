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
	gDefaultShellFlag  = "/c"
	gDefaultSocketProt = "tcp"
	gDefaultSocketPath = "127.0.0.1:12345"
)

var (
	gUser        *user.User
	gConfigPaths []string
	gFilesPath   string
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

	gFilesPath = filepath.Join(data, "lf", "files")
	gMarksPath = filepath.Join(data, "lf", "marks")
	gHistoryPath = filepath.Join(data, "lf", "history")
}

func detachedCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 8}
	return cmd
}

func shellCommand(s string, args []string) *exec.Cmd {
	args = append([]string{gOpts.shellflag, s}, args...)

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
	ptr, err := syscall.UTF16PtrFromString(filepath.Join(path, f.Name()))
	if err != nil {
		return false
	}
	attrs, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		return false
	}
	return attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0
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
	return err.(*os.LinkError).Err.(syscall.Errno) == 17
}

func exportFiles(f string, fs []string, pwd string) {
	envFile := fmt.Sprintf(`"%s"`, f)

	var quotedFiles []string
	for _, f := range fs {
		quotedFiles = append(quotedFiles, fmt.Sprintf(`"%s"`, f))
	}
	envFiles := strings.Join(quotedFiles, gOpts.filesep)

	os.Setenv("f", envFile)
	os.Setenv("fs", envFiles)
	os.Setenv("PWD", pwd)

	if len(fs) == 0 {
		os.Setenv("fx", envFile)
	} else {
		os.Setenv("fx", envFiles)
	}
}

func getWinDrives() []string {
	// https://stackoverflow.com/a/23135463/8608146

	kernel32, _ := syscall.LoadLibrary("kernel32.dll")
	getLogicalDrivesHandle, _ := syscall.GetProcAddress(kernel32, "GetLogicalDrives")

	var drives []string

	if ret, _, callErr := syscall.Syscall(uintptr(getLogicalDrivesHandle), 0, 0, 0, 0); callErr != 0 {
		// handle error
		panic(callErr)
	} else {
		bitsToDrives := func(bitMap uint32) (drives []string) {
			availableDrives := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
			for i := range availableDrives {
				if bitMap&1 == 1 {
					drives = append(drives, availableDrives[i])
				}
				bitMap >>= 1
			}
			return
		}
		drives = bitsToDrives(uint32(ret))
	}

	return drives
}
