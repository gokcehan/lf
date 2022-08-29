package times

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStat(t *testing.T) {
	fileAndDirTest(t, func(name string) {
		ts, err := Stat(name)
		if err != nil {
			t.Error(err.Error())
		}
		timespecTest(ts, newInterval(time.Now(), time.Second), t)
	})
}

func TestGet(t *testing.T) {
	fileAndDirTest(t, func(name string) {
		fi, err := os.Stat(name)
		if err != nil {
			t.Error(err.Error())
		}
		timespecTest(Get(fi), newInterval(time.Now(), time.Second), t)
	})
}

type tsFunc func(string) (Timespec, error)

var offsetTime = -10 * time.Second

func TestStatSymlink(t *testing.T) {
	testStatSymlink(Stat, time.Now().Add(offsetTime), t)
}

func TestLstatSymlink(t *testing.T) {
	testStatSymlink(Lstat, time.Now(), t)
}

func testStatSymlink(sf tsFunc, expectTime time.Time, t *testing.T) {
	fileAndDirTest(t, func(name string) {
		start := time.Now()

		symname := filepath.Join(filepath.Dir(name), "sym-"+filepath.Base(name))
		if err := os.Symlink(name, symname); err != nil {
			t.Error(err.Error())
		}
		defer os.Remove(symname)

		// modify the realFileTime so symlink and real file see diff values.
		realFileTime := start.Add(offsetTime)
		if err := os.Chtimes(name, realFileTime, realFileTime); err != nil {
			t.Error(err.Error())
		}

		ts, err := sf(symname)
		if err != nil {
			t.Error(err.Error())
		}
		timespecTest(ts, newInterval(expectTime, time.Second), t, Timespec.AccessTime, Timespec.ModTime)
	})
}

func TestStatErr(t *testing.T) {
	_, err := Stat("badfile?")
	if err == nil {
		t.Error("expected an error")
	}
}

func TestLstatErr(t *testing.T) {
	_, err := Lstat("badfile?")
	if err == nil {
		t.Error("expected an error")
	}
}

func TestCheat(t *testing.T) {
	// not all times are available for all platforms
	// this allows us to get 100% test coverage for platforms which do not have
	// ChangeTime/BirthTime
	var c ctime
	if c.HasChangeTime() {
		c.ChangeTime()
	}

	var b btime
	if b.HasBirthTime() {
		b.BirthTime()
	}

	var paniced = false
	var nc noctime
	func() {
		if !nc.HasChangeTime() {
			defer func() {
				recover()
				paniced = true
			}()
		}
		nc.ChangeTime()
	}()

	if !paniced {
		t.Error("expected panic")
	}

	paniced = false
	var nb nobtime
	func() {
		if !nb.HasBirthTime() {
			defer func() {
				recover()
				paniced = true
			}()
		}
		nb.BirthTime()
	}()

	if !paniced {
		t.Error("expected panic")
	}
}
