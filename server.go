package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

var (
	gCopyFile          bool
	gFileList          []string
	gConnList          = make(map[int]net.Conn)
	gQuitChan          = make(chan bool, 1)
	gListener          net.Listener
	gConnectionCount   int
	gSelectedFilesPath string
)

func serve() {
	gSelectedFilesPath = filepath.Join(os.TempDir(), fmt.Sprintf("lf.%s.server.files", gUser.Username))
	f, err := os.Create(gServerLogPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Print("hi!")
	if _, err := os.Stat(gSelectedFilesPath); err == nil {
		if err := getSelectedFiles(); err != nil {
			log.Printf("failed to obtain the previously selected files: %s", err.Error())
		} else {
			log.Printf("read previously selected files from: %s", gSelectedFilesPath)
		}
	}

	l, err := net.Listen(gSocketProt, gSocketPath)
	if err != nil {
		log.Printf("listening socket: %s", err)
		return
	}
	defer l.Close()

	gListener = l

	listen(l)
}

func listen(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			select {
			case <-gQuitChan:
				log.Printf("bye!")
				return
			default:
				log.Printf("failed to accept connection: %s", err)
			}
		}
		log.Println("accepted new connection")
		gConnectionCount++
		go handleConn(c)
	}
}

func handleConn(c net.Conn) {
	s := bufio.NewScanner(c)

Loop:
	for s.Scan() {
		log.Printf("listen: %s", s.Text())
		word, rest := splitWord(s.Text())
		switch word {
		case "save":
			s.Scan()
			switch s.Text() {
			case "copy":
				gCopyFile = true
			case "move":
				gCopyFile = false
			default:
				log.Printf("unexpected option to copy file(s): %s", s.Text())
				break Loop
			}
			gFileList = nil
			for s.Scan() && s.Text() != "" {
				gFileList = append(gFileList, s.Text())
			}
		case "load":
			if gCopyFile {
				fmt.Fprintln(c, "copy")
			} else {
				fmt.Fprintln(c, "move")
			}
			for _, f := range gFileList {
				fmt.Fprintln(c, f)
			}
			fmt.Fprintln(c)
		case "conn":
			if rest != "" {
				word2, _ := splitWord(rest)
				id, err := strconv.Atoi(word2)
				if err != nil {
					log.Print("listen: conn: client id should be a number")
				} else {
					gConnList[id] = c
				}
			} else {
				log.Print("listen: conn: requires a client id")
			}
		case "send":
			if rest != "" {
				word2, rest2 := splitWord(rest)
				id, err := strconv.Atoi(word2)
				if err != nil {
					for _, c := range gConnList {
						fmt.Fprintln(c, rest)
					}
				} else {
					if c, ok := gConnList[id]; ok {
						fmt.Fprintln(c, rest2)
					}
				}
			}
		case "quit":
			close(gQuitChan)
			for _, c := range gConnList {
				fmt.Fprintln(c, "echo server is quitting...")
				c.Close()
			}
			gListener.Close()
			break Loop
		default:
			log.Printf("listen: unexpected command: %s", word)
		}
	}

	if s.Err() != nil {
		log.Printf("listening: %s", s.Err())
	}

	gConnectionCount--
	if gConnectionCount == 0 {
		if gFileList == nil {
			if _, err := os.Stat(gSelectedFilesPath); err == nil {
				if err := os.Remove(gSelectedFilesPath); err != nil {
					log.Printf("failed to remove %s: %s", gSelectedFilesPath, err)
				} else {
					close(gQuitChan)
					gListener.Close()
				}
			} else {
				close(gQuitChan)
				gListener.Close()
			}
		} else {
			if err := writeSelectedFiles(); err != nil {
				log.Printf("failed to save selected files: %s", err)
			} else {
				close(gQuitChan)
				gListener.Close()
			}
		}
	}
	c.Close()
}

func getSelectedFiles() error {
	f, err := os.Open(gSelectedFilesPath)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)

	if !s.Scan() {
		return errors.New("failed to read " + gSelectedFilesPath)
	}
	gCopyFile, err = strconv.ParseBool(s.Text())
	if err != nil {
		return err
	}
	for s.Scan() {
		gFileList = append(gFileList, s.Text())
	}
	return s.Err()
}

func writeSelectedFiles() error {
	f, err := os.Create(gSelectedFilesPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprintln(w, strconv.FormatBool(gCopyFile))
	for _, line := range gFileList {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
