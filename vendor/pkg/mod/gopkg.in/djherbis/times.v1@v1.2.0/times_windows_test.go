package times

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestStatFile(t *testing.T) {
	fileTest(t, func(f *os.File) {
		ts, err := StatFile(f)
		if err != nil {
			t.Error(err.Error())
		}
		timespecTest(ts, newInterval(time.Now(), time.Second), t)
	})
}

func TestStatFileErr(t *testing.T) {
	fileTest(t, func(f *os.File) {
		f.Close()

		_, err := StatFile(f)
		if err == nil {
			t.Error("got nil err, but err was expected!")
		}
	})
}

func TestStatFileProcErr(t *testing.T) {
	fileTest(t, func(f *os.File) {
		findProcErr = errors.New("fake error")
		defer func() { findProcErr = nil }()

		_, err := StatFile(f)
		if err == nil {
			t.Error("got nil err, but err was expected!")
		}
	})
}

func TestStatBadNameErr(t *testing.T) {
	_, err := platformSpecficStat(string([]byte{0}))
	if err != syscall.EINVAL {
		t.Error(err)
	}
}

func TestStatProcErrFallback(t *testing.T) {
	fileAndDirTest(t, func(name string) {
		findProcErr = errors.New("fake error")
		defer func() { findProcErr = nil }()

		ts, err := Stat(name)
		if err != nil {
			t.Error(err.Error())
		}
		timespecTest(ts, newInterval(time.Now(), time.Second), t)
	})
}

func TestLstatProcErrFallback(t *testing.T) {
	fileAndDirTest(t, func(name string) {
		findProcErr = errors.New("fake error")
		defer func() { findProcErr = nil }()

		ts, err := Lstat(name)
		if err != nil {
			t.Error(err.Error())
		}
		timespecTest(ts, newInterval(time.Now(), time.Second), t)
	})
}
