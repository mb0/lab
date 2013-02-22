package ws

import (
	"fmt"
	"go/build"
	"runtime"
	"sync"
	"testing"
	"time"
)

func mountAllSeq(w *Ws, dirs []string) {
	for _, path := range dirs {
		_, err := w.Mount(path)
		if err != nil {
			fmt.Println(err)
		}
	}
}
func mountAllPar(w *Ws, dirs []string) {
	var wg sync.WaitGroup
	wg.Add(len(dirs))
	for _, path := range dirs {
		go func(path string, wg *sync.WaitGroup) {
			_, err := w.Mount(path)
			if err != nil {
				fmt.Println(err)
			}
			wg.Done()
		}(path, &wg)
	}
	wg.Wait()
}

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
	if runtime.GOMAXPROCS(0) > 1 {
		mountAllPar(w, dirs)
	} else {
		mountAllSeq(w, dirs)
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
