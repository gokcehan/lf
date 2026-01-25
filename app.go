package main

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"
)

type app struct {
	ui              *ui            // ui state (screen, windows, input)
	nav             *nav           // navigation state (dirs, cursor, selections, preview, caches)
	ticker          *time.Ticker   // refresh ticker if `period` > 0
	quitChan        chan struct{}  // signals main loop to exit
	cmd             *exec.Cmd      // currently running % (shell-pipe) command
	cmdIn           io.WriteCloser // stdin writer for running % command
	cmdOutBuf       []byte         // output of running % command
	cmdHistory      []string       // command history entries
	cmdHistoryBeg   int            // index where commands from this session start in cmdHistory
	cmdHistoryInd   int            // history navigation offset from most recent
	cmdHistoryInput *string        // initial input used as prefix filter while browsing history
	menuCompActive  bool           // whether completion cycling is active
	menuCompTmp     []string       // token snapshot taken when completion cycling starts, used for `cmd-menu-discard`
	menuComps       []compMatch    // completion candidates for active prompt
	menuCompInd     int            // index of selected completion candidate (-1: none selected)
	selectionOut    []string       // paths to output on exit, used for `-print-selection` and `-selection-path`
	watch           *watch         // fs watcher if `watch` is enabled
	quitting        bool           // guard to prevent re-entering quit logic
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
		if _, err := remote(fmt.Sprintf("drop %d", gClientID)); err != nil {
			log.Printf("dropping connection: %s", err)
		}
		if gOpts.autoquit {
			if _, err := remote("quit"); err != nil {
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
		err = fmt.Errorf("opening file selections file: %w", err)
		return
	}
	defer files.Close()

	s := bufio.NewScanner(files)

	if !s.Scan() {
		err = fmt.Errorf("scanning file list: %w", cmp.Or(s.Err(), io.EOF))
		return
	}

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
		err = fmt.Errorf("scanning file list: %w", s.Err())
		return
	}

	log.Printf("loading clipboard: %v", clipboard.paths)

	return
}

func saveFiles(clipboard clipboard) error {
	if err := os.MkdirAll(filepath.Dir(gFilesPath), os.ModePerm); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	files, err := os.Create(gFilesPath)
	if err != nil {
		return fmt.Errorf("opening file selections file: %w", err)
	}
	defer files.Close()

	log.Printf("saving files: %v", clipboard.paths)

	var clipboardModeStr string
	if clipboard.mode == clipboardCopy {
		clipboardModeStr = "copy"
	} else {
		clipboardModeStr = "move"
	}
	if _, err := fmt.Fprintln(files, clipboardModeStr); err != nil {
		return fmt.Errorf("write clipboard mode to file: %w", err)
	}

	for _, path := range clipboard.paths {
		if _, err := fmt.Fprintln(files, path); err != nil {
			return fmt.Errorf("write path to file: %w", err)
		}
	}

	return files.Sync()
}

func (app *app) readHistory() error {
	f, err := os.Open(gHistoryPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("opening history file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cmd := scanner.Text()
		if len(cmd) < 1 || !slices.Contains([]string{":", "$", "!", "%", "&"}, cmd[:1]) {
			continue
		}
		app.cmdHistory = append(app.cmdHistory, cmd)
	}

	app.cmdHistoryBeg = len(app.cmdHistory)

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading history file: %w", err)
	}

	return nil
}

func (app *app) writeHistory() error {
	if len(app.cmdHistory) == 0 {
		return nil
	}

	local := slices.Clone(app.cmdHistory[app.cmdHistoryBeg:])
	app.cmdHistory = nil

	if err := app.readHistory(); err != nil {
		return fmt.Errorf("reading history file: %w", err)
	}

	app.cmdHistory = append(app.cmdHistory, local...)
	if len(app.cmdHistory) > 1000 {
		app.cmdHistory = app.cmdHistory[len(app.cmdHistory)-1000:]
	}

	if err := os.MkdirAll(filepath.Dir(gHistoryPath), os.ModePerm); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	f, err := os.Create(gHistoryPath)
	if err != nil {
		return fmt.Errorf("creating history file: %w", err)
	}
	defer f.Close()

	for _, cmd := range app.cmdHistory {
		if _, err = fmt.Fprintln(f, cmd); err != nil {
			return fmt.Errorf("writing history file: %w", err)
		}
	}

	return nil
}

// loop is the main event loop of the application. Expressions are read from
// the client and the server on separate goroutines and sent here over channels
// for evaluation. Similarly directories and regular files are also read in
// separate goroutines and sent here for update.
func (app *app) loop() {
	go app.nav.preloadLoop(app.ui)
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

	app.nav.loadDirs()
	app.nav.addJumpList()

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

			log.Printf("*************** closing client, PID: %d ***************", os.Getpid())

			return
		case n := <-app.nav.copyJobsChan:
			app.nav.copyJobs += n
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
			var oldCurrPath string
			if curr := app.nav.currFile(); curr != nil {
				oldCurrPath = curr.path
			}

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

			app.nav.position()

			if curr := app.nav.currFile(); curr != nil {
				if curr.path != oldCurrPath {
					app.ui.loadFile(app, true)
				}
			}

			app.watchDir(d)

			paths := []string{}
			for _, file := range d.allFiles {
				paths = append(paths, file.path)
			}
			onLoad(app, paths)

			if d.path == app.nav.currDir().path {
				app.nav.preload()
			}

			app.ui.draw(app.nav)
		case r := <-app.nav.regChan:
			app.nav.regCache[r.path] = r

			if curr := app.nav.currFile(); curr != nil {
				if r.path == curr.path {
					app.ui.sxScreen.forceClear = true
					if gOpts.preload && r.volatile {
						app.ui.loadFile(app, true)
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

			delete(app.nav.regCache, f.path)
			app.ui.loadFile(app, false)
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
				app.nav.loadDirs()
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
		case <-app.nav.preloadTimer.C:
			app.nav.preload()
		}
	}
}

func (app *app) runCmdSync(cmd *exec.Cmd, pauseAfter bool) {
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
	if pauseAfter {
		anyKey()
	}

	app.ui.loadFile(app, true)
	app.nav.renew()
}

// runShell is used to run a shell command. Modes are as follows:
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

	switch prefix {
	case "$", "!":
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr

		app.runCmdSync(cmd, prefix == "!")
		return
	}

	// We are running the command asynchronously
	var inReader, inWriter, outReader, outWriter *os.File
	if prefix == "%" {
		if app.ui.cmdPrefix == ">" {
			return
		}

		// [exec.Cmd.StdoutPipe] cannot be used as it requires the output to be fully
		// read before calling [exec.Cmd.Wait], however in this case Cmd.Wait should
		// only wait for the command to finish executing regardless of whether the
		// output has been fully read or not.
		inReader, inWriter, err := os.Pipe()
		if err != nil {
			log.Printf("creating input pipe: %s", err)
			return
		}
		cmd.Stdin = inReader
		app.cmdIn = inWriter

		outReader, outWriter, err = os.Pipe()
		if err != nil {
			log.Printf("creating output pipe: %s", err)
			return
		}
		cmd.Stdout = outWriter
		cmd.Stderr = outWriter
	}

	shellSetPG(cmd)
	if err := cmd.Start(); err != nil {
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
			reader := bufio.NewReader(outReader)
			for {
				b, err := reader.ReadByte()
				if err != nil {
					if !errors.Is(err, io.EOF) && !errors.Is(err, fs.ErrClosed) {
						log.Printf("reading command output: %s", err)
					}
					break
				}

				app.cmdOutBuf = append(app.cmdOutBuf, b)
				if reader.Buffered() == 0 {
					app.ui.exprChan <- &callExpr{"echo", []string{string(app.cmdOutBuf)}, 1}
				}

				if b == '\n' || b == '\r' {
					app.cmdOutBuf = nil
				}
			}
		}()

		go func() {
			if err := cmd.Wait(); err != nil {
				log.Printf("running shell: %s", err)
			}
			inReader.Close()
			inWriter.Close()
			outReader.Close()
			outWriter.Close()

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
	var longest string

	switch app.ui.cmdPrefix {
	case ":":
		matches, longest = completeCmd(app.ui.cmdAccLeft)
	case "$", "%", "!", "&":
		matches, longest = completeShell(app.ui.cmdAccLeft)
	case "/", "?":
		matches, longest = completeSearch(app.ui.cmdAccLeft)
	}

	app.ui.cmdAccLeft = []rune(longest)
	app.ui.menu, app.ui.menuSelect = listMatches(app.ui.screen, matches, -1)
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

		toks := slices.Clone(app.menuCompTmp)
		toks[len(toks)-1] = app.menuComps[app.menuCompInd].result
		app.ui.cmdAccLeft = []rune(strings.Join(toks, " "))
	}
	app.ui.menu, app.ui.menuSelect = listMatches(app.ui.screen, app.menuComps, app.menuCompInd)
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
		if app.menuCompActive {
			return "compmenu"
		}

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
