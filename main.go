package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

var (
	envPath  = os.Getenv("PATH")
	envLevel = os.Getenv("LF_LEVEL")
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
	gSelect        string
	gCommand       string
	gVersion       string
)

func init() {
	h, err := os.Hostname()
	if err != nil {
		log.Printf("hostname: %s", err)
	}
	gHostname = h

	if envLevel == "" {
		envLevel = "0"
	}
}

func startServer() {
	cmd := detachedCommand(os.Args[0], "-server")
	if err := cmd.Start(); err != nil {
		log.Printf("starting server: %s", err)
	}
}

func checkServer() {
	if gSocketProt == "unix" {
		if _, err := os.Stat(gSocketPath); os.IsNotExist(err) {
			startServer()
		} else if _, err := net.Dial(gSocketProt, gSocketPath); err != nil {
			os.Remove(gSocketPath)
			startServer()
		}
	} else {
		if _, err := net.Dial(gSocketProt, gSocketPath); err != nil {
			startServer()
		}
	}
}

func main() {
	showDoc := flag.Bool(
		"doc",
		false,
		"show documentation")

	showVersion := flag.Bool(
		"version",
		false,
		"show version")

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
		"path to the file to write the CPU profile")

	memprofile := flag.String(
		"memprofile",
		"",
		"path to the file to write the memory profile")

	flag.StringVar(&gLastDirPath,
		"last-dir-path",
		"",
		"path to the file to write the last dir on exit (to use for cd)")

	flag.StringVar(&gSelectionPath,
		"selection-path",
		"",
		"path to the file to write selected files on open (to use as open file dialog)")

	flag.StringVar(&gCommand,
		"command",
		"",
		"command to execute on client initialization")

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

	switch {
	case *showDoc:
		fmt.Print(genDocString)
	case *showVersion:
		fmt.Println(gVersion)
	case *remoteCmd != "":
		if err := sendRemote(*remoteCmd); err != nil {
			log.Fatalf("remote command: %s", err)
		}
	case *serverMode:
		gServerLogPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.server.log", gUser.Username))
		serve()
	default:
		checkServer()

		gClientID = 1000
		gLogPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.%d.log", gUser.Username, gClientID))
		for _, err := os.Stat(gLogPath); !os.IsNotExist(err); _, err = os.Stat(gLogPath) {
			gClientID++
			gLogPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.%d.log", gUser.Username, gClientID))
		}

		switch flag.NArg() {
		case 0:
			_, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(2)
			}
		case 1:
			gSelect = flag.Arg(0)
		default:
			fmt.Fprintf(os.Stderr, "only single file or directory is allowed\n")
			os.Exit(2)
		}

		run()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
}
