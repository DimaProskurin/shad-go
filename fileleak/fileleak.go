//go:build !solution

package fileleak

import (
	"maps"
	"os"
)

type testingT interface {
	Errorf(msg string, args ...interface{})
	Cleanup(func())
}

const PATH = "/proc/self/fd/"

func VerifyNone(t testingT) {
	openedBefore := getOpenedFiles()
	t.Cleanup(func() {
		openedAfter := getOpenedFiles()
		if d := diff(openedAfter, openedBefore); len(d) > 0 {
			t.Errorf("leak %v", d)
		}
	})
}

func getOpenedFiles() map[string]int {
	entries, err := os.ReadDir(PATH)
	if err != nil {
		panic("doing os.ReadDir: " + err.Error())
	}
	opened := make(map[string]int)
	for _, e := range entries {
		linkV, err := os.Readlink(PATH + e.Name())
		if err != nil {
			continue
		}
		opened[linkV]++
	}
	return opened
}

func diff(a map[string]int, b map[string]int) map[string]int {
	d := make(map[string]int)
	maps.Copy(d, a)
	for k := range b {
		d[k] -= b[k]
		if d[k] == 0 {
			delete(d, k)
		}
	}
	return d
}
