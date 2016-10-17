package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	gCmdWords = []string{
		"set",
		"map",
		"cmd",
		"up",
		"half-up",
		"page-up",
		"down",
		"half-down",
		"page-down",
		"updir",
		"open",
		"quit",
		"bot",
		"top",
		"read",
		"read-shell",
		"read-shell-wait",
		"read-shell-async",
		"search",
		"search-back",
		"toggle",
		"yank",
		"delete",
		"paste",
		"renew",
		"echo",
		"cd",
		"push",
	}

	gOptWords = []string{
		"hidden",
		"nohidden",
		"hidden!",
		"preview",
		"nopreview",
		"preview!",
		"dirfirst",
		"nodirfirst",
		"dirfirst!",
		"scrolloff",
		"tabstop",
		"ifs",
		"previewer",
		"shell",
		"showinfo",
		"sortby",
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
			} else if s != "" {
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
	var words []string

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
				f, err = os.Stat(filepath.Join(p, f.Name()))
				if err != nil {
					log.Printf("getting file information: %s", err)
				}

				if !f.Mode().IsRegular() || f.Mode()&0111 == 0 {
					continue
				}

				words = append(words, f.Name())
			}
		}
	}

	sort.Strings(words)

	if len(words) > 0 {
		uniq := words[:1]
		for i := 1; i < len(words); i++ {
			if words[i] != words[i-1] {
				uniq = append(uniq, words[i])
			}
		}
		words = uniq
	}

	return matchWord(s, words)
}

func matchFile(s string) (matches []string, longest string) {
	dir := strings.Replace(s, "~", envHome, -1)

	if !filepath.IsAbs(dir) {
		wd, err := os.Getwd()
		if err != nil {
			log.Printf("getting current directory: %s", err)
		}
		dir = wd + string(filepath.Separator) + dir
	}

	dir = filepath.Dir(dir)

	fi, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("reading directory: %s", err)
	}

	for _, f := range fi {
		f, err := os.Stat(filepath.Join(dir, f.Name()))
		if err != nil {
			log.Printf("getting file information: %s", err)
			return
		}

		_, last := filepath.Split(s)
		if strings.HasPrefix(f.Name(), last) {
			name := f.Name()
			if isRoot(s) || filepath.Base(s) != s {
				name = filepath.Join(filepath.Dir(s), f.Name())
			}
			item := f.Name()
			if f.Mode().IsDir() {
				item += string(filepath.Separator)
			}
			matches = append(matches, item)
			if longest != "" {
				longest = matchLongest(longest, name)
			} else if s != "" {
				if f.Mode().IsRegular() {
					longest = name + " "
				} else {
					longest = name + string(filepath.Separator)
				}
			}
		}
	}

	if longest == "" {
		longest = s
	}

	return
}

func compCmd(acc []rune) (matches []string, longestAcc []rune) {
	s := string(acc)
	f := strings.Fields(s)

	if len(f) == 0 || s[len(s)-1] == ' ' {
		f = append(f, "")
	}

	var longest string

	switch len(f) {
	case 1:
		words := gCmdWords
		for c, _ := range gOpts.cmds {
			words = append(words, c)
		}
		sort.Strings(words)
		matches, longest = matchWord(s, words)
		longestAcc = []rune(longest)
	case 2:
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
	default:
		switch f[0] {
		case "set", "map", "cmd":
			longestAcc = acc
		default:
			matches, longest = matchFile(f[len(f)-1])
			longestAcc = append(acc[:len(acc)-len(f[len(f)-1])], []rune(longest)...)
		}
	}

	return
}

func compShell(acc []rune) (matches []string, longestAcc []rune) {
	s := string(acc)
	f := strings.Fields(s)

	if len(f) == 0 || s[len(s)-1] == ' ' {
		f = append(f, "")
	}

	var longest string

	switch len(f) {
	case 1:
		matches, longest = matchExec(s)
		longestAcc = []rune(longest)
	default:
		matches, longest = matchFile(f[len(f)-1])
		longestAcc = append(acc[:len(acc)-len(f[len(f)-1])], []rune(longest)...)
	}

	return
}
