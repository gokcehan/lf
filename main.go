package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
)

var (
	envUser  = os.Getenv("USER")
	envHome  = os.Getenv("HOME")
	envHost  = os.Getenv("HOSTNAME")
	envPath  = os.Getenv("PATH")
	envShell = os.Getenv("SHELL")
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

	tmp := os.TempDir()

	gSocketPath = path.Join(tmp, fmt.Sprintf("lf.%s.sock", envUser))

	// TODO: unique log file for each client
	gLogPath = path.Join(tmp, fmt.Sprintf("lf.%s.log", envUser))
	gServerLogPath = path.Join(tmp, fmt.Sprintf("lf.%s.server.log", envUser))

	// TODO: xdg-config-home etc.
	gConfigPath = path.Join(envHome, ".config", "lf", "lfrc")
}

func startServer() {
	cmd := exec.Command(os.Args[0], "-server")
	err := cmd.Start()
	if err != nil {
		log.Printf("starting server: %s", err)
	}
}

func main() {
	serverMode := flag.Bool("server", false, "start server (automatic)")
	flag.StringVar(&gLastDirPath, "last-dir-path", "", "path to the file to write the last dir on exit (to use for cd)")
	flag.StringVar(&gSelectionPath, "selection-path", "", "path to the file to write selected files on exit (to use as open file dialog)")

	flag.Parse()

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
