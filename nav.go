package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/djherbis/times"
)

type linkState byte

const (
	notLink linkState = iota
	working
	broken
)

type file struct {
	os.FileInfo
	linkState  linkState
	linkTarget string
	path       string
	dirCount   int
	dirSize    int64
	accessTime time.Time
	changeTime time.Time
	ext        string
	err        error
}

func (file *file) TotalSize() int64 {
	if file.IsDir() {
		if file.dirSize >= 0 {
			return file.dirSize
		}
		return 0
	}
	return file.Size()
}

type fakeStat struct {
	name string
}

func (fs *fakeStat) Name() string       { return fs.name }
func (fs *fakeStat) Size() int64        { return 0 }
func (fs *fakeStat) Mode() os.FileMode  { return os.FileMode(0000) }
func (fs *fakeStat) ModTime() time.Time { return time.Unix(0, 0) }
func (fs *fakeStat) IsDir() bool        { return false }
func (fs *fakeStat) Sys() any           { return nil }

func readdir(path string) ([]*file, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()

	files := make([]*file, 0, len(names))
	for _, fname := range names {
		fpath := filepath.Join(path, fname)

		lstat, err := os.Lstat(fpath)

		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			log.Printf("getting file information: %s", err)
			files = append(files, &file{
				FileInfo:   &fakeStat{name: fname},
				linkState:  notLink,
				linkTarget: "",
				path:       fpath,
				dirCount:   -1,
				dirSize:    -1,
				accessTime: time.Unix(0, 0),
				changeTime: time.Unix(0, 0),
				ext:        filepath.Ext(fpath),
				err:        err,
			})
			continue
		}

		var linkState linkState
		var linkTarget string

		if lstat.Mode()&os.ModeSymlink != 0 {
			stat, err := os.Stat(fpath)
			if err == nil {
				linkState = working
				lstat = stat
			} else {
				linkState = broken
			}
			linkTarget, err = os.Readlink(fpath)
			if err != nil {
				log.Printf("reading link target: %s", err)
			}
		}

		ts := times.Get(lstat)
		at := ts.AccessTime()
		var ct time.Time
		// from times docs: ChangeTime() panics unless HasChangeTime() is true
		if ts.HasChangeTime() {
			ct = ts.ChangeTime()
		} else {
			// fall back to ModTime if ChangeTime cannot be determined
			ct = lstat.ModTime()
		}

		// returns an empty string if extension could not be determined
		// i.e. directories, filenames without extensions
		ext := filepath.Ext(fpath)

		dirCount := -1
		if lstat.IsDir() && gOpts.dircounts {
			d, err := os.Open(fpath)
			if err != nil {
				dirCount = -2
			} else {
				names, err := d.Readdirnames(1000)
				d.Close()

				if names == nil && err != io.EOF {
					dirCount = -2
				} else {
					dirCount = len(names)
				}
			}
		}

		files = append(files, &file{
			FileInfo:   lstat,
			linkState:  linkState,
			linkTarget: linkTarget,
			path:       fpath,
			dirCount:   dirCount,
			dirSize:    -1,
			accessTime: at,
			changeTime: ct,
			ext:        ext,
			err:        nil,
		})
	}

	return files, err
}

type dir struct {
	loading     bool      // directory is loading from disk
	loadTime    time.Time // current loading or last load time
	ind         int       // index of current entry in files
	pos         int       // position of current entry in ui
	path        string    // full path of directory
	files       []*file   // displayed files in directory including or excluding hidden ones
	allFiles    []*file   // all files in directory including hidden ones (same array as files)
	sortType    sortType  // sort method and options from last sort
	dironly     bool      // dironly value from last sort
	hiddenfiles []string  // hiddenfiles value from last sort
	filter      []string  // last filter for this directory
	ignorecase  bool      // ignorecase value from last sort
	ignoredia   bool      // ignoredia value from last sort
	noPerm      bool      // whether lf has no permission to open the directory
	lines       []string  // lines of text to display if directory previews are enabled
}

func newDir(path string) *dir {
	time := time.Now()

	files, err := readdir(path)
	if err != nil {
		log.Printf("reading directory: %s", err)
	}

	return &dir{
		loading:  gOpts.dirpreviews, // Directory is loaded after previewer function exits.
		loadTime: time,
		path:     path,
		files:    files,
		allFiles: files,
		noPerm:   os.IsPermission(err),
	}
}

func normalize(s1, s2 string, ignorecase, ignoredia bool) (string, string) {
	if gOpts.ignorecase {
		s1 = strings.ToLower(s1)
		s2 = strings.ToLower(s2)
	}
	if gOpts.ignoredia {
		s1 = removeDiacritics(s1)
		s2 = removeDiacritics(s2)
	}
	return s1, s2
}

func (dir *dir) sort() {
	dir.sortType = getSortType(dir.path)
	dir.dironly = getDirOnly(dir.path)
	dir.hiddenfiles = gOpts.hiddenfiles
	dir.ignorecase = gOpts.ignorecase
	dir.ignoredia = gOpts.ignoredia

	dir.files = dir.allFiles

	// reverse order cannot be applied after stable sorting, otherwise the order
	// of equivalent elements will be reversed
	reverse := dir.sortType.option&reverseSort != 0

	switch dir.sortType.method {
	case naturalSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			s1, s2 := normalize(dir.files[i].Name(), dir.files[j].Name(), dir.ignorecase, dir.ignoredia)
			if !reverse {
				return naturalLess(s1, s2)
			} else {
				return naturalLess(s2, s1)
			}
		})
	case nameSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			s1, s2 := normalize(dir.files[i].Name(), dir.files[j].Name(), dir.ignorecase, dir.ignoredia)
			if !reverse {
				return s1 < s2
			} else {
				return s2 < s1
			}
		})
	case sizeSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			if !reverse {
				return dir.files[i].TotalSize() < dir.files[j].TotalSize()
			} else {
				return dir.files[j].TotalSize() < dir.files[i].TotalSize()
			}
		})
	case timeSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			if !reverse {
				return dir.files[i].ModTime().Before(dir.files[j].ModTime())
			} else {
				return dir.files[j].ModTime().Before(dir.files[i].ModTime())
			}
		})
	case atimeSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			if !reverse {
				return dir.files[i].accessTime.Before(dir.files[j].accessTime)
			} else {
				return dir.files[j].accessTime.Before(dir.files[i].accessTime)
			}
		})
	case ctimeSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			if !reverse {
				return dir.files[i].changeTime.Before(dir.files[j].changeTime)
			} else {
				return dir.files[j].changeTime.Before(dir.files[i].changeTime)
			}
		})
	case extSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			ext1, ext2 := normalize(dir.files[i].ext, dir.files[j].ext, dir.ignorecase, dir.ignoredia)

			// if the extension could not be determined (directories, files without)
			// use a zero byte so that these files can be ranked higher
			if ext1 == "" {
				ext1 = "\x00"
			}
			if ext2 == "" {
				ext2 = "\x00"
			}

			name1, name2 := normalize(dir.files[i].Name(), dir.files[j].Name(), dir.ignorecase, dir.ignoredia)

			// in order to also have natural sorting with the filenames
			// combine the name with the ext but have the ext at the front
			if !reverse {
				return ext1 < ext2 || ext1 == ext2 && name1 < name2
			} else {
				return ext2 < ext1 || ext2 == ext1 && name2 < name1
			}
		})
	}

	if dir.sortType.option&dirfirstSort != 0 {
		sort.SliceStable(dir.files, func(i, j int) bool {
			if dir.files[i].IsDir() == dir.files[j].IsDir() {
				return i < j
			}
			return dir.files[i].IsDir()
		})
	}

	// when dironly option is enabled, we move files to the beginning of our file
	// list and then set the beginning of displayed files to the first directory
	// in the list
	if dir.dironly {
		sort.SliceStable(dir.files, func(i, j int) bool {
			if !dir.files[i].IsDir() && !dir.files[j].IsDir() {
				return i < j
			}
			return !dir.files[i].IsDir()
		})
		dir.files = func() []*file {
			for i, f := range dir.files {
				if f.IsDir() {
					return dir.files[i:]
				}
			}
			return dir.files[len(dir.files):]
		}()
	}

	// when hidden option is disabled, we move hidden files to the
	// beginning of our file list and then set the beginning of displayed
	// files to the first non-hidden file in the list
	if dir.sortType.option&hiddenSort == 0 {
		sort.SliceStable(dir.files, func(i, j int) bool {
			if isHidden(dir.files[i], dir.path, dir.hiddenfiles) && isHidden(dir.files[j], dir.path, dir.hiddenfiles) {
				return i < j
			}
			return isHidden(dir.files[i], dir.path, dir.hiddenfiles)
		})
		for i, f := range dir.files {
			if !isHidden(f, dir.path, dir.hiddenfiles) {
				dir.files = dir.files[i:]
				break
			}
		}
		if len(dir.files) > 0 && isHidden(dir.files[len(dir.files)-1], dir.path, dir.hiddenfiles) {
			dir.files = dir.files[len(dir.files):]
		}
	}

	if len(dir.filter) != 0 {
		sort.SliceStable(dir.files, func(i, j int) bool {
			if isFiltered(dir.files[i], dir.filter) && isFiltered(dir.files[j], dir.filter) {
				return i < j
			}
			return isFiltered(dir.files[i], dir.filter)
		})
		for i, f := range dir.files {
			if !isFiltered(f, dir.filter) {
				dir.files = dir.files[i:]
				break
			}
		}
		if len(dir.files) > 0 && isFiltered(dir.files[len(dir.files)-1], dir.filter) {
			dir.files = dir.files[len(dir.files):]
		}
	}

	dir.ind = max(dir.ind, 0)
	dir.ind = min(dir.ind, len(dir.files)-1)
}

func (dir *dir) name() string {
	if len(dir.files) == 0 {
		return ""
	}

	return dir.files[dir.ind].Name()
}

func (dir *dir) sel(name string, height int) {
	if len(dir.files) == 0 {
		dir.ind, dir.pos = 0, 0
		return
	}

	dir.ind = max(dir.ind, 0)
	dir.ind = min(dir.ind, len(dir.files)-1)

	if dir.files[dir.ind].Name() != name {
		for i, f := range dir.files {
			if f.Name() == name {
				dir.ind = i
				break
			}
		}
	}

	dir.boundPos(height)
}

func (dir *dir) boundPos(height int) {
	if len(dir.files) <= height {
		dir.pos = dir.ind
		return
	}

	edge := min(height/2, gOpts.scrolloff)
	dir.pos = max(dir.pos, edge)

	// use a smaller value for half when the height is even and scrolloff is
	// maxed in order to stay at the same row while scrolling up and down
	if height%2 == 0 {
		edge = min(height/2-1, gOpts.scrolloff)
	}
	dir.pos = min(dir.pos, height-1-edge)

	dir.pos = min(dir.pos, dir.ind)
	dir.pos = max(dir.pos, height-(len(dir.files)-dir.ind))
}

type nav struct {
	init            bool
	dirs            []*dir
	copyBytes       int64
	copyTotal       int64
	copyUpdate      int
	moveCount       int
	moveTotal       int
	moveUpdate      int
	deleteCount     int
	deleteTotal     int
	deleteUpdate    int
	copyBytesChan   chan int64
	copyTotalChan   chan int64
	moveCountChan   chan int
	moveTotalChan   chan int
	deleteCountChan chan int
	deleteTotalChan chan int
	previewChan     chan string
	dirPreviewChan  chan *dir
	dirChan         chan *dir
	regChan         chan *reg
	dirCache        map[string]*dir
	regCache        map[string]*reg
	saves           map[string]bool
	marks           map[string]string
	renameOldPath   string
	renameNewPath   string
	selections      map[string]int
	tags            map[string]string
	selectionInd    int
	height          int
	find            string
	findBack        bool
	search          string
	searchBack      bool
	searchInd       int
	searchPos       int
	prevFilter      []string
	volatilePreview bool
	previewTimer    *time.Timer
	previewLoading  bool
	jumpList        []string
	jumpListInd     int
}

func (nav *nav) loadDirInternal(path string) *dir {
	nav.startPreview()

	d := &dir{
		loading:     true,
		loadTime:    time.Now(),
		path:        path,
		sortType:    getSortType(path),
		hiddenfiles: gOpts.hiddenfiles,
		ignorecase:  gOpts.ignorecase,
		ignoredia:   gOpts.ignoredia,
	}
	go func() {
		d := newDir(path)
		d.sort()
		d.ind, d.pos = 0, 0
		if gOpts.dirpreviews {
			nav.dirPreviewChan <- d
		}
		nav.dirChan <- d

	}()
	return d
}

func (nav *nav) loadDir(path string) *dir {
	if gOpts.dircache {
		d, ok := nav.dirCache[path]
		if !ok {
			d = nav.loadDirInternal(path)
			nav.dirCache[path] = d
			return d
		}

		nav.checkDir(d)

		return d
	}
	return nav.loadDirInternal(path)
}

func (nav *nav) checkDir(dir *dir) {
	s, err := os.Stat(dir.path)
	if err != nil {
		log.Printf("getting directory info: %s", err)
		return
	}

	switch {
	case s.ModTime().After(dir.loadTime):
		now := time.Now()

		// XXX: Linux builtin exFAT drivers are able to predict modifications in the future
		// https://bugs.launchpad.net/ubuntu/+source/ubuntu-meta/+bug/1872504
		if s.ModTime().After(now) {
			return
		}

		nav.startPreview()
		dir.loading = true
		dir.loadTime = now
		go func() {
			nd := newDir(dir.path)
			nd.filter = dir.filter
			nd.sort()
			if gOpts.dirpreviews {
				nav.dirPreviewChan <- nd
			}
			nav.dirChan <- nd
		}()
	case dir.sortType != getSortType(dir.path) ||
		dir.dironly != getDirOnly(dir.path) ||
		!reflect.DeepEqual(dir.hiddenfiles, gOpts.hiddenfiles) ||
		dir.ignorecase != gOpts.ignorecase ||
		dir.ignoredia != gOpts.ignoredia:
		dir.loading = true
		sd := *dir
		go func() {
			sd.sort()
			sd.loading = false
			nav.dirChan <- &sd
		}()
	}
}

func (nav *nav) getDirs(wd string) {
	var dirs []*dir

	wd = filepath.Clean(wd)

	for curr, base := wd, ""; !isRoot(base); curr, base = filepath.Dir(curr), filepath.Base(curr) {
		dir := nav.loadDir(curr)
		if base != "" {
			dir.sel(base, nav.height)
		}
		dirs = append(dirs, dir)
	}

	for i, j := 0, len(dirs)-1; i < j; i, j = i+1, j-1 {
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}

	nav.dirs = dirs
}

func newNav(height int) *nav {
	nav := &nav{
		copyBytesChan:   make(chan int64, 1024),
		copyTotalChan:   make(chan int64, 1024),
		moveCountChan:   make(chan int, 1024),
		moveTotalChan:   make(chan int, 1024),
		deleteCountChan: make(chan int, 1024),
		deleteTotalChan: make(chan int, 1024),
		previewChan:     make(chan string, 1024),
		dirPreviewChan:  make(chan *dir, 1024),
		dirChan:         make(chan *dir),
		regChan:         make(chan *reg),
		dirCache:        make(map[string]*dir),
		regCache:        make(map[string]*reg),
		saves:           make(map[string]bool),
		marks:           make(map[string]string),
		selections:      make(map[string]int),
		tags:            make(map[string]string),
		selectionInd:    0,
		height:          height,
		previewTimer:    time.NewTimer(0),
		jumpList:        make([]string, 0),
		jumpListInd:     -1,
	}

	return nav
}

func (nav *nav) addJumpList() {
	currPath := nav.currDir().path
	if nav.jumpListInd >= 0 && nav.jumpListInd < len(nav.jumpList)-1 {
		if nav.jumpList[nav.jumpListInd] == currPath {
			// walking the jumpList
			return
		}
		nav.jumpList = nav.jumpList[:nav.jumpListInd+1]
	}
	if len(nav.jumpList) == 0 || nav.jumpList[len(nav.jumpList)-1] != currPath {
		nav.jumpList = append(nav.jumpList, currPath)
	}
	nav.jumpListInd = len(nav.jumpList) - 1
}

func (nav *nav) cdJumpListPrev() {
	if nav.jumpListInd > 0 {
		nav.jumpListInd -= 1
		nav.cd(nav.jumpList[nav.jumpListInd])
	}
}

func (nav *nav) cdJumpListNext() {
	if nav.jumpListInd < len(nav.jumpList)-1 {
		nav.jumpListInd += 1
		nav.cd(nav.jumpList[nav.jumpListInd])
	}
}

func (nav *nav) renew() {
	for _, d := range nav.dirs {
		nav.checkDir(d)
	}

	for m := range nav.selections {
		if _, err := os.Lstat(m); os.IsNotExist(err) {
			delete(nav.selections, m)
		}
	}

	if len(nav.selections) == 0 {
		nav.selectionInd = 0
	}
}

func (nav *nav) reload() error {
	nav.dirCache = make(map[string]*dir)
	nav.regCache = make(map[string]*reg)

	curr, err := nav.currFile()
	nav.getDirs(nav.currDir().path)
	if err == nil {
		last := nav.dirs[len(nav.dirs)-1]
		last.files = append(last.files, curr)
	}

	return nil
}

func (nav *nav) position() {
	if !nav.init {
		return
	}

	path := nav.currDir().path
	for i := len(nav.dirs) - 2; i >= 0; i-- {
		nav.dirs[i].sel(filepath.Base(path), nav.height)
		path = filepath.Dir(path)
	}
}

func (nav *nav) exportFiles() {
	if !nav.init {
		return
	}

	var currFile string
	if curr, err := nav.currFile(); err == nil {
		currFile = quoteString(curr.path)
	}

	var selections []string
	for _, selection := range nav.currSelections() {
		selections = append(selections, quoteString(selection))
	}
	currSelections := strings.Join(selections, gOpts.filesep)

	os.Setenv("f", currFile)
	os.Setenv("fs", currSelections)
	os.Setenv("PWD", quoteString(nav.currDir().path))

	if len(selections) == 0 {
		os.Setenv("fx", currFile)
	} else {
		os.Setenv("fx", currSelections)
	}
}

func (nav *nav) dirPreviewLoop(ui *ui) {
	var prevPath string
	for dir := range nav.dirPreviewChan {
		if dir == nil && len(gOpts.previewer) != 0 && len(gOpts.cleaner) != 0 && nav.volatilePreview {
			cmd := exec.Command(gOpts.cleaner, prevPath)
			if err := cmd.Run(); err != nil {
				log.Printf("cleaning preview: %s", err)
			}
			nav.volatilePreview = false
		} else if dir != nil {
			win := ui.wins[len(ui.wins)-1]
			nav.previewDir(dir, win)
			prevPath = dir.path
		}
	}
}

func (nav *nav) previewLoop(ui *ui) {
	var prev string
	for path := range nav.previewChan {
		clear := len(path) == 0
	loop:
		for {
			select {
			case path = <-nav.previewChan:
				clear = clear || len(path) == 0
			default:
				break loop
			}
		}
		win := ui.wins[len(ui.wins)-1]
		if clear && len(gOpts.previewer) != 0 && len(gOpts.cleaner) != 0 && nav.volatilePreview {
			cmd := exec.Command(gOpts.cleaner, prev,
				strconv.Itoa(win.w),
				strconv.Itoa(win.h),
				strconv.Itoa(win.x),
				strconv.Itoa(win.y),
				path)
			if err := cmd.Run(); err != nil {
				log.Printf("cleaning preview: %s", err)
			}
			nav.volatilePreview = false
		}
		if len(path) != 0 {
			nav.preview(path, win)
			prev = path
		}
	}
}

//lint:ignore U1000 This function is not used on Windows
func matchPattern(pattern, name, path string) bool {
	s := name

	pattern = replaceTilde(pattern)

	if filepath.IsAbs(pattern) {
		s = filepath.Join(path, name)
	}

	// pattern errors are checked when 'hiddenfiles' option is set
	matched, _ := filepath.Match(pattern, s)

	return matched
}

func (nav *nav) previewDir(dir *dir, win *win) {

	defer func() {
		dir.loading = false
		nav.dirChan <- dir
	}()

	var reader io.Reader

	if len(gOpts.previewer) != 0 {
		cmd := exec.Command(gOpts.previewer, dir.path,
			strconv.Itoa(win.w),
			strconv.Itoa(win.h),
			strconv.Itoa(win.x),
			strconv.Itoa(win.y))

		out, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("previewing dir: %s", err)
			return
		}

		if err := cmd.Start(); err != nil {
			log.Printf("previewing dir: %s", err)
			out.Close()
			return
		}

		defer func() {
			if err := cmd.Wait(); err != nil {
				if e, ok := err.(*exec.ExitError); ok {
					if e.ExitCode() != 0 {
						nav.volatilePreview = true
					}
				} else {
					log.Printf("loading dir: %s", err)
				}
			}
		}()
		defer out.Close()
		reader = out
		buf := bufio.NewScanner(reader)

		for i := 0; i < win.h && buf.Scan(); i++ {
			for _, r := range buf.Text() {
				if r == 0 {
					dir.lines = []string{"\033[7mbinary\033[0m"}
					return
				}
			}
			dir.lines = append(dir.lines, buf.Text())
		}

		if buf.Err() != nil {
			log.Printf("loading dir: %s", buf.Err())
		}
	}

}

func (nav *nav) preview(path string, win *win) {

	reg := &reg{loadTime: time.Now(), path: path}
	defer func() { nav.regChan <- reg }()

	var reader *bufio.Reader

	if len(gOpts.previewer) != 0 {
		cmd := exec.Command(gOpts.previewer, path,
			strconv.Itoa(win.w),
			strconv.Itoa(win.h),
			strconv.Itoa(win.x),
			strconv.Itoa(win.y))

		out, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("previewing file: %s", err)
			return
		}

		if err := cmd.Start(); err != nil {
			log.Printf("previewing file: %s", err)
			out.Close()
			return
		}

		defer func() {
			if err := cmd.Wait(); err != nil {
				if e, ok := err.(*exec.ExitError); ok {
					if e.ExitCode() != 0 {
						reg.volatile = true
						nav.volatilePreview = true
					}
				} else {
					log.Printf("loading file: %s", err)
				}
			}
		}()
		defer out.Close()
		reader = bufio.NewReader(out)
	} else {
		f, err := os.Open(path)
		if err != nil {
			log.Printf("opening file: %s", err)
			return
		}

		defer f.Close()
		reader = bufio.NewReader(f)
	}

	if gOpts.sixel {
		prefix, err := reader.Peek(2)
		if err == nil && string(prefix) == gSixelBegin {
			b, err := io.ReadAll(reader)
			if err != nil {
				log.Printf("loading sixel: %s", err)
			}
			str := string(b)
			reg.sixel = &str
			return
		}
	}

	// bufio.Scanner can't handle files containing long lines if they exceed the
	// size of its internal buffer
	addLine := true
	for len(reg.lines) < win.h {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			break
		}

		for _, r := range line {
			if r == 0 {
				reg.lines = []string{"\033[7mbinary\033[0m"}
				return
			}
		}

		if addLine {
			reg.lines = append(reg.lines, string(line))
		}

		addLine = !isPrefix
	}
}

func (nav *nav) loadReg(path string, volatile bool) *reg {
	r, ok := nav.regCache[path]
	if !ok || (volatile && r.volatile) {
		r := &reg{loading: true, loadTime: time.Now(), path: path, volatile: true}
		nav.regCache[path] = r
		nav.startPreview()
		nav.previewChan <- path
		return r
	}

	nav.checkReg(r)

	return r
}

func (nav *nav) checkReg(reg *reg) {
	s, err := os.Stat(reg.path)
	if err != nil {
		return
	}

	now := time.Now()

	// XXX: Linux builtin exFAT drivers are able to predict modifications in the future
	// https://bugs.launchpad.net/ubuntu/+source/ubuntu-meta/+bug/1872504
	if s.ModTime().After(now) {
		return
	}

	if s.ModTime().After(reg.loadTime) {
		reg.loadTime = now
		nav.startPreview()
		nav.previewChan <- reg.path
	}
}

func (nav *nav) startPreview() {
	nav.previewTimer.Stop()
	nav.previewLoading = false
	nav.previewTimer.Reset(100 * time.Millisecond)
}

func (nav *nav) sort() {
	for _, d := range nav.dirs {
		name := d.name()
		d.sort()
		d.sel(name, nav.height)
	}
}

func (nav *nav) setFilter(filter []string) error {
	newfilter := []string{}
	for _, tok := range filter {
		_, err := filepath.Match(tok, "a")
		if err != nil {
			return err
		}
		if tok != "" {
			newfilter = append(newfilter, tok)
		}
	}
	dir := nav.currDir()
	dir.filter = newfilter

	// Apply filter, by sorting current dir (see nav.sort())
	name := dir.name()
	dir.sort()
	dir.sel(name, nav.height)
	return nil
}

func (nav *nav) up(dist int) bool {
	dir := nav.currDir()

	old := dir.ind

	if dir.ind == 0 {
		if gOpts.wrapscroll {
			nav.bottom()
		}
		return old != dir.ind
	}

	dir.ind -= dist
	dir.ind = max(0, dir.ind)

	dir.pos -= dist
	dir.boundPos(nav.height)

	return old != dir.ind
}

func (nav *nav) down(dist int) bool {
	dir := nav.currDir()

	old := dir.ind

	maxind := len(dir.files) - 1

	if dir.ind >= maxind {
		if gOpts.wrapscroll {
			nav.top()
		}
		return old != dir.ind
	}

	dir.ind += dist
	dir.ind = min(maxind, dir.ind)

	dir.pos += dist
	dir.boundPos(nav.height)

	return old != dir.ind
}

func (nav *nav) scrollUp(dist int) bool {
	dir := nav.currDir()

	// when reached top do nothing
	if istop := dir.ind == dir.pos; istop {
		return false
	}

	old := dir.ind

	minedge := min(nav.height/2, gOpts.scrolloff)

	dir.pos += dist

	// jump to ensure minedge when edge < minedge
	edge := nav.height - dir.pos
	delta := min(0, edge-minedge-1)
	dir.pos = min(dir.pos, nav.height-minedge-1)
	// update dir.ind accordingly
	dir.ind = dir.ind + delta

	dir.ind = min(dir.ind, dir.ind-(dir.pos-nav.height+1))

	// prevent cursor disappearing downwards
	dir.pos = min(dir.pos, nav.height-1)

	return old != dir.ind
}

func (nav *nav) scrollDown(dist int) bool {
	dir := nav.currDir()
	maxind := len(dir.files) - 1

	// reached bottom
	if dir.ind-dir.pos > maxind-nav.height {
		return false
	}

	old := dir.ind

	minedge := min(nav.height/2, gOpts.scrolloff)

	dir.pos -= dist

	// jump to ensure minedge when edge < minedge
	delta := min(0, dir.pos-minedge)
	dir.pos = max(dir.pos, minedge)
	// update dir.ind accordingly
	dir.ind = dir.ind - delta
	dir.ind = max(dir.ind, dir.ind-(dir.pos-minedge))

	dir.ind = min(maxind, dir.ind)
	// prevent disappearing
	dir.pos = max(dir.pos, 0)

	return old != dir.ind
}

func (nav *nav) updir() error {
	if len(nav.dirs) <= 1 {
		return nil
	}

	dir := nav.currDir()

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

	path := curr.path

	dir := nav.loadDir(path)

	nav.dirs = append(nav.dirs, dir)

	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("open: %s", err)
	}

	return nil
}

func (nav *nav) top() bool {
	dir := nav.currDir()

	old := dir.ind

	dir.ind = 0
	dir.pos = 0

	return old != dir.ind
}

func (nav *nav) bottom() bool {
	dir := nav.currDir()

	old := dir.ind

	dir.ind = len(dir.files) - 1
	dir.pos = min(dir.ind, nav.height-1)

	return old != dir.ind
}

func (nav *nav) high() bool {
	dir := nav.currDir()

	old := dir.ind
	beg := max(dir.ind-dir.pos, 0)
	offs := gOpts.scrolloff
	if beg == 0 {
		offs = 0
	}

	dir.ind = beg + offs
	dir.pos = offs

	return old != dir.ind
}

func (nav *nav) middle() bool {
	dir := nav.currDir()

	old := dir.ind
	beg := max(dir.ind-dir.pos, 0)
	end := min(beg+nav.height, len(dir.files))

	half := (end - beg) / 2
	dir.ind = beg + half
	dir.pos = half

	return old != dir.ind
}

func (nav *nav) low() bool {
	dir := nav.currDir()

	old := dir.ind
	beg := max(dir.ind-dir.pos, 0)
	end := min(beg+nav.height, len(dir.files))
	offs := gOpts.scrolloff
	if end == len(dir.files) {
		offs = 0
	}

	dir.ind = end - 1 - offs
	dir.pos = end - beg - 1 - offs

	return old != dir.ind
}

func (nav *nav) move(index int) bool {
	old := nav.currDir().ind

	switch {
	case index < old:
		return nav.up(old - index)
	case index > old:
		return nav.down(index - old)
	default:
		return false
	}
}

func (nav *nav) toggleSelection(path string) {
	if _, ok := nav.selections[path]; ok {
		delete(nav.selections, path)
		if len(nav.selections) == 0 {
			nav.selectionInd = 0
		}
	} else {
		nav.selections[path] = nav.selectionInd
		nav.selectionInd++
	}
}

func (nav *nav) toggle() {
	curr, err := nav.currFile()
	if err != nil {
		return
	}

	nav.toggleSelection(curr.path)
}

func (nav *nav) tagToggleSelection(path string, tag string) {
	if _, ok := nav.tags[path]; ok {
		delete(nav.tags, path)
	} else {
		nav.tags[path] = tag
	}
}

func (nav *nav) tagToggle(tag string) error {
	list, err := nav.currFileOrSelections()
	if err != nil {
		return err
	}

	if printLength(tag) != 1 {
		return errors.New("tag should be single width character")
	}

	for _, path := range list {
		nav.tagToggleSelection(path, tag)
	}

	return nil
}

func (nav *nav) tag(tag string) error {
	list, err := nav.currFileOrSelections()
	if err != nil {
		return err
	}

	if printLength(tag) != 1 {
		return errors.New("tag should be single width character")
	}

	for _, path := range list {
		nav.tags[path] = tag
	}

	return nil
}

func (nav *nav) invertAfter(ix int) {
	dir := nav.currDir()
	for _, f := range dir.files[ix:] {
		path := filepath.Join(dir.path, f.Name())
		nav.toggleSelection(path)
	}
}

func (nav *nav) invert() {
	nav.invertAfter(0)
}

func (nav *nav) invertBelow() {
	nav.invertAfter(nav.currDir().ind)
}

func (nav *nav) unselect() {
	nav.selections = make(map[string]int)
	nav.selectionInd = 0
}

func (nav *nav) save(cp bool) error {
	list, err := nav.currFileOrSelections()
	if err != nil {
		return err
	}

	if err := saveFiles(list, cp); err != nil {
		return err
	}

	nav.saves = make(map[string]bool)
	for _, f := range list {
		nav.saves[f] = cp
	}

	return nil
}

func (nav *nav) copyAsync(app *app, srcs []string, dstDir string) {
	echo := &callExpr{"echoerr", []string{""}, 1}

	_, err := os.Stat(dstDir)
	if os.IsNotExist(err) {
		echo.args[0] = err.Error()
		app.ui.exprChan <- echo
		return
	}

	total, err := copySize(srcs)
	if err != nil {
		echo.args[0] = err.Error()
		app.ui.exprChan <- echo
		return
	}

	nav.copyTotalChan <- total

	nums, errs := copyAll(srcs, dstDir)

	errCount := 0
loop:
	for {
		select {
		case n := <-nums:
			nav.copyBytesChan <- n
		case err, ok := <-errs:
			if !ok {
				break loop
			}
			errCount++
			echo.args[0] = fmt.Sprintf("[%d] %s", errCount, err)
			app.ui.exprChan <- echo
		}
	}

	nav.copyTotalChan <- -total

	if gSingleMode {
		nav.renew()
		app.ui.loadFile(app, true)
	} else {
		if err := remote("send load"); err != nil {
			errCount++
			echo.args[0] = fmt.Sprintf("[%d] %s", errCount, err)
			app.ui.exprChan <- echo
		}
	}

	if errCount == 0 {
		app.ui.exprChan <- &callExpr{"echo", []string{"\033[0;32mCopied successfully\033[0m"}, 1}
	}
}

func (nav *nav) moveAsync(app *app, srcs []string, dstDir string) {
	echo := &callExpr{"echoerr", []string{""}, 1}

	_, err := os.Stat(dstDir)
	if os.IsNotExist(err) {
		echo.args[0] = err.Error()
		app.ui.exprChan <- echo
		return
	}

	nav.moveTotalChan <- len(srcs)

	errCount := 0
	for _, src := range srcs {
		nav.moveCountChan <- 1

		srcStat, err := os.Lstat(src)
		if err != nil {
			errCount++
			echo.args[0] = fmt.Sprintf("[%d] %s", errCount, err)
			app.ui.exprChan <- echo
			continue
		}

		file := filepath.Base(src)
		dst := filepath.Join(dstDir, file)

		dstStat, err := os.Stat(dst)
		if os.SameFile(srcStat, dstStat) {
			errCount++
			echo.args[0] = fmt.Sprintf("[%d] rename %s %s: source and destination are the same file", errCount, src, dst)
			app.ui.exprChan <- echo
			continue
		} else if !os.IsNotExist(err) {
			ext := filepath.Ext(file)
			basename := filepath.Base(file[:len(file)-len(ext)])
			var newPath string
			for i := 1; !os.IsNotExist(err); i++ {
				file = strings.ReplaceAll(gOpts.dupfilefmt, "%f", basename+ext)
				file = strings.ReplaceAll(file, "%b", basename)
				file = strings.ReplaceAll(file, "%e", ext)
				file = strings.ReplaceAll(file, "%n", strconv.Itoa(i))
				newPath = filepath.Join(dstDir, file)
				_, err = os.Lstat(newPath)
			}
			dst = newPath
		}

		if err := os.Rename(src, dst); err != nil {
			if errCrossDevice(err) {
				total, err := copySize([]string{src})
				if err != nil {
					echo.args[0] = err.Error()
					app.ui.exprChan <- echo
					continue
				}

				nav.copyTotalChan <- total

				nums, errs := copyAll([]string{src}, dstDir)

				oldCount := errCount
			loop:
				for {
					select {
					case n := <-nums:
						nav.copyBytesChan <- n
					case err, ok := <-errs:
						if !ok {
							break loop
						}
						errCount++
						echo.args[0] = fmt.Sprintf("[%d] %s", errCount, err)
						app.ui.exprChan <- echo
					}
				}

				nav.copyTotalChan <- -total

				if errCount == oldCount {
					if err := os.RemoveAll(src); err != nil {
						errCount++
						echo.args[0] = fmt.Sprintf("[%d] %s", errCount, err)
						app.ui.exprChan <- echo
					}
				}
			} else {
				errCount++
				echo.args[0] = fmt.Sprintf("[%d] %s", errCount, err)
				app.ui.exprChan <- echo
			}
		}
	}

	nav.moveTotalChan <- -len(srcs)

	if gSingleMode {
		nav.renew()
		app.ui.loadFile(app, true)
	} else {
		if err := remote("send load"); err != nil {
			errCount++
			echo.args[0] = fmt.Sprintf("[%d] %s", errCount, err)
			app.ui.exprChan <- echo
		}
	}

	if errCount == 0 {
		app.ui.exprChan <- &callExpr{"echo", []string{"\033[0;32mMoved successfully\033[0m"}, 1}
	}
}

func (nav *nav) paste(app *app) error {
	srcs, cp, err := loadFiles()
	if err != nil {
		return err
	}

	if len(srcs) == 0 {
		return errors.New("no file in copy/cut buffer")
	}

	dstDir := nav.currDir().path

	if cp {
		go nav.copyAsync(app, srcs, dstDir)
	} else {
		go nav.moveAsync(app, srcs, dstDir)
		if err := saveFiles(nil, false); err != nil {
			return fmt.Errorf("clearing copy/cut buffer: %s", err)
		}

		if gSingleMode {
			if err := nav.sync(); err != nil {
				return fmt.Errorf("paste: %s", err)
			}
		} else {
			if err := remote("send sync"); err != nil {
				return fmt.Errorf("paste: %s", err)
			}
		}
	}

	return nil
}

func (nav *nav) del(app *app) error {
	list, err := nav.currFileOrSelections()
	if err != nil {
		return err
	}

	go func() {
		echo := &callExpr{"echoerr", []string{""}, 1}
		errCount := 0

		nav.deleteTotalChan <- len(list)

		for _, path := range list {
			nav.deleteCountChan <- 1

			if err := os.RemoveAll(path); err != nil {
				errCount++
				echo.args[0] = fmt.Sprintf("[%d] %s", errCount, err)
				app.ui.exprChan <- echo
			}
		}

		nav.deleteTotalChan <- -len(list)

		if gSingleMode {
			nav.renew()
			app.ui.loadFile(app, true)
		} else {
			if err := remote("send load"); err != nil {
				errCount++
				echo.args[0] = fmt.Sprintf("[%d] %s", errCount, err)
				app.ui.exprChan <- echo
			}
		}
	}()

	return nil
}

func (nav *nav) rename() error {
	oldPath := nav.renameOldPath
	newPath := nav.renameNewPath

	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	lstat, err := os.Lstat(newPath)
	if err != nil {
		return err
	}

	// It is possible for newPath to already have cache entries if it previously
	// existed and was deleted. In this case the cache entries should be deleted
	// before loading newPath to prevent displaying a stale preview. However,
	// this clears only the current instance of lf, and not any other instances.
	delete(nav.regCache, newPath)
	delete(nav.dirCache, newPath)
	dir := nav.loadDir(filepath.Dir(newPath))

	if dir.loading {
		dir.files = append(dir.files, &file{FileInfo: lstat})
	}

	dir.sel(lstat.Name(), nav.height)

	return nil
}

func (nav *nav) sync() error {
	list, cp, err := loadFiles()
	if err != nil {
		return err
	}

	nav.saves = make(map[string]bool)
	for _, f := range list {
		nav.saves[f] = cp
	}

	oldmarks := nav.marks
	errMarks := nav.readMarks()
	for _, ch := range gOpts.tempmarks {
		tmp := string(ch)
		if v, e := oldmarks[tmp]; e {
			nav.marks[tmp] = v
		}
	}
	err = nav.readTags()
	if errMarks != nil {
		return errMarks
	}
	return err
}

func (nav *nav) cd(wd string) error {
	wd = replaceTilde(wd)
	wd = filepath.Clean(wd)

	if !filepath.IsAbs(wd) {
		wd = filepath.Join(nav.currDir().path, wd)
	}

	if err := os.Chdir(wd); err != nil {
		return fmt.Errorf("cd: %s", err)
	}

	nav.getDirs(wd)
	nav.addJumpList()
	return nil
}

func (nav *nav) sel(path string) error {
	path = replaceTilde(path)
	path = filepath.Clean(path)

	lstat, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("select: %s", err)
	}

	dir := filepath.Dir(path)

	if err := nav.cd(dir); err != nil {
		return fmt.Errorf("select: %s", err)
	}

	base := filepath.Base(path)

	last := nav.currDir()

	if last.loading {
		last.files = append(last.files, &file{FileInfo: lstat})
	}

	for i, f := range last.files {
		if f.Name() == base {
			nav.move(i)
			break
		}
	}

	return nil
}

func (nav *nav) globSel(pattern string, invert bool) error {
	dir := nav.currDir()
	anyMatched := false

	for i := 0; i < len(dir.files); i++ {
		matched, err := filepath.Match(pattern, dir.files[i].Name())
		if err != nil {
			return fmt.Errorf("glob-select: %s", err)
		}
		if matched {
			anyMatched = true
			fpath := filepath.Join(dir.path, dir.files[i].Name())
			if _, ok := nav.selections[fpath]; ok == invert {
				nav.toggleSelection(fpath)
			}
		}
	}

	if !anyMatched {
		return fmt.Errorf("glob-select: pattern not found: %s", pattern)
	}

	return nil
}

func findMatch(name, pattern string) bool {
	if gOpts.ignorecase {
		lpattern := strings.ToLower(pattern)
		if !gOpts.smartcase || lpattern == pattern {
			pattern = lpattern
			name = strings.ToLower(name)
		}
	}
	if gOpts.ignoredia {
		lpattern := removeDiacritics(pattern)
		if !gOpts.smartdia || lpattern == pattern {
			pattern = lpattern
			name = removeDiacritics(name)
		}
	}
	if gOpts.anchorfind {
		return strings.HasPrefix(name, pattern)
	}
	return strings.Contains(name, pattern)
}

func (nav *nav) findSingle() int {
	count := 0
	index := 0
	dir := nav.currDir()
	for i := 0; i < len(dir.files); i++ {
		if findMatch(dir.files[i].Name(), nav.find) {
			count++
			if count > 1 {
				return count
			}
			index = i
		}
	}
	if count == 1 {
		if index > dir.ind {
			nav.down(index - dir.ind)
		} else {
			nav.up(dir.ind - index)
		}
	}
	return count
}

func (nav *nav) findNext() (bool, bool) {
	dir := nav.currDir()
	for i := dir.ind + 1; i < len(dir.files); i++ {
		if findMatch(dir.files[i].Name(), nav.find) {
			return nav.down(i - dir.ind), true
		}
	}
	if gOpts.wrapscan {
		for i := 0; i < dir.ind; i++ {
			if findMatch(dir.files[i].Name(), nav.find) {
				return nav.up(dir.ind - i), true
			}
		}
	}
	return false, false
}

func (nav *nav) findPrev() (bool, bool) {
	dir := nav.currDir()
	for i := dir.ind - 1; i >= 0; i-- {
		if findMatch(dir.files[i].Name(), nav.find) {
			return nav.up(dir.ind - i), true
		}
	}
	if gOpts.wrapscan {
		for i := len(dir.files) - 1; i > dir.ind; i-- {
			if findMatch(dir.files[i].Name(), nav.find) {
				return nav.down(i - dir.ind), true
			}
		}
	}
	return false, false
}

func searchMatch(name, pattern string) (matched bool, err error) {
	if gOpts.ignorecase {
		lpattern := strings.ToLower(pattern)
		if !gOpts.smartcase || lpattern == pattern {
			pattern = lpattern
			name = strings.ToLower(name)
		}
	}
	if gOpts.ignoredia {
		lpattern := removeDiacritics(pattern)
		if !gOpts.smartdia || lpattern == pattern {
			pattern = lpattern
			name = removeDiacritics(name)
		}
	}
	if gOpts.globsearch {
		return filepath.Match(pattern, name)
	}
	return strings.Contains(name, pattern), nil
}

func (nav *nav) searchNext() (bool, error) {
	dir := nav.currDir()
	for i := dir.ind + 1; i < len(dir.files); i++ {
		if matched, err := searchMatch(dir.files[i].Name(), nav.search); err != nil {
			return false, err
		} else if matched {
			return nav.down(i - dir.ind), nil
		}
	}
	if gOpts.wrapscan {
		for i := 0; i < dir.ind; i++ {
			if matched, err := searchMatch(dir.files[i].Name(), nav.search); err != nil {
				return false, err
			} else if matched {
				return nav.up(dir.ind - i), nil
			}
		}
	}
	return false, nil
}

func (nav *nav) searchPrev() (bool, error) {
	dir := nav.currDir()
	for i := dir.ind - 1; i >= 0; i-- {
		if matched, err := searchMatch(dir.files[i].Name(), nav.search); err != nil {
			return false, err
		} else if matched {
			return nav.up(dir.ind - i), nil
		}
	}
	if gOpts.wrapscan {
		for i := len(dir.files) - 1; i > dir.ind; i-- {
			if matched, err := searchMatch(dir.files[i].Name(), nav.search); err != nil {
				return false, err
			} else if matched {
				return nav.down(i - dir.ind), nil
			}
		}
	}
	return false, nil
}

func isFiltered(f os.FileInfo, filter []string) bool {
	for _, pattern := range filter {
		matched, err := searchMatch(f.Name(), strings.TrimPrefix(pattern, "!"))
		if err != nil {
			log.Printf("Filter Error: %s", err)
			return false
		}
		if strings.HasPrefix(pattern, "!") && matched {
			return true
		} else if !strings.HasPrefix(pattern, "!") && !matched {
			return true
		}
	}
	return false
}

func (nav *nav) removeMark(mark string) error {
	if _, ok := nav.marks[mark]; ok {
		delete(nav.marks, mark)
		return nil
	}
	return fmt.Errorf("no such mark")
}

func (nav *nav) readMarks() error {
	nav.marks = make(map[string]string)
	f, err := os.Open(gMarksPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("opening marks file: %s", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		toks := strings.SplitN(scanner.Text(), ":", 2)
		if _, ok := nav.marks[toks[0]]; !ok {
			nav.marks[toks[0]] = toks[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading marks file: %s", err)
	}

	return nil
}

func (nav *nav) writeMarks() error {
	if err := os.MkdirAll(filepath.Dir(gMarksPath), os.ModePerm); err != nil {
		return fmt.Errorf("creating data directory: %s", err)
	}

	f, err := os.Create(gMarksPath)
	if err != nil {
		return fmt.Errorf("creating marks file: %s", err)
	}
	defer f.Close()

	var keys []string
	for k := range nav.marks {
		if !strings.Contains(gOpts.tempmarks, k) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		_, err = f.WriteString(fmt.Sprintf("%s:%s\n", k, nav.marks[k]))
		if err != nil {
			return fmt.Errorf("writing marks file: %s", err)
		}
	}

	return nil
}

func (nav *nav) readTags() error {
	nav.tags = make(map[string]string)
	f, err := os.Open(gTagsPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("opening tags file: %s", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		ind := strings.LastIndex(text, ":")
		path := text[0:ind]
		mark := text[ind+1:]
		if _, ok := nav.tags[path]; !ok {
			nav.tags[path] = mark
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading tags file: %s", err)
	}

	return nil
}

func (nav *nav) writeTags() error {
	if err := os.MkdirAll(filepath.Dir(gTagsPath), os.ModePerm); err != nil {
		return fmt.Errorf("creating data directory: %s", err)
	}

	f, err := os.Create(gTagsPath)
	if err != nil {
		return fmt.Errorf("creating tags file: %s", err)
	}
	defer f.Close()

	var keys []string
	for k := range nav.tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		_, err = f.WriteString(fmt.Sprintf("%s:%s\n", k, nav.tags[k]))
		if err != nil {
			return fmt.Errorf("writing tags file: %s", err)
		}
	}

	return nil
}

func (nav *nav) currDir() *dir {
	return nav.dirs[len(nav.dirs)-1]
}

func (nav *nav) currFile() (*file, error) {
	dir := nav.dirs[len(nav.dirs)-1]

	if len(dir.files) == 0 {
		return nil, fmt.Errorf("empty directory")
	}

	return dir.files[dir.ind], nil
}

type indexedSelections struct {
	paths   []string
	indices []int
}

func (m indexedSelections) Len() int { return len(m.paths) }

func (m indexedSelections) Swap(i, j int) {
	m.paths[i], m.paths[j] = m.paths[j], m.paths[i]
	m.indices[i], m.indices[j] = m.indices[j], m.indices[i]
}

func (m indexedSelections) Less(i, j int) bool { return m.indices[i] < m.indices[j] }

func (nav *nav) currSelections() []string {

	currDirOnly := gOpts.selmode == "dir"
	currDirPath := ""
	if currDirOnly {
		// select only from this directory
		currDirPath = nav.currDir().path
	}

	paths := make([]string, 0, len(nav.selections))
	indices := make([]int, 0, len(nav.selections))
	for path, index := range nav.selections {
		if !currDirOnly || filepath.Dir(path) == currDirPath {
			paths = append(paths, path)
			indices = append(indices, index)
		}
	}
	sort.Sort(indexedSelections{paths: paths, indices: indices})
	return paths
}

func (nav *nav) currFileOrSelections() (list []string, err error) {
	sel := nav.currSelections()

	if len(sel) == 0 {
		curr, err := nav.currFile()
		if err != nil {
			return nil, errors.New("no file selected")
		}
		return []string{curr.path}, nil
	}
	return sel, nil
}

func (nav *nav) calcDirSize() error {
	calc := func(f *file) error {
		if f.IsDir() {
			total, err := copySize([]string{f.path})
			if err != nil {
				return err
			}
			f.dirSize = total
		}
		return nil
	}

	if len(nav.selections) == 0 {
		curr, err := nav.currFile()
		if err != nil {
			return errors.New("no file selected")
		}
		return calc(curr)
	}

	for sel := range nav.selections {
		lstat, err := os.Lstat(sel)
		if err != nil || !lstat.IsDir() {
			continue
		}
		path, name := filepath.Dir(sel), filepath.Base(sel)
		dir := nav.loadDir(path)

		for _, f := range dir.files {
			if f.Name() == name {
				err := calc(f)
				if err != nil {
					return err
				}
				break
			}
		}
	}

	return nil
}
