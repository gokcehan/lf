package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type watch struct {
	watcher     *fsnotify.Watcher
	events      <-chan fsnotify.Event
	quit        chan struct{}
	loads       map[string]bool
	loadTimer   *time.Timer
	updates     map[string]bool
	updateTimer *time.Timer
	dirChan     chan<- *dir
	fileChan    chan<- *file
	delChan     chan<- string
}

func newWatch(dirChan chan<- *dir, fileChan chan<- *file, delChan chan<- string) *watch {
	return &watch{
		quit:        make(chan struct{}),
		loads:       make(map[string]bool),
		loadTimer:   time.NewTimer(0),
		updates:     make(map[string]bool),
		updateTimer: time.NewTimer(0),
		dirChan:     dirChan,
		fileChan:    fileChan,
		delChan:     delChan,
	}
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
	if path != "/dev" {
		watch.watcher.Add(path)
	}
}

func (watch *watch) loop() {
	for {
		select {
		case ev := <-watch.events:
			if ev.Has(fsnotify.Create) {
				for _, path := range watch.getSameDirs(filepath.Dir(ev.Name)) {
					watch.addLoad(path)
					watch.addUpdate(path)
				}
			}

			if ev.Has(fsnotify.Remove) || ev.Has(fsnotify.Rename) {
				dir, file := filepath.Split(ev.Name)
				for _, path := range watch.getSameDirs(dir) {
					watch.delChan <- filepath.Join(path, file)
					watch.addLoad(path)
					watch.addUpdate(path)
				}
			}

			if ev.Has(fsnotify.Write) || ev.Has(fsnotify.Chmod) {
				// skip updates for the log file, otherwise it is possible to
				// have an infinite loop where writing to the log file causes it
				// to be reloaded, which in turn triggers more events that are
				// then logged
				if ev.Name == gLogPath && ev.Has(fsnotify.Write) {
					continue
				}

				dir, file := filepath.Split(ev.Name)
				for _, path := range watch.getSameDirs(dir) {
					watch.addUpdate(filepath.Join(path, file))
				}
			}
		case <-watch.loadTimer.C:
			for path := range watch.loads {
				if _, err := os.Lstat(path); err != nil {
					continue
				}
				dir := newDir(path)
				dir.sort()
				watch.dirChan <- dir
			}
			clear(watch.loads)
		case <-watch.updateTimer.C:
			for path := range watch.updates {
				if _, err := os.Lstat(path); err != nil {
					continue
				}
				watch.fileChan <- newFile(path)
			}
			clear(watch.updates)
		case <-watch.quit:
			return
		}
	}
}

func (watch *watch) addLoad(path string) {
	if len(watch.loads) == 0 {
		watch.loadTimer.Stop()
		watch.loadTimer.Reset(10 * time.Millisecond)
	}
	watch.loads[path] = true
}

func (watch *watch) addUpdate(path string) {
	if len(watch.updates) == 0 {
		watch.updateTimer.Stop()
		watch.updateTimer.Reset(10 * time.Millisecond)
	}
	watch.updates[path] = true
}

// Hacky workaround since fsnotify reports changes for only one path if a
// directory is located at more than one path (e.g. bind mounts).
func (watch *watch) getSameDirs(dir string) []string {
	var paths []string

	dirStat, err := os.Stat(dir)
	if err != nil {
		return nil
	}

	for _, path := range watch.watcher.WatchList() {
		if path == dir {
			paths = append(paths, path)
			continue
		}

		stat, err := os.Stat(path)
		if err != nil {
			continue
		}

		if os.SameFile(stat, dirStat) {
			paths = append(paths, path)
		}
	}

	return paths
}
