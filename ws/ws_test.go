// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"
)

func kb(n uint64) uint64 { return n / (1 << 10) }
func gcandstat() string {
	runtime.GC()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	f := "alloc: %d/%d kb, heap: %d/%d kb, objs: %d, gcs: %d"
	return fmt.Sprintf(f, kb(mem.Alloc), kb(mem.TotalAlloc), kb(mem.HeapAlloc), kb(mem.HeapSys), mem.HeapObjects, mem.NumGC)
}
func TestWalkSrc(t *testing.T) {
	dirs := []string{runtime.GOROOT()}
	t.Log(dirs)
	w := New(Config{CapHint: 8000})
	start := time.Now()
	for i, err := range MountAll(w, dirs) {
		if err != nil {
			t.Errorf("error mounting %s: %s\n", dirs[i], err)
		}
	}
	first := w.all[NewId(dirs[0])]
	if len(first.Children) == 0 {
		t.Errorf("error mount children missing\n")
	}
	if first.Parent == nil || len(first.Parent.Children) == 0 {
		t.Errorf("error mount parent missing\n")
	}
	for id, r := range w.all {
		if rid := NewId(r.Path()); id != rid {
			t.Error(id, "!=", rid, r.Path())
		}
	}
	took := time.Since(start)
	t.Logf("count: %d, took: %s", len(w.all), took)
	t.Log(gcandstat())
	w.Close()
	if len(w.all) != 0 || w.watcher != nil {
		t.Error("not clean after close")
	}
	t.Log(gcandstat())
}

type testhandler chan struct {
	Op
	*Res
}

func (h testhandler) Handle(op Op, r *Res) {
	h <- struct {
		Op
		*Res
	}{op, r}
}

func TestWatch(t *testing.T) {
	dir, err := ioutil.TempDir("", "wsinotify")
	if err != nil {
		t.Fatal("failed to create temp dir")
	}
	events := make(testhandler, 10)
	w := New(Config{CapHint: 100, Watcher: NewInotify, Handler: events})
	defer w.Close()
	fail := func(msg string, err error) {
		if err != nil {
			os.RemoveAll(dir)
			t.Fatalf("%s: %s", msg, err)
		}
	}
	expect := func(path string, ops ...Op) {
		for _, op := range ops {
			select {
			case e := <-events:
				p := e.Path()
				if e.Op != op || p != path {
					t.Errorf("expected event %x %q got %x %q\n", op, path, e.Op, p)
				}
			case <-time.After(1 * time.Second):
				t.Fatalf("expected event %x %q\n got timeout", op, path)
			}
		}
	}
	r, err := w.Mount(dir)
	fail("mount", err)
	expect(dir, Add, Change)
	createdir := func(name string) string {
		subdir := dir + name
		err := os.Mkdir(subdir, 0777)
		fail("subdir", err)
		return subdir
	}
	createfile := func(name string) string {
		file := dir + name
		f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0666)
		fail("create testfile", err)
		f.Close()
		return file
	}
	file := createfile("/testfile")
	expect(file, Add|Create, Change|Modify)

	os.Remove(file)
	expect(file, Remove|Delete)

	file = createfile("/testfile")
	expect(file, Add|Create, Change|Modify)

	subdir := createdir("/testdir")
	expect(subdir, Add|Create, Change|Create)

	subsubdir := createdir("/testdir/sub")
	expect(subsubdir, Add|Create, Change|Create)

	otherfile := createfile("/testdir/sub/otherfile")
	expect(otherfile, Add|Create, Change|Modify)

	err = os.Rename(subsubdir, dir+"/sub")
	fail("mv", err)
	expect(otherfile, Remove|Delete)
	expect(subsubdir, Remove|Delete)
	subsubdir = dir + "/sub"
	otherfile = dir + "/sub/otherfile"
	expect(subsubdir, Add|Create)
	expect(otherfile, Add|Create)
	expect(subsubdir, Change|Create)

	os.RemoveAll(dir)
	expect(file, Remove|Delete)
	expect(subdir, Remove|Delete)
	expect(otherfile, Remove|Delete)
	expect(subsubdir, Remove|Delete)
	expect(dir, Remove|Delete)

	if _, ok := w.all[r.Id]; ok {
		t.Error("dir still exists after remove")
	}
	w.Close()
}

func TestSplitPath(t *testing.T) {
	expect := []string{"c", "b", "a"}
	w := &Ws{fs: Filesystem}
	for i, p := range w.split("/a/b/c") {
		if e := expect[i]; p != e {
			fmt.Printf("%s != %s\n", p, e)
		}
	}
}
