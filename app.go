package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type App struct {
	ui   *UI
	nav  *Nav
	quit chan bool
}

func newApp() *App {
	ui := newUI()
	nav := newNav(ui.wins[0].h)
	quit := make(chan bool)

	return &App{
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

func (app *App) handleInp() {
	clientChan := app.ui.readExpr(app)

	var serverChan chan Expr

	c, err := net.Dial("unix", gSocketPath)
	if err != nil {
		msg := fmt.Sprintf("connecting server: %s", err)
		app.ui.message = msg
		log.Printf(msg)
	} else {
		serverChan = readExpr(c)
	}

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

func (app *App) exportVars() {
	var envFile string
	if f, err := app.nav.currFile(); err == nil {
		envFile = f.Path
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

	os.Setenv("id", strconv.Itoa(gClientId))
}

// This function is used to run a command in shell. Following modes are used:
//
// Prefix  Wait  Async  Stdin/Stdout/Stderr  UI action
// $       No    No     Yes                  Pause and then resume
// !       Yes   No     Yes                  Pause and then resume
// &       No    Yes    No                   Do nothing
//
// Waiting async commands are not used for now.
func (app *App) runShell(s string, args []string, wait bool, async bool) {
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
}
