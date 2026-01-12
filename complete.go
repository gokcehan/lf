package main

import (
	"log"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
)

var (
	gCmdWords = []string{
		"set",
		"setlocal",
		"map",
		"nmap",
		"vmap",
		"cmap",
		"cmd",
		"addcustominfo",
		"bottom",
		"calcdirsize",
		"cd",
		"clear",
		"clearmaps",
		"copy",
		"cut",
		"down",
		"delete",
		"draw",
		"echo",
		"echoerr",
		"echomsg",
		"filter",
		"find",
		"find-back",
		"find-next",
		"find-prev",
		"glob-select",
		"glob-unselect",
		"half-down",
		"half-up",
		"high",
		"invert",
		"jump-next",
		"jump-prev",
		"load",
		"low",
		"mark-load",
		"mark-remove",
		"mark-save",
		"middle",
		"open",
		"page-down",
		"page-up",
		"paste",
		"push",
		"quit",
		"read",
		"redraw",
		"reload",
		"rename",
		"scroll-down",
		"scroll-up",
		"search",
		"search-back",
		"search-next",
		"search-prev",
		"select",
		"setfilter",
		"shell",
		"shell-async",
		"shell-pipe",
		"shell-wait",
		"source",
		"sync",
		"tag",
		"tag-toggle",
		"toggle",
		"top",
		"tty-write",
		"unselect",
		"up",
		"updir",
		"visual",
		"visual-accept",
		"visual-change",
		"visual-discard",
		"visual-unselect",
		"cmd-capitalize-word",
		"cmd-complete",
		"cmd-delete",
		"cmd-delete-back",
		"cmd-delete-end",
		"cmd-delete-home",
		"cmd-delete-unix-word",
		"cmd-delete-word",
		"cmd-delete-word-back",
		"cmd-end",
		"cmd-enter",
		"cmd-escape",
		"cmd-history-next",
		"cmd-history-prev",
		"cmd-home",
		"cmd-interrupt",
		"cmd-left",
		"cmd-lowercase-word",
		"cmd-menu-accept",
		"cmd-menu-complete",
		"cmd-menu-complete-back",
		"cmd-menu-discard",
		"cmd-right",
		"cmd-transpose",
		"cmd-transpose-word",
		"cmd-uppercase-word",
		"cmd-word",
		"cmd-word-back",
		"cmd-yank",
	}

	gOptWords      = getOptWords(gOpts)
	gLocalOptWords = getLocalOptWords(gLocalOpts)
)

func getOptWords(opts any) (optWords []string) {
	t := reflect.TypeOf(opts)
	for i := range t.NumField() {
		field := t.Field(i)
		switch field.Type.Kind() {
		case reflect.Map:
			continue
		case reflect.Bool:
			name := field.Name
			optWords = append(optWords, name, "no"+name, name+"!")
		default:
			optWords = append(optWords, field.Name)
		}
	}
	sort.Strings(optWords)
	return
}

func getLocalOptWords(localOpts any) (localOptWords []string) {
	t := reflect.TypeOf(localOpts)
	for i := range t.NumField() {
		field := t.Field(i)
		name := field.Name
		if field.Type.Kind() != reflect.Map {
			continue
		}
		if field.Type.Elem().Kind() == reflect.Bool {
			localOptWords = append(localOptWords, name, "no"+name, name+"!")
		} else {
			localOptWords = append(localOptWords, name)
		}
	}
	sort.Strings(localOptWords)
	return
}

func commonPrefix(s1, s2 string) string {
	r1 := []rune(s1)
	r2 := []rune(s2)

	i := 0
	for ; i < len(r1) && i < len(r2); i++ {
		if r1[i] != r2[i] {
			break
		}
	}

	return string(r1[:i])
}

type compMatch struct {
	name   string // display name in completion menu
	result string // result when cycling through completion menu
}

func matchWord(s string, words []string) (matches []compMatch, result string) {
	for _, w := range words {
		if !strings.HasPrefix(w, s) {
			continue
		}

		matches = append(matches, compMatch{w, w})
		if len(matches) == 1 {
			result = w
		} else {
			result = commonPrefix(result, w)
		}
	}

	switch len(matches) {
	case 0:
		result = s
	case 1:
		result += " "
	}
	return
}

func matchList(s string, words []string) (matches []compMatch, result string) {
	toks := strings.Split(s, ":")

	for _, w := range words {
		if slices.Contains(toks[:len(toks)-1], w) || !strings.HasPrefix(w, toks[len(toks)-1]) {
			continue
		}

		matchResult := strings.Join(append(slices.Clone(toks[:len(toks)-1]), w), ":")
		matches = append(matches, compMatch{w, matchResult})

		if len(matches) == 1 {
			result = matchResult
		} else {
			result = commonPrefix(result, matchResult)
		}
	}

	switch len(matches) {
	case 0:
		result = s
	case 1:
		if result == s {
			result += " "
		}
	}
	return
}

func matchCmd(s string) (matches []compMatch, result string) {
	words := slices.Concat(gCmdWords, slices.Collect(maps.Keys(gOpts.cmds)))
	slices.Sort(words)
	matches, result = matchWord(s, slices.Compact(words))
	return
}

func matchFile(s string, dirOnly bool, escape, unescape func(string) string) (matches []compMatch, result string) {
	dir, file := filepath.Split(unescape(replaceTilde(s)))

	d := dir
	if dir == "" {
		d = "."
	}
	files, err := os.ReadDir(d)
	if err != nil {
		log.Printf("reading directory: %s", err)
		result = s
		return
	}

	var commonName string

	for _, f := range files {
		isDir := false
		if f.IsDir() {
			isDir = true
		} else if f.Type()&os.ModeSymlink != 0 {
			if stat, err := os.Stat(filepath.Join(d, f.Name())); err == nil && stat.IsDir() {
				isDir = true
			}
		}

		if !isDir && dirOnly {
			continue
		}

		if !strings.HasPrefix(strings.ToLower(f.Name()), strings.ToLower(file)) {
			continue
		}

		name := f.Name()
		if isDir {
			name += string(filepath.Separator)
		}
		matches = append(matches, compMatch{name, escape(dir + name)})

		if len(matches) == 1 {
			commonName = name
		} else {
			// Match case-insensitively without changing the prefix's case.
			p := commonPrefix(strings.ToLower(commonName), strings.ToLower(name))
			commonName = string([]rune(commonName)[:len([]rune(p))])
		}
	}

	switch len(matches) {
	case 0:
		result = s
	case 1:
		result = escape(dir + commonName)
		if !strings.HasSuffix(commonName, string(filepath.Separator)) {
			result += " "
		}
	default:
		result = escape(dir + commonName)
	}
	return
}

func matchCmdFile(s string, dirOnly bool) (matches []compMatch, result string) {
	matches, result = matchFile(s, dirOnly, cmdEscape, cmdUnescape)
	return
}

func matchShellFile(s string) (matches []compMatch, result string) {
	matches, result = matchFile(s, false, shellEscape, shellUnescape)
	return
}

func matchSetlocalDir(s string) (matches []compMatch, result string) {
	matches, result = matchFile(s, true, cmdEscape, cmdUnescape)
	if len(matches) == 0 {
		return
	}

	trimSep := func(path string) string {
		return strings.TrimSuffix(path, string(filepath.Separator))
	}

	trimSepEsc := func(path string) string {
		return cmdEscape(trimSep(cmdUnescape(path)))
	}

	// add separate matches for path and recursive path
	tmp := make([]compMatch, 0, len(matches)*2)
	for _, match := range matches {
		trimmedMatch := compMatch{trimSep(match.name), trimSepEsc(match.result)}
		tmp = append(tmp, trimmedMatch, match)
	}
	matches = tmp

	if result != s {
		result = trimSepEsc(result)
	}
	return
}

func matchExec(s string) (matches []compMatch, result string) {
	var words []string
	for p := range strings.SplitSeq(envPath, string(filepath.ListSeparator)) {
		files, err := os.ReadDir(p)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("reading path: %s", err)
			}
			continue
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

			if finfo.Mode().IsRegular() && isExecutable(finfo) {
				words = append(words, f.Name())
			}
		}
	}

	slices.Sort(words)
	matches, result = matchWord(s, slices.Compact(words))
	return
}

func matchSearch(s string) (matches []compMatch, result string) {
	files, err := os.ReadDir(".")
	if err != nil {
		log.Printf("reading directory: %s", err)
		result = s
		return
	}

	for _, f := range files {
		if !strings.HasPrefix(strings.ToLower(f.Name()), strings.ToLower(s)) {
			continue
		}

		matches = append(matches, compMatch{f.Name(), f.Name()})
		if len(matches) == 1 {
			result = f.Name()
		} else {
			p := commonPrefix(strings.ToLower(result), strings.ToLower(f.Name()))
			result = string([]rune(result)[:len([]rune(p))])
		}
	}

	if len(matches) == 0 {
		result = s
	}
	return
}

func completeCmd(acc []rune) (matches []compMatch, result string) {
	s := string(acc)
	f := tokenize(s)

	if len(f) == 1 {
		matches, result = matchCmd(s)
		return
	}

	result = f[len(f)-1]

	switch f[0] {
	case "set":
		if len(f) == 2 {
			matches, result = matchWord(f[1], gOptWords)
			break
		}
		if len(f) != 3 {
			break
		}
		switch f[1] {
		case "cleaner", "previewer", "rulerfile":
			matches, result = matchCmdFile(f[2], false)
		case "filtermethod", "searchmethod":
			matches, result = matchWord(f[2], []string{"glob", "regex", "text"})
		case "info":
			matches, result = matchList(f[2], []string{"atime", "btime", "ctime", "custom", "group", "perm", "size", "time", "user"})
		case "preserve":
			matches, result = matchList(f[2], []string{"mode", "timestamps"})
		case "selmode":
			matches, result = matchWord(f[2], []string{"all", "dir"})
		case "sizeunits":
			matches, result = matchWord(f[2], []string{"binary", "decimal"})
		case "sortby":
			matches, result = matchWord(f[2], []string{"atime", "btime", "ctime", "custom", "ext", "name", "natural", "size", "time"})
		default:
			if slices.Contains(gOptWords, f[1]+"!") {
				matches, result = matchWord(f[2], []string{"false", "true"})
			}
		}
	case "setlocal":
		if len(f) == 2 {
			matches, result = matchSetlocalDir(f[1])
			break
		}
		if len(f) == 3 {
			matches, result = matchWord(f[2], gLocalOptWords)
			break
		}
		if len(f) != 4 {
			break
		}
		switch f[2] {
		case "info":
			matches, result = matchList(f[3], []string{"atime", "btime", "ctime", "custom", "group", "perm", "size", "time", "user"})
		case "sortby":
			matches, result = matchWord(f[3], []string{"atime", "btime", "ctime", "custom", "ext", "name", "natural", "size", "time"})
		default:
			if slices.Contains(gLocalOptWords, f[2]+"!") {
				matches, result = matchWord(f[3], []string{"false", "true"})
			}
		}
	case "map", "nmap", "vmap", "cmap":
		if len(f) == 3 {
			matches, result = matchCmd(f[2])
		}
	case "cmd":
	case "cd":
		if len(f) == 2 {
			matches, result = matchCmdFile(f[1], true)
		}
	case "addcustominfo", "select", "source":
		if len(f) == 2 {
			matches, result = matchCmdFile(f[1], false)
		}
	case "toggle":
		matches, result = matchCmdFile(f[len(f)-1], false)
	default:
		if !slices.Contains(gCmdWords, f[0]) {
			matches, result = matchCmdFile(f[len(f)-1], false)
		}
	}

	f[len(f)-1] = result
	result = strings.Join(f, " ")
	return
}

func completeShell(acc []rune) (matches []compMatch, result string) {
	f := tokenize(string(acc))

	switch len(f) {
	case 1:
		matches, result = matchExec(f[0])
	default:
		matches, result = matchShellFile(f[len(f)-1])
	}

	f[len(f)-1] = result
	result = strings.Join(f, " ")
	return
}

func completeSearch(acc []rune) (matches []compMatch, result string) {
	matches, result = matchSearch(string(acc))
	return
}
