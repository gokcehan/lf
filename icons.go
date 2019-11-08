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
		// directory name in english
		".git=",
		"Desktop=",
		"Documents=",
		"Downloads=",
		"Dropbox=",
		"Music=",
		"Pictures=",
		"Public=",
		"Templates=",
		"Videos=",
		// file extensions
		".7z=",
		".ai=",
		".apk=",
		".avi=",
		".bat=",
		".bat=",
		".bmp=",
		".bz2=",
		".c++=",
		".c=",
		".cab=",
		".cc=",
		".clj=",
		".cljc=",
		".cljs=",
		".cmd=",
		".coffee=",
		".conf=",
		".cp=",
		".cpio=",
		".cpp=",
		".css=",
		".cxx=",
		".d=",
		".dart=",
		".db=",
		".deb=",
		".diff=",
		".dump=",
		".edn=",
		".ejs=",
		".epub=",
		".erl=",
		".f#=",
		".fish=",
		".flac=",
		".flv=",
		".fs=",
		".fsi=",
		".fsscript=",
		".fsx=",
		".gem=",
		".gif=",
		".go=",
		".gz=",
		".gzip=",
		".hbs=",
		".hrl=",
		".hs=",
		".htm=",
		".html=",
		".ico=",
		".ini=",
		".java=",
		".jl=",
		".jpeg=",
		".jpg=",
		".js=",
		".json=",
		".jsx=",
		".less=",
		".lha=",
		".lhs=",
		".log=",
		".lua=",
		".lzh=",
		".lzma=",
		".markdown=",
		".md=",
		".mkv=",
		".ml=λ",
		".mli=λ",
		".mov=",
		".mp3=",
		".mp4=",
		".mpeg=",
		".mpg=",
		".mustache=",
		".ogg=",
		".pdf=",
		".php=",
		".pl=",
		".pm=",
		".png=",
		".ps1=",
		".psb=",
		".psd=",
		".py=",
		".pyc=",
		".pyd=",
		".pyo=",
		".rar=",
		".rb=",
		".rc=",
		".reg=",
		".rlib=",
		".rpm=",
		".rs=",
		".rss=",
		".scala=",
		".scss=",
		".sh=",
		".slim=",
		".sln=",
		".sql=",
		".styl=",
		".suo=",
		".t=",
		".tar=",
		".tgz=",
		".ts=",
		".twig=",
		".vim=",
		".vimrc=",
		".wav=",
		".xml=",
		".xul=",
		".xz=",
		".yml=",
		".zip=",
		".zsh=",
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

func (im iconMap) get(f *file) string {
	var file, folder string
	var exist bool

	if file, exist = im[filepath.Ext(f.Name())]; !f.IsDir() && exist {
		return file
	} else if !f.IsDir() && !exist {
		// falback icons for unknown extension
		return ""
	}

	if folder, exist = im[f.Name()]; f.IsDir() && exist {
		return folder
	}

	// just return icons of a folder
	return ""
}
