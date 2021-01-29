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
		"top",
		"bottom",
		"toggle",
		"invert",
		"unselect",
		"copy",
		"cut",
		"paste",
		"clear",
		"redraw",
		"reload",
		"read",
		"rename",
		"shell",
		"shell-pipe",
		"shell-wait",
		"shell-async",
		"find",
		"find-back",
		"find-next",
		"find-prev",
		"search",
		"search-back",
		"search-next",
		"search-prev",
		"mark-save",
		"mark-remove",
		"mark-load",
		"draw",
		"load",
		"sync",
		"echo",
		"echomsg",
		"echoerr",
		"cd",
		"select",
		"glob-select",
		"glob-unselect",
		"source",
		"push",
		"delete",
	}

	gOptWords = []string{
		"anchorfind",
		"noanchorfind",
		"anchorfind!",
		"dircounts",
		"nodircounts",
		"dircounts!",
		"dirfirst",
		"nodirfirst",
		"dirfirst!",
		"drawbox",
		"nodrawbox",
		"drawbox!",
		"globsearch",
		"noglobsearch",
		"globsearch!",
		"hidden",
		"nohidden",
		"hidden!",
		"icons",
		"noicons",
		"icons!",
		"ignorecase",
		"noignorecase",
		"ignorecase!",
		"ignoredia",
		"noignoredia",
		"ignoredia!",
		"incsearch",
		"noincsearch",
		"incsearch!",
		"mouse",
		"nomouse",
		"mouse!",
		"number",
		"nonumber",
		"number!",
		"preview",
		"nopreview",
		"preview!",
		"relativenumber",
		"norelativenumber",
		"relativenumber!",
		"reverse",
		"noreverse",
		"reverse!",
		"smartcase",
		"nosmartcase",
		"smartcase!",
		"smartdia",
		"nosmartdia",
		"smartdia!",
		"wrapscan",
		"nowrapscan",
		"wrapscan!",
		"wrapscroll",
		"nowrapscroll",
		"wrapscroll!",
		"findlen",
		"period",
		"scrolloff",
		"tabstop",
		"errorfmt",
		"filesep",
		"hiddenfiles",
		"ifs",
		"info",
		"previewer",
		"cleaner",
		"promptfmt",
		"ratios",
		"shell",
		"shellopts",
		"sortby",
		"timefmt",
		"truncatechar",
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
		if !strings.HasPrefix(w, s) {
			continue
		}

		matches = append(matches, w)
		if longest != "" {
			longest = matchLongest(longest, w)
		} else if s != "" {
			longest = w + " "
		}
	}

	if longest == "" {
		longest = s
	}

	return
}

func matchExec(s string) (matches []string, longest string) {
	var words []string

	paths := strings.Split(envPath, string(filepath.ListSeparator))

	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			continue
		}

		files, err := ioutil.ReadDir(p)
		if err != nil {
			log.Printf("reading path: %s", err)
		}

		for _, f := range files {
			if !strings.HasPrefix(f.Name(), s) {
				continue
			}

			f, err = os.Stat(filepath.Join(p, f.Name()))
			if err != nil {
				log.Printf("getting file information: %s", err)
				continue
			}

			if !f.Mode().IsRegular() || !isExecutable(f) {
				continue
			}

			log.Print(f.Name())
			words = append(words, f.Name())
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
	dir := replaceTilde(s)

	if !filepath.IsAbs(dir) {
		wd, err := os.Getwd()
		if err != nil {
			log.Printf("getting current directory: %s", err)
		} else {
			dir = wd + string(filepath.Separator) + dir
		}
	}

	dir = unescape(filepath.Dir(dir))

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("reading directory: %s", err)
	}

	for _, f := range files {
		f, err := os.Stat(filepath.Join(dir, f.Name()))
		if err != nil {
			log.Printf("getting file information: %s", err)
			return
		}

		_, last := filepath.Split(s)
		if !strings.HasPrefix(escape(f.Name()), last) {
			continue
		}

		name := f.Name()
		if isRoot(s) || filepath.Base(s) != s {
			name = filepath.Join(filepath.Dir(unescape(s)), f.Name())
		}
		name = escape(name)

		item := f.Name()
		if f.Mode().IsDir() {
			item += escape(string(filepath.Separator))
		}
		matches = append(matches, item)

		if longest != "" {
			longest = matchLongest(longest, name)
		} else if s != "" {
			if f.Mode().IsRegular() {
				longest = name + " "
			} else {
				longest = name + escape(string(filepath.Separator))
			}
		}
	}

	if longest == "" {
		longest = s
	}

	return
}

func completeCmd(acc []rune) (matches []string, longestAcc []rune) {
	s := string(acc)
	f := tokenize(s)

	var longest string

	switch len(f) {
	case 1:
		words := gCmdWords
		for c := range gOpts.cmds {
			words = append(words, c)
		}
		sort.Strings(words)
		j := 0
		for i := 1; i < len(words); i++ {
			if words[j] == words[i] {
				continue
			}
			j++
			words[i], words[j] = words[j], words[i]
		}
		words = words[:j+1]
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

func completeFile(acc []rune) (matches []string, longestAcc []rune) {
	s := string(acc)

	wd, err := os.Getwd()
	if err != nil {
		log.Printf("getting current directory: %s", err)
	}

	files, err := ioutil.ReadDir(wd)
	if err != nil {
		log.Printf("reading directory: %s", err)
	}

	var longest string

	for _, f := range files {
		if !strings.HasPrefix(f.Name(), s) {
			continue
		}

		matches = append(matches, f.Name())

		if longest != "" {
			longest = matchLongest(longest, f.Name())
		} else if s != "" {
			longest = f.Name()
		}
	}

	if longest == "" {
		longest = s
	}

	longestAcc = []rune(longest)

	return
}

func completeShell(acc []rune) (matches []string, longestAcc []rune) {
	s := string(acc)
	f := tokenize(s)

	var longest string

	switch len(f) {
	case 1:
		matches, longest = matchExec(s)
		longestAcc = []rune(longest)
	default:
		matches, longest = matchFile(f[len(f)-1])
		longestAcc = append(acc[:len(acc)-len([]rune(f[len(f)-1]))], []rune(longest)...)
	}

	return
}
