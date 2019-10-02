package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
)

func (nav *nav) isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

type byLength []string

func (s byLength) Len() int {
	return len(s)
}
func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byLength) Less(i, j int) bool {
	return len(s[i]) > len(s[j])
}

func (nav *nav) delAll(list []string, errs chan error) {
	// start with files deepest in the file hierarchy
	sort.Sort(byLength(list)) // sort by length, decreasing
	later := make([]string, 0)

	for _, fpath := range list {
		log.Print(fpath)
		_, err := os.Stat(fpath)
		if os.IsNotExist(err) {
			errs <- fmt.Errorf("delete: %s", err)
			continue
		}

		filepath.Walk(fpath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				errs <- fmt.Errorf("walk: %s", err)
				return nil
			}
			if info.IsDir() { // delete only if empty directory
				empty, err := nav.isEmpty(path)
				if err != nil{
					errs <- fmt.Errorf("isempty: %s", err)
				} else if empty {
					if err := os.Remove(path); err != nil {
						errs <- fmt.Errorf("delete: %s", err)
					}
				} else { // non-empty directory
					// have to delete currently full folders later when their insides are cleared
					later = append(later, path)
				}
			} else {
				if err := os.Remove(path); err != nil {
					errs <- fmt.Errorf("delete: %s", err)
				}
			}
			return nil
		})
	}
	// delete dangling empty directories
	if len(later) > 0 {
		nav.delAll(later, errs)
	}
}

func (nav *nav) delAsync(ui *ui, list [] string) {
	echo := &callExpr{"echoerr", []string{""}, 1}
	errs := make(chan error, 1024)
	nav.delAll(list, errs)
	errCount := 0

	for {
		err, ok := <- errs
		if !ok {
			break
		}
		errCount ++
		echo.args[0] = fmt.Sprintf("[%d] %s", errCount, err)
		ui.exprChan <- echo
	}
}

