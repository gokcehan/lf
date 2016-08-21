package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

var (
	gCmdWords = []string{"set", "map", "cmd"}
	gOptWords = []string{
		"preview",
		"nopreview",
		"preview!",
		"hidden",
		"nohidden",
		"hidden!",
		"tabstop",
		"scrolloff",
		"sortby",
		"showinfo",
		"opener",
		"ratios",
	}
)

func matchLongest(s1, s2 string) string {
	i := 0
	for ; i < len(s1) && i < len(s2); i++ {
		if s1[i] != s2[i] {
			break
		}
	}
	return s1[:i]
}

func matchWord(s string, words []string) (matches []string, longest string) {
	for _, w := range words {
		if strings.HasPrefix(w, s) {
			matches = append(matches, w)
			if longest != "" {
				longest = matchLongest(longest, w)
			} else {
				longest = w + " "
			}
		}
	}

	if longest == "" {
		longest = s
	}

	return
}

func matchExec(s string) (matches []string, longest string) {
	paths := strings.Split(envPath, ":")

	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			continue
		}

		fi, err := ioutil.ReadDir(p)
		if err != nil {
			log.Printf("reading path: %s", err)
		}

		for _, f := range fi {
			if strings.HasPrefix(f.Name(), s) {
				f, err = os.Stat(path.Join(p, f.Name()))
				if err != nil {
					log.Printf("getting file information: %s", err)
				}

				if !f.Mode().IsRegular() || f.Mode()&0111 == 0 {
					continue
				}

				matches = append(matches, f.Name())
				if longest != "" {
					longest = matchLongest(longest, f.Name())
				} else {
					longest = f.Name() + " "
				}
			}
		}
	}

	if longest == "" {
		longest = s
	}

	return
}

func matchFile(s string) (matches []string, longest string) {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("getting current directory: %s", err)
	}

	fi, err := ioutil.ReadDir(wd)
	if err != nil {
		log.Printf("reading directory: %s", err)
	}

	for _, f := range fi {
		if strings.HasPrefix(f.Name(), s) {
			matches = append(matches, f.Name())
			if longest != "" {
				longest = matchLongest(longest, f.Name())
			} else {
				longest = f.Name() + " "
			}
		}
	}

	if longest == "" {
		longest = s
	}

	return
}

func compCmd(acc []rune) (matches []string, longestAcc []rune) {
	if len(acc) == 0 || acc[len(acc)-1] == ' ' {
		return matches, acc
	}

	s := string(acc)
	f := strings.Fields(s)

	var longest string

	switch len(f) {
	case 0:
		longestAcc = acc
	case 1:
		words := gCmdWords
		for c, _ := range gOpts.cmds {
			words = append(words, c)
		}
		matches, longest = matchWord(s, words)
		longestAcc = []rune(longest)
	default:
		switch f[0] {
		case "set":
			matches, longest = matchWord(f[1], gOptWords)
			longestAcc = append(acc[:len(acc)-len(f[len(f)-1])], []rune(longest)...)
		case "map", "cmd":
			longestAcc = acc
		default:
			matches, longest = matchFile(f[len(f)-1])
			longestAcc = append(acc[:len(acc)-len(f[len(f)-1])], []rune(longest)...)
		}
	}

	return
}

func compShell(acc []rune) (matches []string, longestAcc []rune) {
	if len(acc) == 0 || acc[len(acc)-1] == ' ' {
		return matches, acc
	}

	s := string(acc)
	f := strings.Fields(s)

	var longest string

	switch len(f) {
	case 0:
		longestAcc = acc
	case 1:
		matches, longest = matchExec(s)
		longestAcc = []rune(longest)
	default:
		matches, longest = matchFile(f[len(f)-1])
		longestAcc = append(acc[:len(acc)-len(f[len(f)-1])], []rune(longest)...)
	}

	return
}
