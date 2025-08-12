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
	ui             *ui
	nav            *nav
	ticker         *time.Ticker
	quitChan       chan struct{}
	cmd            *exec.Cmd
	cmdIn          io.WriteCloser
	cmdOutBuf      []byte
	cmdHistory     []cmdItem
	cmdHistoryBeg  int
	cmdHistoryInd  int
	menuCompActive bool
	menuCompTmp    []string
	menuComps      []compMatch
	menuCompInd    int
	selectionOut   []string
	watch          *watch
	quitting       bool
}

func newApp(ui *ui, nav *nav) *app {
	quitChan := make(chan struct{}, 1)

	app := &app{
		ui:       ui,
		nav:      nav,
		ticker:   new(time.Ticker),
		quitChan: quitChan,
		watch:    newWatch(nav.dirChan, nav.fileChan, nav.delChan),
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		for {
			switch <-sigChan {
			case os.Interrupt:
			case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM:
				app.quit()
				os.Exit(3)
				return
			}
		}
	}()

	return app
}

func (app *app) quit() {
	// Using synchronous shell commands for `on-quit` can cause this to be
	// called again, so a guard variable is introduced here to prevent an
	// infinite loop.
	if app.quitting {
		return
	}
	app.quitting = true

	onQuit(app)

	if gOpts.history {
		if err := app.writeHistory(); err != nil {
			log.Printf("writing history file: %s", err)
		}
	}
	if !gSingleMode {
		if err := remote(fmt.Sprintf("drop %d", gClientID)); err != nil {
			log.Printf("dropping connection: %s", err)
		}
		if gOpts.autoquit {
			if err := remote("quit"); err != nil {
				log.Printf("auto quitting server: %s", err)
			}
		}
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

func loadFiles() (clipboard clipboard, err error) {
	files, err := os.Open(gFilesPath)
	if os.IsNotExist(err) {
		err = nil
		return
	}
	if err != nil {
		err = fmt.Errorf("opening file selections file: %s", err)
		return
	}
	defer files.Close()

	s := bufio.NewScanner(files)

	s.Scan()

	switch s.Text() {
	case "copy":
		clipboard.mode = clipboardCopy
	case "move":
		clipboard.mode = clipboardCut
	default:
		err = fmt.Errorf("unexpected option to copy file(s): %s", s.Text())
		return
	}

	for s.Scan() && s.Text() != "" {
		clipboard.paths = append(clipboard.paths, s.Text())
	}

	if s.Err() != nil {
		err = fmt.Errorf("scanning file list: %s", s.Err())
		return
	}

	log.Printf("loading files: %v", clipboard.paths)

	return
}

func saveFiles(clipboard clipboard) error {
	if err := os.MkdirAll(filepath.Dir(gFilesPath), os.ModePerm); err != nil {
		return fmt.Errorf("creating data directory: %s", err)
	}

	files, err := os.Create(gFilesPath)
	if err != nil {
		return fmt.Errorf("opening file selections file: %s", err)
	}
	defer files.Close()

	log.Printf("saving files: %v", clipboard.paths)

	if clipboard.mode == clipboardCopy {
		fmt.Fprintln(files, "copy")
	} else {
		fmt.Fprintln(files, "move")
	}

	for _, path := range clipboard.paths {
		fmt.Fprintln(files, path)
	}

	files.Sync()
	return nil
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
		if toks[0] != ":" && toks[0] != "$" && toks[0] != "%" && toks[0] != "!" && toks[0] != "&" {
			continue
		}
		if len(toks) < 2 {
			continue
		}
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
		_, err = fmt.Fprintf(f, "%s %s\n", cmd.prefix, cmd.value)
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
	go app.nav.previewLoop(app.ui)

	var serverChan <-chan expr
	if !gSingleMode {
		serverChan = readExpr()
	}

	app.ui.readExpr()

	if gConfigPath != "" {
		if _, err := os.Stat(gConfigPath); !os.IsNotExist(err) {
			app.readFile(gConfigPath)
		} else {
			log.Printf("config file does not exist: %s", err)
		}
	} else {
		for _, path := range gConfigPaths {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				app.readFile(path)
			}
		}
	}

	for _, cmd := range gCommands {
		p := newParser(strings.NewReader(cmd))

		for p.parse() {
			p.expr.eval(app, nil)
		}

		if p.err != nil {
			app.ui.echoerrf("%s", p.err)
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Printf("getting current directory: %s", err)
	}

	app.nav.getDirs(wd)
	app.nav.addJumpList()
	app.nav.init = true

	if gSelect != "" {
		go func() {
			lstat, err := os.Lstat(gSelect)
			if err != nil {
				app.ui.exprChan <- &callExpr{"echoerr", []string{err.Error()}, 1}
			} else if lstat.IsDir() {
				app.ui.exprChan <- &callExpr{"cd", []string{gSelect}, 1}
			} else {
				app.ui.exprChan <- &callExpr{"select", []string{gSelect}, 1}
			}
		}()
	}

	for {
		select {
		case <-app.quitChan:
			if app.nav.copyJobs > 0 {
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

			app.quit()

			app.nav.previewChan <- ""

			log.Print("bye!")

			return
		case <-app.nav.copyJobsChan:
			app.nav.copyJobs += 1
			app.ui.draw(app.nav)
		case n := <-app.nav.copyBytesChan:
			app.nav.copyBytes += n
			// n is usually 32*1024B (default io.Copy() buffer) so update roughly per 32KB x 128 = 4MB copied
			if app.nav.copyUpdate++; app.nav.copyUpdate >= 128 {
				app.nav.copyUpdate = 0
				app.ui.draw(app.nav)
			}
		case n := <-app.nav.copyTotalChan:
			app.nav.copyTotal += n
			if n < 0 {
				app.nav.copyBytes += n
				app.nav.copyJobs -= 1
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
			if gOpts.dircache {
				prev, ok := app.nav.dirCache[d.path]
				if ok {
					d.ind = prev.ind
					d.pos = prev.pos
					d.visualAnchor = min(prev.visualAnchor, len(d.files)-1)
					d.visualWrap = prev.visualWrap
					d.filter = prev.filter
					d.sort()
					d.sel(prev.name(), app.nav.height)
				}

				app.nav.dirCache[d.path] = d
			} else {
				d.sort()
			}

			var oldCurrPath string
			if curr, err := app.nav.currFile(); err == nil {
				oldCurrPath = curr.path
			}

			for i := range app.nav.dirs {
				if app.nav.dirs[i].path == d.path {
					app.nav.dirs[i] = d
				}
			}

			app.nav.position()

			curr, err := app.nav.currFile()
			if err == nil {
				if curr.path != oldCurrPath {
					app.ui.loadFile(app, true)
					if app.ui.msgIsStat {
						app.ui.loadFileInfo(app.nav)
					}
				}
				if d.path == curr.path {
					app.ui.dirPrev = d
				}
			}

			app.watchDir(d)

			paths := []string{}
			for _, file := range d.allFiles {
				paths = append(paths, file.path)
			}
			onLoad(app, paths)

			app.ui.draw(app.nav)
		case r := <-app.nav.regChan:
			app.nav.regCache[r.path] = r

			curr, err := app.nav.currFile()
			if err == nil {
				if r.path == curr.path {
					app.ui.regPrev = r
					if gOpts.sixel {
						app.ui.sxScreen.forceClear = true
					}
				}
			}

			app.ui.draw(app.nav)
		case f := <-app.nav.fileChan:
			for _, dir := range app.nav.dirCache {
				if dir.path != filepath.Dir(f.path) {
					continue
				}

				for i := range dir.allFiles {
					if dir.allFiles[i].path == f.path {
						dir.allFiles[i] = f
						break
					}
				}

				name := dir.name()
				dir.sort()
				dir.sel(name, app.nav.height)
			}

			app.ui.loadFile(app, false)
			if app.ui.msgIsStat {
				app.ui.loadFileInfo(app.nav)
			}

			onLoad(app, []string{f.path})
			app.ui.draw(app.nav)
		case path := <-app.nav.delChan:
			deletePathRecursive(app.nav.selections, path)
			if len(app.nav.selections) == 0 {
				app.nav.selectionInd = 0
			}

			deletePathRecursive(app.nav.regCache, path)

			deletePathRecursive(app.nav.dirCache, path)
			currPath := app.nav.currDir().path
			if currPath == path || strings.HasPrefix(currPath, path+string(filepath.Separator)) {
				if wd, err := os.Getwd(); err == nil {
					app.nav.getDirs(wd)
				}
			}
		case ev := <-app.ui.evChan:
			e := app.ui.readEvent(ev, app.nav)
			if e == nil {
				continue
			}
			e.eval(app, nil)
		loop:
			for {
				select {
				case ev := <-app.ui.evChan:
					e = app.ui.readEvent(ev, app.nav)
					if e == nil {
						continue
					}
					e.eval(app, nil)
				default:
					break loop
				}
			}
			app.ui.draw(app.nav)
		case e := <-app.ui.exprChan:
			e.eval(app, nil)
			app.ui.draw(app.nav)
		case e := <-serverChan:
			e.eval(app, nil)
			app.ui.draw(app.nav)
		case <-app.ticker.C:
			app.nav.renew()
			app.ui.loadFile(app, false)
		case <-app.nav.previewTimer.C:
			app.nav.previewLoading = true
			app.ui.draw(app.nav)
		}
	}
}

func (app *app) runCmdSync(cmd *exec.Cmd, pause_after bool) {
	app.nav.previewChan <- ""

	if err := app.ui.suspend(); err != nil {
		log.Printf("suspend: %s", err)
	}
	defer func() {
		if err := app.ui.resume(); err != nil {
			app.quit()
			os.Exit(3)
		}
	}()

	if err := cmd.Run(); err != nil {
		app.ui.echoerrf("running shell: %s", err)
	}
	if pause_after {
		anyKey()
	}

	app.ui.loadFile(app, true)
	app.nav.renew()
}

// This function is used to run a shell command. Modes are as follows:
//
//	Prefix  Wait  Async  Stdin  Stdout  Stderr  UI action
//	$       No    No     Yes    Yes     Yes     Pause and then resume
//	%       No    No     Yes    Yes     Yes     Statline for input/output
//	!       Yes   No     Yes    Yes     Yes     Pause and then resume
//	&       No    Yes    No     No      No      Do nothing
func (app *app) runShell(s string, args []string, prefix string) {
	app.nav.exportFiles()
	app.ui.exportSizes()
	app.exportMode()
	exportLfPath()
	exportOpts()

	gState.mutex.Lock()
	gState.data["maps"] = listBinds(map[string]map[string]expr{
		"n": gOpts.nkeys,
		"v": gOpts.vkeys,
	})
	gState.data["nmaps"] = listBinds(map[string]map[string]expr{
		"n": gOpts.nkeys,
	})
	gState.data["vmaps"] = listBinds(map[string]map[string]expr{
		"v": gOpts.vkeys,
	})
	gState.data["cmaps"] = listBinds(map[string]map[string]expr{
		"c": gOpts.cmdkeys,
	})
	gState.data["cmds"] = listCmds(gOpts.cmds)
	gState.data["jumps"] = listJumps(app.nav.jumpList, app.nav.jumpListInd)
	gState.data["history"] = listHistory(app.cmdHistory)
	gState.data["files"] = listFilesInCurrDir(app.nav)
	gState.mutex.Unlock()

	cmd := shellCommand(s, args)

	var out io.Reader
	var err error
	switch prefix {
	case "$", "!":
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr

		app.runCmdSync(cmd, prefix == "!")
		return
	}

	// We are running the command asynchronously
	if prefix == "%" {
		if app.ui.cmdPrefix == ">" {
			return
		}
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

	shellSetPG(cmd)
	if err = cmd.Start(); err != nil {
		app.ui.echoerrf("running shell: %s", err)
	}

	switch prefix {
	case "%":
		normal(app)
		app.cmd = cmd
		app.cmdOutBuf = nil
		app.ui.cmdPrefix = ">"
		app.ui.echo("")

		go func() {
			eol := false
			reader := bufio.NewReader(out)
			for {
				b, err := reader.ReadByte()
				if err == io.EOF {
					break
				}
				if eol {
					eol = false
					app.cmdOutBuf = nil
				}
				app.cmdOutBuf = append(app.cmdOutBuf, b)
				if b == '\n' || b == '\r' {
					eol = true
				}
				if reader.Buffered() > 0 {
					continue
				}
				app.ui.exprChan <- &callExpr{"echo", []string{string(app.cmdOutBuf)}, 1}
			}

			if err := cmd.Wait(); err != nil {
				log.Printf("running shell: %s", err)
			}
			app.cmd = nil
			app.ui.cmdPrefix = ""
			app.ui.exprChan <- &callExpr{"load", nil, 1}
		}()
	case "&":
		go func() {
			if err := cmd.Wait(); err != nil {
				log.Printf("running shell: %s", err)
			}
			app.ui.exprChan <- &callExpr{"load", nil, 1}
		}()
	}
}

func (app *app) doComplete() (matches []compMatch) {
	var result string

	switch app.ui.cmdPrefix {
	case ":":
		matches, result = completeCmd(app.ui.cmdAccLeft)
	case "$", "%", "!", "&":
		matches, result = completeShell(app.ui.cmdAccLeft)
	case "/", "?":
		matches, result = completeSearch(app.ui.cmdAccLeft)
	}

	app.ui.cmdAccLeft = []rune(result)
	app.ui.menu = listMatches(app.ui.screen, matches, -1)
	return
}

func (app *app) menuComplete(direction int) {
	if !app.menuCompActive {
		app.menuCompTmp = tokenize(string(app.ui.cmdAccLeft))
		app.menuComps = app.doComplete()
		if len(app.menuComps) > 1 {
			app.menuCompInd = -1
			app.menuCompActive = true
		}
	} else {
		app.menuCompInd += direction
		if app.menuCompInd == len(app.menuComps) {
			app.menuCompInd = 0
		} else if app.menuCompInd < 0 {
			app.menuCompInd = len(app.menuComps) - 1
		}

		app.menuCompTmp[len(app.menuCompTmp)-1] = app.menuComps[app.menuCompInd].result
		app.ui.cmdAccLeft = []rune(strings.Join(app.menuCompTmp, " "))
	}
	app.ui.menu = listMatches(app.ui.screen, app.menuComps, app.menuCompInd)
}

func (app *app) watchDir(dir *dir) {
	if !gOpts.watch {
		return
	}

	app.watch.add(dir.path)

	// ensure dircounts are updated for child directories
	for _, file := range dir.allFiles {
		if file.IsDir() {
			app.watch.add(file.path)
		}
	}
}

func (app *app) exportMode() {
	getMode := func() string {
		if strings.HasPrefix(app.ui.cmdPrefix, "delete") {
			return "delete"
		}

		if strings.HasPrefix(app.ui.cmdPrefix, "replace") || strings.HasPrefix(app.ui.cmdPrefix, "create") {
			return "rename"
		}

		switch app.ui.cmdPrefix {
		case "filter: ":
			return "filter"
		case "find: ", "find-back: ":
			return "find"
		case "mark-save: ", "mark-load: ", "mark-remove: ":
			return "mark"
		case "rename: ":
			return "rename"
		case "/", "?":
			return "search"
		case ":":
			return "command"
		case "$", "%", "!", "&":
			return "shell"
		case ">":
			return "pipe"
		case "":
			if app.nav.isVisualMode() {
				return "visual"
			}
			return "normal"
		default:
			return "unknown"
		}
	}

	os.Setenv("lf_mode", getMode())
}
