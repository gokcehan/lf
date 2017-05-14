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
	envConfig = os.Getenv("XDG_CONFIG_HOME")
)

var (
	gClientID      int
	gLastDirPath   string
	gSelectionPath string
	gSocketPath    string
	gLogPath       string
	gServerLogPath string
	gConfigPath    string
)

func init() {
	if envUser == "" {
		log.Print("$USER not set")
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

	gClientID = 1000
	gLogPath = filepath.Join(tmp, fmt.Sprintf("lf.%s.%d.log", envUser, gClientID))
	for _, err := os.Stat(gLogPath); err == nil; _, err = os.Stat(gLogPath) {
		gClientID++
		gLogPath = filepath.Join(tmp, fmt.Sprintf("lf.%s.%d.log", envUser, gClientID))
	}

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
	showDoc := flag.Bool("doc", false, "show documentation")
	remoteCmd := flag.String("remote", "", "send remote command to server")
	serverMode := flag.Bool("server", false, "start server (automatic)")
	cpuprofile := flag.String("cpuprofile", "", "path to the file to write the cpu profile")
	flag.StringVar(&gLastDirPath, "last-dir-path", "", "path to the file to write the last dir on exit (to use for cd)")
	flag.StringVar(&gSelectionPath, "selection-path", "", "path to the file to write selected files on open (to use as open file dialog)")

	flag.Parse()

	if *showDoc {
		fmt.Print(genDocString)
		return
	}

	if *remoteCmd != "" {
		if err := sendRemote(*remoteCmd); err != nil {
			log.Fatalf("remote command: %s", err)
		}
		return
	}

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
