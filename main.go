package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/pprof"
)

var (
	envPath = os.Getenv("PATH")
)

var (
	gClientID      int
	gHostname      string
	gLastDirPath   string
	gSelectionPath string
	gSocketProt    string
	gSocketPath    string
	gLogPath       string
	gServerLogPath string
)

func init() {
	h, err := os.Hostname()
	if err != nil {
		log.Printf("hostname: %s", err)
	}
	gHostname = h
}

func startServer() {
	cmd := exec.Command(os.Args[0], "-server")
	if err := cmd.Start(); err != nil {
		log.Printf("starting server: %s", err)
	}
}

func main() {
	showDoc := flag.Bool(
		"doc",
		false,
		"show documentation")

	remoteCmd := flag.String(
		"remote",
		"",
		"send remote command to server")

	serverMode := flag.Bool(
		"server",
		false,
		"start server (automatic)")

	cpuprofile := flag.String(
		"cpuprofile",
		"",
		"path to the file to write the cpu profile")

	flag.StringVar(&gLastDirPath,
		"last-dir-path",
		"",
		"path to the file to write the last dir on exit (to use for cd)")

	flag.StringVar(&gSelectionPath,
		"selection-path",
		"",
		"path to the file to write selected files on open (to use as open file dialog)")

	flag.Parse()

	gSocketProt = gDefaultSocketProt
	gSocketPath = gDefaultSocketPath

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatalf("could not create CPU profile: %s", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("could not start CPU profile: %s", err)
		}
		defer pprof.StopCPUProfile()
	}

	switch flag.NArg() {
	case 0:
	case 1:
		dir := flag.Arg(0)
		if err := os.Chdir(dir); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(2)
		}
	default:
		fmt.Fprintf(os.Stderr, "only single directory is allowed\n")
		os.Exit(2)
	}

	switch {
	case *showDoc:
		fmt.Print(genDocString)
	case *remoteCmd != "":
		if err := sendRemote(*remoteCmd); err != nil {
			log.Fatalf("remote command: %s", err)
		}
	case *serverMode:
		gServerLogPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.server.log", gUser.Username))
		serve()
	default:
		// TODO: check if the socket is working
		if _, err := os.Stat(gSocketPath); os.IsNotExist(err) {
			startServer()
		}

		gClientID = 1000
		gLogPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.%d.log", gUser.Username, gClientID))
		for _, err := os.Stat(gLogPath); !os.IsNotExist(err); _, err = os.Stat(gLogPath) {
			gClientID++
			gLogPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.%d.log", gUser.Username, gClientID))
		}

		run()
	}
}
