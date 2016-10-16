package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/nsf/termbox-go"
)

func client() {
	logFile, err := os.Create(gLogPath)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Print("hi!")

	if err := termbox.Init(); err != nil {
		log.Fatalf("initializing termbox: %s", err)
	}
	defer termbox.Close()

	ui := newUI()
	nav := newNav(ui.wins[0].h)
	app := &App{ui, nav}

	app.ui.loadFile(app.nav)

	if _, err := os.Stat(gConfigPath); err == nil {
		log.Printf("reading configuration file: %s", gConfigPath)

		rcFile, err := os.Open(gConfigPath)
		if err != nil {
			msg := fmt.Sprintf("opening configuration file: %s", err)
			app.ui.message = msg
			log.Printf(msg)
		}
		defer rcFile.Close()

		p := newParser(rcFile)
		for p.parse() {
			if p.err != nil {
				app.ui.message = "see the log file for errors in the configuration file"
				log.Print(p.err)
			}
			p.expr.eval(app, nil)
		}
	}

	app.ui.draw(app.nav)

	app.handleInp()
}

func saveFiles(list []string, keep bool) error {
	c, err := net.Dial("unix", gSocketPath)
	if err != nil {
		return fmt.Errorf("dialing to save files: %s", err)
	}
	defer c.Close()

	log.Printf("saving files: %v", list)

	fmt.Fprintln(c, "save")

	if keep {
		fmt.Fprintln(c, "keep")
	} else {
		fmt.Fprintln(c, "move")
	}

	for _, f := range list {
		fmt.Fprintln(c, f)
	}

	return nil
}

func loadFiles() (list []string, keep bool, err error) {
	c, e := net.Dial("unix", gSocketPath)
	if e != nil {
		err = fmt.Errorf("dialing to load files: %s", e)
		return
	}
	defer c.Close()

	fmt.Fprintln(c, "load")

	s := bufio.NewScanner(c)

	switch s.Scan(); s.Text() {
	case "keep":
		keep = true
	case "move":
		keep = false
	default:
		err = fmt.Errorf("unexpected option to keep file(s): %s", s.Text())
		return
	}

	for s.Scan() {
		list = append(list, s.Text())
	}

	if s.Err() != nil {
		err = fmt.Errorf("scanning file list: %s", s.Err())
		return
	}

	log.Printf("loading files: %v", list)

	return
}
