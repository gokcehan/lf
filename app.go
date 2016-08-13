package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type App struct {
	ui  *UI
	nav *Nav
}

func waitKey() error {
	// TODO: this should be done with termbox somehow

	cmd := exec.Command(envShell, "-c", "echo; echo -n 'Press any key to continue'; stty -echo; read -n 1; stty echo; echo")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("waiting key: %s", err)
	}

	return nil
}

func (app *App) handleInp() {
	for {
		if gExitFlag {
			log.Print("bye!")
			return
		}
		e := app.ui.getExpr()
		if e == nil {
			continue
		}
		e.eval(app, nil)
		app.ui.draw(app.nav)
	}
}

func (app *App) exportVars() {
	dir := app.nav.currDir()

	var envFile string
	if len(dir.fi) != 0 {
		envFile = app.nav.currPath()
	}

	marks := app.nav.currMarks()

	envFiles := strings.Join(marks, ":")

	os.Setenv("f", envFile)
	os.Setenv("fs", envFiles)

	if len(marks) == 0 {
		os.Setenv("fx", envFile)
	} else {
		os.Setenv("fx", envFiles)
	}
}

// This function is used to run a command in shell. Following modes are used:
//
// Prefix  Wait  Async  Stdin/Stdout/Stderr  UI action (before/after)
// $       No    No     Yes                  Do nothing and then sync
// !       Yes   No     Yes                  pause and then resume
// &       No    Yes    No                   Do nothing
//
// Waiting async commands are not used for now.
func (app *App) runShell(s string, args []string, wait bool, async bool) {
	app.exportVars()

	if len(gOpts.ifs) != 0 {
		s = fmt.Sprintf("IFS='%s'; %s", gOpts.ifs, s)
	}

	args = append([]string{"-c", s, "--"}, args...)
	cmd := exec.Command(envShell, args...)

	if !async {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if wait {
		app.ui.pause()
		defer app.ui.resume()
	} else {
		defer app.ui.sync()
	}

	defer app.nav.renew(app.ui.wins[0].h)

	var err error
	if async {
		err = cmd.Start()
	} else {
		err = cmd.Run()
	}

	if err != nil {
		msg := fmt.Sprintf("running shell: %s", err)
		app.ui.message = msg
		log.Print(msg)
	}

	if wait {
		if err := waitKey(); err != nil {
			msg := fmt.Sprintf("waiting shell: %s", err)
			app.ui.message = msg
			log.Print(msg)
		}
	}
}
