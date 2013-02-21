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
func TestWalkSrc(t *testing.T) {
	dirs := build.Default.SrcDirs()
	t.Log(dirs)
	w := New()
	start := time.Now()
	if runtime.GOMAXPROCS(0) > 1 {
		mountAllPar(w, dirs)
	} else {
		mountAllSeq(w, dirs)
	}
	for p, r := range w.all {
		if p != r.Path() {
			t.Error(p, "!=", r.Path())
		}
	}
	took := time.Since(start)
	runtime.GC()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	f := "count: %d, took: %s, alloc: %d/%d kb, heap: %d/%d kb, objs: %d, gcs: %d"
	kb := func(n uint64) uint64 { return n / (1 << 10) }
	t.Logf(f, len(w.all), took, kb(mem.Alloc), kb(mem.TotalAlloc), kb(mem.HeapAlloc), kb(mem.HeapSys), mem.HeapObjects, mem.NumGC)
}
