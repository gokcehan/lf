package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type iconMap map[string]string

func parseIcons() iconMap {
	if env := os.Getenv("LF_ICONS"); env != "" {
		return parseIconsEnv(env)
	}

	defaultIcons := []string{
		"tw=🗀",
		"st=🗀",
		"ow=🗀",
		"di=🗀",
		"fi=🗎",
	}

	return parseIconsEnv(strings.Join(defaultIcons, ":"))
}

func parseIconsEnv(env string) iconMap {
	icons := make(iconMap)

	entries := strings.Split(env, ":")
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		pair := strings.Split(entry, "=")
		if len(pair) != 2 {
			log.Printf("invalid $LF_ICONS entry: %s", entry)
			return icons
		}
		key, val := pair[0], pair[1]
		icons[key] = val
	}

	return icons
}

// Maps a file to a possible key in the iconMap.
type iconKeyGenerator func(f *file) (string, bool)

// Slice of transformation functions, each of which tries to convert
// an input file to a key in iconMap. The first key that's found AND
// exists in the iconMap is used to set the icon for the current
// file.
var iconKeyGenerators []iconKeyGenerator = []iconKeyGenerator{
	// Firstly try to match the complete path (without tilde expansion).
	func(f *file) (string, bool) {
		path := f.path
		home := gUser.HomeDir
		if strings.HasPrefix(path, home) {
			path = "~" + path[len(home):]
		}
		if gOpts.iconsdir && f.IsDir() {
			path += string(os.PathSeparator)
		}
		return path, true
	},
	// The exact basename of the file.
	func(f *file) (string, bool) {
		base := filepath.Base(f.Name())
		if gOpts.iconsdir && f.IsDir() {
			base += string(os.PathSeparator)
		}
		return base, true
	},
	// The basename of the file excluding the extension
	func(f *file) (string, bool) {
		return filepath.Base(f.Name()) + ".*", true
	},
	// The file extension
	func(f *file) (string, bool) {
		return "*" + filepath.Ext(f.Name()), true
	},
	// The filetype classification
	func(f *file) (string, bool) {
		var key string

		switch {
		case f.IsDir() && f.Mode()&os.ModeSticky != 0 && f.Mode()&0002 != 0:
			key = "tw"
		case f.IsDir() && f.Mode()&os.ModeSticky != 0:
			key = "st"
		case f.IsDir() && f.Mode()&0002 != 0:
			key = "ow"
		case f.IsDir():
			key = "di"
		case f.linkState == working:
			key = "ln"
		case f.linkState == broken:
			key = "or"
		case f.Mode()&os.ModeNamedPipe != 0:
			key = "pi"
		case f.Mode()&os.ModeSocket != 0:
			key = "so"
		case f.Mode()&os.ModeCharDevice != 0:
			key = "cd"
		case f.Mode()&os.ModeDevice != 0:
			key = "bd"
		case f.Mode()&os.ModeSetuid != 0:
			key = "su"
		case f.Mode()&os.ModeSetgid != 0:
			key = "sg"
		case f.Mode().IsRegular() && f.Mode()&0111 != 0:
			key = "ex"
		default:
			return "", false
		}

		return key, true
	},
	// If all else fails, everything is a file
	func(f *file) (string, bool) {
		return "fi", true
	},
}

func (im iconMap) get(f *file) string {
	for _, gen := range iconKeyGenerators {
		if key, ok := gen(f); ok {
			if val, ok := im[key]; ok {
				return val
			}
		}
	}

	return " "
}
