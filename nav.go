package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type linkState byte

const (
	notLink linkState = iota
	working
	broken
)

type file struct {
	os.FileInfo
	LinkState linkState
	Path      string
	Count     int
}

type filesSortable struct {
	files []*file
	less  func(i, j int) bool
}

func (f filesSortable) Len() int           { return len(f.files) }
func (f filesSortable) Swap(i, j int)      { f.files[i], f.files[j] = f.files[j], f.files[i] }
func (f filesSortable) Less(i, j int) bool { return f.less(i, j) }

// TODO: Replace with `sort.SliceStable` once available
func sortFilesStable(files []*file, less func(i, j int) bool) {
	sort.Stable(filesSortable{files: files, less: less})
}

func getFilesSorted(path string) []*file {
	fi, err := readdir(path)
	if err != nil {
		log.Printf("reading directory: %s", err)
	}

	switch gOpts.sortby {
	case "natural":
		sortFilesStable(fi, func(i, j int) bool {
			return naturalLess(strings.ToLower(fi[i].Name()), strings.ToLower(fi[j].Name()))
		})
	case "name":
		sortFilesStable(fi, func(i, j int) bool {
			return strings.ToLower(fi[i].Name()) < strings.ToLower(fi[j].Name())
		})
	case "size":
		sortFilesStable(fi, func(i, j int) bool {
			return fi[i].Size() < fi[j].Size()
		})
	case "time":
		sortFilesStable(fi, func(i, j int) bool {
			return fi[i].ModTime().Before(fi[j].ModTime())
		})
	default:
		log.Printf("unknown sorting type: %s", gOpts.sortby)
	}

	if gOpts.reverse {
		for i, j := 0, len(fi)-1; i < j; i, j = i+1, j-1 {
			fi[i], fi[j] = fi[j], fi[i]
		}
	}

	if gOpts.dirfirst {
		sortFilesStable(fi, func(i, j int) bool {
			if fi[i].IsDir() == fi[j].IsDir() {
				return i < j
			}
			return fi[i].IsDir()
		})
	}

	return fi
}

func readdir(path string) ([]*file, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	names, err := f.Readdirnames(-1)
	f.Close()

	fi := make([]*file, 0, len(names))
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

		var linkState linkState

		if lstat.Mode()&os.ModeSymlink != 0 {
			stat, serr := os.Stat(fpath)
			if serr == nil {
				linkState = working
				lstat = stat
			} else {
				linkState = broken
				log.Printf("getting link destination info: %s", serr)
			}
		}

		fi = append(fi, &file{
			FileInfo:  lstat,
			LinkState: linkState,
			Path:      fpath,
			Count:     -1,
		})
	}
	return fi, err
}

type dir struct {
	ind  int // which entry is highlighted
	pos  int // which line in the ui highlighted entry is
	path string
	fi   []*file
}

func newDir(path string) *dir {
	return &dir{
		path: path,
		fi:   getFilesSorted(path),
	}
}

func (dir *dir) renew(height int) {
	var name string
	if len(dir.fi) != 0 {
		name = dir.fi[dir.ind].Name()
	}

	dir.fi = getFilesSorted(dir.path)

	dir.load(dir.ind, dir.pos, height, name)
}

func (dir *dir) load(ind, pos, height int, name string) {
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

type nav struct {
	dirs    []*dir
	inds    map[string]int
	poss    map[string]int
	names   map[string]string
	marks   map[string]int
	saves   map[string]bool
	markInd int
	height  int
	search  string
}

func getDirs(wd string, height int) []*dir {
	var dirs []*dir

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

func newNav(height int) *nav {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("getting current directory: %s", err)
	}

	dirs := getDirs(wd, height)

	return &nav{
		dirs:    dirs,
		inds:    make(map[string]int),
		poss:    make(map[string]int),
		names:   make(map[string]string),
		marks:   make(map[string]int),
		saves:   make(map[string]bool),
		markInd: 0,
		height:  height,
	}
}

func (nav *nav) renew(height int) {
	nav.height = height
	for _, d := range nav.dirs {
		d.renew(nav.height)
	}

	for m := range nav.marks {
		if _, err := os.Stat(m); os.IsNotExist(err) {
			delete(nav.marks, m)
		}
	}
	if len(nav.marks) == 0 {
		nav.markInd = 0
	}
}

func (nav *nav) up(dist int) {
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

func (nav *nav) down(dist int) {
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

func (nav *nav) updir() error {
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

func (nav *nav) open() error {
	curr, err := nav.currFile()
	if err != nil {
		return fmt.Errorf("open: %s", err)
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

func (nav *nav) bot() {
	dir := nav.currDir()

	dir.ind = len(dir.fi) - 1
	dir.pos = min(dir.ind, nav.height-1)
}

func (nav *nav) top() {
	dir := nav.currDir()

	dir.ind = 0
	dir.pos = 0
}

func (nav *nav) cd(wd string) error {
	wd = strings.Replace(wd, "~", gUser.HomeDir, -1)
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

func match(pattern, name string) (matched bool, err error) {
	if gOpts.ignorecase {
		lpattern := strings.ToLower(pattern)
		if !gOpts.smartcase || lpattern == pattern {
			pattern = lpattern
			name = strings.ToLower(name)
		}
	}
	if gOpts.globsearch {
		return filepath.Match(pattern, name)
	}
	return strings.Contains(name, pattern), nil
}

func (nav *nav) searchNext() error {
	last := nav.currDir()
	for i := last.ind + 1; i < len(last.fi); i++ {
		matched, err := match(nav.search, last.fi[i].Name())
		if err != nil {
			return err
		}
		if matched {
			nav.down(i - last.ind)
			return nil
		}
	}
	if gOpts.wrapscan {
		for i := 0; i < last.ind; i++ {
			matched, err := match(nav.search, last.fi[i].Name())
			if err != nil {
				return err
			}
			if matched {
				nav.up(last.ind - i)
				return nil
			}
		}
	}
	return nil
}

func (nav *nav) searchPrev() error {
	last := nav.currDir()
	for i := last.ind - 1; i >= 0; i-- {
		matched, err := match(nav.search, last.fi[i].Name())
		if err != nil {
			return err
		}
		if matched {
			nav.up(last.ind - i)
			return nil
		}
	}
	if gOpts.wrapscan {
		for i := len(last.fi) - 1; i > last.ind; i-- {
			matched, err := match(nav.search, last.fi[i].Name())
			if err != nil {
				return err
			}
			if matched {
				nav.down(i - last.ind)
				return nil
			}
		}
	}
	return nil
}

func (nav *nav) toggleMark(path string) {
	if _, ok := nav.marks[path]; ok {
		delete(nav.marks, path)
		if len(nav.marks) == 0 {
			nav.markInd = 0
		}
	} else {
		nav.marks[path] = nav.markInd
		nav.markInd = nav.markInd + 1
	}
}

func (nav *nav) toggle() {
	curr, err := nav.currFile()
	if err != nil {
		return
	}

	nav.toggleMark(curr.Path)

	nav.down(1)
}

func (nav *nav) invert() {
	last := nav.currDir()
	for _, f := range last.fi {
		path := filepath.Join(last.path, f.Name())
		nav.toggleMark(path)
	}
}

func (nav *nav) save(copy bool) error {
	if len(nav.marks) == 0 {
		curr, err := nav.currFile()
		if err != nil {
			return errors.New("no file selected")
		}

		if err := saveFiles([]string{curr.Path}, copy); err != nil {
			return err
		}

		nav.saves = make(map[string]bool)
		nav.saves[curr.Path] = copy
	} else {
		var fs []string
		for f := range nav.marks {
			fs = append(fs, f)
		}

		if err := saveFiles(fs, copy); err != nil {
			return err
		}

		nav.saves = make(map[string]bool)
		for f := range nav.marks {
			nav.saves[f] = copy
		}
	}

	return nil
}

func (nav *nav) put() error {
	list, copy, err := loadFiles()
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return errors.New("no file in yank/delete buffer")
	}

	dir := nav.currDir()

	cmd := putCommand(list, dir, copy)

	if err := cmd.Run(); err != nil {
		// TODO: add name of external command to message
		return fmt.Errorf("putting files: %s", err)
	}

	// TODO: async?

	return nil
}

func (nav *nav) sync() error {
	list, copy, err := loadFiles()
	if err != nil {
		return err
	}

	nav.saves = make(map[string]bool)
	for _, f := range list {
		nav.saves[f] = copy
	}

	return nil
}

func (nav *nav) currDir() *dir {
	return nav.dirs[len(nav.dirs)-1]
}

func (nav *nav) currFile() (*file, error) {
	last := nav.dirs[len(nav.dirs)-1]

	if len(last.fi) == 0 {
		return nil, fmt.Errorf("empty directory")
	}
	return last.fi[last.ind], nil
}

type indexedMarks struct {
	paths   []string
	indices []int
}

func (m indexedMarks) Len() int { return len(m.paths) }

func (m indexedMarks) Swap(i, j int) {
	m.paths[i], m.paths[j] = m.paths[j], m.paths[i]
	m.indices[i], m.indices[j] = m.indices[j], m.indices[i]
}

func (m indexedMarks) Less(i, j int) bool { return m.indices[i] < m.indices[j] }

func (nav *nav) currMarks() []string {
	paths := make([]string, 0, len(nav.marks))
	indices := make([]int, 0, len(nav.marks))
	for path, index := range nav.marks {
		paths = append(paths, path)
		indices = append(indices, index)
	}
	sort.Sort(indexedMarks{paths: paths, indices: indices})
	return paths
}
