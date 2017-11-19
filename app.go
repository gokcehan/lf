package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type cmdItem struct {
	prefix string
	value  string
}

type app struct {
	ui         *ui
	nav        *nav
	quitChan   chan bool
	cmdHist    []cmdItem
	cmdHistInd int
}

func newApp() *app {
	ui := newUI()
	nav := newNav(ui.wins[0].h)

	return &app{
		ui:       ui,
		nav:      nav,
		quitChan: make(chan bool, 1),
	}
}

func (app *app) readFile(path string) {
	log.Printf("reading file: %s", path)

	f, err := os.Open(path)
	if err != nil {
		app.ui.printf("opening file: %s", err)
		return
	}
	defer f.Close()

	p := newParser(f)
	for p.parse() {
		p.expr.eval(app, nil)
	}

	if p.err != nil {
		app.ui.printf("%s", p.err)
	}
}

// This is the main event loop of the application. There are two channels to
// read expressions from client and server. Reading and evaluation are done on
// separate goroutines.
func (app *app) loop() {
	clientChan := app.ui.readExpr()
	serverChan := readExpr()

	for {
		select {
		case <-app.quitChan:
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
		envFile = f.path
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

func waitKey() error {
	cmd := pauseCommand()

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("waiting key: %s", err)
	}

	return nil
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

	cmd := shellCommand(s, args)

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
		app.ui.printf("running shell: %s", err)
	}

	if wait {
		if err := waitKey(); err != nil {
			app.ui.printf("waiting shell: %s", err)
		}
	}

	app.ui.loadFile(app.nav)
	app.ui.loadFileInfo(app.nav)
}
