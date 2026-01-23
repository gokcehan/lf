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

func getLongest(s1, s2 string) string {
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

func matchWord(s string, words []string) (matches []compMatch, longest string) {
	for _, w := range words {
		if !strings.HasPrefix(w, s) {
			continue
		}

		matches = append(matches, compMatch{w, w})
		if len(matches) == 1 {
			longest = w
		} else {
			longest = getLongest(longest, w)
		}
	}

	switch len(matches) {
	case 0:
		longest = s
	case 1:
		longest += " "
	}
	return
}

func matchList(s string, words []string) (matches []compMatch, longest string) {
	toks := strings.Split(s, ":")

	for _, w := range words {
		if slices.Contains(toks[:len(toks)-1], w) || !strings.HasPrefix(w, toks[len(toks)-1]) {
			continue
		}

		matchResult := strings.Join(append(slices.Clone(toks[:len(toks)-1]), w), ":")
		matches = append(matches, compMatch{w, matchResult})

		if len(matches) == 1 {
			longest = matchResult
		} else {
			longest = getLongest(longest, matchResult)
		}
	}

	switch len(matches) {
	case 0:
		longest = s
	case 1:
		if longest == s {
			longest += " "
		}
	}
	return
}

func matchCmd(s string) (matches []compMatch, longest string) {
	words := slices.Concat(gCmdWords, slices.Collect(maps.Keys(gOpts.cmds)))
	slices.Sort(words)
	matches, longest = matchWord(s, slices.Compact(words))
	return
}

func matchFile(s string, dirOnly bool, escape, unescape func(string) string) (matches []compMatch, longest string) {
	dir, file := filepath.Split(unescape(replaceTilde(s)))

	d := dir
	if dir == "" {
		d = "."
	}
	files, err := os.ReadDir(d)
	if err != nil {
		log.Printf("reading directory: %s", err)
		longest = s
		return
	}

	var longestName string

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
			longestName = name
		} else {
			// Match case-insensitively without changing the prefix's case.
			p := getLongest(strings.ToLower(longestName), strings.ToLower(name))
			longestName = string([]rune(longestName)[:len([]rune(p))])
		}
	}

	switch len(matches) {
	case 0:
		longest = s
	case 1:
		longest = escape(dir + longestName)
		if !strings.HasSuffix(longestName, string(filepath.Separator)) {
			longest += " "
		}
	default:
		longest = escape(dir + longestName)
	}
	return
}

func matchCmdFile(s string, dirOnly bool) (matches []compMatch, longest string) {
	matches, longest = matchFile(s, dirOnly, cmdEscape, cmdUnescape)
	return
}

func matchShellFile(s string) (matches []compMatch, longest string) {
	matches, longest = matchFile(s, false, shellEscape, shellUnescape)
	return
}

func matchSetlocalDir(s string) (matches []compMatch, longest string) {
	matches, longest = matchFile(s, true, cmdEscape, cmdUnescape)
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

	if longest != s {
		longest = trimSepEsc(longest)
	}
	return
}

func matchExec(s string) (matches []compMatch, longest string) {
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
	matches, longest = matchWord(s, slices.Compact(words))
	return
}

func matchSearch(s string) (matches []compMatch, longest string) {
	files, err := os.ReadDir(".")
	if err != nil {
		log.Printf("reading directory: %s", err)
		longest = s
		return
	}

	for _, f := range files {
		if !strings.HasPrefix(strings.ToLower(f.Name()), strings.ToLower(s)) {
			continue
		}

		matches = append(matches, compMatch{f.Name(), f.Name()})
		if len(matches) == 1 {
			longest = f.Name()
		} else {
			p := getLongest(strings.ToLower(longest), strings.ToLower(f.Name()))
			longest = string([]rune(longest)[:len([]rune(p))])
		}
	}

	if len(matches) == 0 {
		longest = s
	}
	return
}

func completeCmd(acc []rune) (matches []compMatch, longest string) {
	s := string(acc)
	f := tokenize(s)

	if len(f) == 1 {
		matches, longest = matchCmd(s)
		return
	}

	longest = f[len(f)-1]

	switch f[0] {
	case "set":
		if len(f) == 2 {
			matches, longest = matchWord(f[1], gOptWords)
			break
		}
		if len(f) != 3 {
			break
		}
		switch f[1] {
		case "cleaner", "previewer", "rulerfile":
			matches, longest = matchCmdFile(f[2], false)
		case "filtermethod", "searchmethod":
			matches, longest = matchWord(f[2], []string{"glob", "regex", "text"})
		case "info":
			matches, longest = matchList(f[2], []string{"atime", "btime", "ctime", "custom", "group", "perm", "size", "time", "user"})
		case "preserve":
			matches, longest = matchList(f[2], []string{"mode", "timestamps"})
		case "selmode":
			matches, longest = matchWord(f[2], []string{"all", "dir"})
		case "sizeunits":
			matches, longest = matchWord(f[2], []string{"binary", "decimal"})
		case "sortby":
			matches, longest = matchWord(f[2], []string{"atime", "btime", "ctime", "custom", "ext", "name", "natural", "size", "time"})
		default:
			if slices.Contains(gOptWords, f[1]+"!") {
				matches, longest = matchWord(f[2], []string{"false", "true"})
			}
		}
	case "setlocal":
		if len(f) == 2 {
			matches, longest = matchSetlocalDir(f[1])
			break
		}
		if len(f) == 3 {
			matches, longest = matchWord(f[2], gLocalOptWords)
			break
		}
		if len(f) != 4 {
			break
		}
		switch f[2] {
		case "info":
			matches, longest = matchList(f[3], []string{"atime", "btime", "ctime", "custom", "group", "perm", "size", "time", "user"})
		case "sortby":
			matches, longest = matchWord(f[3], []string{"atime", "btime", "ctime", "custom", "ext", "name", "natural", "size", "time"})
		default:
			if slices.Contains(gLocalOptWords, f[2]+"!") {
				matches, longest = matchWord(f[3], []string{"false", "true"})
			}
		}
	case "map", "nmap", "vmap", "cmap":
		if len(f) == 3 {
			matches, longest = matchCmd(f[2])
		}
	case "cmd":
	case "cd":
		if len(f) == 2 {
			matches, longest = matchCmdFile(f[1], true)
		}
	case "addcustominfo", "select", "source":
		if len(f) == 2 {
			matches, longest = matchCmdFile(f[1], false)
		}
	case "toggle":
		matches, longest = matchCmdFile(f[len(f)-1], false)
	default:
		if !slices.Contains(gCmdWords, f[0]) {
			matches, longest = matchCmdFile(f[len(f)-1], false)
		}
	}

	f[len(f)-1] = longest
	longest = strings.Join(f, " ")
	return
}

func completeShell(acc []rune) (matches []compMatch, longest string) {
	f := tokenize(string(acc))

	switch len(f) {
	case 1:
		matches, longest = matchExec(f[0])
	default:
		matches, longest = matchShellFile(f[len(f)-1])
	}

	f[len(f)-1] = longest
	longest = strings.Join(f, " ")
	return
}

func completeSearch(acc []rune) (matches []compMatch, longest string) {
	matches, longest = matchSearch(string(acc))
	return
}
