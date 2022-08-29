package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/djherbis/times"
)

func main() {
	switch len(os.Args) {
	case 1:
		tempFile()
		fmt.Println()
		tempDir()

	default:
		printTimes(os.Args[1])
	}
}

func tempDir() {
	name, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(name)
	fmt.Println("# DIR: " + name)

	symname := filepath.Join(filepath.Dir(name), "sym-"+filepath.Base(name))
	if err := os.Symlink(name, symname); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(symname)

	newAtime := time.Now().Add(-10 * time.Second)
	newMtime := time.Now().Add(10 * time.Second)
	if err := os.Chtimes(name, newAtime, newMtime); err != nil {
		log.Fatal(err)
	}

	printTimes(symname)
}

func tempFile() {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	fmt.Println("# FILE: " + f.Name())

	symname := filepath.Join(filepath.Dir(f.Name()), "sym-"+filepath.Base(f.Name()))
	if err := os.Symlink(f.Name(), symname); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(symname)

	newAtime := time.Now().Add(-10 * time.Second)
	newMtime := time.Now().Add(10 * time.Second)
	if err := os.Chtimes(f.Name(), newAtime, newMtime); err != nil {
		log.Fatal(err)
	}

	printTimes(symname)
}

func printTimes(name string) {
	fmt.Println("## Stat:", name)
	printTimespec(times.Stat(name))

	fmt.Println("\n## Lstat:", name)
	printTimespec(times.Lstat(name))
}

func printTimespec(ts times.Timespec, err error) {
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("AccessTime:", ts.AccessTime())
	fmt.Println("ModTime:", ts.ModTime())

	if ts.HasChangeTime() {
		fmt.Println("ChangeTime:", ts.ChangeTime())
	}

	if ts.HasBirthTime() {
		fmt.Println("BirthTime:", ts.BirthTime())
	}
}
