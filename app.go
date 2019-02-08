package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type cmdItem struct {
	prefix string
	value  string
}

type app struct {
	ui            *ui
	nav           *nav
	ticker        *time.Ticker
	quitChan      chan bool
	cmd           *exec.Cmd
	cmdIn         io.WriteCloser
	cmdOutBuf     []byte
	cmdHistory    []cmdItem
	cmdHistoryBeg int
	cmdHistoryInd int
}

func newApp() *app {
	ui := newUI()
	nav := newNav(ui.wins[0].h)

	return &app{
		ui:       ui,
		nav:      nav,
		ticker:   new(time.Ticker),
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

func (app *app) readHistory() error {
	f, err := os.Open(gHistoryPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("opening history file: %s", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		toks := strings.SplitN(scanner.Text(), " ", 2)
		app.cmdHistory = append(app.cmdHistory, cmdItem{toks[0], toks[1]})
	}

	app.cmdHistoryBeg = len(app.cmdHistory)

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading history file: %s", err)
	}

	return nil
}

func (app *app) writeHistory() error {
	if len(app.cmdHistory) == 0 {
		return nil
	}

	local := make([]cmdItem, len(app.cmdHistory)-app.cmdHistoryBeg)
	copy(local, app.cmdHistory[app.cmdHistoryBeg:])
	app.cmdHistory = nil

	if err := app.readHistory(); err != nil {
		return fmt.Errorf("reading history file: %s", err)
	}

	app.cmdHistory = append(app.cmdHistory, local...)

	if err := os.MkdirAll(filepath.Dir(gHistoryPath), os.ModePerm); err != nil {
		return fmt.Errorf("creating data directory: %s", err)
	}

	f, err := os.Create(gHistoryPath)
	if err != nil {
		return fmt.Errorf("creating history file: %s", err)
	}
	defer f.Close()

	if len(app.cmdHistory) > 1000 {
		app.cmdHistory = app.cmdHistory[len(app.cmdHistory)-1000:]
	}

	for _, cmd := range app.cmdHistory {
		_, err = f.WriteString(fmt.Sprintf("%s %s\n", cmd.prefix, cmd.value))
		if err != nil {
			return fmt.Errorf("writing history file: %s", err)
		}
	}

	return nil
}

// This is the main event loop of the application. Expressions are read from
// the client and the server on separate goroutines and sent here over channels
// for evaluation. Similarly directories and regular files are also read in
// separate goroutines and sent here for update.
func (app *app) loop() {
	clientChan := app.ui.readExpr()
	serverChan := readExpr()

	for _, path := range gConfigPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			app.readFile(path)
		}
	}

	if gCommand != "" {
		p := newParser(strings.NewReader(gCommand))
		if e := p.parseExpr(); e != nil {
			log.Printf("evaluating start command: %s", e.String())
			e.eval(app, nil)
		}
	}

	for {
		select {
		case <-app.quitChan:
			log.Print("bye!")

			if err := app.nav.writeMarks(); err != nil {
				log.Printf("writing marks file: %s", err)
			}

			if err := app.writeHistory(); err != nil {
				log.Printf("writing history file: %s", err)
			}

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
				d.ind = prev.ind
				d.sel(prev.name(), app.nav.height)
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
			e.eval(app, nil)
			app.ui.draw(app.nav)
		case e := <-serverChan:
			e.eval(app, nil)
			app.ui.draw(app.nav)
		case <-app.ticker.C:
			app.nav.renew()
			app.ui.loadFile(app.nav)
			app.ui.draw(app.nav)
		}
	}
}

func (app *app) exportVars() {
	var envFile string
	if f, err := app.nav.currFile(); err == nil {
		envFile = f.path
	}

	selections := app.nav.currSelections()

	envFiles := strings.Join(selections, gOpts.filesep)

	os.Setenv("f", envFile)
	os.Setenv("fs", envFiles)

	if len(selections) == 0 {
		os.Setenv("fx", envFile)
	} else {
		os.Setenv("fx", envFiles)
	}

	os.Setenv("id", strconv.Itoa(gClientID))

	os.Setenv("OPENER", envOpener)
	os.Setenv("EDITOR", envEditor)
	os.Setenv("PAGER", envPager)
	os.Setenv("SHELL", envShell)

	level, err := strconv.Atoi(envLevel)
	if err != nil {
		log.Printf("reading lf level: %s", err)
	}

	level++

	os.Setenv("LF_LEVEL", strconv.Itoa(level))
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

// This function is used to run a shell command. Modes are as follows:
//
//     Prefix  Wait  Async  Stdin  Stdout  Stderr  UI action
//     $       No    No     Yes    Yes     Yes     Pause and then resume
//     %       No    No     Yes    Yes     Yes     Statline for input/output
//     !       Yes   No     Yes    Yes     Yes     Pause and then resume
//     &       No    Yes    No     No      No      Do nothing
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
		defer app.nav.renew()
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
		app.cmd = cmd
		app.cmdOutBuf = nil
		app.ui.msg = ""
		app.ui.cmdPrefix = ">"

		reader := bufio.NewReader(out)
		bytes := make(chan byte, 1000)

		go func() {
			for {
				b, err := reader.ReadByte()
				if err == io.EOF {
					break
				}
				bytes <- b
			}
			close(bytes)
		}()

		go func() {
		loop:
			for {
				select {
				case b, ok := <-bytes:
					switch {
					case !ok:
						break loop
					case b == '\n' || b == '\r':
						app.ui.exprChan <- &callExpr{"echo", []string{string(app.cmdOutBuf)}, 1}
						app.cmdOutBuf = nil
					default:
						app.cmdOutBuf = append(app.cmdOutBuf, b)
					}
				default:
					app.ui.exprChan <- &callExpr{"echo", []string{string(app.cmdOutBuf)}, 1}
				}
			}

			if err := cmd.Wait(); err != nil {
				log.Printf("running shell: %s", err)
			}
			app.cmd = nil
			app.ui.cmdPrefix = ""
			app.ui.exprChan <- &callExpr{"load", nil, 1}
		}()
	}
}
