package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type watch struct {
	watcher   *fsnotify.Watcher
	events    <-chan fsnotify.Event
	quit      chan struct{}
	pending   map[watchUpdate]bool
	timeout   chan watchUpdate
	dirChan   chan<- *dir
	fileChan  chan<- *file
	delChan   chan<- string
	paths     map[string]os.FileInfo
	pathsLock sync.Mutex
}

func newWatch(dirChan chan<- *dir, fileChan chan<- *file, delChan chan<- string) *watch {
	watch := &watch{
		quit:     make(chan struct{}),
		pending:  make(map[watchUpdate]bool),
		timeout:  make(chan watchUpdate, 1024),
		dirChan:  dirChan,
		fileChan: fileChan,
		delChan:  delChan,
		paths:    make(map[string]os.FileInfo),
	}

	return watch
}

func (watch *watch) start() {
	if watch.watcher != nil {
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("start watcher: %s", err)
		return
	}

	watch.watcher = watcher
	watch.events = watcher.Events

	go watch.loop()
}

func (watch *watch) stop() {
	if watch.watcher == nil {
		return
	}

	watch.quit <- struct{}{}
	watch.watcher.Close()

	watch.watcher = nil
	watch.events = nil
}

func (watch *watch) add(path string) {
	if watch.watcher == nil {
		return
	}

	// ignore /dev since write updates to /dev/tty causes high cpu usage
	if path == "/dev" {
		return
	}

	// the stat is kept so it is not repeated for every watched path on every event
	stat, err := os.Stat(path)
	if err != nil {
		log.Printf("watch path %s: %s", path, err)
		return
	}

	if err := watch.watcher.Add(path); err != nil {
		log.Printf("watch path %s: %s", path, err)
		return
	}

	watch.pathsLock.Lock()
	watch.paths[path] = stat
	watch.pathsLock.Unlock()
}

func (watch *watch) loop() {
	for {
		select {
		case ev := <-watch.events:
			if ev.Has(fsnotify.Create) {
				for _, path := range watch.getSameDirs(filepath.Dir(ev.Name)) {
					watch.addUpdate(watchUpdate{"dir", path})
					watch.addUpdate(watchUpdate{"file", path})
				}
			}

			if ev.Has(fsnotify.Remove) || ev.Has(fsnotify.Rename) {
				dir, file := filepath.Split(ev.Name)
				for _, path := range watch.getSameDirs(dir) {
					watch.delChan <- filepath.Join(path, file)
					watch.addUpdate(watchUpdate{"dir", path})
					watch.addUpdate(watchUpdate{"file", path})
				}
			}

			if ev.Has(fsnotify.Write) || ev.Has(fsnotify.Chmod) {
				// skip write updates for the log file, otherwise it is possible
				// to have an infinite loop where writing to the log file causes
				// it to be reloaded, which in turn triggers more events that
				// are then logged
				if ev.Name == gLogPath && ev.Has(fsnotify.Write) {
					continue
				}

				dir, file := filepath.Split(ev.Name)
				for _, path := range watch.getSameDirs(dir) {
					watch.addUpdate(watchUpdate{"file", filepath.Join(path, file)})
				}
			}
		case update := <-watch.timeout:
			if watch.pending[update] {
				watch.processUpdate(update)
				time.AfterFunc(100*time.Millisecond, func() { watch.timeout <- update })
				watch.pending[update] = false
			} else {
				delete(watch.pending, update)
			}
		case <-watch.quit:
			watch.pathsLock.Lock()
			clear(watch.paths)
			watch.pathsLock.Unlock()
			return
		}
	}
}

type watchUpdate struct {
	kind string
	path string
}

func (watch *watch) addUpdate(update watchUpdate) {
	// process an update immediately if is the first time, otherwise store it
	// and process only after a timeout to reduce the number of actual loads
	if _, ok := watch.pending[update]; !ok {
		watch.processUpdate(update)
		time.AfterFunc(100*time.Millisecond, func() { watch.timeout <- update })
		watch.pending[update] = false
	} else {
		watch.pending[update] = true
	}
}

func (watch *watch) processUpdate(update watchUpdate) {
	switch update.kind {
	case "dir":
		if _, err := os.Lstat(update.path); err == nil {
			watch.dirChan <- newDir(update.path)
		}
	case "file":
		if _, err := os.Lstat(update.path); err == nil {
			watch.fileChan <- newFile(update.path)
		}
	}
}

// fsnotify silently aliases watches by inode, so two paths to the same
// directory (a symlink and its target) end up as a single entry in
// watcher.WatchList. Track every Added path ourselves so getSameDirs can
// fan events out to all aliases.
func (watch *watch) getSameDirs(dir string) []string {
	var paths []string

	dirStat, err := os.Stat(dir)
	if err != nil {
		return nil
	}

	// the kept stats tell which paths could be the same directory without asking the disk
	var candidates []string
	watch.pathsLock.Lock()
	for path, stat := range watch.paths {
		if path == dir || os.SameFile(stat, dirStat) {
			candidates = append(candidates, path)
		}
	}
	watch.pathsLock.Unlock()

	for _, path := range candidates {
		if path == dir {
			paths = append(paths, path)
			continue
		}

		// a kept stat can be out of date, so the match is confirmed
		if stat, err := os.Stat(path); err == nil && os.SameFile(stat, dirStat) {
			paths = append(paths, path)
		}
	}

	return paths
}
