package times

import (
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"
)

type timeRange struct {
	start time.Time
	end   time.Time
}

func newInterval(t time.Time, dur time.Duration) timeRange {
	return timeRange{start: t.Add(-dur), end: t.Add(dur)}
}

func (t timeRange) Contains(findTime time.Time) bool {
	return !findTime.Before(t.start) && !findTime.After(t.end)
}

func fileAndDirTest(t testing.TB, testFunc func(name string)) {
	filenameTest(t, testFunc)
	dirTest(t, testFunc)
}

// creates a file and cleans it up after the test is run.
func fileTest(t testing.TB, testFunc func(f *os.File)) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	testFunc(f)
}

func filenameTest(t testing.TB, testFunc func(filename string)) {
	fileTest(t, func(f *os.File) {
		testFunc(f.Name())		
	})
}

// creates a dir and cleans it up after the test is run.
func dirTest(t testing.TB, testFunc func(dirname string)) {
	dirname, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dirname)
	testFunc(dirname)
}

type timeFetcher func(Timespec) time.Time

func timespecTest(ts Timespec, r timeRange, t testing.TB, getTimes ...timeFetcher) {
	if len(getTimes) == 0 {
		getTimes = append(getTimes, Timespec.AccessTime, Timespec.ModTime)

		if ts.HasChangeTime() {
			getTimes = append(getTimes, Timespec.ChangeTime)
		}

		if ts.HasBirthTime() {
			getTimes = append(getTimes, Timespec.BirthTime)
		}
	}

	for _, getTime := range getTimes {
		if !r.Contains(getTime(ts)) {
			t.Errorf("expected %s=%s to be in range: \n[%s, %s]\n", GetFunctionName(getTime), getTime(ts), r.start, r.end)
		}
	}
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
