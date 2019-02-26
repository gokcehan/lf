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
	"sort"
	"strconv"
	"strings"
	"time"
)

type linkState byte

const (
	notLink linkState = iota
	working
	broken
)

type file struct {
	os.FileInfo
	linkState linkState
	path      string
	dirCount  int
}

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
			return files, err
		}

		var linkState linkState

		if lstat.Mode()&os.ModeSymlink != 0 {
			stat, err := os.Stat(fpath)
			if err == nil {
				linkState = working
				lstat = stat
			} else {
				linkState = broken
			}
		}

		files = append(files, &file{
			FileInfo:  lstat,
			linkState: linkState,
			path:      fpath,
			dirCount:  -1,
		})
	}

	return files, err
}

type dir struct {
	loading  bool      // directory is loading from disk
	loadTime time.Time // current loading or last load time
	ind      int       // index of current entry in files
	pos      int       // position of current entry in ui
	path     string    // full path of directory
	files    []*file   // displayed files in directory including or excluding hidden ones
	allFiles []*file   // all files in directory including hidden ones (same array as files)
	sortType sortType  // sort method and options from last sort
}

func newDir(path string) *dir {
	time := time.Now()

	files, err := readdir(path)
	if err != nil {
		log.Printf("reading directory: %s", err)
	}

	return &dir{
		loadTime: time,
		path:     path,
		files:    files,
		allFiles: files,
	}
}

func (dir *dir) sort() {
	dir.sortType = gOpts.sortType

	dir.files = dir.allFiles

	switch gOpts.sortType.method {
	case naturalSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			return naturalLess(strings.ToLower(dir.files[i].Name()), strings.ToLower(dir.files[j].Name()))
		})
	case nameSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			return strings.ToLower(dir.files[i].Name()) < strings.ToLower(dir.files[j].Name())
		})
	case sizeSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			return dir.files[i].Size() < dir.files[j].Size()
		})
	case timeSort:
		sort.SliceStable(dir.files, func(i, j int) bool {
			return dir.files[i].ModTime().Before(dir.files[j].ModTime())
		})
	}

	if gOpts.sortType.option&reverseSort != 0 {
		for i, j := 0, len(dir.files)-1; i < j; i, j = i+1, j-1 {
			dir.files[i], dir.files[j] = dir.files[j], dir.files[i]
		}
	}

	if gOpts.sortType.option&dirfirstSort != 0 {
		sort.SliceStable(dir.files, func(i, j int) bool {
			if dir.files[i].IsDir() == dir.files[j].IsDir() {
				return i < j
			}
			return dir.files[i].IsDir()
		})
	}

	// when hidden option is disabled, we move hidden files to the
	// beginning of our file list and then set the beginning of displayed
	// files to the first non-hidden file in the list
	if gOpts.sortType.option&hiddenSort == 0 {
		sort.SliceStable(dir.files, func(i, j int) bool {
			if dir.files[i].Name()[0] == '.' && dir.files[j].Name()[0] == '.' {
				return i < j
			}
			return dir.files[i].Name()[0] == '.'
		})
		for i, f := range dir.files {
			if f.Name()[0] != '.' {
				dir.files = dir.files[i:]
				return
			}
		}
		dir.files = dir.files[len(dir.files):]
	}
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

	dir.ind = min(dir.ind, len(dir.files)-1)

	if dir.files[dir.ind].Name() != name {
		for i, f := range dir.files {
			if f.Name() == name {
				dir.ind = i
				break
			}
		}
	}

	edge := min(min(height/2, gOpts.scrolloff), len(dir.files)-dir.ind-1)
	dir.pos = min(dir.ind, height-edge-1)
}

type nav struct {
	dirs         []*dir
	dirChan      chan *dir
	regChan      chan *reg
	dirCache     map[string]*dir
	regCache     map[string]*reg
	saves        map[string]bool
	marks        map[string]string
	selections   map[string]int
	selectionInd int
	height       int
	find         string
	findBack     bool
	search       string
	searchBack   bool
	searchInd    int
	searchPos    int
}

func (nav *nav) loadDir(path string) *dir {
	d, ok := nav.dirCache[path]
	if !ok {
		go func() {
			d := newDir(path)
			d.sort()
			d.ind, d.pos = 0, 0
			nav.dirChan <- d
		}()
		d := &dir{loading: true, path: path, sortType: gOpts.sortType}
		nav.dirCache[path] = d
		return d
	}

	s, err := os.Stat(d.path)
	if err != nil {
		return d
	}

	switch {
	case s.ModTime().After(d.loadTime):
		go func() {
			d.loadTime = time.Now()
			nd := newDir(path)
			nd.sort()
			nd.sel(d.name(), nav.height)
			nav.dirChan <- nd
		}()
	case d.sortType != gOpts.sortType:
		go func() {
			d.loading = true
			name := d.name()
			d.sort()
			d.sel(name, nav.height)
			d.loading = false
			nav.dirChan <- d
		}()
	}

	return d
}

func (nav *nav) getDirs(wd string) {
	var dirs []*dir

	for curr, base := wd, ""; !isRoot(base); curr, base = filepath.Dir(curr), filepath.Base(curr) {
		dir := nav.loadDir(curr)
		dir.sel(base, nav.height)
		dirs = append(dirs, dir)
	}

	for i, j := 0, len(dirs)-1; i < j; i, j = i+1, j-1 {
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}

	nav.dirs = dirs
}

func newNav(height int) *nav {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("getting current directory: %s", err)
	}

	nav := &nav{
		dirChan:      make(chan *dir),
		regChan:      make(chan *reg),
		dirCache:     make(map[string]*dir),
		regCache:     make(map[string]*reg),
		saves:        make(map[string]bool),
		marks:        make(map[string]string),
		selections:   make(map[string]int),
		selectionInd: 0,
		height:       height,
	}

	nav.getDirs(wd)

	return nav
}

func (nav *nav) renew() {
	for _, d := range nav.dirs {
		go func(d *dir) {
			s, err := os.Stat(d.path)
			if err != nil {
				log.Printf("getting directory info: %s", err)
				return
			}
			if d.loadTime.After(s.ModTime()) {
				return
			}
			d.loadTime = time.Now()
			nd := newDir(d.path)
			nd.sort()
			nav.dirChan <- nd
		}(d)
	}

	for m := range nav.selections {
		if _, err := os.Stat(m); os.IsNotExist(err) {
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

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %s", err)
	}

	curr, err := nav.currFile()
	nav.getDirs(wd)
	if err == nil {
		last := nav.dirs[len(nav.dirs)-1]
		last.files = append(last.files, curr)
	}

	return nil
}

func (nav *nav) position() {
	path := nav.currDir().path
	for i := len(nav.dirs) - 2; i >= 0; i-- {
		nav.dirs[i].sel(filepath.Base(path), nav.height)
		path = filepath.Dir(path)
	}
}

func (nav *nav) preview() {
	curr, err := nav.currFile()
	if err != nil {
		return
	}

	var reader io.Reader

	if len(gOpts.previewer) != 0 {
		cmd := exec.Command(gOpts.previewer, curr.path, strconv.Itoa(nav.height))

		out, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("previewing file: %s", err)
		}

		if err := cmd.Start(); err != nil {
			log.Printf("previewing file: %s", err)
		}

		defer cmd.Wait()
		defer out.Close()
		reader = out
	} else {
		f, err := os.Open(curr.path)
		if err != nil {
			log.Printf("opening file: %s", err)
		}

		defer f.Close()
		reader = f
	}

	reg := &reg{loadTime: time.Now(), path: curr.path}

	buf := bufio.NewScanner(reader)

	for i := 0; i < nav.height && buf.Scan(); i++ {
		for _, r := range buf.Text() {
			if r == 0 {
				reg.lines = []string{"\033[1mbinary\033[0m"}
				nav.regChan <- reg
				return
			}
		}
		reg.lines = append(reg.lines, buf.Text())
	}

	if buf.Err() != nil {
		log.Printf("loading file: %s", buf.Err())
	}

	nav.regChan <- reg
}

func (nav *nav) loadReg(ui *ui, path string) *reg {
	r, ok := nav.regCache[path]
	if !ok {
		go nav.preview()
		r := &reg{loading: true, path: path}
		nav.regCache[path] = r
		return r
	}

	s, err := os.Stat(r.path)
	if err != nil {
		return r
	}

	if s.ModTime().After(r.loadTime) {
		r.loadTime = time.Now()
		go nav.preview()
	}

	return r
}

func (nav *nav) sort() {
	for _, d := range nav.dirs {
		name := d.name()
		d.sort()
		d.sel(name, nav.height)
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
	edge := min(min(nav.height/2, gOpts.scrolloff), dir.ind)
	dir.pos = max(dir.pos, edge)
}

func (nav *nav) down(dist int) {
	dir := nav.currDir()

	maxind := len(dir.files) - 1

	if dir.ind >= maxind {
		return
	}

	dir.ind += dist
	dir.ind = min(maxind, dir.ind)

	dir.pos += dist
	edge := min(min(nav.height/2, gOpts.scrolloff), maxind-dir.ind)

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

func (nav *nav) top() {
	dir := nav.currDir()

	dir.ind = 0
	dir.pos = 0
}

func (nav *nav) bottom() {
	dir := nav.currDir()

	dir.ind = len(dir.files) - 1
	dir.pos = min(dir.ind, nav.height-1)
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

	nav.down(1)
}

func (nav *nav) invert() {
	last := nav.currDir()
	for _, f := range last.files {
		path := filepath.Join(last.path, f.Name())
		nav.toggleSelection(path)
	}
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

func copyAsync(ui *ui, srcs []string, dstDir string) {
	echo := &callExpr{"echo", []string{""}, 1}

	_, err := os.Stat(dstDir)
	if os.IsNotExist(err) {
		echo.args[0] = fmt.Sprintf("error: %s", err)
		ui.exprChan <- echo
		return
	}

	total, err := copySize(srcs)
	if err != nil {
		echo.args[0] = fmt.Sprintf("error: %s", err)
		ui.exprChan <- echo
		return
	}

	nums, errs := copyAll(srcs, dstDir)

	curr := int64(0)
	lastUpdate := 0
	errCount := 0
loop:
	for {
		select {
		case n := <-nums:
			curr += n
			// n is usually 1024B so update roughly per 1024B x 1024 = 1MB copied
			if lastUpdate++; lastUpdate >= 1024 {
				lastUpdate = 0
				percentage := int((100 * float64(curr)) / float64(total))
				echo.args[0] = fmt.Sprintf("%d%%", percentage)
				ui.exprChan <- echo
			}
		case err, ok := <-errs:
			if !ok {
				break loop
			}
			errCount++
			echo.args[0] = fmt.Sprintf("[%d] error: %s", errCount, err)
			ui.exprChan <- echo
		}
	}

	if err := sendRemote("send load"); err != nil {
		errCount++
		echo.args[0] = fmt.Sprintf("[%d] error: %s", errCount, err)
		ui.exprChan <- echo
	}
}

func moveAsync(ui *ui, srcs []string, dstDir string) {
	echo := &callExpr{"echo", []string{""}, 1}

	_, err := os.Stat(dstDir)
	if os.IsNotExist(err) {
		echo.args[0] = fmt.Sprintf("error: %s", err)
		ui.exprChan <- echo
		return
	}

	errCount := 0
	for _, src := range srcs {
		dst := filepath.Join(dstDir, filepath.Base(src))

		_, err := os.Stat(dst)
		if !os.IsNotExist(err) {
			var newPath string
			for i := 1; !os.IsNotExist(err); i++ {
				newPath = fmt.Sprintf("%s.~%d~", dst, i)
				_, err = os.Stat(newPath)
			}
			dst = newPath
		}

		if err := os.Rename(src, dst); err != nil {
			errCount++
			echo.args[0] = fmt.Sprintf("[%d] error: %s", errCount, err)
			ui.exprChan <- echo
		}
	}

	if err := sendRemote("send load"); err != nil {
		errCount++
		echo.args[0] = fmt.Sprintf("[%d] error: %s", errCount, err)
		ui.exprChan <- echo
	}
}

func (nav *nav) paste(ui *ui) error {
	srcs, cp, err := loadFiles()
	if err != nil {
		return err
	}

	if len(srcs) == 0 {
		return errors.New("no file in copy/cut buffer")
	}

	dstDir := nav.currDir().path

	if cp {
		go copyAsync(ui, srcs, dstDir)
	} else {
		go moveAsync(ui, srcs, dstDir)
	}

	if err := saveFiles(nil, false); err != nil {
		return fmt.Errorf("clearing copy/cut buffer: %s", err)
	}

	if err := sendRemote("send sync"); err != nil {
		return fmt.Errorf("paste: %s", err)
	}

	return nil
}

func (nav *nav) del() error {
	list, err := nav.currFileOrSelections()

	if err != nil {
		return err
	}

	for _, path := range list {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

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

	return nil
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

	nav.getDirs(wd)

	return nil
}

func (nav *nav) sel(path string) error {
	lstat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("select: %s", err)
	}

	dir := filepath.Dir(path)

	if err := nav.cd(dir); err != nil {
		return fmt.Errorf("select: %s", err)
	}

	base := filepath.Base(path)

	last := nav.dirs[len(nav.dirs)-1]
	if last.loading {
		last.files = append(last.files, &file{FileInfo: lstat})
	} else {
		last.sel(base, nav.height)
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
	last := nav.currDir()
	for i := 0; i < len(last.files); i++ {
		if findMatch(last.files[i].Name(), nav.find) {
			count++
			if count > 1 {
				return count
			}
			index = i
		}
	}
	if count == 1 {
		if index > last.ind {
			nav.down(index - last.ind)
		} else {
			nav.up(last.ind - index)
		}
	}
	return count
}

func (nav *nav) findNext() bool {
	last := nav.currDir()
	for i := last.ind + 1; i < len(last.files); i++ {
		if findMatch(last.files[i].Name(), nav.find) {
			nav.down(i - last.ind)
			return true
		}
	}
	if gOpts.wrapscan {
		for i := 0; i < last.ind; i++ {
			if findMatch(last.files[i].Name(), nav.find) {
				nav.up(last.ind - i)
				return true
			}
		}
	}
	return false
}

func (nav *nav) findPrev() bool {
	last := nav.currDir()
	for i := last.ind - 1; i >= 0; i-- {
		if findMatch(last.files[i].Name(), nav.find) {
			nav.up(last.ind - i)
			return true
		}
	}
	if gOpts.wrapscan {
		for i := len(last.files) - 1; i > last.ind; i-- {
			if findMatch(last.files[i].Name(), nav.find) {
				nav.down(i - last.ind)
				return true
			}
		}
	}
	return false
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

func (nav *nav) searchNext() error {
	last := nav.currDir()
	for i := last.ind + 1; i < len(last.files); i++ {
		matched, err := searchMatch(last.files[i].Name(), nav.search)
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
			matched, err := searchMatch(last.files[i].Name(), nav.search)
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
		matched, err := searchMatch(last.files[i].Name(), nav.search)
		if err != nil {
			return err
		}
		if matched {
			nav.up(last.ind - i)
			return nil
		}
	}
	if gOpts.wrapscan {
		for i := len(last.files) - 1; i > last.ind; i-- {
			matched, err := searchMatch(last.files[i].Name(), nav.search)
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

func (nav *nav) readMarks() error {
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
	if len(nav.marks) == 0 {
		return nil
	}

	if err := nav.readMarks(); err != nil {
		return fmt.Errorf("reading marks file: %s", err)
	}

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
		keys = append(keys, k)
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

func (nav *nav) currDir() *dir {
	return nav.dirs[len(nav.dirs)-1]
}

func (nav *nav) currFile() (*file, error) {
	last := nav.dirs[len(nav.dirs)-1]

	if len(last.files) == 0 {
		return nil, fmt.Errorf("empty directory")
	}
	return last.files[last.ind], nil
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
	paths := make([]string, 0, len(nav.selections))
	indices := make([]int, 0, len(nav.selections))
	for path, index := range nav.selections {
		paths = append(paths, path)
		indices = append(indices, index)
	}
	sort.Sort(indexedSelections{paths: paths, indices: indices})
	return paths
}

func (nav *nav) currFileOrSelections() (list []string, err error) {
	if len(nav.selections) == 0 {
		curr, err := nav.currFile()
		if err != nil {
			return nil, errors.New("no file selected")
		}

		return []string{curr.path}, nil
	}

	return nav.currSelections(), nil
}
