package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
	cmd        *exec.Cmd
	cmdIn      io.WriteCloser
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
		case d := <-app.nav.dirChan:
			prev, ok := app.nav.dirCache[d.path]
			if ok {
				d.find(prev.name(), app.nav.height)
			}

			app.nav.dirCache[d.path] = d

			for i := range app.nav.dirs {
				if app.nav.dirs[i].path == d.path {
					app.nav.dirs[i] = d
				}
			}

			app.nav.position()

			curr, err := app.nav.currFile()
			if err == nil {
				if d.path == app.nav.currDir().path {
					app.ui.loadFile(app.nav)
				}
				if d.path == curr.path {
					app.ui.dirPrev = d
				}
			}

			app.ui.draw(app.nav)
		case r := <-app.nav.regChan:
			app.nav.regCache[r.path] = r

			curr, err := app.nav.currFile()
			if err == nil {
				if r.path == curr.path {
					app.ui.regPrev = r
				}
			}

			app.ui.draw(app.nav)
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
// Prefix  Wait  Async  Stdin  Stdout  Stderr  UI action
// $       No    No     Yes    Yes     Yes     Pause and then resume
// %       No    No     No     Yes     Yes     Display output in statline
// !       Yes   No     Yes    Yes     Yes     Pause and then resume
// &       No    Yes    No     No      No      Do nothing
func (app *app) runShell(s string, args []string, prefix string) {
	app.exportVars()

	cmd := shellCommand(s, args)

	var out io.Reader
	switch prefix {
	case "$", "!":
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		app.ui.pause()
		defer app.ui.resume()
		defer app.nav.renew(app.ui.wins[0].h)
	case "%":
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Printf("writing stdin: %s", err)
		}
		app.cmdIn = stdin
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("reading stdout: %s", err)
		}
		out = stdout
		cmd.Stderr = cmd.Stdout
	}

	var err error
	switch prefix {
	case "$", "!":
		err = cmd.Run()
	case "%", "&":
		err = cmd.Start()
	}

	if err != nil {
		app.ui.printf("running shell: %s", err)
	}

	switch prefix {
	case "!":
		if err := waitKey(); err != nil {
			app.ui.printf("waiting key: %s", err)
		}
	}

	app.ui.loadFile(app.nav)
	app.ui.loadFileInfo(app.nav)

	switch prefix {
	case "%":
		go func() {
			app.cmd = cmd
			app.ui.msg = ""
			app.ui.cmdPrefix = ">"

			reader := bufio.NewReader(out)
			var buf []byte
			for {
				b, err := reader.ReadByte()
				if err == io.EOF {
					break
				}
				buf = append(buf, b)
				app.ui.exprChan <- multiExpr{&callExpr{"echo", []string{string(buf)}}, 1}
				if b == '\n' || b == '\r' {
					buf = nil
				}
			}

			if err := cmd.Wait(); err != nil {
				log.Printf("running shell: %s", err)
			}
			app.cmd = nil
			app.ui.cmdPrefix = ""
			app.ui.exprChan <- multiExpr{&callExpr{"reload", nil}, 1}
		}()
	}
}
