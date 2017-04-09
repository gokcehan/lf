package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type app struct {
	ui   *ui
	nav  *nav
	quit chan bool
}

func newApp() *app {
	ui := newUI()
	nav := newNav(ui.wins[0].h)
	quit := make(chan bool, 1)

	return &app{
		ui:   ui,
		nav:  nav,
		quit: quit,
	}
}

func waitKey() error {
	// TODO: this should be done with termbox somehow

	c := `echo
	      echo -n 'Press any key to continue'
	      old=$(stty -g)
	      stty raw -echo
	      eval "ignore=\$(dd bs=1 count=1 2> /dev/null)"
	      stty $old
	      echo`

	cmd := exec.Command(gOpts.shell, "-c", c)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("waiting key: %s", err)
	}

	return nil
}

// This is the main event loop of the application. There are two channels to
// read expressions from client and server. Reading and evaluation are done on
// separate goroutines.
func (app *app) handleInp(serverChan <-chan expr) {
	clientChan := app.ui.readExpr()

	for {
		select {
		case <-app.quit:
			log.Print("bye!")

			if gLastDirPath != "" {
				f, err := os.Create(gLastDirPath)
				if err != nil {
					log.Printf("opening last dir file: %s", err)
				}
				defer f.Close()

				dir := app.nav.currDir()

				_, err = f.WriteString(dir.path)
				if err != nil {
					log.Printf("writing last dir file: %s", err)
				}
			}

			return
		case e := <-clientChan:
			for i := 0; i < e.count; i++ {
				e.expr.eval(app, nil)
			}
			app.ui.draw(app.nav)
		case e := <-serverChan:
			e.eval(app, nil)
			app.ui.draw(app.nav)
		}
	}
}

func (app *app) exportVars() {
	var envFile string
	if f, err := app.nav.currFile(); err == nil {
		envFile = f.Path
	}

	marks := app.nav.currMarks()

	envFiles := strings.Join(marks, gOpts.filesep)

	os.Setenv("f", envFile)
	os.Setenv("fs", envFiles)

	if len(marks) == 0 {
		os.Setenv("fx", envFile)
	} else {
		os.Setenv("fx", envFiles)
	}

	os.Setenv("id", strconv.Itoa(gClientID))
}

// This function is used to run a command in shell. Following modes are used:
//
// Prefix  Wait  Async  Stdin/Stdout/Stderr  UI action
// $       No    No     Yes                  Pause and then resume
// !       Yes   No     Yes                  Pause and then resume
// &       No    Yes    No                   Do nothing
//
// Waiting async commands are not used for now.
func (app *app) runShell(s string, args []string, wait bool, async bool) {
	app.exportVars()

	if len(gOpts.ifs) != 0 {
		s = fmt.Sprintf("IFS='%s'; %s", gOpts.ifs, s)
	}

	args = append([]string{"-c", s, "--"}, args...)
	cmd := exec.Command(gOpts.shell, args...)

	if !async {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		app.ui.pause()
		defer app.ui.resume()
		defer app.nav.renew(app.ui.wins[0].h)
	}

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

	app.ui.loadFile(app.nav)
	app.ui.loadFileInfo(app.nav)
}
