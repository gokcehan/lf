package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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

	quitChan := make(chan bool, 1)

	osChan := make(chan os.Signal, 1)
	signal.Notify(osChan, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	go func() {
		<-osChan
		nav.copyTotalChan <- -nav.copyTotal
		nav.moveTotalChan <- -nav.moveTotal
		nav.deleteTotalChan <- -nav.deleteTotal
		quitChan <- true
		return
	}()

	return &app{
		ui:       ui,
		nav:      nav,
		ticker:   new(time.Ticker),
		quitChan: quitChan,
	}
}

func (app *app) readFile(path string) {
	log.Printf("reading file: %s", path)

	f, err := os.Open(path)
	if err != nil {
		app.ui.echoerrf("opening file: %s", err)
		return
	}
	defer f.Close()

	p := newParser(f)

	for p.parse() {
		p.expr.eval(app, nil)
	}

	if p.err != nil {
		app.ui.echoerrf("%s", p.err)
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

	if gSelect != "" {
		go func() {
			stat, err := os.Stat(gSelect)
			if err != nil {
				app.ui.exprChan <- &callExpr{"echoerr", []string{err.Error()}, 1}
			} else if stat.IsDir() {
				app.ui.exprChan <- &callExpr{"cd", []string{gSelect}, 1}
			} else {
				app.ui.exprChan <- &callExpr{"select", []string{gSelect}, 1}
			}
		}()
	}

	for _, path := range gConfigPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			app.readFile(path)
		}
	}

	if gCommand != "" {
		p := newParser(strings.NewReader(gCommand))

		for p.parse() {
			p.expr.eval(app, nil)
		}

		if p.err != nil {
			app.ui.echoerrf("%s", p.err)
		}
	}

	for {
		select {
		case <-app.quitChan:
			if app.nav.copyTotal > 0 {
				app.ui.echoerr("quit: copy operation in progress")
				continue
			}

			if app.nav.moveTotal > 0 {
				app.ui.echoerr("quit: move operation in progress")
				continue
			}

			if app.nav.deleteTotal > 0 {
				app.ui.echoerr("quit: delete operation in progress")
				continue
			}

			log.Print("bye!")

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
		case n := <-app.nav.copyBytesChan:
			app.nav.copyBytes += n
			// n is usually 4096B so update roughly per 4096B x 1024 = 4MB copied
			if app.nav.copyUpdate++; app.nav.copyUpdate >= 1024 {
				app.nav.copyUpdate = 0
				app.ui.draw(app.nav)
			}
		case n := <-app.nav.copyTotalChan:
			app.nav.copyTotal += n
			if n < 0 {
				app.nav.copyBytes += n
			}
			if app.nav.copyTotal == 0 {
				app.nav.copyUpdate = 0
			}
			app.ui.draw(app.nav)
		case n := <-app.nav.moveCountChan:
			app.nav.moveCount += n
			if app.nav.moveUpdate++; app.nav.moveUpdate >= 1000 {
				app.nav.moveUpdate = 0
				app.ui.draw(app.nav)
			}
		case n := <-app.nav.moveTotalChan:
			app.nav.moveTotal += n
			if n < 0 {
				app.nav.moveCount += n
			}
			if app.nav.moveTotal == 0 {
				app.nav.moveUpdate = 0
			}
			app.ui.draw(app.nav)
		case n := <-app.nav.deleteCountChan:
			app.nav.deleteCount += n
			if app.nav.deleteUpdate++; app.nav.deleteUpdate >= 1000 {
				app.nav.deleteUpdate = 0
				app.ui.draw(app.nav)
			}
		case n := <-app.nav.deleteTotalChan:
			app.nav.deleteTotal += n
			if n < 0 {
				app.nav.deleteCount += n
			}
			if app.nav.deleteTotal == 0 {
				app.nav.deleteUpdate = 0
			}
			app.ui.draw(app.nav)
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

func (app *app) exportFiles() {
	var currFile string
	if f, err := app.nav.currFile(); err == nil {
		currFile = f.path
	}

	currSelections := app.nav.currSelections()

	exportFiles(currFile, currSelections)
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
	app.exportFiles()

	cmd := shellCommand(s, args)

	var out io.Reader
	var err error
	switch prefix {
	case "$", "!":
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		app.ui.pause()
		defer app.ui.resume()
		defer app.nav.renew()

		err = cmd.Run()
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
		fallthrough
	case "&":
		err = cmd.Start()
	}

	if err != nil {
		app.ui.echoerrf("running shell: %s", err)
	}

	switch prefix {
	case "!":
		if err := waitKey(); err != nil {
			app.ui.echoerrf("waiting key: %s", err)
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

		go func() {
			reader := bufio.NewReader(out)
			for {
				b, err := reader.ReadByte()
				if err == io.EOF {
					break
				}
				app.cmdOutBuf = append(app.cmdOutBuf, b)
				app.ui.exprChan <- &callExpr{"echo", []string{string(app.cmdOutBuf)}, 1}
				if b == '\n' || b == '\r' {
					app.cmdOutBuf = nil
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
