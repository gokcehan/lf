package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	gCmdWords = []string{"set", "map", "cmd"}
	gOptWords = []string{
		"preview",
		"nopreview",
		"hidden",
		"nohidden",
		"tabstop",
		"scrolloff",
		"sortby",
		"showinfo",
		"opener",
		"ratios",
	}
)

func matchWord(s string, words []string) string {
	var match string

	for _, w := range words {
		if strings.HasPrefix(w, s) {
			if match != "" {
				return s
			}
			match = w
		}
	}

	if match != "" {
		return match + " "
	}

	return s
}

func matchExec(s string) string {
	var match string

	paths := strings.Split(envPath, ":")

	for _, p := range paths {
		fi, err := ioutil.ReadDir(p)
		if err != nil {
			log.Print(err)
		}

		for _, f := range fi {
			if strings.HasPrefix(f.Name(), s) {
				if !f.Mode().IsRegular() || f.Mode()&0111 == 0 {
					continue
				}
				if match != "" {
					return s
				}
				match = f.Name()
			}
		}
	}

	if match != "" {
		return match + " "
	}

	return s
}

func matchFile(s string) string {
	var match string

	wd, err := os.Getwd()
	if err != nil {
		log.Print(err)
	}

	fi, err := ioutil.ReadDir(wd)
	if err != nil {
		log.Print(err)
	}

	for _, f := range fi {
		if strings.HasPrefix(f.Name(), s) {
			if match != "" {
				return s
			}
			match = f.Name()
		}
	}

	if match != "" {
		return match + " "
	}

	return s
}

func compCmd(acc []rune) []rune {
	if len(acc) == 0 || acc[len(acc)-1] == ' ' {
		return acc
	}

	s := string(acc)
	f := strings.Fields(s)

	switch len(f) {
	case 0: // do nothing
	case 1:
		words := gCmdWords
		for c, _ := range gOpts.cmds {
			words = append(words, c)
		}
		return []rune(matchWord(s, words))
	default:
		switch f[0] {
		case "set":
			opt := matchWord(f[1], gOptWords)
			ret := []rune(f[0])
			ret = append(ret, ' ')
			ret = append(ret, []rune(opt)...)
			return ret
		case "map", "cmd": // do nothing
		default:
			ret := []rune(f[0])
			ret = append(ret, ' ')
			for i := 1; i < len(f); i++ {
				name := matchFile(f[i])
				ret = append(ret, []rune(name)...)
			}
			return ret
		}
	}

	return acc
}

func compShell(acc []rune) []rune {
	if len(acc) == 0 || acc[len(acc)-1] == ' ' {
		return acc
	}

	s := string(acc)
	f := strings.Fields(s)

	switch len(f) {
	case 0: // do nothing
	case 1:
		return []rune(matchExec(s))
	default:
		ret := []rune(f[0])
		ret = append(ret, ' ')
		for i := 1; i < len(f); i++ {
			name := matchFile(f[i])
			ret = append(ret, []rune(name)...)
		}
		return ret
	}

	return acc
}
