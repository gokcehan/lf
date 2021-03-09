package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
)

var (
	envPath  = os.Getenv("PATH")
	envLevel = os.Getenv("LF_LEVEL")
)

type arrayFlag []string

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
	gCustomConfig  string
	gCommands      arrayFlag
	gVersion       string
)

func (a *arrayFlag) Set(v string) error {
	*a = append(*a, v)
	return nil
}

func (a *arrayFlag) String() string {
	return strings.Join(*a, ", ")
}

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

func exportEnvVars() {
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

// used by exportOpts below
func fieldToString(field reflect.Value) string {
	kind := field.Kind()
	var value string

	switch kind {
	case reflect.Int:
		value = strconv.Itoa(int(field.Int()))
	case reflect.Bool:
		value = strconv.FormatBool(field.Bool())
	case reflect.Slice:
		for i := 0; i < field.Len(); i++ {
			element := field.Index(i)

			if i == 0 {
				value = fieldToString(element)
			} else {
				value += ":" + fieldToString(element)
			}
		}
	default:
		value = field.String()
	}

	return value
}

func exportOpts() {
	e := reflect.ValueOf(&gOpts).Elem()

	for i := 0; i < e.NumField(); i++ {
		// Get name and prefix it with lf_
		name := e.Type().Field(i).Name
		name = fmt.Sprintf("lf_%s", name)

		// Skip maps
		if name == "lf_keys" || name == "lf_cmdkeys" || name == "lf_cmds" {
			continue
		}

		// Get string representation of the value
		if name == "lf_sortType" {
			var sortby string

			switch gOpts.sortType.method {
			case naturalSort:
				sortby = "natural"
			case nameSort:
				sortby = "name"
			case sizeSort:
				sortby = "size"
			case timeSort:
				sortby = "time"
			case ctimeSort:
				sortby = "ctime"
			case atimeSort:
				sortby = "atime"
			case extSort:
				sortby = "ext"
			}

			os.Setenv("lf_sortby", sortby)

			reverse := strconv.FormatBool(gOpts.sortType.option&reverseSort != 0)
			os.Setenv("lf_reverse", reverse)

			hidden := strconv.FormatBool(gOpts.sortType.option&hiddenSort != 0)
			os.Setenv("lf_hidden", hidden)

			dirfirst := strconv.FormatBool(gOpts.sortType.option&dirfirstSort != 0)
			os.Setenv("lf_dirfirst", dirfirst)
		} else {
			field := e.Field(i)
			value := fieldToString(field)

			os.Setenv(name, value)
		}
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

	flag.StringVar(&gCustomConfig,
		"config",
		"",
		"path to a custom config file to be used, instead of normal lfrc file")

	flag.Var(&gCommands,
		"command",
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
		if err := remote(*remoteCmd); err != nil {
			log.Fatalf("remote command: %s", err)
		}
	case *serverMode:
		os.Chdir(gUser.HomeDir)
		gServerLogPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.server.log", gUser.Username))
		serve()
	default:
		checkServer()

		gClientID = os.Getpid()
		gLogPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.%d.log", gUser.Username, gClientID))
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

		exportEnvVars()

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
