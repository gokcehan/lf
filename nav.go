package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type LinkState int8

const (
	NotLink LinkState = iota
	Working
	Broken
)

type File struct {
	os.FileInfo
	LinkState LinkState
	Path      string
}

type ByName []*File

func (a ByName) Len() int      { return len(a) }
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a ByName) Less(i, j int) bool {
	return strings.ToLower(a[i].Name()) < strings.ToLower(a[j].Name())
}

type BySize []*File

func (a BySize) Len() int           { return len(a) }
func (a BySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySize) Less(i, j int) bool { return a[i].Size() < a[j].Size() }

type ByTime []*File

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].ModTime().Before(a[j].ModTime()) }

type ByDir []*File

func (a ByDir) Len() int      { return len(a) }
func (a ByDir) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a ByDir) Less(i, j int) bool {
	if a[i].IsDir() == a[j].IsDir() {
		return i < j
	}
	return a[i].IsDir()
}

type ByNum []*File

func (a ByNum) Len() int      { return len(a) }
func (a ByNum) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a ByNum) Less(i, j int) bool {
	nums1, rest1, numFirst1 := extractNums(a[i].Name())
	nums2, rest2, numFirst2 := extractNums(a[j].Name())

	if numFirst1 != numFirst2 {
		return i < j
	}

	if numFirst1 {
		if nums1[0] != nums2[0] {
			return nums1[0] < nums2[0]
		}
		nums1 = nums1[1:]
		nums2 = nums2[1:]
	}

	for k := 0; k < len(nums1) && k < len(nums2); k++ {
		if rest1[k] != rest2[k] {
			return i < j
		}
		if nums1[k] != nums2[k] {
			return nums1[k] < nums2[k]
		}
	}

	return i < j
}

func getFilesSorted(path string) []*File {
	fi, err := readdir(path)
	if err != nil {
		log.Printf("reading directory: %s", err)
	}

	switch gOpts.sortby {
	case "name":
		sort.Sort(ByName(fi))
	case "size":
		sort.Sort(BySize(fi))
	case "time":
		sort.Sort(ByTime(fi))
	default:
		log.Printf("unknown sorting type: %s", gOpts.sortby)
	}

	if gOpts.dirfirst {
		sort.Stable(ByDir(fi))
	}
	//TODO this should be optional
	sort.Stable(ByNum(fi))

	return fi
}

func readdir(path string) ([]*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	names, err := f.Readdirnames(-1)
	fi := make([]*File, 0, len(names))
	for _, filename := range names {
		if !gOpts.hidden && filename[0] == '.' {
			continue
		}

		fpath := filepath.Join(path, filename)

		lstat, lerr := os.Lstat(fpath)
		if os.IsNotExist(lerr) {
			continue
		}
		if lerr != nil {
			return fi, lerr
		}

		var linkState LinkState

		if lstat.Mode()&os.ModeSymlink != 0 {
			stat, serr := os.Stat(fpath)
			if serr == nil {
				linkState = Working
				lstat = stat
			} else {
				linkState = Broken
				log.Printf("getting link destination info: %s", serr)
			}
		}

		fi = append(fi, &File{
			FileInfo:  lstat,
			LinkState: linkState,
			Path:      fpath,
		})
	}
	return fi, err
}

type Dir struct {
	ind  int // which entry is highlighted
	pos  int // which line in the ui highlighted entry is
	path string
	fi   []*File
}

func newDir(path string) *Dir {
	return &Dir{
		path: path,
		fi:   getFilesSorted(path),
	}
}

func (dir *Dir) renew(height int) {
	var name string
	if len(dir.fi) != 0 {
		name = dir.fi[dir.ind].Name()
	}

	dir.fi = getFilesSorted(dir.path)

	dir.load(dir.ind, dir.pos, height, name)
}

func (dir *Dir) load(ind, pos, height int, name string) {
	if len(dir.fi) == 0 {
		dir.ind, dir.pos = 0, 0
		return
	}

	ind = max(0, min(ind, len(dir.fi)-1))

	if dir.fi[ind].Name() != name {
		for i, f := range dir.fi {
			if f.Name() == name {
				ind = i
				break
			}
		}

		edge := min(gOpts.scrolloff, len(dir.fi)-ind-1)
		pos = min(ind, height-edge-1)
	}

	dir.ind = ind
	dir.pos = pos
}

type Nav struct {
	dirs   []*Dir
	inds   map[string]int
	poss   map[string]int
	names  map[string]string
	marks  map[string]bool
	height int
}

func getDirs(wd string, height int) []*Dir {
	var dirs []*Dir

	for curr, base := wd, ""; !isRoot(base); curr, base = filepath.Dir(curr), filepath.Base(curr) {
		dir := newDir(curr)
		for i, f := range dir.fi {
			if f.Name() == base {
				dir.ind = i
				edge := min(gOpts.scrolloff, len(dir.fi)-dir.ind-1)
				dir.pos = min(i, height-edge-1)
				break
			}
		}
		dirs = append(dirs, dir)
	}

	for i, j := 0, len(dirs)-1; i < j; i, j = i+1, j-1 {
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}

	return dirs
}

func newNav(height int) *Nav {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("getting current directory: %s", err)
	}

	dirs := getDirs(wd, height)

	return &Nav{
		dirs:   dirs,
		inds:   make(map[string]int),
		poss:   make(map[string]int),
		names:  make(map[string]string),
		marks:  make(map[string]bool),
		height: height,
	}
}

func (nav *Nav) renew(height int) {
	nav.height = height
	for _, d := range nav.dirs {
		d.renew(nav.height)
	}

	for m := range nav.marks {
		if _, err := os.Stat(m); os.IsNotExist(err) {
			delete(nav.marks, m)
		}
	}
}

func (nav *Nav) up(dist int) {
	dir := nav.currDir()

	if dir.ind == 0 {
		return
	}

	dir.ind -= dist
	dir.ind = max(0, dir.ind)

	dir.pos -= dist
	edge := min(gOpts.scrolloff, dir.ind)
	dir.pos = max(dir.pos, edge)
}

func (nav *Nav) down(dist int) {
	dir := nav.currDir()

	maxind := len(dir.fi) - 1

	if dir.ind >= maxind {
		return
	}

	dir.ind += dist
	dir.ind = min(maxind, dir.ind)

	dir.pos += dist
	edge := min(gOpts.scrolloff, maxind-dir.ind)

	// use a smaller value when the height is even and scrolloff is maxed
	// in order to stay at the same row as much as possible while up/down
	edge = min(edge, nav.height/2+nav.height%2-1)

	dir.pos = min(dir.pos, nav.height-edge-1)
	dir.pos = min(dir.pos, maxind)
}

func (nav *Nav) updir() error {
	if len(nav.dirs) <= 1 {
		return nil
	}

	dir := nav.currDir()

	nav.inds[dir.path] = dir.ind
	nav.poss[dir.path] = dir.pos

	if len(dir.fi) != 0 {
		nav.names[dir.path] = dir.fi[dir.ind].Name()
	}

	nav.dirs = nav.dirs[:len(nav.dirs)-1]

	if err := os.Chdir(filepath.Dir(dir.path)); err != nil {
		return fmt.Errorf("updir: %s", err)
	}

	return nil
}

var ErrNotDir = fmt.Errorf("not a directory")

func (nav *Nav) open() error {
	curr, err := nav.currFile()
	if err != nil {
		return err
	}
	if !curr.IsDir() {
		return ErrNotDir
	}
	path := curr.Path

	dir := newDir(path)

	dir.load(nav.inds[path], nav.poss[path], nav.height, nav.names[path])

	nav.dirs = append(nav.dirs, dir)

	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("open: %s", err)
	}

	return nil
}

func (nav *Nav) bot() {
	dir := nav.currDir()

	dir.ind = len(dir.fi) - 1
	dir.pos = min(dir.ind, nav.height-1)
}

func (nav *Nav) top() {
	dir := nav.currDir()

	dir.ind = 0
	dir.pos = 0
}

func (nav *Nav) cd(wd string) error {
	wd = strings.Replace(wd, "~", envHome, -1)
	wd = filepath.Clean(wd)

	if !filepath.IsAbs(wd) {
		wd = filepath.Join(nav.currDir().path, wd)
	}

	if err := os.Chdir(wd); err != nil {
		return fmt.Errorf("cd: %s", err)
	}

	nav.dirs = getDirs(wd, nav.height)

	// TODO: save/load ind and pos from the map

	return nil
}

func (nav *Nav) toggleMark(path string) {
	if nav.marks[path] {
		delete(nav.marks, path)
	} else {
		nav.marks[path] = true
	}
}

func (nav *Nav) toggle() {
	curr, err := nav.currFile()
	if err != nil {
		return
	}

	nav.toggleMark(curr.Path)

	nav.down(1)
}

func (nav *Nav) invert() {
	last := nav.currDir()
	for _, f := range last.fi {
		path := filepath.Join(last.path, f.Name())
		nav.toggleMark(path)
	}
}

func (nav *Nav) save(copy bool) error {
	if len(nav.marks) == 0 {
		curr, err := nav.currFile()
		if err != nil {
			return errors.New("no file selected")
		}

		if err := saveFiles([]string{curr.Path}, copy); err != nil {
			return err
		}
	} else {
		var fs []string
		for f := range nav.marks {
			fs = append(fs, f)
		}

		if err := saveFiles(fs, copy); err != nil {
			return err
		}
	}

	return nil
}

func (nav *Nav) put() error {
	list, copy, err := loadFiles()
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return errors.New("no file in yank/delete buffer")
	}

	dir := nav.currDir()

	var args []string

	var sh string
	if copy {
		sh = "cp"
		args = append(args, "-r")
	} else {
		sh = "mv"
	}

	args = append(args, "--backup=numbered")
	args = append(args, list...)
	args = append(args, dir.path)

	cmd := exec.Command(sh, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", sh, err)
	}

	// TODO: async?

	return nil
}

func (nav *Nav) currDir() *Dir {
	return nav.dirs[len(nav.dirs)-1]
}

func (nav *Nav) currFile() (*File, error) {
	last := nav.dirs[len(nav.dirs)-1]

	if len(last.fi) == 0 {
		return nil, fmt.Errorf("empty directory")
	}
	return last.fi[last.ind], nil
}

func (nav *Nav) currMarks() []string {
	marks := make([]string, 0, len(nav.marks))
	for m := range nav.marks {
		marks = append(marks, m)
	}
	return marks
}
