package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type watchThrottle struct {
	callback   func(string)
	duration   time.Duration
	updates    map[string]struct{}
	timer      *time.Timer
	throttling bool
}

func newWatchThrottle(callback func(string), duration time.Duration) *watchThrottle {
	return &watchThrottle{
		callback: callback,
		duration: duration,
		updates:  make(map[string]struct{}),
		timer:    time.NewTimer(0),
	}
}

func (throttle *watchThrottle) addUpdate(path string) {
	if throttle.throttling {
		throttle.updates[path] = struct{}{}
	} else {
		throttle.callback(path)
		throttle.timer.Reset(throttle.duration)
		throttle.throttling = true
	}
}

func (throttle *watchThrottle) onTimeout() {
	if len(throttle.updates) > 0 {
		for path := range throttle.updates {
			throttle.callback(path)
		}
		clear(throttle.updates)
		throttle.timer.Reset(throttle.duration)
	} else {
		throttle.throttling = false
	}
}

type watch struct {
	watcher      *fsnotify.Watcher
	events       <-chan fsnotify.Event
	quit         chan struct{}
	dirThrottle  *watchThrottle
	fileThrottle *watchThrottle
	dirChan      chan<- *dir
	fileChan     chan<- *file
	delChan      chan<- string
}

func newWatch(dirChan chan<- *dir, fileChan chan<- *file, delChan chan<- string) *watch {
	watch := &watch{
		quit:     make(chan struct{}),
		dirChan:  dirChan,
		fileChan: fileChan,
		delChan:  delChan,
	}

	watch.dirThrottle = newWatchThrottle(watch.processDir, 500*time.Millisecond)
	watch.fileThrottle = newWatchThrottle(watch.processFile, 500*time.Millisecond)
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
	if path != "/dev" {
		if err := watch.watcher.Add(path); err != nil {
			log.Printf("watch path %s: %s", path, err)
		}
	}
}

func (watch *watch) loop() {
	for {
		select {
		case ev := <-watch.events:
			if ev.Has(fsnotify.Create) {
				for _, path := range watch.getSameDirs(filepath.Dir(ev.Name)) {
					watch.dirThrottle.addUpdate(path)
					watch.fileThrottle.addUpdate(path)
				}
			}

			if ev.Has(fsnotify.Remove) || ev.Has(fsnotify.Rename) {
				dir, file := filepath.Split(ev.Name)
				for _, path := range watch.getSameDirs(dir) {
					watch.delChan <- filepath.Join(path, file)
					watch.dirThrottle.addUpdate(path)
					watch.fileThrottle.addUpdate(path)
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
					watch.fileThrottle.addUpdate(filepath.Join(path, file))
				}
			}
		case <-watch.dirThrottle.timer.C:
			watch.dirThrottle.onTimeout()
		case <-watch.fileThrottle.timer.C:
			watch.fileThrottle.onTimeout()
		case <-watch.quit:
			return
		}
	}
}

func (watch *watch) processDir(path string) {
	if _, err := os.Lstat(path); err == nil {
		dir := newDir(path)
		dir.sort()
		watch.dirChan <- dir
	}
}

func (watch *watch) processFile(path string) {
	if _, err := os.Lstat(path); err == nil {
		watch.fileChan <- newFile(path)
	}
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
