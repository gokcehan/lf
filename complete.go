package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	gCmdWords = []string{
		"set",
		"setlocal",
		"map",
		"cmap",
		"cmd",
		"quit",
		"up",
		"half-up",
		"page-up",
		"scroll-up",
		"down",
		"half-down",
		"page-down",
		"scroll-down",
		"updir",
		"open",
		"jump-next",
		"jump-prev",
		"top",
		"bottom",
		"high",
		"middle",
		"low",
		"toggle",
		"invert",
		"invert-below",
		"unselect",
		"glob-select",
		"glob-unselect",
		"calcdirsize",
		"clearmaps",
		"copy",
		"cut",
		"paste",
		"clear",
		"sync",
		"draw",
		"redraw",
		"load",
		"reload",
		"echo",
		"echomsg",
		"echoerr",
		"cd",
		"select",
		"delete",
		"rename",
		"source",
		"push",
		"read",
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
		"filter",
		"setfilter",
		"mark-save",
		"mark-load",
		"mark-remove",
		"tag",
		"tag-toggle",
		"cmd-escape",
		"cmd-complete",
		"cmd-menu-complete",
		"cmd-menu-complete-back",
		"cmd-menu-accept",
		"cmd-enter",
		"cmd-interrupt",
		"cmd-history-next",
		"cmd-history-prev",
		"cmd-left",
		"cmd-right",
		"cmd-home",
		"cmd-end",
		"cmd-delete",
		"cmd-delete-back",
		"cmd-delete-home",
		"cmd-delete-end",
		"cmd-delete-unix-word",
		"cmd-yank",
		"cmd-transpose",
		"cmd-transpose-word",
		"cmd-word",
		"cmd-word-back",
		"cmd-delete-word",
		"cmd-delete-word-back",
		"cmd-capitalize-word",
		"cmd-uppercase-word",
		"cmd-lowercase-word",
	}

	gOptWords = []string{
		"anchorfind",
		"noanchorfind",
		"anchorfind!",
		"autoquit",
		"noautoquit",
		"autoquit!",
		"borderfmt",
		"cursoractivefmt",
		"cursorparentfmt",
		"cursorpreviewfmt",
		"dircache",
		"nodircache",
		"dircache!",
		"dircounts",
		"nodircounts",
		"dircounts!",
		"dirfirst",
		"nodirfirst",
		"dirfirst!",
		"dironly",
		"nodironly",
		"dironly!",
		"dirpreviews",
		"nodirpreviews",
		"dirpreviews!",
		"drawbox",
		"nodrawbox",
		"drawbox!",
		"dupfilefmt",
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
		"incfilter",
		"noincfilter",
		"incfilter!",
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
		"ruler",
		"rulerfmt",
		"preserve",
		"sixel",
		"nosixel",
		"sixel!",
		"smartcase",
		"nosmartcase",
		"smartcase!",
		"smartdia",
		"nosmartdia",
		"smartdia!",
		"waitmsg",
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
		"history",
		"ifs",
		"info",
		"numberfmt",
		"previewer",
		"cleaner",
		"promptfmt",
		"ratios",
		"selmode",
		"shell",
		"shellflag",
		"shellopts",
		"sortby",
		"statfmt",
		"timefmt",
		"tempmarks",
		"tagfmt",
		"infotimefmtnew",
		"infotimefmtold",
		"truncatechar",
		"truncatepct",
	}

	gLocalOptWords = []string{
		"dirfirst",
		"nodirfirst",
		"dirfirst!",
		"dironly",
		"nodironly",
		"dironly!",
		"hidden",
		"nohidden",
		"hidden!",
		"info",
		"reverse",
		"noreverse",
		"reverse!",
		"sortby",
	}
)

func matchLongest(s1, s2 []rune) []rune {
	i := 0
	for ; i < len(s1) && i < len(s2); i++ {
		if s1[i] != s2[i] {
			break
		}
	}
	return s1[:i]
}

func matchWord(s string, words []string) (matches []string, longest []rune) {
	for _, w := range words {
		if !strings.HasPrefix(w, s) {
			continue
		}

		matches = append(matches, w)
		if len(longest) != 0 {
			longest = matchLongest(longest, []rune(w))
		} else if s != "" {
			longest = []rune(w + " ")
		}
	}

	if len(longest) == 0 {
		longest = []rune(s)
	}

	return
}

func matchExec(s string) (matches []string, longest []rune) {
	var words []string

	paths := strings.Split(envPath, string(filepath.ListSeparator))

	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			continue
		}

		files, err := os.ReadDir(p)
		if err != nil {
			log.Printf("reading path: %s", err)
		}

		for _, f := range files {
			if !strings.HasPrefix(f.Name(), s) {
				continue
			}

			finfo, err := f.Info()
			if err != nil {
				log.Printf("getting file information: %s", err)
				continue
			}

			if !finfo.Mode().IsRegular() || !isExecutable(finfo) {
				continue
			}

			words = append(words, finfo.Name())
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

func matchFile(s string) (matches []string, longest []rune) {
	dir := replaceTilde(s)

	if !filepath.IsAbs(dir) {
		wd, err := os.Getwd()
		if err != nil {
			log.Printf("getting current directory: %s", err)
		} else {
			dir = wd + string(filepath.Separator) + dir
		}
	}

	dir = filepath.Dir(unescape(dir))

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("reading directory: %s", err)
	}

	for _, f := range files {
		name := filepath.Join(dir, f.Name())
		f, err := os.Lstat(name)
		if err != nil {
			log.Printf("getting file information: %s", err)
			continue
		}

		name = strings.ToLower(escape(f.Name()))
		_, last := filepath.Split(s)
		if !strings.HasPrefix(name, strings.ToLower(last)) {
			continue
		}

		name = f.Name()
		if isRoot(s) || filepath.Base(s) != s {
			name = filepath.Join(filepath.Dir(unescape(s)), f.Name())
		}
		name = escape(replaceTilde(name))

		item := f.Name()
		if f.Mode().IsDir() {
			item += escape(string(filepath.Separator))
		}
		matches = append(matches, item)

		if longest == nil {
			if f.Mode().IsRegular() {
				longest = []rune(name + " ")
			} else {
				longest = []rune(name + escape(string(filepath.Separator)))
			}
		} else {
			longest = matchLongest(longest, []rune(name))
		}
	}

	if len(longest) < len([]rune(s)) {
		longest = []rune(s)
	}

	return
}

func matchCmd(s string) (matches []string, longest []rune) {
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
	return
}

func completeCmd(acc []rune) (matches []string, longestAcc []rune) {
	s := string(acc)
	f := tokenize(s)

	var longest []rune

	switch len(f) {
	case 1:
		matches, longestAcc = matchCmd(s)
	case 2:
		switch f[0] {
		case "set":
			matches, longest = matchWord(f[1], gOptWords)
			longestAcc = append(acc[:len(acc)-len([]rune(f[len(f)-1]))], longest...)
		case "map", "cmap", "cmd":
			longestAcc = acc
		default:
			matches, longest = matchFile(f[len(f)-1])
			longestAcc = append(acc[:len(acc)-len([]rune(f[len(f)-1]))], longest...)
		}
	case 3:
		switch f[0] {
		case "setlocal":
			matches, longest = matchWord(f[2], gLocalOptWords)
			longestAcc = append(acc[:len(acc)-len([]rune(f[len(f)-1]))], longest...)
		case "map", "cmap":
			matches, longest = matchCmd(f[2])
			longestAcc = append(acc[:len(acc)-len([]rune(f[len(f)-1]))], longest...)
		default:
			matches, longest = matchFile(f[len(f)-1])
			longestAcc = append(acc[:len(acc)-len([]rune(f[len(f)-1]))], longest...)
		}
	default:
		switch f[0] {
		case "set", "setlocal", "map", "cmap", "cmd":
			longestAcc = acc
		default:
			matches, longest = matchFile(f[len(f)-1])
			longestAcc = append(acc[:len(acc)-len([]rune(f[len(f)-1]))], longest...)
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

	files, err := os.ReadDir(wd)
	if err != nil {
		log.Printf("reading directory: %s", err)
	}

	for _, f := range files {
		if !strings.HasPrefix(strings.ToLower(f.Name()), strings.ToLower(s)) {
			continue
		}

		matches = append(matches, f.Name())

		if longestAcc == nil {
			longestAcc = []rune(f.Name())
		} else {
			longestAcc = matchLongest(longestAcc, []rune(f.Name()))
		}
	}

	if len(longestAcc) < len(acc) {
		longestAcc = acc
	}

	return
}

func completeShell(acc []rune) (matches []string, longestAcc []rune) {
	s := string(acc)
	f := tokenize(s)

	var longest []rune

	switch len(f) {
	case 1:
		matches, longestAcc = matchExec(s)
	default:
		matches, longest = matchFile(f[len(f)-1])
		longestAcc = append(acc[:len(acc)-len([]rune(f[len(f)-1]))], longest...)
	}

	return
}
