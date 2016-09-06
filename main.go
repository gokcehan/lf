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
	envUser   = os.Getenv("USER")
	envHome   = os.Getenv("HOME")
	envHost   = os.Getenv("HOSTNAME")
	envPath   = os.Getenv("PATH")
	envShell  = os.Getenv("SHELL")
	envConfig = os.Getenv("XDG_CONFIG_HOME")
)

var (
	gExitFlag      bool
	gLastDirPath   string
	gSelectionPath string
	gSocketPath    string
	gLogPath       string
	gServerLogPath string
	gConfigPath    string
)

func init() {
	if envUser == "" {
		log.Fatal("$USER not set")
	}
	if envHome == "" {
		envHome = "/home/" + envUser
	}
	if envHost == "" {
		host, err := os.Hostname()
		if err != nil {
			log.Printf("hostname: %s", err)
		}
		envHost = host
	}
	if envConfig == "" {
		envConfig = filepath.Join(envHome, ".config")
	}

	tmp := os.TempDir()

	gSocketPath = filepath.Join(tmp, fmt.Sprintf("lf.%s.sock", envUser))

	// TODO: unique log file for each client
	gLogPath = filepath.Join(tmp, fmt.Sprintf("lf.%s.log", envUser))
	gServerLogPath = filepath.Join(tmp, fmt.Sprintf("lf.%s.server.log", envUser))

	gConfigPath = filepath.Join(envConfig, "lf", "lfrc")
}

func startServer() {
	cmd := exec.Command(os.Args[0], "-server")
	if err := cmd.Start(); err != nil {
		log.Printf("starting server: %s", err)
	}
}

func main() {
	serverMode := flag.Bool("server", false, "start server (automatic)")
	cpuprofile := flag.String("cpuprofile", "", "path to the file to write the cpu profile")
	flag.StringVar(&gLastDirPath, "last-dir-path", "", "path to the file to write the last dir on exit (to use for cd)")
	flag.StringVar(&gSelectionPath, "selection-path", "", "path to the file to write selected files on exit (to use as open file dialog)")

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if *serverMode {
		serve()
	} else {
		// TODO: check if the socket is working
		if _, err := os.Stat(gSocketPath); os.IsNotExist(err) {
			startServer()
		}

		client()
	}
}
