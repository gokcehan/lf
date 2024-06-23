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

func (watch *watch) set(paths map[string]bool) {
	if watch.watcher == nil {
		return
	}

	for _, path := range watch.watcher.WatchList() {
		if !paths[path] {
			watch.watcher.Remove(path)
		}
	}

	for path := range paths {
		watch.watcher.Add(path)
	}
}

func (watch *watch) loop() {
	for {
		select {
		case ev := <-watch.events:
			if ev.Has(fsnotify.Create) {
				dir := filepath.Dir(ev.Name)
				watch.addLoad(dir)
				watch.addUpdate(dir)
			}

			if ev.Has(fsnotify.Remove) || ev.Has(fsnotify.Rename) {
				watch.delChan <- ev.Name
				dir := filepath.Dir(ev.Name)
				watch.addLoad(dir)
				watch.addUpdate(dir)
			}

			if ev.Has(fsnotify.Write) || ev.Has(fsnotify.Chmod) {
				watch.addUpdate(ev.Name)
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
			watch.loads = make(map[string]bool)
		case <-watch.updateTimer.C:
			for path := range watch.updates {
				if _, err := os.Lstat(path); err != nil {
					continue
				}
				watch.fileChan <- newFile(path)
			}
			watch.updates = make(map[string]bool)
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
