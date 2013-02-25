package ws

import (
	"fmt"
	"go/build"
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
	dirs := build.Default.SrcDirs()
	t.Log(dirs)
	w := New(Config{CapHint: 8000})
	start := time.Now()
	for i, err := range MountAll(w, dirs) {
		if err != nil {
			fmt.Printf("error mounting %s: %s\n", dirs[i], err)
		}
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
func TestWatch(t *testing.T) {
	dir, err := ioutil.TempDir("", "wsinotify")
	if err != nil {
		t.Fatal("failed to create temp dir")
	}
	events := make(chan struct {
		Op
		*Res
	}, 10)
	w := New(Config{CapHint: 100, Watcher: NewInotify, Handler: func(op Op, r *Res) {
		events <- struct {
			Op
			*Res
		}{op, r}
	}})
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
	file := dir + "/testfile"
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0666)
	fail("create testfile", err)
	f.Close()
	expect(file, Add|Create, Change|Modify)
	os.Remove(file)
	expect(file, Remove|Delete)
	f, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0666)
	fail("recreate testfile", err)
	f.Close()
	expect(file, Add|Create, Change|Modify)
	os.RemoveAll(dir)
	expect(file, Remove|Delete)
	expect(dir, Remove|Delete)
	if _, ok := w.all[r.Id]; ok {
		t.Error("dir still exists after remove")
	}
	w.Close()
}

func TestSplitPath(t *testing.T) {
	expect := []string{"c", "b", "a"}
	for i, p := range split("/a/b/c") {
		if e := expect[i]; p != e {
			fmt.Printf("%s != %s\n", p, e)
		}
	}
}
