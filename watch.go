package main

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

type watch struct {
	watcher     *fsnotify.Watcher
	events      <-chan fsnotify.Event
	loads       map[string]bool
	loadTimer   *time.Timer
	updates     map[string]bool
	updateTimer *time.Timer
}

func newWatch() *watch {
	return &watch{
		loads:       make(map[string]bool),
		loadTimer:   time.NewTimer(0),
		updates:     make(map[string]bool),
		updateTimer: time.NewTimer(0),
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
}

func (watch *watch) stop() {
	if watch.watcher == nil {
		return
	}

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
